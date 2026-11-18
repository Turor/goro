package config

import (
	"path/filepath"
	"testing"
)

func TestMaxFPSDefaultSavedAndCLIOverride(t *testing.T) {
	isolateUserConfig(t)
	cfg, err := LoadConfig(nil)
	if err != nil || cfg.Render.MaxFPS != DefaultMaxFPS {
		t.Fatalf("MaxFPS should default to %d: got %d err=%v", DefaultMaxFPS, cfg.Render.MaxFPS, err)
	}
	cfg.ConfigPath = filepath.Join(t.TempDir(), "goro.ini")
	if _, err := cfg.SaveUserSettings(UserSettings{MaxFPS: -1}); err == nil {
		t.Fatal("negative MaxFPS saved")
	}
	for _, fps := range []int{0, 60} {
		if _, err := cfg.SaveUserSettings(UserSettings{MaxFPS: fps}); err != nil {
			t.Fatal(err)
		}
		saved, err := LoadConfig([]string{"--config", cfg.ConfigPath})
		if err != nil || saved.Render.MaxFPS != fps {
			t.Fatalf("saved MaxFPS=%d, want %d; error=%v", saved.Render.MaxFPS, fps, err)
		}
		overridden, err := LoadConfig([]string{"--config", cfg.ConfigPath, "-max-fps=240"})
		if err != nil || overridden.Render.MaxFPS != 240 {
			t.Fatalf("CLI MaxFPS=%d, want 240; error=%v", overridden.Render.MaxFPS, err)
		}
	}
}
