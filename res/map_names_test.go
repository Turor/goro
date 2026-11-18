package res

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseMapNameTable(t *testing.T) {
	data := []byte("// comment\r\n" +
		"PRONTERA.RSW#Prontera, Capital of Rune Midgard#\r\n" +
		"payon.gat#Payon Town#\r\n" +
		"geffen#Geffen, the Magic City#\n" +
		"prontera.rsw#ignored duplicate#\n" +
		"badline-without-hashes\n" +
		"#skipped\n" +
		"incomplete.rsw#no end\n" +
		"  izlude.rsw  # Izlude, the Satellite City #\n")
	got := parseMapNameTable(data)
	want := map[string]string{
		"prontera.rsw": "Prontera, Capital of Rune Midgard",
		"payon.rsw":    "Payon Town",
		"geffen.rsw":   "Geffen, the Magic City",
		"izlude.rsw":   "Izlude, the Satellite City",
	}
	if len(got) != len(want) {
		t.Fatalf("got %d entries %#v, want %d", len(got), got, len(want))
	}
	for key, name := range want {
		if got[key] != name {
			t.Errorf("%s = %q, want %q", key, got[key], name)
		}
	}
}

func TestMapDisplayNameLoadsAndCaches(t *testing.T) {
	root := t.TempDir()
	dataDir := filepath.Join(root, "data")
	if err := os.Mkdir(dataDir, 0o700); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dataDir, "mapnametable.txt")
	if err := os.WriteFile(path, []byte("prontera.rsw#Prontera, Capital of Rune Midgard#\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	m, err := NewManager(root)
	if err != nil {
		t.Fatal(err)
	}
	name, ok := m.MapDisplayName("prontera")
	if !ok || name != "Prontera, Capital of Rune Midgard" {
		t.Fatalf("bare name = %q ok=%t", name, ok)
	}
	for _, key := range []string{"PRONTERA.GAT", "data\\\\prontera.rsw", "prontera.rsw"} {
		got, ok := m.MapDisplayName(key)
		if !ok || got != name {
			t.Errorf("MapDisplayName(%q) = %q ok=%t", key, got, ok)
		}
	}
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	if got, ok := m.MapDisplayName("prontera.gat"); !ok || got != name {
		t.Fatal("map names were not cached")
	}
	if _, ok := (*Manager)(nil).MapDisplayName("prontera"); ok {
		t.Fatal("nil manager returned a name")
	}
	if _, ok := m.MapDisplayName("missing_map"); ok {
		t.Fatal("unknown map returned a name")
	}
}

func TestMapDisplayNamesRealWhenConfigured(t *testing.T) {
	m := realDataManager(t)
	name, ok := m.MapDisplayName("prontera")
	if !ok || name != "Prontera, Capital of Rune Midgard" {
		t.Fatalf("prontera = %q ok=%t", name, ok)
	}
	payon, ok := m.MapDisplayName("payon.gat")
	if !ok || payon != "Payon Town" {
		t.Fatalf("payon = %q ok=%t", payon, ok)
	}
	if len(m.mapDisplayNames) < 1000 {
		t.Fatalf("only %d map names", len(m.mapDisplayNames))
	}
	t.Logf("%d map display names", len(m.mapDisplayNames))
}
