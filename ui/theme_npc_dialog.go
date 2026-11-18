package ui

import (
	"github.com/gogpu/ui/widget"

	"github.com/kivutar/goro/ui/rotheme"
)

// NPC dialog, menu and input geometry and visual-novel colours.
var (
	npcDialogWidth             = 360
	npcDialogHeight            = 260
	npcDialogMinW              = 260
	npcDialogMinH              = 130
	npcDialogPad               = 12
	npcDialogLineH             = 14
	npcDialogLineGap           = 2
	npcDialogButtonW           = 78
	npcDialogRuneW             = 7
	npcDialogMinRunes          = 12
	npcDialogWrapMin           = 8
	npcDialogCharWidth float32 = 0.6
	// Copied from the shared text colours at startup; ui-layout.ini can retint the dialog on its own.
	npcDialogTextColor  = TextColor
	npcDialogMutedColor = MutedTextColor
	npcDialogColorCodeA = 255

	npcMenuWidth   = 260
	npcMenuMinW    = 220
	npcMenuRowH    = 20
	npcMenuPad     = 8
	npcMenuMinRows = 4
	npcMenuMaxRows = 5
	npcPromptGap   = 8

	npcInputTextWidth            = 274
	npcInputTextHeight           = 32
	npcInputNumberWidth          = 174
	npcInputNumberHeight         = 52
	npcInputMinW                 = 140
	npcInputPadX                 = 10
	npcInputPadY                 = 5
	npcInputGap                  = 3
	npcInputRowGap               = 8
	npcInputFieldH               = 22
	npcInputButtonW              = 42
	npcInputRadius       float32 = 5
	npcInputBorderW      float32 = 1
	npcInputBg                   = widget.RGBA8(255, 255, 255, 255)
	npcInputBorder               = widget.RGBA8(193, 198, 194, 255)
)

func init() {
	s := rotheme.Section("npc_dialog", "NPC dialog, menu and number or text prompts")
	s.Int("width", &npcDialogWidth, 260, 800, "NPC dialog width")
	s.Int("height", &npcDialogHeight, 130, 600, "NPC dialog height")
	s.Int("min_width", &npcDialogMinW, 80, 800, "minimum dialog width on a small screen")
	s.Int("min_height", &npcDialogMinH, 40, 600, "minimum dialog height on a small screen")
	s.Int("padding", &npcDialogPad, 0, 32, "NPC dialog text padding")
	s.Int("line_height", &npcDialogLineH, 10, 32, "NPC dialog line height")
	s.Int("line_gap", &npcDialogLineGap, 0, 16, "gap between dialog lines")
	s.Int("button_width", &npcDialogButtonW, 40, 200, "NPC dialog Next/Close button width")
	s.Int("rune_width", &npcDialogRuneW, 1, 32, "estimated pixels per rune when wrapping dialog text")
	s.Int("min_runes", &npcDialogMinRunes, 4, 80, "minimum runes per dialog line")
	s.Int("wrap_min_runes", &npcDialogWrapMin, 1, 40, "smallest wrap width the dialog will use")
	s.Float("char_width", &npcDialogCharWidth, 0.2, 2, "estimated character width as a fraction of text size")
	s.RGBA("text_color", &npcDialogTextColor, "dialog and menu text")
	s.RGBA("muted_color", &npcDialogMutedColor, "waiting text")
	s.Int("color_code_alpha", &npcDialogColorCodeA, 0, 255, "alpha of ^rrggbb visual-novel colour codes")
	s.Int("menu_width", &npcMenuWidth, 220, 600, "NPC menu width")
	s.Int("menu_min_width", &npcMenuMinW, 80, 600, "minimum menu width on a small screen")
	s.Int("menu_row_height", &npcMenuRowH, 14, 40, "NPC menu row height")
	s.Int("menu_padding", &npcMenuPad, 0, 32, "NPC menu padding")
	s.Int("menu_min_rows", &npcMenuMinRows, 1, 20, "minimum menu rows")
	s.Int("menu_max_rows", &npcMenuMaxRows, 1, 24, "visible menu rows before scrolling")
	s.Int("prompt_gap", &npcPromptGap, 0, 64, "gap between the dialog and the menu or input")
	s.Int("input_text_width", &npcInputTextWidth, 80, 800, "text input width")
	s.Int("input_text_height", &npcInputTextHeight, 16, 160, "text input height")
	s.Int("input_number_width", &npcInputNumberWidth, 40, 600, "number input width")
	s.Int("input_number_height", &npcInputNumberHeight, 16, 200, "number input height")
	s.Int("input_min_width", &npcInputMinW, 40, 600, "minimum input width on a small screen")
	s.Int("input_padding_x", &npcInputPadX, 0, 48, "input horizontal padding")
	s.Int("input_padding_y", &npcInputPadY, 0, 32, "input vertical padding")
	s.Int("input_gap", &npcInputGap, 0, 24, "gap between the input label and field")
	s.Int("input_row_gap", &npcInputRowGap, 0, 32, "gap between the input field and OK")
	s.Int("input_field_height", &npcInputFieldH, 12, 64, "input field height")
	s.Int("input_button_width", &npcInputButtonW, 24, 160, "input OK button width")
	s.Float("input_radius", &npcInputRadius, 0, 24, "input corner radius")
	s.Float("input_border_width", &npcInputBorderW, 0, 8, "input border width")
	s.Color("input_background", &npcInputBg, "input background")
	s.Color("input_border", &npcInputBorder, "input border")
}
