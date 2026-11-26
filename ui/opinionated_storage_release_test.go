package ui

import (
	"testing"

	"github.com/kivutar/goro/network"
	"github.com/kivutar/goro/session"
)

func TestOpinionatedStorageTabReleasesTheServerLock(t *testing.T) {
	client, packets := connectedStoreClient(t)
	window := storageDropWindow()
	ctx := Context{
		Session: &session.Session{OpinionatedUI: true, Storage: session.Storage{Open: true, Items: []session.InventoryItem{{Index: 4, ItemID: 501, Amount: 1}}}},
		Network: client,
	}
	window.UpdatePresentation(ctx, nil)
	got := readStorePacket(t, packets)
	want := network.BuildCloseStoragePacketForClientDate(20250716)
	if string(got) != string(want) {
		t.Fatalf("close packet %x, want %x", got, want)
	}
	window.UpdatePresentation(ctx, nil)
	select {
	case extra := <-packets:
		t.Fatalf("released storage twice: %x", extra)
	default:
	}
}
