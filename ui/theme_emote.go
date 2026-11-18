package ui

import "github.com/kivutar/goro/ui/rotheme"

// Emotion icon list.
var (
	emoteWindowWidth          = 250
	emoteWindowHeight         = 330
	emoteCols                 = 5
	emoteRows                 = 6
	emoteCellSize             = 40
	emoteCellGap              = 4
	emoteContentPadX          = 10
	emoteContentPadY          = 6
	emoteSectionGap           = 4
	emoteNavHeight            = 24
	emoteNavButtonW           = 48
	emoteNavLabelW            = 56
	emoteNavControlH          = 22
	emoteNavLabelPad          = 4
	emoteNavGap               = 8
	emoteDefaultX             = 600
	emoteDefaultY             = 200
	emoteHoverRadius  float32 = 5
	emoteHoverBorder  float32 = 1
)

func init() {
	s := rotheme.Section("emote", "emotion icon list")
	s.Int("width", &emoteWindowWidth, 120, 800, "emote window width")
	s.Int("height", &emoteWindowHeight, 160, 900, "emote window height")
	s.Int("columns", &emoteCols, 1, 12, "emote columns")
	s.Int("rows", &emoteRows, 1, 12, "emote rows per page")
	s.Int("cell_size", &emoteCellSize, 16, 96, "emote cell size")
	s.Int("cell_gap", &emoteCellGap, 0, 24, "gap between emote cells")
	s.Int("padding_x", &emoteContentPadX, 0, 48, "emote content horizontal padding")
	s.Int("padding_y", &emoteContentPadY, 0, 32, "emote content vertical padding")
	s.Int("section_gap", &emoteSectionGap, 0, 24, "gap between the grid and the pager")
	s.Int("nav_height", &emoteNavHeight, 16, 64, "pager row height")
	s.Int("nav_button_width", &emoteNavButtonW, 24, 160, "previous and next button width")
	s.Int("nav_label_width", &emoteNavLabelW, 24, 200, "page counter width")
	s.Int("nav_control_height", &emoteNavControlH, 12, 64, "pager button and counter height")
	s.Int("nav_label_pad", &emoteNavLabelPad, 0, 24, "page counter top padding")
	s.Int("nav_gap", &emoteNavGap, 0, 32, "gap between pager controls")
	s.Int("default_x", &emoteDefaultX, 0, 4000, "default distance from the left screen edge")
	s.Int("default_y", &emoteDefaultY, 0, 4000, "default distance from the top screen edge")
	s.Float("hover_radius", &emoteHoverRadius, 0, 24, "hovered cell corner radius")
	s.Float("hover_border", &emoteHoverBorder, 0, 8, "hovered cell border width")
}
