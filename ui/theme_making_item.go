package ui

import "github.com/kivutar/goro/ui/rotheme"

// [making_item] defaults: the item creation window.
var (
	makingItemWindowWidth  = 312
	makingItemTableHeaderH = 24
	makingItemRowH         = 32
	makingItemRows         = 6
	makingItemWindowHeight int
)

func init() {
	s := rotheme.Section("making_item", "item creation window")
	s.Int("width", &makingItemWindowWidth, 200, 900, "window width")
	s.Int("header_height", &makingItemTableHeaderH, 12, 80, "table header height")
	s.Int("row_height", &makingItemRowH, 12, 96, "item row height")
	s.Int("visible_rows", &makingItemRows, 1, 24, "visible item rows")
	rotheme.OnLayoutApplied(recomputeMakingItemMetrics)
	recomputeMakingItemMetrics()
}

func recomputeMakingItemMetrics() {
	makingItemWindowHeight = ROWindowTitleHeight + makingItemTableHeaderH + makingItemRows*makingItemRowH + ROWindowFooterHeight
}
