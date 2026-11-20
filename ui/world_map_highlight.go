package ui

import (
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

// worldMapLayers stacks the cached atlas under a highlight that can redraw
// on its own. The highlight is a sibling, not a child of the atlas: a child
// boundary is skipped while the atlas records, and is not composited afterwards.
type worldMapLayers struct {
	widget.WidgetBase
	board     *worldMapWidget
	highlight *worldMapHighlight
}

func newWorldMapLayers(board *worldMapWidget) *worldMapLayers {
	highlight := &worldMapHighlight{board: board}
	highlight.SetVisible(true)
	highlight.SetEnabled(true)
	highlight.SetRepaintBoundary(true)
	board.highlight = highlight
	layers := &worldMapLayers{board: board, highlight: highlight}
	layers.SetVisible(true)
	layers.SetEnabled(true)
	board.SetParent(layers)
	highlight.SetParent(layers)
	layers.AddChild(board)
	layers.AddChild(highlight)
	return layers
}

func (s *worldMapLayers) Layout(ctx widget.Context, constraints geometry.Constraints) geometry.Size {
	size := s.board.Layout(ctx, constraints)
	bounds := geometry.FromPointSize(s.Position(), size)
	s.SetBounds(bounds)
	s.board.SetBounds(bounds)
	s.highlight.SetBounds(bounds)
	return size
}

func (s *worldMapLayers) Draw(ctx widget.Context, canvas widget.Canvas) {
	// The parent positions this stack after Layout, so copy that frame onto
	// both layers before they are stamped. Otherwise the map is hit-tested
	// at the window origin and the click lands on the region rail.
	bounds := s.Bounds()
	s.board.SetBounds(bounds)
	s.highlight.SetBounds(bounds)
	widget.DrawChild(s.board, ctx, canvas)
	widget.DrawChild(s.highlight, ctx, canvas)
}

func (s *worldMapLayers) Event(ctx widget.Context, e event.Event) bool {
	if s.board == nil {
		return false
	}
	return s.board.Event(ctx, e)
}

// worldMapHighlight draws only the hover cell and the preview. It is its own
// repaint boundary, so a hover does not record the atlas again.
type worldMapHighlight struct {
	widget.WidgetBase
	board *worldMapWidget
}

func (h *worldMapHighlight) Layout(_ widget.Context, constraints geometry.Constraints) geometry.Size {
	size := constraints.Constrain(geometry.Sz(h.Bounds().Width(), h.Bounds().Height()))
	h.SetBounds(geometry.FromPointSize(h.Position(), size))
	return size
}

func (h *worldMapHighlight) Event(ctx widget.Context, e event.Event) bool {
	if h.board == nil {
		return false
	}
	return h.board.Event(ctx, e)
}

func (h *worldMapHighlight) Draw(_ widget.Context, canvas widget.Canvas) {
	board := h.board
	if board == nil || board.hovered < 0 || board.hovered >= len(board.entries) {
		return
	}
	entry := board.entries[board.hovered]
	cell := board.mapRect(entry)
	canvas.DrawRect(cell, worldMapHoverFill)
	canvas.StrokeRect(cell, widget.ColorWhite, worldMapHoverStroke)
	board.drawPreview(canvas, entry)
}

func (w *worldMapWidget) markHoverDirty(ctx widget.Context) {
	if w.highlight != nil {
		w.highlight.SetNeedsRedraw(true)
		if ctx != nil {
			ctx.InvalidateRect(w.highlight.Bounds())
		}
		return
	}
	w.SetNeedsRedraw(true)
	if ctx != nil {
		ctx.InvalidateRect(w.Bounds())
	}
}
