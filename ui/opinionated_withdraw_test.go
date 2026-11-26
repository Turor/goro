package ui

import (
	"testing"

	"github.com/kivutar/goro/input"
	"github.com/kivutar/goro/network"
	"github.com/kivutar/goro/session"
)

func TestOpinionatedWithdrawQueuesUntilStorageOpens(t *testing.T) {
	client, packets := connectedStoreClient(t)
	window := storageDropWindow()
	window.dragActive = true
	window.dragPane = opinionatedPaneOther
	window.dragItem = session.InventoryItem{Index: 4, ItemID: 501, Amount: 1, Type: 5}
	ctx := Context{
		Session: &session.Session{
			OpinionatedUI: true,
			Storage:       session.Storage{Items: []session.InventoryItem{{Index: 4, ItemID: 501, Amount: 1}}},
		},
		Input:   &input.State{MouseX: 4, MouseY: inventoryOpinionatedChromeH() + 4},
		Network: client,
	}
	if !window.UpdateDrag(ctx, nil, nil, nil, nil, nil) {
		t.Fatal("drop on the inventory pane was ignored")
	}
	if window.pendingWithdraw.Index != 4 {
		t.Fatalf("pending withdraw = %+v", window.pendingWithdraw)
	}
	opened := readStorePacket(t, packets)
	if indexOf(opened, []byte("@storage")) < 0 {
		t.Fatalf("open packet %q, want @storage", opened)
	}
	ctx.Session.Storage.Open = true
	window.flushPendingStore(ctx)
	got := readStorePacket(t, packets)
	want := network.BuildMoveFromStoragePacketForClientDate(4, 1, 20250716)
	if string(got) != string(want) {
		t.Fatalf("withdraw packet %x, want %x", got, want)
	}
}

func TestOpinionatedItemChangeRefreshesTheOpenGrid(t *testing.T) {
	window := storageDropWindow()
	window.width = opinionatedFrameW()
	window.height = opinionatedBodyH()
	window.bagGrid = newInventoryGridWidget(inventoryGridConfig{})
	window.sideGrid = newInventoryGridWidget(inventoryGridConfig{})
	sessionState := &session.Session{OpinionatedUI: true}
	ctx := Context{Session: sessionState}
	window.chromeKey = window.inventoryChromeKey()
	window.snapshot = window.inventorySnapshot(sessionState)
	sessionState.Inventory.Items = []session.InventoryItem{{Index: 8, ItemID: 501, Amount: 1}}
	window.UpdatePresentation(ctx, nil)
	if len(window.bagGrid.cfg.items) != 1 || window.bagGrid.cfg.items[0].Index != 8 {
		t.Fatalf("bag grid = %+v, want the withdrawn item on this update", window.bagGrid.cfg.items)
	}
}
