package ui

import (
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/widget"
	"github.com/kivutar/goro/ui/rotheme"
)

// roFooterWidget decorates the existing footer layout without changing its
// sizing or event handling. Its embedded box has a transparent background.
type roFooterWidget struct {
	*primitives.BoxWidget
}

func (w *roFooterWidget) Draw(ctx widget.Context, canvas widget.Canvas) {
	if !w.IsVisible() || w.Bounds().IsEmpty() {
		return
	}
	bounds := w.Bounds()
	canvas.DrawRect(bounds, rotheme.Default.Colors.WindowFooter)
	for y := bounds.Min.Y; y < bounds.Max.Y; y += footerStripeH * 2 {
		canvas.DrawRect(geometry.NewRect(bounds.Min.X, y, bounds.Width(), min(footerStripeH, bounds.Max.Y-y)), footerStripeColor)
	}
	w.BoxWidget.Draw(ctx, canvas)
}
