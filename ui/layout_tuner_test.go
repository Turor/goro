package ui

import (
	"strings"
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

func TestLayoutTunerFiltersBySectionAndSearch(t *testing.T) {
	t.Cleanup(rotheme.ResetLayout)
	tuner := NewLayoutTuner(t.TempDir() + "/ui-layout.ini")
	ctx := Context{ScreenW: 1280, ScreenH: 720}
	tuner.Toggle(ctx)
	all := len(tuner.filtered())
	if all != len(rotheme.Tokens()) {
		t.Fatalf("all sections shows %d of %d keys", all, len(rotheme.Tokens()))
	}
	sections := tunerSections()
	for i, name := range sections {
		if name == "tuner" {
			tuner.chooseSection(ctx, i-tuner.section)
		}
	}
	for _, info := range tuner.filtered() {
		if info.Section != "tuner" {
			t.Fatalf("section filter let %s.%s through", info.Section, info.Key)
		}
	}
	tuner.chooseSection(ctx, -tuner.section)
	tuner.search(ctx, "KEY_WIDTH")
	got := tuner.filtered()
	if len(got) == 0 || len(got) >= all {
		t.Fatalf("search matched %d of %d keys", len(got), all)
	}
	for _, info := range got {
		if !strings.Contains(info.Section+"."+info.Key, "key_width") && !strings.Contains(strings.ToLower(info.Doc), "key_width") {
			t.Fatalf("search let %s.%s through", info.Section, info.Key)
		}
	}
	tuner.chooseSection(ctx, -1) // wraps to the last section
	if tuner.section != len(sections)-1 {
		t.Fatalf("section %d after wrapping back, want %d", tuner.section, len(sections)-1)
	}
}
