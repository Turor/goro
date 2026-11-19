package ui

import "github.com/kivutar/goro/ui/rotheme"

var (
	enchantUIWindowW = 360
	enchantUIWindowH = 280
	enchantUIRowH    = 24
	enchantUIRows    = 5
	enchantUIPad     = 10
	enchantUIGap     = 6
	enchantUILabelW  = 80
)

func init() {
	s := rotheme.Section("enchant_ui", "item enchant UI")
	s.Int("width", &enchantUIWindowW, 240, 800, "window width")
	s.Int("height", &enchantUIWindowH, 160, 700, "window height")
	s.Int("row_height", &enchantUIRowH, 12, 48, "item row height")
	s.Int("visible_rows", &enchantUIRows, 3, 20, "visible item rows")
	s.Int("padding", &enchantUIPad, 0, 32, "content padding")
	s.Int("gap", &enchantUIGap, 0, 24, "row gap")
	s.Int("label_width", &enchantUILabelW, 40, 160, "field label width")
}
