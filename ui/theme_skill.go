package ui

import (
	"github.com/gogpu/ui/widget"
	"github.com/kivutar/goro/ui/rotheme"
)

// [skill] defaults: the skill list window. Grid cell metrics live in [skill_grid].
var (
	skillTabW         = 32
	skillTabH         = 44
	skillTableViewW   = 396
	skillWindowHeight = 388
	skillRowH         = 32
	skillIconSize     = 24
	skillHeaderH      = 24

	skillColIconW      float32 = 34
	skillColTypeW      float32 = 16
	skillColNameW      float32 = 142
	skillColLevelDownW float32 = 18
	skillColLevelW     float32 = 40
	skillColLevelSelW  float32 = 18
	skillColSPW        float32 = 38
	skillColRangeW     float32 = 56
	skillColLevelUpW   float32 = 22

	skillTooltipW  = 292
	skillTooltipDX = 16
	skillTooltipDY = 18

	skillPassiveColor = widget.RGBA8(34, 142, 158, 255)
	skillActiveColor  = widget.RGBA8(44, 92, 184, 255)

	skillTabRailW         int
	skillWindowWidth      int
	skillTableViewH       int
	skillTableBodyH       int
	skillGridViewW        int
	skillGridViewH        int
	skillGridWindowWidth  int
	skillGridWindowHeight int
	skillTableColumns     []rotheme.TableViewColumn
)

func recomputeSkillWindowMetrics() {
	skillTabRailW = skillTabW + skillTabOver*2
	skillWindowWidth = skillTabRailW + verticalTabDividerW + skillTableViewW
	skillTableViewH = skillWindowHeight - ROWindowTitleHeight - ROWindowFooterHeight
	skillTableBodyH = skillTableViewH - skillHeaderH
	skillGridViewW = skillGridColumns*skillGridCellW + ROScrollbarGutter
	skillGridViewH = skillGridMinRows * skillGridCellH
	skillGridWindowWidth = skillTabRailW + verticalTabDividerW + skillGridViewW
	skillGridWindowHeight = ROWindowTitleHeight + skillGridViewH + ROWindowFooterHeight
	skillTableColumns = []rotheme.TableViewColumn{
		{Key: "icon", Width: skillColIconW},
		{Key: "type", Width: skillColTypeW},
		{Key: "name", Title: "Name", Width: skillColNameW},
		{Key: "leveldown", Width: skillColLevelDownW},
		{Key: "level", Title: "Lv", Width: skillColLevelW, Align: widget.TextAlignCenter},
		{Key: "levelupselect", Width: skillColLevelSelW},
		{Key: "sp", Title: "SP", Width: skillColSPW},
		{Key: "range", Title: "Range", Width: skillColRangeW},
		{Key: "levelup", Width: skillColLevelUpW},
		{Key: "fill", Flex: 1},
	}
}

func init() {
	s := rotheme.Section("skill", "skill list window")
	s.Int("tab_width", &skillTabW, 16, 96, "skill tab width")
	s.Int("tab_height", &skillTabH, 20, 120, "skill tab height")
	s.Int("table_width", &skillTableViewW, 200, 1200, "skill table width")
	s.Int("height", &skillWindowHeight, 160, 1200, "skill list window height")
	s.Int("row_height", &skillRowH, 16, 80, "skill row height")
	s.Int("icon_size", &skillIconSize, 12, 64, "skill icon size")
	s.Int("header_height", &skillHeaderH, 12, 64, "skill table header height")
	s.Float("column_icon_width", &skillColIconW, 12, 96, "skill icon column width")
	s.Float("column_type_width", &skillColTypeW, 8, 64, "skill type column width")
	s.Float("column_name_width", &skillColNameW, 40, 400, "skill name column width")
	s.Float("column_level_down_width", &skillColLevelDownW, 8, 64, "skill level down column width")
	s.Float("column_level_width", &skillColLevelW, 16, 120, "skill level column width")
	s.Float("column_level_select_width", &skillColLevelSelW, 8, 64, "skill level select column width")
	s.Float("column_sp_width", &skillColSPW, 16, 120, "skill sp column width")
	s.Float("column_range_width", &skillColRangeW, 16, 160, "skill range column width")
	s.Float("column_level_up_width", &skillColLevelUpW, 8, 80, "skill level up column width")
	s.Int("tooltip_width", &skillTooltipW, 120, 800, "skill tooltip width")
	s.Int("tooltip_offset_x", &skillTooltipDX, 0, 64, "skill tooltip offset to the right of the cursor")
	s.Int("tooltip_offset_y", &skillTooltipDY, 0, 64, "skill tooltip offset below the cursor")
	s.Color("passive_color", &skillPassiveColor, "passive skill type color")
	s.Color("active_color", &skillActiveColor, "active skill type color")
	rotheme.OnLayoutApplied(recomputeSkillWindowMetrics)
	recomputeSkillWindowMetrics()
}
