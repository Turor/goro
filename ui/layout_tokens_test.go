package ui

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/kivutar/goro/ui/rotheme"
)

// docs/ui-layout.default.ini is the reference players and designers copy
// from. Regenerate it with: go run . --dump-ui-layout docs/ui-layout.default.ini
func TestDefaultLayoutReferenceIsCurrent(t *testing.T) {
	want, err := os.ReadFile("../docs/ui-layout.default.ini")
	if err != nil {
		t.Fatal(err)
	}
	var got bytes.Buffer
	if err := rotheme.WriteDefaultLayout(&got); err != nil {
		t.Fatal(err)
	}
	if got.String() != string(want) {
		t.Fatal("docs/ui-layout.default.ini is stale; run: go run . --dump-ui-layout docs/ui-layout.default.ini")
	}
}

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
