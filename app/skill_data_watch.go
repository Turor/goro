package app

import (
	"os"
	"time"

	"github.com/gogpu/gpucontext"
	"github.com/kivutar/goro/glog"
	"github.com/kivutar/goro/input"
	"github.com/kivutar/goro/res"
)

// skillDataPollInterval matches the UI layout watcher: edits to a loose skill
// table show within about a second.
const skillDataPollInterval = time.Second

// skillDataWatcher live-reloads the skill tree when a loose skill table on
// disk changes (or on F10). Like the UI layout watcher it only exists with
// --ui-dev or [ui_layout] watch = true, and runs on the game thread.
// Streamed tables are reloaded through Manager.InvalidateData instead.
type skillDataWatcher struct {
	resource *res.Manager
	nextPoll time.Time
	stamps   map[string]fileStamp
}

type fileStamp struct {
	mod  time.Time
	size int64
}

func newSkillDataWatcher(enabled bool, resource *res.Manager, now time.Time) *skillDataWatcher {
	if !enabled || resource == nil {
		return nil
	}
	w := &skillDataWatcher{resource: resource, nextPoll: now.Add(skillDataPollInterval)}
	w.stamp()
	return w
}

func (w *skillDataWatcher) stamp() bool {
	stamps := make(map[string]fileStamp)
	for _, path := range w.resource.SkillDataLocalFiles() {
		if info, err := os.Stat(path); err == nil {
			stamps[path] = fileStamp{mod: info.ModTime(), size: info.Size()}
		}
	}
	changed := len(stamps) != len(w.stamps)
	for path, stamp := range stamps {
		if old, ok := w.stamps[path]; !ok || !old.mod.Equal(stamp.mod) || old.size != stamp.size {
			changed = true
		}
	}
	w.stamps = stamps
	return changed
}

// Update polls the loose tables and handles F10; it reports a reload.
func (w *skillDataWatcher) Update(in *input.State, now time.Time) bool {
	if w == nil {
		return false
	}
	forced := in != nil && in.KeyCodeJustPressed(gpucontext.KeyF10)
	if !forced {
		if now.Before(w.nextPoll) {
			return false
		}
		w.nextPoll = now.Add(skillDataPollInterval)
		if !w.stamp() {
			return false
		}
	} else {
		w.stamp()
	}
	glog.Infof("skill tree: reloading client tables")
	return w.resource.ReloadSkillData()
}
