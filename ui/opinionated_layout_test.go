package ui

import (
	"testing"

	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
	"github.com/kivutar/goro/db"
	"github.com/kivutar/goro/session"
)

func TestOpinionatedSplitIsSixAndSix(t *testing.T) {
	if opinionatedBagCols != opinionatedSplitCols*2+1 {
		t.Fatalf("bag %d is not 6 + middle + 6", opinionatedBagCols)
	}
	if opinionatedSplitContentW() != opinionatedSplitCols*inventoryBagCell {
		t.Fatalf("left pane %d, want %d cells", opinionatedSplitContentW(), opinionatedSplitCols)
	}
	if opinionatedSplitContentWRight() != opinionatedSplitCols*inventoryBagCell {
		t.Fatalf("right pane %d, want %d cells", opinionatedSplitContentWRight(), opinionatedSplitCols)
	}
	if opinionatedMiddleW() != inventoryBagCell {
		t.Fatalf("middle column %d, want one cell", opinionatedMiddleW())
	}
	if opinionatedDividerW() >= opinionatedMiddleW() {
		t.Fatalf("divider %d fills the middle column %d", opinionatedDividerW(), opinionatedMiddleW())
	}
	if opinionatedRightGridX() != (opinionatedSplitCols+1)*inventoryBagCell {
		t.Fatalf("right grid starts at %d, want the 8th inventory column", opinionatedRightGridX())
	}
	if opinionatedFrameW() != opinionatedContentW()+ROScrollbarGutter {
		t.Fatalf("frame %d is not the 13-column grid plus one gutter", opinionatedFrameW())
	}
}

func TestOpinionatedSourceTabsShareTheWidth(t *testing.T) {
	for _, count := range []int{2, 3, 4} {
		sum := 0
		for i := 0; i < count; i++ {
			sum += opinionatedEqualTabWidth(i, count)
		}
		if sum != opinionatedFrameW() {
			t.Fatalf("%d tabs sum to %d, want %d", count, sum, opinionatedFrameW())
		}
	}
}

func TestOpinionatedSourcesFollowCartAndGuild(t *testing.T) {
	window := &InventoryBagWindow{}
	ctx := Context{Session: &session.Session{OpinionatedUI: true}}
	if got := len(window.opinionatedSources(ctx)); got != 2 {
		t.Fatalf("sources = %d, want inventory and storage", got)
	}
	ctx.Session.Cart.MaxAmount = 1
	if got := len(window.opinionatedSources(ctx)); got != 3 {
		t.Fatalf("sources with a cart = %d, want 3", got)
	}
	ctx.Session.GuildID = 7
	if got := len(window.opinionatedSources(ctx)); got != 4 {
		t.Fatalf("sources with a guild = %d, want 4", got)
	}
	window.source = opinionatedSourceGuild
	ctx.Session.GuildID = 0
	window.normalizeOpinionatedSource(ctx)
	if window.source != opinionatedSourceBag {
		t.Fatalf("hidden guild tab left source %d", window.source)
	}
}

func TestOpinionatedCategoryTabsAreOneRow(t *testing.T) {
	row := (&InventoryBagWindow{}).filterRows(Context{Session: &session.Session{OpinionatedUI: true}})
	row.Layout(widget.NewContext(), geometry.Tight(geometry.Sz(float32(opinionatedFrameW()), float32(inventoryBagFilterH))))
	children := row.Children()
	if len(children) == 1 {
		if _, ok := children[0].(*tabWidget); !ok {
			children = children[0].Children()
		}
	}
	seen := 0
	for _, child := range children {
		tab, ok := child.(*tabWidget)
		if !ok {
			t.Fatalf("category child = %T", child)
		}
		if tab.cfg.label != opinionatedInventoryTabs[seen].label {
			t.Fatalf("category %d = %q", seen, tab.cfg.label)
		}
		seen++
	}
	if seen != len(opinionatedInventoryTabs) {
		t.Fatalf("saw %d categories, want %d", seen, len(opinionatedInventoryTabs))
	}
}

