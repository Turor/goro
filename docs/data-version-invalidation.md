# Live data invalidation (design)

Status: implemented in goro (`res` streaming layer) and remoteclient2
(`DATA_VERSION=1`: header, `/version`, `/changes`, `/data-manifest`). Enable
the remoteclient2 flag separately; old clients ignore the header. Signed GDAT1
manifests are produced offline with `goro-sign-release sign-data` (same
Ed25519 key as client updates) and copied onto the data host. The signing
seed never lives on the data server.

## Goal

When the data server's files change (a new skill table overlay, a fixed
translation, a new map), every running client should drop just the changed
files from its streaming cache and live-reload what it built from them, such
as the skill trees, without a restart and without re-downloading everything.

## Decision

- **remoteclient2 stamps a data version on every response**:
  `X-Data-Version: <n>`, a monotonic unsigned 64-bit integer.
- **Clients track the newest version they have seen.** When any response
  carries a version newer than the one their cache is consistent with, they
  ask once for the delta: `GET <data-url>changes?since=<cachedVersion>`.
- **Idle clients poll** a tiny endpoint every 60 to 120 seconds (random
  jitter) just to read the header.
- **No Hercules packet.** The header plus the poll covers active and idle
  clients alike, so the map server needs no change and the custom packet
  range 0x0EF0 to 0x0EFF stays untouched. (An earlier draft proposed a
  version-only nudge packet; it is dropped.)

## remoteclient2

### Version and change log

- State lives next to the served tree: `state/data-version.json`
  (`{"version": n, "snapshot": "<hash>"}`) and an append-only change log,
  `state/changes.jsonl`, one line per version:
  `{"version": n, "at": "<RFC 3339>", "paths": ["data/...", "System/..."], "full": false}`.
- **Bumping.** remoteclient2 already hashes the served tree for `list-files`
  (the archive index ETag). Extend that scan into a per-file snapshot
  (server path to size, mtime and content hash, plus the entry hashes inside
  each GRF). It rescans when its file watcher fires on the overlay directories
  or archives, debounced about 5 seconds so a multi-file upload becomes one
  version, and on start. If the snapshot differs from the last one, it writes
  version n+1 with the changed paths: loose overlay files, `System/` files,
  and changed GRF entries listed by their entry path. A replaced archive whose
  entries cannot be diffed cheaply is logged as `"full": true`.
- **Monotonic across restarts and lost state.** The first version is the
  Unix time in seconds, and a missing state file restarts from
  `max(now, last known)`. A reset therefore never goes backwards; it looks
  like a version too old for the change log (see below).
- **Retention.** Keep the last 1000 versions or 30 days, whichever is
  larger. Older lines are compacted away.

### Endpoints and headers

- `X-Data-Version: <n>` on **every** response: files, 304s, 404s,
  `list-files`, `changes` and `version`. It is also listed in
  `Access-Control-Expose-Headers`, so roBrowser can read it too.
- `GET changes?since=<v>` returns JSON:
  - `{"version": n, "paths": [...]}`, the union of the paths changed in
    versions v+1 through n, de-duplicated (at most a few thousand paths;
    beyond that it answers `full`);
  - `{"version": n, "full": true}` when v is older than the retained log, in
    the future (a server reset), or the range holds a `full` entry: "too old,
    revalidate everything";
  - `{"version": n, "paths": []}` when v is already current.
  Responses carry `Cache-Control: public, max-age=5`, so a burst of clients
  asking about the same version costs one computation behind a cache.
- `GET version` (or `HEAD` on it) returns `{"version": n}`, a few bytes with
  `Cache-Control: no-cache`. That is the poll target.
- Paths use the server spelling (`data/...` lowercase, `System/...`), the same
  names `list-files` and the file URLs use, including the Turoran overlay
  names (`skillinfolist.turoran13.lub`).

## goro client

### Tracking the version

- `remoteSource` parses `X-Data-Version` from every response (file GETs,
  revalidations, `list-files`) into an atomic `seenVersion`, keeping the max.
- `cachedVersion`, the version the cache is consistent with, is persisted
  in the cache directory (`data-version.json`, keyed by host like the rest of
  the cache). A fresh cache adopts the first version it sees without a delta,
  because everything in it was just downloaded.

