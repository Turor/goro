package ui

import "github.com/kivutar/goro/ui/rotheme"

var tabBorderW float32 = 1

func init() {
	s := rotheme.Section("tabs", "tab buttons")
	s.Float("border_width", &tabBorderW, 0, 8, "tab outline thickness")
}
