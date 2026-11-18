package main

import (
	"os"
	"strings"
	"testing"

	"github.com/kivutar/goro/ui/rotheme"
)

// TestBuiltinThemesUseKnownKeys parses every shipped theme with every token
// package linked, so a renamed setting cannot silently drop out of a theme.
func TestBuiltinThemesUseKnownKeys(t *testing.T) {
	if problems := rotheme.CheckBuiltinThemes(); len(problems) > 0 {
		t.Fatalf("built-in theme problems:\n%s", strings.Join(problems, "\n"))
	}
}

// The dark example in the docs is the shipped Dark theme.
func TestDarkExampleMatchesBuiltinTheme(t *testing.T) {
	settings := func(path string) []string {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		var out []string
		for _, line := range strings.Split(string(data), "\n") {
			line = strings.TrimSpace(line)
			if line != "" && !strings.HasPrefix(line, "#") {
				out = append(out, line)
			}
		}
		return out
	}
	doc, shipped := settings("docs/ui-palette-dark.example.ini"), settings("ui/rotheme/themes/dark.ini")
	if strings.Join(doc, "\n") != strings.Join(shipped, "\n") {
		t.Fatalf("docs/ui-palette-dark.example.ini differs from ui/rotheme/themes/dark.ini:\n%s\n---\n%s",
			strings.Join(doc, "\n"), strings.Join(shipped, "\n"))
	}
}
