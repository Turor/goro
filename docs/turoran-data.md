# TuroranRO data directory

## Streaming (players: nothing to install)

goro streams client data from the TuroranRO asset server (remoteclient2,
`https://ro.turoran.com/client/`), the same source roBrowser and Korangar
read. With no GRF in the data directory (`[data] remote = auto`, the
default) every file the directory lacks is fetched on first use and kept in
`<user cache>/goro/remote/<host>/` (`$XDG_CACHE_HOME`, `%LocalAppData%`).

- **Priority.** Loose files in `--data-dir`, then local archives, then the
  server. The server applies roBrowser's order itself: `System/`, the text
  overlays (`OVERLAY_PATH`), then slot-0 `translation.grf` and the other
  `DATA.INI` archives. The Turoran table names below are tried before the
  canonical ones (`res.TuroranRemoteAliases`; a test keeps it equal to the
  assemble script).
- **Paths.** Names are decoded from EUC-KR to UTF-8 and percent-encoded per
  segment (`data/texture/%EC%9C%A0...`), as roBrowser requests them. Only
  `data/` and `System/` are requested.
- **Startup.** Nothing waits for the network. The server's archive name
  list (`GET list-files`, 22 MB JSON, about 1.4 MB gzip on the wire) loads in
  the background from the cache, then is revalidated with its ETag (an
  unchanged list is not reparsed or rewritten). Once it is in, archive files
  it does not list are not requested at all. Text tables
  (`.lub .lua .txt .xml .json`) are always eligible because overlays are not
  in that list. `resnametable.txt` is reloaded when streaming starts, so its
  aliases come from the server too.
- **Frames.** Map loading prefetches each stage (gnd/rsw/gat, then models,
  then textures) in parallel in the background; the loading screen keeps
  drawing until the batch is done, and the load tasks then read the cache.
  Per-frame lookups (item and skill icons, chat-room boards) wait at most
  15 ms (`scene.FrameRemoteBudget`), draw without the asset, and pick it up
  on a later frame (`ReadFileCandidatesWithin`, `ErrRemotePending`). Other
  game-thread lookups outside the loading screen (for example the sprites of
  an actor that walks in) still block until their file arrives; they are
  small and cached after the first time.
- **Freshness.** Archive files are reused from the cache and revalidated
  (`If-None-Match`) when the server's archive list changes. Cached text
  tables are served at once and refreshed in the background, so an overlay
  change shows on the next run.
- **Failures.** A 404 is remembered for the run and the lookup falls through
  to the next candidate. A bad file (over 128 MB, cut short) only fails that
  file. If the server cannot be reached (dial, TLS, no response), cached
  files are used and new requests pause for 30 s; nothing aborts.
- **Limits.** At most 6 connections to the server and 4 background
  refreshes or prefetches at a time; files up to 128 MB, the index up to
  64 MB. Cache entries are keyed by host, so one `cache_dir` can serve
  several servers.
- **Settings.** `[data] remote = auto|on|off`, `remote_url`, `cache_dir`, or
  `--data-remote`, `--data-url`, `--data-cache`. `on` also streams files that
  local archives lack; a streamed exact name then wins over a local
  archive's suffix-fallback match. `clientinfo.xml` and `sclientinfo.xml`
  are never streamed, in any directory: the server list stays local.

Check a map without any local data:

```sh
go run ./cmd/goro-datacheck --data-url https://ro.turoran.com/client/ --stream-map prontera
```

## Assembled data directory (offline, developers)

goro reads the same data roBrowser serves for TuroranRO: the five GRFs in
`DATA.INI` order plus remoteclient2's loose overlay. Loose files beat every
archive in goro (as `OVERLAY_PATH` does in remoteclient2), so the overlay is
laid over the GRFs as plain files.

```sh
scripts/data/assemble-turoran-data.sh GRF_DIR LOOSE_DIR OUT_DIR
goro --data-dir OUT_DIR
```

- `GRF_DIR`: rodata-dev `/opt/remoteclient2/resources-0716` (`DATA.INI`,
  `translation.grf`, `2025-07-16_en.grf`, `data.grf`, `iro_assets.grf`,
  `region_assets.grf`).
- `LOOSE_DIR`: rodata-dev `/opt/remoteclient2` (`overlay-live-0716/`,
  `System/`).

The output is symlinks only. The script also links the tables roBrowser's
`DBManager.js` loads under Turoran names to their canonical client names:

| Canonical name | Turoran source |
|---|---|
| `System/itemInfo.lua` | `System/itemInfo.turoran.renewal11.lua` (29,596 items, English) |
| `lua files/skillinfoz/skillid.lub` | `skillid.turoran4.lub` |
| `lua files/skillinfoz/skillinfolist.lub` | `skillinfolist.turoran13.lub` |
| `lua files/skillinfoz/skilltreeview.lub` | `skilltreeview.fourth.lub` |
| `lua files/skillinfoz/skilldescript.lub` | `System/skilldescript.turoran.re2.lub` |
| `lua files/datainfo/petinfo.lub`, `addrandomoptionnametable.lub` | `*.turoran.lub` |
| `data/cardprefixnametable.txt`, `data/num2cardillustnametable.txt` | `*.turoran.txt` |

When roBrowser bumps one of these names, bump the script and
`res.TuroranRemoteAliases` in the same change (`TestRemoteTuroranAliasesPreferVersionedTables`
fails if they differ).

## Lua tables (.lub)

- **Compiled Lua 5.1 bytecode** (421 files in our GRFs) is translated in
  process into gopher-lua prototypes. All 421 translate; no decompiler.
- **Lua source** (the 77 translated overlay tables, the 23 MB itemInfo) is
  parsed with gopher-lua's parser. Files that are only global table
  assignments (every big table) are evaluated straight from the syntax tree;
  anything with code runs in gopher-lua. gopher-lua's compiler de-duplicates
  constants in quadratic time: the Turoran itemInfo takes 53 s to compile but
  0.5 s through the tree walk.
- Measured on CT119 with an assembled dir: items 0.5 s, skills (ids, info
  list, descriptions) 0.8 s, skill trees 0.4 s, NPC names 8 ms, all lazy.
- **Map names:** `mapnametable.txt` is loaded lazily (`Manager.MapDisplayName`). The minimap footer uses the display name when present, otherwise the map file stem. Renewal maps that the English table does not list (for example `rag_fes`, `bl_grass`) still show the stem.

## Checking a data set

```sh
go run ./cmd/goro-datacheck --data-dir OUT_DIR
GORO_DATA_DIR=OUT_DIR go test ./res/
```