func TestOpinionatedSourceTabsUseTheStackStyle(t *testing.T) {
	window := &InventoryBagWindow{source: opinionatedSourceStorage}
	strip := window.sourceTabs(Context{Session: &session.Session{OpinionatedUI: true}})
	strip.Layout(widget.NewContext(), geometry.Tight(geometry.Sz(float32(opinionatedFrameW()), float32(inventoryBagFilterH))))
	tabs := strip.Children()
	if len(tabs) != 2 {
		t.Fatalf("source tabs = %d, want inventory and storage", len(tabs))
	}
	active, ok := tabs[1].(*tabWidget)
	if !ok || !active.cfg.active || !active.cfg.omitBottom {
		t.Fatalf("storage tab = %#v, want the front tab with no bottom border", tabs[1])
	}
	if active.cfg.activeFill != inventoryBagTabActive {
		t.Fatal("selected tab does not use the theme highlight")
	}
}

func TestOpinionatedGridFillsAWiderPane(t *testing.T) {
	grid := newInventoryGridWidget(inventoryGridConfig{
		cols:      opinionatedBagCols,
		fillWidth: true,
		cellSize:  inventoryBagCell,
		viewWidth: opinionatedBagCols * inventoryBagCell,
		items:     []session.InventoryItem{{Index: 1, ItemID: 1, Type: db.ItemTypeEtc, Amount: 2}},
	})
	width := float32(opinionatedBagCols * inventoryBagCell * 2)
	grid.Layout(widget.NewContext(), geometry.Tight(geometry.Sz(width, 200)))
	if grid.laidCell < float32(inventoryBagCell)*1.5 {
		t.Fatalf("cell = %v, want the pane width divided across the columns", grid.laidCell)
	}
}

func TestOpinionatedPaneHitUsesTheSplit(t *testing.T) {
	window := &InventoryBagWindow{source: opinionatedSourceStorage}
	width, height := inventoryBagFrameSize(true)
	window.EnsureWindow(width, height)
	window.x, window.y = 10, 20
	bodyY := window.y + inventoryOpinionatedChromeH() + 4
	if got := window.opinionatedPaneAt(window.x+4, bodyY); got != opinionatedPaneBag {
		t.Fatalf("left pane = %d", got)
	}
	right := window.x + opinionatedSideViewW() + opinionatedMiddleW() + 4
	if got := window.opinionatedPaneAt(right, bodyY); got != opinionatedPaneOther {
		t.Fatalf("right pane = %d", got)
	}
	mid := window.x + opinionatedSideViewW() + 1
	if got := window.opinionatedPaneAt(mid, bodyY); got != -1 {
		t.Fatalf("middle column = %d, want no pane", got)
	}
}

func TestOpinionatedSplitDividerSitsInTheMiddleColumn(t *testing.T) {
	window := &InventoryBagWindow{source: opinionatedSourceStorage, opinionated: true}
	panes := window.opinionatedPanes(Context{Session: &session.Session{OpinionatedUI: true}}, nil)
	panes.Layout(widget.NewContext(), geometry.Tight(geometry.Sz(float32(opinionatedFrameW()), float32(inventoryBagViewH))))
	children := panes.Children()
	if len(children) != 3 {
		t.Fatalf("split children = %d, want left, middle, right", len(children))
	}
	center, ok := children[1].(*opinionatedCenterColumn)
	if !ok {
		t.Fatalf("middle child = %T, want the divider column", children[1])
	}
	if center.Bounds().Width() != float32(opinionatedMiddleW()) {
		t.Fatalf("middle width %v, want one cell", center.Bounds().Width())
	}
	if len(center.Children()) != 0 {
		t.Fatalf("middle column holds %d controls, want only the divider", len(center.Children()))
	}
	placed, ok := children[2].(interface{ Bounds() geometry.Rect })
	if !ok || placed.Bounds().Min.X != float32(opinionatedRightGridX()) {
		t.Fatalf("right grid did not stay on the inventory columns")
	}
	if opinionatedBodyH() != inventoryBagViewH+inventoryOpinionatedChromeH() {
		t.Fatalf("window height %d stretched past the rows", opinionatedBodyH())
	}
}

func TestClosedOpinionatedSearchDoesNotBlockShortcuts(t *testing.T) {
	window := &InventoryBagWindow{opinionated: true, searchArmed: true}
	if window.KeyboardShortcutsBlocked() {
		t.Fatal("a closed bag blocked the inventory hotkeys")
	}
}
