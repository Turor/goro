package ui

import "github.com/kivutar/goro/ui/rotheme"

// Login form metrics. The screen backdrop colours live in game/login and
// share this section.
var (
	loginWindowW            = 304
	loginWindowFormTopPad   = 18
	loginWindowFormLeftPad  = 16
	loginWindowFormRightPad = 8
	loginWindowFieldGap     = 11
	loginWindowLabelW       = 56
	loginWindowLabelGap     = 12
	loginWindowFieldH       = 22
	loginWindowKeepW        = 64
	loginWindowKeepGap      = 12
	loginWindowBottomPad    = 16
	loginWindowScreenMargin = 8
	loginWindowBiasMinY     = 48
)

func init() {
	s := rotheme.Section("login", "login screen and account form")
	s.Int("width", &loginWindowW, 200, 900, "account form width")
	s.Int("form_top_pad", &loginWindowFormTopPad, 0, 64, "space above the account fields")
	s.Int("form_left_pad", &loginWindowFormLeftPad, 0, 64, "space left of the account fields")
	s.Int("form_right_pad", &loginWindowFormRightPad, 0, 64, "space right of the keep checkbox")
	s.Int("field_gap", &loginWindowFieldGap, 0, 40, "gap between account and password")
	s.Int("label_width", &loginWindowLabelW, 24, 160, "account and password label width")
	s.Int("label_gap", &loginWindowLabelGap, 0, 40, "gap between a label and its field")
	s.Int("field_height", &loginWindowFieldH, 14, 48, "account and password field height")
	s.Int("keep_width", &loginWindowKeepW, 24, 160, "keep-id checkbox width")
	s.Int("keep_gap", &loginWindowKeepGap, 0, 48, "gap between the fields and keep-id")
	s.Int("bottom_pad", &loginWindowBottomPad, 0, 48, "space under the password field")
	s.Int("screen_margin", &loginWindowScreenMargin, 0, 64, "minimum distance from the screen edge")
	s.Int("bias_min_y", &loginWindowBiasMinY, 0, 200, "minimum top when the form sits low on the screen")
}
