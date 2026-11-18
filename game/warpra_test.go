package game

import (
	"strings"
	"testing"
	"time"

	"github.com/kivutar/goro/client"
	"github.com/kivutar/goro/game/internal/testutil"
	"github.com/kivutar/goro/input"
	"github.com/kivutar/goro/network"
	"github.com/kivutar/goro/session"
)

func warpraListPacket(t *testing.T, list network.WarpraList) network.Packet {
	t.Helper()
	data := network.EncodeWarpraList(list)
	return network.Packet{ID: network.PacketZCWarpraList, Data: data}
}

func warpraResultPacket(t *testing.T, result network.WarpraResult) network.Packet {
	t.Helper()
	data := network.EncodeWarpraResult(result)
	return network.Packet{ID: network.PacketZCWarpraResult, Data: data}
}

func TestWarpraPacketsAssembleCatalogAndBanners(t *testing.T) {
	ctx := client.Context{Session: &session.Session{}, Input: input.NewState(), UIManager: &worldModeTestUIManager{}, ScreenW: 1024, ScreenH: 768}
	m := &WorldMode{}
	m.bindWarpra(ctx)

	first := warpraListPacket(t, network.WarpraList{
		Zeny: 100, Price: network.WarpraDefaultPrice, Flags: network.WarpraListFlagMore, Total: 2,
		Locs: []network.WarpraLoc{{Kind: network.WarpraKindTown, ID: 0, Map: "prontera", Name: "Prontera", Unlocked: true}},
	})
	if _, stop := m.handleNetworkPacket(ctx, first, time.Now()); stop {
		t.Fatal("list chunk stopped the frame")
	}
	if ctx.Session.Warpra.Ready() {
		t.Fatal("MORE chunk completed the catalog")
	}

	second := warpraListPacket(t, network.WarpraList{
		Zeny: 100, Price: network.WarpraDefaultPrice, Total: 2,
		Locs: []network.WarpraLoc{{Kind: network.WarpraKindField, ID: 1, Map: "prt_fild08", Name: "Prontera Field 8"}},
	})
	m.handleNetworkPacket(ctx, second, time.Now())
	if !ctx.Session.Warpra.Ready() || len(ctx.Session.Warpra.Locs) != 2 {
		t.Fatalf("catalog = %+v", ctx.Session.Warpra)
	}

	m.handleNetworkPacket(ctx, warpraResultPacket(t, network.WarpraResult{
		Result: network.WarpraNoZeny, Kind: uint16(network.WarpraKindField), ID: 1,
	}), time.Now())
	if messages := m.ui.console.Messages(); len(messages) == 0 || !strings.Contains(messages[len(messages)-1].Text, "50,000,000") {
		t.Fatalf("no-zeny banner = %+v", m.ui.console.Messages())
	}

	m.handleNetworkPacket(ctx, warpraResultPacket(t, network.WarpraResult{
		Result: network.WarpraBusy,
	}), time.Now())
	if messages := m.ui.console.Messages(); !strings.Contains(messages[len(messages)-1].Text, "cannot warp") {
		t.Fatalf("busy banner = %+v", messages)
	}
}

func TestWarpraBuyThenWarpSendsCZRequest(t *testing.T) {
	netClient, server := testutil.NewConnection(t, 20080910)
	ctx := client.Context{Network: netClient, Session: &session.Session{}, Input: input.NewState(), UIManager: &worldModeTestUIManager{}, ScreenW: 1024, ScreenH: 768}
	m := &WorldMode{}
	m.bindWarpra(ctx)
	loc := network.WarpraLoc{Kind: network.WarpraKindTown, ID: 2, Map: "geffen", Name: "Geffen"}
	m.ui.worldMap.UnlockOrWarp(ctx, loc, true)
	m.sendWarpraBuy(ctx, loc)
	readBotTestPackets(t, server, network.BuildWarpraRequestPacket(network.WarpraActBuy, uint16(loc.Kind), loc.ID))

	m.handleNetworkPacket(ctx, warpraResultPacket(t, network.WarpraResult{
		Result: network.WarpraOK, Kind: uint16(loc.Kind), ID: loc.ID,
	}), time.Now())
	readBotTestPackets(t, server, network.BuildWarpraRequestPacket(network.WarpraActWarp, uint16(loc.Kind), loc.ID))
	if messages := m.ui.console.Messages(); len(messages) == 0 || !strings.Contains(messages[len(messages)-1].Text, "Unlocked") {
		t.Fatalf("unlock banner = %+v", m.ui.console.Messages())
	}

	m.handleNetworkPacket(ctx, warpraListPacket(t, network.WarpraList{
		Zeny: 0, Price: network.WarpraDefaultPrice, Total: 1,
		Locs: []network.WarpraLoc{{Kind: loc.Kind, ID: loc.ID, Map: loc.Map, Name: loc.Name, Unlocked: true}},
	}), time.Now())
	m.handleNetworkPacket(ctx, warpraResultPacket(t, network.WarpraResult{
		Result: network.WarpraOK, Kind: uint16(loc.Kind), ID: loc.ID,
	}), time.Now())
	if m.ui.worldMap.IsOpen() {
		t.Fatal("world map stayed open after a successful warp")
	}
}

func TestWarpraLevelGateRefusesBuy(t *testing.T) {
	ctx := client.Context{Session: &session.Session{}, Input: input.NewState()}
	m := &WorldMode{}
	m.bindWarpra(ctx)
	loc := network.WarpraLoc{Kind: network.WarpraKindSpecial, ID: 1, Map: "jor_back1", Name: "Jormungandr Dungeon (Lv 180+)"}
	m.ui.worldMap.UnlockOrWarp(ctx, loc, true)
	if _, _, ok := m.ui.worldMap.ConsumePendingBuy(); ok {
		t.Fatal("level-gated destination opened a buy confirm")
	}
	if messages := m.ui.console.Messages(); len(messages) == 0 || !strings.Contains(messages[0].Text, "base level 180") {
		t.Fatalf("level banner = %+v", m.ui.console.Messages())
	}
}

func TestWarpraWindowSurvivesMapTransitionAndMenuOpens(t *testing.T) {
	ctx := client.Context{Session: &session.Session{}, Input: input.NewState(), UIManager: &worldModeTestUIManager{}, ScreenW: 1024, ScreenH: 768}
	m := &WorldMode{}
	m.ui.warpra.OpenWindow(ctx)
	if m.ui.interactionModalOpen() {
		t.Fatal("Warpra list was treated as a modal")
	}
	next := m.nextWorldMode()
	if !next.ui.warpra.IsOpen() {
		t.Fatal("map transition dropped the Warpra window")
	}
	next.rebindPersistentUI(ctx)
	next.ui.warpra.Close()
	next.basicMenuCallbacks(ctx).OnWarpra()
	if !next.ui.warpra.IsOpen() {
		t.Fatal("basic-menu Warpra callback did not open the list")
	}
}

func TestWarpraConfirmBlocksGameplay(t *testing.T) {
	ctx := client.Context{Session: &session.Session{}, Input: input.NewState(), UIManager: &worldModeTestUIManager{}, ScreenW: 1024, ScreenH: 768}
	m := &WorldMode{}
	m.ui.warpra.Confirm.Open(ctx, "Unlock Prontera?", "Spend zeny?", func() {}, func() {})
	if !m.ui.interactionModalOpen() || !m.ui.nonConsoleKeyboardInputBlocked(ctx) {
		t.Fatal("unlock confirm did not block gameplay")
	}
}
