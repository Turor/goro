package ui

import (
	"github.com/gogpu/ui/widget"
	"github.com/kivutar/goro/ui/rotheme"
)

// [whisper] defaults: one 1:1 chat window.
var (
	whisperWindowW        = 286
	whisperWindowContentH = 120
	whisperMessagePadX    = 8
	whisperMessagePadY    = 6
	whisperInputH         = 24
	whisperWindowBg       = widget.RGBA8(0, 0, 0, 0)
	whisperPanelBg        = widget.RGBA8(14, 18, 24, 188)
	whisperPanelBorder    = widget.RGBA8(180, 198, 218, 95)
	whisperIncomingText   = widget.RGBA8(181, 222, 239, 255)
	whisperOutgoingText   = widget.RGBA8(255, 255, 120, 255)
	whisperEmptyText      = widget.RGBA8(150, 165, 182, 255)
)

func whisperWindowH() int {
	return ROWindowTitleHeight + whisperWindowContentH + ROWindowFooterHeight
}

func init() {
	s := rotheme.Section("whisper", "1:1 chat window")
	s.Int("width", &whisperWindowW, 160, 900, "window width")
	s.Int("content_height", &whisperWindowContentH, 48, 400, "message area height")
	s.Int("padding_x", &whisperMessagePadX, 0, 32, "message area horizontal padding")
	s.Int("padding_y", &whisperMessagePadY, 0, 32, "message area vertical padding")
	s.Int("input_height", &whisperInputH, 16, 64, "message field height")
	s.Color("window_background", &whisperWindowBg, "window background")
	s.Color("panel_background", &whisperPanelBg, "message panel background")
	s.Color("panel_border", &whisperPanelBorder, "message panel border")
	s.Color("incoming_text", &whisperIncomingText, "incoming message text")
	s.Color("outgoing_text", &whisperOutgoingText, "outgoing message text")
	s.Color("empty_text", &whisperEmptyText, "empty conversation text")
}
