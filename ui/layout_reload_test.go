package ui

import (
	"strings"
	"testing"

	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/widget"
	"github.com/kivutar/goro/ui/rotheme"
)

func TestOpenWindowFrameFollowsReloadedTokens(t *testing.T) {
	t.Cleanup(rotheme.ResetLayout)
	frame := Win(Title("Test"), Content(primitives.Box())).(*windowFrame)
	layout, _, err := rotheme.ParseLayout(strings.NewReader("[theme]\nwindow_body = #20304080\nwindow_border = #ff0000\ntitle_text = #00ff00\n[window]\nborder_width = 3\nradius = 0\n"), "test")
	if err != nil {
		t.Fatal(err)
	}
	layout.Apply()
	frame.Layout(widget.NewContext(), geometry.Tight(geometry.Sz(300, 240)))
	style := frame.Style()
	if style.Background != widget.RGBA8(0x20, 0x30, 0x40, 0x80) || style.Border.Color != widget.RGBA8(255, 0, 0, 255) || style.Border.Width != 3 || style.Radius != 0 {
		t.Fatalf("frame kept old styling: %+v", style)
	}
	if frame.title == nil || frame.title.Style().Color != widget.RGBA8(0, 255, 0, 255) {
		t.Fatal("theme title color not applied on reload")
	}
	custom := widget.RGBA8(1, 2, 3, 255)
	frame.title.Color(custom)
	rotheme.ResetLayout()
	frame.Layout(widget.NewContext(), geometry.Tight(geometry.Sz(300, 240)))
	if frame.title.Style().Color != custom {
		t.Fatal("a window's custom title color was lost on reload")
	}
}

func TestShortcutBarSizeFollowsTokens(t *testing.T) {
	t.Cleanup(rotheme.ResetLayout)
	before := shortcutBarWidth()
	layout, _, err := rotheme.ParseLayout(strings.NewReader("[hotbar]\nslot_size = 44\n"), "test")
	if err != nil {
		t.Fatal(err)
	}
	layout.Apply()
	if got := shortcutBarWidth(); got != before+shortcutCols*10 {
		t.Fatalf("hotbar width %d, want %d", got, before+shortcutCols*10)
	}
}
