package db

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

func testdataFile(t *testing.T, name string) []byte {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("caller")
	}
	data, err := os.ReadFile(filepath.Join(filepath.Dir(file), "testdata", name))
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func TestParseEffectTablesFixture(t *testing.T) {
	tables, err := ParseEffectTablesJSON(testdataFile(t, "effects.json"))
	if err != nil {
		t.Fatal(err)
	}
	spec, ok := tables.Effects[7]
	if !ok {
		t.Fatal("missing effect 7")
	}
	if spec.Duration != 400*time.Millisecond || spec.Layer != EffectLayerActor || spec.ZOrder != 2 || spec.Depth != 0.25 {
		t.Fatalf("effect 7 = %+v", spec)
	}
	if len(spec.SFX) != 1 || spec.SFX[0] != "custom.wav" {
		t.Fatalf("sfx = %v", spec.SFX)
	}
	if len(spec.Components) != 1 || spec.Components[0].Kind != EffectComponentSTR || spec.Components[0].STRFile != "custom" || !spec.Components[0].AttachedEntity {
		t.Fatalf("component = %+v", spec.Components)
	}
	skill, ok := tables.SkillEffects[5]
	if !ok || skill.EffectTiming != SkillEffectAtActionStart || len(skill.SuccessEffectIDsSelf) != 1 || skill.SuccessEffectIDsSelf[0] != 7 {
		t.Fatalf("skill 5 = %+v", skill)
	}
	if tables.SkillActions[5] != SkillActionAttack {
		t.Fatalf("action = %v", tables.SkillActions[5])
	}
}

func TestParseEffectTablesBadFile(t *testing.T) {
	if _, err := ParseEffectTablesJSON([]byte(`{`)); err == nil {
		t.Fatal("truncated JSON accepted")
	}
	if _, err := ParseEffectTablesJSON([]byte(`{"effects": {"nope": {}}}`)); err == nil {
		t.Fatal("non-numeric effect id accepted")
	}
	if _, err := ParseEffectTablesJSON([]byte(`{"effects": {"1": {"components": [{"kind": "nope"}]}}}`)); err == nil {
		t.Fatal("unknown kind accepted")
	}
	if _, err := ParseEffectTablesJSON([]byte(`{"skill_actions": {"1": "fly"}}`)); err == nil {
		t.Fatal("unknown action accepted")
	}
}

func TestOverlayEffectTablesReplacesSpec(t *testing.T) {
	overlay, err := ParseEffectTablesJSON(testdataFile(t, "effects.json"))
	if err != nil {
		t.Fatal(err)
	}
	specs, skills, actions := OverlayEffectTables(overlay)
	if specs[7].SFX[0] != "custom.wav" {
		t.Fatalf("overlay did not replace effect 7: %+v", specs[7])
	}
	if compiledEffectSpecs[7].Duration == specs[7].Duration && len(compiledEffectSpecs[7].SFX) > 0 && compiledEffectSpecs[7].SFX[0] == "custom.wav" {
		t.Fatal("compiled fallback mutated")
	}
	if skills[5].SuccessEffectIDsSelf[0] != 7 {
		t.Fatalf("skill overlay = %+v", skills[5])
	}
	if actions[5] != SkillActionAttack {
		t.Fatalf("action overlay = %v", actions[5])
	}
	if _, ok := specs[effectLoud]; !ok {
		t.Fatal("compiled loud effect dropped")
	}
}

func TestCompiledEffectTablesRoundTripLoud(t *testing.T) {
	tables := EffectTables{
		Effects:      map[int]EffectSpec{effectLoud: cloneEffectSpec(compiledEffectSpecs[effectLoud])},
		SkillEffects: map[uint16]SkillEffectSpec{SkillMCLoud: cloneSkillEffectSpec(compiledSkillEffects[SkillMCLoud])},
		SkillActions: map[uint16]SkillActionKind{SkillMCLoud: compiledSkillActions[SkillMCLoud]},
	}
	raw, err := MarshalEffectTablesJSON(tables)
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := ParseEffectTablesJSON(raw)
	if err != nil {
		t.Fatal(err)
	}
	got := parsed.Effects[effectLoud]
	if got.Layer != EffectLayerActor || got.Depth != 0.05 || got.ZOrder != 10 {
		t.Fatalf("loud layering = layer %d z %d depth %v", got.Layer, got.ZOrder, got.Depth)
	}
	if len(got.Components) != 1 || got.Components[0].STRFile != "loud" {
		t.Fatalf("loud component = %+v", got.Components)
	}
	if parsed.SkillEffects[SkillMCLoud].SuccessEffectIDsSelf[0] != effectLoud {
		t.Fatalf("MC_LOUD = %+v", parsed.SkillEffects[SkillMCLoud])
	}
}

func TestEffectDrawsBeforeEntitiesUsesLayer(t *testing.T) {
	spec := EffectSpec{Layer: EffectLayerActor}
	if EffectDrawsBeforeEntities(spec, EffectComponent{RenderBefore: true}) {
		t.Fatal("actor layer still used RenderBefore")
	}
	spec.Layer = EffectLayerGround
	if !EffectDrawsBeforeEntities(spec, EffectComponent{}) {
		t.Fatal("ground layer not before entities")
	}
	if EffectDrawsBeforeEntities(EffectSpec{}, EffectComponent{}) {
		t.Fatal("default drew before entities")
	}
	if !EffectDrawsBeforeEntities(EffectSpec{}, EffectComponent{RenderBefore: true}) {
		t.Fatal("RenderBefore ignored when layer is auto")
	}
	if EffectDrawDepth(EffectSpec{Depth: 0.05}, EffectComponent{}) != 0.05 {
		t.Fatal("spec depth not used")
	}
	if EffectDrawZOrder(EffectSpec{ZOrder: 1}, EffectComponent{ZOrder: 9}) != 9 {
		t.Fatal("component z-order not preferred")
	}
}

func TestLoudEffectSpecLayering(t *testing.T) {
	spec := compiledEffectSpecs[effectLoud]
	if spec.Layer != EffectLayerActor || spec.Depth != 0.05 || spec.ZOrder != 10 {
		t.Fatalf("compiled loud layering = %+v", spec)
	}
	if EffectDrawsBeforeEntities(spec, spec.Components[0]) {
		t.Fatal("Crazy Uproar would draw behind NPCs")
	}
}
