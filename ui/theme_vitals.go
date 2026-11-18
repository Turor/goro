package ui

import "github.com/kivutar/goro/ui/rotheme"

// [vitals] defaults: the shared HP and SP row.
var (
	vitalsLabelWidth   float32 = 22
	vitalsPercentWidth float32 = 34
	vitalsGap          float32 = 6
)

func init() {
	s := rotheme.Section("vitals", "hp and sp rows shared by the character and companion windows")
	s.Float("label_width", &vitalsLabelWidth, 8, 80, "vitals label width")
	s.Float("percent_width", &vitalsPercentWidth, 16, 120, "vitals percent width")
	s.Float("gap", &vitalsGap, 0, 24, "gap between the vitals label, bar and percent")
}
