package ui

import (
	"github.com/gogpu/ui/geometry"
)

func equipmentSlotScale(bounds geometry.Rect, slot equipmentSlotDef) float32 {
	designW := float32(equipmentLeftColW)
	designH := float32(equipmentRowH)
	switch slot.side {
	case equipmentSlotRight:
		designW = float32(equipmentRightColW)
	case equipmentSlotCenter:
		designW = float32(equipmentCenterColW)
		if equipmentCenterColH > 0 {
			designH = float32(equipmentCenterColH)
		}
	}
	if designW <= 0 || designH <= 0 || bounds.Width() <= 0 || bounds.Height() <= 0 {
		return 1
	}
	sx := bounds.Width() / designW
	sy := bounds.Height() / designH
	if sx < sy {
		return sx
	}
	return sy
}

func equipmentSlotIconRect(bounds geometry.Rect, slot equipmentSlotDef) geometry.Rect {
	size := bounds.Height() * 0.72
	if size > bounds.Width()*0.4 {
		size = bounds.Width() * 0.4
	}
	if size < 1 {
		size = 1
	}
	pad := equipmentIconPad * equipmentSlotScale(bounds, slot)
	y := bounds.Min.Y + (bounds.Height()-size)/2
	switch slot.side {
	case equipmentSlotRight:
		return geometry.NewRect(bounds.Max.X-size-pad, y, size, size)
	case equipmentSlotCenter:
		return geometry.NewRect(bounds.Min.X+(bounds.Width()-size)/2, y, size, size)
	default:
		return geometry.NewRect(bounds.Min.X+pad, y, size, size)
	}
}
