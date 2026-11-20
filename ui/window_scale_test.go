package ui

import (
	"testing"

	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/widget"

	"github.com/kivutar/goro/client"
	"github.com/kivutar/goro/input"
	"github.com/kivutar/goro/session"
)

func TestResizedWindowsScaleElements(t *testing.T) {
	ctx := Context{Input: input.NewState(), UIManager: NewManager(), ScreenW: 1600, ScreenH: 1200, Session: &session.Session{}}
	clientCtx := client.Context{Input: ctx.Input, UIManager: ctx.UIManager, ScreenW: ctx.ScreenW, ScreenH: ctx.ScreenH}
	gameCtx := ctx

	t.Run("escape", func(t *testing.T) {
		menu := &EscapeMenu{}
		menu.Toggle(clientCtx)
		menu.Update(clientCtx)
		assertWindowScales(t, &menu.Window, 0.6)
	})

	t.Run("basic menu", func(t *testing.T) {
		menu := &BasicMenu{}
		menu.Update(clientCtx, BasicMenuCallbacks{})
		assertWindowScales(t, &menu.Window, 0.6)
	})

	t.Run("gear", func(t *testing.T) {
		window := &EquipmentWindow{}
		window.Toggle(Context{Session: &session.Session{}, ScreenW: ctx.ScreenW, ScreenH: ctx.ScreenH, Input: ctx.Input, UIManager: ctx.UIManager})
		assertWindowScales(t, &window.Window, 0.6)
	})

	t.Run("inventory", func(t *testing.T) {
		window := &InventoryBagWindow{}
		window.Toggle(Context{Session: &session.Session{}, ScreenW: ctx.ScreenW, ScreenH: ctx.ScreenH, Input: ctx.Input, UIManager: ctx.UIManager})
		assertWindowScales(t, &window.Window, 0.6)
	})

	t.Run("skill", func(t *testing.T) {
		window := &SkillWindow{}
		window.Toggle(gameCtx)
		assertWindowScales(t, &window.Window, 0.6)
	})
	for _, tc := range []struct {
		name string
		open func() *Window
	}{
		{"friends", func() *Window { w := &FriendsWindow{}; w.Toggle(gameCtx); return &w.Window }},
		{"guild", func() *Window { w := &GuildWindow{}; w.Toggle(gameCtx); return &w.Window }},
		{"quest", func() *Window { w := &QuestWindow{}; w.Toggle(gameCtx); return &w.Window }},
		{"cart", func() *Window { w := &CartWindow{}; w.OpenWindow(gameCtx); return &w.Window }},
		{"equip switch", func() *Window { w := &EquipSwitchWindow{}; w.Toggle(gameCtx); return &w.Window }},
		{"rodex", func() *Window { w := &RodexWindow{}; w.Toggle(gameCtx); return &w.Window }},
		{"warpra", func() *Window { w := &WarpraWindow{}; w.Toggle(gameCtx); return &w.Window }},
		{"search store", func() *Window { w := &SearchStoreWindow{}; w.Toggle(gameCtx); return &w.Window }},
		{"roulette", func() *Window { w := &RouletteWindow{}; w.Toggle(gameCtx); return &w.Window }},
		{"instance", func() *Window {
			w := &InstanceWindow{}
			gameCtx.Session.Instance.Name = "Test"
			w.Toggle(gameCtx)
			return &w.Window
		}},
		{"stats", func() *Window { w := &StatsWindow{}; w.Toggle(gameCtx); return &w.Window }},
		{"mail", func() *Window { w := &MailWindow{}; w.Open(gameCtx, nil); return &w.Window }},
		{"bank", func() *Window { w := &BankWindow{}; w.Open(gameCtx, 0, 0); return &w.Window }},
		{"storage", func() *Window { w := &StorageWindow{}; w.OpenWindow(gameCtx); return &w.Window }},
		{"attendance", func() *Window { w := &AttendanceWindow{}; w.Open(gameCtx, 0); return &w.Window }},
		{"achievement", func() *Window { w := &AchievementWindow{}; w.Open(gameCtx); return &w.Window }},
		{"companion", func() *Window { w := &CompanionWindow{}; w.OpenWindow(gameCtx); return &w.Window }},
		{"shortcut", func() *Window { w := &ShortcutBar{}; w.Update(gameCtx, nil); return &w.window }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			assertWindowScales(t, tc.open(), 0.6)
		})
	}
}

func assertWindowScales(t *testing.T, window *Window, factor float64) {
	t.Helper()
	root := window.content
	layoutWidget(root, window.width, window.height)
	before := explicitSizes(root)
	if len(before) == 0 {
		t.Fatal("window has no sized elements")
	}
	width := int(float64(window.width) * factor)
	height := int(float64(window.height) * factor)
	if width < windowMinWidth {
		width = windowMinWidth
	}
	if height < windowMinHeight {
		height = windowMinHeight
	}
	window.resizeTo(Context{Input: input.NewState(), UIManager: NewManager(), ScreenW: 1600, ScreenH: 1200}, width, height)
	root = window.content
	layoutWidget(root, window.width, window.height)
	after := explicitSizes(root)
	shrunk := 0
	for key, was := range before {
		now, ok := after[key]
		if !ok || was < 8 {
			continue
		}
		if now < was*0.9 {
			shrunk++
		}
	}
	if shrunk == 0 {
		t.Fatalf("no element shrank; before %v after %v", before, after)
	}
}

func layoutWidget(root widget.Widget, width, height int) {
	if root == nil {
		return
	}
	root.Layout(widget.NewContext(), geometry.Tight(geometry.Sz(float32(width), float32(height))))
}

func explicitSizes(root widget.Widget) map[string]float32 {
	out := make(map[string]float32)
	var walk func(widget.Widget, string)
	walk = func(current widget.Widget, path string) {
		if current == nil {
			return
		}
		if box, ok := current.(*primitives.BoxWidget); ok {
			style := box.Style()
			if style.ExplicitWidth >= 8 {
				out[path+".w"] = style.ExplicitWidth
			}
			if style.ExplicitHeight >= 8 {
				out[path+".h"] = style.ExplicitHeight
			}
		}
		if text, ok := current.(*primitives.TextWidget); ok && text.Style().FontSize >= 8 {
			out[path+".font"] = text.Style().FontSize
		}
		for i, child := range current.Children() {
			walk(child, path+"/"+string(rune('a'+i)))
		}
	}
	walk(root, "f")
	return out
}
