package ui

import (
	"testing"

	"github.com/gogpu/ui/primitives"
	"github.com/kivutar/goro/input"
)

func TestWindowResizePersistsSize(t *testing.T) {
	manager := NewManager()
	ctx := Context{Input: input.NewState(), UIManager: manager, ScreenW: 1280, ScreenH: 720}
	window := NewWindow(300, 200)
	window.Open(ctx, Win(Title("Box"), Size(300, 200), Content(primitives.Box())))
	window.Publish(ctx)
	pressX, pressY := window.x+window.width-2, window.y+window.height-2
	ctx.Input.SetMousePosition(pressX, pressY)
	ctx.Input.SetMouseButton(input.MouseButtonLeft, true)
	if !window.Update(ctx) || !window.Resizing() {
		t.Fatal("bottom-right grip did not start a resize")
	}
	ctx.Input.EndFrame()
	ctx.Input.SetMousePosition(pressX+40, pressY+30)
	if !window.Update(ctx) || window.width != 340 || window.height != 230 {
		t.Fatalf("resized = %dx%d, want 340x230", window.width, window.height)
	}
	ctx.Input.EndFrame()
	ctx.Input.SetMouseButton(input.MouseButtonLeft, false)
	window.Update(ctx)
	saved := window.SavedPosition()
	if !saved.Set || saved.Width != 340 || saved.Height != 230 {
		t.Fatalf("saved size = %+v", saved)
	}
	restored := Window{}
	restored.RestorePosition(ctx, saved)
	restored.EnsureWindow(300, 200)
	if restored.width != 340 || restored.height != 230 {
		t.Fatalf("restored size = %dx%d, want 340x230", restored.width, restored.height)
	}
}

func TestWindowResizeRejectsBelowMinimum(t *testing.T) {
	ctx := Context{Input: input.NewState(), UIManager: NewManager(), ScreenW: 1280, ScreenH: 720}
	window := NewWindow(300, 200)
	window.Open(ctx, Win(Title("Box"), Size(300, 200), Content(primitives.Box())))
	window.resizeTo(ctx, 10, 10)
	if window.width < windowMinWidth || window.height < windowMinHeight {
		t.Fatalf("size = %dx%d, below minimum", window.width, window.height)
	}
	if window.SavedPosition().Width != window.width {
		t.Fatal("clamped size was not saved")
	}
}

func TestSavedPositionWithoutResizeOmitsSize(t *testing.T) {
	window := NewWindow(300, 200)
	window.userMoved = true
	window.x, window.y = 12, 18
	saved := window.SavedPosition()
	if saved.Width != 0 || saved.Height != 0 {
		t.Fatalf("moved-only position = %+v", saved)
	}
}
