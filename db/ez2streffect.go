package db

import (
	"path"
	"sort"
	"strings"
	"time"
)

// Ez2STREffect is one named entry from ez2streffect.bson.
type Ez2STREffect struct {
	FilePath       string
	SoundPath      string
	IsToGround     bool
	IsFloor        bool
	PosX           float64
	PosY           float64
	StartDelayTime float64
	IsInfinite     bool
}

// Ez2Table is the named BSON effect table, with decode order preserved so
// mergeEz2Effects matching is stable.
type Ez2Table struct {
	Order   []string
	Entries map[string]Ez2STREffect
}

var ez2SuffixFields = []struct {
	suffix string
	field  string
}{
	{"hit", "hitEffectId"},
	{"hitsub", "hitEffectId"},
	{"target", "hitEffectId"},
	{"cast", "effectId"},
	{"cast_bottom", "effectId"},
	{"single", "effectId"},
	{"bottom", "effectId"},
	{"begin", "effectId"},
	{"start", "effectId"},
	{"end", "effectId"},
	{"loop", "effectId"},
	{"buff", "effectId"},
	{"aura", "effectId"},
	{"main", "effectId"},
	{"sync", "effectId"},
}

var ez2HardcodedFields = map[string]string{
	"broken_limit_buff":  "effectId",
	"abyss_flame_target": "hitEffectId",
}

// ParseEz2STREffectBSON reads official (or fixture) ez2streffect.bson bytes.
func ParseEz2STREffectBSON(data []byte) (Ez2Table, error) {
	docs, err := DecodeBSONDocuments(data)
	if err != nil {
		return Ez2Table{}, err
	}
	if len(docs) == 0 {
		return Ez2Table{Entries: map[string]Ez2STREffect{}}, nil
	}
	root := docs[0]
	order := bsonFieldOrder(root)
	// roBrowser unwraps a single-key root document (a collection wrapper).
	// A single named effect at the root must stay as that effect.
	if len(order) == 1 {
		if inner, ok := root[order[0]].(map[string]any); ok {
			innerOrder := bsonFieldOrder(inner)
			wrapped := len(innerOrder) > 0
			for _, key := range innerOrder {
				if _, ok := inner[key].(map[string]any); !ok {
					wrapped = false
					break
				}
			}
			if wrapped {
				root = inner
				order = innerOrder
			}
		}
	}
	out := Ez2Table{Order: make([]string, 0, len(order)), Entries: map[string]Ez2STREffect{}}
	for _, name := range order {
		raw, ok := root[name]
		if !ok {
			continue
		}
		entry, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		out.Order = append(out.Order, name)
		out.Entries[name] = ez2EntryFromMap(entry)
	}
	return out, nil
}

func ez2EntryFromMap(entry map[string]any) Ez2STREffect {
	return Ez2STREffect{
		FilePath:       bsonString(entry["FilePath"]),
		SoundPath:      bsonString(entry["SoundPath"]),
		IsToGround:     bsonBool(entry["IsToGround"]),
		IsFloor:        bsonBool(entry["IsFloor"]),
		PosX:           bsonFloat(entry["PosX"]),
		PosY:           bsonFloat(entry["PosY"]),
		StartDelayTime: bsonFloat(entry["StartDelayTime"]),
		IsInfinite:     bsonBool(entry["IsInfinite"]),
	}
}

func bsonString(v any) string {
	switch t := v.(type) {
	case string:
		return t
	default:
		return ""
	}
}

func bsonBool(v any) bool {
	switch t := v.(type) {
	case bool:
		return t
	case int32:
		return t != 0
	case int64:
		return t != 0
	case float64:
		return t != 0
	default:
		return false
	}
}

func bsonFloat(v any) float64 {
	switch t := v.(type) {
	case float64:
		return t
	case int32:
		return float64(t)
	case int64:
		return float64(t)
	case int:
		return float64(t)
	default:
		return 0
	}
}

type ez2Meta struct {
	baseName   string
	isSuffixed bool
	pathParts  []string
	soundBase  string
	field      string
}

// MergeEz2Effects applies roBrowserLegacy mergeEz2Effects onto the live
// tables: each BSON name becomes an STR EffectSpec, and matching skill names
// (exact, path segment, sound, then suffix inheritance) append the new ID to
// the corresponding SkillEffectSpec field. skillNames maps SKID-style names
// to skill IDs (compiled names plus client Lua SKID).
func MergeEz2Effects(specs map[int]EffectSpec, skills map[uint16]SkillEffectSpec, table Ez2Table, skillNames map[string]uint16) map[string]int {
	if specs == nil || skills == nil {
		return nil
	}
	fuzzy := buildEz2SkillFuzzy(skillNames)
	meta := make(map[string]ez2Meta, len(table.Entries))
	nameToSkill := map[string]uint16{}
	skillBaseToID := map[string]uint16{}

	for _, name := range table.Order {
		entry := table.Entries[name]
		base, suffixed, field := ez2NameField(name, entry)
		filePath := entry.FilePath
		soundPath := entry.SoundPath
		pathParts := splitEz2Path(filePath)
		soundFile := soundPath
		if i := strings.LastIndexAny(soundPath, `\/`); i >= 0 {
			soundFile = soundPath[i+1:]
		}
		soundBase := strings.TrimSuffix(soundFile, path.Ext(soundFile))
		meta[name] = ez2Meta{baseName: base, isSuffixed: suffixed, pathParts: pathParts, soundBase: soundBase, field: field}

		if id, ok := fuzzy[strings.ToUpper(base)]; ok {
			nameToSkill[name] = id
			continue
		}
		if !suffixed {
			for _, segment := range pathParts {
				if segment == "" {
					continue
				}
				if id, ok := skillBaseToID[segment]; ok {
					nameToSkill[name] = id
					break
				}
				if id, ok := fuzzy[strings.ToUpper(segment)]; ok {
					skillBaseToID[segment] = id
					nameToSkill[name] = id
					break
				}
			}
		}
	}

	nextID := nextEz2EffectID(specs)
	nameIDs := map[string]int{}
	for _, name := range table.Order {
		entry := table.Entries[name]
		info := meta[name]
		skillID, mapped := nameToSkill[name]
		if !mapped {
			for _, segment := range info.pathParts {
				if segment == "" {
					continue
				}
				if id, ok := skillBaseToID[segment]; ok {
					skillID, mapped = id, true
					break
				}
				if id, ok := fuzzy[strings.ToUpper(segment)]; ok {
					skillID, mapped = id, true
					break
				}
			}
		}
		if !mapped && info.soundBase != "" {
			if id, ok := fuzzy[strings.ToUpper(info.soundBase)]; ok {
				skillID, mapped = id, true
			}
		}
		if !mapped && info.isSuffixed {
			if id, ok := nameToSkill[info.baseName]; ok {
				skillID, mapped = id, true
			}
		}
		if mapped {
			nameToSkill[name] = skillID
		}

		effectID := nextID
		nextID++
		nameIDs[name] = effectID
		specs[effectID] = ez2EffectSpec(entry)

		if mapped {
			spec := skills[skillID]
			spec = appendEz2SkillField(spec, info.field, effectID)
			skills[skillID] = spec
		}
	}
	return nameIDs
}

