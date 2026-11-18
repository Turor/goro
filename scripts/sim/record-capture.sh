#!/usr/bin/env bash
# Record a golden capture from a kgap headless session (Phase 5.1).
#
#   GORO_LIVE_TEST=1 GORO_LIVE_ACCOUNT_FILE=/path/to/kgap01.json \
#     GORO_CAPTURE=sim/testdata/new.json \
#     scripts/sim/record-capture.sh scenarios/login.lua
#
# Before committing: replace any live IPs with 192.0.2.0/24, drop credentials
# (s2c only is recorded), and add a replay assertion. Full TEST captures of
# warp/fight/cast/storage/item marks/Warpra were not taken in the Phase 5 job;
# sim/testdata/login-walk.json is the reduced in-repo fixture.
set -euo pipefail
cd "$(git rev-parse --show-toplevel)"
if [[ -z "${GORO_CAPTURE:-}" ]]; then
  echo "record-capture: set GORO_CAPTURE to an output JSON path" >&2
  exit 2
fi
exec "$(dirname "$0")/live-runner.sh" "${1:-scenarios/login.lua}"
