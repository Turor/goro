package ui

import "github.com/kivutar/goro/ui/rotheme"

var mercenaryContextMenuWidth = 120

func init() {
	s := rotheme.Section("mercenary_context_menu", "mercenary right-click menu")
	s.Int("width", &mercenaryContextMenuWidth, 60, 400, "mercenary menu width")
}
