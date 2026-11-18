package ui

import (
	"image/color"

	"github.com/kivutar/goro/ui/rotheme"
)

var (
	homunculusInfoWindowW     = 300
	homunculusInfoWindowH     = 292
	homunculusInfoContentPad  = 10
	homunculusInfoColumnGap   = 10
	homunculusInfoRowH        = 17
	homunculusInfoNameH       = 24
	homunculusInfoStatLabelW  = 40
	homunculusInfoStatValueW  = 40
	homunculusInfoInfoLabelW  = 50
	homunculusInfoInfoGap     = 4
	homunculusInfoNameGap     = 4
	homunculusInfoNameButtonW = 44
	homunculusInfoBarH        = 7
	homunculusInfoBarBorder   = 1
	homunculusInfoBarGap      = 2
	homunculusInfoRowGap      = 4
	homunculusInfoButtonW     = 72
	homunculusInfoBarBack     = color.RGBA{R: 66, G: 66, B: 66, A: 255}
	homunculusInfoExpColor    = WindowBorderColor
)

func init() {
	s := rotheme.Section("homunculus_info", "homunculus information window")
	s.Int("width", &homunculusInfoWindowW, 200, 1000, "window width")
	s.Int("height", &homunculusInfoWindowH, 160, 900, "window height")
	s.Int("padding", &homunculusInfoContentPad, 0, 40, "content padding")
	s.Int("column_gap", &homunculusInfoColumnGap, 0, 40, "gap between the stat table and details")
	s.Int("row_height", &homunculusInfoRowH, 12, 48, "text row height")
	s.Int("name_height", &homunculusInfoNameH, 16, 64, "name row height")
	s.Int("stat_label_width", &homunculusInfoStatLabelW, 16, 160, "stat label column width")
	s.Int("stat_value_width", &homunculusInfoStatValueW, 16, 160, "stat value column width")
	s.Int("info_label_width", &homunculusInfoInfoLabelW, 16, 160, "detail label width")
	s.Int("info_gap", &homunculusInfoInfoGap, 0, 24, "gap between a detail label and its value")
	s.Int("name_gap", &homunculusInfoNameGap, 0, 24, "gap in the name row")
	s.Int("name_button_width", &homunculusInfoNameButtonW, 24, 160, "rename button width")
	s.Int("bar_height", &homunculusInfoBarH, 2, 32, "custom bar height")
	s.Int("bar_border_width", &homunculusInfoBarBorder, 0, 8, "custom bar border width")
	s.Int("bar_gap", &homunculusInfoBarGap, 0, 16, "gap between a bar label and the bar")
	s.Int("row_gap", &homunculusInfoRowGap, 0, 24, "gap between detail rows")
	s.Int("button_width", &homunculusInfoButtonW, 32, 200, "footer button width")
	s.RGBA("bar_background", &homunculusInfoBarBack, "custom bar track")
	s.RGBA("exp_color", &homunculusInfoExpColor, "experience bar fill")
}
