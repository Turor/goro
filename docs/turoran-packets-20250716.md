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

## Items, skills and chat (milestone 2)

- **32-bit item IDs.** Item and card IDs are `uint32` across `network`,
  `session`, `game`, `ui`, `world` and `client`. The 2008 parsers still read
  2-byte IDs and the 2008 builders still write them.
- **Inventory.** `0x0B08`/`0x0B0B` bracket each list. `0x0B09` (34-byte
  `NORMALITEM_INFO`) and `0x0B39` (83-byte `EQUIPITEM_INFO` with options and
  forge data) are routed by their inventory type byte: 0 inventory, 1 cart,
  2 storage. `0x0B41` pickup, `0x0ADD` and 19-byte `0x009D` floor items,
  `0x0999`/`0x099A` equip acks (result 0 is success here), `0x01C8` use ack.
- **Skills.** `0x0B32` list and `0x0B31`/`0x0B33` updates (15-byte
  `SKILLDATA`, no name on the wire), `0x0B1A` cast.
- **Hotkeys.** `0x0B20` tab 0 fills the 27-slot skill bar.
- **Client packets.** `network/testdata/hercules_cz_handlers_20250716.txt`
  is the Hercules opcode-to-handler map for this date
  (`scripts/packets/gen-handlers-20250716.sh`). A test checks that every
  date-aware builder lands on the intended handler with the Hercules length.
  That moved storage to `0x0364`/`0x0365` (`0x00F3` is global chat here),
  wear to `0x0998`, ground skills to `0x0366`, and the ground-skill talkbox
  to 31 bytes.

## Not yet in this profile

Turoran custom features (`0x0EF1`, `0x0EF5` and the rest), equipment
locations above 16 bits (shadow gear), item options and grades, guild storage
(inventory type 3), the extra hotkey tabs, and the remaining map packets
(`0x0A9B`, `0x097B`, `0x0446`, `0x02D9`, `0x0AE2`, `0x0A3B`, `0x099B`,
`0x0441`, `0x007F`) are framed and skipped. Builders that do not take a client
date (party, guild, trade, vending, mail) are not checked against the 20250716
handler map yet.
