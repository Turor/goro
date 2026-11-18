package ui

import "github.com/kivutar/goro/ui/rotheme"

// Item book reader.
var (
	bookWindowWidth        = 555
	bookWindowHeight       = 455
	bookWindowPaddingX     = 28
	bookWindowPaddingY     = 12
	bookWindowLineHeight   = 14
	bookWindowLinesPerPage = 25
	bookWindowLineRunes    = 70
	bookWindowNavigationH  = 28
)

func init() {
	s := rotheme.Section("book", "item book reader")
	s.Int("width", &bookWindowWidth, 200, 1200, "book window width")
	s.Int("height", &bookWindowHeight, 200, 1200, "book window height")
	s.Int("padding_x", &bookWindowPaddingX, 0, 80, "book page horizontal padding")
	s.Int("padding_y", &bookWindowPaddingY, 0, 48, "book page vertical padding")
	s.Int("line_height", &bookWindowLineHeight, 8, 40, "book line height")
	s.Int("lines_per_page", &bookWindowLinesPerPage, 1, 80, "book lines per page")
	s.Int("line_runes", &bookWindowLineRunes, 10, 200, "book runes per line")
	s.Int("navigation_height", &bookWindowNavigationH, 16, 80, "previous and next row height")
}
