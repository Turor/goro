package ui

import "github.com/kivutar/goro/ui/rotheme"

// [identify] defaults: the item appraisal window.
var (
	identifyWindowWidth  = 312
	identifyTableHeaderH = 36
	identifyRowH         = 32
	identifyRows         = 6
	identifyWindowHeight int
)

func init() {
	s := rotheme.Section("identify", "item appraisal window")
	s.Int("width", &identifyWindowWidth, 200, 900, "window width")
	s.Int("header_height", &identifyTableHeaderH, 12, 96, "table header height")
	s.Int("row_height", &identifyRowH, 12, 96, "item row height")
	s.Int("visible_rows", &identifyRows, 1, 24, "visible item rows")
	rotheme.OnLayoutApplied(recomputeIdentifyMetrics)
	recomputeIdentifyMetrics()
}

func recomputeIdentifyMetrics() {
	identifyWindowHeight = ROWindowTitleHeight + identifyTableHeaderH + identifyRows*identifyRowH + ROWindowFooterHeight
}
