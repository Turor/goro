package ui

import (
	"image"
	"testing"

	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/widget"
	"github.com/kivutar/goro/session"
)

func TestGearOpenDefersPortrait(t *testing.T) {
	window := &EquipmentWindow{}
	assets := &countingPreview{}
	ctx := Context{Session: &session.Session{}, ScreenW: 1280, ScreenH: 720}
	window.Toggle(ctx)
	window.Update(ctx, nil, nil, assets)
	if assets.calls != 0 {
		t.Fatalf("open frame loaded the portrait %d times", assets.calls)
	}
	window.Update(ctx, nil, nil, assets)
	if assets.calls != 1 {
		t.Fatalf("next frame portrait calls = %d, want 1", assets.calls)
	}
}

func TestGearCompositeCellsUseTheWindowWidth(t *testing.T) {
	window := &EquipmentWindow{}
	ctx := Context{Session: &session.Session{}, ScreenW: 1280, ScreenH: 900}
	window.Toggle(ctx)
	tree := window.widgetTree(ctx, nil, nil)
	width, height := gearWindowSize()
	tree.Layout(widget.NewContext(), geometry.Tight(geometry.Sz(float32(width), float32(height))))
	if !hasWideStatRow(tree, float32(gearStatsColumnWidth())) {
		t.Fatal("composite row is not wider than the status column")
	}
}

func hasWideStatRow(root widget.Widget, column float32) bool {
	found := false
	var walk func(widget.Widget)
	walk = func(current widget.Widget) {
		if box, ok := current.(*primitives.BoxWidget); ok {
			bounds := box.Bounds()
			if bounds.Height() == float32(statsRowH) && bounds.Width() > column {
				found = true
			}
		}
		for _, child := range current.Children() {
			walk(child)
		}
	}
	walk(root)
	return found
}

type countingPreview struct {
	skillWindowTestRenderer
	calls int
}

func (c *countingPreview) EquipmentPreviewImage(Context, int, int) image.Image {
	c.calls++
	return nil
}
