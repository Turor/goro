package ui

import (
	"github.com/gogpu/ui/widget"
	"github.com/kivutar/goro/ui/rotheme"
)

// [shop] defaults: NPC shop buy and sell windows.
var (
	shopBuyListWindowW = 420
	shopBuyCartWindowW = 420
	shopTableHeaderH   = 36
	shopRowH           = 32
	shopListRows       = 7
	shopBuyCartRows    = 4
	shopSellCartRows   = 10
	shopWindowGap      = 20
	shopItemMinW       = float32(120)
	shopPriceW         = float32(104)
	shopQtyW           = float32(66)
	shopPriceOnlyW     = float32(124)
	shopQtyColor       = widget.RGBA8(54, 128, 76, 255)
)

func init() {
	s := rotheme.Section("shop", "NPC shop buy and sell windows")
	s.Int("buy_list_width", &shopBuyListWindowW, 240, 1400, "buy list window width")
	s.Int("buy_cart_width", &shopBuyCartWindowW, 240, 1400, "buy cart window width")
	s.Int("header_height", &shopTableHeaderH, 16, 96, "table header height")
	s.Int("row_height", &shopRowH, 12, 96, "item row height")
	s.Int("list_rows", &shopListRows, 1, 24, "visible shop list rows")
	s.Int("buy_cart_rows", &shopBuyCartRows, 1, 16, "visible buy cart rows")
	s.Int("sell_cart_rows", &shopSellCartRows, 1, 30, "visible sell cart rows")
	s.Int("window_gap", &shopWindowGap, 0, 80, "gap between the list and cart windows")
	s.Float("item_min_width", &shopItemMinW, 40, 480, "item column minimum width")
	s.Float("price_width", &shopPriceW, 40, 320, "price column width when quantity is shown")
	s.Float("quantity_width", &shopQtyW, 24, 240, "quantity column width")
	s.Float("price_only_width", &shopPriceOnlyW, 40, 400, "price column width without quantity")
	s.Color("quantity_color", &shopQtyColor, "quantity text color")
}
