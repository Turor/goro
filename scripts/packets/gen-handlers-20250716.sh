#!/usr/bin/env bash
# Regenerates network/testdata/hercules_cz_handlers_20250716.txt from a
# Hercules source tree:
#
#   scripts/packets/gen-handlers-20250716.sh /opt/hercules-test/src > network/testdata/hercules_cz_handlers_20250716.txt
#
# It compiles map/packets.h plus map/packets_shuffle_main.h the way
# clif.c packetdb_loaddb() does for PACKETVER 20250716 (main client) and
# prints the final client->server opcode -> handler map. Later entries
# overwrite earlier ones, exactly like packetdb_addpacket.
set -euo pipefail
src="${1:?usage: gen-handlers-20250716.sh <hercules src dir>}"
tmp=$(mktemp -d)
trap "rm -rf \"$tmp\"" EXIT
cat >"$tmp/gen.c" <<C
#include <stdio.h>
#define PACKETVER 20250716
#define PACKETVER_MAIN_NUM 20250716
#define PACKETVER_RE_NUM 0
#define PACKETVER_ZERO_NUM 0
#define PACKETVER_SAK_NUM 0
#define PACKETVER_AD_NUM 0
#define packet(id, fn, ...) printf("0x%04X %s\n", id, #fn);
int main(void) {
#include "map/packets.h"
#include "map/packets_shuffle_main.h"
	return 0;
}
C
cc -I"$src" -o "$tmp/gen" "$tmp/gen.c"
"$tmp/gen" | awk '{m[$1]=$2} END {for (k in m) print k, m[k]}' | sort
