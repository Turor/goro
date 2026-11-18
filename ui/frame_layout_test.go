package ui

import (
	"image/color"
	"testing"

	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/uitest"
	"github.com/gogpu/ui/widget"
	"github.com/kivutar/goro/client"
	"github.com/kivutar/goro/input"
	"github.com/kivutar/goro/ui/rotheme"
)

func TestFrameLayoutReloadAppliesSizeAndColor(t *testing.T) {
	t.Cleanup(rotheme.ResetLayout)
	rotheme.ResetLayout()

	manager := NewManager()
	ctx := client.Context{Input: input.NewState(), UIManager: manager, ScreenW: 800, ScreenH: 600}
	w := NewWindow(300, 200)
	tree := func() widget.Widget {
		return Win(Title("Frame"), Size(300, 200), Footer(rotheme.Text("Ok")))
	}
	w.Open(ctx, tree())
	w.Publish(ctx)
	w.OnLayoutReload(func() {
		w.SetSize(300, 200)
		w.SetContent(tree())
	})
	if w.height != 200 || w.titleHeight != ROWindowTitleHeight {
		t.Fatalf("open size %dx%d title %d", w.width, w.height, w.titleHeight)
	}
	baseContent := w.content
	baseCart := cartWindowWidth
	baseText := TextColor

	if err := rotheme.SetToken("window", "title_height", "40"); err != nil {
		t.Fatal(err)
	}
	if err := rotheme.SetToken("footer", "stripe_color", "#ff00aa80"); err != nil {
		t.Fatal(err)
	}
	if err := rotheme.SetToken("surface", "text", "#112233"); err != nil {
		t.Fatal(err)
	}
	if err := rotheme.SetToken("scrollbar", "gutter", "20"); err != nil {
		t.Fatal(err)
	}
	if err := rotheme.SetToken("window", "screen_margin", "24"); err != nil {
		t.Fatal(err)
	}
	if characterWindowX != 24 || characterWindowY != 24 || basicMenuX != 24 {
		t.Fatalf("screen margin copies character %d,%d menu %d", characterWindowX, characterWindowY, basicMenuX)
	}
	if got := basicMenuY(); got != 24+characterWindowHeight+basicMenuFollowGap {
		t.Fatalf("basic menu y %d", got)
	}
	if x, y, _, _ := minimapBounds(800, 600); y != 24 || x != 800-minimapWidth-24 {
		t.Fatalf("minimap at %d,%d, want the 24px screen margin", x, y)
	}
	bar := roTitleBar(primitives.Box())
	if got := bar.Layout(widget.NewContext(), geometry.Constraints{MaxWidth: 300, MaxHeight: 80}); got.Height != 40 {
		t.Fatalf("title bar height %v, want 40 from the token at layout time", got.Height)
	}
	if n := manager.RebuildStaleWindows(); n != 1 {
		t.Fatalf("rebuilt %d windows, want 1", n)
	}
	if w.height != 212 || w.titleHeight != 40 {
		t.Fatalf("reloaded size height %d title %d, want 212 and 40", w.height, w.titleHeight)
	}
	if w.content == baseContent {
		t.Fatal("reload kept the window content built at the old title height")
	}
	if cartWindowWidth != baseCart+8 {
		t.Fatalf("cart width %d, want %d after a wider gutter", cartWindowWidth, baseCart+8)
	}
	if TextColor != (color.RGBA{R: 0x11, G: 0x22, B: 0x33, A: 255}) {
		t.Fatalf("text color %v", TextColor)
	}
	footer := findFooter(w.content)
	if footer == nil {
		t.Fatal("rebuilt window has no footer")
	}
	footer.SetBounds(geometry.NewRect(0, 0, 40, 8))
	canvas := &uitest.MockCanvas{}
	footer.Draw(widget.NewContext(), canvas)
	wantStripe := widget.RGBA8(0xff, 0x00, 0xaa, 0x80)
	if len(canvas.Rects) < 2 || canvas.Rects[1].Color != wantStripe || canvas.Rects[1].Bounds.Height() != footerStripeH {
		t.Fatalf("footer stripe = %+v, want %v height %v", canvas.Rects, wantStripe, footerStripeH)
	}

	rotheme.ResetLayout()
	if n := manager.RebuildStaleWindows(); n != 1 {
		t.Fatalf("restore rebuilt %d windows, want 1", n)
	}
	if w.height != 200 || w.titleHeight != ROWindowTitleHeight {
		t.Fatalf("restored height %d title %d", w.height, w.titleHeight)
	}
	if TextColor != baseText || cartWindowWidth != baseCart || footerStripeColor != widget.RGBA(1, 1, 1, 0.5) {
		t.Fatalf("defaults not restored: text %v cart %d stripe %v", TextColor, cartWindowWidth, footerStripeColor)
	}
	if _, my, _, _ := minimapBounds(800, 600); characterWindowX != 8 || basicMenuX != 8 || basicMenuY() != 8+characterWindowHeight+basicMenuFollowGap || my != 8 {
		t.Fatalf("screen margin not restored: character %d menu %d,%d minimap y %d", characterWindowX, basicMenuX, basicMenuY(), my)
	}
	if got := bar.Layout(widget.NewContext(), geometry.Constraints{MaxWidth: 300, MaxHeight: 80}); got.Height != float32(ROWindowTitleHeight) {
		t.Fatalf("restored title bar height %v", got.Height)
	}
}

type widgetChildren interface{ Children() []widget.Widget }

func findFooter(root widget.Widget) *roFooterWidget {
	if root == nil {
		return nil
	}
	if footer, ok := root.(*roFooterWidget); ok {
		return footer
	}
	parent, ok := root.(widgetChildren)
	if !ok {
		return nil
	}
	for _, child := range parent.Children() {
		if footer := findFooter(child); footer != nil {
			return footer
		}
	}
	return nil
}
