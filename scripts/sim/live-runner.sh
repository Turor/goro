#!/usr/bin/env bash
# Opt-in live scenario runner (Phase 5.2). Never part of the default gate.
#
#   GORO_LIVE_TEST=1 GORO_LIVE_ACCOUNT_FILE=/path/to/kgap01.json \
#     scripts/sim/live-runner.sh scenarios/login.lua
#
# kgap pool characters only. Credentials are read from the account file and
# never printed. The TEST login host comes from that uncommitted file or from
# data/clientinfo.xml — never commit LAN addresses.
set -euo pipefail
# Do not set -x: it would print --password.
cd "$(git rev-parse --show-toplevel)"
export PATH="/usr/local/go/bin:${PATH}"

if [[ "${GORO_LIVE_TEST:-}" != "1" ]]; then
  echo "live runner skipped: set GORO_LIVE_TEST=1" >&2
  exit 0
fi
if [[ -z "${GORO_LIVE_ACCOUNT_FILE:-}" ]]; then
  echo "live runner skipped: set GORO_LIVE_ACCOUNT_FILE to a kgap JSON" >&2
  exit 0
fi

script=${1:-scenarios/login.lua}
if [[ ! -f "$script" ]]; then
  echo "live runner: missing script $script" >&2
  exit 1
fi

python3 - "$script" <<'PY'
import json, os, re, subprocess, sys

path = os.environ["GORO_LIVE_ACCOUNT_FILE"]
script = sys.argv[1]
with open(path, encoding="utf-8") as f:
    acct = json.load(f)
user = acct.get("username") or ""
password = acct.get("password") or ""
if not re.fullmatch(r"kgap[0-9]+", user):
    sys.stderr.write("live runner: username is not a kgap pool character\n")
    sys.exit(1)
if not password:
    sys.stderr.write("live runner: account file has no password\n")
    sys.exit(1)

goro = os.environ.get("GORO_BIN", "./goro")
if not os.path.isfile(goro):
    subprocess.check_call(["go", "build", "-tags", "nofakecgo", "-o", "goro", "."])
    goro = "./goro"

data_dir = os.environ.get("GORO_DATA_DIR", "")
cmd = [goro, "--headless", "--username", user, "--password", password, "--char-slot", os.environ.get("GORO_LIVE_CHAR_SLOT", "0"), "--script", script]
if data_dir:
    cmd[1:1] = ["--data-dir", data_dir]
sys.stderr.write("live runner: starting headless %s as %s\n" % (script, user))
# argv contains the password; never log the command line or a CalledProcessError.
timeout = int(os.environ.get("GORO_LIVE_TIMEOUT", "60"))
try:
    result = subprocess.run(cmd, check=False, timeout=timeout)
except subprocess.TimeoutExpired:
    sys.stderr.write("live runner: timed out after %ss\n" % timeout)
    sys.exit(1)
except OSError as err:
    sys.stderr.write("live runner: failed to start goro: %s\n" % err)
    sys.exit(1)
if result.returncode != 0:
    sys.stderr.write("live runner: goro exited %d\n" % result.returncode)
    sys.exit(result.returncode)
PY
