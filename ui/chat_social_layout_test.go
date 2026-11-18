package ui

import (
	"testing"

	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/widget"
	"github.com/kivutar/goro/client"
	"github.com/kivutar/goro/input"
	"github.com/kivutar/goro/session"
	"github.com/kivutar/goro/ui/rotheme"
)

func TestChatSocialLayoutReload(t *testing.T) {
	t.Cleanup(rotheme.ResetLayout)
	manager := NewManager()
	ctx := client.Context{
		Input:     input.NewState(),
		UIManager: manager,
		Session:   session.New(),
		ScreenW:   800,
		ScreenH:   600,
	}

	var friends FriendsWindow
	friends.OpenWindow(ctx)
	var settings FriendSettingsWindow
	settings.Open(ctx)
	var whisper WhisperWindow
	whisper.Open(ctx, "Alice")
	var room ChatRoomWindow
	room.Open(ctx, "Room", 20, true, []string{"Alice"})
	var create ChatRoomCreateWindow
	create.Open(ctx)
	var shortcuts ChatShortcutsWindow
	shortcuts.Toggle(ctx, nil, nil)

	defaults := []int{friends.width, settings.width, whisper.width, room.width, create.width, shortcuts.width}
	panel := whisperPanelBg

	set := func(section, key, value string) {
		t.Helper()
		if err := rotheme.SetToken(section, key, value); err != nil {
			t.Fatal(err)
		}
	}
	set("friends", "width", "400")
	set("friend_settings", "width", "420")
	set("whisper", "width", "440")
	set("whisper", "panel_background", "#112233")
	set("chat_room", "width", "640")
	set("chat_room_create", "width", "420")
	set("chat_shortcuts", "width", "400")

	if n := manager.RebuildStaleWindows(); n != 6 {
		t.Fatalf("rebuilt %d windows, want 6", n)
	}
	if friends.width != 400 || settings.width != 420 || whisper.width != 440 || room.width != 640 || create.width != 420 || shortcuts.width != 400 {
		t.Fatalf("widths friends=%d settings=%d whisper=%d room=%d create=%d shortcuts=%d", friends.width, settings.width, whisper.width, room.width, create.width, shortcuts.width)
	}
	tint := widget.RGBA8(0x11, 0x22, 0x33, 0xff)
	if !layoutTreeHasBackground(whisper.content, tint) || layoutTreeHasBackground(whisper.content, panel) {
		t.Fatal("whisper panel did not take the reloaded colour")
	}

	rotheme.ResetLayout()
	if n := manager.RebuildStaleWindows(); n != 6 {
		t.Fatalf("restore rebuilt %d windows, want 6", n)
	}
	got := []int{friends.width, settings.width, whisper.width, room.width, create.width, shortcuts.width}
	for i := range defaults {
		if got[i] != defaults[i] {
			t.Fatalf("restored widths %v, want %v", got, defaults)
		}
	}
	if !layoutTreeHasBackground(whisper.content, whisperPanelBg) || layoutTreeHasBackground(whisper.content, tint) {
		t.Fatal("whisper panel colour was not restored")
	}
}

func layoutTreeHasBackground(root widget.Widget, color widget.Color) bool {
	if root == nil {
		return false
	}
	if box, ok := root.(*primitives.BoxWidget); ok && box.Style().Background == color {
		return true
	}
	parent, ok := root.(interface{ Children() []widget.Widget })
	if !ok {
		return false
	}
	for _, child := range parent.Children() {
		if layoutTreeHasBackground(child, color) {
			return true
		}
	}
	return false
}
