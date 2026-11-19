package ui

import "github.com/kivutar/goro/ui/rotheme"

var (
	gradeWindowW = 360
	gradeWindowH = 320
	gradeRowH    = 24
	gradePad     = 10
	gradeGap     = 6
)

func init() {
	s := rotheme.Section("grade_enchant", "grade enchant UI")
	s.Int("width", &gradeWindowW, 240, 800, "window width")
	s.Int("height", &gradeWindowH, 160, 700, "window height")
	s.Int("row_height", &gradeRowH, 12, 48, "row height")
	s.Int("padding", &gradePad, 0, 32, "content padding")
	s.Int("gap", &gradeGap, 0, 24, "row gap")
}
