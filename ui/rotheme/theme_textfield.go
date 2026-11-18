package rotheme

import "github.com/gogpu/ui/widget"

// [textfield] defaults: single-line text field.
var (
	textFieldPadX            float32 = 6
	textFieldPadY            float32 = 4
	textFieldShadowRows              = 4
	textFieldShadowInset             = 2
	textFieldShadowAlpha             = 34
	textFieldShadowAlphaStep         = 7
	textFieldShadow                  = widget.RGBA8(82, 108, 138, 255)
)

func init() {
	s := Section("textfield", "single-line text field")
	s.Float("pad_x", &textFieldPadX, 0, 24, "text inset from the left and right")
	s.Float("pad_y", &textFieldPadY, 0, 16, "text and caret inset from the top and bottom")
	s.Int("shadow_rows", &textFieldShadowRows, 0, 16, "inset shadow rows")
	s.Int("shadow_inset", &textFieldShadowInset, 0, 8, "rows kept clear at the bottom of the inset shadow")
	s.Int("shadow_alpha", &textFieldShadowAlpha, 0, 255, "inset shadow alpha of the first row")
	s.Int("shadow_alpha_step", &textFieldShadowAlphaStep, 0, 32, "inset shadow alpha drop per row")
	s.Color("shadow_color", &textFieldShadow, "inset shadow colour")
}
