package app

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/kivutar/goro/config"
	"github.com/kivutar/goro/ui/rotheme"
)

func TestUILayoutWatcherAppliesRejectsAndResets(t *testing.T) {
	t.Cleanup(rotheme.ResetLayout)
	if newUILayoutWatcher(config.Config{}, time.Now()) != nil {
		t.Fatal("players must not get a watcher")
	}
	path := filepath.Join(t.TempDir(), "ui-layout.ini")
	cfg := config.Config{UILayout: config.UILayoutConfig{Path: path, Watch: true}}
	now := time.Now()
	w := newUILayoutWatcher(cfg, now)
	builtIn := rotheme.Default.Typography.TextSize
	write := func(src string, age time.Duration) {
		t.Helper()
		if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
			t.Fatal(err)
		}
		stamp := time.Unix(1_700_000_000, 0).Add(age)
		if err := os.Chtimes(path, stamp, stamp); err != nil {
			t.Fatal(err)
		}
	}
	tick := func() bool {
		now = now.Add(uiLayoutPollInterval)
		return w.Update(nil, now)
	}

	write("[text]\nsize = 15\n", time.Second)
	if !tick() || rotheme.Default.Typography.TextSize != 15 {
		t.Fatalf("edit not applied, size %v", rotheme.Default.Typography.TextSize)
	}
	if tick() {
		t.Fatal("an unchanged file must not reapply")
	}
	write("[text]\nsize = 99\n", 2*time.Second)
	if tick() || rotheme.Default.Typography.TextSize != 15 {
		t.Fatal("an invalid edit must keep the last good values")
	}
	if !w.toast.Visible(now) {
		t.Fatal("an invalid edit must show an error toast")
	}
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	if !tick() || rotheme.Default.Typography.TextSize != builtIn {
		t.Fatal("removing the file must restore the built-in layout")
	}
}
