package ui

import (
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/widget"
	"github.com/kivutar/goro/client"
	"github.com/kivutar/goro/input"
	"github.com/kivutar/goro/ui/rotheme"
)

func (w *Window) Resizing() bool { return w.resizing }

func (w *Window) holdsUserSize() bool {
	return w != nil && (w.userSized || w.resizing)
}

func (w *Window) rememberSize(width, height int) {
	if width < 1 || height < 1 {
		return
	}
	w.userSized = true
	w.userMoved = true
	w.userWidth, w.userHeight = width, height
	w.width, w.height = width, height
}

func (w *Window) clampResize(ctx client.Context, width, height int) (int, int) {
	minW := windowMinWidth
	minH := maxInt(windowMinHeight, w.titleHeight)
	if minW < 1 {
		minW = 1
	}
	if minH < 1 {
		minH = 1
	}
	screenW, screenH := 0, 0
	if ctx.ScreenW > 0 || ctx.ScreenH > 0 {
		screenW, screenH = ctx.ScreenSize()
	}
	maxW := width
	maxH := height
	if screenW > 0 {
		maxW = maxInt(minW, screenW-windowScreenMargin-w.x)
	}
	if screenH > 0 {
		maxH = maxInt(minH, screenH-windowScreenMargin-w.y)
	}
	return clampWindowInt(width, minW, maxW), clampWindowInt(height, minH, maxH)
}

func (w *Window) resizeHit(x, y int) bool {
	grip := windowResizeGrip
	if grip < 1 || w.width < grip || w.height < grip {
		return false
	}
	localX, localY := x-w.x, y-w.y
	fromRight := w.width - localX
	fromBottom := w.height - localY
	return fromRight > 0 && fromBottom > 0 && fromRight+fromBottom <= grip
}

func (w *Window) resizeTo(ctx client.Context, width, height int) {
	designW, designH := w.resizeDesign()
	width, height = lockResizeAspect(designW, designH, width, height)
	width, height = w.clampResize(ctx, width, height)
	width, height = lockResizeAspect(designW, designH, width, height)
	width, height = w.clampResize(ctx, width, height)
	if width == w.width && height == w.height {
		return
	}
	w.rememberSize(width, height)
	w.applyUserFrame()
	if overlay := w.positionedOverlay(); overlay != nil {
		overlay.setFrame(w.x, w.y, w.width, w.height)
	}
	invalidateWindowLayout(ctx)
}

func (w *Window) applyUserFrame() {
	if !w.userSized {
		return
	}
	if overlay := w.positionedOverlay(); overlay != nil {
		overlay.contentScale = 1
	}
	frame, ok := w.content.(*windowFrame)
	if !ok {
		w.scaleLooseContent()
		return
	}
	w.width, w.height = w.userWidth, w.userHeight
	if frame.designW <= 0 || frame.designH <= 0 {
		frame.designW = frame.Style().ExplicitWidth
		frame.designH = frame.Style().ExplicitHeight
	}
	frame.Width(float32(w.userWidth))
	frame.Height(float32(w.userHeight))
	frame.fillUserSize = true
	w.heightDelta = frame.heightDelta
	if overlay := w.positionedOverlay(); overlay != nil {
		overlay.contentScale = frameScale(frame.designW, frame.designH, float32(w.userWidth), float32(w.userHeight))
	}
}

func (w *Window) scaleLooseContent() {
	box, ok := w.content.(*primitives.BoxWidget)
	if !ok || w.userWidth <= 0 || w.userHeight <= 0 {
		return
	}
	if w.looseDesignW <= 0 || w.looseDesignH <= 0 {
		w.looseDesignW = box.Style().ExplicitWidth
		w.looseDesignH = box.Style().ExplicitHeight
	}
	scale := frameScale(w.looseDesignW, w.looseDesignH, float32(w.userWidth), float32(w.userHeight))
	if w.looseTokens == nil {
		w.looseTokens = make(map[widget.Widget]*scaleToken)
	}
	frame := &windowFrame{contentScale: scale, scaleTokens: w.looseTokens}
	for _, child := range box.Children() {
		frame.scaleWidgetTree(child)
	}
	box.Width(float32(w.userWidth))
	box.Height(float32(w.userHeight))
	if overlay := w.positionedOverlay(); overlay != nil {
		overlay.contentScale = scale
	}
}

func (w *Window) startResize(ctx client.Context) {
	w.resizing = true
	w.userMoved = true
	w.resizeStartW = w.width
	w.resizeStartH = w.height
	w.resizeAnchorX = ctx.Input.MouseX
	w.resizeAnchorY = ctx.Input.MouseY
}

func (w *Window) updateResize(ctx client.Context) bool {
	if !w.resizing {
		return false
	}
	if ctx.Input.MousePressed(input.MouseButtonLeft) {
		w.resizeTo(ctx, w.resizeStartW+ctx.Input.MouseX-w.resizeAnchorX, w.resizeStartH+ctx.Input.MouseY-w.resizeAnchorY)
		return true
	}
	w.resizing = false
	w.endDragLayer(ctx)
	w.Publish(ctx)
	return true
}

func drawWindowResizeGrip(canvas widget.Canvas, width, height int) {
	grip := float32(windowResizeGrip)
	if grip <= 0 || width < 1 || height < 1 {
		return
	}
	row := windowBorderWidth
	if row <= 0 {
		row = grip
	}
	color := rotheme.Default.Colors.ButtonBorder
	y := float32(height) - grip
	bottom := float32(height)
	for y < bottom {
		span := bottom - y
		if span > grip {
			span = grip
		}
		canvas.DrawRect(geometry.NewRect(float32(width)-span, y, span, row), color)
		y += row
	}
}

func (w *Window) resizeDesign() (int, int) {
	if frame, ok := w.content.(*windowFrame); ok && frame.designW > 0 && frame.designH > 0 {
		return int(frame.designW), int(frame.designH)
	}
	if w.looseDesignW > 0 && w.looseDesignH > 0 {
		return int(w.looseDesignW), int(w.looseDesignH)
	}
	return w.width, w.height
}

func lockResizeAspect(designW, designH, width, height int) (int, int) {
	if designW <= 0 || designH <= 0 || width <= 0 || height <= 0 {
		return width, height
	}
	deltaW := width - designW
	if deltaW < 0 {
		deltaW = -deltaW
	}
	deltaH := height - designH
	if deltaH < 0 {
		deltaH = -deltaH
	}
	scale := float32(width) / float32(designW)
	if deltaH > deltaW {
		scale = float32(height) / float32(designH)
	}
	if scale <= 0 {
		return width, height
	}
	return int(float32(designW) * scale), int(float32(designH) * scale)
}

func frameScale(designW, designH, userW, userH float32) float32 {
	if designW <= 0 || designH <= 0 || userW <= 0 || userH <= 0 {
		return 1
	}
	scale := userW / designW
	if heightScale := userH / designH; heightScale < scale {
		scale = heightScale
	}
	if scale <= 0 {
		return 1
	}
	return scale
}
