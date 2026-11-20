package ui

func (w *Window) elementScale() float32 {
	if w == nil {
		return 1
	}
	frame, ok := w.content.(*windowFrame)
	if !ok || frame == nil || frame.contentScale <= 0 {
		return 1
	}
	return frame.contentScale
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
