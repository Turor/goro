package res

import "strings"

// HatEffectPlayback maps a server hat-effect id to an effect-table id.
// resource-only entries (no hatEffectID) are omitted; those need a sprite player.
type HatEffectPlayback struct {
	EffectID int
}

func (m *Manager) HatEffectPlaybackID(id int) (int, bool) {
	if m == nil || id == 0 {
		return 0, false
	}
	m.loadHatEffects()
	effectID, ok := m.hatEffectPlayback[id]
	return effectID, ok && effectID != 0
}

func (m *Manager) loadHatEffects() {
	if m == nil || m.hatEffectsLoaded {
		return
	}
	m.hatEffectsLoaded = true
	m.hatEffectPlayback = map[int]int{}
	_, data, ok := m.ReadFirst([]string{
		`data\luafiles514\lua files\hateffectinfo\hateffectinfo.lub`,
		`data\luafiles514\lua files\hateffectinfo\hateffectinfo.lub`,
		`lua files\hateffectinfo\hateffectinfo.lub`,
		`hateffectinfo\hateffectinfo.lub`,
		"hateffectinfo.lub",
	})
	if !ok {
		return
	}
	globals := map[string]luaValue{}
	if err := executeLuaChunk(data, globals); err != nil {
		return
	}
	for id, effectID := range hatEffectPlaybackFromGlobals(globals) {
		m.hatEffectPlayback[id] = effectID
	}
}

func hatEffectPlaybackFromGlobals(globals map[string]luaValue) map[int]int {
	table, ok := globals["hatEffectTable"]
	if !ok || table.kind != luaTable {
		return nil
	}
	out := map[int]int{}
	for key, value := range table.table {
		id, ok := luaMapKeyInt(key)
		if !ok || value.kind != luaTable {
			continue
		}
		effect, ok := value.table["hatEffectID"]
		if !ok || effect.kind != luaNumber || effect.num == 0 {
			continue
		}
		out[id] = int(effect.num)
	}
	return out
}

func luaMapKeyInt(key interface{}) (int, bool) {
	switch key := key.(type) {
	case int:
		return key, key != 0
	case float64:
		return int(key), key != 0
	case string:
		key = strings.TrimSpace(key)
		if key == "" {
			return 0, false
		}
		n := 0
		for _, r := range key {
			if r < '0' || r > '9' {
				return 0, false
			}
			n = n*10 + int(r-'0')
		}
		return n, n != 0
	default:
		return 0, false
	}
}

func (m *Manager) SetHatEffectPlaybackForTest(playback map[int]int) {
	if m == nil {
		return
	}
	m.hatEffectsLoaded = true
	if playback == nil {
		m.hatEffectPlayback = map[int]int{}
		return
	}
	m.hatEffectPlayback = playback
}

func ParseHatEffectPlayback(data []byte) (map[int]int, error) {
	globals := map[string]luaValue{}
	if err := executeLuaChunk(data, globals); err != nil {
		return nil, err
	}
	return hatEffectPlaybackFromGlobals(globals), nil
}
