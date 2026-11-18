package rotheme

import (
	"image/color"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// resetThemes restores the default theme, no user themes folder and no user
// layout after a test.
func resetThemes(t *testing.T) {
	t.Helper()
	t.Cleanup(func() {
		SetThemesDir("")
		if _, err := SetTheme(DefaultThemeID); err != nil {
			t.Errorf("restore default theme: %v", err)
		}
		ResetLayout()
	})
}

func themeTokenColor(t *testing.T, id string) color.RGBA {
	t.Helper()
	tok := tokenIndex[id]
	if tok == nil || !isColorToken(tok) {
		t.Fatalf("no color token %s", id)
	}
	return tokenRawRGBA(tok, tok.current())
}

func applyLayoutText(t *testing.T, text string) {
	t.Helper()
	layout, _, err := ParseLayout(strings.NewReader(text), "test")
	if err != nil {
		t.Fatal(err)
	}
	layout.Apply()
}

func hexRGBA(t *testing.T, s string) color.RGBA {
	t.Helper()
	r, g, b, a, err := parseHexColor(s)
	if err != nil {
		t.Fatal(err)
	}
	return color.RGBA{R: r, G: g, B: b, A: a}
}

func writeFile(t *testing.T, path, text string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestListThemesBuiltinThenUserFiles(t *testing.T) {
	resetThemes(t)
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "ocean.ini"), "# name: Ocean Blue\n[palette]\naccent = #1060a0\n")
	writeFile(t, filepath.Join(dir, "Sunset_Glow.INI"), "[palette]\naccent = #e06020\n")
	writeFile(t, filepath.Join(dir, "dark.ini"), "[palette]\naccent = #ff0000\n") // shadows a built-in: skipped
	writeFile(t, filepath.Join(dir, "bad name.ini"), "[palette]\n")               // invalid ID: skipped
	writeFile(t, filepath.Join(dir, "notes.txt"), "not a theme\n")
	if err := os.Mkdir(filepath.Join(dir, "folder.ini"), 0o755); err != nil {
		t.Fatal(err)
	}
	SetThemesDir(dir)
	var ids, names []string
	for _, info := range ListThemes() {
		ids = append(ids, info.ID)
		names = append(names, info.Name)
		if info.Builtin != (info.Path == "") {
			t.Errorf("%s: builtin=%v path=%q", info.ID, info.Builtin, info.Path)
		}
	}
	wantIDs := []string{"default", "dark", "complementary", "analogous", "triadic", "ocean", "sunset_glow"}
	wantNames := []string{"Default", "Dark", "Complementary", "Analogous", "Triadic", "Ocean Blue", "Sunset Glow"}
	if !reflect.DeepEqual(ids, wantIDs) || !reflect.DeepEqual(names, wantNames) {
		t.Fatalf("themes = %v %v, want %v %v", ids, names, wantIDs, wantNames)
	}
	SetThemesDir(filepath.Join(dir, "missing"))
	if got := len(ListThemes()); got != len(builtinThemes) {
		t.Fatalf("missing themes folder lists %d themes, want the %d built-in", got, len(builtinThemes))
	}
}

func TestBuiltinThemesParseCleanly(t *testing.T) {
	for _, id := range builtinThemes {
		layout, warnings, err := loadTheme(id)
		if err != nil {
			t.Fatalf("theme %s: %v", id, err)
		}
		// Keys of the ui package are unknown here; the root package test
		// (TestBuiltinThemesUseKnownKeys) checks them with every token linked.
		for _, warning := range warnings {
			if !strings.Contains(warning, "unknown key [theme]") {
				t.Errorf("theme %s: %s", id, warning)
			}
		}
		if id == DefaultThemeID {
			if layout.Len() != 0 {
				t.Fatalf("default theme sets %d keys; it must be the compiled defaults", layout.Len())
			}
			continue
		}
		for _, key := range PaletteKeys() {
			if _, ok := layout.values[paletteToken(key)]; !ok {
				t.Errorf("theme %s does not set palette.%s", id, key)
			}
		}
	}
}

