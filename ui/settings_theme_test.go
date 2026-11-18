package ui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gogpu/gpucontext"
	uiapp "github.com/gogpu/ui/app"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/uitest"
	"github.com/kivutar/goro/client"
	"github.com/kivutar/goro/config"
	"github.com/kivutar/goro/input"
	"github.com/kivutar/goro/session"
	"github.com/kivutar/goro/ui/rotheme"
)

func TestSettingsThemeDropdownSwitchesLiveAndPersists(t *testing.T) {
	dir := t.TempDir()
	themes := filepath.Join(dir, config.ThemesDirName)
	if err := os.Mkdir(themes, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(themes, "ocean.ini"), []byte("# name: Ocean\n[palette]\naccent = #1060a0\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg := config.Config{ConfigPath: filepath.Join(dir, "goro.ini")}
	rotheme.SetThemesDir(cfg.ThemesDir())
	t.Cleanup(func() {
		rotheme.SetThemesDir("")
		if _, err := rotheme.SetTheme(rotheme.DefaultThemeID); err != nil {
			t.Error(err)
		}
		rotheme.ResetLayout()
	})
	defaultBody := rotheme.Default.Colors.WindowBody

	app := uiapp.New(uiapp.WithWindowProvider(gpucontext.NullWindowProvider{W: 640, H: 480}))
	manager := NewManager()
	manager.SetUIApp(basicMenuTestApp{app: app})
	ctx := client.Context{
		Input: input.NewState(), Session: session.New(), UIManager: manager, UIApp: basicMenuTestApp{app: app},
		ScreenW: 640, ScreenH: 480, Config: cfg,
	}
	var window SettingsWindow
	window.OpenWindow(ctx)
	draw := func() map[string]geometry.Rect {
		manager.RebuildStaleWindows()
		app.Frame()
		canvas := &settingsLabelCanvas{labels: make(map[string]geometry.Rect)}
		app.Window().DrawTo(canvas)
		return canvas.labels
	}
	click := func(label string) {
		t.Helper()
		bounds, ok := draw()[label]
		if !ok {
			t.Fatalf("control %q is not drawn", label)
		}
		p := bounds.Center()
		app.Window().HandleEvent(uitest.Click(p.X, p.Y))
		app.Window().HandleEvent(uitest.Release(p.X, p.Y))
	}
	click("Gameplay")
	labels := draw()
	theme, hasTheme := labels["Theme"]
	current, hasCurrent := labels["Default"]
	if !hasTheme || !hasCurrent {
		t.Fatalf("Theme row or its selection is missing: %v", labels)
	}
	if theme.Center().Y != current.Center().Y {
		t.Fatalf("Theme label and dropdown are misaligned: %v %v", theme, current)
	}
	click("Default")
	menu := draw()
	for _, name := range []string{"Dark", "Complementary", "Analogous", "Triadic", "Ocean"} {
		if _, ok := menu[name]; !ok {
			t.Fatalf("theme %q is not in the open menu: %v", name, menu)
		}
	}
	click("Dark")
	if rotheme.ActiveTheme() != "dark" {
		t.Fatalf("active theme = %q after choosing Dark", rotheme.ActiveTheme())
	}
	if rotheme.Default.Colors.WindowBody == defaultBody {
		t.Fatal("window body color did not change with the Dark theme")
	}
	if _, ok := draw()["Dark"]; !ok || window.tab != settingsTabGameplay {
		t.Fatal("dropdown does not show Dark or the tab changed")
	}
	data, err := os.ReadFile(cfg.ConfigPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "[ui_layout]\ntheme = dark\n") {
		t.Fatalf("goro.ini does not keep the theme:\n%s", data)
	}
	loaded, err := config.LoadConfig([]string{"--config", cfg.ConfigPath})
	if err != nil {
		t.Fatal(err)
	}
	if loaded.UILayout.Theme != "dark" {
		t.Fatalf("reloaded theme = %q", loaded.UILayout.Theme)
	}
}
