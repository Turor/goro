package ui

import (
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
	"github.com/kivutar/goro/client"
	"github.com/kivutar/goro/input"
	"github.com/kivutar/goro/ui/rotheme"
)

func (w *Window) Resizing() bool { return w.resizing }

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
	return localX >= w.width-grip && localX < w.width && localY >= w.height-grip && localY < w.height
}

func (w *Window) resizeTo(ctx client.Context, width, height int) {
	width, height = w.clampResize(ctx, width, height)
	if width == w.width && height == w.height && w.userSized {
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
	frame, ok := w.content.(*windowFrame)
	if !ok {
		return
	}
	frame.Width(float32(w.width))
	frame.Height(float32(w.height))
	w.heightDelta = frame.heightDelta
}

func (w *Window) startResize(ctx client.Context) {
	w.resizing = true
	w.userMoved = true
	w.resizeStartW = w.width
	w.resizeStartH = w.height
	w.resizeAnchorX = ctx.Input.MouseX
	w.resizeAnchorY = ctx.Input.MouseY
	w.beginDragLayer(ctx)
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
	canvas.DrawRect(
		geometry.NewRect(float32(width)-grip, float32(height)-grip, grip, grip),
		rotheme.Default.Colors.ButtonBorder,
	)
}
