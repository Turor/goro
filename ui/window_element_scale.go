package ui

func (w *Window) elementScale() float32 {
	if w == nil {
		return 1
	}
	if overlay := w.positionedOverlay(); overlay != nil && overlay.contentScale > 0 {
		return overlay.contentScale
	}
	frame, ok := w.content.(*windowFrame)
	if ok && frame != nil && frame.contentScale > 0 {
		return frame.contentScale
	}
	return 1
}

func (w *Window) currentSize(designW, designH int) (int, int) {
	if w != nil && w.width > 0 && w.height > 0 {
		return w.width, w.height
	}
	return designW, designH
}

func scaleInt(v int, scale float32) int {
	if scale <= 0 || scale == 1 {
		return v
	}
	return int(float32(v) * scale)
}

func scaleF(v, scale float32) float32 {
	if scale <= 0 {
		return v
	}
	return v * scale
}

func scaledScrollMax(content, view int, scale float32) float32 {
	if scale <= 0 {
		scale = 1
	}
	return float32(maxInt(0, content-view)) * scale
}
