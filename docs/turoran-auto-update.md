# TuroranRO auto-update (goro)

Release builds of the Turoran goro client check for a signed update on every
launch, before the window opens. The design and the signing key are shared
with the Korangar client (`Turor/korangar`, `korangar-update`); only the
manifest magic, file names and hosting prefix differ.

## Launch behaviour

1. `--update-self-check` and `--build-info` are answered first and exit.
2. The check runs only when the binary was linked with a release channel
   (`updater.Channel=main`). Plain `go build` binaries never update, and there
   is no override. It is also skipped with `--headless` (simulation runs) and
   with `GORO_SKIP_UPDATE=1`.
   The check takes `updates/.lock` (a random token; only the owner removes it;
   a lock older than 15 minutes is taken over by an atomic rename). If another
   instance holds it, the check is skipped. `<name>.old` is only deleted by the
   lock holder.
3. The client fetches `https://ro.turoran.com/client/goro/main/latest`. That
   remoteclient2 endpoint returns JSON with a short-lived presigned
   `manifest_url` and per-file URLs for the private Garage bucket.
   `GORO_UPDATE_MANIFEST_URL` overrides the URL (a catalog or a manifest), but
   only allowlisted URLs are fetched: that catalog, presigned
   `https://garage.turoran.com/korangar-releases/goro/main/<name>` objects,
   loopback `http://`, and `file://`. Redirects are refused.
   The catalog and manifest fetches share one **12 s deadline** (5 s dial, 5 s
   TLS), so a stalled server delays the window by at most that; a timeout is a
   quiet skip. Only the binary download gets the long (10 minute) timeout.
4. The manifest signature is checked against the embedded Turoran release
   public key (`updater.EmbeddedPublicKey`). Nothing is downloaded before that.
5. Rollback protection: the manifest build must be newer than both the build
   linked into the running binary and `updates/applied.build`. Equal means up
   to date.
6. Only the binary for the running OS and architecture is downloaded
   (`goro-<os>-<arch>-<sha>[.exe]`), streamed into
   `updates/staging/<build>/` (named by the numeric build, never the version
   string), and checked against the signed size and SHA-256.
7. The staged binary must pass `--update-self-check`, printing the same build
   as the manifest. Then the running binary is renamed to `<name>.old` (the
   supported way to replace a running `.exe` on Windows), the new one is moved
   into place, and the self-check runs again. Any failure moves the old binary
   back.
8. `updates/applied.build` and `applied.version` are written, the new binary is
   started with the same arguments (`GORO_SKIP_UPDATE=1` so it cannot loop),
   and the old process exits. `<name>.old` is deleted on the next launch.

Network trouble (DNS, refused, timeout, 404, 5xx) skips the check quietly. A
bad signature, hash or self-check is logged as a warning. In every failure
case the installed binary is unchanged and the game still starts.

## Manifest (GORO1)

```text
GORO1SIG
<128 lowercase hex Ed25519 signature>
GORO1
<version>            e.g. main-0123456789ab
<build>              unix time at signing, strictly increasing
<file count>
<path>               repeated per file, sorted by path bytes
<sha256 lowercase hex>
<size decimal>
<url>                relative to the manifest unless overridden
```

The signature covers everything from `GORO1` on. `version` matches
`[0-9A-Za-z._+-]{1,64}`, must not start with `.`, and cannot be `.`/`..`. Korangar's `KRUP1` uses the
same layout and the same key; the different magic makes a manifest for one
client useless to the other.

## Release flow (local, no GitHub Actions)

```
merge to main (local gate + independent review PASS)
  → controller timer runs the PINNED poller (~/Work/goro-release-publisher/pinned/)
      → verify pinned files against PINNED.sha256
      → bare mirror: commit must be on main and descend from the last published commit
      → CT119, user gorobuild (no seed, no sudo, no keys): wipe its home,
        temp files, crontab and processes; fresh run dir, HOME, GOCACHE and
        GOMODCACHE (GOTOOLCHAIN=local, -mod=readonly); then
        scripts/release/local-gate.sh + build-release.sh of that commit;
        kill its processes and delete the run dir afterwards
      → copy back only the two binaries, read as gorobuild (symlinks and
        non-regular files refused); check names, size, ELF/PE magic,
        embedded version string (never executed on the controller)
      → pinned goro-sign-release sign   (host-local seed)
      → pinned goro-sign-release verify (embedded public key only)
      → pinned publish-goro-main-channel.sh
          → Garage korangar-releases/goro/main/  (manifest written last, keep 3 SHAs)
```

**Why pinned.** The controller user owns the signing seed (shared with
Korangar). It never builds, tests or runs code from the commit being released.
The signer and the publisher scripts are pinned from a reviewed commit with:

```sh
scripts/release/install-pinned-publisher.sh <reviewed commit sha>
```

Re-pin only after an independent review of a change to `cmd/goro-sign-release`,
`updater/manifest.go` or `scripts/release/`. The build user on CT119 is created
once with `scripts/release/setup-build-user.sh` (as root there).

`DRY_RUN=1` does everything except the upload and then deletes the signed
output. A non-main commit (`ALLOW_NON_MAIN=1`) is signed with a throwaway key,
never the release seed. `SHA=<commit> ALLOW_NON_MAIN=1`
is accepted only together with `DRY_RUN=1`. The systemd user units in
`packaging/release/` are templates and are not enabled by this change.

The signing seed and the Garage writer key stay on the controller, outside any
git tree. Never commit or print them.

## Key rotation

`updater.EmbeddedPublicKey` is the same key as Korangar's
`EMBEDDED_UPDATE_PUBLIC_KEY`, and `TestEmbeddedKeyIsTheTuroranReleaseKey` pins
it. Rotating the key needs a coordinated change in both clients, a release of
each signed with the old key that carries the new key, and only then a switch
of the controller seed.

## Local gate

```sh
scripts/release/local-gate.sh
```

gofmt, `go vet`, `go test`, Linux and Windows builds with `CGO_ENABLED=0
-tags nofakecgo`, and the signer build.
