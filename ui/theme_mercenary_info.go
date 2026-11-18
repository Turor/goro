package ui

import (
	"image/color"

	"github.com/kivutar/goro/ui/rotheme"
)

var (
	mercenaryInfoWindowH      = 272
	mercenaryInfoButtonW      = 72
	mercenaryInfoTimeLowRatio = 0.25
	mercenaryInfoTimeColor    = WindowBorderColor
	mercenaryInfoTimeLow      = color.RGBA{R: 0xFF, G: 0x1E, B: 0x00, A: 0xFF}
	mercenaryInfoKillsColor   = WindowBorderColor
)

func init() {
	s := rotheme.Section("mercenary_info", "mercenary information window")
	s.Int("height", &mercenaryInfoWindowH, 160, 900, "window height")
	s.Int("button_width", &mercenaryInfoButtonW, 32, 200, "footer button width")
	s.Float64("time_low_ratio", &mercenaryInfoTimeLowRatio, 0, 1, "remaining time ratio that switches the time bar color")
	s.RGBA("time_color", &mercenaryInfoTimeColor, "contract time bar fill")
	s.RGBA("time_low_color", &mercenaryInfoTimeLow, "contract time bar fill when time is low")
	s.RGBA("kills_color", &mercenaryInfoKillsColor, "kills bar fill")
}
