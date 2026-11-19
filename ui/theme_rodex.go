package ui

import "github.com/kivutar/goro/ui/rotheme"

var (
	rodexWindowW = 420
	rodexWindowH = 340
	rodexRowH    = 22
	rodexRows    = 8
	rodexPad     = 10
	rodexGap     = 6
	rodexLabelW  = 56
)

func init() {
	s := rotheme.Section("rodex", "RoDEX mailbox")
	s.Int("width", &rodexWindowW, 280, 900, "window width")
	s.Int("height", &rodexWindowH, 180, 800, "window height")
	s.Int("row_height", &rodexRowH, 12, 48, "mail row height")
	s.Int("visible_rows", &rodexRows, 4, 24, "visible mail rows")
	s.Int("padding", &rodexPad, 0, 32, "content padding")
	s.Int("gap", &rodexGap, 0, 24, "row gap")
	s.Int("label_width", &rodexLabelW, 40, 160, "compose label width")
}
