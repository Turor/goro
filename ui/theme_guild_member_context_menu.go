package ui

import "github.com/kivutar/goro/ui/rotheme"

var guildMemberContextMenuWidth = 128

func init() {
	s := rotheme.Section("guild_member_context_menu", "guild member right-click menu")
	s.Int("width", &guildMemberContextMenuWidth, 60, 400, "guild member menu width")
}
