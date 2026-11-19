package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDataRemoteDefaultsToAutoWithNoServer(t *testing.T) {
	cfg, err := LoadConfig([]string{"--data-dir", t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Data.Remote != DataRemoteAuto || cfg.Data.RemoteURL != "" || cfg.Data.Streaming() {
		t.Fatalf("defaults must not stream without a configured server: %+v", cfg.Data)
	}
}

func TestDataRemoteBuildDefault(t *testing.T) {
	old := DefaultDataRemoteURL
	DefaultDataRemoteURL = "https://build.example.invalid/client/"
	t.Cleanup(func() { DefaultDataRemoteURL = old })
	cfg, err := LoadConfig([]string{"--data-dir", t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Data.RemoteURL != DefaultDataRemoteURL || !cfg.Data.Streaming() {
		t.Fatalf("link-time default ignored: %+v", cfg.Data)
	}
}

func TestDataServerFile(t *testing.T) {
	dir := t.TempDir()
	ini := "; operator file, never committed\n[data]\nremote_url = https://assets.example.invalid/client/\n"
	if err := os.WriteFile(filepath.Join(dir, DataServerFile), []byte(ini), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, err := LoadConfig([]string{"--data-dir", dir})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Data.RemoteURL != "https://assets.example.invalid/client/" || !cfg.Data.Streaming() {
		t.Fatalf("data-server.ini: %+v", cfg.Data)
	}
	// goro.ini and flags still win over data-server.ini.
	if err := os.WriteFile(filepath.Join(dir, "goro.ini"), []byte("[data]\nremote_url = https://other.example.invalid/\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, err = LoadConfig([]string{"--data-dir", dir})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Data.RemoteURL != "https://other.example.invalid/" {
		t.Fatalf("goro.ini must override data-server.ini: %+v", cfg.Data)
	}
	cfg, err = LoadConfig([]string{"--data-dir", dir, "--data-remote", "off"})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Data.Streaming() {
		t.Fatalf("--data-remote off must win: %+v", cfg.Data)
	}
}

func TestDataServerFileOnlyData(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, DataServerFile), []byte("[login]\nhost = x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadConfig([]string{"--data-dir", dir}); err == nil || !strings.Contains(err.Error(), "only the [data] section") {
		t.Fatalf("non-data section accepted: %v", err)
	}
}

func TestDataRemoteOnNeedsURL(t *testing.T) {
	if _, err := LoadConfig([]string{"--data-dir", t.TempDir(), "--data-remote", "on"}); err == nil {
		t.Fatal("on without a URL accepted")
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

func TestDataManifestMode(t *testing.T) {
	dir := t.TempDir()
	cfg, err := LoadConfig([]string{"--data-dir", dir})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Data.ManifestMode != DataManifestReport {
		t.Fatalf("default mode %q", cfg.Data.ManifestMode)
	}
	cfg, err = LoadConfig([]string{"--data-dir", dir, "--data-manifest-mode", "enforce"})
	if err != nil || cfg.Data.ManifestMode != DataManifestEnforce {
		t.Fatalf("flag: %+v %v", cfg.Data, err)
	}
	ini := "[data]\nmanifest_mode = ENFORCE\n"
	if err := os.WriteFile(filepath.Join(dir, "goro.ini"), []byte(ini), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, err = LoadConfig([]string{"--data-dir", dir})
	if err != nil || cfg.Data.ManifestMode != DataManifestEnforce {
		t.Fatalf("ini: %+v %v", cfg.Data, err)
	}
	if _, err := LoadConfig([]string{"--data-dir", t.TempDir(), "--data-manifest-mode", "off"}); err == nil {
		t.Fatal("bad manifest mode accepted")
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
	dir, err := DataConfig{RemoteURL: "https://assets.example.invalid:8443/client/"}.RemoteCacheDir()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(filepath.ToSlash(dir), "goro/remote/assets.example.invalid_8443") {
		t.Fatalf("cache dir %s", dir)
	}
}
