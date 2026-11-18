#!/usr/bin/env bash
# Controller-only goro release publisher. Run the PINNED copy installed by
# install-pinned-publisher.sh, never a checkout of main.
#
# Trust split:
#   * The controller user owns the signing seed. It runs only pinned, reviewed
#     code: this script, the publish script and goro-sign-release. It never
#     executes, builds or tests code from the commit being released; it only
#     hashes and signs the produced artifacts.
#   * The gate and the release build run on CT119 as the unprivileged user
#     `gorobuild`, which has no access to the seed or to the controller.
#   * Only a commit on main that descends from the last published commit is
#     published, so an older commit can never be re-shipped as an "upgrade".
#
#   DRY_RUN=1 .../pinned/poll-and-publish.sh          # everything except upload
#   SHA=<commit> DRY_RUN=1 ALLOW_NON_MAIN=1 ...       # review a branch commit
#
# Never prints key material.
set -euo pipefail

SELF_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_URL="${REPO_URL:-https://github.com/Turor/goro.git}"
BRANCH="${BRANCH:-main}"
WORK="${WORK:-$HOME/Work/goro-release-publisher}"
SIGNING="${SIGNING:-$HOME/Work/korangar-signing}"
SEED="${SEED:-$SIGNING/update.seed}"
STATE="${STATE:-$SIGNING/goro-last-published-sha}"
LOG="${LOG:-$WORK/publish.log}"
BUILD_HOST="${BUILD_HOST:-root@192.168.11.54}"
BUILD_USER="${BUILD_USER:-gorobuild}"
DRY_RUN="${DRY_RUN:-0}"
ALLOW_NON_MAIN="${ALLOW_NON_MAIN:-0}"
MAX_BYTES=$((512 * 1024 * 1024))

mkdir -p "$WORK/staging" "$WORK/failures"
chmod 700 "$WORK" "$WORK/staging" "$WORK/failures"
touch "$LOG"
chmod 600 "$LOG"
exec 9>"$WORK/publish.lock"
flock -n 9 || { echo "another publish is running" >>"$LOG"; exit 0; }
# Remote build output is untrusted: drop control characters (escape sequences).
sanitize() { tr -d '\000-\010\013-\037\177'; }
log() { printf '%s %s\n' "$(date -Iseconds)" "$*" | tee -a "$LOG"; }
die() { log "refused: $*"; exit 1; }

# 1. Only pinned, unmodified tools.
[[ -f "$SELF_DIR/PINNED.sha256" ]] || die "not running from a pinned install (see install-pinned-publisher.sh)"
(cd "$SELF_DIR" && sha256sum --quiet -c PINNED.sha256) || die "pinned publisher files changed"
SIGNER="$SELF_DIR/goro-sign-release"
PUBLISH="$SELF_DIR/publish-goro-main-channel.sh"
[[ -f "$SEED" ]] || die "missing signing seed"
if [[ "$ALLOW_NON_MAIN" == 1 && "$DRY_RUN" != 1 ]]; then die "ALLOW_NON_MAIN needs DRY_RUN=1"; fi

# 2. Resolve the commit and check ancestry in a controller-side bare mirror
#    (git data only; nothing from the tree runs here).
mirror="$WORK/mirror.git"
[[ -d "$mirror" ]] || git init -q --bare "$mirror"
git -C "$mirror" fetch -q --prune "$REPO_URL" "+refs/heads/$BRANCH:refs/heads/$BRANCH"
sha="${SHA:-$(git -C "$mirror" rev-parse "refs/heads/$BRANCH")}"
[[ "$sha" =~ ^[0-9a-f]{40}$ ]] || die "bad commit $sha"
git -C "$mirror" cat-file -e "$sha^{commit}" 2>/dev/null || git -C "$mirror" fetch -q "$REPO_URL" "$sha"
if [[ "$ALLOW_NON_MAIN" != 1 ]]; then
  git -C "$mirror" merge-base --is-ancestor "$sha" "refs/heads/$BRANCH" || die "$sha is not on $BRANCH"
