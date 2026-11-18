package ui

import (
	"github.com/gogpu/ui/widget"
	"github.com/kivutar/goro/ui/rotheme"
)

// Character select window. Preview sprite scale lives in game/login and shares
// this section.
var (
	characterSelectWindowW               = 576
	characterSelectBaseH                 = 356
	characterSelectInfoRowCount          = 6
	characterSelectInfoRowH              = 16
	characterSelectInfoTableGap          = 1
	characterSelectInfoPanelPad          = 2
	characterSelectInfoPanelW            = 318
	characterSelectSlotW                 = 139
	characterSelectSlotH                 = 144
	characterSelectArrowGap              = 13
	characterSelectSlotGap               = 25
	characterSelectPageSlots             = 3
	characterSelectContentPadTop         = 12
	characterSelectContentPadX           = 24
	characterSelectContentGap            = 9
	characterSelectEmptyPad              = 12
	characterSelectEmptyGap              = 8
	characterSelectInfoLabelW    float32 = 44
	characterSelectStatLabelW    float32 = 32
	characterSelectBorderWidth   float32 = 1
	characterSelectSelectedBG            = widget.RGBA8(222, 237, 252, 255)

	characterSelectInfoTableH  int
	characterSelectInfoPanelH  int
	characterSelectWindowH     int
	characterSelectInfoColumnW float32
	characterSelectInfoValueW  float32
	characterSelectStatValueW  float32
)

func recomputeCharacterSelectMetrics() {
	characterSelectInfoTableH = characterSelectInfoRowCount*characterSelectInfoRowH + (characterSelectInfoRowCount-1)*characterSelectInfoTableGap
	characterSelectInfoPanelH = characterSelectInfoTableH + 2*characterSelectInfoPanelPad
	characterSelectWindowH = characterSelectBaseH + characterSelectInfoRowH + characterSelectInfoTableGap
	inner := float32(characterSelectInfoPanelW - 2*characterSelectInfoPanelPad)
	characterSelectInfoColumnW = (inner - rotheme.TableGap) / 2
	characterSelectInfoValueW = characterSelectInfoColumnW - characterSelectInfoLabelW - rotheme.TableGap
	characterSelectStatValueW = characterSelectInfoColumnW - characterSelectStatLabelW - rotheme.TableGap
}

func init() {
	s := rotheme.Section("char_select", "character select screen")
	s.Int("width", &characterSelectWindowW, 320, 1400, "character select width")
	s.Int("base_height", &characterSelectBaseH, 180, 900, "window height before the info row grows")
	s.Int("info_row_count", &characterSelectInfoRowCount, 1, 12, "character info rows")
	s.Int("info_row_height", &characterSelectInfoRowH, 10, 40, "character info row height")
	s.Int("info_table_gap", &characterSelectInfoTableGap, 0, 12, "gap between character info rows")
	s.Int("info_panel_pad", &characterSelectInfoPanelPad, 0, 24, "padding inside the character info panel")
	s.Int("info_panel_width", &characterSelectInfoPanelW, 160, 800, "character info panel width")
	s.Int("slot_width", &characterSelectSlotW, 64, 320, "character slot width")
	s.Int("slot_height", &characterSelectSlotH, 64, 360, "character slot height")
	s.Int("arrow_gap", &characterSelectArrowGap, 0, 48, "gap between page arrows and slots")
	s.Int("slot_gap", &characterSelectSlotGap, 0, 64, "gap between character slots")
	s.Int("page_slots", &characterSelectPageSlots, 1, 12, "character slots shown on one page")
	s.Int("content_pad_top", &characterSelectContentPadTop, 0, 64, "space above the slot row")
	s.Int("content_pad_x", &characterSelectContentPadX, 0, 80, "space at the sides of the slot row")
	s.Int("content_gap", &characterSelectContentGap, 0, 40, "gap between the slot row, page number and info panel")
	s.Int("empty_padding", &characterSelectEmptyPad, 0, 48, "padding of an empty slot info panel")
	s.Int("empty_gap", &characterSelectEmptyGap, 0, 32, "gap between empty slot info lines")
	s.Float("info_label_width", &characterSelectInfoLabelW, 16, 160, "name and job label column width")
	s.Float("stat_label_width", &characterSelectStatLabelW, 16, 120, "stat label column width")
	s.Float("border_width", &characterSelectBorderWidth, 0, 8, "slot and info panel border width")
	s.Color("selected_background", &characterSelectSelectedBG, "selected slot and info header background")
	rotheme.OnLayoutApplied(recomputeCharacterSelectMetrics)
	recomputeCharacterSelectMetrics()
}
