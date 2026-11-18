package ui

import "github.com/kivutar/goro/ui/rotheme"

// [chat_text] defaults: approximate glyph widths used to wrap chat lines.
var (
	chatRuneSpaceW   = 4
	chatRuneNarrowW  = 3
	chatRuneWideW    = 9
	chatRuneCJKW     = 11
	chatRuneDefaultW = 7
)

func init() {
	s := rotheme.Section("chat_text", "chat text wrapping glyph widths")
	s.Int("space_width", &chatRuneSpaceW, 1, 16, "space and tab width")
	s.Int("narrow_width", &chatRuneNarrowW, 1, 16, "narrow glyph width")
	s.Int("wide_width", &chatRuneWideW, 1, 32, "wide glyph width")
	s.Int("cjk_width", &chatRuneCJKW, 1, 32, "wide east-asian glyph width")
	s.Int("default_width", &chatRuneDefaultW, 1, 32, "default glyph width")
}
