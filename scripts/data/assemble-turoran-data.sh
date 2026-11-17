#!/usr/bin/env bash
# Assemble a goro data directory that matches what roBrowser serves for
# TuroranRO: the GRFs in DATA.INI order, the loose overlay (remoteclient2
# OVERLAY_PATH) on top, and the Turoran table picks roBrowser's DBManager
# loads under their canonical client names.
#
# Everything is a symlink into the source trees; nothing is copied or
# modified. Sources (read-only copies on CT119):
#   GRF_DIR    rodata-dev /opt/remoteclient2/resources-0716 (DATA.INI + *.grf)
#   LOOSE_DIR  rodata-dev /opt/remoteclient2 (overlay-live-0716/, System/)
#
# Usage: assemble-turoran-data.sh GRF_DIR LOOSE_DIR OUT_DIR
set -euo pipefail

[[ $# == 3 ]] || { echo "usage: $0 GRF_DIR LOOSE_DIR OUT_DIR" >&2; exit 2; }
grf_dir=$(realpath "$1")
loose_dir=$(realpath "$2")
out=$3
overlay="$loose_dir/overlay-live-0716"
system="$loose_dir/System"
[[ -f "$grf_dir/DATA.INI" ]] || { echo "no DATA.INI in $grf_dir" >&2; exit 1; }
[[ -d "$overlay" && -d "$system" ]] || { echo "need $overlay and $system" >&2; exit 1; }
if [[ -e "$out" ]] && [[ -n "$(ls -A "$out")" ]]; then
  echo "$out exists and is not empty" >&2
  exit 1
fi
mkdir -p "$out"
out=$(realpath "$out")

# 1. Archives, in DATA.INI order (goro reads DATA.INI itself).
ln -s "$grf_dir/DATA.INI" "$out/DATA.INI"
for grf in "$grf_dir"/*.grf; do
  ln -s "$grf" "$out/$(basename "$grf")"
done

# 2. Loose overlay: loose files win over every archive, as OVERLAY_PATH does.
cp -as "$overlay/data" "$out/data"
# cp -as copies the source directory modes; only our new directories change.
find "$out/data" -type d -exec chmod u+w,go+rx {} +
mkdir -p "$out/System"

# 3. Canonical names for the tables roBrowser loads under Turoran names
#    (robrowserlegacy src/DB/DBManager.js).
lua="$out/data/luafiles514/lua files"
pick() { # pick SOURCE DEST
  [[ -f "$1" ]] || { echo "missing $1" >&2; exit 1; }
  mkdir -p "$(dirname "$2")"
  ln -sfn "$1" "$2"
}
pick "$overlay/data/luafiles514/lua files/skillinfoz/skillid.turoran4.lub" "$lua/skillinfoz/skillid.lub"
pick "$overlay/data/luafiles514/lua files/skillinfoz/skillinfolist.turoran13.lub" "$lua/skillinfoz/skillinfolist.lub"
pick "$overlay/data/luafiles514/lua files/skillinfoz/skilltreeview.fourth.lub" "$lua/skillinfoz/skilltreeview.lub"
pick "$system/skilldescript.turoran.re2.lub" "$lua/skillinfoz/skilldescript.lub"
pick "$overlay/data/luafiles514/lua files/datainfo/petinfo.turoran.lub" "$lua/datainfo/petinfo.lub"
pick "$overlay/data/luafiles514/lua files/datainfo/addrandomoptionnametable.turoran.lub" "$lua/datainfo/addrandomoptionnametable.lub"
pick "$overlay/data/cardprefixnametable.turoran.txt" "$out/data/cardprefixnametable.txt"
pick "$overlay/data/num2cardillustnametable.turoran.txt" "$out/data/num2cardillustnametable.txt"
pick "$system/itemInfo.turoran.renewal11.lua" "$out/System/itemInfo.lua"

echo "assembled $out"
