package res

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestMissingCheckerboardIsOpaqueMagentaAndBlack(t *testing.T) {
	img := MissingCheckerboard()
	if img == nil || img.Bounds() != image.Rect(0, 0, 64, 64) {
		t.Fatalf("checkerboard bounds = %v, want 64x64", img.Bounds())
	}
	if img != MissingCheckerboard() {
		t.Fatal("MissingCheckerboard must return the same image")
	}
	magenta := color.NRGBA{R: 255, B: 255, A: 255}
	black := color.NRGBA{A: 255}
	if got := img.NRGBAAt(0, 0); got != magenta {
		t.Fatalf("cell (0,0) = %+v, want opaque magenta", got)
	}
	if got := img.NRGBAAt(8, 0); got != black {
		t.Fatalf("cell (8,0) = %+v, want opaque black", got)
	}
	if got := img.NRGBAAt(0, 8); got != black {
		t.Fatalf("cell (0,8) = %+v, want opaque black", got)
	}
	if got := img.NRGBAAt(8, 8); got != magenta {
		t.Fatalf("cell (8,8) = %+v, want opaque magenta", got)
	}
	before := img.NRGBAAt(0, 0)
	_ = applyROTransparency(img)
	if img.NRGBAAt(0, 0) != before {
		t.Fatal("shared placeholder must stay opaque; RO chroma-keying copies")
	}
	if !IsMissingCheckerboard(img) || IsMissingCheckerboard(image.NewNRGBA(img.Bounds())) {
		t.Fatal("IsMissingCheckerboard must match the shared placeholder only")
	}
}

func TestRecordMissingTextureLogsEachPathOnce(t *testing.T) {
	path := fmt.Sprintf("g61-missing-once/%s/%d", t.Name(), time.Now().UnixNano())
	if !RecordMissingTexture(path) {
		t.Fatal("first recording of a path must be new")
	}
	if RecordMissingTexture(path) || RecordMissingTexture("  "+path+"  ") {
		t.Fatal("later recordings of the same path must be ignored")
	}
	if RecordMissingTexture("") || RecordMissingTexture("   ") {
		t.Fatal("empty paths must not be recorded")
	}
}

func TestLoadTextureSubstitutesCheckerboardForPermanentMiss(t *testing.T) {
	root := t.TempDir()
	manager := &Manager{Root: root}
	path := filepath.Join("g61-missing", t.Name(), "hole.png")
	img, source, err := LoadTexture(manager, []string{path})
	if err != nil {
		t.Fatalf("LoadTexture: %v", err)
	}
	if !IsMissingCheckerboard(img) {
		t.Fatal("missing file must return the checkerboard")
	}
	if source != path {
		t.Fatalf("source = %q, want first candidate %q", source, path)
	}
	again, _, err := LoadTexture(manager, []string{path})
	if err != nil || !IsMissingCheckerboard(again) {
		t.Fatalf("retry: img missing=%t err=%v", IsMissingCheckerboard(again), err)
	}
	if RecordMissingTexture(path) {
		t.Fatal("LoadTexture must record the missing path only once")
	}
}

func TestLoadTextureExactReadsPresentFile(t *testing.T) {
	root := t.TempDir()
	name := "present.png"
	writeSolidPNG(t, filepath.Join(root, name), color.NRGBA{G: 200, A: 255})
	img, source, err := LoadTextureExact(&Manager{Root: root}, []string{name})
	if err != nil {
		t.Fatal(err)
	}
	if IsMissingCheckerboard(img) {
		t.Fatal("present file must not be replaced with the checkerboard")
	}
	if source != name {
		t.Fatalf("source = %q, want %q", source, name)
	}
	got := color.NRGBAModel.Convert(img.At(0, 0)).(color.NRGBA)
	if got != (color.NRGBA{G: 200, A: 255}) {
		t.Fatalf("pixel = %+v", got)
	}
}

func TestTextureFromLoadLeavesRemotePendingUnlogged(t *testing.T) {
	path := fmt.Sprintf("g61-pending/%s/%d", t.Name(), time.Now().UnixNano())
	img, source, err := textureFromLoad(nil, "", ErrRemotePending, []string{path})
	if !errors.Is(err, ErrRemotePending) {
		t.Fatalf("err = %v, want ErrRemotePending", err)
	}
	if img != nil || source != "" {
		t.Fatalf("pending result img=%v source=%q", img, source)
	}
	if !RecordMissingTexture(path) {
		t.Fatal("ErrRemotePending must not record the path as missing")
	}
}

func writeSolidPNG(t *testing.T, path string, c color.NRGBA) {
	t.Helper()
	img := image.NewNRGBA(image.Rect(0, 0, 1, 1))
	img.SetNRGBA(0, 0, c)
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		t.Fatal(err)
	}
}
