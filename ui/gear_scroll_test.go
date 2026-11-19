package ui

import (
	"testing"

	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
	"github.com/kivutar/goro/session"
)

func TestGearStatsColumnLeavesRoomForScrollbar(t *testing.T) {
	ctx := Context{Session: &session.Session{}, ScreenW: 800, ScreenH: 600}
	column := gearStatsColumn(ctx)
	column.Layout(widget.NewContext(), geometry.Loose(geometry.Sz(800, 600)))
	bounds, ok := column.(interface{ Bounds() geometry.Rect })
	if !ok {
		t.Fatal("gear stats column has no bounds")
	}
	got := bounds.Bounds().Width()
	want := statsPrimaryColumnWidth + float32(ROScrollbarGutter)
	if got != want {
		t.Fatalf("stats column width = %.1f, want %.1f", got, want)
	}
}
