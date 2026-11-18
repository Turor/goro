#!/bin/bash
# Offline renewal-data screenshots: enters each map without a server and grabs
# the screen from a virtual X display. Needs Xvfb, ImageMagick import and a
# Vulkan driver (Mesa lavapipe works on CPU-only machines).
#
#   scripts/data/preview-shots.sh BINARY DATA_DIR OUT_DIR "name map,x,y[,job] [actors]" ...
#
# Example:
#   scripts/data/preview-shots.sh ./goro /srv/goro-data /tmp/shots \
#     "rag_fes rag_fes,200,200" "jobs prontera,156,120,4218 0,7,4054,4252"
set -euo pipefail
bin=$1 data=$2 out=$3
shift 3
wait_s=${PREVIEW_WAIT:-60}
mkdir -p "$out"
for shot in "$@"; do
	read -r name spec actors <<<"$shot"
	display=":$((100 + RANDOM % 400))"
	Xvfb "$display" -screen 0 1280x720x24 >/dev/null 2>&1 &
	xvfb=$!
	sleep 1
	: >"$out/$name.ini"
	DISPLAY=$display GORO_SKIP_UPDATE=1 GORO_PREVIEW_MAP="$spec" GORO_PREVIEW_ACTORS="${actors:-}" \
		timeout $((wait_s + 20)) "$bin" --data-dir "$data" --no-audio --no-ui --windowed \
		--width 1280 --height 720 --config "$out/$name.ini" --log-file "$out/$name.log" >/dev/null 2>&1 &
	game=$!
	sleep "$wait_s"
	DISPLAY=$display import -window root "$out/$name.png"
	kill "$game" 2>/dev/null || true
	wait "$game" 2>/dev/null || true
	kill "$xvfb" 2>/dev/null || true
	echo "$out/$name.png"
done
