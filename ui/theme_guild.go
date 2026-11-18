package ui

import (
	"github.com/gogpu/ui/widget"
	"github.com/kivutar/goro/ui/rotheme"
)

// [guild] defaults: guild info, members, positions, skills, history and notice.
var (
	guildWindowWidth     = 400
	guildWindowHeight    = 325
	guildWindowTabHeight = 23
	guildWindowTabWidth  = 68
	guildTabStripGap     = -1
	guildContentGap      = 2
	guildEmblemSize      = 24
	guildPortraitSize    = 24
	guildTableRowH       = 32
	guildTableHeaderH    = 24
	guildTableControlH   = 22
	guildCellPadLeft     = 6
	guildPortraitPadLeft = 2
	guildMemberDropItemH = 22
	guildMemberDropRows  = 8
	guildEmblemDropRows  = 5
	guildNoticeFieldH    = 24
	guildNoticeTitleH    = 28
	guildNoticePadX      = 9
	guildNoticePadY      = 10
	guildNoticeGap       = 5
	guildNoticeBoxPadX   = 5
	guildNoticeBoxPadY   = 4
	guildNoticeLineH     = 16
	guildInfoPadX        = 9
	guildInfoPadY        = 10
	guildInfoColGap      = 10
	guildInfoRowGap      = 3
	guildInfoLeftW       = 188
	guildInfoRightW      = 174
	guildInfoRowH        = 16
	guildEmblemEditorW   = 132
	guildEmblemEditorH   = 22
	guildEmblemEditorGap = 6
	guildTendencyLabelW  = 16
	guildTendencyTrackH  = 90
	guildTendencyGap     = 4
	guildTendencyPadL    = 3
	guildTendencyPadT    = 3
	guildTendencyCaption = 19
	guildTendencyBoxGap  = 2
	guildRelationRows    = 3
	guildRelationW       = 168
	guildRelationH       = 42
	guildRelationPadX    = 3
	guildRelationPadY    = 1
	guildRelationRowW    = 160
	guildRelationRowH    = 13
	guildRelationTextX   = 2
	guildRelationTextW   = 4
	guildFramePad        = 4
	guildPlaceholderPad  = 10
	guildTooltipW        = 292
	guildTooltipOffX     = 16
	guildTooltipOffY     = 18
	guildRowTipOffX      = 150
	guildRowTipOffY      = 18
	guildRowTipAbove     = 6
	guildRowTipW         = 300

	guildMemberPortraitW = 28
	guildMemberNameW     = 78
	guildMemberPosW      = 62
	guildMemberJobW      = 52
	guildMemberLevelW    = 26
	guildMemberNoteW     = 62
	guildMemberDevotionW = 38
	guildMemberTaxW      = 42
	guildPosRankW        = 44
	guildPosTitleW       = 164
	guildPosInviteW      = 58
	guildPosPunishW      = 48
	guildPosTaxW         = 46
	guildSkillIconW      = 34
	guildSkillTypeW      = 16
	guildSkillNameW      = 234
	guildSkillLevelW     = 40
	guildSkillLevelUpW   = 22
	guildHistoryNameW    = 116

	guildTabFrameBg = widget.RGBA8(255, 255, 255, 255)
)

