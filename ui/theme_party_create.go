package ui

import "github.com/kivutar/goro/ui/rotheme"

var (
	partyCreateW        = 306
	partyCreateContentH = 158
	partyCreatePad      = 14
	partyCreateGap      = 6
	partyCreateNameH    = 24
)

func init() {
	s := rotheme.Section("party_create", "create party window")
	s.Int("width", &partyCreateW, 180, 800, "create party window width")
	s.Int("content_height", &partyCreateContentH, 60, 400, "create party content height")
	s.Int("padding", &partyCreatePad, 0, 40, "create party padding")
	s.Int("gap", &partyCreateGap, 0, 24, "create party row gap")
	s.Int("name_field_height", &partyCreateNameH, 14, 64, "party name field height")
}

func partyCreateHeight() int {
	return ROWindowTitleHeight + partyCreateContentH + ROWindowFooterHeight
}
