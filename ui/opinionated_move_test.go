package ui

import (
	"testing"
	"time"

	"github.com/kivutar/goro/session"
)

func TestOpinionatedCartDoubleClickDoesNotDrag(t *testing.T) {
	window := &InventoryBagWindow{source: opinionatedSourceCart, lastClickItem: 4, lastClickAt: time.Now()}
	ctx := Context{Session: &session.Session{OpinionatedUI: true}}
	window.startItemDragOrActivate(ctx, session.InventoryItem{Index: 4, ItemID: 12, Amount: 2})
	if window.dragActive {
		t.Fatal("cart double-click started an inventory drag")
	}
	if window.lastClickItem != 0 {
		t.Fatal("withdraw did not consume the double-click")
	}
}

func TestOpinionatedDepositStaysOnTheBag(t *testing.T) {
	window := &InventoryBagWindow{source: opinionatedSourceBag, lastClickItem: 3}
	ctx := Context{Session: &session.Session{OpinionatedUI: true, Inventory: session.Inventory{Items: []session.InventoryItem{{Index: 3, ItemID: 501, Amount: 1}}}}}
	window.depositSelected(ctx, opinionatedSourceStorage)
	if window.source != opinionatedSourceBag {
		t.Fatal("deposit changed the source tab")
	}
}

func TestOpinionatedFavoritesTabKeepsMarkedItems(t *testing.T) {
	window := &InventoryBagWindow{source: opinionatedSourceBag, tab: opinionatedBagTabFavorite}
	ctx := Context{Session: &session.Session{OpinionatedUI: true, Inventory: session.Inventory{Items: []session.InventoryItem{
		{Index: 1, ItemID: 501, Amount: 1},
		{Index: 2, ItemID: 502, Amount: 1, Favorite: true},
	}}}}
	items := window.tabItems(ctx.Session)
	if len(items) != 1 || items[0].Index != 2 {
		t.Fatalf("favorites tab = %+v", items)
	}
}
