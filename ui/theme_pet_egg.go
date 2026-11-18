package ui

import "github.com/kivutar/goro/ui/rotheme"

// [pet_egg] defaults: the pet egg selection window.
var (
	petEggWindowWidth  = 312
	petEggTableHeaderH = 24
	petEggRowH         = 32
	petEggRows         = 6
	petEggWindowHeight int
)

func init() {
	s := rotheme.Section("pet_egg", "pet egg selection window")
	s.Int("width", &petEggWindowWidth, 200, 900, "window width")
	s.Int("header_height", &petEggTableHeaderH, 12, 80, "table header height")
	s.Int("row_height", &petEggRowH, 12, 96, "item row height")
	s.Int("visible_rows", &petEggRows, 1, 24, "visible item rows")
	rotheme.OnLayoutApplied(recomputePetEggMetrics)
	recomputePetEggMetrics()
}

func recomputePetEggMetrics() {
	petEggWindowHeight = ROWindowTitleHeight + petEggTableHeaderH + petEggRows*petEggRowH + ROWindowFooterHeight
}
