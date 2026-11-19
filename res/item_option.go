package res

import "strconv"

// Turoran forge metadata lives in option slots and is not a random option name.
const (
	turoranForgeOption  = 0x7FFE
	turoranForgeOwnerLo = 0x7FFD
	turoranForgeOwnerHi = 0x7FFC
)

var enumVarLuaCandidates = []string{
	"data\\luafiles514\\lua files\\datainfo\\enumvar.lub",
	"data\\lua files\\datainfo\\enumvar.lub",
	"lua files\\datainfo\\enumvar.lub",
}

var randomOptionNameLuaCandidates = []string{
	"data\\luafiles514\\lua files\\datainfo\\addrandomoptionnametable.lub",
	"data\\lua files\\datainfo\\addrandomoptionnametable.lub",
	"lua files\\datainfo\\addrandomoptionnametable.lub",
}

// IsForgeMetaOption reports option indexes that carry forge metadata rather
// than a random option from addrandomoptionnametable.
func IsForgeMetaOption(index int16) bool {
	switch int(index) {
	case turoranForgeOption, turoranForgeOwnerLo, turoranForgeOwnerHi:
		return true
	default:
		return false
	}
}

// RandomOptionName returns the client format string for a random option.
// The string may contain %d for the rolled value.
func (m *Manager) RandomOptionName(index int) (string, bool) {
	if m == nil || index == 0 || IsForgeMetaOption(int16(index)) {
		return "", false
	}
	if !m.randomOptionNamesLoaded {
		m.loadRandomOptionNames()
	}
	name, ok := m.randomOptionNames[index]
	return name, ok && name != ""
}

func (m *Manager) loadRandomOptionNames() {
	m.randomOptionNamesLoaded = true
	m.randomOptionNames = map[int]string{}
	globals := map[string]luaValue{}
	for _, candidates := range [][]string{enumVarLuaCandidates, randomOptionNameLuaCandidates} {
		_, data, ok := m.ReadFirst(candidates)
		if !ok {
			continue
		}
		if err := executeLuaChunk(data, globals); err != nil {
			continue
		}
	}
	table := globals["NameTable_VAR"]
	if table.kind != luaTable {
		return
	}
	for key, value := range table.table {
		id, ok := luaTableIntKey(key)
		if !ok || id == 0 || value.kind != luaString || value.str == "" {
			continue
		}
		m.randomOptionNames[id] = decodeROText([]byte(value.str))
	}
}

func luaTableIntKey(key interface{}) (int, bool) {
	switch key := key.(type) {
	case int:
		return key, true
	case float64:
		return int(key), true
	case string:
		id, err := strconv.Atoi(key)
		if err != nil {
			return 0, false
		}
		return id, true
	default:
		return 0, false
	}
}
