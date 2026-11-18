package ui

import "github.com/kivutar/goro/ui/rotheme"

var playerContextMenuWidth = 118

func init() {
	s := rotheme.Section("player_context_menu", "player right-click menu")
	s.Int("width", &playerContextMenuWidth, 60, 400, "player menu width")
}
