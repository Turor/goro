package ui

import "github.com/kivutar/goro/ui/rotheme"

// [item_table] defaults: the shared one-column item list.
var (
	itemTableMinW = float32(120)
)

func init() {
	s := rotheme.Section("item_table", "shared item list table")
	s.Float("item_min_width", &itemTableMinW, 40, 480, "item column minimum width")
}
