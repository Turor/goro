package session

import (
	"testing"

	"github.com/kivutar/goro/network"
)

func TestWarpraCatalogAssemblesChunksAndIndexesMaps(t *testing.T) {
	var catalog WarpraCatalog
	if catalog.Ready() {
		t.Fatal("empty catalog was ready")
	}
	if catalog.ApplyList(network.WarpraList{
		Zeny: 100, Price: network.WarpraDefaultPrice, Flags: network.WarpraListFlagMore, Total: 3,
		Locs: []network.WarpraLoc{{Kind: network.WarpraKindTown, ID: 0, Map: "prontera", Name: "Prontera", Unlocked: true}},
	}) {
		t.Fatal("more-flag chunk completed the catalog")
	}
	if catalog.Ready() || len(catalog.Locs) != 0 {
		t.Fatal("partial catalog leaked into Locs")
	}
	if catalog.ApplyList(network.WarpraList{
		Zeny: 100, Total: 3,
		Locs: []network.WarpraLoc{
			{Kind: network.WarpraKindDungeon, ID: 0, Group: 1, Map: "iz_dun00", Name: "Byalan 1", GroupName: "Byalan", Unlocked: true},
			{Kind: network.WarpraKindDungeon, ID: 1, Group: 1, Map: "iz_dun01", Name: "Byalan 2", GroupName: "Byalan"},
		},
	}) != true {
		t.Fatal("final chunk did not complete")
	}
	if !catalog.Ready() || len(catalog.Locs) != 3 || catalog.Zeny != 100 {
		t.Fatalf("catalog = %+v", catalog)
	}
	loc, ok := catalog.FindByMap("IZ_DUN00.rsw")
	if !ok || loc.Name != "Byalan 1" {
		t.Fatalf("FindByMap = %+v ok=%t", loc, ok)
	}
	dungeon, ok := catalog.DungeonFloors("iz_dun01")
	if !ok || dungeon.Name != "Byalan" || len(dungeon.Floors) != 2 || !dungeon.Unlocked {
		t.Fatalf("dungeon = %+v ok=%t", dungeon, ok)
	}
	if _, ok := catalog.FindByMap("geffen"); ok {
		t.Fatal("missing map was found")
	}
}

func TestWarpraLevelGate(t *testing.T) {
	loc := network.WarpraLoc{Name: "Jormungandr Dungeon (Lv 180+)", Map: "jor_back1"}
	if WarpraLevelGate(loc) != 180 {
		t.Fatalf("level = %d", WarpraLevelGate(loc))
	}
	if msg := WarpraLevelGateMessage(loc); msg != "Jormungandr Dungeon requires base level 180." {
		t.Fatalf("message = %q", msg)
	}
	if WarpraLevelGate(network.WarpraLoc{Name: "Prontera"}) != 0 {
		t.Fatal("plain name was treated as level-gated")
	}
}

func TestSelectCharacterClearsWarpraCatalog(t *testing.T) {
	s := New()
	s.Warpra.Locs = []network.WarpraLoc{{Map: "prontera"}}
	s.SelectCharacter(Character{ID: 1, Name: "New"})
	if len(s.Warpra.Locs) != 0 || s.Warpra.Zeny != 0 {
		t.Fatalf("warpra leaked across character selection: %+v", s.Warpra)
	}
}
