package ui

import (
	"testing"

	"github.com/kivutar/goro/session"
)

func TestOpinionatedStorageTabDoesNotOpenTheClassicWindow(t *testing.T) {
	bag := &InventoryBagWindow{opinionated: true, source: opinionatedSourceStorage}
	storage := &StorageWindow{}
	ctx := Context{Session: &session.Session{OpinionatedUI: true, Storage: session.Storage{Open: true}}, ScreenW: 800, ScreenH: 600}
	PresentClassicStorage(storage, bag, ctx)
	if storage.IsOpen() {
		t.Fatal("storage tab opened the classic window")
	}
	if !ctx.Session.Storage.Open {
		t.Fatal("hiding the classic window closed server storage")
	}
}

func TestClassicStorageWindowStillOpens(t *testing.T) {
	bag := &InventoryBagWindow{}
	storage := &StorageWindow{}
	ctx := Context{Session: &session.Session{Storage: session.Storage{Open: true}}, ScreenW: 800, ScreenH: 600}
	PresentClassicStorage(storage, bag, ctx)
	if !storage.IsOpen() {
		t.Fatal("classic storage window did not open")
	}
}
