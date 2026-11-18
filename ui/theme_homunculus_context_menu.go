package ui

import "github.com/kivutar/goro/ui/rotheme"

var homunculusContextMenuWidth = 120

func init() {
	s := rotheme.Section("homunculus_context_menu", "homunculus right-click menu")
	s.Int("width", &homunculusContextMenuWidth, 60, 400, "homunculus menu width")
}
