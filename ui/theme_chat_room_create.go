package ui

import "github.com/kivutar/goro/ui/rotheme"

// [chat_room_create] defaults: make-a-room and room settings form.
var (
	chatRoomCreateW         = 306
	chatRoomCreateContentH  = 180
	chatRoomCreateFieldH    = 24
	chatRoomCreateLimitW    = 72
	chatRoomCreateLabelGap  = 4
	chatRoomCreateColumnGap = 12
	chatRoomCreatePad       = 14
	chatRoomCreateGap       = 6
)

func chatRoomCreateWindowH() int {
	return ROWindowTitleHeight + chatRoomCreateContentH + ROWindowFooterHeight
}

func init() {
	s := rotheme.Section("chat_room_create", "chat room create and settings form")
	s.Int("width", &chatRoomCreateW, 200, 900, "window width")
	s.Int("content_height", &chatRoomCreateContentH, 80, 560, "form content height")
	s.Int("field_height", &chatRoomCreateFieldH, 16, 64, "text field height")
	s.Int("limit_width", &chatRoomCreateLimitW, 40, 220, "member limit column width")
	s.Int("label_gap", &chatRoomCreateLabelGap, 0, 16, "gap between a label and its field")
	s.Int("column_gap", &chatRoomCreateColumnGap, 0, 48, "gap between the limit and type columns")
	s.Int("padding", &chatRoomCreatePad, 0, 48, "form padding")
	s.Int("gap", &chatRoomCreateGap, 0, 24, "gap between form rows")
}
