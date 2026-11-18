package ui

import (
	"strings"
	"testing"

	"github.com/kivutar/goro/network"
	"github.com/kivutar/goro/session"
)

func TestWarpraWindowRowsAndUnlock(t *testing.T) {
	var catalog session.WarpraCatalog
	catalog.ApplyList(network.WarpraList{
		Zeny: 10, Price: network.WarpraDefaultPrice, Total: 4,
		Locs: []network.WarpraLoc{
			{Kind: network.WarpraKindTown, ID: 0, Map: "prontera", Name: "Prontera", Unlocked: true},
			{Kind: network.WarpraKindField, ID: 0, Map: "prt_fild08", Name: "Prontera Field 8", Group: 1, GroupName: "Prontera"},
			{Kind: network.WarpraKindDungeon, ID: 0, Group: 1, Map: "iz_dun00", Name: "Byalan 1", GroupName: "Byalan"},
			{Kind: network.WarpraKindSpecial, ID: 0, Map: "turbo_room", Name: "Turbo Track"},
		},
	})
	w := &WarpraWindow{catalog: catalog}
	w.refreshRows()
	if len(w.rows) != 1 || !strings.Contains(w.rows[0].label, "Prontera") || !strings.Contains(w.rows[0].label, "Visited") {
		t.Fatalf("towns = %+v", w.rows)
	}
	w.tab = warpraTabField
	w.refreshRows()
	if len(w.rows) != 1 || !w.rows[0].drill || !strings.Contains(w.rows[0].label, "Prontera") {
		t.Fatalf("field groups = %+v", w.rows)
	}
	w.tab = warpraTabDungeon
	w.refreshRows()
	if len(w.rows) != 1 || w.rows[0].drill || !strings.Contains(w.rows[0].label, "Byalan") {
		t.Fatalf("locked dungeon should offer unlock, got %+v", w.rows)
	}
	w.tab = warpraTabSpecial
	w.refreshRows()
	if len(w.rows) != 1 || !strings.Contains(w.rows[0].label, "Turbo Track") {
		t.Fatalf("special = %+v", w.rows)
	}

	var bought network.WarpraLoc
	w.onBuy = func(loc network.WarpraLoc) { bought = loc }
	ctx := Context{Session: &session.Session{Warpra: catalog}}
	w.UnlockOrWarp(ctx, catalog.Locs[2], false)
	if _, warpAfter, ok := w.ConsumePendingBuy(); !ok || warpAfter {
		t.Fatal("locked dungeon did not queue a buy")
	}
	w.Confirm.Confirm(ctx)
	if bought.Map != "iz_dun00" {
		t.Fatalf("buy = %+v", bought)
	}
}

func TestWarpraFloorLabelAndUnlockMessage(t *testing.T) {
	if got := warpraFloorLabel(network.WarpraLoc{Name: "Byalan 1", Map: "iz_dun00", Unlocked: true}, 50_000_000); !strings.Contains(got, "Warp") || !strings.Contains(got, "✓") {
		t.Fatalf("unlocked floor = %q", got)
	}
	if got := warpraFloorLabel(network.WarpraLoc{Name: "Byalan 2", Map: "iz_dun01"}, 50_000_000); !strings.Contains(got, "Unlock") || !strings.Contains(got, "✕") {
		t.Fatalf("locked floor = %q", got)
	}
	title, body := warpraUnlockMessage(network.WarpraLoc{Kind: network.WarpraKindTown, Name: "Geffen", Map: "geffen"}, 50_000_000, 1_000)
	if title != "Unlock Geffen?" || !strings.Contains(body, "50,000,000") || !strings.Contains(body, "1,000") {
		t.Fatalf("unlock message = %q %q", title, body)
	}
}
