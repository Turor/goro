#!/usr/bin/env bash
# Controller-only goro release publisher (the goro twin of Korangar's
# korangar-release-publish.timer). No GitHub Actions: it polls Turor/goro main,
# runs the local gate on that commit, cross-compiles Linux and Windows (pure Go,
# no Windows VM needed), signs with the host-local seed, verifies the envelope
# with the client's embedded public key, and publishes to Garage.
#
#   DRY_RUN=1 scripts/release/poll-and-publish.sh      # sign + verify, no upload
#   SHA=<commit> DRY_RUN=1 scripts/release/poll-and-publish.sh
#
# Never prints key material. Never run on ephemeral VMs.
set -euo pipefail

# The checkout below moves this very file; run from a private copy instead.
if [[ -z "${GORO_PUBLISH_COPY:-}" ]]; then
  copy=$(mktemp)
  cp "$0" "$copy"
  GORO_PUBLISH_COPY="$copy" exec bash "$copy" "$@"
fi

REPO_URL="${REPO_URL:-https://github.com/Turor/goro.git}"
BRANCH="${BRANCH:-main}"
WORK="${WORK:-$HOME/Work/goro-release-publisher}"
SIGNING="${SIGNING:-$HOME/Work/korangar-signing}"
SEED="${SEED:-$SIGNING/update.seed}"
STATE="${STATE:-$SIGNING/goro-last-published-sha}"
LOG="${LOG:-$WORK/publish.log}"
DRY_RUN="${DRY_RUN:-0}"
export PATH="${GO_ROOT:-$HOME/Work/toolchains/go}/bin:$PATH"

mkdir -p "$WORK/staging" "$WORK/failures"
chmod 700 "$WORK" "$WORK/staging"
exec 9>"$WORK/publish.lock"
flock -n 9 || { echo "another publish is running" >>"$LOG"; exit 0; }
log() { printf '%s %s\n' "$(date -Iseconds)" "$*" | tee -a "$LOG"; }

[[ -f "$SEED" ]] || { log "missing signing seed"; exit 1; }
sha="${SHA:-$(git ls-remote "$REPO_URL" "refs/heads/$BRANCH" | cut -f1)}"
[[ "$sha" =~ ^[0-9a-f]{40}$ ]] || { log "could not resolve $BRANCH"; exit 1; }
if [[ "$DRY_RUN" != 1 && -f "$STATE" && "$(tr -d ' \n' <"$STATE")" == "$sha" ]]; then
  exit 0
fi
fails=$(cat "$WORK/failures/$sha" 2>/dev/null || echo 0)
if [[ "$DRY_RUN" != 1 ]] && (( fails >= 3 )); then exit 0; fi
on_exit() {
  local rc=$?
  rm -f "$GORO_PUBLISH_COPY"
  if (( rc != 0 )); then
    [[ "$DRY_RUN" == 1 ]] || echo $((fails + 1)) >"$WORK/failures/$sha"
    log "publish failed sha=$sha rc=$rc dry_run=$DRY_RUN"
  fi
}
trap on_exit EXIT

src="$WORK/src"
if [[ ! -d "$src/.git" ]]; then git clone -q "$REPO_URL" "$src"; fi
git -C "$src" fetch -q origin "$sha"
git -C "$src" checkout -q --detach "$sha"
git -C "$src" clean -qfdx

log "gate sha=$sha"
(cd "$src" && scripts/release/local-gate.sh) >>"$LOG" 2>&1

build=$(date +%s)
version="main-${sha:0:12}"
stage="$WORK/staging/$sha"
rm -rf "$stage"
mkdir -p "$stage"
chmod 700 "$stage"
(cd "$src" && scripts/release/build-release.sh --build "$build" --out "$stage") >>"$LOG" 2>&1
(cd "$src" && CGO_ENABLED=0 go build -trimpath -o "$stage/goro-sign-release" ./cmd/goro-sign-release)

linux="goro-linux-amd64-${sha}"
windows="goro-windows-amd64-${sha}.exe"
"$stage/goro-sign-release" sign --private-key "$SEED" --version "$version" --build "$build" \
  --out "$stage/manifest.goroup" --file "$linux=$stage/$linux" --file "$windows=$stage/$windows" >>"$LOG"
# Public-key-only check with the key compiled into the client.
"$stage/goro-sign-release" verify --manifest "$stage/manifest.goroup" >>"$LOG"

publish=("$src/scripts/release/publish-goro-main-channel.sh" --sha "$sha" --manifest "$stage/manifest.goroup" \
  --linux "$stage/$linux" --windows "$stage/$windows")
if [[ "$DRY_RUN" == 1 ]]; then
  "${publish[@]}" --dry-run | tee -a "$LOG"
  log "dry run ok sha=$sha version=$version build=$build stage=$stage"
  exit 0
fi
"${publish[@]}" >>"$LOG"
printf '%s\n' "$sha" >"$STATE"
chmod 600 "$STATE"
log "published sha=$sha version=$version build=$build"
