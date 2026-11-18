package config

import (
	"os"
	"path/filepath"
	"strconv"
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
	if _, err := cfg.SaveUserSettings(UserSettings{MaxFPS: MaxAllowedFPS + 1}); err == nil {
		t.Fatal("oversized MaxFPS saved")
	}
	for _, fps := range []int{0, 60, MaxAllowedFPS} {
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

func TestMaxFPSRejectsOutOfRangeOnLoad(t *testing.T) {
	isolateUserConfig(t)
	path := filepath.Join(t.TempDir(), "goro.ini")
	for _, fps := range []int{-1, MaxAllowedFPS + 1, 2000000000} {
		if err := os.WriteFile(path, []byte("[render]\nmax_fps = "+strconv.Itoa(fps)+"\n"), 0o600); err != nil {
			t.Fatal(err)
		}
		if _, err := LoadConfig([]string{"--config", path}); err == nil {
			t.Fatalf("accepted ini max_fps=%d", fps)
		}
		if _, err := LoadConfig([]string{"-max-fps", strconv.Itoa(fps)}); err == nil {
			t.Fatalf("accepted CLI max_fps=%d", fps)
		}
	}
	if err := os.WriteFile(path, []byte("[render]\nmax_fps = 0\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	cfg, err := LoadConfig([]string{"--config", path})
	if err != nil || cfg.Render.MaxFPS != 0 {
		t.Fatalf("unlimited MaxFPS=%d error=%v", cfg.Render.MaxFPS, err)
	}
	cfg, err = LoadConfig([]string{"--config", path, "-max-fps", strconv.Itoa(MaxAllowedFPS)})
	if err != nil || cfg.Render.MaxFPS != MaxAllowedFPS {
		t.Fatalf("CLI MaxFPS=%d, want %d; error=%v", cfg.Render.MaxFPS, MaxAllowedFPS, err)
	}
}

func TestMaxFPSINIRejectsInlineComment(t *testing.T) {
	isolateUserConfig(t)
	path := filepath.Join(t.TempDir(), "goro.ini")
	if err := os.WriteFile(path, []byte("[render]\nmax_fps = 144 # 0 = unlimited\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadConfig([]string{"--config", path}); err == nil {
		t.Fatal("accepted inline comment on max_fps")
	}
	if err := os.WriteFile(path, []byte("[render]\n# 0 = unlimited\nmax_fps = 144\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	cfg, err := LoadConfig([]string{"--config", path})
	if err != nil || cfg.Render.MaxFPS != 144 {
		t.Fatalf("own-line comment MaxFPS=%d error=%v", cfg.Render.MaxFPS, err)
	}
}
