package ui

import "github.com/kivutar/goro/ui/rotheme"

var (
	achievementWindowW = 360
	achievementWindowH = 320
	achievementRowH    = 22
	achievementPad     = 10
	achievementGap     = 4
)

func init() {
	s := rotheme.Section("achievement", "achievement window")
	s.Int("width", &achievementWindowW, 240, 800, "window width")
	s.Int("height", &achievementWindowH, 160, 700, "window height")
	s.Int("row_height", &achievementRowH, 12, 48, "row height")
	s.Int("padding", &achievementPad, 0, 32, "content padding")
	s.Int("gap", &achievementGap, 0, 24, "row gap")
}
