package ui

import "github.com/kivutar/goro/ui/rotheme"

// [amount_prompt] defaults: the number prompt used by shops, mail and drops.
var (
	amountPromptW      = 218
	amountPromptH      = 112
	amountPromptPad    = 10
	amountPromptGap    = 7
	amountPromptFieldH = 24
)

func init() {
	s := rotheme.Section("amount_prompt", "amount entry prompt")
	s.Int("width", &amountPromptW, 140, 640, "prompt width")
	s.Int("height", &amountPromptH, 72, 320, "prompt height")
	s.Int("padding", &amountPromptPad, 0, 40, "content padding")
	s.Int("gap", &amountPromptGap, 0, 32, "gap between the label and the field")
	s.Int("field_height", &amountPromptFieldH, 12, 64, "amount field height")
}
