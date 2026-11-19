package ui

import "github.com/kivutar/goro/ui/rotheme"

var (
	leftoverWindowW = 320
	leftoverWindowH = 180
	leftoverLabelW  = 56
	leftoverRowH    = 24
	leftoverPad     = 10
	leftoverGap     = 4
	edenWindowW     = 460
	edenWindowH     = 320
)

func init() {
	s := rotheme.Section("leftover_ui", "reform, laphine, styling, and eden windows")
	s.Int("width", &leftoverWindowW, 180, 700, "small leftover window width")
	s.Int("height", &leftoverWindowH, 80, 500, "small leftover window height")
	s.Int("label_width", &leftoverLabelW, 32, 160, "index label width")
	s.Int("row_height", &leftoverRowH, 12, 48, "row height")
	s.Int("padding", &leftoverPad, 0, 32, "content padding")
	s.Int("gap", &leftoverGap, 0, 24, "row gap")
	s.Int("eden_width", &edenWindowW, 240, 800, "eden window width")
	s.Int("eden_height", &edenWindowH, 140, 700, "eden window height")
}
