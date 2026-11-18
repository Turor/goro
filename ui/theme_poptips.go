package ui

import (
	"image/color"

	"github.com/kivutar/goro/ui/rotheme"
)

// Poptip stack defaults. The solid alpha is the colour's alpha before fading.
var (
	poptipMaxItems   = 2
	poptipBaseY      = 90
	poptipPadding    = 4
	poptipLineH      = 20
	poptipTextNudge  = 2
	poptipBackground = color.RGBA{A: 204}
	poptipText       = color.RGBA{R: 255, G: 255, B: 255, A: 255}
)

func init() {
	s := rotheme.Section("poptips", "short server messages near the top of the game view")
	s.Int("max_items", &poptipMaxItems, 1, 8, "poptips shown at once")
	s.Int("base_y", &poptipBaseY, 0, 400, "vertical center of the newest poptip")
	s.Int("padding", &poptipPadding, 0, 24, "padding around the poptip text")
	s.Int("line_height", &poptipLineH, 8, 64, "distance between stacked poptips")
	s.Int("text_nudge", &poptipTextNudge, 0, 16, "text offset inside the poptip box")
	s.RGBA("background", &poptipBackground, "poptip background")
	s.RGBA("text_color", &poptipText, "poptip text")
}
