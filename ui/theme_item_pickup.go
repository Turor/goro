package ui

import (
	"image/color"

	"github.com/kivutar/goro/ui/rotheme"
)

// [item_pickup] defaults: the "item obtained" toast.
var (
	itemPickupNotificationTopY      = 72
	itemPickupNotificationH         = 36
	itemPickupNotificationIconSize  = 24
	itemPickupNotificationIconGap   = 5
	itemPickupNotificationMargin    = 8
	itemPickupNotificationOuter     = 3
	itemPickupNotificationBorderW   = 1
	itemPickupNotificationPadX      = 5
	itemPickupNotificationPadRight  = 14
	itemPickupNotificationTextH     = 20
	itemPickupNotificationTextNudge = 2
	itemPickupNotificationRadius    = 4
	itemPickupNotificationOutline   = color.RGBA{R: 255, G: 255, B: 255, A: 245}
	itemPickupNotificationBorder    = color.RGBA{R: 74, G: 138, B: 202, A: 245}
	itemPickupNotificationFill      = color.RGBA{R: 255, G: 255, B: 255, A: 245}
	itemPickupNotificationText      = color.RGBA{R: 30, G: 34, B: 40, A: 255}
)

func init() {
	s := rotheme.Section("item_pickup", "item obtained toast")
	s.Int("top", &itemPickupNotificationTopY, 0, 300, "distance from the screen top")
	s.Int("height", &itemPickupNotificationH, 12, 120, "toast height")
	s.Int("icon_size", &itemPickupNotificationIconSize, 8, 96, "item icon size")
	s.Int("icon_gap", &itemPickupNotificationIconGap, 0, 32, "gap between the icon and the text")
	s.Int("margin", &itemPickupNotificationMargin, 0, 64, "minimum distance from the screen sides")
	s.Int("outer", &itemPickupNotificationOuter, 0, 16, "outer outline thickness")
	s.Int("border_width", &itemPickupNotificationBorderW, 0, 8, "inner border thickness")
	s.Int("pad_x", &itemPickupNotificationPadX, 0, 32, "content left padding")
	s.Int("pad_right", &itemPickupNotificationPadRight, 0, 64, "content right padding")
	s.Int("text_height", &itemPickupNotificationTextH, 8, 48, "text line height")
	s.Int("text_nudge", &itemPickupNotificationTextNudge, 0, 16, "text vertical nudge")
	s.Int("radius", &itemPickupNotificationRadius, 0, 24, "corner radius")
	s.RGBA("outline", &itemPickupNotificationOutline, "outer outline")
	s.RGBA("border", &itemPickupNotificationBorder, "inner border")
	s.RGBA("fill", &itemPickupNotificationFill, "toast fill")
	s.RGBA("text", &itemPickupNotificationText, "toast text")
}
