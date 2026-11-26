package ui

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/gogpu/ui/widget"
	"github.com/kivutar/goro/db"
	"github.com/kivutar/goro/input"
	"github.com/kivutar/goro/network"
	"github.com/kivutar/goro/session"
)

type storeBlocker struct{}

func (storeBlocker) AddOverlay(widget.Widget)     {}
func (storeBlocker) RemoveOverlay(widget.Widget)  {}
func (storeBlocker) Clear()                       {}
func (storeBlocker) PointerBlocked(int, int) bool { return true }

func TestOpinionatedDropOnStorageStoresWhileTheWindowBlocksThePointer(t *testing.T) {
	client, packets := connectedStoreClient(t)
	window := storageDropWindow()
	window.dragActive = true
	window.dragPane = opinionatedPaneBag
	window.dragItem = session.InventoryItem{Index: 7, ItemID: 501, Amount: 3, Type: db.ItemTypeWeapon}
	ctx := Context{
		Session:   &session.Session{OpinionatedUI: true, Storage: session.Storage{Open: true}},
		Input:     &input.State{MouseX: opinionatedSideViewW() + opinionatedMiddleW() + 4, MouseY: inventoryOpinionatedChromeH() + 4},
		Network:   client,
		UIManager: storeBlocker{},
	}
	if !window.UpdateDrag(ctx, nil, nil, nil, nil, nil) {
		t.Fatal("drop on the storage pane was ignored")
	}
	got := readStorePacket(t, packets)
	want := network.BuildStorageMovePacket(network.StorageWherePersonal, network.StorageActionDeposit, 7, 3)
	if string(got) != string(want) {
		t.Fatalf("stored packet %x, want %x", got, want)
	}
}

func TestOpinionatedDropStoresWithoutOpeningStorage(t *testing.T) {
	client, packets := connectedStoreClient(t)
	window := storageDropWindow()
	window.dragActive = true
	window.dragPane = opinionatedPaneBag
	window.dragItem = session.InventoryItem{Index: 7, ItemID: 501, Amount: 1, Type: db.ItemTypeWeapon}
	ctx := Context{
		Session: &session.Session{OpinionatedUI: true},
		Input:   &input.State{MouseX: opinionatedSideViewW() + opinionatedMiddleW() + 4, MouseY: inventoryOpinionatedChromeH() + 4},
		Network: client,
	}
	if !window.UpdateDrag(ctx, nil, nil, nil, nil, nil) {
		t.Fatal("drop on the storage pane was ignored")
	}
	if window.pendingStoreAmount != 0 {
		t.Fatal("drop waited for storage to open")
	}
	got := readStorePacket(t, packets)
	want := network.BuildStorageMovePacket(network.StorageWherePersonal, network.StorageActionDeposit, 7, 1)
	if string(got) != string(want) {
		t.Fatalf("stored packet %x, want %x", got, want)
	}
}

func storageDropWindow() *InventoryBagWindow {
	window := &InventoryBagWindow{source: opinionatedSourceStorage, opinionated: true}
	window.open = true
	return window
}

func connectedStoreClient(t *testing.T) (*network.Client, <-chan []byte) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = ln.Close() })
	packets := make(chan []byte, 4)
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		buf := make([]byte, 256)
		for {
			_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
			n, err := conn.Read(buf)
			if n > 0 {
				packets <- append([]byte(nil), buf[:n]...)
			}
			if err != nil {
				return
			}
		}
	}()
	client := network.NewClient(20250716, false)
	addr := ln.Addr().(*net.TCPAddr)
	if err := client.Connect(context.Background(), "127.0.0.1", addr.Port); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { client.Close() })
	return client, packets
}

func readStorePacket(t *testing.T, packets <-chan []byte) []byte {
	t.Helper()
	select {
	case packet := <-packets:
		return packet
	case <-time.After(2 * time.Second):
		t.Fatal("no packet")
		return nil
	}
}

func indexOf(packet, text []byte) int {
	for i := 0; i+len(text) <= len(packet); i++ {
		if string(packet[i:i+len(text)]) == string(text) {
			return i
		}
	}
	return -1
}
