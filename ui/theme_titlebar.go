package ui

import "github.com/kivutar/goro/ui/rotheme"

var titlebarBorderH float32 = 1

func init() {
	s := rotheme.Section("titlebar", "window title bar")
	s.Float("border_height", &titlebarBorderH, 0, 8, "title bar bottom border thickness")
}
