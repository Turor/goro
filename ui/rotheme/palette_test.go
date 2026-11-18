package rotheme

import (
	"bytes"
	"image/color"
	"math/rand"
	"strings"
	"testing"

	"github.com/gogpu/ui/widget"
)

func TestColorExprParseAndCanonicalForm(t *testing.T) {
	cases := map[string]string{
		"@panel":                             "@panel",
		"@ACCENT   darken(15%)":              "@accent darken(15%)",
		"@bad alpha(0.8)":                    "@bad alpha(0.8)",
		"@bad alpha(80%)":                    "@bad alpha(0.8)",
		"@text lighten(5) spin(-30deg)":      "@text lighten(5%) spin(-30)",
		"@sp saturate(10%) desaturate(2.5%)": "@sp saturate(10%) desaturate(2.5%)",
	}
	for raw, want := range cases {
		e, err := parseColorExpr(raw)
		if err != nil {
			t.Fatalf("%q: %v", raw, err)
		}
		if got := e.String(); got != want {
			t.Errorf("%q canonical = %q, want %q", raw, got, want)
		}
	}
	for _, bad := range []string{
		"panel", "@", "@nope", "@panel darken", "@panel darken(x)", "@panel darken(101%)",
		"@panel blur(2)", "@panel alpha(2)", "@panel spin(400)", "@panel spin(10%)",
		"@panel " + strings.Repeat("lighten(1%) ", 9),
	} {
		if _, err := parseColorExpr(bad); err == nil {
			t.Errorf("%q parsed, want an error", bad)
		}
	}
}

func TestColorModifiers(t *testing.T) {
	gray := color.RGBA{R: 128, G: 128, B: 128, A: 255}
	if got := applyColorMods(gray, []colorMod{{op: "lighten", v: 100}}); got != (color.RGBA{R: 255, G: 255, B: 255, A: 255}) {
		t.Fatalf("lighten(100%%) = %v", got)
	}
	if got := applyColorMods(gray, []colorMod{{op: "darken", v: 100}}); got != (color.RGBA{A: 255}) {
		t.Fatalf("darken(100%%) = %v", got)
	}
	odd := color.RGBA{R: 13, G: 200, B: 77, A: 255}
	if got := applyColorMods(odd, []colorMod{{op: "alpha", v: 0.5}}); got != (color.RGBA{R: 13, G: 200, B: 77, A: 128}) {
		t.Fatalf("alpha only must keep the base bytes, got %v", got)
	}
	red := color.RGBA{R: 255, A: 255}
	if got := applyColorMods(red, []colorMod{{op: "spin", v: 120}}); got != (color.RGBA{G: 255, A: 255}) {
		t.Fatalf("spin(120) of red = %v", got)
	}
	if got := applyColorMods(red, []colorMod{{op: "desaturate", v: 100}}); got.R != got.G || got.G != got.B {
		t.Fatalf("desaturate(100%%) is not gray: %v", got)
	}
}

// Every 8-bit color can be reached exactly from every palette color, so the
// default references never need an explicit color.
func TestSuggestPaletteRefIsExact(t *testing.T) {
	rng := rand.New(rand.NewSource(7))
	for i := 0; i < 400; i++ {
		target := color.RGBA{R: uint8(rng.Intn(256)), G: uint8(rng.Intn(256)), B: uint8(rng.Intn(256)), A: uint8(rng.Intn(256))}
		for _, key := range PaletteKeys() {
			ref, ok := SuggestPaletteRef(key, target)
			if !ok {
				t.Fatalf("no exact reference from @%s to %v", key, target)
			}
			e, err := parseColorExpr(ref)
			if err != nil {
				t.Fatalf("%s: %v", ref, err)
			}
			if got := e.resolveWith(defaultPaletteColor); got != target {
				t.Fatalf("%s resolves to %v, want %v", ref, got, target)
			}
		}
	}
}

func TestPaletteReferencesInLayoutFiles(t *testing.T) {
	resetLayoutAfter(t)
	apply(t, "[palette]\nbackground = #102030\n")
	if got := Default.Colors.WindowBody; got != widget.RGBA8(0x10, 0x20, 0x30, 0xff) {
		t.Fatalf("window_body did not follow @background: %+v", got)
	}
	apply(t, "[palette]\nbackground = #102030\n[theme]\nwindow_body = #ff0000\n")
	if got := Default.Colors.WindowBody; got != widget.RGBA8(0xff, 0, 0, 0xff) {
		t.Fatalf("an explicit color must win over the palette: %+v", got)
	}
	apply(t, "[theme]\nwindow_body = @accent alpha(0.5)\n")
	if got := Default.Colors.WindowBody; got != widget.RGBA8(118, 160, 206, 128) {
		t.Fatalf("own reference: %+v", got)
	}
	ResetLayout()
	if Default.Colors.WindowBody != tokenIndex["theme.window_body"].def.c {
		t.Fatal("reset did not restore the built-in color")
	}
	for name, src := range map[string]string{
		"palette reference in palette": "[palette]\npanel = @accent\n",
		"unknown palette color":        "[theme]\nwindow_body = @nope\n",
		"bad modifier":                 "[theme]\nwindow_body = @panel glow(3)\n",
	} {
		if _, _, err := ParseLayout(strings.NewReader(src), "ui-layout.ini"); err == nil || !strings.Contains(err.Error(), "ui-layout.ini:2:") {
			t.Errorf("%s: want a line-numbered error, got %v", name, err)
		}
	}
}

func TestTunerKeepsPaletteReferences(t *testing.T) {
	resetLayoutAfter(t)
	if err := SetToken("palette", "background", "#102030"); err != nil {
		t.Fatal(err)
	}
	var body TokenInfo
	for _, info := range Tokens() {
		if info.Section == "theme" && info.Key == "window_body" {
			body = info
		}
	}
	if body.Value != "@background" || body.Color != "#102030" || body.Value != body.Default {
		t.Fatalf("window_body after a palette change: value %q color %q default %q", body.Value, body.Color, body.Default)
	}
	var out bytes.Buffer
	if err := WriteOverrides(&out); err != nil {
		t.Fatal(err)
	}
	if got := out.String(); !strings.Contains(got, "[palette]\nbackground = #102030\n") || strings.Contains(got, "window_body") {
		t.Fatalf("saving after a palette change must keep only the palette line:\n%s", got)
	}
	if err := SetToken("theme", "window_body", "@panel darken(10%)"); err != nil {
		t.Fatal(err)
	}
	out.Reset()
	if err := WriteOverrides(&out); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "window_body = @panel darken(10%)") {
		t.Fatalf("a typed reference must be saved as written:\n%s", out.String())
	}
}

func TestDefaultDumpWritesPaletteFirstAndReferences(t *testing.T) {
	var out bytes.Buffer
	if err := WriteDefaultLayout(&out); err != nil {
		t.Fatal(err)
	}
	dump := out.String()
	palette, theme := strings.Index(dump, "\n[palette]\n"), strings.Index(dump, "\n[theme]\n")
	if palette < 0 || theme < 0 || palette > theme {
		t.Fatal("the dump must start with [palette]")
	}
	if !strings.Contains(dump, "window_body = @background\n") || !strings.Contains(dump, "(= #ffffff)") {
		t.Fatal("the dump must write palette references with their resolved color")
	}
}
