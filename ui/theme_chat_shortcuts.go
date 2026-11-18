package ui

import "github.com/kivutar/goro/ui/rotheme"

// [chat_shortcuts] defaults: Alt+1 through Alt+0 shortcut editor.
var (
	chatShortcutsW        = 280
	chatShortcutsRows     = 10
	chatShortcutsRowH     = 24
	chatShortcutsRowGap   = 4
	chatShortcutsPad      = 12
	chatShortcutsLabelW   = 54
	chatShortcutsLabelGap = 8
)

// chatShortcutsH is the window height for the current row, gap and padding tokens.
// Ten rows of 24 with a gap of 4 and padding of 12 are the same 300px content
// the old 10*28+20 formula produced.
func chatShortcutsH() int {
	rows := chatShortcutsRows
	if rows < 1 {
		rows = 1
	}
	return ROWindowTitleHeight + rows*chatShortcutsRowH + (rows-1)*chatShortcutsRowGap + chatShortcutsPad*2 + ROWindowFooterHeight
}

func init() {
	s := rotheme.Section("chat_shortcuts", "chat shortcut editor")
	s.Int("width", &chatShortcutsW, 160, 900, "window width")
	s.Int("rows", &chatShortcutsRows, 1, 20, "shortcut rows used for the window height")
	s.Int("row_height", &chatShortcutsRowH, 16, 64, "shortcut row height")
	s.Int("row_gap", &chatShortcutsRowGap, 0, 16, "gap between shortcut rows")
	s.Int("padding", &chatShortcutsPad, 0, 48, "content padding")
	s.Int("label_width", &chatShortcutsLabelW, 24, 160, "alt-key label width")
	s.Int("label_gap", &chatShortcutsLabelGap, 0, 32, "gap between the label and the field")
}
