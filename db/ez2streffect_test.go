package db

import (
	"testing"
)

func TestMergeEz2EffectsMatchesRobrowserPrecedence(t *testing.T) {
	table := fixtureEz2Table()
	specs := map[int]EffectSpec{1: {SFX: []string{"keep"}}}
	skills := map[uint16]SkillEffectSpec{
		SkillMCLoud: {SuccessEffectIDsSelf: []int{effectLoud}},
		SkillSMBash: {HitEffectIDs: []int{1}},
	}
	names := map[string]uint16{
		"MC_LOUD": SkillMCLoud,
		"SM_BASH": SkillSMBash,
	}
	ids := MergeEz2Effects(specs, skills, table, names)

	loudID, ok := ids["MC_LOUD"]
	if !ok {
		t.Fatal("MC_LOUD not assigned")
	}
	if specs[loudID].Components[0].STRFile != "loud" {
		t.Fatalf("MC_LOUD str = %+v", specs[loudID])
	}
	if !containsEffectID(skills[SkillMCLoud].EffectIDs, loudID) {
		t.Fatalf("MC_LOUD effectId not merged: %+v", skills[SkillMCLoud])
	}
	if skills[SkillMCLoud].SuccessEffectIDsSelf[0] != effectLoud {
		t.Fatal("existing SuccessEffectIDsSelf dropped")
	}

	hitID := ids["MC_LOUD_hit"]
	if !containsEffectID(skills[SkillMCLoud].HitEffectIDs, hitID) {
		t.Fatalf("suffix _hit did not map to hitEffectId: %+v", skills[SkillMCLoud])
	}

	pathID := ids["unknown_from_path"]
	if !containsEffectID(skills[SkillSMBash].EffectIDs, pathID) {
		t.Fatalf("path segment SM_BASH not mapped: %+v", skills[SkillSMBash])
	}

	soundID := ids["sound_only"]
	if !containsEffectID(skills[SkillMCLoud].EffectIDs, soundID) {
		t.Fatalf("sound fallback not mapped: %+v", skills[SkillMCLoud])
	}

	floorID := ids["MC_LOUD_bottom"]
	if !containsEffectID(skills[SkillMCLoud].GroundEffectIDs, floorID) && !containsEffectID(skills[SkillMCLoud].EffectIDs, floorID) {
		t.Fatalf("floor/suffix bottom not mapped: %+v", skills[SkillMCLoud])
	}
	if specs[floorID].Layer != EffectLayerActor {
		t.Fatalf("IsFloor layer = %d", specs[floorID].Layer)
	}

	inherited := ids["MC_LOUD_cast"]
	if !containsEffectID(skills[SkillMCLoud].EffectIDs, inherited) {
		t.Fatalf("suffix inheritance failed: %+v ids=%v", skills[SkillMCLoud], ids)
	}

	hardID := ids["abyss_flame_target"]
	if !containsEffectID(skills[SkillSMBash].HitEffectIDs, hardID) {
		t.Fatalf("hardcoded abyss_flame_target not hitEffectId: %+v", skills[SkillSMBash])
	}

	if _, ok := specs[1]; !ok || specs[1].SFX[0] != "keep" {
		t.Fatal("existing spec dropped")
	}
}

func TestParseEz2STREffectBSONRoundTrip(t *testing.T) {
	raw := EncodeBSONDocument(map[string]any{
		"MC_LOUD": map[string]any{
			"FilePath":  `data\texture\effect\loud.str`,
			"SoundPath": `effect\loud.wav`,
			"IsFloor":   false,
			"PosX":      1.5,
			"":          []string{"FilePath", "SoundPath", "IsFloor", "PosX"},
		},
	}, []string{"MC_LOUD"})
	table, err := ParseEz2STREffectBSON(raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(table.Order) != 1 || table.Order[0] != "MC_LOUD" {
		t.Fatalf("order = %v", table.Order)
	}
	entry := table.Entries["MC_LOUD"]
	if entry.FilePath != `data\texture\effect\loud.str` || entry.PosX != 1.5 {
		t.Fatalf("entry = %+v", entry)
	}
}

func TestParseEz2STREffectBSONRejectsCorrupt(t *testing.T) {
	if _, err := ParseEz2STREffectBSON([]byte{1, 2, 3}); err == nil {
		t.Fatal("corrupt bson accepted")
	}
}

func fixtureEz2Table() Ez2Table {
	raw := EncodeBSONDocument(map[string]any{
		"MC_LOUD":            ez2Doc(`data\texture\effect\loud.str`, "", false, false),
		"MC_LOUD_hit":        ez2Doc(`data\texture\effect\loud_hit.str`, "", false, false),
		"unknown_from_path":  ez2Doc(`data\texture\effect\SM_BASH\foo.str`, "", false, false),
		"sound_only":         ez2Doc(`data\texture\effect\other.str`, `effect\MC_LOUD.wav`, false, false),
		"MC_LOUD_bottom":     ez2Doc(`data\texture\effect\loud_bottom.str`, "", true, false),
		"MC_LOUD_cast":       ez2Doc(`data\texture\effect\loud_cast.str`, "", false, false),
		"abyss_flame_target": ez2Doc(`data\texture\effect\SM_BASH\flame.str`, "", false, false),
	}, []string{"MC_LOUD", "MC_LOUD_hit", "unknown_from_path", "sound_only", "MC_LOUD_bottom", "MC_LOUD_cast", "abyss_flame_target"})
	table, err := ParseEz2STREffectBSON(raw)
	if err != nil {
		panic(err)
	}
	return table
}

func ez2Doc(file, sound string, floor, ground bool) map[string]any {
	return map[string]any{
		"FilePath":   file,
		"SoundPath":  sound,
		"IsFloor":    floor,
		"IsToGround": ground,
		"PosX":       0,
		"PosY":       0,
		"IsInfinite": false,
		"":           []string{"FilePath", "SoundPath", "IsFloor", "IsToGround", "PosX", "PosY", "IsInfinite"},
	}
}

func containsEffectID(ids []int, id int) bool {
	for _, v := range ids {
		if v == id {
			return true
		}
	}
	return false
}
