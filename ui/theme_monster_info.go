package ui

import "github.com/kivutar/goro/ui/rotheme"

var (
	monsterInfoWindowW       = 400
	monsterInfoWindowH       = 250
	monsterInfoPreviewW      = 110
	monsterInfoPreviewH      = 160
	monsterInfoStatsW        = 260
	monsterInfoRowH          = 19
	monsterInfoElementH      = 23
	monsterInfoPadX          = 10
	monsterInfoPadY          = 5
	monsterInfoContentGap    = 10
	monsterInfoStatsGap      = 7
	monsterInfoNameLabelW    = 38
	monsterInfoNameValueW    = 92
	monsterInfoExtraLabelW   = 50
	monsterInfoExtraValueW   = 80
	monsterInfoElementGap    = 4
	monsterInfoElementRowGap = 3
	monsterInfoElementW      = 84
	monsterInfoElementRadius = 3
	monsterInfoElementBorder = 1
)

func init() {
	s := rotheme.Section("monster_info", "monster information window")
	s.Int("width", &monsterInfoWindowW, 200, 1200, "window width")
	s.Int("height", &monsterInfoWindowH, 120, 800, "window height")
	s.Int("preview_width", &monsterInfoPreviewW, 40, 480, "preview panel width")
	s.Int("preview_height", &monsterInfoPreviewH, 40, 640, "preview panel height")
	s.Int("stats_width", &monsterInfoStatsW, 80, 800, "stats column width")
	s.Int("row_height", &monsterInfoRowH, 12, 64, "stat row height")
	s.Int("element_height", &monsterInfoElementH, 12, 72, "element cell height")
	s.Int("padding_x", &monsterInfoPadX, 0, 48, "content horizontal padding")
	s.Int("padding_y", &monsterInfoPadY, 0, 48, "content vertical padding")
	s.Int("content_gap", &monsterInfoContentGap, 0, 48, "gap between preview and stats")
	s.Int("stats_gap", &monsterInfoStatsGap, 0, 32, "gap between stats and elements")
	s.Int("name_label_width", &monsterInfoNameLabelW, 16, 160, "left stat label width")
	s.Int("name_value_width", &monsterInfoNameValueW, 24, 280, "left stat value width")
	s.Int("extra_label_width", &monsterInfoExtraLabelW, 16, 200, "right stat label width")
	s.Int("extra_value_width", &monsterInfoExtraValueW, 24, 240, "right stat value width")
	s.Int("element_gap", &monsterInfoElementGap, 0, 24, "gap between element cells")
	s.Int("element_row_gap", &monsterInfoElementRowGap, 0, 24, "gap between element rows")
	s.Int("element_width", &monsterInfoElementW, 32, 280, "element cell width")
	s.Int("element_radius", &monsterInfoElementRadius, 0, 16, "element cell corner radius")
	s.Int("element_border_width", &monsterInfoElementBorder, 0, 8, "element cell border width")
}