fi
last=""
[[ -f "$STATE" ]] && last="$(tr -d ' \n' <"$STATE")"
if [[ -n "$last" ]]; then
  if [[ "$last" == "$sha" && "$DRY_RUN" != 1 ]]; then exit 0; fi
  git -C "$mirror" cat-file -e "$last^{commit}" 2>/dev/null || die "last published $last is missing from the repo history"
  git -C "$mirror" merge-base --is-ancestor "$last" "$sha" || die "$sha does not descend from last published $last"
fi

fails=$(cat "$WORK/failures/$sha" 2>/dev/null || echo 0)
if [[ "$DRY_RUN" != 1 ]] && (( fails >= 3 )); then exit 0; fi
stage=""
remote_base=""
on_exit() {
  local rc=$?
  # Never leave a build user process or run directory behind on CT119.
  if [[ -n "$remote_base" ]]; then
    remote_root "pkill -KILL -u $BUILD_USER || true; rm -rf -- '$remote_base'" || true
  fi
  # A dry run never keeps signed output.
  if [[ "$DRY_RUN" == 1 && -n "$stage" && -d "$stage" ]]; then
    chmod -R u+w "$stage" 2>/dev/null || true
    rm -rf -- "$stage"
  fi
  if (( rc != 0 )); then
    [[ "$DRY_RUN" == 1 ]] || echo $((fails + 1)) >"$WORK/failures/$sha"
    log "publish failed sha=$sha rc=$rc dry_run=$DRY_RUN"
  fi
}
trap on_exit EXIT

# 3. Gate and build on CT119 as the unprivileged build user, in a fresh
#    directory with a fresh HOME, Go cache and module cache every run. Before
#    the run the build user's home, temp files, crontab and processes are wiped,
#    so nothing an earlier commit planted (git hooks or config, cache entries,
#    daemons) can touch this build. The build number is chosen here and linked
#    in; the client self-check enforces it.
remote_root() { ssh -o BatchMode=yes "$BUILD_HOST" -- "$1"; }
build=$(date +%s)
version="main-${sha:0:12}"
linux="goro-linux-amd64-${sha}"
windows="goro-windows-amd64-${sha}.exe"
remote_root "set -e; pkill -KILL -u $BUILD_USER || true; crontab -r -u $BUILD_USER 2>/dev/null || true
  find /home/$BUILD_USER -mindepth 1 -maxdepth 1 -exec rm -rf -- {} +
  find /tmp /var/tmp /dev/shm -xdev -user $BUILD_USER -delete 2>/dev/null || true"
# Only a validated path ever reaches remote_base (and the EXIT trap's rm -rf).
new_base=$(remote_root "runuser -u $BUILD_USER -- mktemp -d /home/$BUILD_USER/run.XXXXXXXX")
[[ "$new_base" =~ ^/home/$BUILD_USER/run\.[A-Za-z0-9]{8}$ ]] || die "bad remote run dir"
remote_base="$new_base"
log "remote gate+build sha=$sha build=$build on $BUILD_HOST as $BUILD_USER in $remote_base"
# shellcheck disable=SC2087
ssh -o BatchMode=yes "$BUILD_HOST" -- runuser -u "$BUILD_USER" -- env -i \
  HOME="$remote_base/home" GOCACHE="$remote_base/gocache" GOMODCACHE="$remote_base/gomod" \
  GOPATH="$remote_base/gopath" GOTOOLCHAIN=local GOFLAGS=-mod=readonly \
  GIT_CONFIG_NOSYSTEM=1 PATH=/usr/local/go/bin:/usr/bin:/bin LANG=C.UTF-8 \
  bash -s -- "$REPO_URL" "$sha" "$build" "$remote_base" 2>&1 <<'REMOTE' | sanitize >>"$LOG"
