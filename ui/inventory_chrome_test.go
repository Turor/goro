package ui

import (
	"testing"

	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
	"github.com/kivutar/goro/input"
	"github.com/kivutar/goro/session"
)

func TestOpinionatedGridsKeepTheSameCellScale(t *testing.T) {
	full := newInventoryGridWidget(inventoryGridConfig{
		cols:      opinionatedBagCols,
		grid:      true,
		cellSize:  inventoryBagCell,
		viewWidth: opinionatedContentW(),
	})
	split := newInventoryGridWidget(inventoryGridConfig{
		cols:      opinionatedSplitCols,
		grid:      true,
		cellSize:  inventoryBagCell,
		viewWidth: opinionatedSplitContentW(),
	})
	wide := geometry.Tight(geometry.Sz(float32(opinionatedFrameW()), float32(inventoryBagViewH)))
	narrow := geometry.Tight(geometry.Sz(float32(opinionatedSplitContentW()), float32(inventoryBagViewH)))
	full.Layout(widget.NewContext(), wide)
	split.Layout(widget.NewContext(), narrow)
	if full.laidCell != split.laidCell {
		t.Fatalf("full cell %v, split cell %v", full.laidCell, split.laidCell)
	}
	if full.laidCell != float32(inventoryBagCell) {
		t.Fatalf("cell = %v, want the design cell", full.laidCell)
	}
}

func TestOpinionatedTabDragLeavesTheCloseBox(t *testing.T) {
	window := &InventoryBagWindow{opinionated: true}
	window.width = opinionatedFrameW()
	window.syncOpinionatedDrag()
	if window.dragBounds.Min.Y != 0 || window.dragBounds.Height() != float32(inventoryBagFilterH) {
		t.Fatalf("drag = %v, want the tab row", window.dragBounds)
	}
	if window.dragBounds.Max.X > float32(window.width-inventoryBagFilterH) {
		t.Fatalf("drag covers the close box: %v", window.dragBounds)
	}
}

func TestOpinionatedSearchTypesWithoutTheField(t *testing.T) {
	window := &InventoryBagWindow{opinionated: true, searchArmed: true}
	ctx := Context{Session: &session.Session{OpinionatedUI: true}, Input: input.NewState()}
	window.searchCtx = ctx
	window.ensureSearchField()
	ctx.Input.AddTextInput("ab")
	window.syncSearchTyping(ctx)
	if window.search != "ab" {
		t.Fatalf("search = %q, want ab", window.search)
	}
	ctx.Input.EndFrame()
	ctx.Input.SetKey(input.KeyBackspace, true)
	window.syncSearchTyping(ctx)
	if window.search != "a" {
		t.Fatalf("search = %q, want a", window.search)
	}
}

func TestOpinionatedSortButtonFitsTheSearchBar(t *testing.T) {
	window := &InventoryBagWindow{}
	row := window.searchRow(Context{}).(*opinionatedSearchRow)
	width := float32(opinionatedFrameW())
	row.Layout(widget.NewContext(), geometry.Tight(geometry.Sz(width, float32(inventorySearchRowH()))))
	if row.sort.Bounds().Height() > row.search.fieldRect().Height() {
		t.Fatalf("sort height %v is taller than the search bar %v", row.sort.Bounds().Height(), row.search.fieldRect().Height())
	}
	if row.search.Bounds().Max.X > row.sort.Bounds().Min.X {
		t.Fatalf("search %v overlaps the sort button %v", row.search.Bounds(), row.sort.Bounds())
	}
	if row.search.Bounds().Width() >= width-float32(inventoryBagSortW)+1 {
		t.Fatalf("search width %v did not shrink for the sort button", row.search.Bounds().Width())
	}
	if row.sort.Bounds().Max.X > width-float32(inventoryBagSearchInset)+1 {
		t.Fatalf("sort button %v is flush with the row edge %v", row.sort.Bounds(), width)
	}
}
