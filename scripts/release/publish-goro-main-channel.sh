#!/usr/bin/env bash
# Upload a signed goro main-channel release to Garage. Controller only.
# Never prints key material. Mirrors Korangar's publish-korangar-main-channel.sh.
#
#   scripts/release/publish-goro-main-channel.sh --sha SHA --manifest PATH \
#       --linux PATH [--windows PATH] [--dry-run]
#
# Objects (bucket korangar-releases, prefix goro/main/):
#   goro/main/manifest.goroup                (written last, via a temp key)
#   goro/main/goro-linux-amd64-<sha>
#   goro/main/goro-windows-amd64-<sha>.exe
# The last 3 SHAs are kept. The bucket's existing write-only key is reused;
# remoteclient2 serves the update catalog (updater.DefaultCatalogURL) from it.
set -euo pipefail

SHA="" MANIFEST="" LINUX="" WINDOWS="" DRY_RUN=0
while [[ $# -gt 0 ]]; do
  case "$1" in
    --sha) SHA="$2"; shift 2 ;;
    --manifest) MANIFEST="$2"; shift 2 ;;
    --linux) LINUX="$2"; shift 2 ;;
    --windows) WINDOWS="$2"; shift 2 ;;
    --dry-run) DRY_RUN=1; shift ;;
    *) echo "unknown arg: $1" >&2; exit 1 ;;
  esac
done
[[ "$SHA" =~ ^[0-9a-f]{40}$ && -f "$MANIFEST" && -f "$LINUX" && ( -z "$WINDOWS" || -f "$WINDOWS" ) ]] || {
  echo "usage: $0 --sha SHA --manifest PATH --linux PATH [--windows PATH] [--dry-run]" >&2
  exit 1
}

BUCKET="${BUCKET:-korangar-releases}"
PREFIX="goro/main"
LINUX_KEY="${PREFIX}/goro-linux-amd64-${SHA}"
WIN_KEY="${PREFIX}/goro-windows-amd64-${SHA}.exe"
MANIFEST_KEY="${PREFIX}/manifest.goroup"
STAGING_KEY="${PREFIX}/.manifest.goroup.${SHA}.tmp"

if [[ "$DRY_RUN" == 1 ]]; then
  echo "dry run: would upload to s3://${BUCKET}/"
  echo "  ${LINUX_KEY} ($(stat -c %s "$LINUX") bytes)"
  [[ -n "$WINDOWS" ]] && echo "  ${WIN_KEY} ($(stat -c %s "$WINDOWS") bytes)"
  echo "  ${STAGING_KEY} -> ${MANIFEST_KEY} ($(stat -c %s "$MANIFEST") bytes)"
  echo "  then prune ${PREFIX}/ to the newest 3 SHAs"
  exit 0
fi

CRED="${GORO_GARAGE_ENV:-$HOME/Work/korangar-signing/garage-writer.env}"
[[ -f "$CRED" ]] || { echo "missing garage writer env" >&2; exit 1; }
# Hosts are operator config, never committed (see publish.env.example).
CONF="${GORO_PUBLISH_CONFIG:-$HOME/Work/goro-release-publisher/publish.env}"
pconf() { [[ -f "$CONF" ]] || return 0; sed -n "s/^[[:space:]]*$1[[:space:]]*=[[:space:]]*//p" "$CONF" | tail -n1 | tr -d '\r"' | sed 's/[[:space:]]*$//'; }
GARAGE_HOST="${GARAGE_HOST:-$(pconf garage_host)}"
SSH_EXTRA="${SSH_EXTRA-$(pconf ssh_extra)}"
[[ -n "$GARAGE_HOST" ]] || { echo "set garage_host in $CONF" >&2; exit 1; }
ENDPOINT="${ENDPOINT:-http://127.0.0.1:3900}"
REMOTE=/root/goro-release-upload

WORKDIR=$(mktemp -d)
trap 'rm -rf "$WORKDIR"' EXIT
chmod 700 "$WORKDIR"
cp "$MANIFEST" "$WORKDIR/manifest.goroup"
cp "$LINUX" "$WORKDIR/linux.bin"
if [[ -n "$WINDOWS" ]]; then cp "$WINDOWS" "$WORKDIR/windows.exe"; else : >"$WORKDIR/windows.exe"; fi
HAS_WINDOWS=$([[ -n "$WINDOWS" ]] && echo 1 || echo 0)
cp "$CRED" "$WORKDIR/writer.env"
chmod 600 "$WORKDIR/writer.env"

# shellcheck disable=SC2086
ssh -o BatchMode=yes $SSH_EXTRA "$GARAGE_HOST" -- "mkdir -p $REMOTE && chmod 700 $REMOTE"
# shellcheck disable=SC2086
scp -q -o BatchMode=yes $SSH_EXTRA "$WORKDIR/writer.env" "$WORKDIR/manifest.goroup" \
  "$WORKDIR/linux.bin" "$WORKDIR/windows.exe" "${GARAGE_HOST}:${REMOTE}/"

# shellcheck disable=SC2086
ssh -o BatchMode=yes $SSH_EXTRA "$GARAGE_HOST" -- bash -s <<EOS
set -euo pipefail
cd $REMOTE
trap 'rm -f $REMOTE/writer.env $REMOTE/linux.bin $REMOTE/windows.exe $REMOTE/manifest.goroup' EXIT
set -a; source ./writer.env; set +a
export AWS_DEFAULT_REGION="\${AWS_DEFAULT_REGION:-garage}" AWS_REQUEST_CHECKSUM_CALCULATION=when_required AWS_EC2_METADATA_DISABLED=true
s3() { aws --endpoint-url ${ENDPOINT} s3api "\$@"; }
s3 put-object --bucket ${BUCKET} --key "${LINUX_KEY}" --body linux.bin >/dev/null
if [[ ${HAS_WINDOWS} == 1 ]]; then s3 put-object --bucket ${BUCKET} --key "${WIN_KEY}" --body windows.exe >/dev/null; fi
s3 put-object --bucket ${BUCKET} --key "${STAGING_KEY}" --body manifest.goroup >/dev/null
s3 copy-object --bucket ${BUCKET} --copy-source "${BUCKET}/${STAGING_KEY}" --key "${MANIFEST_KEY}" >/dev/null
s3 delete-object --bucket ${BUCKET} --key "${STAGING_KEY}" >/dev/null

# Keep the newest 3 SHAs (by LastModified of the Linux binary).
mapfile -t keep < <(s3 list-objects-v2 --bucket ${BUCKET} --prefix "${PREFIX}/goro-linux-amd64-" \
  --query 'reverse(sort_by(Contents,&LastModified))[].Key' --output text | tr '\t' '\n' | sed '/^\$/d;/^None\$/d' | sed 's/.*-//' | head -n 3)
mapfile -t all < <(s3 list-objects-v2 --bucket ${BUCKET} --prefix "${PREFIX}/goro-" --query 'Contents[].Key' --output text | tr '\t' '\n' | sed '/^\$/d;/^None\$/d')
for k in "\${all[@]}"; do
  s=\${k##*-}; s=\${s%.exe}
  drop=1
  for x in "\${keep[@]}"; do [[ "\$x" == "\$s" ]] && drop=0; done
  if [[ \$drop == 1 ]]; then s3 delete-object --bucket ${BUCKET} --key "\$k" >/dev/null; fi
done
echo "published ${MANIFEST_KEY} sha=${SHA}"
EOS
