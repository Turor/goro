package ui

import (
	"testing"

	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/widget"
)

func TestContextMenuUsesCompactInsetRows(t *testing.T) {
	first := contextMenuItem("First", nil)
	second := contextMenuItem("Second", nil)
	wantHeight := contextMenuHeight(2)
	root := contextMenu(100, first, second)
	size := root.Layout(widget.NewContext(), geometry.Tight(geometry.Sz(100, float32(wantHeight))))

	if size != geometry.Sz(100, float32(wantHeight)) {
		t.Fatalf("context menu size = %v, want 100x%d", size, wantHeight)
	}
	if got := first.(interface{ Bounds() geometry.Rect }).Bounds(); got != geometry.NewRect(float32(contextMenuPadding), float32(contextMenuPadding), float32(100-2*contextMenuPadding), float32(contextMenuRowH())) {
		t.Fatalf("first context menu row bounds = %v, want inset compact row", got)
	}
	if got := second.(interface{ Bounds() geometry.Rect }).Bounds(); got != geometry.NewRect(float32(contextMenuPadding), float32(contextMenuPadding+contextMenuRowH()), float32(100-2*contextMenuPadding), float32(contextMenuRowH())) {
		t.Fatalf("second context menu row bounds = %v, want adjacent inset compact row", got)
	}
	frame, ok := root.(*windowFrame)
	box := (*primitives.BoxWidget)(nil)
	if ok {
		box = frame.BoxWidget
	}
	if !ok {
		t.Fatalf("context menu root = %T, want box", root)
	}
	if box.Style().Radius != float32(contextMenuRadius) {
		t.Fatalf("context menu radius = %.1f, want %.1f", box.Style().Radius, float32(contextMenuRadius))
	}
}