set -euo pipefail
repo="$1" sha="$2" build="$3" base="$4"
mkdir -p "$HOME"
git clone -q --no-checkout "$repo" "$base/src"
cd "$base/src"
git fetch -q origin "$sha"
git -c advice.detachedHead=false checkout -q "$sha"
scripts/release/local-gate.sh
scripts/release/build-release.sh --build "$build" --out "$base/out"
REMOTE
[[ "${PIPESTATUS[0]}" == 0 ]] || die "remote gate or build failed"

# 4. Copy back only the two artifacts, read as the build user (so a symlink
#    can never expose a root-only file), refusing anything that is not a
#    regular file; sanity-check them without running them.
stage="$WORK/staging/$sha"
rm -rf "$stage"
mkdir -p "$stage"
chmod 700 "$stage"
# Kill every build-user process first, so nothing left running by the build
# can swap an artifact between the checks and the read, and cap the read at
# MAX_BYTES+1 so even a swapped-in endless file cannot fill this disk (the
# size check below then rejects it).
remote_root "pkill -KILL -u $BUILD_USER || true"
read_cap=$((MAX_BYTES + 1))
for f in "$linux" "$windows"; do
  remote_root "runuser -u $BUILD_USER -- sh -c 'test -f \"\$1\" && test ! -L \"\$1\" && head -c $read_cap -- \"\$1\"' sh '$remote_base/out/$f'" >"$stage/$f" \
    || die "artifact $f missing, not a regular file, or a symlink"
done
for f in "$linux" "$windows"; do
  [[ -f "$stage/$f" && ! -L "$stage/$f" ]] || die "missing artifact $f"
  size=$(stat -c %s "$stage/$f")
  (( size > 0 && size < MAX_BYTES )) || die "bad size for $f"
  grep -aqF "$version" "$stage/$f" || die "$f does not carry version $version"
  chmod 0444 "$stage/$f"
done
[[ "$(head -c 4 "$stage/$linux" | od -An -tx1 | tr -d ' ')" == "7f454c46" ]] || die "linux artifact is not ELF"
[[ "$(head -c 2 "$stage/$windows")" == "MZ" ]] || die "windows artifact is not PE"
remote_root "pkill -KILL -u $BUILD_USER || true; rm -rf -- '$remote_base'"
remote_base=""

# 5. Sign with the pinned signer; verify with the client's embedded key only.
#    A commit that is not on main (ALLOW_NON_MAIN, dry runs only) is signed
#    with a throwaway key, so it never gets a real signature; every other dry
#    run signs with the real seed to prove it, and the trap deletes the output.
if [[ "$ALLOW_NON_MAIN" == 1 ]]; then
  throwaway="$stage/throwaway.seed"
  (umask 077 && head -c 32 /dev/urandom >"$throwaway")
  pub=$("$SIGNER" public-key --private-key "$throwaway" | head -n1)
  "$SIGNER" sign --private-key "$throwaway" --version "$version" --build "$build" \
    --out "$stage/manifest.goroup" --file "$linux=$stage/$linux" --file "$windows=$stage/$windows" >>"$LOG"
  rm -f "$throwaway"
  "$SIGNER" verify --manifest "$stage/manifest.goroup" --public-key "$pub" >>"$LOG"
  log "non-main dry run: signed with a throwaway key, not the release seed"
else
  "$SIGNER" sign --private-key "$SEED" --version "$version" --build "$build" \
    --out "$stage/manifest.goroup" --file "$linux=$stage/$linux" --file "$windows=$stage/$windows" >>"$LOG"
  "$SIGNER" verify --manifest "$stage/manifest.goroup" >>"$LOG"
fi

publish=("$PUBLISH" --sha "$sha" --manifest "$stage/manifest.goroup" --linux "$stage/$linux" --windows "$stage/$windows")
if [[ "$DRY_RUN" == 1 ]]; then
  "${publish[@]}" --dry-run | tee -a "$LOG"
  log "dry run ok sha=$sha version=$version build=$build (signed output deleted)"
  exit 0
fi
"${publish[@]}" >>"$LOG"
printf '%s\n' "$sha" >"$STATE"
chmod 600 "$STATE"
log "published sha=$sha version=$version build=$build"
