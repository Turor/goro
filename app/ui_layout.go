package app

import (
	"fmt"
	"image/color"
	"io"
	"os"
	"time"
	"unicode/utf8"

	"github.com/gogpu/gpucontext"
	"github.com/kivutar/goro/input"
	gameui "github.com/kivutar/goro/ui"

	"github.com/kivutar/goro/config"
	"github.com/kivutar/goro/glog"
	"github.com/kivutar/goro/render"
	"github.com/kivutar/goro/ui/rotheme"
)

// loadUILayout applies the optional ui-layout.ini override once at startup.
// A missing file costs one failed open and keeps the built-in layout; an
// invalid file is logged and also keeps the built-in layout. It never fails
// startup.
func loadUILayout(cfg config.Config) {
	path := cfg.UILayoutPath()
	applied, warnings, err := rotheme.LoadLayoutFile(path)
	// ParseLayout already caps warnings at MaxLayoutWarnings plus a total.
	for _, warning := range warnings {
		glog.Warnf("ui layout: %s", warning)
	}
	switch {
	case err != nil:
		glog.Errorf("ui layout rejected, using built-in layout: %v", err)
	case applied:
		glog.Infof("ui layout loaded path=%s", path)
		if err := loadUIFont(path); err != nil {
			glog.Errorf("ui layout: %v (using the built-in font)", err)
		}
	}
}

// DumpUILayout writes every layout key with its built-in default to path, or
// to stdout when path is "-".
func DumpUILayout(path string) error {
	if path == "-" {
		return rotheme.WriteDefaultLayout(os.Stdout)
	}
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("dump ui layout: %w", err)
	}
	if err := writeAndClose(file, rotheme.WriteDefaultLayout); err != nil {
		return fmt.Errorf("dump ui layout: %w", err)
	}
	return nil
}

func writeAndClose(file *os.File, write func(io.Writer) error) error {
	if err := write(file); err != nil {
		file.Close()
		return err
	}
	return file.Close()
}

// uiLayoutPollInterval keeps edits visible within about a second while
// costing one stat call per interval, and only with --ui-dev.
const uiLayoutPollInterval = 500 * time.Millisecond

// uiLayoutWatcher applies ui-layout.ini edits while the game runs. It only
// exists with --ui-dev or [ui_layout] watch = true; players never poll. It
// runs on the game thread, so token writes never race the UI.
type uiLayoutWatcher struct {
	path     string
	nextPoll time.Time
	modTime  time.Time
	size     int64
	exists   bool
	toast    gameui.Announcement
}

func newUILayoutWatcher(cfg config.Config, now time.Time) *uiLayoutWatcher {
	if !cfg.UILayout.Watch {
		return nil
	}
	w := &uiLayoutWatcher{path: cfg.UILayoutPath()}
	w.stamp()
	w.nextPoll = now.Add(uiLayoutPollInterval)
	glog.Infof("ui layout watch on path=%s (F10 reloads)", w.path)
	return w
}

func (w *uiLayoutWatcher) stamp() bool {
	info, err := os.Stat(w.path)
	exists := err == nil
	var mod time.Time
	var size int64
	if exists {
		mod, size = info.ModTime(), info.Size()
	}
	changed := exists != w.exists || !mod.Equal(w.modTime) || size != w.size
	w.exists, w.modTime, w.size = exists, mod, size
	return changed
}

// Update polls the file and handles F10. It reports whether new values were
// applied, so the caller can refresh open windows.
func (w *uiLayoutWatcher) Update(in *input.State, now time.Time) bool {
	if w == nil {
		return false
	}
	forced := in != nil && in.KeyCodeJustPressed(gpucontext.KeyF10)
	if !forced {
		if now.Before(w.nextPoll) {
			return false
		}
		w.nextPoll = now.Add(uiLayoutPollInterval)
		if !w.stamp() {
			return false
		}
	} else {
		w.stamp()
	}
	return w.reload(now)
}

func (w *uiLayoutWatcher) reload(now time.Time) bool {
	if !w.exists {
		rotheme.ResetLayout()
		glog.Infof("ui layout file gone, built-in layout restored path=%s", w.path)
		w.show("UI layout: file removed, built-in layout restored", uiLayoutOKColor, now, 4*time.Second)
		return true
	}
	applied, warnings, err := rotheme.LoadLayoutFile(w.path)
	for _, warning := range warnings {
		glog.Warnf("ui layout: %s", warning)
	}
	if err != nil {
		glog.Errorf("ui layout rejected, keeping previous values: %v", err)
		w.show("UI layout NOT applied: "+bannerText(err.Error()), uiLayoutErrorColor, now, 15*time.Second)
		return false
	}
	if !applied {
		return false
	}
	glog.Infof("ui layout reloaded path=%s", w.path)
	if err := loadUIFont(w.path); err != nil {
		glog.Errorf("ui layout: %v (using the built-in font)", err)
		warnings = append(warnings, err.Error())
	}
	msg := "UI layout reloaded" + warningSummary(warnings)
	w.show(msg, uiLayoutOKColor, now, 4*time.Second)
	return true
}

// updateUITuner toggles the --ui-dev tuning panel with F11 and runs it.
func (g *Game) updateUITuner() {
	if g.uiTuner == nil || g.cfg.Headless {
		return
	}
	ctx := g.modeContext()
	if g.input.KeyCodeJustPressed(gpucontext.KeyF11) {
		g.uiTuner.Toggle(ctx)
	}
	g.uiTuner.Update(ctx)
}

var (
	uiLayoutOKColor    = color.RGBA{R: 160, G: 255, B: 160, A: 255}
	uiLayoutErrorColor = color.RGBA{R: 255, G: 120, B: 120, A: 255}
)

func (w *uiLayoutWatcher) show(text string, c color.RGBA, now time.Time, life time.Duration) {
	w.toast.ShowFor(text, c, gameui.AnnouncementStyle{Y: 12}, now, life)
}

func (w *uiLayoutWatcher) Draw(screen *render.Frame, now time.Time) {
	if w == nil {
		return
	}
	w.toast.Draw(screen, now)
}

// uiLayoutBannerMax keeps banner text to about two lines; the log has it all.
const uiLayoutBannerMax = 160

func bannerText(s string) string {
	if len(s) <= uiLayoutBannerMax {
		return s
	}
	cut := uiLayoutBannerMax
	for cut > 0 && !utf8.RuneStart(s[cut]) {
		cut--
	}
	return s[:cut] + "…"
}

// warningSummary shows one warning, or only a count when there are several.
func warningSummary(warnings []string) string {
	switch {
	case len(warnings) == 0:
		return ""
	case len(warnings) == 1:
		return " (" + bannerText(warnings[0]) + ")"
	case len(warnings) > rotheme.MaxLayoutWarnings:
		return fmt.Sprintf(" (over %d warnings, see log)", rotheme.MaxLayoutWarnings)
	default:
		return fmt.Sprintf(" (%d warnings, see log)", len(warnings))
	}
}
