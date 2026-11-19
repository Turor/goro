package ui

import "github.com/kivutar/goro/ui/rotheme"

var (
	searchStoreWindowW = 380
	searchStoreWindowH = 280
	searchStoreRowH    = 22
	searchStorePad     = 10
	searchStoreGap     = 4
)

func init() {
	s := rotheme.Section("search_store", "vending search window")
	s.Int("width", &searchStoreWindowW, 240, 800, "window width")
	s.Int("height", &searchStoreWindowH, 140, 700, "window height")
	s.Int("row_height", &searchStoreRowH, 12, 48, "result row height")
	s.Int("padding", &searchStorePad, 0, 32, "content padding")
	s.Int("gap", &searchStoreGap, 0, 24, "row gap")
}
