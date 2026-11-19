package ui

import (
	"testing"

	"github.com/gogpu/ui/widget"
	"github.com/kivutar/goro/input"
	"github.com/kivutar/goro/network"
	"github.com/kivutar/goro/session"
	"github.com/kivutar/goro/ui/rotheme"
)

func TestCharacterGroupLayoutReload(t *testing.T) {
	t.Cleanup(rotheme.ResetLayout)
	manager := NewManager()
	ctx := Context{
		Session:   session.New(),
		Input:     input.NewState(),
		ScreenW:   1280,
		ScreenH:   720,
		UIManager: manager,
	}

	character := &CharacterWindow{}
	character.Update(ctx)
	equipment := &EquipmentWindow{}
	equipment.Toggle(ctx)
	stats := &StatsWindow{}
	stats.OpenWindow(ctx)
	skill := &SkillWindow{}
	skill.OpenWindow(ctx)
	inventory := &InventoryBagWindow{}
	inventory.Toggle(ctx)
	itemInfo := &ItemInfoWindow{}
	itemInfo.openItem(ctx, session.InventoryItem{ItemID: 501, Identified: true}, 200, 200)
	viewed := &ViewEquipmentWindow{}
	viewed.Open(ctx, network.ViewedEquipment{Name: "Seen"}, nil)

	defCharW := character.width
	defEquipW := equipment.width
	defStatsH := stats.height
	defSkillH := skill.height
	defInvH := inventory.height
	defInfoW := itemInfo.width
	defViewW := viewed.width
	bar := findCharacterBar(character.content)
	if bar == nil {
		t.Fatal("character window has no exp bar")
	}
	defFill := bar.fill

	set := func(section, key, value string) {
		t.Helper()
		if err := rotheme.SetToken(section, key, value); err != nil {
			t.Fatal(err)
		}
	}
	set("character", "width", "480")
	set("character", "exp_color", "#ff0000")
	set("equipment", "width", "420")
	set("status", "row_height", "24")
	set("skill", "height", "460")
	set("inventory", "cell_size", "40")
	set("item_info", "width", "420")

	if n := manager.RebuildStaleWindows(); n != 7 {
		t.Fatalf("rebuilt %d windows, want 7", n)
	}
	if character.width != 480 {
		t.Fatalf("character width = %d, want 480", character.width)
	}
	bar = findCharacterBar(character.content)
	if bar == nil || bar.fill == defFill || bar.fill.R < 0.99 {
		t.Fatalf("exp bar fill = %+v, want the reloaded red", bar)
	}
	if equipment.width != gearWindowSizeWidth() || viewed.width != 420 {
		t.Fatalf("equipment width = %d, viewed = %d, want gear %d and paperdoll 420", equipment.width, viewed.width, gearWindowSizeWidth())
	}
	if stats.height != defStatsH+6*6 {
		t.Fatalf("status height = %d, want %d", stats.height, defStatsH+6*6)
	}
	if skill.height != 460 {
		t.Fatalf("skill height = %d, want 460", skill.height)
	}
	if inventory.height != defInvH+5*8 {
		t.Fatalf("inventory height = %d, want %d", inventory.height, defInvH+5*8)
	}
	if itemInfo.width != 420 {
		t.Fatalf("item info width = %d, want 420", itemInfo.width)
	}

	rotheme.ResetLayout()
	if n := manager.RebuildStaleWindows(); n != 7 {
		t.Fatalf("restored %d windows, want 7", n)
	}
	if character.width != defCharW || equipment.width != defEquipW || stats.height != defStatsH || skill.height != defSkillH || inventory.height != defInvH || itemInfo.width != defInfoW || viewed.width != defViewW {
		t.Fatalf("defaults not restored: character %d equipment %d status %d skill %d inventory %d item info %d viewed %d",
			character.width, equipment.width, stats.height, skill.height, inventory.height, itemInfo.width, viewed.width)
	}
	bar = findCharacterBar(character.content)
	if bar == nil || bar.fill != defFill {
		t.Fatalf("exp bar fill = %+v, want the default %+v", barFill(bar), defFill)
	}
}

func findCharacterBar(root widget.Widget) *characterBarWidget {
	if root == nil {
		return nil
	}
	if bar, ok := root.(*characterBarWidget); ok {
		return bar
	}
	for _, child := range root.Children() {
		if bar := findCharacterBar(child); bar != nil {
			return bar
		}
	}
	return nil
}

func barFill(bar *characterBarWidget) widget.Color {
	if bar == nil {
		return widget.Color{}
	}
	return bar.fill
}
