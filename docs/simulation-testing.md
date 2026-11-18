# Simulation testing (Phase 5)

goro's lean test suite stays offline and under 10 seconds on CT119. Phase 5
adds a golden replay harness, an opt-in live scenario runner, and an opt-in
budget gate. Phase 6 (swappable classic/modern UI) is not in this work.

## 5.1 Golden replay

Captures live in `sim/testdata/*.json` (`GORO-REPLAY-1`). Each stream is one
TCP connection (login, char, map). Records are hex-encoded packets. The char
stream may set `prefix_hex` for Hercules' unframed 4-byte account ID.

`sim/testdata/login-walk.json` is a **reduced** PACKETVER 20250716 fixture:
login accept (`0x0AC4`), character list, zone notify, map enter, walk ack,
plus framed fight, cast, storage, warp, item-mark (`0x0EE8`) and Warpra
(`0x0EFE`) packets. Addresses are RFC 5737 documentation hosts (`192.0.2.10`),
never live TEST IPs. A full TEST capture of those flows was not taken here.

Replay goes through `network.Framer`, the 20250716 decoders, login session
apply, and `WorldMode.ApplyReplayPacket`. Default `go test ./sim` does this
without opening a socket.

### Recording more fixtures

kgap pool characters only. Point `data/clientinfo.xml` at TEST from
uncommitted config.

```sh
GORO_LIVE_TEST=1 GORO_LIVE_ACCOUNT_FILE=/path/to/kgapNN.json \
  GORO_CAPTURE=sim/testdata/new.json \
  scripts/sim/record-capture.sh scenarios/login.lua
```

`--capture path.json` is the same flag. The file is s2c only (no passwords).
Before committing:

1. Replace live addresses with `192.0.2.0/24`.
2. Frame with `goro-replay frame FILE` or `goro-replay pack`.
3. Add a replay assertion in `sim/replay_test.go`.

Raw TCP dumps (one file per connection) pack the same way:

```sh
goro-replay pack -out sim/testdata/new.json -note "login+walk" \
  login=login.bin char=char.bin map=map.bin
```

Char dumps that start with the 4-byte account ID are split into `prefix_hex`
automatically. `goro-replay frame -skip 4 char.bin` lists opcodes.

## 5.2 Live scenario runner

```sh
GORO_LIVE_TEST=1 GORO_LIVE_ACCOUNT_FILE=/path/to/kgapNN.json \
  scripts/sim/live-runner.sh scenarios/login.lua
```

Without `GORO_LIVE_TEST=1` or an account file the runner and the default
`go test ./sim` checks skip. Usernames that are not `kgap[0-9]+` are refused.
`go test -tags live ./sim` runs the same runner when those env vars are set.

Scripts:

- `scenarios/login.lua` — enter the world, walk one cell, `goro.quit()`.
- `scenarios/warp.lua` — confirm in-world, then quit. A full Warpra list and
  locked-buy check is Phase 7.

### TEST database

If a scenario needs a known character state, use the existing TEST-safe
backup/restore. Take the backup and restore only while Hercules TEST is
offline. Never `korangar_m1_1108`. This job does not run that procedure.

## 5.3 Budgets (opt-in)

`scripts/release/budget-gate.sh` fails when a step exceeds its CT119 budget:

| Step | Default |
| --- | --- |
| Clean Linux build (private `GOCACHE`) | 15 s |
| `go test -tags nofakecgo -count=1 ./...` | 15 s |
| Headless startup-to-login (`-tags budget`) | 3 s |

Override with `CLEAN_BUILD_BUDGET_S`, `TEST_SUITE_BUDGET_S`,
`HEADLESS_LOGIN_BUDGET_S`. The default local gate does not run this script.

## 5.4 Lean-suite policy

Default `go test` (with `-tags nofakecgo` as in the local gate) stays under
10 seconds on CT119. Slow and real-data tests use a build tag:

| Tag | When |
| --- | --- |
| *(none)* | Offline lean suite, including golden replay |
| `slow` | Real client data (`GORO_DATA_DIR`) and long fixtures |
| `live` | Headless kgap run against TEST |
| `budget` | Headless startup-to-login timing |

Keep new tests that need GRFs, GPU, or TEST behind one of those tags.