func nextEz2EffectID(specs map[int]EffectSpec) int {
	next := Ez2EffectIDBase
	for id := range specs {
		if id >= next {
			next = id + 1
		}
	}
	return next
}

func ez2NameField(name string, entry Ez2STREffect) (base string, suffixed bool, field string) {
	field = "effectId"
	if entry.IsToGround || entry.IsFloor {
		field = "groundEffectId"
	}
	base = name
	for _, s := range ez2SuffixFields {
		if strings.HasSuffix(name, "_"+s.suffix) {
			base = name[:len(name)-len(s.suffix)-1]
			suffixed = true
			field = s.field
			break
		}
	}
	if mapped, ok := ez2HardcodedFields[name]; ok {
		field = mapped
	}
	return base, suffixed, field
}

func splitEz2Path(filePath string) []string {
	normalized := strings.ReplaceAll(filePath, "/", "\\")
	parts := strings.Split(normalized, "\\")
	out := parts[:0]
	for _, p := range parts {
		p = strings.TrimSuffix(p, ".str")
		p = strings.TrimSuffix(p, ".STR")
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func buildEz2SkillFuzzy(skillNames map[string]uint16) map[string]uint16 {
	names := make([]string, 0, len(skillNames))
	for name := range skillNames {
		names = append(names, name)
	}
	sort.Strings(names)
	fuzzy := map[string]uint16{}
	for _, name := range names {
		id := skillNames[name]
		upper := strings.ToUpper(name)
		fuzzy[upper] = id
	}
	for _, name := range names {
		id := skillNames[name]
		upper := strings.ToUpper(name)
		for i := 0; i < len(upper); i++ {
			if upper[i] != '_' {
				continue
			}
			suffix := upper[i+1:]
			if suffix != "" {
				if _, exists := fuzzy[suffix]; !exists {
					fuzzy[suffix] = id
				}
			}
		}
	}
	return fuzzy
}

func ez2EffectSpec(entry Ez2STREffect) EffectSpec {
	file := strings.ReplaceAll(entry.FilePath, "\\", "/")
	file = strings.TrimSuffix(file, ".str")
	file = strings.TrimSuffix(file, ".STR")
	file = strings.TrimPrefix(file, "data/texture/effect/")
	file = strings.TrimPrefix(file, "data/texture/")
	texturePath := ""
	if i := strings.LastIndex(entry.FilePath, "\\"); i >= 0 {
		texturePath = entry.FilePath[:i+1]
	} else if i := strings.LastIndex(entry.FilePath, "/"); i >= 0 {
		texturePath = entry.FilePath[:i+1]
	}
	wav := strings.ReplaceAll(entry.SoundPath, "\\", "/")
	wav = strings.TrimSuffix(wav, ".wav")
	wav = strings.TrimSuffix(wav, ".WAV")
	component := EffectComponent{
		Kind:           EffectComponentSTR,
		STRFile:        file,
		TexturePath:    texturePath,
		RenderBefore:   !entry.IsFloor,
		PosX:           entry.PosX,
		PosY:           entry.PosY,
		Delay:          time.Duration(entry.StartDelayTime * float64(time.Second)),
		Repeat:         entry.IsInfinite,
		AttachedEntity: true,
	}
	if entry.IsFloor {
		component.Layer = EffectLayerActor
	} else {
		component.Layer = EffectLayerGround
		component.RenderBefore = true
	}
	spec := EffectSpec{
		Components: []EffectComponent{component},
		Layer:      component.Layer,
	}
	if wav != "" {
		spec.SFX = []string{wav}
	}
	return spec
}

func appendEz2SkillField(spec SkillEffectSpec, field string, effectID int) SkillEffectSpec {
	switch field {
	case "hitEffectId":
		spec.HitEffectIDs = appendUniqueInt(spec.HitEffectIDs, effectID)
	case "groundEffectId":
		spec.GroundEffectIDs = appendUniqueInt(spec.GroundEffectIDs, effectID)
	default:
		spec.EffectIDs = appendUniqueInt(spec.EffectIDs, effectID)
	}
	return spec
}

func appendUniqueInt(ids []int, id int) []int {
	for _, existing := range ids {
		if existing == id {
			return ids
		}
	}
	return append(ids, id)
}
