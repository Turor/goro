package ui

import "github.com/kivutar/goro/ui/rotheme"

var petContextMenuWidth = 140

func init() {
	s := rotheme.Section("pet_context_menu", "pet right-click menu")
	s.Int("width", &petContextMenuWidth, 60, 400, "pet menu width")
}
