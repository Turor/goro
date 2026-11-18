package ui

import "github.com/kivutar/goro/ui/rotheme"

// Server announcement banner defaults.
var (
	announcementTop      = 40
	announcementMaxWidth = 500
	announcementPadding  = 10
	announcementFontMin  = 8
	announcementFontMax  = 32
)

func init() {
	s := rotheme.Section("announcement", "server announcement banner")
	s.Int("top", &announcementTop, 0, 400, "distance from the top of the screen")
	s.Int("max_width", &announcementMaxWidth, 80, 1600, "maximum banner width")
	s.Int("padding", &announcementPadding, 0, 64, "horizontal padding inside the banner width")
	s.Int("font_min", &announcementFontMin, 6, 48, "smallest announcement font size")
	s.Int("font_max", &announcementFontMax, 8, 96, "largest announcement font size")
}
