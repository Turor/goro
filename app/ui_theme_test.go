package app

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/kivutar/goro/config"
	"github.com/kivutar/goro/ui/rotheme"
)

func TestStartupThemeSitsUnderUILayoutAndSurvivesReload(t *testing.T) {
	t.Cleanup(func() {
		rotheme.SetThemesDir("")
		if _, err := rotheme.SetTheme(rotheme.DefaultThemeID); err != nil {
			t.Error(err)
		}
		rotheme.ResetLayout()
	})
	defaultBody := rotheme.Default.Colors.WindowBody
	dir := t.TempDir()
	layout := filepath.Join(dir, config.DefaultUILayoutFile)
	if err := os.WriteFile(layout, []byte("[theme]\ntitle_text = #010203\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg := config.Config{
		ConfigPath: filepath.Join(dir, "goro.ini"),
		UILayout:   config.UILayoutConfig{Theme: "dark", Watch: true},
	}
	loadUILayout(cfg)
	if rotheme.ActiveTheme() != "dark" || rotheme.ThemesDir() != cfg.ThemesDir() {
		t.Fatalf("theme=%q dir=%q", rotheme.ActiveTheme(), rotheme.ThemesDir())
	}
	if rotheme.Default.Colors.WindowBody == defaultBody {
		t.Fatal("dark theme not applied at startup")
	}
	if r, g, b, _ := rotheme.Default.Colors.TitleText.RGBA8(); r != 1 || g != 2 || b != 3 {
		t.Fatalf("ui-layout.ini title_text lost to the theme: %d,%d,%d", r, g, b)
	}
	// The --ui-dev watcher removing ui-layout.ini keeps the theme.
	now := time.Now()
	w := newUILayoutWatcher(cfg, now)
	if err := os.Remove(layout); err != nil {
		t.Fatal(err)
	}
	if !w.Update(nil, now.Add(uiLayoutPollInterval)) {
		t.Fatal("watcher did not notice the removed file")
	}
	if rotheme.ActiveTheme() != "dark" || rotheme.Default.Colors.WindowBody == defaultBody {
		t.Fatal("reload dropped the theme")
	}
	// An unknown saved theme falls back to the default theme.
	if _, err := rotheme.SetTheme(rotheme.DefaultThemeID); err != nil {
		t.Fatal(err)
	}
	cfg.UILayout.Theme = "gone"
	loadUILayout(cfg)
	if rotheme.ActiveTheme() != rotheme.DefaultThemeID || rotheme.Default.Colors.WindowBody != defaultBody {
		t.Fatal("an unknown theme must leave the default theme")
	}
}
