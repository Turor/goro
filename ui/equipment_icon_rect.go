package ui

import (
	"github.com/gogpu/ui/geometry"
)

func equipmentSlotIconRect(bounds geometry.Rect, slot equipmentSlotDef) geometry.Rect {
	size := bounds.Height() * 0.72
	if size > bounds.Width()*0.4 {
		size = bounds.Width() * 0.4
	}
	if size < 1 {
		size = 1
	}
	y := bounds.Min.Y + (bounds.Height()-size)/2
	switch slot.side {
	case equipmentSlotRight:
		return geometry.NewRect(bounds.Max.X-size-equipmentIconPad, y, size, size)
	case equipmentSlotCenter:
		return geometry.NewRect(bounds.Min.X+(bounds.Width()-size)/2, y, size, size)
	default:
		return geometry.NewRect(bounds.Min.X+equipmentIconPad, y, size, size)
	}
}
