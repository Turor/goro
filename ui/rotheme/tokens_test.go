package rotheme

import (
	"bytes"
	"errors"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gogpu/ui/widget"
)

func resetLayoutAfter(t *testing.T) {
	t.Helper()
	t.Cleanup(ResetLayout)
}

func TestLayoutDefaultsRoundTrip(t *testing.T) {
	resetLayoutAfter(t)
	var first bytes.Buffer
	if err := WriteDefaultLayout(&first); err != nil {
		t.Fatal(err)
	}
	layout, warnings, err := ParseLayout(bytes.NewReader(first.Bytes()), "dump")
	if err != nil || len(warnings) != 0 {
		t.Fatalf("dump does not parse cleanly: err=%v warnings=%v", err, warnings)
	}
	if layout.Len() != len(tokens) {
		t.Fatalf("dump names %d keys, want every token (%d)", layout.Len(), len(tokens))
	}
	before := Default
	layout.Apply()
	if Default != before {
		t.Fatalf("applying the dumped defaults changed the theme:\n got %+v\nwant %+v", Default, before)
	}
	var second bytes.Buffer
	if err := WriteLayout(&second); err != nil {
		t.Fatal(err)
	}
	if first.String() != second.String() {
		t.Fatal("dump is not stable across a parse and apply")
	}
}

func TestLayoutApplyOverridesAndOmittedKeysRevert(t *testing.T) {
	resetLayoutAfter(t)
	gen := Generation()
	apply(t, "[theme]\nwindow_body = #10203080\n[text]\nsize = 14\n[button]\npadding_x = 12 # wider\n")
	if got := Default.Colors.WindowBody; got != widget.RGBA8(0x10, 0x20, 0x30, 0x80) {
		t.Fatalf("window_body = %+v", got)
	}
	if Default.Typography.TextSize != 14 || ButtonPaddingX != 12 {
		t.Fatalf("text size %v, padding_x %v", Default.Typography.TextSize, ButtonPaddingX)
	}
	if Generation() == gen {
		t.Fatal("apply did not advance the generation")
	}
	apply(t, "[text]\nsize = 13\n")
	if Default.Colors.WindowBody != tokenIndex["theme.window_body"].def.c || ButtonPaddingX != 8 {
		t.Fatal("keys removed from the file did not return to their defaults")
	}
}

func TestLayoutRejectsBadFilesAndKeepsLastGood(t *testing.T) {
	resetLayoutAfter(t)
	apply(t, "[text]\nsize = 15\n")
	cases := map[string]string{
		"out of range": "[text]\nsize = 15\n[hotbar]\n[button]\nradius = 99\n",
		"nan":          "[button]\nradius = NaN\n",
		"bad color":    "[theme]\nwindow_body = #12345\n",
		"no equals":    "[theme]\nwindow_body\n",
		"bad header":   "[theme\n",
		"not a number": "[progress_bar]\nheight = tall\n",
	}
	for name, src := range cases {
		_, _, err := ParseLayout(strings.NewReader(src), "ui-layout.ini")
		var layoutErr *LayoutError
		if !errors.As(err, &layoutErr) || layoutErr.Line == 0 || !strings.Contains(err.Error(), "ui-layout.ini:") {
			t.Errorf("%s: want a line-numbered LayoutError, got %v", name, err)
		}
	}
	if Default.Typography.TextSize != 15 {
		t.Fatal("a rejected file changed the current values")
	}
}

func TestLayoutUnknownKeysWarn(t *testing.T) {
	layout, warnings, err := ParseLayout(strings.NewReader("[theme]\nnot_a_key = 1\ntext = #000000\n"), "f")
	if err != nil || len(warnings) != 1 || !strings.Contains(warnings[0], "f:2") || layout.Len() != 1 {
		t.Fatalf("err=%v warnings=%v len=%d", err, warnings, layout.Len())
	}
}

func TestLoadLayoutFileMissingIsQuiet(t *testing.T) {
	resetLayoutAfter(t)
	applied, warnings, err := LoadLayoutFile(filepath.Join(t.TempDir(), "none.ini"))
	if applied || warnings != nil || err != nil {
		t.Fatalf("applied=%v warnings=%v err=%v", applied, warnings, err)
	}
	path := filepath.Join(t.TempDir(), "ui-layout.ini")
	if err := os.WriteFile(path, []byte("[text]\nsize = 12\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if applied, _, err := LoadLayoutFile(path); !applied || err != nil || Default.Typography.TextSize != 12 {
		t.Fatalf("applied=%v err=%v size=%v", applied, err, Default.Typography.TextSize)
	}
}

// Random partial and corrupted files must never panic or apply bad values.
func TestLayoutParseNeverPanics(t *testing.T) {
	var dump bytes.Buffer
	if err := WriteDefaultLayout(&dump); err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(dump.String(), "\n")
	rng := rand.New(rand.NewSource(4))
	junk := []string{"", "=", "[", "]", "#", "-1", "1e9", "NaN", "#ffffffff", "0x10", "\x00", " = 3", "[x]"}
	for i := 0; i < 300; i++ {
		var b strings.Builder
		for _, line := range lines {
			switch rng.Intn(6) {
			case 0:
				continue
			case 1:
				line += junk[rng.Intn(len(junk))]
			case 2:
				line = junk[rng.Intn(len(junk))] + line
			}
			b.WriteString(line + "\n")
		}
		layout, _, err := ParseLayout(strings.NewReader(b.String()), "fuzz")
		if err == nil {
			for tok, v := range layout.values {
				if tok.kind == tokenInt && (float64(v.i) < tok.min || float64(v.i) > tok.max) ||
					tok.kind == tokenFloat && (float64(v.f) < tok.min || float64(v.f) > tok.max) {
					t.Fatalf("accepted out-of-range %s.%s", tok.section, tok.key)
				}
			}
		}
	}
}

func apply(t *testing.T, src string) {
	t.Helper()
	layout, _, err := ParseLayout(strings.NewReader(src), "test")
	if err != nil {
		t.Fatal(err)
	}
	layout.Apply()
}
