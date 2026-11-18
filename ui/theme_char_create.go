package ui

import (
	"github.com/gogpu/ui/widget"
	"github.com/kivutar/goro/ui/rotheme"
)

// Character create window. The preview sprite scale lives in game/login and
// shares this section.
var (
	characterCreateWindowW                    = 576
	characterCreateWindowH                    = 342
	characterCreatePanelW                     = 142
	characterCreatePanelH                     = 166
	characterCreateGraphW                     = 194
	characterCreateGraphH                     = 230
	characterCreateColumnGap                  = 18
	characterCreateGraphTopPad                = 15
	characterCreateNameColumnTopPad           = 15
	characterCreateNameLabelH                 = 22
	characterCreateNameGap                    = 5
	characterCreateNameTopPad                 = 8
	characterCreateNameInputH                 = 22
	characterCreateGraphOuterRadius           = 58
	characterCreateGraphInnerRadius           = 29
	characterCreateStatButtonRadius           = 90
	characterCreateStatButtonW                = 38
	characterCreateStatButtonH                = 36
	characterCreateStatLabelW                 = 32
	characterCreateStatValueW                 = 95
	characterCreateStatRowH                   = 22
	characterCreateStatPanelPad               = 2
	characterCreateHairColorButtonY           = 5
	characterCreateHairStyleButtonY           = 39
	characterCreateContentPadX                = 32
	characterCreatePreviewButtonInset         = 16
	characterCreateStatScaleMin       float32 = 0.22
	characterCreateStatScaleSpan      float32 = 0.78
	characterCreateBorderWidth        float32 = 1
	characterCreateGraphSpoke                 = widget.RGBA8(185, 204, 224, 150)
	characterCreateGraphFill                  = widget.RGBA8(36, 92, 154, 220)

	characterCreatePreviewTopPad int
	characterCreateNameLift      int
)

func recomputeCharacterCreateMetrics() {
	characterCreatePreviewTopPad = (characterCreateGraphH - characterCreatePanelH) / 2
	characterCreateNameLift = characterCreatePreviewTopPad - characterCreateNameColumnTopPad
}

func init() {
	s := rotheme.Section("char_create", "character create screen")
	s.Int("width", &characterCreateWindowW, 360, 1400, "character create width")
	s.Int("height", &characterCreateWindowH, 200, 900, "character create height")
	s.Int("panel_width", &characterCreatePanelW, 64, 400, "preview and name column width")
	s.Int("panel_height", &characterCreatePanelH, 64, 400, "preview panel height")
	s.Int("graph_width", &characterCreateGraphW, 80, 500, "stat graph width")
	s.Int("graph_height", &characterCreateGraphH, 80, 500, "stat graph height")
	s.Int("column_gap", &characterCreateColumnGap, 0, 64, "gap between the preview and the stat graph")
	s.Int("graph_top_pad", &characterCreateGraphTopPad, 0, 64, "space above the stat graph")
	s.Int("name_column_top_pad", &characterCreateNameColumnTopPad, 0, 80, "space above the preview column")
	s.Int("name_label_height", &characterCreateNameLabelH, 12, 48, "name label height")
	s.Int("name_gap", &characterCreateNameGap, 0, 32, "gap between the name label and the field")
	s.Int("name_top_pad", &characterCreateNameTopPad, 0, 40, "space above the name label")
	s.Int("name_input_height", &characterCreateNameInputH, 14, 48, "name field height")
	s.Int("graph_outer_radius", &characterCreateGraphOuterRadius, 16, 200, "outer stat hexagon radius")
	s.Int("graph_inner_radius", &characterCreateGraphInnerRadius, 8, 160, "inner stat hexagon radius")
	s.Int("stat_button_radius", &characterCreateStatButtonRadius, 24, 240, "distance from the graph center to a stat button")
	s.Int("stat_button_width", &characterCreateStatButtonW, 16, 96, "stat button width")
	s.Int("stat_button_height", &characterCreateStatButtonH, 16, 96, "stat button height")
	s.Int("stat_label_width", &characterCreateStatLabelW, 16, 160, "stat name column width")
	s.Int("stat_value_width", &characterCreateStatValueW, 24, 240, "stat value column width")
	s.Int("stat_row_height", &characterCreateStatRowH, 12, 48, "stat table row height")
	s.Int("stat_panel_pad", &characterCreateStatPanelPad, 0, 24, "padding around the stat table")
	s.Int("hair_color_button_y", &characterCreateHairColorButtonY, 0, 80, "hair color button offset from the preview top")
	s.Int("hair_style_button_y", &characterCreateHairStyleButtonY, 0, 120, "hair style button offset from the preview top")
	s.Int("content_pad_x", &characterCreateContentPadX, 0, 96, "space at the sides of the create form")
	s.Int("preview_button_inset", &characterCreatePreviewButtonInset, 0, 64, "hair arrow inset from the preview sides")
	s.Float("stat_scale_min", &characterCreateStatScaleMin, 0, 1, "minimum stat polygon scale")
	s.Float("stat_scale_span", &characterCreateStatScaleSpan, 0, 2, "extra stat polygon scale at the stat cap")
	s.Float("stroke_width", &characterCreateBorderWidth, 0, 8, "stat table border and graph line width")
	s.Color("graph_spoke", &characterCreateGraphSpoke, "stat graph spoke color")
	s.Color("graph_fill", &characterCreateGraphFill, "stat polygon fill")
	rotheme.OnLayoutApplied(recomputeCharacterCreateMetrics)
	recomputeCharacterCreateMetrics()
}
