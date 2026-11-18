package app

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/kivutar/goro/ui/rotheme"
)

func TestLoadUIFontIsOptionalAndRejectsBadFiles(t *testing.T) {
	t.Cleanup(func() { rotheme.FontFile = ""; loadedUIFont = uiFontState{} })
	dir := t.TempDir()
	layout := filepath.Join(dir, "ui-layout.ini")
	if err := loadUIFont(layout); err != nil {
		t.Fatalf("no font_file must be a no-op, got %v", err)
	}
	rotheme.FontFile = "missing.ttf"
	if err := loadUIFont(layout); err == nil {
		t.Fatal("a missing font_file was accepted")
	}
	if err := os.WriteFile(filepath.Join(dir, "bad.ttf"), []byte("not a font"), 0o644); err != nil {
		t.Fatal(err)
	}
	rotheme.FontFile = "bad.ttf"
	if err := loadUIFont(layout); err == nil {
		t.Fatal("a file that is not a font was accepted")
	}
	if err := os.Mkdir(filepath.Join(dir, "dir.ttf"), 0o755); err != nil {
		t.Fatal(err)
	}
	rotheme.FontFile = "dir.ttf"
	if err := loadUIFont(layout); err == nil {
		t.Fatal("a directory was accepted as font_file")
	}
	big, err := os.Create(filepath.Join(dir, "big.ttf"))
	if err != nil {
		t.Fatal(err)
	}
	if err := big.Truncate(maxUIFontFileSize + 1); err != nil {
		t.Fatal(err)
	}
	big.Close()
	rotheme.FontFile = "big.ttf"
	if err := loadUIFont(layout); err == nil {
		t.Fatal("a font_file over the size limit was accepted")
	}
}
