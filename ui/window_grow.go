package ui

import (
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/widget"
)

// growingContent lets the main panes inside a window fill extra space when
// the window is resized. A fixed Width/Height is a minimum; the tallest body
// and the right-hand pane pick up the leftover.
type growingContent struct {
	widget.WidgetBase
	child widget.Widget
}

func newGrowingContent(child widget.Widget) widget.Widget {
	g := &growingContent{child: child}
	g.SetVisible(true)
	g.SetEnabled(true)
	if child != nil {
		type parentSetter interface{ SetParent(widget.Widget) }
		if ps, ok := child.(parentSetter); ok {
			ps.SetParent(g)
		}
	}
	return g
}

func (g *growingContent) Layout(ctx widget.Context, constraints geometry.Constraints) geometry.Size {
	if g.child == nil {
		return geometry.Size{}
	}
	if constraints.HasBoundedWidth() && constraints.HasBoundedHeight() {
		growWindowPanes(g.child, constraints.MaxWidth, constraints.MaxHeight)
	}
	size := widget.LayoutChild(g.child, ctx, constraints)
	if setter, ok := g.child.(interface{ SetBounds(geometry.Rect) }); ok {
		setter.SetBounds(geometry.FromPointSize(geometry.Point{}, size))
	}
	if constraints.HasBoundedWidth() && constraints.HasBoundedHeight() {
		size = geometry.Sz(constraints.MaxWidth, constraints.MaxHeight)
	}
	g.SetBounds(geometry.FromPointSize(g.Position(), size))
	return size
}

func (g *growingContent) Draw(ctx widget.Context, canvas widget.Canvas) {
	if g.child == nil || !g.IsVisible() {
		return
	}
	canvas.PushTransform(g.Bounds().Min)
	widget.StampScreenOrigin(g.child, canvas)
	g.child.Draw(ctx, canvas)
	canvas.PopTransform()
}

func (g *growingContent) Event(ctx widget.Context, ev event.Event) bool {
	if g.child == nil || !g.IsVisible() || !g.IsEnabled() {
		return false
	}
	if me, ok := ev.(*event.MouseEvent); ok {
		local := *me
		local.Position = me.Position.Sub(g.Bounds().Min)
		return g.child.Event(ctx, &local)
	}
	if we, ok := ev.(*event.WheelEvent); ok {
		local := *we
		local.Position = we.Position.Sub(g.Bounds().Min)
		return g.child.Event(ctx, &local)
	}
	return g.child.Event(ctx, ev)
}

func (g *growingContent) Children() []widget.Widget {
	if g.child == nil {
		return nil
	}
	return []widget.Widget{g.child}
}

func growWindowPanes(root widget.Widget, availW, availH float32) {
	growWindowHeight(root, availH)
	growWindowWidth(root, availW)
}

func growWindowHeight(root widget.Widget, availH float32) {
	if availH <= 0 {
		return
	}
	boxes := windowBoxes(root)
	maxH := float32(0)
	for _, box := range boxes {
		if h := box.Style().ExplicitHeight; h > maxH {
			maxH = h
		}
	}
	if maxH < float32(windowMinHeight) || availH <= maxH {
		return
	}
	for _, box := range boxes {
		style := box.Style()
		if style.ExplicitHeight != maxH {
			continue
		}
		token := style.MinHeight
		if token <= 0 || token > maxH {
			token = maxH
			box.MinHeightValue(token)
		}
		if availH > token {
			box.Height(availH)
		}
	}
}

func growWindowWidth(root widget.Widget, availW float32) {
	box, ok := root.(*primitives.BoxWidget)
	if !ok || availW <= 0 {
		return
	}
	var sum float32
	var count int
	var rightmost *primitives.BoxWidget
	for _, child := range box.Children() {
		childBox, ok := child.(*primitives.BoxWidget)
		if !ok {
			continue
		}
		width := childBox.Style().ExplicitWidth
		if width <= 0 {
			continue
		}
		sum += width
		count++
		rightmost = childBox
	}
	if rightmost == nil || count == 0 {
		return
	}
	gaps := box.Style().Gap * float32(count-1)
	extra := availW - sum - gaps
	if extra <= 0 {
		return
	}
	token := rightmost.Style().MinWidth
	current := rightmost.Style().ExplicitWidth
	if token <= 0 || token > current {
		token = current
		rightmost.MinWidthValue(token)
	}
	rightmost.Width(token + extra)
}

func windowBoxes(root widget.Widget) []*primitives.BoxWidget {
	var boxes []*primitives.BoxWidget
	var walk func(widget.Widget)
	walk = func(current widget.Widget) {
		if current == nil {
			return
		}
		if box, ok := current.(*primitives.BoxWidget); ok {
			boxes = append(boxes, box)
		}
		for _, child := range current.Children() {
			walk(child)
		}
	}
	walk(root)
	return boxes
}
