package ui

import "github.com/kivutar/goro/ui/rotheme"

// [character] defaults: the player character window.
var (
	characterWindowWidth                 = 324
	characterWindowHeight                = 140
	characterWindowCompactHeight         = 80
	characterEXPPanelPaddingX    float32 = 6
	characterEXPPanelPaddingY    float32 = 4
	characterEXPPanelGap         float32 = 2
	characterEXPPanelRadius      float32 = 5
	characterEXPLabelWidth       float32 = 76
	characterEXPLabelBarGap      float32 = 6
	characterEXPBarHeight        float32 = 6
	characterTextLineHeight      float32 = 1.2
	characterBodyPadX            float32 = 12
	characterBodyPadY            float32 = 9
	characterVitalsGap           float32 = 4
	characterBodyGap             float32 = 8
	characterStatCellW           float32 = 146
	characterWindowEXPColor              = WindowBorderColor
	characterWindowJobEXPColor           = WindowBorderColor
)

func init() {
	s := rotheme.Section("character", "player character window")
	s.Int("width", &characterWindowWidth, 180, 960, "character window width")
	s.Int("height", &characterWindowHeight, 80, 420, "character window height")
	s.Int("compact_height", &characterWindowCompactHeight, 40, 240, "compact character window height")
	s.Float("exp_panel_padding_x", &characterEXPPanelPaddingX, 0, 24, "exp panel horizontal padding")
	s.Float("exp_panel_padding_y", &characterEXPPanelPaddingY, 0, 24, "exp panel vertical padding")
	s.Float("exp_panel_gap", &characterEXPPanelGap, 0, 16, "gap between base and job exp rows")
	s.Float("exp_panel_radius", &characterEXPPanelRadius, 0, 20, "exp panel corner radius")
	s.Float("exp_label_width", &characterEXPLabelWidth, 32, 240, "exp row label width")
	s.Float("exp_label_gap", &characterEXPLabelBarGap, 0, 24, "gap between the exp label and bar")
	s.Float("exp_bar_height", &characterEXPBarHeight, 2, 24, "exp bar height")
	s.Float("text_line_height", &characterTextLineHeight, 1, 3, "character text line height scale")
	s.Float("body_padding_x", &characterBodyPadX, 0, 40, "character body horizontal padding")
	s.Float("body_padding_y", &characterBodyPadY, 0, 36, "character body vertical padding")
	s.Float("vitals_gap", &characterVitalsGap, 0, 16, "gap between the hp and sp rows")
	s.Float("body_gap", &characterBodyGap, 0, 32, "gap between character body rows")
	s.Float("stat_cell_width", &characterStatCellW, 48, 400, "hp, sp, weight and zeny cell width")
	s.RGBA("exp_color", &characterWindowEXPColor, "base exp bar")
	s.RGBA("job_exp_color", &characterWindowJobEXPColor, "job exp bar")
}
