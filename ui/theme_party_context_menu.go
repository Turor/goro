package ui

import "github.com/kivutar/goro/ui/rotheme"

var partyContextMenuWidth = 132

func init() {
	s := rotheme.Section("party_context_menu", "party member right-click menu")
	s.Int("width", &partyContextMenuWidth, 60, 400, "party menu width")
}
