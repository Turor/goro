package ui

import "github.com/kivutar/goro/ui/rotheme"

// Window frame sizes. The keys already registered in layout_tokens.go keep
// their names; this file is the defaults table for those literals.
var (
	WindowRadius      float32 = 6
	windowBorderWidth float32 = 1
	windowTitlePadL   float32 = 12

	ROWindowFooterPadding = 10
	ROWindowFooterGap     = 8
	windowTitleButtonGap  = 2
	windowTitleButtonPadR = 7
	windowTitleHeight     = ROWindowTitleHeight
	windowFooterHeight    = ROWindowFooterHeight

	windowScreenMargin    = 8
	windowDirtyPadding    = float32(2)
	windowTitleButtonSize = 17
)

func init() {
	s := rotheme.Section("window", "draggable window frame")
	s.Int("screen_margin", &windowScreenMargin, 0, 64, "gap between a window and the screen edge")
	s.Float("dirty_padding", &windowDirtyPadding, 0, 16, "extra pixels invalidated around a moved window")
	s.Int("title_button_size", &windowTitleButtonSize, 8, 64, "title bar button width")
	rotheme.OnLayoutApplied(recomputeWindowScreenMargins)
}

// Copies of the screen margin live in other windows' position vars. Recompute
// them when the token changes so those positions follow a reload.
func recomputeWindowScreenMargins() {
	characterWindowX = windowScreenMargin
	characterWindowY = windowScreenMargin
	basicMenuX = windowScreenMargin
	basicMenuY = characterWindowY + characterWindowHeight + basicMenuFollowGap
	consoleMargin = windowScreenMargin
	minimapMargin = windowScreenMargin
}
