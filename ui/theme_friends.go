package ui

import (
	"github.com/gogpu/ui/widget"
	"github.com/kivutar/goro/ui/rotheme"
)

// [friends] defaults: friends and party list window.
var (
	friendsWindowWidth   = 286
	friendsWindowHeight  = 256
	friendsRowHeight     = 24
	friendsTabWidth      = 72
	friendsTabHeight     = 24
	friendsRowTextX      = 8
	friendsRowTextRight  = 88
	friendsRowTextH      = 14
	friendsRowStateRight = 72
	friendsRowStateW     = 64
	friendsPartyTextY    = 3
	friendsPartyTextH    = 13
	friendsPartyBarInset = 8
	friendsPartyBarH     = 5
	friendsPartyBarTrack = widget.RGBA8(224, 232, 242, 255)
)

func init() {
	s := rotheme.Section("friends", "friends and party list window")
	s.Int("width", &friendsWindowWidth, 160, 900, "window width")
	s.Int("height", &friendsWindowHeight, 140, 800, "window height")
	s.Int("row_height", &friendsRowHeight, 16, 64, "friend and party row height")
	s.Int("tab_width", &friendsTabWidth, 40, 220, "friends and party tab width")
	s.Int("tab_height", &friendsTabHeight, 16, 64, "friends and party tab height")
	s.Int("row_text_x", &friendsRowTextX, 0, 32, "row text left inset")
	s.Int("row_text_right", &friendsRowTextRight, 32, 240, "space kept clear of the state text")
	s.Int("row_text_height", &friendsRowTextH, 8, 40, "friend row text height")
	s.Int("row_state_right", &friendsRowStateRight, 24, 200, "state text inset from the row's right edge")
	s.Int("row_state_width", &friendsRowStateW, 24, 200, "state text width")
	s.Int("party_text_y", &friendsPartyTextY, 0, 24, "party row text offset from the top")
	s.Int("party_text_height", &friendsPartyTextH, 8, 40, "party row text height")
	s.Int("party_bar_inset", &friendsPartyBarInset, 0, 32, "party hp bar inset")
	s.Int("party_bar_height", &friendsPartyBarH, 1, 20, "party hp bar height")
	s.Color("party_bar_track", &friendsPartyBarTrack, "party hp bar track")
}
