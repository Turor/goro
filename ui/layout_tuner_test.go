package ui

import (
	"testing"

	"github.com/kivutar/goro/ui/rotheme"
)

func TestLayoutTunerPagesWrapAndRebuildAfterEdits(t *testing.T) {
	t.Cleanup(rotheme.ResetLayout)
	tuner := NewLayoutTuner(t.TempDir() + "/ui-layout.ini")
	ctx := Context{ScreenW: 1280, ScreenH: 720}
	tuner.Toggle(ctx)
	if !tuner.IsOpen() || tuner.height <= ROWindowTitleHeight {
		t.Fatal("tuner did not open")
	}
	tuner.turn(ctx, -1)
	if want := tuner.pages(len(rotheme.Tokens())) - 1; tuner.page != want {
		t.Fatalf("page %d after turning back from the first, want the last (%d)", tuner.page, want)
	}
	before := tuner.content
	if err := rotheme.SetToken("text", "size", "14"); err != nil {
		t.Fatal(err)
	}
	tuner.Update(ctx)
	if tuner.content == before {
		t.Fatal("tuner did not rebuild after a value changed")
	}
	tuner.pending["theme.text"] = "#nothex"
	tuner.apply()
	if tuner.status == "applied" || rotheme.Default.Typography.TextSize != 14 {
		t.Fatal("an invalid typed value was applied or dropped the other edits")
	}
}
