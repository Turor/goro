package ui

import "github.com/kivutar/goro/ui/rotheme"

// [card_composition] defaults: the insert-card window.
var (
	cardCompositionWindowWidth  = 328
	cardCompositionTableHeaderH = 24
	cardCompositionRowH         = 32
	cardCompositionRows         = 6
	cardCompositionWindowHeight int
)

func init() {
	s := rotheme.Section("card_composition", "insert card window")
	s.Int("width", &cardCompositionWindowWidth, 200, 900, "window width")
	s.Int("header_height", &cardCompositionTableHeaderH, 12, 80, "table header height")
	s.Int("row_height", &cardCompositionRowH, 12, 96, "item row height")
	s.Int("visible_rows", &cardCompositionRows, 1, 24, "visible item rows")
	rotheme.OnLayoutApplied(recomputeCardCompositionMetrics)
	recomputeCardCompositionMetrics()
}

func recomputeCardCompositionMetrics() {
	cardCompositionWindowHeight = ROWindowTitleHeight + cardCompositionTableHeaderH + cardCompositionRows*cardCompositionRowH + ROWindowFooterHeight
}
