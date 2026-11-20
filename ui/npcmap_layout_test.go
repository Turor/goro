package ui

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/uitest"
	"github.com/gogpu/ui/widget"
	"github.com/kivutar/goro/network"
	"github.com/kivutar/goro/res"
	"github.com/kivutar/goro/ui/rotheme"
)

func TestNPCMapWindowsFollowLayoutTokens(t *testing.T) {
	t.Cleanup(rotheme.ResetLayout)
	ctx, manager, _ := newWindowInstanceTest()

	var input NPCDialog
	input.Apply(network.NPCDialog{Kind: network.NPCDialogNumberInput, NPCID: 100})
	if !input.Update(ctx) || !input.inputWindow.IsOpen() {
		t.Fatal("number input did not open")
	}
	var menu NPCDialog
	menu.Apply(network.NPCDialog{Kind: network.NPCDialogMenu, NPCID: 7, Options: []string{"a", "b", "c", "d"}})
	if !menu.Update(ctx) || !menu.menuWindow.IsOpen() {
		t.Fatal("menu did not open")
	}

	book := &BookWindow{}
	if err := book.Open(bookLayoutContext(t, ctx), 7277, "Empty Book"); err != nil {
		t.Fatal(err)
	}
	cardCtx := ctx
	cardCtx.Resources = cardIllustrationTestManager(t)
	card := &CardIllustrationWindow{}
	if err := card.Open(cardCtx, 4001, "Poring Card"); err != nil {
		t.Fatal(err)
	}
	emote := &EmoteWindow{}
	emote.OpenWindow(ctx, nil)

	mapCtx := worldMapTestContext(t)
	mapCtx.UIManager = manager
	mapCtx.Input = ctx.Input
	atlas := &WorldMapWindow{}
	if err := atlas.OpenWindow(mapCtx); err != nil {
		t.Fatal(err)
	}

	defaultInputW := input.inputWindow.width
	defaultMenuW := menu.menuWindow.width
	defaultBookW := book.width
	defaultCardW := card.width
	defaultEmoteW := emote.width
	defaultHover := worldMapHoverFill
	defaultPreviewW := previewPanelWidth(t, atlas)

	if err := rotheme.SetToken("npc_dialog", "input_number_width", "240"); err != nil {
		t.Fatal(err)
	}
	if err := rotheme.SetToken("npc_dialog", "input_background", "#112233"); err != nil {
		t.Fatal(err)
	}
	if err := rotheme.SetToken("npc_dialog", "menu_width", "400"); err != nil {
		t.Fatal(err)
	}
	if err := rotheme.SetToken("book", "width", "620"); err != nil {
		t.Fatal(err)
	}
	if err := rotheme.SetToken("card_illustration", "width", "360"); err != nil {
		t.Fatal(err)
	}
	if err := rotheme.SetToken("emote", "width", "320"); err != nil {
		t.Fatal(err)
	}
	if err := rotheme.SetToken("world_map", "preview_width", "100"); err != nil {
		t.Fatal(err)
	}
	if err := rotheme.SetToken("world_map", "hover_fill", "#00ff0080"); err != nil {
		t.Fatal(err)
	}
	if n := manager.RebuildStaleWindows(); n < 6 {
		t.Fatalf("rebuilt %d windows, want the open npc, book, card, emote and map windows", n)
	}

	box, ok := input.inputWindow.content.(*primitives.BoxWidget)
	if !ok {
		t.Fatalf("input content = %T", input.inputWindow.content)
	}
	if input.inputWindow.width != 240 || box.Style().Background != npcInputBg || box.Style().Background == widget.RGBA8(255, 255, 255, 255) {
		t.Fatalf("input width %d background %+v, token %+v", input.inputWindow.width, box.Style().Background, npcInputBg)
	}
	if menu.menuWindow.width != 400 || book.width != 620 || card.width != 360 || emote.width != 320 {
		t.Fatalf("sizes menu %d book %d card %d emote %d", menu.menuWindow.width, book.width, card.width, emote.width)
	}
	if got := previewPanelWidth(t, atlas); got != 100 || !hoverDrawn(t, atlas, worldMapHoverFill) || worldMapHoverFill == defaultHover {
		t.Fatalf("preview width %.1f hover %+v", got, worldMapHoverFill)
	}

	rotheme.ResetLayout()
	if n := manager.RebuildStaleWindows(); n < 6 {
		t.Fatalf("restore rebuilt %d windows", n)
	}
	box = input.inputWindow.content.(*primitives.BoxWidget)
	if input.inputWindow.width != defaultInputW || menu.menuWindow.width != defaultMenuW || book.width != defaultBookW || card.width != defaultCardW || emote.width != defaultEmoteW {
		t.Fatalf("defaults not restored: input %d menu %d book %d card %d emote %d", input.inputWindow.width, menu.menuWindow.width, book.width, card.width, emote.width)
	}
	if box.Style().Background != npcInputBg || npcInputBg != widget.RGBA8(255, 255, 255, 255) {
		t.Fatalf("input background %+v", box.Style().Background)
	}
	if got := previewPanelWidth(t, atlas); got != defaultPreviewW || !hoverDrawn(t, atlas, defaultHover) {
		t.Fatalf("map restore width %.1f hover %+v, want %.1f %+v", got, worldMapHoverFill, defaultPreviewW, defaultHover)
	}
}

func bookLayoutContext(t *testing.T, ctx Context) Context {
	t.Helper()
	root := t.TempDir()
	path := filepath.Join(root, "data", "book", "7277.txt")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	ctx.Resources = &res.Manager{Root: root}
	return ctx
}

func previewPanelWidth(t *testing.T, atlas *WorldMapWindow) float32 {
	t.Helper()
	canvas := drawAtlas(t, atlas)
	if len(canvas.RoundRects) == 0 {
		t.Fatal("world map drew no preview panel")
	}
	return canvas.RoundRects[0].Bounds.Width()
}

func hoverDrawn(t *testing.T, atlas *WorldMapWindow, color widget.Color) bool {
	t.Helper()
	for _, rect := range drawAtlas(t, atlas).Rects {
		if rect.Color == color {
			return true
		}
	}
	return false
}

func drawAtlas(t *testing.T, atlas *WorldMapWindow) *uitest.MockCanvas {
	t.Helper()
	if atlas.board == nil || atlas.scaled == nil || len(atlas.board.entries) == 0 {
		t.Fatal("world map has no board")
	}
	atlas.board.hovered = 0
	size := atlas.scaled.Bounds().Size()
	tight := geometry.Tight(geometry.Sz(float32(size.X), float32(size.Y)))
	if parent := atlas.board.Parent(); parent != nil {
		parent.Layout(widget.NewContext(), tight)
	} else {
		atlas.board.Layout(widget.NewContext(), tight)
	}
	atlas.board.placeHighlight()
	canvas := &uitest.MockCanvas{}
	atlas.board.Draw(widget.NewContext(), canvas)
	if atlas.board.highlight != nil {
		atlas.board.highlight.Draw(widget.NewContext(), canvas)
	}
	if atlas.board.previewLayer != nil {
		atlas.board.previewLayer.Draw(widget.NewContext(), canvas)
	}
	return canvas
}
