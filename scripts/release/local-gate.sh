#!/usr/bin/env bash
# Local merge gate for Turor/goro. We do not use GitHub Actions: a PR merges
# after this gate passes and an independent review says PASS.
#
#   scripts/release/local-gate.sh
#
# Steps: gofmt, go vet, go test, Linux and Windows release builds (pure Go,
# CGO off), and the release signer build.
set -euo pipefail
cd "$(git rev-parse --show-toplevel)"
export CGO_ENABLED=0
TAGS="${GORO_TAGS:-nofakecgo}"
out=$(mktemp -d)
trap 'rm -rf "$out"' EXIT
start=$SECONDS

step() {
  local s=$SECONDS
  echo "== $*"
  "$@"
  echo "   ok ($((SECONDS - s))s)"
}

echo "== gofmt"
unformatted=$(git ls-files -z '*.go' | xargs -0 gofmt -l)
if [[ -n "$unformatted" ]]; then
  echo "gofmt needed:" >&2
  echo "$unformatted" >&2
  exit 1
fi
step go vet -tags "$TAGS" ./...
step go test -tags "$TAGS" -count=1 ./...
step env GOOS=linux GOARCH=amd64 go build -tags "$TAGS" -trimpath -o "$out/goro" .
step env GOOS=windows GOARCH=amd64 go build -tags "$TAGS" -trimpath -o "$out/goro.exe" .
step go build -trimpath -o "$out/goro-sign-release" ./cmd/goro-sign-release
step go build -trimpath -o "$out/goro-replay" ./cmd/goro-replay
echo "LOCAL GATE PASS $(git rev-parse --short=12 HEAD) in $((SECONDS - start))s"
