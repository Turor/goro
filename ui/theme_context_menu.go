package ui

import "github.com/kivutar/goro/ui/rotheme"

// [context_menu] defaults shared by every right-click menu.
var (
	contextMenuPadding = 2
	contextMenuRadius  = 5
)

func init() {
	s := rotheme.Section("context_menu", "shared right-click menu chrome")
	s.Int("padding", &contextMenuPadding, 0, 16, "context menu padding")
	s.Int("radius", &contextMenuRadius, 0, 24, "context menu corner radius")
}
