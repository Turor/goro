package ui

import (
	"github.com/gogpu/ui/widget"
	"github.com/kivutar/goro/ui/rotheme"
)

// [level_up] defaults: the base and job level-up buttons.
var (
	levelUpNotificationDefaultSize         = 43
	levelUpNotificationBottom              = 0
	levelUpNotificationFill                = widget.RGBA8(70, 126, 189, 255)
	levelUpNotificationBorder              = widget.RGBA8(35, 72, 120, 255)
	levelUpNotificationText                = widget.RGBA8(255, 255, 255, 255)
	levelUpNotificationFont        float32 = 11
)

func init() {
	s := rotheme.Section("level_up", "base and job level-up buttons")
	s.Int("size", &levelUpNotificationDefaultSize, 8, 160, "fallback button size")
	s.Int("bottom", &levelUpNotificationBottom, 0, 200, "distance from the screen bottom")
	s.Color("fill", &levelUpNotificationFill, "fallback button fill")
	s.Color("border", &levelUpNotificationBorder, "fallback button border")
	s.Color("text", &levelUpNotificationText, "fallback button text")
	s.Float("font_size", &levelUpNotificationFont, 6, 32, "fallback button text size")
}
