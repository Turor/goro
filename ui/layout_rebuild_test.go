package ui

import (
	"testing"

	"github.com/gogpu/ui/primitives"
	"github.com/kivutar/goro/client"
	"github.com/kivutar/goro/input"
	"github.com/kivutar/goro/ui/rotheme"
)

func TestRebuildStaleWindowsRebuildsOpenWindowsInPlace(t *testing.T) {
	t.Cleanup(rotheme.ResetLayout)
	manager := NewManager()
	ctx := client.Context{Input: input.NewState(), UIManager: manager, ScreenW: 800, ScreenH: 600}
	below, above := NewWindow(100, 80), NewWindow(120, 90)
	builds := map[*Window]int{}
	for _, w := range []*Window{&below, &above} {
		w := w
		w.Open(ctx, primitives.Box())
		w.Publish(ctx)
		w.OnLayoutReload(func() {
			builds[w]++
			w.SetSize(w.width+10, w.baseHeight())
			w.SetContent(primitives.Box())
		})
	}
	if n := manager.RebuildStaleWindows(); n != 0 {
		t.Fatalf("rebuilt %d windows with no layout change", n)
	}
	rotheme.ResetLayout() // bumps the generation like a reload
	if n := manager.RebuildStaleWindows(); n != 2 {
		t.Fatalf("rebuilt %d windows, want 2", n)
	}
	if builds[&below] != 1 || builds[&above] != 1 || below.width != 110 || above.width != 130 {
		t.Fatalf("builds=%v widths=%d,%d", builds, below.width, above.width)
	}
	if manager.overlayIndex(below.published) > manager.overlayIndex(above.published) {
		t.Fatal("rebuild changed the stacking order")
	}
	if n := manager.RebuildStaleWindows(); n != 0 {
		t.Fatalf("second pass rebuilt %d windows", n)
	}

	// A closed window, or one whose manager was cleared, is never republished.
	above.Close()
	manager.Clear()
	rotheme.ResetLayout()
	if n := manager.RebuildStaleWindows(); n != 0 {
		t.Fatalf("rebuilt %d windows after clear", n)
	}
	if len(manager.overlays) != 0 {
		t.Fatal("a cleared window was republished")
	}
	below.Close()
	if _, ok := layoutRebuildWindows[&below]; ok {
		t.Fatal("closed window still tracked")
	}
}
