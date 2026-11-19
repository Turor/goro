#!/usr/bin/env bash
# Sign a client-update or data manifest that carries a next-key list, using a
# seed that lives only at a local path. The seed is never printed.
#
# Usage:
#   scripts/release/rotate-signing-key.sh \
#     --private-key PATH \
#     --new-public-key HEX \
#     --out PATH \
#     --kind update|data \
#     [--version V] [--build N] \
#     [--file REL=SOURCE]... [--root DIR] [--url REL=URL]...
#
# The current public key (derived from --private-key) and --new-public-key are
# both written into the signed next-key list so the client already trusts the
# new key before the old one retires.
set -euo pipefail

usage() {
  sed -n '2,16p' "$0" | sed 's/^# \?//'
  exit 2
}

key_path=""
new_pub=""
out_path=""
kind=""
version=""
build=""
files=()
urls=()
root=""

while [[ $# -gt 0 ]]; do
  case "$1" in
    --private-key) key_path="${2:-}"; shift 2 ;;
    --new-public-key) new_pub="${2:-}"; shift 2 ;;
    --out) out_path="${2:-}"; shift 2 ;;
    --kind) kind="${2:-}"; shift 2 ;;
    --version) version="${2:-}"; shift 2 ;;
    --build) build="${2:-}"; shift 2 ;;
    --file) files+=("${2:-}"); shift 2 ;;
    --url) urls+=("${2:-}"); shift 2 ;;
    --root) root="${2:-}"; shift 2 ;;
    -h|--help) usage ;;
    *) echo "rotate-signing-key: unknown argument $1" >&2; usage ;;
  esac
done

if [[ -z "$key_path" || -z "$new_pub" || -z "$out_path" || -z "$kind" ]]; then
  echo "rotate-signing-key: --private-key, --new-public-key, --out and --kind are required" >&2
  exit 2
fi
if [[ ! -f "$key_path" ]]; then
  echo "rotate-signing-key: private key path is not a file" >&2
  exit 1
fi
new_pub="$(printf '%s' "$new_pub" | tr 'A-F' 'a-f')"
if [[ ! "$new_pub" =~ ^[0-9a-f]{64}$ ]]; then
  echo "rotate-signing-key: --new-public-key must be 64 lowercase hex characters" >&2
  exit 1
fi

root_dir="$(cd "$(dirname "$0")/../.." && pwd)"
signer=(go run ./cmd/goro-sign-release)
if [[ -n "${GORO_SIGN_RELEASE:-}" ]]; then
  signer=("$GORO_SIGN_RELEASE")
fi

# Derive the current public key; goro-sign-release never prints the seed.
old_pub="$("${signer[@]}" public-key --private-key "$key_path" | awk 'NR==1{print $1}')"
old_pub="$(printf '%s' "$old_pub" | tr 'A-F' 'a-f')"
if [[ ! "$old_pub" =~ ^[0-9a-f]{64}$ ]]; then
  echo "rotate-signing-key: could not derive the current public key" >&2
  exit 1
fi

cd "$root_dir"
args=()
case "$kind" in
  update)
    if [[ -z "$version" || -z "$build" || ${#files[@]} -eq 0 ]]; then
      echo "rotate-signing-key: --kind update needs --version, --build and --file" >&2
      exit 2
    fi
    args=(sign --private-key "$key_path" --version "$version" --build "$build" --out "$out_path")
    for f in "${files[@]}"; do args+=(--file "$f"); done
    for u in "${urls[@]}"; do args+=(--url "$u"); done
    ;;
  data)
    if [[ -z "$version" ]]; then
      echo "rotate-signing-key: --kind data needs --version" >&2
      exit 2
    fi
    if [[ ${#files[@]} -eq 0 && -z "$root" ]]; then
      echo "rotate-signing-key: --kind data needs --file or --root" >&2
      exit 2
    fi
    args=(sign-data --private-key "$key_path" --version "$version" --out "$out_path")
    for f in "${files[@]}"; do args+=(--file "$f"); done
    if [[ -n "$root" ]]; then
      args+=(--root "$root")
    fi
    ;;
  *)
    echo "rotate-signing-key: --kind must be update or data" >&2
    exit 2
    ;;
esac
args+=(--next-key "$old_pub" --next-key "$new_pub")
"${signer[@]}" "${args[@]}"
echo "rotate-signing-key: wrote $out_path (next-keys: current + new; seed not printed)"
