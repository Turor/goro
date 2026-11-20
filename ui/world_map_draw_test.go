package ui

import (
	"image"
	"testing"

	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
	"github.com/kivutar/goro/res"
)

func TestWorldMapHoverDoesNotRedrawTheBitmap(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 32, 24))
	layer := newWorldMapImage(img)
	board := &worldMapWidget{
		image: img, imageLayer: layer, sourceSize: img.Bounds().Size(),
		entries: []res.WorldMapEntry{{MapName: "prontera", Bounds: image.Rect(0, 0, 16, 12)}},
		hovered: -1,
	}
	layer.SetParent(board)
	board.SetRepaintBoundary(true)
	board.Layout(widget.NewContext(), geometry.Tight(geometry.Sz(32, 24)))
	layer.ClearRedraw()
	board.hovered = 0
	board.SetNeedsRedraw(true)
	if layer.NeedsRedraw() {
		t.Fatal("hover marked the world map bitmap dirty")
	}
	if !layer.IsRepaintBoundary() {
		t.Fatal("world map bitmap is not cached")
	}
	if len(board.Children()) != 1 || board.Children()[0] != layer {
		t.Fatal("bitmap is not a child of the overlay")
	}
}
