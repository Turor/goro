package rotheme

import "github.com/gogpu/ui/widget"

// Button chrome. radius, padding_x, padding_y, large_padding_y and reflect_alpha
// stay registered in theme_tokens.go; registering them again panics.
var (
	ButtonRadius        float32 = 6
	ButtonPaddingX      float32 = 8
	ButtonPaddingY      float32 = 5.5
	LargeButtonPaddingY float32 = 8.5
	buttonReflectInset  float32 = 3
	buttonReflectAlpha  float32 = 0.4
	buttonGradientMix   float32 = 0.42
	buttonShadowOffsetY float32 = 2
	buttonHighlight             = widget.RGBA(1, 1, 1, 1)
	buttonShadowFar             = widget.RGBA(0.4, 0.4, 0.4, 0.15)
	buttonShadowNear            = widget.RGBA(0.4, 0.4, 0.4, 0.35)
)

func init() {
	s := Section("button", "standard button chrome")
	s.Float("reflect_inset", &buttonReflectInset, 0, 16, "glossy highlight inset")
	s.Float("gradient_mix", &buttonGradientMix, 0, 1, "gradient blend toward the highlight")
	s.Float("shadow_offset_y", &buttonShadowOffsetY, 0, 8, "softer drop shadow offset")
	s.Color("highlight", &buttonHighlight, "gradient and glossy highlight tint")
	s.Color("shadow_far", &buttonShadowFar, "softer drop shadow")
	s.Color("shadow_near", &buttonShadowNear, "tighter drop shadow")
}
