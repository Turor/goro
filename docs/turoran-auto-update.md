# TuroranRO auto-update (goro)

Release builds of the Turoran goro client check for a signed update on every
launch, before the window opens. The design and the signing key are shared
with the Korangar client (`Turor/korangar`, `korangar-update`); only the
manifest magic, file names and hosting prefix differ.

## Launch behaviour

1. `--update-self-check` and `--build-info` are answered first and exit.
2. The check runs only when the binary was linked with a release channel
   (`updater.Channel=main`). Plain `go build` binaries never update. It is also
   skipped with `--headless` (simulation runs) and with `GORO_SKIP_UPDATE=1`.
3. The client fetches `https://ro.turoran.com/client/goro/main/latest`. That
   remoteclient2 endpoint returns JSON with a short-lived presigned
   `manifest_url` and per-file URLs for the private Garage bucket.
   `GORO_UPDATE_MANIFEST_URL` overrides the URL (a catalog or a manifest), but
   only allowlisted URLs are fetched: that catalog, presigned
   `https://garage.turoran.com/korangar-releases/goro/main/<name>` objects,
   loopback `http://`, and `file://`. Redirects are refused.
4. The manifest signature is checked against the embedded Turoran release
   public key (`updater.EmbeddedPublicKey`). Nothing is downloaded before that.
5. Rollback protection: the manifest build must be newer than both the build
   linked into the running binary and `updates/applied.build`. Equal means up
   to date.
6. Only the binary for the running OS and architecture is downloaded
   (`goro-<os>-<arch>-<sha>[.exe]`), streamed into
   `updates/staging/<version>/`, and checked against the signed size and
   SHA-256.
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

The signature covers everything from `GORO1` on. Korangar's `KRUP1` uses the
same layout and the same key; the different magic makes a manifest for one
client useless to the other.

## Release flow (local, no GitHub Actions)

```
merge to main (local gate + independent review PASS)
  → controller poller: scripts/release/poll-and-publish.sh
      → scripts/release/local-gate.sh on that commit
      → scripts/release/build-release.sh  (Linux + Windows, pure Go cross-compile)
      → goro-sign-release sign   (seed: ~/Work/korangar-signing/update.seed)
      → goro-sign-release verify (embedded public key only)
      → scripts/release/publish-goro-main-channel.sh
          → Garage korangar-releases/goro/main/  (manifest written last, keep 3 SHAs)
```

`DRY_RUN=1` signs and verifies but does not upload. The systemd user units in
`packaging/release/` are templates and are not enabled by this change.

The signing seed and the Garage writer key stay on the controller, outside any
git tree. Never commit or print them.

## Local gate

```sh
scripts/release/local-gate.sh
```

gofmt, `go vet`, `go test`, Linux and Windows builds with `CGO_ENABLED=0
-tags nofakecgo`, and the signer build.
