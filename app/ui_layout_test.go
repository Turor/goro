package app

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
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

func TestUILayoutBannerIsBounded(t *testing.T) {
	many := make([]string, rotheme.MaxLayoutWarnings+1)
	for i := range many {
		many[i] = fmt.Sprintf("f:%d: unknown key [x] k (ignored)", i)
	}
	if got := warningSummary(many); got != fmt.Sprintf(" (over %d warnings, see log)", rotheme.MaxLayoutWarnings) {
		t.Fatalf("summary %q", got)
	}
	if got := warningSummary(many[:3]); got != " (3 warnings, see log)" {
		t.Fatalf("summary %q", got)
	}
	if got := bannerText(strings.Repeat("é", 500)); len(got) > uiLayoutBannerMax+len("…") || !strings.HasSuffix(got, "…") {
		t.Fatalf("banner not truncated: %d bytes", len(got))
	}
}
