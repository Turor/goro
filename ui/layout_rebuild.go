package ui

import (
	"sort"

	"github.com/gogpu/ui/widget"
	"github.com/kivutar/goro/ui/rotheme"
)

// Live layout rebuild. A window that copies layout tokens when it builds its
// widget tree (sizes, paddings, row heights, colors) registers a rebuild
// function with OnLayoutReload. After ui-layout.ini or the F11 tuner changes
// any token, Manager.RebuildStaleWindows calls it once for every window that
// is still on screen, in stacking order, so login, server and character
// select screens and transient pop-ups (shops, trades, prompts) follow the
// file without being reopened. Players never change the generation, so in a
// release build this costs one atomic load per frame.

// layoutRebuildWindows holds every published window with a rebuild function.
var layoutRebuildWindows = map[*Window]struct{}{}

// OnLayoutReload sets how the window rebuilds its content (and size) after a
// layout reload. fn runs on the UI thread with the window still open; it
// normally calls SetSize and SetContent with a freshly built tree. The window
// is republished in its stack slot afterwards. Call it again (for example on
// every Open) to replace fn; pass nil to stop rebuilding.
func (w *Window) OnLayoutReload(fn func()) {
	w.layoutRebuild = fn
	w.layoutGen = rotheme.Generation()
	if fn == nil {
		delete(layoutRebuildWindows, w)
		return
	}
	if w.published != nil {
		layoutRebuildWindows[w] = struct{}{}
	}
}

// trackLayoutRebuild keeps the registry in step with Publish and Unpublish.
func (w *Window) trackLayoutRebuild() {
	if w.layoutRebuild != nil && w.open && w.published != nil {
		layoutRebuildWindows[w] = struct{}{}
		return
	}
	delete(layoutRebuildWindows, w)
}

// HasOverlay reports whether root is published on this manager.
func (m *Manager) HasOverlay(root widget.Widget) bool {
	if m == nil || root == nil {
		return false
	}
	for _, child := range m.overlays {
		if child == root {
			return true
		}
	}
	return false
}

func (m *Manager) overlayIndex(root widget.Widget) int {
	for i, child := range m.overlays {
		if child == root {
			return i
		}
	}
	return -1
}

// RebuildStaleWindows rebuilds every window on this manager that registered
// OnLayoutReload and was built before the current layout generation. It
// returns how many windows it rebuilt. Windows no longer on this manager (a
// mode change cleared the overlays) are forgotten, never republished.
func (m *Manager) RebuildStaleWindows() int {
	if m == nil || len(layoutRebuildWindows) == 0 {
		return 0
	}
	gen := rotheme.Generation()
	type stale struct {
		w     *Window
		index int
	}
	var list []stale
	for w := range layoutRebuildWindows {
		if w.layoutGen == gen {
			continue
		}
		index := m.overlayIndex(w.published)
		if index < 0 {
			if !w.open || w.published == nil {
				delete(layoutRebuildWindows, w)
			}
			continue
		}
		list = append(list, stale{w: w, index: index})
	}
	sort.Slice(list, func(i, j int) bool { return list[i].index < list[j].index })
	for _, entry := range list {
		w := entry.w
		w.layoutGen = gen
		if !w.open || w.layoutRebuild == nil {
			continue
		}
		w.layoutRebuild()
		if w.open && w.content != nil {
			w.Publish(w.ctx)
		}
	}
	return len(list)
}
