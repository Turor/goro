package ui

import (
	"strconv"

	"github.com/kivutar/goro/network"
	"github.com/kivutar/goro/res"
	"github.com/kivutar/goro/session"
)

// World-map section id → Warpra catalog map. Dungeon tiles whose WM id
// differs from the catalog still open the floor picker.
var worldMapDungeonAlias = map[string]string{
	"izlu2dun":   "iz_dun00",
	"mosk_dun01": "mosk_dun01",
	"mosk_dun02": "mosk_dun01",
	"mosk_dun03": "mosk_dun01",
	"odin_tem01": "odin_tem01",
	"odin_tem02": "odin_tem01",
	"odin_tem03": "odin_tem01",
	"tur_dun01":  "tur_dun01",
	"gon_dun01":  "gon_dun01",
	"lou_dun01":  "lou_dun01",
	"dew_dun01":  "dew_dun01",
}

type worldMapExploreMark struct {
	Present  bool
	Unlocked bool
}

type worldMapWarpAction struct {
	Loc        network.WarpraLoc
	Dungeon    session.WarpraDungeon
	OpenFloors bool
	OK         bool
}

func warpraLocForWorldMap(catalog session.WarpraCatalog, mapID string) (network.WarpraLoc, string, bool) {
	id := network.NormalizeWarpraMapID(mapID)
	alias := worldMapDungeonAlias[id]
	lookup := id
	if alias != "" {
		lookup = alias
	}
	if loc, ok := catalog.FindByMap(lookup); ok {
		return loc, alias, true
	}
	if alias != "" {
		if loc, ok := catalog.FindByMap(id); ok {
			return loc, alias, true
		}
	}
	return network.WarpraLoc{}, alias, false
}

func worldMapExploreMarks(entries []res.WorldMapEntry, catalog session.WarpraCatalog) []worldMapExploreMark {
	return appendWorldMapExploreMarks(nil, entries, catalog)
}

func appendWorldMapExploreMarks(dst []worldMapExploreMark, entries []res.WorldMapEntry, catalog session.WarpraCatalog) []worldMapExploreMark {
	dst = dst[:0]
	if !catalog.Ready() {
		return dst
	}
	for _, entry := range entries {
		loc, _, ok := warpraLocForWorldMap(catalog, entry.MapName)
		if !ok {
			dst = append(dst, worldMapExploreMark{})
			continue
		}
		unlocked := loc.Unlocked
		if loc.Kind == network.WarpraKindDungeon {
			floors := catalog.DungeonFloorsByGroup(loc.Group)
			unlocked = false
			for _, floor := range floors {
				if floor.Unlocked {
					unlocked = true
					break
				}
			}
		}
		dst = append(dst, worldMapExploreMark{Present: true, Unlocked: unlocked})
	}
	return dst
}

func resolveWorldMapWarp(catalog session.WarpraCatalog, mapID string, region int, regionMaps []string) worldMapWarpAction {
	id := network.NormalizeWarpraMapID(mapID)
	loc, alias, ok := warpraLocForWorldMap(catalog, id)
	if !ok || loc.Kind != network.WarpraKindDungeon {
		if group := dungeonLocFromRegion(catalog, regionMaps); group.Map != "" {
			loc = group
			ok = true
			if alias == "" {
				alias = worldMapDungeonAlias[network.NormalizeWarpraMapID(group.Map)]
			}
		}
	}
	if !ok {
		return worldMapWarpAction{}
	}
	if loc.Kind == network.WarpraKindDungeon || alias != "" {
		floorID := loc.Map
		if loc.Kind != network.WarpraKindDungeon && alias != "" {
			floorID = alias
		}
		info, found := catalog.DungeonFloors(floorID)
		if !found {
			info, found = catalog.DungeonFloors(loc.Map)
		}
		if !found || len(info.Floors) == 0 {
			return worldMapWarpAction{}
		}
		return worldMapWarpAction{Loc: loc, Dungeon: info, OpenFloors: true, OK: true}
	}
	return worldMapWarpAction{Loc: loc, OK: true}
}

func dungeonLocFromRegion(catalog session.WarpraCatalog, regionMaps []string) network.WarpraLoc {
	for _, name := range regionMaps {
		loc, _, ok := warpraLocForWorldMap(catalog, name)
		if ok && loc.Kind == network.WarpraKindDungeon {
			return loc
		}
	}
	return network.WarpraLoc{}
}

func regionMapNames(entries []res.WorldMapEntry, region int) []string {
	var names []string
	for _, entry := range entries {
		if entry.Region == region {
			names = append(names, entry.MapName)
		}
	}
	return names
}

func formatWarpraZeny(n uint32) string {
	s := strconv.FormatUint(uint64(n), 10)
	if len(s) <= 3 {
		return s
	}
	buf := make([]byte, 0, len(s)+len(s)/3)
	lead := len(s) % 3
	if lead == 0 {
		lead = 3
	}
	buf = append(buf, s[:lead]...)
	for i := lead; i < len(s); i += 3 {
		buf = append(buf, ',')
		buf = append(buf, s[i:i+3]...)
	}
	return string(buf)
}

func warpraUnlockMessage(loc network.WarpraLoc, price, zeny uint32) (title, body string) {
	label := loc.Name
	if loc.Kind == network.WarpraKindDungeon {
		if loc.GroupName != "" {
			label = loc.GroupName
		}
		title = "Unlock " + label + "?"
		body = label + " is locked. Spend " + formatWarpraZeny(price) + " zeny to unlock every floor and teleport? (You have " + formatWarpraZeny(zeny) + ".)"
		return title, body
	}
	if label == "" {
		label = loc.Map
	}
	title = "Unlock " + label + "?"
	body = label + " is locked. Spend " + formatWarpraZeny(price) + " zeny to unlock and teleport there? (You have " + formatWarpraZeny(zeny) + ".)"
	return title, body
}
