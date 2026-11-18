package ui

import "github.com/kivutar/goro/ui/rotheme"

// [vending] defaults: the vending setup, buy and own-shop windows.
var (
	vendingWindowW        = 420
	vendingSetupRows      = 7
	vendingBuyRows        = 7
	vendingCartRows       = 4
	vendingNameH          = 30
	vendingWindowGap      = 20
	vendingNameLabelW     = 54
	vendingFieldH         = 22
	vendingNameFieldInset = 72
	vendingPadX           = 8
	vendingPadY           = 4
	vendingContentGap     = 4
	vendingPriceW         = 90
	vendingNameGap        = 8
)

func init() {
	s := rotheme.Section("vending", "vending setup and buy windows")
	s.Int("width", &vendingWindowW, 240, 1400, "window width")
	s.Int("setup_rows", &vendingSetupRows, 1, 24, "visible rows while setting up a shop")
	s.Int("buy_rows", &vendingBuyRows, 1, 24, "visible rows in a vendor's shop")
	s.Int("cart_rows", &vendingCartRows, 1, 16, "visible rows in the buying cart")
	s.Int("name_height", &vendingNameH, 16, 80, "shop name row height")
	s.Int("window_gap", &vendingWindowGap, 0, 80, "gap between the two vending windows")
	s.Int("name_label_width", &vendingNameLabelW, 24, 200, "name label width")
	s.Int("field_height", &vendingFieldH, 12, 64, "name and price field height")
	s.Int("name_field_inset", &vendingNameFieldInset, 0, 240, "width kept beside the shop name field")
	s.Int("padding_x", &vendingPadX, 0, 32, "setup content horizontal padding")
	s.Int("padding_y", &vendingPadY, 0, 32, "setup content vertical padding")
	s.Int("content_gap", &vendingContentGap, 0, 24, "gap inside the setup content")
	s.Int("price_width", &vendingPriceW, 40, 280, "price field width")
	s.Int("name_gap", &vendingNameGap, 0, 32, "gap below the shop name row")
}
