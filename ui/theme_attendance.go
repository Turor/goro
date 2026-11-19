package ui

import "github.com/kivutar/goro/ui/rotheme"

var (
	attendanceWindowW = 488
	attendanceWindowH = 413
	attendancePad     = 12
	attendanceGap     = 6
	attendanceCellW   = 88
	attendanceCellH   = 72
	attendanceDays    = 20
	attendanceCols    = 5
)

func init() {
	s := rotheme.Section("attendance", "attendance check window")
	s.Int("width", &attendanceWindowW, 180, 700, "window width")
	s.Int("height", &attendanceWindowH, 80, 600, "window height")
	s.Int("padding", &attendancePad, 0, 32, "content padding")
	s.Int("gap", &attendanceGap, 0, 24, "cell gap")
	s.Int("cell_width", &attendanceCellW, 48, 160, "day cell width")
	s.Int("cell_height", &attendanceCellH, 40, 140, "day cell height")
	s.Int("days", &attendanceDays, 1, 31, "calendar days")
	s.Int("columns", &attendanceCols, 1, 10, "calendar columns")
}
