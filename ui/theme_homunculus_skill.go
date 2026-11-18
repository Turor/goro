package ui

import "github.com/kivutar/goro/ui/rotheme"

var (
	homunculusSkillWindowWidth  = 360
	homunculusSkillWindowHeight = 300
	homunculusSkillColIcon      = 34
	homunculusSkillColType      = 16
	homunculusSkillColName      = 142
	homunculusSkillColLevel     = 40
	homunculusSkillColSP        = 38
	homunculusSkillColRange     = 56
	homunculusSkillColLevelUp   = 22
	homunculusSkillTooltipW     = 292
	homunculusSkillTooltipDX    = 16
	homunculusSkillTooltipDY    = 18
	homunculusSkillPosGapX      = 12
	homunculusSkillPosGapY      = 32
)

func init() {
	s := rotheme.Section("homunculus_skill", "homunculus and mercenary skill windows")
	s.Int("width", &homunculusSkillWindowWidth, 240, 1200, "window width")
	s.Int("height", &homunculusSkillWindowHeight, 160, 1000, "window height")
	s.Int("icon_column_width", &homunculusSkillColIcon, 12, 96, "skill icon column width")
	s.Int("type_column_width", &homunculusSkillColType, 8, 80, "skill type column width")
	s.Int("name_column_width", &homunculusSkillColName, 40, 400, "skill name column width")
	s.Int("level_column_width", &homunculusSkillColLevel, 16, 120, "skill level column width")
	s.Int("sp_column_width", &homunculusSkillColSP, 16, 120, "skill sp column width")
	s.Int("range_column_width", &homunculusSkillColRange, 16, 160, "skill range column width")
	s.Int("levelup_column_width", &homunculusSkillColLevelUp, 12, 80, "level up button column width")
	s.Int("tooltip_width", &homunculusSkillTooltipW, 120, 800, "skill tooltip width")
	s.Int("tooltip_offset_x", &homunculusSkillTooltipDX, 0, 80, "tooltip offset to the right of the pointer")
	s.Int("tooltip_offset_y", &homunculusSkillTooltipDY, 0, 80, "tooltip offset below the pointer")
	s.Int("position_gap_x", &homunculusSkillPosGapX, 0, 80, "gap to the right of the character window")
	s.Int("position_gap_y", &homunculusSkillPosGapY, 0, 160, "offset below the character window top")
}
