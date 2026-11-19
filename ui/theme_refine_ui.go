package ui

import "github.com/kivutar/goro/ui/rotheme"

var (
	refineUIWindowW = 262
	refineUIWindowH = 348
	refineUIRowH    = 24
	refineUIRows    = 6
	refineUIPad     = 10
	refineUIGap     = 6
)

func init() {
	s := rotheme.Section("refine_ui", "refinery UI")
	s.Int("width", &refineUIWindowW, 240, 800, "window width")
	s.Int("height", &refineUIWindowH, 180, 700, "window height")
	s.Int("row_height", &refineUIRowH, 12, 48, "item row height")
	s.Int("visible_rows", &refineUIRows, 3, 20, "visible item rows")
	s.Int("padding", &refineUIPad, 0, 32, "content padding")
	s.Int("gap", &refineUIGap, 0, 24, "row gap")
}
