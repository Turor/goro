package ui

import "github.com/kivutar/goro/ui/rotheme"

// Generic text and password prompt.
var (
	textPromptW        = 306
	textPromptContentH = 70
	textPromptFieldH   = 24
	textPromptPad      = 14
	textPromptGap      = 8
)

func textPromptHeight() int {
	return ROWindowTitleHeight + textPromptContentH + ROWindowFooterHeight
}

func init() {
	s := rotheme.Section("text_prompt", "text and password prompt")
	s.Int("width", &textPromptW, 180, 800, "prompt width")
	s.Int("content_height", &textPromptContentH, 32, 240, "prompt content height")
	s.Int("field_height", &textPromptFieldH, 14, 48, "text field height")
	s.Int("padding", &textPromptPad, 0, 48, "padding around the label and field")
	s.Int("gap", &textPromptGap, 0, 32, "gap between the label and the field")
}
