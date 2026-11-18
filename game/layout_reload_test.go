package game

import (
	"strings"
	"testing"

	"github.com/kivutar/goro/client"
	"github.com/kivutar/goro/input"
	"github.com/kivutar/goro/session"
	"github.com/kivutar/goro/ui/rotheme"
)

func TestWorldRebuildsWindowsOnlyAfterLayoutReload(t *testing.T) {
	t.Cleanup(rotheme.ResetLayout)
	ctx := client.Context{
		Session:   &session.Session{},
		Input:     input.NewState(),
		UIManager: &worldModeTestUIManager{},
		ScreenW:   1280,
		ScreenH:   720,
	}
	mode := &WorldMode{}
	mode.ui.statsWindow.OpenWindow(ctx)
	mode.rebindPersistentUI(ctx)
	built := mode.ui.layoutGen
	mode.rebuildUIAfterLayoutReload(ctx)
	if mode.ui.layoutGen != built {
		t.Fatal("rebuilt without a layout reload")
	}
	layout, _, err := rotheme.ParseLayout(strings.NewReader("[text]\nsize = 14\n"), "test")
	if err != nil {
		t.Fatal(err)
	}
	layout.Apply()
	mode.rebuildUIAfterLayoutReload(ctx)
	if mode.ui.layoutGen != rotheme.Generation() || !mode.ui.statsWindow.IsOpen() {
		t.Fatal("open windows were not rebuilt at the reloaded layout")
	}
}
