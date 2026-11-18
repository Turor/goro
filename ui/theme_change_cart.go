package ui

import "github.com/kivutar/goro/ui/rotheme"

// [change_cart] defaults: the change-cart picker.
var (
	changeCartWindowWidth  = 248
	changeCartWindowHeight = 188
	changeCartPad          = 14
	changeCartGap          = 8
	changeCartButtonW      = 64
	changeCartColumns      = 3
)

func init() {
	s := rotheme.Section("change_cart", "change cart window")
	s.Int("width", &changeCartWindowWidth, 160, 720, "window width")
	s.Int("height", &changeCartWindowHeight, 100, 560, "window height")
	s.Int("padding", &changeCartPad, 0, 48, "content padding")
	s.Int("gap", &changeCartGap, 0, 32, "gap between cart buttons")
	s.Int("button_width", &changeCartButtonW, 32, 200, "cart number button width")
	s.Int("columns", &changeCartColumns, 1, 9, "cart buttons per row")
}
