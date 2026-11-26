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
	wantW, wantH := lockResizeAspect(300, 200, 340, 230)
	if !window.Update(ctx) || window.width != wantW || window.height != wantH {
		t.Fatalf("resized = %dx%d, want %dx%d", window.width, window.height, wantW, wantH)
	}
	drift := window.width*200 - window.height*300
	if drift < 0 {
		drift = -drift
	}
	if drift > 300 {
		t.Fatalf("resized = %dx%d, aspect drifted from 300x200", window.width, window.height)
	}
	ctx.Input.EndFrame()
	ctx.Input.SetMouseButton(input.MouseButtonLeft, false)
	window.Update(ctx)
	saved := window.SavedPosition()
	if !saved.Set || saved.Width != wantW || saved.Height != wantH {
		t.Fatalf("saved size = %+v", saved)
	}
	restored := Window{}
	restored.RestorePosition(ctx, saved)
	restored.EnsureWindow(300, 200)
	if restored.width != wantW || restored.height != wantH {
		t.Fatalf("restored size = %dx%d, want %dx%d", restored.width, restored.height, wantW, wantH)
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
