package ui

import (
	"image"
	"testing"

	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
	"github.com/kivutar/goro/res"
)

func TestWorldMapLeaveClearsHoverOnTheBoard(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 64, 48))
	board := &worldMapWidget{
		image: img, sourceSize: img.Bounds().Size(), hovered: 2,
		entries: []res.WorldMapEntry{{MapName: "prontera", Bounds: image.Rect(0, 0, 32, 24)}},
	}
	board.SetVisible(true)
	board.SetEnabled(true)
	board.Layout(widget.NewContext(), geometry.Tight(geometry.Sz(64, 48)))
	if len(board.Children()) != 0 {
		t.Fatal("world map bitmap must not be a child; it steals hover")
	}
	leave := &event.MouseEvent{MouseType: event.MouseLeave, Position: geometry.Pt(1, 1)}
	if !board.Event(widget.NewContext(), leave) {
		t.Fatal("board did not handle leave")
	}
	if board.hovered != -1 {
		t.Fatalf("hover = %d, want cleared", board.hovered)
	}
}
