package ui

import "github.com/kivutar/goro/ui/rotheme"

var (
	partySettingsW       = 286
	partySettingsContent = 132
	partySettingsPad     = 14
	partySettingsGap     = 8
)

func init() {
	s := rotheme.Section("party_settings", "party exp and invite settings")
	s.Int("width", &partySettingsW, 160, 800, "party settings window width")
	s.Int("content_height", &partySettingsContent, 40, 400, "party settings content height")
	s.Int("padding", &partySettingsPad, 0, 40, "party settings padding")
	s.Int("gap", &partySettingsGap, 0, 24, "party settings row gap")
}

func partySettingsHeight() int {
	return ROWindowTitleHeight + partySettingsContent + ROWindowFooterHeight
}
