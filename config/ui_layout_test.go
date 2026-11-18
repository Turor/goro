package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestUILayoutConfig(t *testing.T) {
	dir := t.TempDir()
	ini := filepath.Join(dir, "goro.ini")
	if err := os.WriteFile(ini, []byte("[ui_layout]\nwatch = true\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, err := LoadConfig([]string{"--config", ini})
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.UILayout.Watch || cfg.UILayoutPath() != filepath.Join(dir, DefaultUILayoutFile) {
		t.Fatalf("watch=%v path=%q", cfg.UILayout.Watch, cfg.UILayoutPath())
	}
	cfg, err = LoadConfig([]string{"--config", ini, "--ui-layout", "/tmp/modern.ini", "--ui-dev=false", "--dump-ui-layout", "-"})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.UILayout.Watch || cfg.UILayoutPath() != "/tmp/modern.ini" || cfg.UILayout.DumpPath != "-" {
		t.Fatalf("flags not applied: %+v", cfg.UILayout)
	}
	if defaultConfig().UILayout.Watch {
		t.Fatal("players must not poll the layout file by default")
	}
}
