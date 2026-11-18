package ui

import "github.com/kivutar/goro/ui/rotheme"

// [equipment_choice] defaults: the repair and weapon-refine item lists.
var (
	equipmentChoiceWindowWidth  = 312
	equipmentChoiceTableHeaderH = 24
	equipmentChoiceRowH         = 32
	equipmentChoiceRows         = 6
	equipmentChoiceWindowHeight int
)

func init() {
	s := rotheme.Section("equipment_choice", "repair and weapon refine windows")
	s.Int("width", &equipmentChoiceWindowWidth, 200, 900, "window width")
	s.Int("header_height", &equipmentChoiceTableHeaderH, 12, 80, "table header height")
	s.Int("row_height", &equipmentChoiceRowH, 12, 96, "item row height")
	s.Int("visible_rows", &equipmentChoiceRows, 1, 24, "visible item rows")
	rotheme.OnLayoutApplied(recomputeEquipmentChoiceMetrics)
	recomputeEquipmentChoiceMetrics()
}

func recomputeEquipmentChoiceMetrics() {
	equipmentChoiceWindowHeight = ROWindowTitleHeight + equipmentChoiceTableHeaderH + equipmentChoiceRows*equipmentChoiceRowH + ROWindowFooterHeight
}
