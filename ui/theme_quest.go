package ui

import "github.com/kivutar/goro/ui/rotheme"

var (
	questWindowW          = 350
	questWindowH          = 310
	questTabW             = 34
	questTabH             = 70
	questTabOverlap       = -1
	questDetailW          = 360
	questDetailH          = 390
	questDetailOffset     = 8
	questDetailPad        = 10
	questDetailGap        = 8
	questRowH             = 32
	questDescriptionRunes = 40
)

func init() {
	s := rotheme.Section("quest", "quest journal and detail windows")
	s.Int("width", &questWindowW, 200, 1200, "journal width")
	s.Int("height", &questWindowH, 160, 1000, "journal height")
	s.Int("tab_width", &questTabW, 16, 120, "side tab width")
	s.Int("tab_height", &questTabH, 24, 240, "side tab height")
	s.Int("tab_overlap", &questTabOverlap, -16, 16, "overlap between side tabs")
	s.Int("detail_width", &questDetailW, 200, 1200, "detail window width")
	s.Int("detail_height", &questDetailH, 160, 1200, "detail window height")
	s.Int("detail_offset", &questDetailOffset, 0, 64, "gap between the journal and the detail window")
	s.Int("detail_padding", &questDetailPad, 0, 48, "detail content padding")
	s.Int("detail_gap", &questDetailGap, 0, 32, "gap between detail lines")
	s.Int("row_height", &questRowH, 16, 96, "quest list row height")
	s.Int("description_runes", &questDescriptionRunes, 10, 160, "description wrap width in runes")
}
