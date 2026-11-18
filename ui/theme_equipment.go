package ui

import "github.com/kivutar/goro/ui/rotheme"

// [equipment] defaults: the equipment window and the viewed-equipment pop-up,
// which shares this layout.
var (
	equipmentWindowWidth            = 300
	equipmentWindowPad              = 10
	equipmentLeftColW               = 112
	equipmentCenterColW             = 56
	equipmentRightColW              = 112
	equipmentRowH                   = 24
	equipmentPreviewImageH          = 106
	equipmentFooterCheckW           = 120
	equipmentFooterCheckH           = 20
	equipmentIconPad        float32 = 4
	equipmentTextPad        float32 = 4
	equipmentTextInset      float32 = 8
	equipmentIconTextLead   float32 = 8
	equipmentIconTextTrail  float32 = 12
	equipmentCenterTextTop  float32 = 8
	equipmentTooltipBelow           = 18
	equipmentTooltipAbove           = 6
	equipmentCenterColH     int
	equipmentWindowContentH int
	equipmentWindowHeight   int
)

func recomputeEquipmentWindowMetrics() {
	equipmentCenterColH = equipmentPreviewImageH + equipmentRowH
	equipmentWindowContentH = equipmentCenterColH + equipmentWindowPad*2
	equipmentWindowHeight = ROWindowTitleHeight + equipmentWindowContentH + ROWindowFooterHeight
}

func init() {
	s := rotheme.Section("equipment", "equipment window and viewed equipment pop-up")
	s.Int("width", &equipmentWindowWidth, 180, 900, "equipment window width")
	s.Int("padding", &equipmentWindowPad, 0, 40, "equipment window padding")
	s.Int("left_column_width", &equipmentLeftColW, 48, 360, "left equipment column width")
	s.Int("center_column_width", &equipmentCenterColW, 32, 200, "center equipment column width")
	s.Int("right_column_width", &equipmentRightColW, 48, 360, "right equipment column width")
	s.Int("row_height", &equipmentRowH, 16, 72, "equipment slot row height")
	s.Int("preview_height", &equipmentPreviewImageH, 40, 360, "character preview height")
	s.Int("footer_check_width", &equipmentFooterCheckW, 40, 360, "show equip checkbox width")
	s.Int("footer_check_height", &equipmentFooterCheckH, 12, 64, "show equip checkbox height")
	s.Float("icon_pad", &equipmentIconPad, 0, 16, "equipment slot icon inset")
	s.Float("text_pad", &equipmentTextPad, 0, 16, "equipment slot text inset")
	s.Float("text_inset", &equipmentTextInset, 0, 32, "equipment slot text margin")
	s.Float("icon_text_lead", &equipmentIconTextLead, 0, 32, "space before the label when an icon is on the left")
	s.Float("icon_text_trail", &equipmentIconTextTrail, 0, 40, "space kept clear of the icon when laying out the label")
	s.Float("center_text_top", &equipmentCenterTextTop, 0, 32, "ammo label top inset")
	s.Int("tooltip_below", &equipmentTooltipBelow, 0, 64, "tooltip offset below the cursor")
	s.Int("tooltip_above", &equipmentTooltipAbove, 0, 48, "tooltip offset above the cursor")
	rotheme.OnLayoutApplied(recomputeEquipmentWindowMetrics)
	recomputeEquipmentWindowMetrics()
}
