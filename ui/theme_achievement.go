package ui

import "github.com/kivutar/goro/ui/rotheme"

var (
	achievementWindowW  = 750
	achievementWindowH  = 468
	achievementRowH     = 22
	achievementPad      = 10
	achievementGap      = 4
	achievementSidebarW = 140
	achievementDetailH  = 140
)

func init() {
	s := rotheme.Section("achievement", "achievement window")
	s.Int("width", &achievementWindowW, 240, 900, "window width")
	s.Int("height", &achievementWindowH, 160, 700, "window height")
	s.Int("row_height", &achievementRowH, 12, 48, "row height")
	s.Int("padding", &achievementPad, 0, 32, "content padding")
	s.Int("gap", &achievementGap, 0, 24, "row gap")
	s.Int("sidebar_width", &achievementSidebarW, 80, 240, "category sidebar width")
	s.Int("detail_height", &achievementDetailH, 80, 280, "selected achievement detail height")
}
