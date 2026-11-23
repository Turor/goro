package ui

import (
	"testing"

	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
	"github.com/kivutar/goro/session"
)

func TestOpinionatedSearchKeepsFocusAndTiles(t *testing.T) {
	window := &InventoryBagWindow{}
	ctx := Context{Session: &session.Session{OpinionatedUI: true, Inventory: session.Inventory{Items: []session.InventoryItem{
		{Index: 1, ItemID: 12, Amount: 1, Identified: true},
		{Index: 2, ItemID: 34, Amount: 1, Identified: true},
	}}}}
	window.searchCtx = ctx
	field := window.ensureSearchField()
	field.SetFocused(true)
	grid := newInventoryGridWidget(inventoryGridConfig{cols: opinionatedBagCols, viewWidth: opinionatedContentW(), fillWidth: true})
	grid.Layout(widget.NewContext(), geometry.Tight(geometry.Sz(float32(opinionatedContentW()), float32(inventoryBagViewH))))
	cell := grid.laidCell
	window.bagGrid = grid
	window.ensureScrollSignal().Set(12)
	window.snapshot = window.inventorySnapshot(ctx.Session)
	window.applySearch(ctx, "12")
	if window.searchField != field || !field.IsFocused() {
		t.Fatal("typing replaced the search field or dropped focus")
	}
	if window.bagGrid != grid {
		t.Fatal("typing rebuilt the inventory grid")
	}
	if window.ensureScrollSignal().Get() != 12 {
		t.Fatal("typing moved the inventory scroll")
	}
	if grid.laidCell != cell {
		t.Fatalf("cell moved from %v to %v", cell, grid.laidCell)
	}
	if len(grid.cfg.items) != 1 || grid.cfg.items[0].ItemID != 12 {
		t.Fatalf("filter items = %+v", grid.cfg.items)
	}
	if window.snapshot != window.inventorySnapshot(ctx.Session) {
		t.Fatal("search left the snapshot stale, so the next frame would rebuild")
	}
	box := newInventorySearchBox(window)
	box.Layout(widget.NewContext(), geometry.Tight(geometry.Sz(float32(opinionatedFrameW()), float32(inventoryBagSearchH))))
	if !box.clearRect().Contains(geometry.Pt(box.Bounds().Max.X-1, 1)) {
		t.Fatal("clear mark is not on the right of the search box")
	}
}

func TestOpinionatedCategoryTabsFillTheRow(t *testing.T) {
	row := (&InventoryBagWindow{}).filterRows(Context{})
	width := float32(opinionatedFrameW() * 2)
	row.Layout(widget.NewContext(), geometry.Tight(geometry.Sz(width, float32(inventoryBagFilterH))))
	children := row.Children()
	if len(children) != len(opinionatedInventoryTabs) {
		t.Fatalf("category tabs = %d, want %d on one row", len(children), len(opinionatedInventoryTabs))
	}
	last, ok := children[len(children)-1].(*tabWidget)
	if !ok {
		t.Fatalf("last category = %T", children[len(children)-1])
	}
	if last.Bounds().Max.X < width-1 {
		t.Fatalf("last category ends at %v, want the row width %v", last.Bounds().Max.X, width)
	}
}

func TestResizeLocksTheDesignAspect(t *testing.T) {
	ctx := Context{ScreenW: 1600, ScreenH: 1200}
	window := NewWindow(300, 200)
	window.resizeTo(ctx, 480, 200)
	if window.width*200 != window.height*300 {
		t.Fatalf("size = %dx%d, aspect drifted from 300x200", window.width, window.height)
	}
	if window.width <= 300 {
		t.Fatalf("width = %d, drag did not grow the window", window.width)
	}
}
