package ui

import "github.com/kivutar/goro/ui/rotheme"

// [skill_grid] defaults: the skill tree grid inside the skill window.
var (
	skillGridColumns              = 7
	skillGridCellW                = 62
	skillGridCellH                = 58
	skillGridMinRows              = 6
	skillGridSlotSize     float32 = 28
	skillGridSlotRadius   float32 = 3
	skillGridSlotTop      float32 = 14
	skillGridOutlineAlpha float32 = 0.25
	skillGridNameH        float32 = 14
	skillGridNameInsetX   float32 = 2
	skillGridLevelInsetY  float32 = 15
	skillGridLevelH       float32 = 14
	skillGridBadgeW       float32 = 16
	skillGridBadgeH       float32 = 14
	skillGridBadgeInsetY  float32 = 14
	skillGridIconInset    float32 = 2
)

func init() {
	s := rotheme.Section("skill_grid", "skill tree grid")
	s.Int("columns", &skillGridColumns, 1, 16, "skill grid columns")
	s.Int("cell_width", &skillGridCellW, 32, 180, "skill grid cell width")
	s.Int("cell_height", &skillGridCellH, 32, 180, "skill grid cell height")
	s.Int("min_rows", &skillGridMinRows, 1, 16, "minimum skill grid rows")
	s.Float("slot_size", &skillGridSlotSize, 12, 80, "skill grid icon slot size")
	s.Float("slot_radius", &skillGridSlotRadius, 0, 16, "skill grid slot corner radius")
	s.Float("slot_top", &skillGridSlotTop, 0, 48, "skill grid slot top inset")
	s.Float("outline_alpha", &skillGridOutlineAlpha, 0, 1, "skill grid slot outline opacity")
	s.Float("name_height", &skillGridNameH, 8, 40, "skill grid name line height")
	s.Float("name_inset_x", &skillGridNameInsetX, 0, 16, "skill grid name horizontal inset")
	s.Float("level_inset_y", &skillGridLevelInsetY, 0, 40, "skill grid level line inset from the bottom")
	s.Float("level_height", &skillGridLevelH, 8, 40, "skill grid level line height")
	s.Float("badge_width", &skillGridBadgeW, 8, 48, "skill requirement badge width")
	s.Float("badge_height", &skillGridBadgeH, 8, 40, "skill requirement badge height")
	s.Float("badge_inset_y", &skillGridBadgeInsetY, 0, 40, "skill requirement badge inset from the slot bottom")
	s.Float("icon_inset", &skillGridIconInset, 0, 16, "skill grid icon inset inside the slot")
}
