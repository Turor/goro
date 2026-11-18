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

func TestWarpraCatalogSecondRequestDropsStalePartial(t *testing.T) {
	var catalog WarpraCatalog
	if catalog.ApplyList(network.WarpraList{
		Zeny: 10, Price: network.WarpraDefaultPrice, Flags: network.WarpraListFlagMore, Total: 2,
		Locs: []network.WarpraLoc{{Kind: network.WarpraKindTown, ID: 0, Map: "prontera", Name: "Prontera"}},
	}) {
		t.Fatal("first MORE chunk completed the catalog")
	}
	catalog.BeginList()
	if len(catalog.assemble) != 0 {
		t.Fatal("second request kept the in-flight partial")
	}
	// Same declared total as the abandoned chain. Concatenating would still
	// satisfy Total, which is how the old append path hid the overlap.
	if catalog.ApplyList(network.WarpraList{
		Zeny: 40, Price: network.WarpraDefaultPrice, Flags: network.WarpraListFlagMore, Total: 2,
		Locs: []network.WarpraLoc{{Kind: network.WarpraKindTown, ID: 1, Map: "geffen", Name: "Geffen"}},
	}) {
		t.Fatal("new MORE chunk completed the catalog")
	}
	if !catalog.ApplyList(network.WarpraList{
		Zeny: 40, Price: network.WarpraDefaultPrice, Total: 2,
		Locs: []network.WarpraLoc{{Kind: network.WarpraKindField, ID: 2, Map: "gef_fild00", Name: "Geffen Field"}},
	}) {
		t.Fatal("new final chunk did not complete")
	}
	if len(catalog.Locs) != 2 || catalog.Locs[0].Map != "geffen" || catalog.Locs[1].Map != "gef_fild00" {
		t.Fatalf("overlapping lists were concatenated: %+v", catalog.Locs)
	}

	// A finished chunk that arrives while a partial is buffered must replace
	// that partial, even when the caller did not retire the chain.
	var stale WarpraCatalog
	stale.ApplyList(network.WarpraList{
		Flags: network.WarpraListFlagMore, Total: 3,
		Locs: []network.WarpraLoc{{Map: "morocc", Name: "Morocc"}},
	})
	if !stale.ApplyList(network.WarpraList{
		Total: 1,
		Locs:  []network.WarpraLoc{{Map: "payon", Name: "Payon"}},
	}) {
		t.Fatal("replacement list did not complete")
	}
	if len(stale.Locs) != 1 || stale.Locs[0].Map != "payon" || len(stale.assemble) != 0 {
		t.Fatalf("stale partial was prepended: locs=%+v assemble=%+v", stale.Locs, stale.assemble)
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
