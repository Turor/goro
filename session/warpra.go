package session

import (
	"regexp"
	"strconv"

	"github.com/kivutar/goro/network"
)

// WarpraCatalog is the account's Warpra destination list, keyed by map. The
// server splits the catalog across 0x0EFE chunks; ApplyList concatenates them
// until the more flag clears.
type WarpraCatalog struct {
	Zeny     uint32
	Price    uint32
	Locs     []network.WarpraLoc
	assemble []network.WarpraLoc
}

type WarpraDungeon struct {
	Group    uint8
	Name     string
	Floors   []network.WarpraLoc
	Unlocked bool
}

var warpraLevelGateRe = regexp.MustCompile(`\(Lv (\d+)\+\)`)

func (c *WarpraCatalog) ApplyList(list network.WarpraList) bool {
	if c == nil {
		return false
	}
	c.Zeny = list.Zeny
	if list.Price != 0 {
		c.Price = list.Price
	} else if c.Price == 0 {
		c.Price = network.WarpraDefaultPrice
	}
	incoming := append([]network.WarpraLoc(nil), list.Locs...)
	if list.More() {
		c.assemble = append(c.assemble, incoming...)
		return false
	}
	if len(c.assemble) > 0 {
		c.Locs = append(c.assemble, incoming...)
		c.assemble = nil
	} else {
		c.Locs = incoming
	}
	return true
}

func (c *WarpraCatalog) Ready() bool {
	return c != nil && len(c.Locs) > 0 && len(c.assemble) == 0
}

func (c *WarpraCatalog) UnlockPrice() uint32 {
	if c == nil || c.Price == 0 {
		return network.WarpraDefaultPrice
	}
	return c.Price
}

func (c *WarpraCatalog) FindByMap(mapID string) (network.WarpraLoc, bool) {
	if c == nil {
		return network.WarpraLoc{}, false
	}
	id := network.NormalizeWarpraMapID(mapID)
	if id == "" {
		return network.WarpraLoc{}, false
	}
	for _, loc := range c.Locs {
		if loc.Map == id {
			return loc, true
		}
	}
	return network.WarpraLoc{}, false
}

func (c *WarpraCatalog) Find(kind uint8, id uint16) (network.WarpraLoc, bool) {
	if c == nil {
		return network.WarpraLoc{}, false
	}
	for _, loc := range c.Locs {
		if loc.Kind == kind && loc.ID == id {
			return loc, true
		}
	}
	return network.WarpraLoc{}, false
}

func (c *WarpraCatalog) MapIndex() map[string]network.WarpraLoc {
	out := make(map[string]network.WarpraLoc)
	if c == nil {
		return out
	}
	for _, loc := range c.Locs {
		if loc.Map != "" {
			out[loc.Map] = loc
		}
	}
	return out
}

func (c *WarpraCatalog) OfKind(kind uint8) []network.WarpraLoc {
	if c == nil {
		return nil
	}
	var out []network.WarpraLoc
	for _, loc := range c.Locs {
		if loc.Kind == kind {
			out = append(out, loc)
		}
	}
	return out
}

func (c *WarpraCatalog) DungeonFloorsByGroup(group uint8) []network.WarpraLoc {
	if c == nil {
		return nil
	}
	var out []network.WarpraLoc
	for _, loc := range c.Locs {
		if loc.Kind == network.WarpraKindDungeon && loc.Group == group {
			out = append(out, loc)
		}
	}
	return out
}

func (c *WarpraCatalog) DungeonFloors(mapID string) (WarpraDungeon, bool) {
	loc, ok := c.FindByMap(mapID)
	if !ok || loc.Kind != network.WarpraKindDungeon {
		return WarpraDungeon{}, false
	}
	floors := c.DungeonFloorsByGroup(loc.Group)
	name := loc.GroupName
	if name == "" {
		name = loc.Name
	}
	if name == "" {
		name = "Dungeon"
	}
	unlocked := false
	for _, floor := range floors {
		if floor.Unlocked {
			unlocked = true
			break
		}
	}
	return WarpraDungeon{Group: loc.Group, Name: name, Floors: floors, Unlocked: unlocked}, true
}

func (c *WarpraCatalog) Groups(kind uint8) []WarpraGroup {
	if c == nil {
		return nil
	}
	index := map[uint8]int{}
	var groups []WarpraGroup
	for _, loc := range c.Locs {
		if loc.Kind != kind {
			continue
		}
		i, ok := index[loc.Group]
		if !ok {
			index[loc.Group] = len(groups)
			name := loc.GroupName
			if name == "" {
				name = loc.Name
			}
			groups = append(groups, WarpraGroup{ID: loc.Group, Name: name})
			i = len(groups) - 1
		}
		groups[i].Locs = append(groups[i].Locs, loc)
		if loc.Unlocked {
			groups[i].Unlocked++
		}
	}
	return groups
}

type WarpraGroup struct {
	ID       uint8
	Name     string
	Locs     []network.WarpraLoc
	Unlocked int
}

func (g WarpraGroup) Parent() network.WarpraLoc {
	if len(g.Locs) == 0 {
		return network.WarpraLoc{}
	}
	parent := g.Locs[0]
	parent.Name = g.Name
	parent.Unlocked = g.Unlocked > 0
	return parent
}

// WarpraLevelGate reads the "(Lv N+)" suffix used for iRO-parity destinations.
// Those entries are free at that base level and cannot be bought.
func WarpraLevelGate(loc network.WarpraLoc) int {
	m := warpraLevelGateRe.FindStringSubmatch(loc.Name)
	if len(m) != 2 {
		return 0
	}
	n, err := strconv.Atoi(m[1])
	if err != nil {
		return 0
	}
	return n
}

func WarpraLevelGateMessage(loc network.WarpraLoc) string {
	lv := WarpraLevelGate(loc)
	if lv == 0 {
		return ""
	}
	name := warpraLevelGateRe.ReplaceAllString(loc.Name, "")
	name = trimWarpraSpaces(name)
	if name == "" {
		name = loc.Map
	}
	return name + " requires base level " + strconv.Itoa(lv) + "."
}

func trimWarpraSpaces(s string) string {
	start, end := 0, len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t') {
		end--
	}
	return s[start:end]
}