func TestSetThemeReresolvesEveryReference(t *testing.T) {
	resetThemes(t)
	defaults := map[string]color.RGBA{}
	for _, ref := range ColorTokenDefaults() {
		defaults[ref.Section+"."+ref.Key] = themeTokenColor(t, ref.Section+"."+ref.Key)
	}
	if _, err := SetTheme("dark"); err != nil {
		t.Fatal(err)
	}
	if ActiveTheme() != "dark" {
		t.Fatalf("active theme = %q", ActiveTheme())
	}
	if got, want := themeTokenColor(t, "palette.background"), hexRGBA(t, "#1c1f26"); got != want {
		t.Fatalf("palette.background = %v, want %v", got, want)
	}
	// Every setting that follows plain @background now has the dark color.
	followers := 0
	for _, ref := range ColorTokenDefaults() {
		if ref.Ref != "@background" {
			continue
		}
		followers++
		if got := themeTokenColor(t, ref.Section+"."+ref.Key); got != hexRGBA(t, "#1c1f26") {
			t.Errorf("%s.%s = %v, want the dark background", ref.Section, ref.Key, got)
		}
	}
	if followers == 0 {
		t.Fatal("no setting follows @background")
	}
	// The theme's own [theme] reference applies.
	if got, want := themeTokenColor(t, "theme.title_text"), themeTokenColor(t, "palette.text"); got != want {
		t.Fatalf("theme.title_text = %v, want dark @text %v", got, want)
	}
	changed := 0
	for id, def := range defaults {
		if themeTokenColor(t, id) != def {
			changed++
		}
	}
	if changed < len(defaults)/2 {
		t.Fatalf("dark theme changed only %d of %d colors", changed, len(defaults))
	}
	// Back to the default theme: every color is the compiled one again.
	if _, err := SetTheme("default"); err != nil {
		t.Fatal(err)
	}
	for id, def := range defaults {
		if got := themeTokenColor(t, id); got != def {
			t.Fatalf("%s = %v after returning to default, want %v", id, got, def)
		}
	}
}

func TestUserLayoutOverridesWinOverTheme(t *testing.T) {
	resetThemes(t)
	applyLayoutText(t, "[palette]\naccent = #123456\n[theme]\ntitle_text = #010203\n")
	if _, err := SetTheme("dark"); err != nil {
		t.Fatal(err)
	}
	if got := themeTokenColor(t, "palette.accent"); got != hexRGBA(t, "#123456") {
		t.Fatalf("user palette.accent lost to the theme: %v", got)
	}
	if got := themeTokenColor(t, "theme.title_text"); got != hexRGBA(t, "#010203") {
		t.Fatalf("user theme.title_text lost to the theme: %v", got)
	}
	if got := themeTokenColor(t, "palette.background"); got != hexRGBA(t, "#1c1f26") {
		t.Fatalf("palette.background = %v, want the theme's", got)
	}
	// A user setting that follows @accent uses the user's accent.
	if got := themeTokenColor(t, "theme.label_text"); got != hexRGBA(t, "#123456") {
		t.Fatalf("theme.label_text (= @accent in dark) = %v, want the user accent", got)
	}
	// Reloading ui-layout.ini (here: removing it) keeps the theme.
	ResetLayout()
	if got := themeTokenColor(t, "palette.accent"); got != hexRGBA(t, "#c8a24a") {
		t.Fatalf("palette.accent = %v after the user layout went away, want dark's", got)
	}
	if ActiveTheme() != "dark" {
		t.Fatalf("active theme = %q after reset", ActiveTheme())
	}
}

func TestTunerEditsDoNotCopyTheTheme(t *testing.T) {
	resetThemes(t)
	if _, err := SetTheme("triadic"); err != nil {
		t.Fatal(err)
	}
	if n := currentLayout().Len(); n != 0 {
		t.Fatalf("with only a theme active the user layout has %d values", n)
	}
	if err := SetToken("palette", "accent", "#111111"); err != nil {
		t.Fatal(err)
	}
	if n := currentLayout().Len(); n != 1 {
		t.Fatalf("one tuner edit saved %d values", n)
	}
	var b strings.Builder
	if err := WriteOverrides(&b); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(b.String(), "background") {
		t.Fatalf("saved overrides copied the theme:\n%s", b.String())
	}
	for _, info := range Tokens() {
		if info.Section == "palette" && info.Key == "background" && info.Default != "#191523" {
			t.Fatalf("tuner default for palette.background = %s, want the theme's", info.Default)
		}
	}
}

