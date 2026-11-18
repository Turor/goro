package ui

import "github.com/kivutar/goro/ui/rotheme"

// [status] defaults: the status window. Row height, padding and the column
// gap were the first tokens; the column widths follow them.
var (
	statsWindowPad  float32 = 10
	statsRowH               = 18
	statsSectionGap float32 = 10

	statsPrimaryLabelWidth float32 = 44
	statsPrimaryValueWidth float32 = 56
	statsPrimaryCostWidth  float32 = 28
	statsPrimaryButtonGap  float32 = 7
	statsDerivedLabelWidth float32 = 46
	statsDerivedValueWidth float32 = 62
	statsPrimaryRows               = 6
	statsPositionGap               = 12

	statsPrimaryColumnWidth float32
	statsDerivedPairWidth   float32
	statsDerivedColumnWidth float32
	statsWindowWidth        int
	statsWindowHeight       int
)

func init() {
	s := rotheme.Section("status", "status window")
	s.Int("row_height", &statsRowH, 12, 40, "status window row height")
	s.Float("padding", &statsWindowPad, 0, 32, "status window padding")
	s.Float("section_gap", &statsSectionGap, 0, 48, "gap between the status window's stat columns")
	s.Float("primary_label_width", &statsPrimaryLabelWidth, 20, 160, "primary stat label width")
	s.Float("primary_value_width", &statsPrimaryValueWidth, 24, 200, "primary stat value width")
	s.Float("primary_cost_width", &statsPrimaryCostWidth, 16, 120, "primary stat cost width")
	s.Float("primary_button_gap", &statsPrimaryButtonGap, 0, 28, "gap before the stat plus button")
	s.Float("derived_label_width", &statsDerivedLabelWidth, 20, 160, "derived stat label width")
	s.Float("derived_value_width", &statsDerivedValueWidth, 24, 200, "derived stat value width")
	s.Int("primary_rows", &statsPrimaryRows, 1, 12, "primary stat rows used to size the window")
	s.Int("position_gap", &statsPositionGap, 0, 48, "gap between the character window and the status window")
	rotheme.OnLayoutApplied(recomputeStatsWindowMetrics)
	recomputeStatsWindowMetrics()
}
