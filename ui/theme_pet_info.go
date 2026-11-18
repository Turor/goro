package ui

import "github.com/kivutar/goro/ui/rotheme"

var (
	petInfoWindowW       = 292
	petInfoWindowH       = 178
	petInfoContentPad    = 12
	petInfoLabelW        = 70
	petInfoFieldW        = 120
	petInfoRowH          = 20
	petInfoNameFieldH    = 24
	petInfoRenameButtonW = 58
	petInfoRowGap        = 6
)

func init() {
	s := rotheme.Section("pet_info", "pet information window")
	s.Int("width", &petInfoWindowW, 180, 900, "window width")
	s.Int("height", &petInfoWindowH, 100, 700, "window height")
	s.Int("padding", &petInfoContentPad, 0, 48, "content padding")
	s.Int("label_width", &petInfoLabelW, 24, 200, "row label width")
	s.Int("field_width", &petInfoFieldW, 40, 400, "value and name field width")
	s.Int("row_height", &petInfoRowH, 12, 48, "info row height")
	s.Int("name_field_height", &petInfoNameFieldH, 16, 64, "name field height")
	s.Int("rename_button_width", &petInfoRenameButtonW, 24, 160, "rename button width")
	s.Int("row_gap", &petInfoRowGap, 0, 24, "gap between rows")
}