func TestSetThemeRejectsUnknownAndBrokenThemes(t *testing.T) {
	resetThemes(t)
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "broken.ini"), "[palette]\naccent = not-a-color\n")
	writeFile(t, filepath.Join(dir, "mine.ini"), "[palette]\naccent = #abcdef\n")
	SetThemesDir(dir)
	if _, err := SetTheme("dark"); err != nil {
		t.Fatal(err)
	}
	if _, err := SetTheme("nope"); err == nil {
		t.Fatal("unknown theme accepted")
	}
	if _, err := SetTheme("broken"); err == nil {
		t.Fatal("broken theme accepted")
	}
	if ActiveTheme() != "dark" || themeTokenColor(t, "palette.background") != hexRGBA(t, "#1c1f26") {
		t.Fatal("a rejected theme changed the current one")
	}
	if _, err := SetTheme(" Mine "); err != nil {
		t.Fatal(err)
	}
	if ActiveTheme() != "mine" || themeTokenColor(t, "palette.accent") != hexRGBA(t, "#abcdef") {
		t.Fatal("user theme file not applied")
	}
	// Keys the user theme does not set are the compiled defaults, not dark's.
	if got := themeTokenColor(t, "palette.background"); got != defaultPaletteColor("background") {
		t.Fatalf("palette.background = %v, want the compiled default", got)
	}
}

func relLuminance(c color.RGBA) float64 {
	lin := func(v uint8) float64 {
		s := float64(v) / 255
		if s <= 0.04045 {
			return s / 12.92
		}
		return math.Pow((s+0.055)/1.055, 2.4)
	}
	return 0.2126*lin(c.R) + 0.7152*lin(c.G) + 0.0722*lin(c.B)
}

// contrastRatio is the WCAG 2 contrast ratio of two opaque colors.
func contrastRatio(a, b color.RGBA) float64 {
	la, lb := relLuminance(a), relLuminance(b)
	if la < lb {
		la, lb = lb, la
	}
	return (la + 0.05) / (lb + 0.05)
}

func TestContrastRatioMatchesWCAG(t *testing.T) {
	black, white := color.RGBA{A: 255}, color.RGBA{R: 255, G: 255, B: 255, A: 255}
	if got := contrastRatio(black, white); math.Abs(got-21) > 1e-9 {
		t.Fatalf("black on white = %g, want 21", got)
	}
	if got := contrastRatio(hexRGBA(t, "#777777"), white); math.Abs(got-4.48) > 0.01 {
		t.Fatalf("#777 on white = %g, want 4.48", got)
	}
}

// TestBuiltinThemeTextContrast keeps body text readable in every shipped
// theme: WCAG AA (4.5:1) for text on the background and panel colors, and
// for the window text, title and label settings on what they are drawn on.
func TestBuiltinThemeTextContrast(t *testing.T) {
	resetThemes(t)
	const aa = 4.5
	pairs := [][2]string{
		{"palette.text", "palette.background"},
		{"palette.text", "palette.panel"},
		{"palette.text", "palette.panel_alt"},
		{"palette.text_dim", "palette.background"},
		{"palette.text_dim", "palette.panel"},
		{"theme.text", "theme.window_body"},
		{"theme.title_text", "theme.window_title"},
		{"theme.label_text", "theme.window_body"},
	}
	for _, id := range builtinThemes {
		if _, err := SetTheme(id); err != nil {
			t.Fatal(err)
		}
		for _, pair := range pairs {
			fg, bg := themeTokenColor(t, pair[0]), themeTokenColor(t, pair[1])
			fg.A, bg.A = 255, 255
			if got := contrastRatio(fg, bg); got < aa {
				t.Errorf("theme %s: %s on %s contrast %.2f:1, want at least %.1f:1", id, pair[0], pair[1], got, aa)
			}
		}
	}
}
