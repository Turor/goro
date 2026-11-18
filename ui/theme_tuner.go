package ui

import "github.com/kivutar/goro/ui/rotheme"

// [tuner] defaults: the --ui-dev F11 UI tuner panel.
var (
	layoutTunerW            = 520
	layoutTunerRows         = 14
	layoutTunerRowH         = 22
	layoutTunerRowGap       = 2
	layoutTunerHeaderH      = 52
	layoutTunerContentPad   = 8
	layoutTunerKeyW         = 230
	layoutTunerSearchLabelW = 48
	layoutTunerStatusW      = 130
	layoutTunerScreenRight  = 16
	layoutTunerScreenTop    = 60
)

func init() {
	s := rotheme.Section("tuner", "the --ui-dev F11 UI tuner panel")
	s.Int("width", &layoutTunerW, 360, 1200, "tuner width")
	s.Int("rows", &layoutTunerRows, 4, 40, "keys shown per page")
	s.Int("row_height", &layoutTunerRowH, 14, 40, "key row height")
	s.Int("row_gap", &layoutTunerRowGap, 0, 16, "gap between key rows")
	s.Int("header_height", &layoutTunerHeaderH, 20, 120, "section and search rows height")
	s.Int("padding", &layoutTunerContentPad, 0, 32, "content padding")
	s.Int("key_width", &layoutTunerKeyW, 80, 600, "key name column width")
	s.Int("search_label_width", &layoutTunerSearchLabelW, 0, 200, "search label width")
	s.Int("status_width", &layoutTunerStatusW, 40, 400, "page and status text width")
	s.Int("screen_right", &layoutTunerScreenRight, 0, 400, "gap from the right screen edge when opened")
	s.Int("screen_top", &layoutTunerScreenTop, 0, 400, "distance from the screen top when opened")
}
