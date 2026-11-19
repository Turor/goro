#!/usr/bin/env bash
# Opt-in CT119 budgets for Phase 5.3. Not part of the default local gate.
# Override the seconds with CLEAN_BUILD_BUDGET_S, TEST_SUITE_BUDGET_S, and
# HEADLESS_LOGIN_BUDGET_S. Uses a private GOCACHE so a clean-build measurement
# does not wipe other worktrees' caches.
set -euo pipefail
cd "$(git rev-parse --show-toplevel)"
export PATH="/usr/local/go/bin:${PATH}"
export CGO_ENABLED=0
TAGS="${GORO_TAGS:-nofakecgo}"
CLEAN_BUILD_BUDGET_S="${CLEAN_BUILD_BUDGET_S:-15}"
TEST_SUITE_BUDGET_S="${TEST_SUITE_BUDGET_S:-15}"
HEADLESS_LOGIN_BUDGET_S="${HEADLESS_LOGIN_BUDGET_S:-3}"

fail_over() {
  local name="$1" elapsed="$2" budget="$3"
  echo "BUDGET FAIL $name ${elapsed}s > ${budget}s" >&2
  exit 1
}

time_step() {
  local name="$1" budget="$2"
  shift 2
  local start=$SECONDS
  echo "== $name (budget ${budget}s)"
  "$@"
  local elapsed=$((SECONDS - start))
  echo "   ${elapsed}s"
  if (( elapsed > budget )); then
    fail_over "$name" "$elapsed" "$budget"
  fi
}

echo "== budget-gate CLEAN=${CLEAN_BUILD_BUDGET_S}s TEST=${TEST_SUITE_BUDGET_S}s LOGIN=${HEADLESS_LOGIN_BUDGET_S}s"
cache=$(mktemp -d)
out=$(mktemp -d)
trap 'rm -rf "$cache" "$out"' EXIT

time_step "clean build" "$CLEAN_BUILD_BUDGET_S" \
  env GOCACHE="$cache" GOOS=linux GOARCH=amd64 go build -tags "$TAGS" -trimpath -o "$out/goro" .

time_step "test suite" "$TEST_SUITE_BUDGET_S" \
  go test -tags "$TAGS" -count=1 ./...

echo "== headless startup-to-login (budget ${HEADLESS_LOGIN_BUDGET_S}s)"
env HEADLESS_LOGIN_BUDGET_S="$HEADLESS_LOGIN_BUDGET_S" \
  go test -tags "$TAGS,budget" -count=1 -run TestHeadlessStartupToLoginBudget ./app
echo "   ok"

echo "== extra tool"
go build -trimpath -o "$out/goro-replay" ./cmd/goro-replay

echo "BUDGET GATE PASS $(git rev-parse --short=12 HEAD)"
