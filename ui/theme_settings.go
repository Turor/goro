package ui

import "github.com/kivutar/goro/ui/rotheme"

var (
	settingsWindowW       = 300
	settingsWindowH       = 359
	settingsTabW          = 88
	settingsTabH          = 24
	settingsTabOverlap    = -1
	settingsPad           = 14
	settingsGap           = 8
	settingsOptionButtonW = 48
	settingsDividerH      = 1
	// Theme dropdown on the Gameplay tab.
	settingsThemeDropdownW = 132
	settingsThemeItemH     = 20
	settingsThemeMenuRows  = 8
)

func init() {
	s := rotheme.Section("settings", "settings window")
	s.Int("width", &settingsWindowW, 200, 1000, "window width")
	s.Int("height", &settingsWindowH, 160, 1000, "window height")
	s.Int("tab_width", &settingsTabW, 40, 240, "tab width")
	s.Int("tab_height", &settingsTabH, 16, 64, "tab height")
	s.Int("tab_overlap", &settingsTabOverlap, -16, 16, "overlap between tabs")
	s.Int("padding", &settingsPad, 0, 48, "tab content padding")
	s.Int("gap", &settingsGap, 0, 32, "gap between settings rows")
	s.Int("option_button_width", &settingsOptionButtonW, 24, 160, "option cycle button width")
	s.Int("divider_height", &settingsDividerH, 0, 8, "line under the tabs")
	s.Int("theme_dropdown_width", &settingsThemeDropdownW, 60, 400, "Theme dropdown width")
	s.Int("theme_item_height", &settingsThemeItemH, 12, 48, "Theme menu row height")
	s.Int("theme_menu_rows", &settingsThemeMenuRows, 1, 24, "Theme menu visible rows before it scrolls")
}
