package rotheme

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

func TestTunerEditsKeepOtherValuesAndSaveOnlyChanges(t *testing.T) {
	resetLayoutAfter(t)
	if err := SetToken("text", "size", "15"); err != nil {
		t.Fatal(err)
	}
	if err := SetToken("button", "radius", "99"); err == nil {
		t.Fatal("an out-of-range value was applied")
	}
	var radius TokenInfo
	for _, info := range Tokens() {
		if info.Section == "button" && info.Key == "radius" {
			radius = info
		}
	}
	for range 40 {
		if err := StepToken(radius, 1); err != nil {
			t.Fatal(err)
		}
		radius.Value = fmtFloat(ButtonRadius)
	}
	if ButtonRadius != 16 || Default.Typography.TextSize != 15 {
		t.Fatalf("radius %v (want clamped to 16), text size %v (want kept)", ButtonRadius, Default.Typography.TextSize)
	}
	var out bytes.Buffer
	if err := WriteOverrides(&out); err != nil {
		t.Fatal(err)
	}
	if got := out.String(); !strings.Contains(got, "[text]\nsize = 15\n") || !strings.Contains(got, "radius = 16") || strings.Contains(got, "window_body") {
		t.Fatalf("overrides file:\n%s", got)
	}
	path := filepath.Join(t.TempDir(), "ui-layout.ini")
	if err := SaveOverrides(path); err != nil {
		t.Fatal(err)
	}
	ResetLayout()
	if applied, _, err := LoadLayoutFile(path); !applied || err != nil || ButtonRadius != 16 || Default.Typography.TextSize != 15 {
		t.Fatalf("saved file did not restore the edits: applied=%v err=%v", applied, err)
	}
}

func fmtFloat(f float32) string {
	return (&token{kind: tokenFloat}).format(tokenValue{f: f})
}
