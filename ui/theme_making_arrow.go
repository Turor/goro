package ui

import "github.com/kivutar/goro/ui/rotheme"

// [making_arrow] defaults: the arrow crafting window.
var (
	makingArrowWindowWidth  = 312
	makingArrowTableHeaderH = 24
	makingArrowRowH         = 32
	makingArrowRows         = 6
	makingArrowWindowHeight int
)

func init() {
	s := rotheme.Section("making_arrow", "arrow crafting window")
	s.Int("width", &makingArrowWindowWidth, 200, 900, "window width")
	s.Int("header_height", &makingArrowTableHeaderH, 12, 80, "table header height")
	s.Int("row_height", &makingArrowRowH, 12, 96, "item row height")
	s.Int("visible_rows", &makingArrowRows, 1, 24, "visible item rows")
	rotheme.OnLayoutApplied(recomputeMakingArrowMetrics)
	recomputeMakingArrowMetrics()
}

func recomputeMakingArrowMetrics() {
	makingArrowWindowHeight = ROWindowTitleHeight + makingArrowTableHeaderH + makingArrowRows*makingArrowRowH + ROWindowFooterHeight
}
