#!/usr/bin/env bash
# Build signed-channel release binaries for one commit. Run from a clean
# checkout of that commit. Prints nothing secret; signing is a separate step.
#
#   scripts/release/build-release.sh --build UNIXTIME --out DIR [--channel main]
#
# Produces DIR/goro-linux-amd64-<sha> and DIR/goro-windows-amd64-<sha>.exe with
# the release identity linked in (updater.Channel/Version/Build) and the
# server URLs, which are never committed, linked in from the build host's
# release config (GORO_RELEASE_CONFIG, default /etc/goro-release/release.env;
# see scripts/release/release.env.example):
#   update_catalog_url  -> updater.DefaultCatalogURL   (required)
#   release_url_prefix  -> updater.ReleaseURLPrefix    (required)
#   data_remote_url     -> config.DefaultDataRemoteURL (optional)
# A release without the update URLs could never update again, so the build
# refuses to run without them.
set -euo pipefail
cd "$(git rev-parse --show-toplevel)"

BUILD=""
OUT=""
CHANNEL="main"
while [[ $# -gt 0 ]]; do
  case "$1" in
    --build) BUILD="$2"; shift 2 ;;
    --out) OUT="$2"; shift 2 ;;
    --channel) CHANNEL="$2"; shift 2 ;;
    *) echo "unknown arg: $1" >&2; exit 1 ;;
  esac
done
[[ "$BUILD" =~ ^[1-9][0-9]*$ && -n "$OUT" ]] || { echo "usage: $0 --build N --out DIR [--channel main]" >&2; exit 1; }
if [[ -n "$(git status --porcelain --untracked-files=no)" ]]; then
  echo "refusing to build a release from a dirty tree" >&2
  exit 1
fi

CONF="${GORO_RELEASE_CONFIG:-/etc/goro-release/release.env}"
[[ -f "$CONF" ]] || { echo "missing release config $CONF (see scripts/release/release.env.example)" >&2; exit 1; }
conf() { sed -n "s/^[[:space:]]*$1[[:space:]]*=[[:space:]]*//p" "$CONF" | tail -n1 | tr -d '\r"' | sed 's/[[:space:]]*$//'; }
https_url='^https://[A-Za-z0-9.-]+(:443)?/[A-Za-z0-9._~/%-]*$'
catalog_url=$(conf update_catalog_url)
release_prefix=$(conf release_url_prefix)
data_url=$(conf data_remote_url)
for v in "$catalog_url" "$release_prefix"; do
  [[ "$v" =~ $https_url ]] || { echo "release config: update_catalog_url and release_url_prefix must be plain https URLs" >&2; exit 1; }
done
[[ -z "$data_url" || "$data_url" =~ $https_url ]] || { echo "release config: data_remote_url must be a plain https URL" >&2; exit 1; }

sha=$(git rev-parse HEAD)
version="${CHANNEL}-${sha:0:12}"
pkg="github.com/kivutar/goro/updater"
ldflags="-s -w -X ${pkg}.Channel=${CHANNEL} -X ${pkg}.Version=${version} -X ${pkg}.Build=${BUILD}"
ldflags+=" -X ${pkg}.DefaultCatalogURL=${catalog_url} -X ${pkg}.ReleaseURLPrefix=${release_prefix}"
if [[ -n "$data_url" ]]; then
  ldflags+=" -X github.com/kivutar/goro/config.DefaultDataRemoteURL=${data_url}"
fi
mkdir -p "$OUT"
export CGO_ENABLED=0
GOOS=linux GOARCH=amd64 go build -tags nofakecgo -trimpath -ldflags "$ldflags" -o "$OUT/goro-linux-amd64-${sha}" .
GOOS=windows GOARCH=amd64 go build -tags nofakecgo -trimpath -ldflags "$ldflags" -o "$OUT/goro-windows-amd64-${sha}.exe" .

# The Linux binary must answer the self-check with the build we are about to sign.
check=$("$OUT/goro-linux-amd64-${sha}" --update-self-check)
[[ "$check" == "goro-update-self-check-ok build=${BUILD} version=${version}" ]] || {
  echo "self-check mismatch: $check" >&2
  exit 1
}
# Both binaries must carry the update URLs, or existing installs would stop
# updating after this release.
for f in "$OUT/goro-linux-amd64-${sha}" "$OUT/goro-windows-amd64-${sha}.exe"; do
  grep -aqF "$catalog_url" "$f" && grep -aqF "$release_prefix" "$f" || {
    echo "$(basename "$f") does not carry the update URLs" >&2
    exit 1
  }
done
echo "built ${version} build=${BUILD} in ${OUT}"
