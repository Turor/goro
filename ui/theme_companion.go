package ui

import (
	"image/color"

	"github.com/kivutar/goro/ui/rotheme"
)

var (
	companionWindowW         = 640
	companionWindowH         = 500
	companionWindowTabW      = 92
	companionWindowTabH      = 23
	companionWindowPad       = 10
	companionWindowRowH      = 22
	companionListRowH        = 30
	companionDetailLabelW    = 78
	companionFrameGap        = 2
	companionRowTextGap      = 2
	companionRowGap          = 8
	companionListGap         = 8
	companionStatRowGap      = 6
	companionActionGap       = 6
	companionDetailGap       = 4
	companionDetailPad       = 6
	companionEmptyPad        = 4
	companionButtonW         = 72
	companionMediumButtonW   = 80
	companionStorePetButtonW = 88
	companionWideButtonW     = 100
	companionFrameBackground = color.RGBA{R: 255, G: 255, B: 255, A: 255}
)

func init() {
	s := rotheme.Section("companion", "companion window (pet, homunculus, mercenary, stable)")
	s.Int("width", &companionWindowW, 400, 900, "companion window width")
	s.Int("height", &companionWindowH, 280, 800, "companion window height")
	s.Int("tab_width", &companionWindowTabW, 60, 160, "companion tab width")
	s.Int("tab_height", &companionWindowTabH, 16, 40, "companion tab height")
	s.Int("padding", &companionWindowPad, 0, 32, "companion tab content padding")
	s.Int("row_height", &companionWindowRowH, 12, 64, "detail row height")
	s.Int("list_row_height", &companionListRowH, 16, 96, "stable list row height")
	s.Int("detail_label_width", &companionDetailLabelW, 24, 240, "detail label width")
	s.Int("frame_gap", &companionFrameGap, 0, 16, "gap between the tab strip and the content")
	s.Int("row_text_gap", &companionRowTextGap, 0, 16, "gap between a list row title and its subtitle")
	s.Int("row_gap", &companionRowGap, 0, 32, "gap between items in a row")
	s.Int("list_gap", &companionListGap, 0, 32, "gap between list rows")
	s.Int("stat_row_gap", &companionStatRowGap, 0, 24, "gap between pet stat rows")
	s.Int("action_gap", &companionActionGap, 0, 24, "gap between pet detail buttons")
	s.Int("detail_gap", &companionDetailGap, 0, 24, "gap between pet detail rows")
	s.Int("detail_padding", &companionDetailPad, 0, 32, "pet detail panel padding")
	s.Int("empty_padding", &companionEmptyPad, 0, 32, "empty tab message padding")
	s.Int("button_width", &companionButtonW, 32, 200, "footer and row button width")
	s.Int("medium_button_width", &companionMediumButtonW, 32, 200, "Summon and Unequip button width")
	s.Int("store_pet_button_width", &companionStorePetButtonW, 32, 200, "Store pet button width")
	s.Int("wide_button_width", &companionWideButtonW, 32, 240, "Store homun and Deposit egg button width")
	s.RGBA("frame_background", &companionFrameBackground, "background behind the tab strip and content")
}