### Delta fetch, one at a time

- When `seenVersion > cachedVersion`, a single background worker wakes
  (singleflight: at most one `changes` request in flight). It waits about
  2 seconds of debounce, so a burst of responses during a map load triggers
  one request, then asks `changes?since=cachedVersion`.
- The unsigned JSON cursor follows the same rollback rule as the signed
  manifest: a `version` that is not strictly greater than `cachedVersion`
  is ignored, and the worker stops instead of issuing another GET in a
  tight loop. The next wake is a later header or the idle poll.
- **Paths and full** are applied only after `GET data-manifest` verifies
  and the accepted manifest version is at least the changes version. Until
  then nothing is evicted and `cachedVersion` stays put.
- **Paths:** for each path, evict it from the streaming cache: mark it stale
  and forget any cached "not on server" answer. Then queue it for the game
  thread with `Manager.QueueDataInvalidation(paths...)`. Persist
  `cachedVersion = version` only after that, and only when it is greater
  than the cursor already stored, so a crash re-asks.
- **Full:** mark every cached path stale (one generation counter that each
  entry's metadata is compared against), reload the archive index, and queue
  an invalidation of every loaded table.
- If versions moved on while the request ran, the worker loops again after
  the same debounce it uses before the first request. It never runs two
  requests at once. Failures retry with the offline back-off.
- Accepting a newer signed manifest bumps that same cache generation, so a
  body stored before the manifest is revalidated before reuse. A conditional
  304 still has its cached bytes digest-checked; a mismatch is deleted and
  not returned.

### Interaction with ETag caching

- Eviction is lazy and conditional. The cached bytes and their ETag stay on
  disk, and a stale path's next read is an `If-None-Match` request: an
  unchanged file (say a path touched but identical) costs a 304 and no body.
- Only tables the client already built (for example the skill trees) are
  re-read at once, on the game thread. Every other changed file is
  revalidated the next time something asks for it. A delta listing 500
  textures does not trigger 500 downloads.
- ETags stay the per-file truth. The version is only a hint about when to
  ask. The existing once-per-run background refresh of text tables stays as
  a safety net: if it brings a changed table, it already queues the same
  invalidation.

### Background poll for idle clients

- Every 60 to 120 seconds, uniformly jittered so clients do not synchronise,
  the client sends `GET version` (or `HEAD`). It reads only the header or the
  tiny body. If that shows a newer version, it follows the delta path above.
- The poll is skipped when any other response in that interval already
  carried the header.
- It is **off when streaming is off**: no `remote_url`, or `remote = off`.
- **Offline:** it reuses the streaming layer's offline back-off and
  doubles the interval up to 10 minutes while the server is unreachable.
  It resets on the first successful response.
- It runs at most one poll and at most one delta at a time, both on the
  streaming layer's bounded background slots, so it never competes with
  frame-critical downloads.

### Client hook (shipped with the Lua skill trees)

- `Manager.QueueDataInvalidation(paths...)` is safe from any goroutine. The
  delta worker calls it.
- `Game.Update` calls `Manager.ApplyDataInvalidations()` once a frame on the
  game thread. It costs one mutex check when nothing is queued.
- `Manager.InvalidateData(paths)` evicts each path from the streaming cache
  and, if any is a skill table (`IsSkillDataFile`, which matches by stem so
  the Turoran overlay names count), calls `Manager.ReloadSkillData()`. That
  clears the load-once flags, rebuilds the tree, validates it, and installs it
  with `db.SetSkillTreeData`. The new tree generation makes open skill windows
  rebuild. Effect overlays (`effects.json`, `ez2streffect.bson`) call
  `Manager.ReloadEffectData()` the same way. Other loaded tables (item names,
  map names) can register later.

## Rollout

1. remoteclient2: version state, change log, header, `changes` and
   `version` endpoints, behind a config flag. Check it with curl against the
   test data tree.
2. goro: header parsing, persisted `cachedVersion`, the delta worker and the
   poll, all calling the hook above. Test them against a fake server that
   bumps versions.
3. Enable the remoteclient2 flag. Old clients ignore the header.
