package ui

import "github.com/kivutar/goro/ui/rotheme"

// [trade] defaults: the player trade window.
var (
	tradeWindowW     = 560
	tradeWindowH     = 360
	tradePanelPad    = 10
	tradePanelGap    = 12
	tradePanelH      = 216
	tradeRowH        = 28
	tradeZenyFieldW  = 96
	tradeZenyFieldH  = 24
	tradeVisibleRows = 7
	tradeTitleH      = 20
	tradeZenyLabelW  = 42
	tradeZenyRowH    = 30
	tradeStatusH     = 18
	tradeRowGap      = 2
	tradePanelExtra  = 40
	tradeQtyW        = 42
	tradeIconExtra   = 8
	tradePanelW      int
)

func init() {
	s := rotheme.Section("trade", "player trade window")
	s.Int("width", &tradeWindowW, 320, 1600, "window width")
	s.Int("height", &tradeWindowH, 200, 1000, "window height")
	s.Int("padding", &tradePanelPad, 0, 40, "content padding")
	s.Int("gap", &tradePanelGap, 0, 48, "gap between the two panels")
	s.Int("panel_height", &tradePanelH, 80, 640, "item panel height before the extra footer")
	s.Int("row_height", &tradeRowH, 12, 80, "item row height")
	s.Int("zeny_field_width", &tradeZenyFieldW, 40, 320, "zeny field width")
	s.Int("zeny_field_height", &tradeZenyFieldH, 12, 64, "zeny field height")
	s.Int("visible_rows", &tradeVisibleRows, 1, 24, "item rows shown in each panel")
	s.Int("title_height", &tradeTitleH, 12, 64, "panel title row height")
	s.Int("zeny_label_width", &tradeZenyLabelW, 20, 160, "zeny label width")
	s.Int("zeny_row_height", &tradeZenyRowH, 12, 80, "zeny row height")
	s.Int("status_height", &tradeStatusH, 10, 48, "ok status row height")
	s.Int("row_gap", &tradeRowGap, 0, 16, "gap between panel rows")
	s.Int("panel_extra", &tradePanelExtra, 0, 160, "extra panel height below the item list")
	s.Int("quantity_width", &tradeQtyW, 20, 160, "quantity column width")
	s.Int("icon_extra", &tradeIconExtra, 0, 32, "extra width around the item icon")
	rotheme.OnLayoutApplied(recomputeTradeMetrics)
	recomputeTradeMetrics()
}

func recomputeTradeMetrics() {
	tradePanelW = (tradeWindowW - 2*tradePanelPad - tradePanelGap) / 2
}
