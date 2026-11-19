package ui

import "github.com/kivutar/goro/ui/rotheme"

var (
	pincodeWindowW = 280
	pincodeWindowH = 160
	pincodeFieldH  = 28
	pincodePad     = 12
	pincodeGap     = 8
	pincodeLabelW  = 80
)

func init() {
	s := rotheme.Section("pincode", "character PIN window")
	s.Int("width", &pincodeWindowW, 200, 600, "window width")
	s.Int("height", &pincodeWindowH, 120, 400, "window height")
	s.Int("field_height", &pincodeFieldH, 16, 48, "PIN field height")
	s.Int("padding", &pincodePad, 0, 32, "content padding")
	s.Int("gap", &pincodeGap, 0, 24, "row gap")
	s.Int("label_width", &pincodeLabelW, 40, 160, "field label width")
}
