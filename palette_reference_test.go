package main

import (
	"bytes"
	"fmt"
	"go/format"
	"os"
	"sort"
	"strings"
	"testing"

	"github.com/kivutar/goro/ui/rotheme"
)

const paletteRefsFile = "ui/rotheme/palette_refs.go"

// Every color setting defaults to a [palette] reference, and with the default
// palette each reference must give exactly the built-in color, so the palette
// changes nothing until a player edits it.
//
// After adding a color setting, add it to ui/rotheme/palette_refs.go with the
// palette color it belongs to ("@panel") and run
//
//	GORO_UPDATE_PALETTE_REFS=1 go test -run TestPaletteDefaults .
//
// to compute the exact modifiers. A setting missing from the file gets the
// nearest palette color. The test lives in package main so every package that
// registers color settings is linked.
func TestPaletteDefaultsReproduceBuiltInColors(t *testing.T) {
	if os.Getenv("GORO_UPDATE_PALETTE_REFS") != "" {
		writePaletteRefs(t)
		t.Skip("regenerated " + paletteRefsFile + "; rerun without GORO_UPDATE_PALETTE_REFS")
	}
	problems := rotheme.CheckPaletteDefaults()
	for _, id := range rotheme.UnusedPaletteRefs() {
		problems = append(problems, id+" is in paletteRefs but is not a color setting")
	}
	if len(problems) > 0 {
		t.Fatalf("%d palette reference problems (regenerate with GORO_UPDATE_PALETTE_REFS=1 go test -run TestPaletteDefaults .):\n%s",
			len(problems), strings.Join(problems, "\n"))
	}
	if n := len(rotheme.ColorTokenDefaults()); n < 100 {
		t.Fatalf("only %d color settings registered; is every token package linked?", n)
	}
}

// Changing a palette color must reach every setting that refers to it, and
// nothing else.
func TestPaletteChangeReachesEveryReference(t *testing.T) {
	t.Cleanup(rotheme.ResetLayout)
	rotheme.ResetLayout()
	baseline := colorsByID()
	refs := map[string]string{}
	for _, ref := range rotheme.ColorTokenDefaults() {
		refs[ref.Section+"."+ref.Key] = rotheme.PaletteBase(ref.Section, ref.Key)
	}
	for _, key := range rotheme.PaletteKeys() {
		users := 0
		for _, base := range refs {
			if base == key {
				users++
			}
		}
		// Two unrelated colors: a reference that clamps (lighten to white)
		// for one still moves for the other.
		changed := map[string]bool{}
		for _, probe := range []string{"#3c7a52", "#d2a04e"} {
			layout, warnings, err := rotheme.ParseLayout(strings.NewReader("[palette]\n"+key+" = "+probe+"\n"), "probe")
			if err != nil || len(warnings) != 0 {
				t.Fatalf("palette %s: %v %v", key, err, warnings)
			}
			layout.Apply()
			for id, c := range colorsByID() {
				if c != baseline[id] {
					changed[id] = true
				}
			}
			rotheme.ResetLayout()
		}
		for id, base := range refs {
			if base == key && !changed[id] {
				t.Errorf("changing [palette] %s did not change %s", key, id)
			}
			if base != key && changed[id] {
				t.Errorf("changing [palette] %s changed %s, which follows @%s", key, id, base)
			}
		}
		t.Logf("@%s: %d settings", key, users)
	}
	if after := colorsByID(); fmt.Sprint(after) != fmt.Sprint(baseline) {
		t.Fatal("resetting the layout did not restore every color")
	}
}

func colorsByID() map[string]string {
	out := map[string]string{}
	for _, info := range rotheme.Tokens() {
		if info.Color != "" && info.Section != "palette" {
			out[info.Section+"."+info.Key] = info.Color
		}
	}
	return out
}

func writePaletteRefs(t *testing.T) {
	t.Helper()
	var entries []string
	for _, ref := range rotheme.ColorTokenDefaults() {
		id := ref.Section + "." + ref.Key
		base := ""
		if ref.Ref != "" {
			base = strings.TrimPrefix(strings.Fields(ref.Ref)[0], "@")
		}
		expr, ok := rotheme.SuggestPaletteRef(base, ref.Color)
		if !ok {
			base = rotheme.NearestPaletteColor(ref.Color)
			expr, ok = rotheme.SuggestPaletteRef(base, ref.Color)
		}
		if !ok {
			t.Fatalf("%s: no exact palette reference for its built-in color", id)
		}
		entries = append(entries, fmt.Sprintf("\t%q: %q,\n", id, expr))
	}
	sort.Strings(entries)
	var b bytes.Buffer
	b.WriteString(`package rotheme

// paletteRefs is the default [palette] reference of every color setting,
// keyed by section.key. With the default palette each one resolves to exactly
// the built-in color (palette_reference_test.go in package main checks this).
//
// To move a setting to another palette color, change its @name and run
//
//	GORO_UPDATE_PALETTE_REFS=1 go test -run TestPaletteDefaults .
//
// which recomputes the modifiers; new settings are added the same way.
var paletteRefs = map[string]string{
`)
	for _, e := range entries {
		b.WriteString(e)
	}
	b.WriteString("}\n")
	src, err := format.Source(b.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(paletteRefsFile, src, 0o644); err != nil {
		t.Fatal(err)
	}
}

// The example palette in docs must stay a valid layout file.
func TestDarkPaletteExampleLoads(t *testing.T) {
	t.Cleanup(rotheme.ResetLayout)
	applied, warnings, err := rotheme.LoadLayoutFile("docs/ui-palette-dark.example.ini")
	if err != nil || !applied || len(warnings) != 0 {
		t.Fatalf("applied=%v warnings=%v err=%v", applied, warnings, err)
	}
	for _, info := range rotheme.Tokens() {
		if info.Section == "theme" && info.Key == "window_body" && info.Color != "#1c1f26" {
			t.Fatalf("window_body = %s, want the example background", info.Color)
		}
	}
}
