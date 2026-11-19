package ui

import "github.com/kivutar/goro/ui/rotheme"

var (
	equipSwitchWindowW = 280
	equipSwitchWindowH = 140
	equipSwitchPad     = 12
)

func init() {
	s := rotheme.Section("equip_switch", "equip switch window")
	s.Int("width", &equipSwitchWindowW, 180, 600, "window width")
	s.Int("height", &equipSwitchWindowH, 80, 400, "window height")
	s.Int("padding", &equipSwitchPad, 0, 32, "content padding")
}
