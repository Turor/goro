package ui

import (
	"image"
	"testing"

	"github.com/kivutar/goro/network"
	"github.com/kivutar/goro/res"
	"github.com/kivutar/goro/session"
)

func testWarpraCatalog() session.WarpraCatalog {
	var catalog session.WarpraCatalog
	catalog.ApplyList(network.WarpraList{
		Zeny: 1_000_000, Price: network.WarpraDefaultPrice, Total: 4,
		Locs: []network.WarpraLoc{
			{Kind: network.WarpraKindTown, ID: 0, Map: "prontera", Name: "Prontera", Unlocked: true},
			{Kind: network.WarpraKindField, ID: 0, Map: "prt_fild08", Name: "Prontera Field 8", GroupName: "Prontera"},
			{Kind: network.WarpraKindDungeon, ID: 0, Group: 1, Map: "iz_dun00", Name: "Byalan 1", GroupName: "Byalan", Unlocked: true},
			{Kind: network.WarpraKindDungeon, ID: 1, Group: 1, Map: "iz_dun01", Name: "Byalan 2", GroupName: "Byalan"},
		},
	})
	return catalog
}

func TestWorldMapExploreMarks(t *testing.T) {
	catalog := testWarpraCatalog()
	entries := []res.WorldMapEntry{
		{MapName: "prontera", Bounds: image.Rect(0, 0, 10, 10)},
		{MapName: "prt_fild08", Bounds: image.Rect(10, 0, 20, 10)},
		{MapName: "izlu2dun", Bounds: image.Rect(20, 0, 30, 10)},
		{MapName: "geffen", Bounds: image.Rect(30, 0, 40, 10)},
	}
	marks := worldMapExploreMarks(entries, catalog)
	if !marks[0].Present || !marks[0].Unlocked {
		t.Fatalf("prontera mark = %+v", marks[0])
	}
	if !marks[1].Present || marks[1].Unlocked {
		t.Fatalf("field mark = %+v", marks[1])
	}
	if !marks[2].Present || !marks[2].Unlocked {
		t.Fatalf("byalan alias should show group unlock: %+v", marks[2])
	}
	if marks[3].Present {
		t.Fatalf("unlisted map grew a mark: %+v", marks[3])
	}
}

func TestResolveWorldMapWarp(t *testing.T) {
	catalog := testWarpraCatalog()
	town := resolveWorldMapWarp(catalog, "prontera", 0, nil)
	if !town.OK || town.OpenFloors || town.Loc.Map != "prontera" {
		t.Fatalf("town = %+v", town)
	}
	locked := resolveWorldMapWarp(catalog, "prt_fild08", 0, nil)
	if !locked.OK || locked.Loc.Unlocked || locked.OpenFloors {
		t.Fatalf("field = %+v", locked)
	}
	dungeon := resolveWorldMapWarp(catalog, "izlu2dun", 4, []string{"izlu2dun", "iz_dun00"})
	if !dungeon.OK || !dungeon.OpenFloors || dungeon.Dungeon.Name != "Byalan" || len(dungeon.Dungeon.Floors) != 2 {
		t.Fatalf("dungeon alias = %+v", dungeon)
	}
	if resolveWorldMapWarp(catalog, "geffen", 0, nil).OK {
		t.Fatal("missing map resolved")
	}
}

func TestFormatWarpraZeny(t *testing.T) {
	if got := formatWarpraZeny(50_000_000); got != "50,000,000" {
		t.Fatalf("got %q", got)
	}
	if got := formatWarpraZeny(0); got != "0" {
		t.Fatalf("zero = %q", got)
	}
}
