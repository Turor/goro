package ui

import (
	"github.com/gogpu/ui/widget"
	"github.com/kivutar/goro/ui/rotheme"
)

// [pvp_counter] defaults: the corner PvP rank readout and its rank font.
var (
	pvpCounterWidth            = 240
	pvpCounterHeight           = 96
	pvpCounterShadow           = widget.RGBA8(20, 28, 38, 190)
	pvpCounterText             = widget.RGBA8(245, 248, 252, 255)
	pvpCounterFont     float32 = 22
	pvpCounterShadowDX float32 = 1
	pvpCounterShadowDY float32 = 2
	pvpRankBaseline            = 52
	pvpRankStep                = 28
	pvpRankDigitShift          = 6
)

func init() {
	s := rotheme.Section("pvp_counter", "corner pvp rank readout")
	s.Int("width", &pvpCounterWidth, 40, 800, "counter width")
	s.Int("height", &pvpCounterHeight, 16, 400, "counter height")
	s.Color("shadow", &pvpCounterShadow, "fallback rank shadow")
	s.Color("text", &pvpCounterText, "fallback rank text")
	s.Float("font_size", &pvpCounterFont, 8, 72, "fallback rank text size")
	s.Float("shadow_x", &pvpCounterShadowDX, -16, 16, "fallback rank shadow x offset")
	s.Float("shadow_y", &pvpCounterShadowDY, -16, 16, "fallback rank shadow y offset")
	s.Int("baseline", &pvpRankBaseline, 0, 200, "rank font baseline")
	s.Int("step", &pvpRankStep, 4, 96, "rank font glyph advance")
	s.Int("digit_shift", &pvpRankDigitShift, 0, 48, "rank and total digit shift from the baseline")
}
