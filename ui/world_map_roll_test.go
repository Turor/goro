package ui

import (
	"image"
	"testing"

	"github.com/gogpu/ui/geometry"
	"github.com/kivutar/goro/res"
)

func TestWorldMapRollHitsGappedCells(t *testing.T) {
	board := rollBoard([]res.WorldMapEntry{
		{MapName: "a", Bounds: image.Rect(0, 0, 10, 10)},
		{MapName: "b", Bounds: image.Rect(11, 0, 21, 10)},
		{MapName: "c", Bounds: image.Rect(22, 0, 32, 10)},
	})
	if got := board.mapAt(geometry.Pt(5, 5)); got != 0 {
		t.Fatalf("first cell = %d", got)
	}
	if got := board.mapAt(geometry.Pt(10.5, 5)); got < 0 {
		t.Fatalf("gap between cells = %d, want a cell", got)
	}
	if got := board.mapAt(geometry.Pt(16, 5)); got != 1 {
		t.Fatalf("middle cell = %d", got)
	}
	if got := board.mapAt(geometry.Pt(27, 5)); got != 2 {
		t.Fatalf("last cell = %d", got)
	}
	wide := rollBoard([]res.WorldMapEntry{
		{MapName: "a", Bounds: image.Rect(0, 0, 10, 10)},
		{MapName: "b", Bounds: image.Rect(22, 0, 32, 10)},
	})
	if got := wide.mapAt(geometry.Pt(16, 5)); got < 0 {
		t.Fatalf("wide gap = %d, want a neighboring cell", got)
	}
}

func TestWorldMapOverlapPrefersLaterCell(t *testing.T) {
	board := rollBoard([]res.WorldMapEntry{
		{MapName: "field", Bounds: image.Rect(0, 0, 40, 40)},
		{MapName: "town", Bounds: image.Rect(10, 10, 25, 25)},
	})
	if got := board.mapAt(geometry.Pt(15, 15)); got != 1 {
		t.Fatalf("overlap = %d, want the later town", got)
	}
	if got := board.mapAt(geometry.Pt(2, 2)); got != 0 {
		t.Fatalf("field = %d", got)
	}
}

func TestWorldMapHighlightStaysOnTheCell(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 64, 48))
	board := &worldMapWidget{
		image: img, sourceSize: img.Bounds().Size(), hovered: 1,
		entries: []res.WorldMapEntry{
			{MapName: "a", Bounds: image.Rect(0, 0, 10, 10)},
			{MapName: "b", Bounds: image.Rect(20, 10, 30, 20)},
		},
	}
	board.SetBounds(geometry.NewRect(0, 0, 64, 48))
	_ = newWorldMapLayers(board)
	board.placeHighlight()
	cell := board.mapRect(board.entries[1])
	if board.highlight.Bounds() != cell {
		t.Fatalf("highlight = %v, want cell %v", board.highlight.Bounds(), cell)
	}
	if board.highlight.Bounds().Width() >= board.Bounds().Width() {
		t.Fatal("highlight is the whole map")
	}
}

func rollBoard(entries []res.WorldMapEntry) *worldMapWidget {
	img := image.NewRGBA(image.Rect(0, 0, 64, 48))
	board := &worldMapWidget{image: img, sourceSize: img.Bounds().Size(), entries: entries}
	board.SetBounds(geometry.NewRect(0, 0, 64, 48))
	return board
}
