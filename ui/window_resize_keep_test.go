package ui

import (
	"testing"

	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/widget"
	"github.com/kivutar/goro/input"
)

func TestSetSizeKeepsRememberedResize(t *testing.T) {
	ctx := Context{Input: input.NewState(), UIManager: NewManager(), ScreenW: 1280, ScreenH: 720}
	window := NewWindow(300, 200)
	window.Open(ctx, Win(Title("Box"), Size(300, 200), Content(primitives.Box())))
	window.Publish(ctx)
	window.resizeTo(ctx, 450, 300)
	window.SetSize(100, 80)
	if window.width != 450 || window.height != 300 {
		t.Fatalf("SetSize clobbered size to %dx%d", window.width, window.height)
	}
	window.SetContent(Win(Title("Box"), Size(100, 80), Content(primitives.Box())))
	frame, ok := window.content.(*windowFrame)
	if !ok {
		t.Fatal("content is not a window frame")
	}
	frame.Layout(widget.NewContext(), geometry.Loose(geometry.Sz(800, 600)))
	bounds := frame.Bounds()
	if bounds.Width() != 450 || bounds.Height() != 300 {
		t.Fatalf("frame = %.0fx%.0f, want 450x300", bounds.Width(), bounds.Height())
	}
}

func TestGripPressWithoutMovementDoesNotSaveSize(t *testing.T) {
	ctx := Context{Input: input.NewState(), UIManager: NewManager(), ScreenW: 1280, ScreenH: 720}
	window := NewWindow(300, 200)
	window.userMoved = true
	window.x, window.y = 40, 50
	window.Open(ctx, Win(Title("Box"), Size(300, 200), Content(primitives.Box())))
	pressX, pressY := window.x+window.width-2, window.y+window.height-2
	ctx.Input.SetMousePosition(pressX, pressY)
	ctx.Input.SetMouseButton(input.MouseButtonLeft, true)
	window.Update(ctx)
	ctx.Input.EndFrame()
	window.Update(ctx)
	ctx.Input.EndFrame()
	ctx.Input.SetMouseButton(input.MouseButtonLeft, false)
	window.Update(ctx)
	saved := window.SavedPosition()
	if saved.Width != 0 || saved.Height != 0 {
		t.Fatalf("unchanged grip press saved size %+v", saved)
	}
}

func TestUserSizeDoesNotLookLikeALayoutMismatch(t *testing.T) {
	ctx := Context{Input: input.NewState(), UIManager: NewManager(), ScreenW: 1280, ScreenH: 720}
	window := NewWindow(300, 200)
	window.Open(ctx, Win(Title("Box"), Size(300, 200), Content(primitives.Box())))
	window.resizeTo(ctx, 450, 300)
	if !window.holdsUserSize() {
		t.Fatal("resized window does not hold its size")
	}
	window.SetSize(300, 200)
	if window.width != 450 || window.height != 300 {
		t.Fatalf("layout SetSize changed user size to %dx%d", window.width, window.height)
	}
}

func TestSetContentDuringResizeInstallsTree(t *testing.T) {
	ctx := Context{Input: input.NewState(), UIManager: NewManager(), ScreenW: 1280, ScreenH: 720}
	window := NewWindow(300, 200)
	window.Open(ctx, Win(Title("Box"), Size(300, 200), Content(primitives.Box())))
	window.Publish(ctx)
	pressX, pressY := window.x+window.width-2, window.y+window.height-2
	ctx.Input.SetMousePosition(pressX, pressY)
	ctx.Input.SetMouseButton(input.MouseButtonLeft, true)
	window.Update(ctx)
	if !window.Resizing() {
		t.Fatal("grip press did not start a resize")
	}
	next := Win(Title("Updated"), Size(300, 200), Content(primitives.Box()))
	window.SetContent(next)
	if window.content != next {
		t.Fatal("SetContent did not install the new tree")
	}
	if window.width != 300 || window.height != 200 {
		t.Fatalf("resize start changed size to %dx%d", window.width, window.height)
	}
}
