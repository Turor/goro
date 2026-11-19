package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSaveUIThemeRoundTripsAndKeepsOtherSettings(t *testing.T) {
	dir := t.TempDir()
	ini := filepath.Join(dir, "goro.ini")
	if err := os.WriteFile(ini, []byte("[ui_layout]\nwatch = true\n\n[ui]\nskills_open = true\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg := Config{ConfigPath: ini}
	if _, err := cfg.SaveUITheme(" Dark "); err != nil {
		t.Fatal(err)
	}
	if _, err := cfg.SaveUITheme("triadic"); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(ini)
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	if strings.Count(text, "theme =") != 1 || !strings.Contains(text, "theme = triadic\n") ||
		strings.Index(text, "theme =") > strings.Index(text, "[ui]") {
		t.Fatalf("theme not saved in [ui_layout] once:\n%s", data)
	}
	loaded, err := LoadConfig([]string{"--config", ini})
	if err != nil {
		t.Fatal(err)
	}
	if loaded.UILayout.Theme != "triadic" || !loaded.UILayout.Watch || !loaded.UI.SkillsOpen {
		t.Fatalf("reloaded theme=%q watch=%v skills=%v", loaded.UILayout.Theme, loaded.UILayout.Watch, loaded.UI.SkillsOpen)
	}
	if got, want := loaded.ThemesDir(), filepath.Join(dir, ThemesDirName); got != want {
		t.Fatalf("themes dir = %q, want %q", got, want)
	}
}

func TestSaveUIThemeAddsSectionToNewFile(t *testing.T) {
	ini := filepath.Join(t.TempDir(), "goro.ini")
	cfg := Config{ConfigPath: ini}
	if _, err := cfg.SaveUITheme("analogous"); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(ini)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "[ui_layout]\ntheme = analogous\n" {
		t.Fatalf("new goro.ini = %q", data)
	}
}

func TestUIThemeNameValidation(t *testing.T) {
	cfg := Config{ConfigPath: filepath.Join(t.TempDir(), "goro.ini")}
	for _, bad := range []string{"../evil", "a b", "x\ny", strings.Repeat("a", maxThemeNameLen+1)} {
		if _, err := cfg.SaveUITheme(bad); err == nil {
			t.Errorf("SaveUITheme(%q) accepted", bad)
		}
	}
	ini := filepath.Join(t.TempDir(), "goro.ini")
	if err := os.WriteFile(ini, []byte("[ui_layout]\ntheme = ../evil\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadConfig([]string{"--config", ini}); err == nil {
		t.Fatal("invalid [ui_layout] theme accepted")
	}
}
