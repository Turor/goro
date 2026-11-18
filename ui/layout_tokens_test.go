package ui

import (
	"strings"
	"testing"

	"github.com/kivutar/goro/ui/rotheme"
)

func TestInventoryLayoutTokensRecomputeDerivedSizes(t *testing.T) {
	t.Cleanup(rotheme.ResetLayout)
	defaultHeight := inventoryBagHeight
	layout, _, err := rotheme.ParseLayout(strings.NewReader("[inventory]\ncell_size = 40\ncols = 6\nrows = 3\n"), "test")
	if err != nil {
		t.Fatal(err)
	}
	layout.Apply()
	if inventoryBagGridW != 240 || inventoryBagViewH != 120 || inventoryBagHeight != ROWindowTitleHeight+120 {
		t.Fatalf("grid %d view %d height %d", inventoryBagGridW, inventoryBagViewH, inventoryBagHeight)
	}
	rotheme.ResetLayout()
	if inventoryBagHeight != defaultHeight || inventoryBagCols != 8 || inventoryBagCell != 32 {
		t.Fatal("reset did not restore the built-in inventory metrics")
	}
}

func TestTitleFooterAndStatusTokensResizeWindows(t *testing.T) {
	t.Cleanup(rotheme.ResetLayout)
	layout, _, err := rotheme.ParseLayout(strings.NewReader("[window]\ntitle_height = 36\nfooter_height = 50\n[status]\nrow_height = 20\n"), "test")
	if err != nil {
		t.Fatal(err)
	}
	layout.Apply()
	w := NewWindow(300, 200)
	w.SetContent(Win(Title("t"), Size(300, 200), Footer()))
	if w.height != 216 || w.titleHeight != 36 {
		t.Fatalf("window height %d title %d, want 216 and 36", w.height, w.titleHeight)
	}
	w.SetSize(300, 180)
	if w.height != 196 {
		t.Fatalf("SetSize height %d, want the base plus the 16px token delta", w.height)
	}
	if statsWindowHeight != 161+6*2 {
		t.Fatalf("status window height %d, want it to follow row_height", statsWindowHeight)
	}
	rotheme.ResetLayout()
	if statsWindowHeight != 161 || windowTitleHeight != ROWindowTitleHeight {
		t.Fatal("reset did not restore the built-in window metrics")
	}
}
