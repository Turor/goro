package ui

import (
	"testing"

	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/widget"
	"github.com/kivutar/goro/ui/rotheme"
)

func TestGuildPartyMenusFollowLayoutReload(t *testing.T) {
	t.Cleanup(rotheme.ResetLayout)
	manager := NewManager()
	ctx := Context{UIManager: manager, ScreenW: 1024, ScreenH: 768}

	var settings PartySettingsWindow
	var create PartyCreateWindow
	var invite PartyInviteWindow
	var menu PetContextMenu
	var guild GuildWindow
	settings.Open(ctx)
	create.Open(ctx)
	invite.Open(ctx)
	menu.Open(ctx, 20, 30)
	guild.OpenWindow(ctx)

	if settings.width != 286 || create.width != 306 || invite.width != 286 || menu.width != 140 || guild.width != 400 {
		t.Fatalf("default sizes settings=%d create=%d invite=%d menu=%d guild=%d", settings.width, create.width, invite.width, menu.width, guild.width)
	}
	if menuFrameRadius(menu.content) != float32(contextMenuRadius) {
		t.Fatalf("menu radius = %v", menuFrameRadius(menu.content))
	}
	if !widgetHasBackground(guild.content, guildTabFrameBg) {
		t.Fatal("guild tab frame is missing its default background")
	}

	if err := rotheme.SetToken("party_settings", "width", "420"); err != nil {
		t.Fatal(err)
	}
	if err := rotheme.SetToken("party_create", "width", "480"); err != nil {
		t.Fatal(err)
	}
	if err := rotheme.SetToken("party_invite", "width", "360"); err != nil {
		t.Fatal(err)
	}
	if err := rotheme.SetToken("pet_context_menu", "width", "220"); err != nil {
		t.Fatal(err)
	}
	if err := rotheme.SetToken("context_menu", "radius", "12"); err != nil {
		t.Fatal(err)
	}
	if err := rotheme.SetToken("guild", "width", "560"); err != nil {
		t.Fatal(err)
	}
	if err := rotheme.SetToken("guild", "tab_frame_background", "#102030"); err != nil {
		t.Fatal(err)
	}
	tuned := guildTabFrameBg
	if n := manager.RebuildStaleWindows(); n != 5 {
		t.Fatalf("rebuilt %d windows, want 5", n)
	}
	if settings.width != 420 || create.width != 480 || invite.width != 360 || menu.width != 220 || guild.width != 560 {
		t.Fatalf("tuned sizes settings=%d create=%d invite=%d menu=%d guild=%d", settings.width, create.width, invite.width, menu.width, guild.width)
	}
	if menuFrameRadius(menu.content) != 12 {
		t.Fatalf("tuned menu radius = %v, want 12", menuFrameRadius(menu.content))
	}
	if !widgetHasBackground(guild.content, tuned) {
		t.Fatal("guild tab frame did not pick up the colour token")
	}

	rotheme.ResetLayout()
	if n := manager.RebuildStaleWindows(); n != 5 {
		t.Fatalf("restore rebuilt %d windows, want 5", n)
	}
	if settings.width != 286 || create.width != 306 || invite.width != 286 || menu.width != 140 || guild.width != 400 {
		t.Fatalf("restored sizes settings=%d create=%d invite=%d menu=%d guild=%d", settings.width, create.width, invite.width, menu.width, guild.width)
	}
	if menuFrameRadius(menu.content) != float32(contextMenuRadius) || contextMenuRadius != 5 {
		t.Fatalf("restored radius = %v token %d", menuFrameRadius(menu.content), contextMenuRadius)
	}
	if widgetHasBackground(guild.content, tuned) || !widgetHasBackground(guild.content, guildTabFrameBg) {
		t.Fatal("guild tab frame colour did not return to the default")
	}
}

func menuFrameRadius(root widget.Widget) float32 {
	frame, ok := root.(*windowFrame)
	if !ok || frame.BoxWidget == nil {
		return -1
	}
	return frame.Style().Radius
}

func widgetHasBackground(root widget.Widget, color widget.Color) bool {
	if root == nil {
		return false
	}
	if box, ok := root.(*primitives.BoxWidget); ok && box.Style().Background == color {
		return true
	}
	type kids interface{ Children() []widget.Widget }
	if k, ok := root.(kids); ok {
		for _, child := range k.Children() {
			if widgetHasBackground(child, color) {
				return true
			}
		}
	}
	return false
}
