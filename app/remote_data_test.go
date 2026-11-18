package app

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/kivutar/goro/config"
	"github.com/kivutar/goro/res"
)

func TestEnableRemoteDataModes(t *testing.T) {
	empty, err := res.NewManager(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	data := config.DataConfig{Remote: config.DataRemoteAuto, RemoteURL: "http://127.0.0.1:1/", CacheDir: t.TempDir()}
	if err := enableRemoteData(data, empty); err != nil || !empty.RemoteEnabled() {
		t.Fatalf("auto without archives must stream: %v", err)
	}

	withArchive := managerWithArchive(t)
	if err := enableRemoteData(data, withArchive); err != nil || withArchive.RemoteEnabled() {
		t.Fatalf("auto with local archives must not stream: %v", err)
	}
	data.Remote = config.DataRemoteOn
	if err := enableRemoteData(data, withArchive); err != nil || !withArchive.RemoteEnabled() {
		t.Fatalf("on must stream: %v", err)
	}

	off, _ := res.NewManager(t.TempDir())
	data.Remote = config.DataRemoteOff
	if err := enableRemoteData(data, off); err != nil || off.RemoteEnabled() {
		t.Fatalf("off must not stream: %v", err)
	}
}

func managerWithArchive(t *testing.T) *res.Manager {
	t.Helper()
	src, dir := t.TempDir(), t.TempDir()
	if err := os.MkdirAll(filepath.Join(src, "data"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(src, "data", "a.txt"), []byte("a"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := res.PackGRF(filepath.Join(dir, "data.grf"), src); err != nil {
		t.Fatal(err)
	}
	m, err := res.NewManager(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(m.Archives) == 0 {
		t.Fatal("archive not opened")
	}
	t.Cleanup(func() {
		for _, a := range m.Archives {
			_ = a.Close()
		}
	})
	return m
}
