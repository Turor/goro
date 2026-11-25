package ui

import (
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"

	"github.com/kivutar/goro/ui/rotheme"
)

// opinionatedSplitRow keeps the right grid on the same columns as the full
// bag. The 7th column is removed for the divider; the left scrollbar sits in
// that column instead of pushing the right tiles over.
type opinionatedSplitRow struct {
	widget.WidgetBase
	left  widget.Widget
	mid   widget.Widget
	right widget.Widget
	scale float32
}

func newOpinionatedSplitRow(left, mid, right widget.Widget) *opinionatedSplitRow {
	row := &opinionatedSplitRow{left: left, mid: mid, right: right}
	row.SetVisible(true)
	row.SetEnabled(true)
	return row
}

func (r *opinionatedSplitRow) Layout(ctx widget.Context, constraints geometry.Constraints) geometry.Size {
	scale := rotheme.ContentScale()
	if scale <= 0 {
		scale = 1
	}
	r.scale = scale
	width := float32(opinionatedFrameW()) * scale
	height := float32(inventoryBagViewH) * scale
	size := constraints.Constrain(geometry.Sz(width, height))
	r.SetBounds(geometry.FromPointSize(r.Position(), size))
	r.place(ctx, scale, height)
	return size
}

func (r *opinionatedSplitRow) place(ctx widget.Context, scale, height float32) {
	leftW := float32(opinionatedScrollW(opinionatedSplitContentW())) * scale
	midX := float32(opinionatedSideViewW()) * scale
	midW := float32(opinionatedMiddleW()) * scale
	rightX := float32(opinionatedRightGridX()) * scale
	rightW := float32(opinionatedFrameW())*scale - rightX
	r.placeChild(ctx, r.left, 0, leftW, height)
	r.placeChild(ctx, r.mid, midX, midW, height)
	r.placeChild(ctx, r.right, rightX, rightW, height)
}

func (r *opinionatedSplitRow) placeChild(ctx widget.Context, child widget.Widget, x, width, height float32) {
	if child == nil {
		return
	}
	child.Layout(ctx, geometry.Tight(geometry.Sz(width, height)))
	if placed, ok := child.(interface{ SetBounds(geometry.Rect) }); ok {
		placed.SetBounds(geometry.NewRect(x, 0, width, height))
	}
}

func (r *opinionatedSplitRow) Draw(ctx widget.Context, canvas widget.Canvas) {
	if !r.IsVisible() {
		return
	}
	bounds := r.Bounds()
	canvas.PushTransform(bounds.Min)
	widget.DrawChild(r.mid, ctx, canvas)
	widget.DrawChild(r.left, ctx, canvas)
	widget.DrawChild(r.right, ctx, canvas)
	canvas.PopTransform()
}

func (r *opinionatedSplitRow) Event(ctx widget.Context, e event.Event) bool {
	mouse, ok := e.(*event.MouseEvent)
	if !ok || !r.IsVisible() || !r.IsEnabled() {
		return false
	}
	local := *mouse
	local.Position = mouse.Position.Sub(r.Bounds().Min)
	for _, child := range []widget.Widget{r.right, r.left, r.mid} {
		bounded, ok := child.(interface{ Bounds() geometry.Rect })
		if child == nil || !ok || !bounded.Bounds().Contains(local.Position) {
			continue
		}
		return child.Event(ctx, &local)
	}
	return false
}

func (r *opinionatedSplitRow) Children() []widget.Widget {
	return []widget.Widget{r.left, r.mid, r.right}
}
