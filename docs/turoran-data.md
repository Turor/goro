# TuroranRO data directory

goro reads the same data roBrowser serves for TuroranRO: the five GRFs in
`DATA.INI` order plus remoteclient2's loose overlay. Loose files beat every
archive in goro (as `OVERLAY_PATH` does in remoteclient2), so the overlay is
laid over the GRFs as plain files.

```sh
scripts/data/assemble-turoran-data.sh GRF_DIR LOOSE_DIR OUT_DIR
goro --data-dir OUT_DIR
```

- `GRF_DIR`: rodata-dev `/opt/remoteclient2/resources-0716` (`DATA.INI`,
  `translation.grf`, `2025-07-16_en.grf`, `data.grf`, `iro_assets.grf`,
  `region_assets.grf`).
- `LOOSE_DIR`: rodata-dev `/opt/remoteclient2` (`overlay-live-0716/`,
  `System/`).

The output is symlinks only. The script also links the tables roBrowser's
`DBManager.js` loads under Turoran names to their canonical client names:

| Canonical name | Turoran source |
|---|---|
| `System/itemInfo.lua` | `System/itemInfo.turoran.renewal11.lua` (29,596 items, English) |
| `lua files/skillinfoz/skillid.lub` | `skillid.turoran4.lub` |
| `lua files/skillinfoz/skillinfolist.lub` | `skillinfolist.turoran13.lub` |
| `lua files/skillinfoz/skilltreeview.lub` | `skilltreeview.fourth.lub` |
| `lua files/skillinfoz/skilldescript.lub` | `System/skilldescript.turoran.re2.lub` |
| `lua files/datainfo/petinfo.lub`, `addrandomoptionnametable.lub` | `*.turoran.lub` |
| `data/cardprefixnametable.txt`, `data/num2cardillustnametable.txt` | `*.turoran.txt` |

When roBrowser bumps one of these names, bump the script in the same change.

## Lua tables (.lub)

- **Compiled Lua 5.1 bytecode** (421 files in our GRFs) is translated in
  process into gopher-lua prototypes. All 421 translate; no decompiler.
- **Lua source** (the 77 translated overlay tables, the 23 MB itemInfo) is
  parsed with gopher-lua's parser. Files that are only global table
  assignments (every big table) are evaluated straight from the syntax tree;
  anything with code runs in gopher-lua. gopher-lua's compiler de-duplicates
  constants in quadratic time: the Turoran itemInfo takes 53 s to compile but
  0.5 s through the tree walk.
- Measured on CT119 with an assembled dir: items 0.5 s, skills (ids, info
  list, descriptions) 0.8 s, skill trees 0.4 s, NPC names 8 ms, all lazy.
- **Map names:** `mapnametable.txt` is loaded lazily (`Manager.MapDisplayName`). The minimap footer uses the display name when present, otherwise the map file stem. Renewal maps that the English table does not list (for example `rag_fes`, `bl_grass`) still show the stem.

## Checking a data set

```sh
go run ./cmd/goro-datacheck --data-dir OUT_DIR
GORO_DATA_DIR=OUT_DIR go test ./res/
```
