package ui

import (
	"github.com/gogpu/ui/widget"
	"github.com/kivutar/goro/ui/rotheme"
)

// [chat_room] defaults: chat room conversation window.
var (
	chatRoomWindowW        = 480
	chatRoomMembersW       = 160
	chatRoomWindowContentH = 152
	chatRoomMessagePadX    = 8
	chatRoomMessagePadY    = 6
	chatRoomMessageGap     = 1
	chatRoomSummaryH       = 20
	chatRoomSummaryReserve = 24
	chatRoomSummaryPadY    = 3
	chatRoomTextInset      = 6
	chatRoomMemberRowH     = 22
	chatRoomMemberHeaderH  = 24
	chatRoomInputH         = 24
	chatRoomOwnerGap       = 4
	chatRoomOwnerPad       = 4
	chatRoomOwnerH         = 34
	chatRoomWindowBg       = widget.RGBA8(0, 0, 0, 0)
	chatRoomPanelBg        = widget.RGBA8(14, 18, 24, 188)
	chatRoomPanelBorder    = widget.RGBA8(180, 198, 218, 95)
	chatRoomMessageText    = widget.RGBA8(232, 238, 245, 255)
	chatRoomSystemText     = widget.RGBA8(160, 190, 230, 255)
	chatRoomSummaryText    = widget.RGBA8(181, 222, 239, 255)
	chatRoomEmptyText      = widget.RGBA8(150, 165, 182, 255)
)

func chatRoomWindowH() int {
	return ROWindowTitleHeight + chatRoomWindowContentH + ROWindowFooterHeight
}

func init() {
	s := rotheme.Section("chat_room", "chat room conversation window")
	s.Int("width", &chatRoomWindowW, 280, 1400, "window width")
	s.Int("members_width", &chatRoomMembersW, 80, 480, "member list width")
	s.Int("content_height", &chatRoomWindowContentH, 64, 500, "conversation area height")
	s.Int("padding_x", &chatRoomMessagePadX, 0, 32, "message area horizontal padding")
	s.Int("padding_y", &chatRoomMessagePadY, 0, 32, "message area vertical padding")
	s.Int("message_gap", &chatRoomMessageGap, 0, 8, "gap between message lines")
	s.Int("summary_height", &chatRoomSummaryH, 8, 64, "member summary row height")
	s.Int("summary_reserve", &chatRoomSummaryReserve, 0, 80, "height reserved above the messages for the summary")
	s.Int("summary_padding_y", &chatRoomSummaryPadY, 0, 16, "member summary vertical padding")
	s.Int("text_inset", &chatRoomTextInset, 0, 24, "extra inset when measuring wrapped message text")
	s.Int("member_row_height", &chatRoomMemberRowH, 12, 64, "member list row height")
	s.Int("member_header_height", &chatRoomMemberHeaderH, 12, 64, "member list header height")
	s.Int("input_height", &chatRoomInputH, 16, 64, "message field height")
	s.Int("owner_gap", &chatRoomOwnerGap, 0, 16, "gap between owner buttons")
	s.Int("owner_padding", &chatRoomOwnerPad, 0, 16, "owner button row padding")
	s.Int("owner_height", &chatRoomOwnerH, 20, 96, "owner button row height")
	s.Color("window_background", &chatRoomWindowBg, "window background")
	s.Color("panel_background", &chatRoomPanelBg, "message panel background")
	s.Color("panel_border", &chatRoomPanelBorder, "message panel border")
	s.Color("message_text", &chatRoomMessageText, "chat message text")
	s.Color("system_text", &chatRoomSystemText, "system message text")
	s.Color("summary_text", &chatRoomSummaryText, "member summary text")
	s.Color("empty_text", &chatRoomEmptyText, "empty conversation text")
}
