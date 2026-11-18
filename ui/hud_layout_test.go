package ui

import (
	"testing"

	"github.com/gogpu/ui/primitives"
	"github.com/kivutar/goro/client"
	"github.com/kivutar/goro/input"
	"github.com/kivutar/goro/session"
	"github.com/kivutar/goro/ui/rotheme"
	worldstate "github.com/kivutar/goro/world"
)

func TestHUDWindowsFollowLayoutTokens(t *testing.T) {
	t.Cleanup(rotheme.ResetLayout)
	manager := NewManager()
	world := worldstate.New()
	world.MapName = "prontera"
	ctx := client.Context{
		Input:     input.NewState(),
		UIManager: manager,
		World:     world,
		Session:   session.New(),
		ScreenW:   1280,
		ScreenH:   720,
	}

	console := &ChatConsole{}
	console.UpdatePresentation(ctx)
	menu := &BasicMenu{}
	menu.Update(ctx, BasicMenuCallbacks{})
	minimap := &Minimap{}
	if minimap.Update(ctx) {
		t.Fatal("opening the minimap consumed input")
	}
	if !console.window.IsOpen() || !menu.IsOpen() || !minimap.window.IsOpen() {
		t.Fatal("hud windows did not open")
	}

	consoleW, consoleH := console.window.width, console.window.height
	menuH := menu.height
	mapW := minimap.window.width
	consoleBg := consoleBackground
	playerColor := minimapPlayerColor
	box, ok := console.window.content.(*primitives.BoxWidget)
	if !ok || box.Style().Background != consoleBg {
		t.Fatalf("console background = %#v", console.window.content)
	}

	if err := rotheme.SetToken("console", "width", "600"); err != nil {
		t.Fatal(err)
	}
	if err := rotheme.SetToken("console", "background", "#ff00aa"); err != nil {
		t.Fatal(err)
	}
	if err := rotheme.SetToken("basic_menu", "button_height", "40"); err != nil {
		t.Fatal(err)
	}
	if err := rotheme.SetToken("minimap", "width", "240"); err != nil {
		t.Fatal(err)
	}
	if err := rotheme.SetToken("minimap", "player_color", "#00ff00"); err != nil {
		t.Fatal(err)
	}
	if n := manager.RebuildStaleWindows(); n != 3 {
		t.Fatalf("rebuilt %d hud windows, want 3", n)
	}

	box, ok = console.window.content.(*primitives.BoxWidget)
	if !ok || console.window.width != 600 || console.window.width == consoleW || box.Style().Background == consoleBg || box.Style().Background != consoleBackground {
		t.Fatalf("console size/color width=%d background=%v", console.window.width, box.Style().Background)
	}
	if _, want := basicMenuSize(false); menu.height == menuH || menu.height != want {
		t.Fatalf("basic menu height = %d, was %d, want %d", menu.height, menuH, want)
	}
	frame, ok := minimap.window.content.(*windowFrame)
	if !ok || minimap.window.width != 240 || minimap.window.width == mapW || frame.Style().ExplicitWidth != float32(minimapWidth) || minimapPlayerColor == playerColor {
		t.Fatalf("minimap width=%d frame=%v color=%v", minimap.window.width, minimap.window.content, minimapPlayerColor)
	}

	rotheme.ResetLayout()
	if n := manager.RebuildStaleWindows(); n != 3 {
		t.Fatalf("restored %d hud windows, want 3", n)
	}
	box, ok = console.window.content.(*primitives.BoxWidget)
	if !ok || console.window.width != consoleW || console.window.height != consoleH || box.Style().Background != consoleBg {
		t.Fatalf("console did not restore size=%dx%d background=%v", console.window.width, console.window.height, box.Style().Background)
	}
	if _, want := basicMenuSize(false); menu.height != want || menu.height != menuH {
		t.Fatalf("basic menu height = %d, want %d", menu.height, menuH)
	}
	frame, ok = minimap.window.content.(*windowFrame)
	if !ok || minimap.window.width != mapW || frame.Style().ExplicitWidth != float32(mapW) || minimapPlayerColor != playerColor {
		t.Fatalf("minimap did not restore width=%d color=%v", minimap.window.width, minimapPlayerColor)
	}
}
