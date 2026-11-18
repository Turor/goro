package ui

import (
	"github.com/gogpu/ui/widget"
	"github.com/kivutar/goro/ui/rotheme"
)

// [teleport] defaults: the teleport and warp-portal destination list.
var (
	teleportModalWidth   = 260
	teleportModalMinH    = 168
	teleportModalMaxRows = 6
	teleportModalPad     = 14
	teleportModalGap     = 8
	teleportModalRowH    = 20
	teleportStatusColor  = widget.RGBA8(204, 48, 48, 255)
)

func init() {
	s := rotheme.Section("teleport", "teleport and warp portal window")
	s.Int("width", &teleportModalWidth, 160, 800, "window width")
	s.Int("min_height", &teleportModalMinH, 80, 480, "minimum window height")
	s.Int("max_rows", &teleportModalMaxRows, 1, 20, "maximum destination rows")
	s.Int("padding", &teleportModalPad, 0, 48, "content padding")
	s.Int("gap", &teleportModalGap, 0, 32, "gap between destination rows and text")
	s.Int("row_height", &teleportModalRowH, 12, 48, "destination row height")
	s.Color("status_color", &teleportStatusColor, "error status text color")
}
