package ui

import "github.com/kivutar/goro/ui/rotheme"

var (
	instanceWindowW = 320
	instanceWindowH = 178
	instancePad     = 12
	instanceGap     = 6
)

func init() {
	s := rotheme.Section("instance", "memorial dungeon window")
	s.Int("width", &instanceWindowW, 180, 1000, "window width")
	s.Int("height", &instanceWindowH, 100, 700, "window height")
	s.Int("padding", &instancePad, 0, 48, "content padding")
	s.Int("gap", &instanceGap, 0, 32, "gap between status lines")
}
