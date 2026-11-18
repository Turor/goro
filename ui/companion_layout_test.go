package ui

import (
	"image/color"
	"testing"
	"time"

	"github.com/gogpu/ui/widget"
	"github.com/kivutar/goro/input"
	"github.com/kivutar/goro/network"
	"github.com/kivutar/goro/session"
	"github.com/kivutar/goro/ui/rotheme"
)

func TestCompanionWindowsFollowLayoutReload(t *testing.T) {
	t.Cleanup(rotheme.ResetLayout)
	manager := NewManager()
	ctx := Context{
		Input: input.NewState(), UIManager: manager, ScreenW: 1280, ScreenH: 800,
		Session: session.New(),
	}
	ctx.Session.Homunculus = session.Companion{Active: true, Name: "Lif", Flags: 5}
	ctx.Session.Instance = session.Instance{Name: "Endless Tower"}
	ctx.Session.Quests.Set(session.Quest{ID: 1, Active: true})

	monster := &MonsterInfoWindow{}
	monster.OpenInfo(ctx, MonsterInfoView{Name: "Poring"})
	quest := &QuestWindow{}
	quest.OpenWindow(ctx)
	quest.openDetail(ctx, 1)
	hom := &HomunculusInfoWindow{}
	hom.OpenInfo(ctx, ctx.Session.Homunculus)
	merc := &MercenaryInfoWindow{}
	merc.OpenInfo(ctx, session.Companion{Name: "Merc", ExpireTick: uint32(time.Now().Unix() + 1200)})
	skill := &HomunculusSkillWindow{}
	skill.Open(ctx, nil)
	pet := &PetInfoWindow{}
	pet.OpenInfo(ctx, network.PetProperty{Name: "Poring", Modified: true})
	settings := &SettingsWindow{}
	settings.OpenWindow(ctx)
	instance := &InstanceWindow{}
	instance.Toggle(ctx)

	type opened struct {
		name       string
		win        *Window
		x, y, w, h int
		content    widget.Widget
	}
	windows := []opened{
		{name: "monster", win: &monster.Window, x: monster.x, y: monster.y, w: monster.width, h: monster.height, content: monster.content},
		{name: "quest", win: &quest.Window, x: quest.x, y: quest.y, w: quest.width, h: quest.height, content: quest.content},
		{name: "quest detail", win: &quest.detail, x: quest.detail.x, y: quest.detail.y, w: quest.detail.width, h: quest.detail.height, content: quest.detail.content},
		{name: "homunculus", win: &hom.Window, x: hom.x, y: hom.y, w: hom.width, h: hom.height, content: hom.content},
		{name: "mercenary", win: &merc.Window, x: merc.x, y: merc.y, w: merc.width, h: merc.height, content: merc.content},
		{name: "skill", win: &skill.Window, x: skill.x, y: skill.y, w: skill.width, h: skill.height, content: skill.content},
		{name: "pet", win: &pet.Window, x: pet.x, y: pet.y, w: pet.width, h: pet.height, content: pet.content},
		{name: "settings", win: &settings.Window, x: settings.x, y: settings.y, w: settings.width, h: settings.height, content: settings.content},
		{name: "instance", win: &instance.Window, x: instance.x, y: instance.y, w: instance.width, h: instance.height, content: instance.content},
	}
	for _, item := range windows {
		if !item.win.IsOpen() || item.w == 0 || item.content == nil {
			t.Fatalf("%s did not open (open=%v width=%d)", item.name, item.win.IsOpen(), item.w)
		}
	}

	set := func(section, key, value string) {
		t.Helper()
		if err := rotheme.SetToken(section, key, value); err != nil {
			t.Fatal(err)
		}
	}
	set("monster_info", "width", "520")
	set("monster_info", "preview_width", "140")
	set("quest", "width", "420")
	set("quest", "detail_width", "440")
	set("homunculus_info", "width", "360")
	set("homunculus_info", "exp_color", "#ff00aa")
	set("homunculus_info", "bar_background", "#101010")
	set("mercenary_info", "height", "320")
	set("mercenary_info", "time_color", "#123456")
	set("homunculus_skill", "width", "480")
	set("homunculus_skill", "icon_column_width", "48")
	set("pet_info", "width", "360")
	set("settings", "width", "420")
	set("instance", "width", "400")

	if n := manager.RebuildStaleWindows(); n != len(windows) {
		t.Fatalf("rebuilt %d windows, want %d", n, len(windows))
	}
	if monster.width != 520 || quest.width != 420 || quest.detail.width != 440 {
		t.Fatalf("widths monster=%d quest=%d detail=%d", monster.width, quest.width, quest.detail.width)
	}
	// Mercenary width follows the homunculus window width, as it did before tokens.
	if hom.width != 360 || merc.width != 360 || merc.height != 320 {
		t.Fatalf("companion sizes hom=%d merc=%dx%d", hom.width, merc.width, merc.height)
	}
	if skill.width != 480 || pet.width != 360 || settings.width != 420 || instance.width != 400 {
		t.Fatalf("widths skill=%d pet=%d settings=%d instance=%d", skill.width, pet.width, settings.width, instance.width)
	}
	for _, item := range windows {
		if item.win.content == item.content {
			t.Fatalf("%s content was not rebuilt", item.name)
		}
		if item.win.x != item.x || item.win.y != item.y {
			t.Fatalf("%s moved from %d,%d to %d,%d", item.name, item.x, item.y, item.win.x, item.win.y)
		}
	}
	if !widgetHasBarFill(hom.content, Color(color.RGBA{R: 0xff, G: 0x00, B: 0xaa, A: 0xff})) {
		t.Fatal("homunculus exp bar did not take the layout color")
	}
	if !widgetHasBarFill(merc.content, Color(color.RGBA{R: 0x12, G: 0x34, B: 0x56, A: 0xff})) {
		t.Fatal("mercenary time bar did not take the layout color")
	}
	if previewW := staticImageWidth(monster.content); previewW != 140 {
		t.Fatalf("monster preview width = %d, want 140", previewW)
	}
	if iconW := companionSkillTableColumns()[0].Width; iconW != 48 {
		t.Fatalf("skill icon column = %v, want 48", iconW)
	}

	rotheme.ResetLayout()
	if n := manager.RebuildStaleWindows(); n != len(windows) {
		t.Fatalf("restore rebuilt %d windows, want %d", n, len(windows))
	}
	if monster.width != 400 || quest.width != 350 || quest.detail.width != 360 || hom.width != 300 || merc.width != 300 || merc.height != 272 || skill.width != 360 || pet.width != 292 || settings.width != 300 || instance.width != 320 {
		t.Fatalf("restored monster=%d quest=%d detail=%d hom=%d merc=%dx%d skill=%d pet=%d settings=%d instance=%d",
			monster.width, quest.width, quest.detail.width, hom.width, merc.width, merc.height, skill.width, pet.width, settings.width, instance.width)
	}
	if !widgetHasBarFill(hom.content, Color(homunculusInfoExpColor)) {
		t.Fatal("homunculus exp bar did not return to the default color")
	}
	if previewW := staticImageWidth(monster.content); previewW != monsterInfoPreviewW {
		t.Fatalf("restored preview width = %d, want %d", previewW, monsterInfoPreviewW)
	}
	for _, item := range windows {
		if item.win.x != item.x || item.win.y != item.y {
			t.Fatalf("%s moved on restore from %d,%d to %d,%d", item.name, item.x, item.y, item.win.x, item.win.y)
		}
	}
}

func widgetHasBarFill(root widget.Widget, fill widget.Color) bool {
	found := false
	walkWidgets(root, func(w widget.Widget) {
		if bar, ok := w.(*homunculusBarWidget); ok && bar.fill == fill {
			found = true
		}
	})
	return found
}

func staticImageWidth(root widget.Widget) int {
	width := 0
	walkWidgets(root, func(w widget.Widget) {
		if img, ok := w.(*staticImageWidget); ok {
			width = img.width
		}
	})
	return width
}

func walkWidgets(root widget.Widget, visit func(widget.Widget)) {
	if root == nil {
		return
	}
	visit(root)
	parent, ok := root.(interface{ Children() []widget.Widget })
	if !ok {
		return
	}
	for _, child := range parent.Children() {
		walkWidgets(child, visit)
	}
}
