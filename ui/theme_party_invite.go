package ui

import "github.com/kivutar/goro/ui/rotheme"

var partyInviteW = 286

func init() {
	s := rotheme.Section("party_invite", "party invitation name prompt")
	s.Int("width", &partyInviteW, 160, 800, "party invite window width")
}

func partyInviteHeight() int {
	return ROWindowTitleHeight + textPromptContentH + ROWindowFooterHeight
}
