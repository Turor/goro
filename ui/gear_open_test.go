package ui

import (
	"image"
	"math"
	"testing"

	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/widget"
	"github.com/kivutar/goro/session"
)

func TestGearOpenDefersPortrait(t *testing.T) {
	window := &EquipmentWindow{}
	assets := &countingPreview{}
	itemInfo := &ItemWindows{}
	cart := &CartWindow{}
	ctx := Context{Session: &session.Session{}, ScreenW: 1280, ScreenH: 720}
	window.Toggle(ctx)
	window.Update(ctx, itemInfo, cart, assets)
	if assets.calls != 0 {
		t.Fatalf("open frame loaded the portrait %d times", assets.calls)
	}
	window.Update(ctx, itemInfo, cart, assets)
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
	widths := compositeCellWidths(tree)
	if len(widths) < gearCompositeBandCols*2 {
		t.Fatalf("found %d composite cells, want at least %d", len(widths), gearCompositeBandCols*2)
	}
	for _, cell := range widths {
		if math.Abs(float64(cell-widths[0])) > 1 {
			t.Fatalf("composite cells are not equal width: %v", widths)
		}
	}
}

func compositeCellWidths(root widget.Widget) []float32 {
	var widths []float32
	var walk func(widget.Widget)
	walk = func(current widget.Widget) {
		if current == nil {
			return
		}
		if box, ok := current.(*primitives.BoxWidget); ok && box.Bounds().Width() > float32(gearStatsColumnWidth()) {
			children := box.Children()
			if len(children) != gearCompositeBandCols {
				for _, child := range children {
					walk(child)
				}
				return
			}
			row := make([]float32, 0, len(children))
			for _, child := range children {
				expanded, isExpanded := child.(*primitives.ExpandedWidget)
				if !isExpanded {
					row = nil
					break
				}
				row = append(row, expanded.Bounds().Width())
			}
			if len(row) == gearCompositeBandCols {
				widths = append(widths, row...)
			}
		}
		for _, child := range current.Children() {
			walk(child)
		}
	}
	walk(root)
	return widths
}

type countingPreview struct {
	skillWindowTestRenderer
	calls int
}

func (c *countingPreview) EquipmentPreviewImage(Context, int, int) image.Image {
	c.calls++
	return nil
}