func init() {
	s := rotheme.Section("guild", "guild info, members, positions, skills, history and notice")
	s.Int("width", &guildWindowWidth, 240, 1200, "guild window width")
	s.Int("height", &guildWindowHeight, 180, 1000, "guild window height")
	s.Int("tab_height", &guildWindowTabHeight, 14, 64, "guild tab height")
	s.Int("tab_width", &guildWindowTabWidth, 32, 200, "guild tab width")
	s.Int("tab_strip_gap", &guildTabStripGap, -16, 16, "gap between guild tabs")
	s.Int("content_gap", &guildContentGap, 0, 16, "gap under the guild tab strip")
	s.Int("emblem_size", &guildEmblemSize, 8, 96, "guild emblem preview size")
	s.Int("portrait_size", &guildPortraitSize, 8, 96, "member portrait size")
	s.Int("table_row_height", &guildTableRowH, 16, 80, "guild table row height")
	s.Int("table_header_height", &guildTableHeaderH, 12, 64, "guild table header height")
	s.Int("table_control_height", &guildTableControlH, 12, 64, "guild table editor height")
	s.Int("cell_padding_left", &guildCellPadLeft, 0, 24, "guild table text inset")
	s.Int("portrait_padding_left", &guildPortraitPadLeft, 0, 24, "member portrait inset")
	s.Int("member_dropdown_item_height", &guildMemberDropItemH, 12, 48, "member position menu row height")
	s.Int("member_dropdown_visible", &guildMemberDropRows, 1, 24, "member position menu visible rows")
	s.Int("emblem_dropdown_visible", &guildEmblemDropRows, 1, 24, "emblem menu visible rows")
	s.Int("notice_field_height", &guildNoticeFieldH, 14, 64, "notice title field height")
	s.Int("notice_title_height", &guildNoticeTitleH, 14, 80, "read-only notice title height")
	s.Int("notice_padding_x", &guildNoticePadX, 0, 32, "notice tab horizontal padding")
	s.Int("notice_padding_y", &guildNoticePadY, 0, 32, "notice tab vertical padding")
	s.Int("notice_gap", &guildNoticeGap, 0, 24, "notice tab row gap")
	s.Int("notice_box_padding_x", &guildNoticeBoxPadX, 0, 24, "notice title box horizontal padding")
	s.Int("notice_box_padding_y", &guildNoticeBoxPadY, 0, 24, "notice title box vertical padding")
	s.Int("notice_line_height", &guildNoticeLineH, 8, 40, "notice title line height")
	s.Int("info_padding_x", &guildInfoPadX, 0, 32, "info tab horizontal padding")
	s.Int("info_padding_y", &guildInfoPadY, 0, 32, "info tab vertical padding")
	s.Int("info_column_gap", &guildInfoColGap, 0, 48, "gap between info columns")
	s.Int("info_row_gap", &guildInfoRowGap, 0, 24, "info column row gap")
	s.Int("info_left_width", &guildInfoLeftW, 80, 500, "info left column width")
	s.Int("info_right_width", &guildInfoRightW, 80, 500, "info right column width")
	s.Int("info_row_height", &guildInfoRowH, 10, 40, "info row height")
	s.Int("emblem_editor_width", &guildEmblemEditorW, 40, 400, "emblem dropdown width")
	s.Int("emblem_editor_height", &guildEmblemEditorH, 12, 48, "emblem dropdown height")
	s.Int("emblem_editor_gap", &guildEmblemEditorGap, 0, 24, "gap between emblem and dropdown")
	s.Int("tendency_label_width", &guildTendencyLabelW, 8, 48, "tendency compass label width")
	s.Int("tendency_track_height", &guildTendencyTrackH, 24, 240, "tendency compass track size")
	s.Int("tendency_gap", &guildTendencyGap, 0, 24, "gap inside the tendency compass")
	s.Int("tendency_padding_left", &guildTendencyPadL, 0, 16, "tendency label left padding")
	s.Int("tendency_padding_top", &guildTendencyPadT, 0, 16, "tendency caption top padding")
	s.Int("tendency_caption_height", &guildTendencyCaption, 8, 48, "tendency caption height")
	s.Int("tendency_box_gap", &guildTendencyBoxGap, 0, 16, "gap around the tendency compass")
	s.Int("relation_rows", &guildRelationRows, 1, 12, "alliance names shown")
	s.Int("relation_width", &guildRelationW, 60, 400, "alliance list width")
	s.Int("relation_height", &guildRelationH, 16, 160, "alliance list height")
	s.Int("relation_padding_x", &guildRelationPadX, 0, 16, "alliance list horizontal padding")
	s.Int("relation_padding_y", &guildRelationPadY, 0, 16, "alliance list vertical padding")
	s.Int("relation_row_width", &guildRelationRowW, 40, 400, "alliance row width")
	s.Int("relation_row_height", &guildRelationRowH, 8, 40, "alliance row height")
	s.Int("relation_text_x", &guildRelationTextX, 0, 16, "alliance name left inset")
	s.Int("relation_text_inset", &guildRelationTextW, 0, 24, "alliance name width inset")
	s.Int("frame_padding", &guildFramePad, 0, 16, "framed box padding")
	s.Int("placeholder_padding", &guildPlaceholderPad, 0, 32, "empty tab padding")
	s.Int("tooltip_width", &guildTooltipW, 80, 800, "skill tooltip width")
	s.Int("tooltip_offset_x", &guildTooltipOffX, 0, 64, "skill tooltip horizontal offset")
	s.Int("tooltip_offset_y", &guildTooltipOffY, 0, 64, "skill tooltip vertical offset")
	s.Int("row_tooltip_offset_x", &guildRowTipOffX, 0, 400, "row tooltip horizontal offset")
	s.Int("row_tooltip_offset_y", &guildRowTipOffY, 0, 64, "row tooltip vertical offset")
	s.Int("row_tooltip_above", &guildRowTipAbove, 0, 64, "row tooltip above-cursor offset")
	s.Int("row_tooltip_width", &guildRowTipW, 80, 800, "row tooltip width")
	s.Int("member_portrait_width", &guildMemberPortraitW, 12, 80, "member portrait column width")
	s.Int("member_name_width", &guildMemberNameW, 24, 240, "member name column width")
	s.Int("member_position_width", &guildMemberPosW, 24, 200, "member position column width")
	s.Int("member_job_width", &guildMemberJobW, 24, 200, "member job column width")
	s.Int("member_level_width", &guildMemberLevelW, 16, 80, "member level column width")
	s.Int("member_note_width", &guildMemberNoteW, 24, 200, "member note column width")
	s.Int("member_devotion_width", &guildMemberDevotionW, 16, 120, "member devotion column width")
	s.Int("member_tax_width", &guildMemberTaxW, 16, 120, "member tax column width")
	s.Int("position_rank_width", &guildPosRankW, 16, 120, "position rank column width")
	s.Int("position_title_width", &guildPosTitleW, 40, 400, "position title column width")
	s.Int("position_invite_width", &guildPosInviteW, 24, 160, "position invite column width")
	s.Int("position_punish_width", &guildPosPunishW, 24, 160, "position punish column width")
	s.Int("position_tax_width", &guildPosTaxW, 16, 120, "position tax column width")
	s.Int("skill_icon_width", &guildSkillIconW, 12, 80, "skill icon column width")
	s.Int("skill_type_width", &guildSkillTypeW, 8, 64, "skill type column width")
	s.Int("skill_name_width", &guildSkillNameW, 40, 500, "skill name column width")
	s.Int("skill_level_width", &guildSkillLevelW, 16, 120, "skill level column width")
	s.Int("skill_levelup_width", &guildSkillLevelUpW, 12, 64, "skill level-up column width")
	s.Int("history_name_width", &guildHistoryNameW, 40, 400, "expel history name column width")
	s.Color("tab_frame_background", &guildTabFrameBg, "guild tab strip background")
}
