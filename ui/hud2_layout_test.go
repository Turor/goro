package ui

import (
	"image/color"
	"testing"
	"time"

	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/uitest"
	"github.com/gogpu/ui/widget"
	"github.com/kivutar/goro/client"
	"github.com/kivutar/goro/input"
	"github.com/kivutar/goro/session"
	"github.com/kivutar/goro/ui/rotheme"
	worldstate "github.com/kivutar/goro/world"
)

func TestHUDLayoutReloadAppliesSizeAndColor(t *testing.T) {
	rotheme.ResetLayout()
	t.Cleanup(rotheme.ResetLayout)

	manager := NewManager()
	ctx := client.Context{
		Input:     input.NewState(),
		UIManager: manager,
		ScreenW:   800,
		ScreenH:   600,
		Session:   &session.Session{},
	}

	bar := &ShortcutBar{}
	bar.Update(ctx, nil)
	hotbarW, hotbarH := bar.window.width, bar.window.height
	content := bar.window.content
	if hotbarW == 0 || content == nil {
		t.Fatal("hotbar did not open")
	}
	if err := rotheme.SetToken("hotbar", "slot_size", "44"); err != nil {
		t.Fatal(err)
	}
	if n := manager.RebuildStaleWindows(); n != 0 {
		t.Fatalf("hotbar rebuilds from its Update path, RebuildStaleWindows rebuilt %d", n)
	}
	bar.Update(ctx, nil)
	if bar.window.width == hotbarW || bar.window.height == hotbarH || bar.window.content == content {
		t.Fatalf("hotbar stayed at %dx%d content %p", bar.window.width, bar.window.height, bar.window.content)
	}

	var notes LevelUpNotifications
	notes.NotifyBase()
	notes.Update(ctx)
	levelW := notes.base.root.(*positionedOverlay).Bounds().Width()
	icon := newLevelUpNotificationWidget(nil)
	icon.Layout(widget.NewContext(), geometry.Tight(geometry.Sz(levelW, levelW)))
	beforeCanvas := &uitest.MockCanvas{}
	icon.Draw(widget.NewContext(), beforeCanvas)
	if len(beforeCanvas.Rects) == 0 {
		t.Fatal("level-up fallback drew no rect")
	}
	if err := rotheme.SetToken("level_up", "size", "60"); err != nil {
		t.Fatal(err)
	}
	if err := rotheme.SetToken("level_up", "fill", "#ff00aa"); err != nil {
		t.Fatal(err)
	}
	notes.Update(ctx)
	if got := notes.base.root.(*positionedOverlay).Bounds().Width(); got == levelW || got != 60 {
		t.Fatalf("level-up width = %v, want 60 (was %v)", got, levelW)
	}
	afterCanvas := &uitest.MockCanvas{}
	icon.Draw(widget.NewContext(), afterCanvas)
	if len(afterCanvas.Rects) == 0 {
		t.Fatal("level-up fallback drew no rect after the colour change")
	}
	uitest.AssertColorEqual(t, afterCanvas.Rects[0].Color, widget.RGBA8(255, 0, 170, 255))
	if afterCanvas.Rects[0].Color == beforeCanvas.Rects[0].Color {
		t.Fatal("level-up fill colour did not change")
	}

	world := worldstate.New()
	world.MapProperty = worldstate.MapPropertyFreePvPZone
	world.Player.PvPRank = 3
	world.Player.PvPTotal = 12
	world.Player.HasPvPRanking = true
	pctx := ctx
	pctx.World = world
	var counter PvPCounter
	counter.Update(pctx)
	pvpW := counter.root.(*positionedOverlay).Bounds().Width()
	if err := rotheme.SetToken("pvp_counter", "width", "300"); err != nil {
		t.Fatal(err)
	}
	counter.Update(pctx)
	if got := counter.root.(*positionedOverlay).Bounds().Width(); got == pvpW || got != 300 {
		t.Fatalf("pvp counter width = %v, want 300 (was %v)", got, pvpW)
	}

	_, _, _, pickupH := itemPickupNotificationBounds(800, "Apple")
	beforePixel := notificationFillPixel(t)
	if err := rotheme.SetToken("item_pickup", "height", "48"); err != nil {
		t.Fatal(err)
	}
	if err := rotheme.SetToken("item_pickup", "fill", "#112233"); err != nil {
		t.Fatal(err)
	}
	if _, _, _, got := itemPickupNotificationBounds(800, "Apple"); got == pickupH || got != 48 {
		t.Fatalf("pickup height = %d, want 48 (was %d)", got, pickupH)
	}
	if got := notificationFillPixel(t); got == beforePixel {
		t.Fatalf("pickup fill stayed %v", got)
	}

	_, _, _, iconH := statusIconOverlayBounds(800, 600, 2)
	now := time.Unix(20, 0)
	effect := session.StatusEffect{HasDuration: true, StartedAt: now, ExpiresAt: now.Add(time.Second)}
	statusWidget := newStatusIconsWidget()
	statusWidget.now = now
	if err := rotheme.SetToken("status_icons", "spacing", "50"); err != nil {
		t.Fatal(err)
	}
	if err := rotheme.SetToken("status_icons", "duration_fill", "#ff0000"); err != nil {
		t.Fatal(err)
	}
	if _, _, _, got := statusIconOverlayBounds(800, 600, 2); got == iconH {
		t.Fatalf("status icon strip height stayed %d", got)
	}
	statusCanvas := &uitest.MockCanvas{}
	statusWidget.drawStatusIcon(statusCanvas, 1, effect, 0, 0)
	uitest.AssertColorEqual(t, statusCanvas.Rects[len(statusCanvas.Rects)-1].Color, widget.RGBA8(255, 0, 0, 255))

	rotheme.ResetLayout()
	bar.Update(ctx, nil)
	if bar.window.width != hotbarW || bar.window.height != hotbarH {
		t.Fatalf("hotbar reset size = %dx%d, want %dx%d", bar.window.width, bar.window.height, hotbarW, hotbarH)
	}
	notes.Update(ctx)
	if got := notes.base.root.(*positionedOverlay).Bounds().Width(); got != levelW {
		t.Fatalf("level-up reset width = %v, want %v", got, levelW)
	}
	counter.Update(pctx)
	if got := counter.root.(*positionedOverlay).Bounds().Width(); got != pvpW {
		t.Fatalf("pvp counter reset width = %v, want %v", got, pvpW)
	}
	if _, _, _, got := itemPickupNotificationBounds(800, "Apple"); got != pickupH {
		t.Fatalf("pickup reset height = %d, want %d", got, pickupH)
	}
	if _, _, _, got := statusIconOverlayBounds(800, 600, 2); got != iconH {
		t.Fatalf("status icon reset height = %d, want %d", got, iconH)
	}
}

func notificationFillPixel(t *testing.T) color.RGBA {
	t.Helper()
	img := cachedItemPickupNotificationSurface(80, itemPickupNotificationH)
	if img == nil || img.RGBA() == nil {
		t.Fatal("pickup surface was not drawn")
	}
	rgba := img.RGBA()
	off := rgba.PixOffset(40, itemPickupNotificationH/2)
	return color.RGBA{R: rgba.Pix[off], G: rgba.Pix[off+1], B: rgba.Pix[off+2], A: rgba.Pix[off+3]}
}
