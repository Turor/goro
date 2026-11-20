package ui

import (
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

// worldMapLayers stacks the cached atlas under a cell highlight and a preview.
// Those two are siblings, not children of the atlas: a child boundary is skipped
// while the atlas records. Each is only as large as what it paints, so a hover
// does not record a texture the size of the map.
type worldMapLayers struct {
	widget.WidgetBase
	board     *worldMapWidget
	highlight *worldMapHighlight
	preview   *worldMapPreview
}

func newWorldMapLayers(board *worldMapWidget) *worldMapLayers {
	highlight := &worldMapHighlight{board: board}
	preview := &worldMapPreview{board: board}
	for _, layer := range []widget.Widget{highlight, preview} {
		layer.(interface{ SetVisible(bool) }).SetVisible(true)
		layer.(interface{ SetEnabled(bool) }).SetEnabled(true)
		layer.(interface{ SetRepaintBoundary(bool) }).SetRepaintBoundary(true)
	}
	board.highlight = highlight
	board.previewLayer = preview
	board.hoverArmed = false
	layers := &worldMapLayers{board: board, highlight: highlight, preview: preview}
	layers.SetVisible(true)
	layers.SetEnabled(true)
	board.SetParent(layers)
	highlight.SetParent(layers)
	preview.SetParent(layers)
	layers.AddChild(board)
	layers.AddChild(highlight)
	layers.AddChild(preview)
	return layers
}

func (s *worldMapLayers) Layout(ctx widget.Context, constraints geometry.Constraints) geometry.Size {
	size := s.board.Layout(ctx, constraints)
	bounds := geometry.FromPointSize(s.Position(), size)
	s.SetBounds(bounds)
	s.board.SetBounds(bounds)
	s.board.placeHighlight()
	return size
}

func (s *worldMapLayers) Draw(ctx widget.Context, canvas widget.Canvas) {
	// The parent positions this stack after Layout. Copy that frame onto the
	// atlas, then place the highlight on the cell. Stretching the highlight to
	// the whole map records a full-size texture on every hover.
	bounds := s.Bounds()
	s.board.SetBounds(bounds)
	s.board.placeHighlight()
	widget.DrawChild(s.board, ctx, canvas)
	widget.DrawChild(s.highlight, ctx, canvas)
	widget.DrawChild(s.preview, ctx, canvas)
}

func (s *worldMapLayers) Event(ctx widget.Context, e event.Event) bool {
	if s.board == nil {
		return false
	}
	return s.board.Event(ctx, e)
}

// worldMapHighlight paints only the hovered cell.
type worldMapHighlight struct {
	widget.WidgetBase
	board *worldMapWidget
}

func (h *worldMapHighlight) Layout(widget.Context, geometry.Constraints) geometry.Size {
	// placeHighlight owns the cell frame. A layout walk must not move it to the origin.
	return h.Bounds().Size()
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
	cell := board.mapRect(board.entries[board.hovered])
	canvas.DrawRect(cell, worldMapHoverFill)
	canvas.StrokeRect(cell, widget.ColorWhite, worldMapHoverStroke)
}

// worldMapPreview paints the minimap panel. It stays empty until the hover has
// settled, so rolling across cells does not decode an image per cell.
type worldMapPreview struct {
	widget.WidgetBase
	board *worldMapWidget
}

func (p *worldMapPreview) Layout(widget.Context, geometry.Constraints) geometry.Size {
	return p.Bounds().Size()
}

func (p *worldMapPreview) Event(ctx widget.Context, e event.Event) bool {
	if p.board == nil {
		return false
	}
	return p.board.Event(ctx, e)
}

func (p *worldMapPreview) Draw(_ widget.Context, canvas widget.Canvas) {
	board := p.board
	if board == nil || board.hovered < 0 || board.hovered >= len(board.entries) {
		return
	}
	board.drawPreview(canvas, board.entries[board.hovered])
}

func (w *worldMapWidget) placeHighlight() {
	if w.highlight == nil {
		return
	}
	if w.hovered < 0 || w.hovered >= len(w.entries) {
		w.highlight.SetBounds(geometry.Rect{})
		if w.previewLayer != nil {
			w.previewLayer.SetBounds(geometry.Rect{})
		}
		return
	}
	entry := w.entries[w.hovered]
	w.highlight.SetBounds(w.mapRect(entry))
	if w.previewLayer != nil {
		w.previewLayer.SetBounds(w.previewPanel(entry))
	}
}

func (w *worldMapWidget) markHoverDirty(ctx widget.Context) {
	var oldCell, oldPreview geometry.Rect
	if w.highlight != nil {
		oldCell = w.highlight.Bounds()
	}
	if w.previewLayer != nil {
		oldPreview = w.previewLayer.Bounds()
	}
	w.placeHighlight()
	w.hoverArmed = false
	if w.highlight != nil {
		w.highlight.SetNeedsRedraw(true)
	}
	if w.previewLayer != nil {
		w.previewLayer.SetNeedsRedraw(true)
	}
	if ctx == nil {
		return
	}
	ctx.InvalidateRect(oldCell)
	ctx.InvalidateRect(oldPreview)
	if w.highlight != nil {
		ctx.InvalidateRect(w.highlight.Bounds())
	}
	if w.previewLayer != nil {
		ctx.InvalidateRect(w.previewLayer.Bounds())
	}
}
