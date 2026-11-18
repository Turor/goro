package ui

import "github.com/kivutar/goro/ui/rotheme"

var friendContextMenuWidth = 124

func init() {
	s := rotheme.Section("friend_context_menu", "friend list right-click menu")
	s.Int("width", &friendContextMenuWidth, 60, 400, "friend menu width")
}
