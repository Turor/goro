package ui

import "github.com/kivutar/goro/ui/rotheme"

// [friend_settings] defaults: friend whisper setup window.
var (
	friendSettingsW        = 286
	friendSettingsContentH = 112
	friendSettingsPad      = 14
	friendSettingsGap      = 8
)

func friendSettingsWindowH() int {
	return ROWindowTitleHeight + friendSettingsContentH + ROWindowFooterHeight
}

func init() {
	s := rotheme.Section("friend_settings", "friend whisper setup window")
	s.Int("width", &friendSettingsW, 160, 900, "window width")
	s.Int("content_height", &friendSettingsContentH, 48, 400, "content height")
	s.Int("padding", &friendSettingsPad, 0, 48, "content padding")
	s.Int("gap", &friendSettingsGap, 0, 32, "gap between options")
}
