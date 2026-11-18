package ui

import "github.com/kivutar/goro/ui/rotheme"

// Basic menu defaults. The origin follows the character window and the screen margin.
var (
	basicMenuFollowGap = 2
	basicMenuToggleH   = 10
	basicMenuToggleGap = 2
	basicMenuCols      = 4
	basicMenuRows      = 2
	basicMenuButtonW   = 72
	basicMenuButtonH   = 24
	basicMenuGapX      = 6
	basicMenuGapY      = 5
	basicMenuPad       = 8

	basicMenuX = windowScreenMargin
	// basicMenuY() follows the character window and follow_gap.
)

func init() {
	s := rotheme.Section("basic_menu", "basic info menu under the character window")
	s.Int("follow_gap", &basicMenuFollowGap, 0, 32, "gap below the character window")
	s.Int("toggle_height", &basicMenuToggleH, 6, 40, "collapse toggle height")
	s.Int("toggle_gap", &basicMenuToggleGap, 0, 16, "gap between the toggle and the buttons")
	s.Int("columns", &basicMenuCols, 1, 8, "button columns")
	s.Int("rows", &basicMenuRows, 1, 8, "button rows")
	s.Int("button_width", &basicMenuButtonW, 24, 240, "button width")
	s.Int("button_height", &basicMenuButtonH, 12, 80, "button height")
	s.Int("gap_x", &basicMenuGapX, 0, 32, "horizontal gap between buttons")
	s.Int("gap_y", &basicMenuGapY, 0, 32, "vertical gap between button rows")
	s.Int("padding", &basicMenuPad, 0, 32, "padding around the buttons")
	rotheme.OnLayoutApplied(recomputeBasicMenuOrigin)
}

func recomputeBasicMenuOrigin() {
	basicMenuX = windowScreenMargin
}
