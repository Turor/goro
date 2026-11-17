#!/usr/bin/env bash
# Build signed-channel release binaries for one commit. Run from a clean
# checkout of that commit. Prints nothing secret; signing is a separate step.
#
#   scripts/release/build-release.sh --build UNIXTIME --out DIR [--channel main]
#
# Produces DIR/goro-linux-amd64-<sha> and DIR/goro-windows-amd64-<sha>.exe with
# the release identity linked in (updater.Channel/Version/Build).
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

sha=$(git rev-parse HEAD)
version="${CHANNEL}-${sha:0:12}"
pkg="github.com/kivutar/goro/updater"
ldflags="-s -w -X ${pkg}.Channel=${CHANNEL} -X ${pkg}.Version=${version} -X ${pkg}.Build=${BUILD}"
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
echo "built ${version} build=${BUILD} in ${OUT}"
