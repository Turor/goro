package app

import (
	"os"
	"time"

	"github.com/gogpu/gpucontext"
	"github.com/kivutar/goro/glog"
	"github.com/kivutar/goro/input"
	"github.com/kivutar/goro/res"
)

const effectDataPollInterval = time.Second

// effectDataWatcher live-reloads effect tables when a loose overlay on disk
// changes (or on F10). Streamed tables reload through Manager.InvalidateData.
type effectDataWatcher struct {
	resource *res.Manager
	nextPoll time.Time
	stamps   map[string]fileStamp
}

func newEffectDataWatcher(enabled bool, resource *res.Manager, now time.Time) *effectDataWatcher {
	if !enabled || resource == nil {
		return nil
	}
	w := &effectDataWatcher{resource: resource, nextPoll: now.Add(effectDataPollInterval)}
	w.stamp()
	return w
}

func (w *effectDataWatcher) stamp() bool {
	stamps := make(map[string]fileStamp)
	for _, path := range w.resource.EffectDataLocalFiles() {
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

func (w *effectDataWatcher) Update(in *input.State, now time.Time) bool {
	if w == nil {
		return false
	}
	forced := in != nil && in.KeyCodeJustPressed(gpucontext.KeyF10)
	if !forced {
		if now.Before(w.nextPoll) {
			return false
		}
		w.nextPoll = now.Add(effectDataPollInterval)
		if !w.stamp() {
			return false
		}
	} else {
		w.stamp()
	}
	glog.Infof("effects: reloading client tables")
	return w.resource.ReloadEffectData()
}
