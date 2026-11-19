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
	window.resizeTo(ctx, 340, 230)
	window.SetSize(100, 80)
	if window.width != 340 || window.height != 230 {
		t.Fatalf("SetSize clobbered size to %dx%d", window.width, window.height)
	}
	window.SetContent(Win(Title("Box"), Size(100, 80), Content(primitives.Box())))
	frame, ok := window.content.(*windowFrame)
	if !ok {
		t.Fatal("content is not a window frame")
	}
	frame.Layout(widget.NewContext(), geometry.Loose(geometry.Sz(800, 600)))
	bounds := frame.Bounds()
	if bounds.Width() != 340 || bounds.Height() != 230 {
		t.Fatalf("frame = %.0fx%.0f, want 340x230", bounds.Width(), bounds.Height())
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
