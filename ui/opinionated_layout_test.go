package ui

import (
	"testing"

	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
	"github.com/kivutar/goro/db"
	"github.com/kivutar/goro/session"
)

func TestOpinionatedSplitIsFourAndFour(t *testing.T) {
	if opinionatedSplitCols*2+opinionatedDividerCols != opinionatedBagCols {
		t.Fatalf("split %d+%d+%d, want %d columns", opinionatedSplitCols, opinionatedDividerCols, opinionatedSplitCols, opinionatedBagCols)
	}
	if opinionatedSideViewW()*2+opinionatedDividerW() != opinionatedFrameW() {
		t.Fatalf("panes %d divider %d frame %d", opinionatedSideViewW(), opinionatedDividerW(), opinionatedFrameW())
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
	if len(row.Children()) != 1 {
		t.Fatalf("category rows = %d, want 1", len(row.Children()))
	}
	seen := 0
	for _, child := range row.Children()[0].Children() {
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
	if got := window.opinionatedPaneAt(window.x+4, window.y+windowTitleHeight+inventoryOpinionatedChromeH()+4); got != opinionatedPaneBag {
		t.Fatalf("left pane = %d", got)
	}
	right := window.x + opinionatedSideViewW() + opinionatedDividerW() + 4
	if got := window.opinionatedPaneAt(right, window.y+windowTitleHeight+inventoryOpinionatedChromeH()+4); got != opinionatedPaneOther {
		t.Fatalf("right pane = %d", got)
	}
	mid := window.x + opinionatedSideViewW() + 1
	if got := window.opinionatedPaneAt(mid, window.y+windowTitleHeight+inventoryOpinionatedChromeH()+4); got != -1 {
		t.Fatalf("divider = %d, want no pane", got)
	}
}
