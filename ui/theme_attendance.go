package ui

import "github.com/kivutar/goro/ui/rotheme"

var (
	attendanceWindowW = 280
	attendanceWindowH = 140
	attendancePad     = 12
)

func init() {
	s := rotheme.Section("attendance", "attendance check window")
	s.Int("width", &attendanceWindowW, 180, 600, "window width")
	s.Int("height", &attendanceWindowH, 80, 400, "window height")
	s.Int("padding", &attendancePad, 0, 32, "content padding")
}
