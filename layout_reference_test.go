package main

import (
	"bytes"
	"os"
	"testing"

	"github.com/kivutar/goro/ui/rotheme"
)

// docs/ui-layout.default.ini is the reference players and designers copy
// from. Regenerate it with: go run . --dump-ui-layout docs/ui-layout.default.ini
//
// The test lives in package main so every package that registers layout
// tokens (ui, ui/rotheme, game/login, game/map) is linked, exactly as in the
// client that writes the dump.
func TestDefaultLayoutReferenceIsCurrent(t *testing.T) {
	want, err := os.ReadFile("docs/ui-layout.default.ini")
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
