# Turoran packet profile 20250716

TuroranRO runs Hercules with `PACKETVER 20250716` (main client, renewal). goro
keeps the upstream 2008-09-10 profile and adds this one next to it. Select it
with `--packet-client-date 20250716` or:

```ini
[packet]
client_date = 20250716
```

## What the profile changes

- **Length table.** `network/packet_lengths_20250716.go` is generated from
  Hercules `common/packets_len.h`
  (`scripts/packets/gen-lengths-20250716.sh <hercules>/src`). Hercules
  statically asserts that every fixed packet struct matches this table.
  Turoran plugin packets (`0x0EE8`–`0x0EFF`) are added from
  `turoranCustomLengths` in `network/packet_profile.go`. `0x0EF1` (features)
  goes to every client after every map load, and `0x0EF5` (MVP minimap) to
  every client on MVP maps, so both must be known even though goro ignores them.
- **Framing.** Each connection starts with a fresh framer. In this profile the
  bare 4-byte account ID that the char server writes before its first packet
  is skipped explicitly instead of by resync.
- **Login.** `0x0AC4` accept (64-byte header, 160-byte server rows) and
  `0x083E`/`0x0B02` refusals. A 24-character password sends client type 0,
  because Hercules reads the password with no NUL straight into that byte.
- **Char server.** `0x006B` and `0x0B72` lists with 175-byte `CHARACTER_INFO`
  (64-bit EXP and HP/SP), `0x0B6F` make-char accept, `0x0AC5` zone server.
- **Map.** `0x02EB`/`0x0A18` accept enter, `0x09FF`/`0x09FE`/`0x09FD` unit
  entries (account ID, 32-bit shield, robe, level), `0x0ACB` 64-bit parameters
  and `0x0ACC` 64-bit EXP notifications.

Outgoing packets already used the unshuffled 2018+ opcodes (`0x0436` enter,
`0x035F` walk, `0x0360` tick, `0x0361` direction, `0x0368` name, `0x0437`
action), which match Hercules `packets_shuffle_main.h` for this date.

## Not yet in this profile

Inventory and equipment (`0x0B08`–`0x0B0B`, `0x0B39`, `0x0B41`, 32-bit item
IDs), hotkeys (`0x0B20`), skills and casting (`0x0B1A`, `0x0B32`), and the
Turoran custom features are parsed in later milestones. Until then the client
frames and skips them.
