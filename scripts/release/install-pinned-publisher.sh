#!/usr/bin/env bash
# Operator step on the controller: pin the release signer and publisher scripts
# from a commit that has passed independent review. The timer only ever runs
# this pinned copy, so code merged to main later can never run as the key owner
# or near the signing seed. Re-run only after reviewing a change to
# cmd/goro-sign-release, updater/manifest.go or scripts/release/.
#
#   scripts/release/install-pinned-publisher.sh <reviewed-commit-sha>
#
# Nothing from the tree is executed here: `go build` compiles the signer (Go
# has no build scripts and CGO is off), and the shell scripts are copied, not run.
set -euo pipefail

SHA="${1:-}"
[[ "$SHA" =~ ^[0-9a-f]{40}$ ]] || { echo "usage: $0 <reviewed 40-char commit sha>" >&2; exit 1; }
REPO_URL="${REPO_URL:-https://github.com/Turor/goro.git}"
PIN="${PIN:-$HOME/Work/goro-release-publisher/pinned}"
export PATH="${GO_ROOT:-$HOME/Work/toolchains/go}/bin:$PATH" CGO_ENABLED=0
# Use the installed toolchain only and never rewrite go.mod/go.sum.
export GOTOOLCHAIN=local GOFLAGS=-mod=readonly

tmp=$(mktemp -d)
trap 'chmod -R u+w "$tmp" 2>/dev/null; rm -rf "$tmp"' EXIT
git clone -q --no-checkout "$REPO_URL" "$tmp/src"
git -C "$tmp/src" fetch -q origin "$SHA"
git -C "$tmp/src" -c advice.detachedHead=false checkout -q "$SHA"
export GOCACHE="$tmp/gocache" GOMODCACHE="$tmp/gomod"

new="$PIN.new"
rm -rf "$new"
mkdir -p "$new"
(cd "$tmp/src" && go build -trimpath -o "$new/goro-sign-release" ./cmd/goro-sign-release)
cp "$tmp/src/scripts/release/poll-and-publish.sh" "$tmp/src/scripts/release/publish-goro-main-channel.sh" "$new/"
chmod 0555 "$new"/*
(cd "$new" && sha256sum goro-sign-release poll-and-publish.sh publish-goro-main-channel.sh >PINNED.sha256)
printf 'commit=%s\npinned_at=%s\n' "$SHA" "$(date -Iseconds)" >"$new/PINNED"
chmod 0444 "$new/PINNED" "$new/PINNED.sha256"
chmod 0555 "$new"
if [[ -d "$PIN" ]]; then chmod u+w "$PIN"; rm -rf "$PIN"; fi
mv "$new" "$PIN"
echo "pinned goro publisher from $SHA in $PIN"
cat "$PIN/PINNED.sha256"
