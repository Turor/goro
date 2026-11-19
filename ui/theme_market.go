package ui

import "github.com/kivutar/goro/ui/rotheme"

var (
	rouletteWindowW = 280
	rouletteWindowH = 160
	roulettePad     = 10
	rouletteGap     = 6

	marketWindowW = 420
	marketWindowH = 280
	marketRowH    = 24
	marketPad     = 10
	marketGap     = 4

	clanWindowW = 280
	clanWindowH = 180
	clanPad     = 10
	clanGap     = 4
)

func init() {
	r := rotheme.Section("roulette", "roulette window")
	r.Int("width", &rouletteWindowW, 180, 600, "window width")
	r.Int("height", &rouletteWindowH, 80, 400, "window height")
	r.Int("padding", &roulettePad, 0, 32, "content padding")
	r.Int("gap", &rouletteGap, 0, 24, "row gap")

	m := rotheme.Section("npc_market", "npc market and barter windows")
	m.Int("width", &marketWindowW, 240, 800, "window width")
	m.Int("height", &marketWindowH, 140, 700, "window height")
	m.Int("row_height", &marketRowH, 12, 48, "item row height")
	m.Int("padding", &marketPad, 0, 32, "content padding")
	m.Int("gap", &marketGap, 0, 24, "row gap")

	c := rotheme.Section("clan", "clan info window")
	c.Int("width", &clanWindowW, 180, 600, "window width")
	c.Int("height", &clanWindowH, 80, 400, "window height")
	c.Int("padding", &clanPad, 0, 32, "content padding")
	c.Int("gap", &clanGap, 0, 24, "row gap")
}
