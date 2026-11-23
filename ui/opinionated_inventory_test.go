package ui

import (
	"testing"

	"github.com/kivutar/goro/db"
	"github.com/kivutar/goro/session"
)

func TestOpinionatedInventorySplitsAmmoAndCards(t *testing.T) {
	if got := opinionatedInventoryTab(session.InventoryItem{Type: db.ItemTypeAmmo}); got != opinionatedBagTabAmmo {
		t.Fatalf("ammo tab = %d, want ammo", got)
	}
	if got := opinionatedInventoryTab(session.InventoryItem{Type: db.ItemTypeCard, Equip: true}); got != opinionatedBagTabCard {
		t.Fatalf("card tab = %d, want cards", got)
	}
	if got := inventoryItemTab(session.InventoryItem{Type: db.ItemTypeAmmo}); got != inventoryBagTabEtc {
		t.Fatalf("classic ammo tab = %d, want etc", got)
	}
	window := InventoryBagWindow{}
	s := &session.Session{OpinionatedUI: true, Inventory: session.Inventory{Items: []session.InventoryItem{
		{Index: 1, Type: db.ItemTypeHealing},
		{Index: 2, Type: db.ItemTypeAmmo},
	}}}
	window.syncStyle(s)
	if window.tab != opinionatedBagTabAll {
		t.Fatalf("opinionated open tab = %d, want all", window.tab)
	}
	if got := len(window.tabItems(s)); got != 2 {
		t.Fatalf("all tab items = %d, want 2", got)
	}
	width, height := inventoryBagFrameSize(true)
	if height <= inventoryBagHeight || width != inventoryBagWidth {
		t.Fatalf("opinionated frame = %d,%d classic height %d", width, height, inventoryBagHeight)
	}
}

func TestOpinionatedInventorySearchSortAndSources(t *testing.T) {
	window := &InventoryBagWindow{search: "item 12", sort: opinionatedSortName}
	s := &session.Session{
		OpinionatedUI: true,
		Inventory: session.Inventory{Items: []session.InventoryItem{
			{Index: 2, ItemID: 12, Type: db.ItemTypeEtc},
			{Index: 1, ItemID: 9, Type: db.ItemTypeEtc},
		}},
		Cart: session.Cart{Items: []session.InventoryItem{{Index: 4, ItemID: 12, Type: db.ItemTypeEtc, Amount: 3}}},
	}
	window.syncStyle(s)
	found := window.tabItems(s)
	if len(found) != 1 || found[0].ItemID != 12 {
		t.Fatalf("search found %+v, want item 12", found)
	}
	window.search = ""
	window.sort = opinionatedSortName
	window.tab = opinionatedBagTabAll
	named := window.tabItems(s)
	if len(named) != 2 || named[0].ItemID != 12 {
		t.Fatalf("name sort = %+v, want item 12 first", named)
	}
	window.source = opinionatedSourceCart
	cart := window.tabItems(s)
	if len(cart) != 1 || cart[0].Index != 4 {
		t.Fatalf("cart source = %+v, want cart index 4", cart)
	}
	classic := &InventoryBagWindow{search: "item 12", tab: inventoryBagTabEtc}
	plain := &session.Session{Inventory: s.Inventory}
	if got := classic.tabItems(plain); len(got) != 2 {
		t.Fatalf("classic search leaked, got %d items", len(got))
	}
}
