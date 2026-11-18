package ui

import "github.com/kivutar/goro/ui/rotheme"

// Escape menu and its death variant.
var (
	escapeMenuWidth        = 252
	escapeMenuHeight       = 200
	escapeMenuReviveHeight = 238
	escapeMenuPad          = 16
	escapeMenuGap          = 8
)

func init() {
	s := rotheme.Section("escape", "escape and death menu")
	s.Int("width", &escapeMenuWidth, 140, 640, "menu width")
	s.Int("height", &escapeMenuHeight, 80, 480, "menu height")
	s.Int("revive_height", &escapeMenuReviveHeight, 80, 560, "menu height when return to life is shown")
	s.Int("padding", &escapeMenuPad, 0, 48, "padding around the menu buttons")
	s.Int("gap", &escapeMenuGap, 0, 32, "gap between menu buttons")
}
