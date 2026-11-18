package ui

import (
	"testing"

	"github.com/gogpu/ui/core/button"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/uitest"
	"github.com/gogpu/ui/widget"
	"github.com/kivutar/goro/client"
	"github.com/kivutar/goro/input"
	"github.com/kivutar/goro/ui/rotheme"
)

func TestLoginScreensRebuildOnLayoutReload(t *testing.T) {
	t.Cleanup(rotheme.ResetLayout)
	manager := NewManager()
	ctx := client.Context{Input: input.NewState(), UIManager: manager, ScreenW: 1280, ScreenH: 720}

	login := NewLoginWindow(ctx, "user", "", false, LoginWindowCallbacks{})
	login.Publish(ctx)
	service := NewServiceWindow(ctx, []string{"Prontera", "Morroc"}, ServiceWindowOptions{Title: "Server"}, ServiceWindowCallbacks{})
	service.Publish(ctx)
	charSelect := NewCharacterSelectWindow(ctx, CharacterSelectWindowOptions{MaxSlots: 9}, CharacterSelectWindowCallbacks{})
	charSelect.Publish(ctx)
	charCreate := NewCharacterCreateWindow(ctx, CharacterCreateWindowOptions{}, CharacterCreateWindowCallbacks{})
	charCreate.Publish(ctx)
	var confirm ConfirmModal
	confirm.Open(ctx, "Confirm", "Leave?", nil, nil)
	var prompt TextPromptWindow
	prompt.Open(ctx, "Prompt", "Name", "", 24)
	var storage StoragePasswordWindow
	storage.OpenPrompt(ctx, StoragePasswordModeEnter, "Enter the password")
	var escape EscapeMenu
	escape.Toggle(ctx)
	t.Cleanup(func() {
		login.Close()
		service.Close()
		charSelect.Close()
		charCreate.Close()
		confirm.Close(ctx)
		prompt.Close()
		storage.Close()
		escape.Close()
	})

	windows := []*Window{
		&login.Window, &service.Window, &charSelect.Window, &charCreate.Window,
		&confirm.Window, &prompt.Window, &storage.Window, &escape.Window,
	}
	widths := make([]int, len(windows))
	contents := make([]widget.Widget, len(windows))
	for i, w := range windows {
		widths[i] = w.width
		contents[i] = w.content
	}

	sets := []struct{ section, key, value string }{
		{"login", "width", "400"},
		{"service", "width", "360"},
		{"char_select", "width", "640"},
		{"char_select", "selected_background", "#ff00ff"},
		{"char_create", "width", "640"},
		{"confirm", "width", "360"},
		{"text_prompt", "width", "400"},
		{"storage_password", "width", "400"},
		{"escape", "width", "320"},
	}
	for _, set := range sets {
		if err := rotheme.SetToken(set.section, set.key, set.value); err != nil {
			t.Fatalf("set %s.%s: %v", set.section, set.key, err)
		}
	}
	if n := manager.RebuildStaleWindows(); n != len(windows) {
		t.Fatalf("rebuilt %d windows, want %d", n, len(windows))
	}
	want := []int{400, 360, 640, 640, 360, 400, 400, 320}
	for i, w := range windows {
		if w.width != want[i] {
			t.Fatalf("window %d width = %d, want %d (was %d)", i, w.width, want[i], widths[i])
		}
		if w.content == nil || w.content == contents[i] {
			t.Fatalf("window %d content was not rebuilt", i)
		}
	}
	if color := selectedSlotBackground(); color != widget.RGBA8(255, 0, 255, 255) {
		t.Fatalf("selected slot background = %+v, want #ff00ff", color)
	}

	rotheme.ResetLayout()
	if n := manager.RebuildStaleWindows(); n != len(windows) {
		t.Fatalf("restore rebuilt %d windows, want %d", n, len(windows))
	}
	defaults := []int{304, 280, 576, 576, 286, 306, 326, 252}
	for i, w := range windows {
		if w.width != defaults[i] {
			t.Fatalf("window %d restored width = %d, want %d", i, w.width, defaults[i])
		}
	}
	if color := selectedSlotBackground(); color != widget.RGBA8(222, 237, 252, 255) {
		t.Fatalf("selected slot background = %+v, want default", color)
	}
}

func selectedSlotBackground() widget.Color {
	canvas := &uitest.MockCanvas{}
	characterSelectSlotPainter{selected: true}.PaintButton(canvas, button.PaintState{
		Bounds: geometry.NewRect(0, 0, 40, 40),
	})
	if len(canvas.Rects) == 0 {
		return widget.Color{}
	}
	return canvas.Rects[0].Color
}
