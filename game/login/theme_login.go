package login

import (
	"image/color"

	"github.com/kivutar/goro/ui/rotheme"
)

// Login screen backdrop and fade. The account form tokens share this section.
var (
	loginFallbackBG = color.RGBA{R: 10, G: 13, B: 22, A: 255}
	loginFadeColor  = color.RGBA{A: 255}
	loginFadeAlpha  = 255
)

func loginFadeOverlay(alpha uint8) color.RGBA {
	c := loginFadeColor
	c.A = alpha
	return c
}

func init() {
	s := rotheme.Section("login", "login screen and account form")
	s.RGBA("fallback_background", &loginFallbackBG, "color behind the login screen when no background loaded")
	s.RGBA("fade_color", &loginFadeColor, "screen fade color")
	s.Int("fade_alpha", &loginFadeAlpha, 0, 255, "screen fade opacity at full cover")
}
