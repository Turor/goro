package res

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/kivutar/goro/db"
)

func writeEffectFixture(t *testing.T, files map[string][]byte) string {
	t.Helper()
	root := t.TempDir()
	for name, body := range files {
		path := filepath.Join(root, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, body, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

func effectManager(t *testing.T, files map[string][]byte) *Manager {
	t.Helper()
	m, err := NewManager(writeEffectFixture(t, files))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(db.RestoreCompiledEffectTables)
	return m
}

func TestEnsureEffectDataLoadsJSONOverlay(t *testing.T) {
	body, err := os.ReadFile(filepath.Join("..", "db", "testdata", "effects.json"))
	if err != nil {
		t.Fatal(err)
	}
	m := effectManager(t, map[string][]byte{
		"data/contentdata/effectdata/effects.json": body,
	})
	if !m.EnsureEffectData() {
		t.Fatal("effects.json not installed")
	}
	spec := db.EffectSpecs[7]
	if spec.Duration != 400*time.Millisecond || spec.Layer != db.EffectLayerActor {
		t.Fatalf("effect 7 = %+v", spec)
	}
	if db.SkillEffects[5].SuccessEffectIDsSelf[0] != 7 {
		t.Fatalf("skill 5 = %+v", db.SkillEffects[5])
	}
	if db.SkillActions[5] != db.SkillActionAttack {
		t.Fatalf("action 5 = %v", db.SkillActions[5])
	}
	if db.EffectSpecs[311].Layer != db.EffectLayerActor {
		t.Fatal("compiled fallback lost Crazy Uproar")
	}
}

func TestEnsureEffectDataMissingFileKeepsCompiled(t *testing.T) {
	m := effectManager(t, nil)
	if m.EnsureEffectData() {
		t.Fatal("installed overlay without files")
	}
	if db.EffectTablesUseDataFile() {
		t.Fatal("marked as data file")
	}
	if db.EffectSpecs[311].Layer != db.EffectLayerActor {
		t.Fatal("compiled loud missing")
	}
}

func TestReloadEffectDataReplacesSpec(t *testing.T) {
	rootJSON := []byte(`{"effects":{"7":{"sfx":["first.wav"],"components":[{"kind":"str","file":"first"}]}}}`)
	m := effectManager(t, map[string][]byte{
		"data/effects.json": rootJSON,
	})
	if !m.EnsureEffectData() {
		t.Fatal("first load failed")
	}
	if db.EffectSpecs[7].SFX[0] != "first.wav" {
		t.Fatalf("first = %+v", db.EffectSpecs[7])
	}
	path := filepath.Join(m.Root, "data", "effects.json")
	second := []byte(`{"effects":{"7":{"sfx":["second.wav"],"components":[{"kind":"str","file":"second"}]}}}`)
	if err := os.WriteFile(path, second, 0o644); err != nil {
		t.Fatal(err)
	}
	before := db.EffectTablesGeneration()
	if !m.ReloadEffectData() {
		t.Fatal("reload failed")
	}
	if db.EffectTablesGeneration() == before {
		t.Fatal("generation unchanged")
	}
	if db.EffectSpecs[7].SFX[0] != "second.wav" || db.EffectSpecs[7].Components[0].STRFile != "second" {
		t.Fatalf("after reload = %+v", db.EffectSpecs[7])
	}
}

func TestBadEffectFileKeepsPreviousTables(t *testing.T) {
	m := effectManager(t, map[string][]byte{
		"data/effects.json": []byte(`{"effects":{"7":{"sfx":["ok.wav"],"components":[{"kind":"str","file":"ok"}]}}}`),
	})
	if !m.EnsureEffectData() {
		t.Fatal("good load failed")
	}
	path := filepath.Join(m.Root, "data", "effects.json")
	if err := os.WriteFile(path, []byte(`{"effects": {"nope": {}}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	before := db.EffectSpecs[7].SFX[0]
	gen := db.EffectTablesGeneration()
	m.ReloadEffectData()
	if db.EffectSpecs[7].SFX[0] != before {
		t.Fatalf("bad file replaced tables: %+v", db.EffectSpecs[7])
	}
	if db.EffectTablesGeneration() != gen {
		t.Fatal("generation bumped after bad file")
	}
}

func TestEz2BSONMergeOnLoad(t *testing.T) {
	bson := db.EncodeBSONDocument(map[string]any{
		"MC_LOUD": map[string]any{
			"FilePath": `data\texture\effect\loud.str`,
			"IsFloor":  false,
			"":         []string{"FilePath", "IsFloor"},
		},
	}, []string{"MC_LOUD"})
	m := effectManager(t, map[string][]byte{
		"data/contentdata/effectdata/ez2streffect.bson": bson,
	})
	if !m.EnsureEffectData() {
		t.Fatal("bson not installed")
	}
	id, ok := db.EffectNameID("MC_LOUD")
	if !ok {
		t.Fatal("MC_LOUD id missing")
	}
	if db.EffectSpecs[id].Components[0].STRFile != "loud" {
		t.Fatalf("merged spec = %+v", db.EffectSpecs[id])
	}
	if !containsInt(db.SkillEffects[db.SkillMCLoud].EffectIDs, id) {
		t.Fatalf("skill merge = %+v", db.SkillEffects[db.SkillMCLoud])
	}
}

func TestInvalidateDataReloadsEffects(t *testing.T) {
	m := effectManager(t, map[string][]byte{
		"data/effects.json": []byte(`{"effects":{"7":{"sfx":["a.wav"],"components":[{"kind":"str","file":"a"}]}}}`),
	})
	if !m.EnsureEffectData() {
		t.Fatal("load failed")
	}
	path := filepath.Join(m.Root, "data", "effects.json")
	if err := os.WriteFile(path, []byte(`{"effects":{"7":{"sfx":["b.wav"],"components":[{"kind":"str","file":"b"}]}}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if m.InvalidateData([]string{"data/texture/x.bmp"}) {
		t.Fatal("non-effect invalidation reloaded")
	}
	if !m.InvalidateData([]string{"data/contentdata/effectdata/effects.json"}) {
		t.Fatal("effects invalidation did not reload")
	}
	if db.EffectSpecs[7].SFX[0] != "b.wav" {
		t.Fatalf("after invalidate = %+v", db.EffectSpecs[7])
	}
}

func TestIsEffectDataFile(t *testing.T) {
	for name, want := range map[string]bool{
		"data/contentdata/effectdata/effects.json":          true,
		"data/contentdata/effectdata/ez2streffect.bson":     true,
		`data\effectdata\effects.overlay.json`:              true,
		"data/texture/effect/loud.str":                      false,
		"data/luafiles514/lua files/skillinfoz/skillid.lub": false,
	} {
		if got := IsEffectDataFile(name); got != want {
			t.Errorf("IsEffectDataFile(%q) = %v", name, got)
		}
	}
}

func TestMissingBSONIsNotAnError(t *testing.T) {
	m := effectManager(t, map[string][]byte{
		"data/effects.json": []byte(`{"effects":{"7":{"sfx":["ok.wav"],"components":[{"kind":"str","file":"ok"}]}}}`),
	})
	if !m.EnsureEffectData() {
		t.Fatal("json load failed without bson")
	}
	if db.EffectSpecs[7].SFX[0] != "ok.wav" {
		t.Fatalf("got %+v", db.EffectSpecs[7])
	}
}

func TestCorruptBSONSkipsMergeKeepsJSON(t *testing.T) {
	m := effectManager(t, map[string][]byte{
		"data/contentdata/effectdata/effects.json":      []byte(`{"effects":{"7":{"sfx":["ok.wav"],"components":[{"kind":"str","file":"ok"}]}}}`),
		"data/contentdata/effectdata/ez2streffect.bson": []byte{1, 2, 3, 4},
	})
	if !m.EnsureEffectData() {
		t.Fatal("json load failed when bson was corrupt")
	}
	if db.EffectSpecs[7].SFX[0] != "ok.wav" {
		t.Fatalf("corrupt bson dropped json overlay: %+v", db.EffectSpecs[7])
	}
	if _, ok := db.EffectNameID("MC_LOUD"); ok {
		t.Fatal("corrupt bson still merged names")
	}
}

func containsInt(ids []int, id int) bool {
	for _, v := range ids {
		if v == id {
			return true
		}
	}
	return false
}
