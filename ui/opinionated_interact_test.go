package ui

import (
	"strings"
	"testing"

	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
	"github.com/kivutar/goro/db"
	"github.com/kivutar/goro/input"
	"github.com/kivutar/goro/session"
)

func TestSubmittedAmountUpdatesTheGrid(t *testing.T) {
	window := &InventoryBagWindow{opinionated: true, source: opinionatedSourceStorage}
	window.bagGrid = newInventoryGridWidget(inventoryGridConfig{
		items: []session.InventoryItem{{Index: 4, ItemID: 501, Amount: 10, Type: db.ItemTypeHealing}},
	})
	window.sideGrid = newInventoryGridWidget(inventoryGridConfig{
		items: []session.InventoryItem{{Index: 2, ItemID: 501, Amount: 1, Type: db.ItemTypeHealing}},
	})
	window.showSubmittedAmount(session.InventoryItem{Index: 4, ItemID: 501, Amount: 10}, opinionatedPaneBag, opinionatedPaneOther, 3)
	if got := window.bagGrid.cfg.items[0].Amount; got != 7 {
		t.Fatalf("bag amount = %d, want 7", got)
	}
	if got := window.sideGrid.cfg.items[0].Amount; got != 4 {
		t.Fatalf("storage amount = %d, want 4", got)
	}
}

func TestTabDragReleaseDoesNotSelectTheTab(t *testing.T) {
	window := &InventoryBagWindow{opinionated: true, source: opinionatedSourceBag}
	window.tabPressed = true
	window.tabPressX = 20
	window.tabPressY = 10
	ctx := Context{ScreenW: 800, ScreenH: 600}
	ctx.Input = &input.State{MouseX: 80, MouseY: 40}
	if window.tabClickAllowed(ctx) {
		t.Fatal("a dragged tab still counted as a click")
	}
	ctx.Input = &input.State{MouseX: 22, MouseY: 11}
	if !window.tabClickAllowed(ctx) {
		t.Fatal("a still click was treated as a drag")
	}
	tab := newTabWidget(tabWidgetConfig{
		label:          "Storage",
		width:          40,
		height:         20,
		clickOnRelease: true,
		allowClick:     func() bool { return false },
		onClick:        func() { t.Fatal("drag release opened the tab") },
	})
	tab.SetBounds(geometry.NewRect(0, 0, 40, 20))
	press := event.NewMouseEvent(event.MousePress, event.ButtonLeft, 0, geometry.Pt(4, 4), geometry.Pt(4, 4), event.ModNone)
	release := event.NewMouseEvent(event.MouseRelease, event.ButtonLeft, 0, geometry.Pt(4, 4), geometry.Pt(4, 4), event.ModNone)
	tab.Event(widget.NewContext(), press)
	tab.Event(widget.NewContext(), release)
}

func TestItemHoverClosesUnlessPinned(t *testing.T) {
	window := &InventoryBagWindow{}
	info := &ItemWindows{}
	item := session.InventoryItem{Index: 3, ItemID: 501, Amount: 2, Identified: true}
	ctx := Context{ScreenW: 800, ScreenH: 600}
	window.showItemHover(ctx, item)
	if window.hoverInfo == nil || !window.hoverInfo.IsOpen() {
		t.Fatal("hover did not open the item window")
	}
	window.hideItemHover(ctx)
	if window.hoverInfo.IsOpen() {
		t.Fatal("mouse off left the item window open")
	}
	window.showItemHover(ctx, item)
	window.pinItemHover(ctx, info, item, 40, 40)
	if window.hoverInfo != nil {
		t.Fatal("pinned window stayed on the hover slot")
	}
	if len(info.descriptions) != 1 || !info.descriptions[0].IsOpen() {
		t.Fatal("right click did not pin the item window")
	}
	window.hideItemHover(ctx)
	if !info.descriptions[0].IsOpen() {
		t.Fatal("mouse off closed the pinned item window")
	}
}

func TestSplitFooterNamesTheTransfer(t *testing.T) {
	window := &InventoryBagWindow{opinionated: true, source: opinionatedSourceStorage}
	root := window.opinionatedWindow(Context{ScreenW: 800, ScreenH: 600}, nil)
	if !widgetHasText(root, opinionatedSplitFooterText) {
		t.Fatal("split inventory has no transfer footer")
	}
	window.source = opinionatedSourceBag
	root = window.opinionatedWindow(Context{ScreenW: 800, ScreenH: 600}, nil)
	if widgetHasText(root, opinionatedSplitFooterText) {
		t.Fatal("footer shows when the bag is not split")
	}
}

func widgetHasText(root widget.Widget, want string) bool {
	var walk func(widget.Widget) bool
	walk = func(current widget.Widget) bool {
		if current == nil {
			return false
		}
		if text, ok := current.(interface{ Content() string }); ok && strings.Contains(text.Content(), want) {
			return true
		}
		for _, child := range current.Children() {
			if walk(child) {
				return true
			}
		}
		return false
	}
	return walk(root)
}
