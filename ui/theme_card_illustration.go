package ui

import "github.com/kivutar/goro/ui/rotheme"

// Card illustration viewer. The window height adds the base title bar.
var (
	cardIllustrationWidth  = 300
	cardIllustrationHeight = 400
)

func init() {
	s := rotheme.Section("card_illustration", "card illustration viewer")
	s.Int("width", &cardIllustrationWidth, 80, 900, "illustration width")
	s.Int("height", &cardIllustrationHeight, 80, 1200, "illustration height")
}
