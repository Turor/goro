package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDataRemoteDefaultsToTuroranAuto(t *testing.T) {
	cfg, err := LoadConfig([]string{"--data-dir", t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Data.Remote != DataRemoteAuto || cfg.Data.RemoteURL != DefaultDataRemoteURL {
		t.Fatalf("defaults: %+v", cfg.Data)
	}
}

func TestDataRemoteINIAndFlags(t *testing.T) {
	dir := t.TempDir()
	ini := "[data]\nremote = ON\nremote_url = http://127.0.0.1:3339/\ncache_dir = cache\n"
	if err := os.WriteFile(filepath.Join(dir, "goro.ini"), []byte(ini), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, err := LoadConfig([]string{"--data-dir", dir})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Data.Remote != DataRemoteOn || cfg.Data.RemoteURL != "http://127.0.0.1:3339/" || cfg.Data.CacheDir != "cache" {
		t.Fatalf("ini: %+v", cfg.Data)
	}
	cfg, err = LoadConfig([]string{"--data-dir", dir, "--data-remote", "off", "--data-cache", "/tmp/x"})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Data.Remote != DataRemoteOff || cfg.Data.CacheDir != "/tmp/x" {
		t.Fatalf("flags: %+v", cfg.Data)
	}
}

func TestDataRemoteRejectsBadValues(t *testing.T) {
	for _, args := range [][]string{
		{"--data-remote", "maybe"},
		{"--data-url", "ftp://example.com/"},
		{"--data-url", "not a url"},
	} {
		if _, err := LoadConfig(append([]string{"--data-dir", t.TempDir()}, args...)); err == nil {
			t.Errorf("%v accepted", args)
		}
	}
}

func TestDataRemoteCacheDirIsPerHost(t *testing.T) {
	t.Setenv("XDG_CACHE_HOME", t.TempDir())
	dir, err := DataConfig{RemoteURL: "https://ro.turoran.com/client/"}.RemoteCacheDir()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(filepath.ToSlash(dir), "goro/remote/ro.turoran.com") {
		t.Fatalf("cache dir %s", dir)
	}
}
