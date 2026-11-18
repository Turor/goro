package ui

import (
	"image/color"

	"github.com/kivutar/goro/ui/rotheme"
)

var (
	WindowBorderColor = color.RGBA{R: 118, G: 160, B: 206, A: 255}
	PanelAltColor     = color.RGBA{R: 236, G: 244, B: 252, A: 255}
	TextColor         = color.RGBA{R: 38, G: 48, B: 58, A: 255}
	MutedTextColor    = color.RGBA{R: 98, G: 112, B: 126, A: 255}
	TitleTextColor    = color.RGBA{R: 22, G: 54, B: 88, A: 255}
	GoodTextColor     = color.RGBA{R: 34, G: 142, B: 72, A: 255}
	ErrorTextColor    = color.RGBA{R: 204, G: 48, B: 48, A: 255}
	PlayerHPBarColor  = color.RGBA{R: 16, G: 239, B: 33, A: 255}
	PlayerSPBarColor  = color.RGBA{R: 24, G: 99, B: 222, A: 255}
	HungerBarColor    = color.RGBA{R: 255, G: 231, B: 231, A: 255}
	HungerBarLowColor = color.RGBA{R: 255, G: 255, B: 0, A: 255}

	HungerBarLowThreshold float32 = 0.25
)

func init() {
	s := rotheme.Section("surface", "rounded surfaces and shared text colors")
	s.RGBA("border", &WindowBorderColor, "shared border color")
	s.RGBA("text", &TextColor, "body text color")
	s.RGBA("muted_text", &MutedTextColor, "secondary text color")
	s.RGBA("title_text", &TitleTextColor, "title text color")
	s.Float("hunger_low_threshold", &HungerBarLowThreshold, 0, 1, "hunger ratio that switches to the low color")
}
