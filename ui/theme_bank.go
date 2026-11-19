package ui

import "github.com/kivutar/goro/ui/rotheme"

var (
	bankWindowW = 320
	bankWindowH = 200
	bankFieldH  = 28
	bankPad     = 12
	bankGap     = 8
	bankLabelW  = 72
)

func init() {
	s := rotheme.Section("bank", "bank window")
	s.Int("width", &bankWindowW, 220, 700, "window width")
	s.Int("height", &bankWindowH, 140, 500, "window height")
	s.Int("field_height", &bankFieldH, 16, 48, "amount field height")
	s.Int("padding", &bankPad, 0, 32, "content padding")
	s.Int("gap", &bankGap, 0, 24, "row gap")
	s.Int("label_width", &bankLabelW, 40, 160, "amount label width")
}
