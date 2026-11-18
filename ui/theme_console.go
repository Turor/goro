package ui

import (
	"image/color"

	"github.com/gogpu/ui/widget"
	"github.com/kivutar/goro/ui/rotheme"
)

// Chat console defaults. Message colours are snapshotted when a line is added.
var (
	consoleWidth        = 480
	consoleHeight       = 176
	consoleMinWidth     = 260
	consoleMinHeight    = 128
	consoleFieldH       = 24
	consoleLineH        = 14
	consoleLineGap      = 1
	consolePadX         = 8
	consolePadY         = 6
	consoleGap          = 4
	consoleTextInset    = 4
	consoleMessageSlack = 4
	consoleMinMessageH  = 20
	consoleScrollLines  = 3

	consoleBorderWidth float32 = 1

	consoleBackground       = widget.RGBA8(14, 18, 24, 188)
	consoleBorderColor      = widget.RGBA8(180, 198, 218, 95)
	consoleColorChat        = color.RGBA{R: 235, G: 242, B: 250, A: 255}
	consoleColorSystem      = color.RGBA{R: 252, G: 221, B: 128, A: 255}
	consoleColorBlue        = color.RGBA{R: 0, G: 255, B: 255, A: 255}
	consoleColorGuildChat   = color.RGBA{R: 180, G: 255, B: 180, A: 255}
	consoleColorGuild       = color.RGBA{R: 255, G: 255, B: 99, A: 255}
	consoleColorError       = color.RGBA{R: 255, G: 132, B: 132, A: 255}
	consoleColorPlaceholder = color.RGBA{R: 150, G: 165, B: 182, A: 255}
)

func init() {
	s := rotheme.Section("console", "chat console at the bottom of the game view")
	s.Int("width", &consoleWidth, 160, 1600, "console width")
	s.Int("height", &consoleHeight, 80, 800, "console height")
	s.Int("min_width", &consoleMinWidth, 80, 800, "minimum console width on a small screen")
	s.Int("min_height", &consoleMinHeight, 48, 400, "minimum console height on a small screen")
	s.Int("field_height", &consoleFieldH, 12, 64, "chat input height")
	s.Int("line_height", &consoleLineH, 8, 40, "message line height")
	s.Int("line_gap", &consoleLineGap, 0, 16, "gap between message lines")
	s.Int("padding_x", &consolePadX, 0, 32, "horizontal padding")
	s.Int("padding_y", &consolePadY, 0, 32, "vertical padding")
	s.Int("gap", &consoleGap, 0, 24, "gap between the message list and the input")
	s.Int("text_inset", &consoleTextInset, 0, 24, "space kept clear of the scrollbar")
	s.Int("message_slack", &consoleMessageSlack, 0, 24, "extra space under the message list")
	s.Int("min_message_height", &consoleMinMessageH, 8, 200, "minimum message list height")
	s.Int("scroll_lines", &consoleScrollLines, 1, 12, "lines scrolled by one step")
	s.Float("border_width", &consoleBorderWidth, 0, 8, "console border width")
	s.Color("background", &consoleBackground, "console background")
	s.Color("border_color", &consoleBorderColor, "console border")
	s.RGBA("chat_color", &consoleColorChat, "normal chat text")
	s.RGBA("system_color", &consoleColorSystem, "system message text")
	s.RGBA("blue_color", &consoleColorBlue, "pickup and confirmation text")
	s.RGBA("guild_chat_color", &consoleColorGuildChat, "guild chat text")
	s.RGBA("guild_color", &consoleColorGuild, "guild notice text")
	s.RGBA("error_color", &consoleColorError, "error text")
	s.RGBA("placeholder_color", &consoleColorPlaceholder, "empty console placeholder")
}
