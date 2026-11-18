package ui

import "github.com/kivutar/goro/ui/rotheme"

// Confirm and alert dialogs. height is the fixed size the shop deal window
// still borrows; the dialog itself grows with its message.
var (
	smallPromptWidth        = 286
	smallPromptHeight       = 128
	smallPromptContentH     = 42
	smallPromptSidePad      = 12
	smallPromptLineH        = 14
	smallPromptDefaultLines = 2
	smallPromptMaxLines     = 4
	smallPromptLineMaxRunes = 42
)

func init() {
	s := rotheme.Section("confirm", "confirm and alert dialog")
	s.Int("width", &smallPromptWidth, 160, 800, "dialog width")
	s.Int("height", &smallPromptHeight, 80, 400, "fixed height borrowed by the shop deal window")
	s.Int("content_height", &smallPromptContentH, 16, 160, "message area height for one extra line")
	s.Int("side_pad", &smallPromptSidePad, 0, 48, "message left and right padding")
	s.Int("line_height", &smallPromptLineH, 8, 40, "message line height")
	s.Int("default_lines", &smallPromptDefaultLines, 1, 8, "message lines when the text is short")
	s.Int("max_lines", &smallPromptMaxLines, 1, 12, "most message lines shown")
	s.Int("line_max_runes", &smallPromptLineMaxRunes, 8, 120, "runes per wrapped message line")
}
