package ui

import (
	"image"
	"testing"

	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
	"github.com/kivutar/goro/res"
)

func TestWorldMapHoverDoesNotRedrawTheAtlas(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 64, 48))
	board := &worldMapWidget{
		image: img, sourceSize: img.Bounds().Size(), hovered: -1,
		entries:  []res.WorldMapEntry{{MapName: "prontera", Bounds: image.Rect(0, 0, 32, 24)}},
		previews: map[string]image.Image{},
	}
	board.SetVisible(true)
	board.SetEnabled(true)
	board.SetRepaintBoundary(true)
	layers := newWorldMapLayers(board)
	layers.Layout(widget.NewContext(), geometry.Tight(geometry.Sz(64, 48)))
	board.SetNeedsRedraw(false)
	board.highlight.SetNeedsRedraw(false)

	move := &event.MouseEvent{MouseType: event.MouseMove, GlobalPosition: geometry.Pt(1, 1)}
	if !board.highlight.Event(widget.NewContext(), move) {
		t.Fatal("highlight did not handle move")
	}
	if board.hovered != 0 {
		t.Fatalf("hover = %d, want 0", board.hovered)
	}
	if board.NeedsRedraw() {
		t.Fatal("hover redrew the atlas")
	}
	if !board.highlight.NeedsRedraw() {
		t.Fatal("hover did not mark the highlight")
	}
	if len(board.Children()) != 0 {
		t.Fatal("atlas must not own the highlight; a child boundary is skipped while recording")
	}
}

func TestWorldMapHighlightLeaveClearsHover(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 64, 48))
	board := &worldMapWidget{
		image: img, sourceSize: img.Bounds().Size(), hovered: 0,
		entries: []res.WorldMapEntry{{MapName: "prontera", Bounds: image.Rect(0, 0, 32, 24)}},
	}
	board.SetVisible(true)
	layers := newWorldMapLayers(board)
	layers.Layout(widget.NewContext(), geometry.Tight(geometry.Sz(64, 48)))
	leave := &event.MouseEvent{MouseType: event.MouseLeave, Position: geometry.Pt(1, 1)}
	if !board.highlight.Event(widget.NewContext(), leave) {
		t.Fatal("highlight did not handle leave")
	}
	if board.hovered != -1 {
		t.Fatalf("hover = %d, want cleared", board.hovered)
	}
	if board.NeedsRedraw() {
		t.Fatal("leave redrew the atlas")
	}
}
