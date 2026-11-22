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
