package ui

import "github.com/kivutar/goro/ui/rotheme"

// Server and character-server list.
var (
	serviceWindowWidth        = 280
	serviceWindowRows         = 6
	serviceWindowRowH         = 22
	serviceWindowPadding      = 10
	serviceWindowScreenMargin = 8
	serviceWindowBiasMinY     = 48
	serviceWindowHeight       int
)

func recomputeServiceWindowMetrics() {
	serviceWindowHeight = ROWindowTitleHeight + serviceWindowPadding*2 + serviceWindowRows*serviceWindowRowH + ROWindowFooterHeight
}

func init() {
	s := rotheme.Section("service", "login and character server list")
	s.Int("width", &serviceWindowWidth, 160, 800, "server list width")
	s.Int("rows", &serviceWindowRows, 1, 24, "visible server rows")
	s.Int("row_height", &serviceWindowRowH, 14, 48, "server row height")
	s.Int("padding", &serviceWindowPadding, 0, 40, "padding around the server list")
	s.Int("screen_margin", &serviceWindowScreenMargin, 0, 64, "minimum distance from the screen edge")
	s.Int("bias_min_y", &serviceWindowBiasMinY, 0, 200, "minimum top when the list sits low on the screen")
	rotheme.OnLayoutApplied(recomputeServiceWindowMetrics)
	recomputeServiceWindowMetrics()
}
