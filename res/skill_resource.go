package res

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/kivutar/goro/db"
	glua "github.com/yuin/gopher-lua"
)

var skillIDLuaCandidates = []string{
	"data\\luafiles514\\lua files\\skillinfoz\\skillid.lub",
	"data\\lua files\\skillinfoz\\skillid.lub",
	"lua files\\skillinfoz\\skillid.lub",
	"data\\luafiles514\\lua files\\skillinfo\\skillid.lub",
	"data\\lua files\\skillinfo\\skillid.lub",
	"lua files\\skillinfo\\skillid.lub",
}

var skillTreeViewLuaCandidates = []string{
	"data\\luafiles514\\lua files\\skillinfoz\\skilltreeview.lub",
	"data\\lua files\\skillinfoz\\skilltreeview.lub",
	"lua files\\skillinfoz\\skilltreeview.lub",
	"data\\luafiles514\\lua files\\skillinfo\\skilltreeview.lub",
	"data\\lua files\\skillinfo\\skilltreeview.lub",
	"lua files\\skillinfo\\skilltreeview.lub",
}

var skillInfoListLuaCandidates = []string{
	"data\\luafiles514\\lua files\\skillinfoz\\skillinfolist.lub",
	"data\\lua files\\skillinfoz\\skillinfolist.lub",
	"lua files\\skillinfoz\\skillinfolist.lub",
}

var skillDescriptLuaCandidates = []string{
	"data\\luafiles514\\lua files\\skillinfoz\\skilldescript.lub",
	"data\\lua files\\skillinfoz\\skilldescript.lub",
	"lua files\\skillinfoz\\skilldescript.lub",
	"System\\skilldescript.lub",
}

var jobIdentityLuaCandidates = []string{
	"data\\luafiles514\\lua files\\datainfo\\jobidentity.lub",
	"data\\lua files\\datainfo\\jobidentity.lub",
	"lua files\\datainfo\\jobidentity.lub",
}

var skillSPAmountCandidates = []string{
	"leveluseskillspamount.txt",
	"data\\leveluseskillspamount.txt",
	"data/leveluseskillspamount.txt",
}

func (m *Manager) SkillResourceName(skillID int) (string, bool) {
	if skillID <= 0 {
		return "", false
	}
	m.loadSkillResourceNames()
	name, ok := m.skillResourceNames[skillID]
	return name, ok && name != ""
}

func (m *Manager) SkillDisplayName(skillID int) (string, bool) {
	if skillID <= 0 {
		return "", false
	}
	m.loadSkillMetadata()
	name, ok := m.skillDisplayNames[skillID]
	return name, ok && name != ""
}

func (m *Manager) SkillDescription(skillID int) ([]string, bool) {
	if skillID <= 0 {
		return nil, false
	}
	m.loadSkillMetadata()
	lines, ok := m.skillDescriptions[skillID]
	if !ok || len(lines) == 0 {
		return nil, false
	}
	return append([]string(nil), lines...), true
}

func (m *Manager) SkillMaxLevel(skillID int) (int, bool) {
	if skillID <= 0 {
		return 0, false
	}
	m.loadSkillMaxLevels()
	level, ok := m.skillMaxLevels[skillID]
	return level, ok && level > 0
}

// SkillTreePositions returns the read-only client-defined grid position for
// each skill belonging to job. Positions are zero-based in seven columns.
func (m *Manager) SkillTreePositions(job int) (map[int]int, bool) {
	if job < 0 {
		return nil, false
	}
	m.loadSkillTreePositions()
	if positions := m.skillTreePositions[job]; len(positions) > 0 {
		return positions, true
	}
	keys := []string{strconv.Itoa(job), strings.ToUpper(db.JobDisplayName(job))}
	switch job {
	case db.JobStar:
		keys = append(keys, "STAR", "TAEKWON MASTER")
	case db.JobLinker:
		keys = append(keys, "LINKER")
	}
	for _, key := range keys {
		if positions := m.skillTreeTextPositions[key]; len(positions) > 0 {
			return positions, true
		}
	}
	return nil, false
}

func (m *Manager) loadSkillResourceNames() {
	if m.skillResourceNamesLoaded {
		return
	}
	m.skillResourceNamesLoaded = true
	m.skillResourceNames = make(map[int]string, len(db.SkillResourceName))
	m.skillIDNames = make(map[string]int, len(db.SkillResourceName))
	for id, name := range db.SkillResourceName {
		m.skillResourceNames[int(id)] = name
		m.skillIDNames[name] = int(id)
	}
	globals := make(map[string]luaValue)
	_, data, ok := m.ReadFirst(skillIDLuaCandidates)
	if !ok {
		return
	}
	if err := executeLuaChunk(data, globals); err != nil {
		return
	}
	table := globals["SKID"]
	if table.kind != luaTable {
		return
	}
	fileIDs := make(map[string]int, len(table.table))
	for key, value := range table.table {
		name, ok := key.(string)
		if !ok || value.kind != luaNumber || name == "" || value.num <= 0 {
			continue
		}
		fileIDs[name] = int(value.num)
	}
	// The client table wins: drop built-in names it assigns to another ID,
	// so the SKID table is never ambiguous.
	for id, name := range m.skillResourceNames {
		if fileID, ok := fileIDs[name]; ok && fileID != id {
			delete(m.skillResourceNames, id)
		}
	}
	// Several names may share one ID (renamed skills keep their old name as
	// an alias). Every name stays resolvable in SKID; the display name per ID
	// is picked deterministically, the smallest name.
	names := make([]string, 0, len(fileIDs))
	for name := range fileIDs {
		names = append(names, name)
	}
	sort.Strings(names)
	fromFile := make(map[int]bool, len(names))
	for _, name := range names {
		id := fileIDs[name]
		m.skillIDNames[name] = id
		if !fromFile[id] {
			m.skillResourceNames[id] = name
			fromFile[id] = true
		}
	}
}

func (m *Manager) loadSkillMetadata() {
	if m.skillMetadataLoaded {
		return
	}
	m.skillMetadataLoaded = true
	m.loadSkillResourceNames()
	m.skillDisplayNames = make(map[int]string)
	m.skillDescriptions = make(map[int][]string)
	nameToID := m.skillNameToID()
	if _, data, ok := m.ReadFirst(skillDataTableCandidates("skillnametable.txt")); ok {
		for id, name := range parseSkillNameTable(data, nameToID) {
			m.skillDisplayNames[id] = name
		}
	}
	if _, data, ok := m.ReadFirst(skillDataTableCandidates("skilldesctable.txt")); ok {
		names, descriptions := parseSkillDescriptionTable(data, nameToID)
		for id, name := range names {
			if m.skillDisplayNames[id] == "" {
				m.skillDisplayNames[id] = name
			}
		}
		for id, lines := range descriptions {
			m.skillDescriptions[id] = lines
		}
	}
	// Renewal clients replaced the text tables with SKILL_INFO_LIST and
	// SKILL_DESCRIPT; when present they are newer, so they win.
	m.loadSkillInfoList()
	for id, name := range m.skillInfoNames {
		m.skillDisplayNames[id] = name
	}
	if _, data, ok := m.ReadFirst(skillDescriptLuaCandidates); ok {
		globals := m.skillIDGlobals()
		if executeLuaChunk(data, globals) == nil {
			if table := globals["SKILL_DESCRIPT"]; table.kind == luaTable {
				for rawID, value := range table.table {
					id, ok := rawID.(int)
					if !ok || id <= 0 {
						continue
					}
					lines := luaStringList(value)
					for i, line := range lines {
						lines[i] = decodeROText([]byte(line))
					}
					if len(lines) == 0 {
						continue
					}
					m.skillDescriptions[id] = lines
					if m.skillDisplayNames[id] == "" {
						m.skillDisplayNames[id] = lines[0]
					}
				}
			}
		}
	}
}

// loadSkillInfoList reads SKILL_INFO_LIST (skillinfolist.lub): display name
// and max level per skill, keyed by SKID.
func (m *Manager) loadSkillInfoList() {
	if m.skillInfoListLoaded {
		return
	}
	m.skillInfoListLoaded = true
	m.skillInfoNames = make(map[int]string)
	m.skillInfoMaxLevels = make(map[int]int)
	m.skillInfoReqs = make(map[int][]db.SkillRequirement)
	m.skillInfoJobReqs = make(map[int]map[int][]db.SkillRequirement)
	m.skillInfoWarnings = nil
	m.skillInfoSource = nil
	_, data, ok := m.ReadFirst(skillInfoListLuaCandidates)
	if !ok {
		return
	}
	m.skillInfoSource = data
	globals := m.skillIDGlobals()
	if executeLuaChunk(data, globals) != nil {
		return
	}
	table := globals["SKILL_INFO_LIST"]
	if table.kind != luaTable {
		return
	}
	for rawID, entry := range table.table {
		id, ok := rawID.(int)
		if !ok || id <= 0 || entry.kind != luaTable {
			continue
		}
		if name := luaField(entry, "SkillName"); name.kind == luaString && name.str != "" {
			m.skillInfoNames[id] = decodeROText([]byte(name.str))
		}
		if level := luaField(entry, "MaxLv"); level.kind == luaNumber && level.num > 0 {
			m.skillInfoMaxLevels[id] = int(level.num)
		}
		m.skillInfoReqs[id] = m.parseSkillRequirements(id, -1, luaField(entry, "_NeedSkillList"))
		if byJob := luaField(entry, "NeedSkillList"); byJob.kind == luaTable {
			for rawJob, list := range byJob.table {
				job, ok := rawJob.(int)
				if !ok || job < 0 {
					continue
				}
				if m.skillInfoJobReqs[id] == nil {
					m.skillInfoJobReqs[id] = make(map[int][]db.SkillRequirement)
				}
				m.skillInfoJobReqs[id][job] = m.parseSkillRequirements(id, job, list)
			}
		}
	}
}

// parseSkillRequirements reads a { { SKID.X, level }, ... } list. An entry
// whose skill did not resolve (a SKID name skillid.lub lacks) is dropped with
// a warning; the source scan in validateSkillTree names it.
func (m *Manager) parseSkillRequirements(skillID, job int, list luaValue) []db.SkillRequirement {
	if list.kind != luaTable {
		return nil
	}
	var out []db.SkillRequirement
	for i := 1; i <= len(list.table)+1; i++ {
		entry, ok := list.table[i]
		if !ok {
			continue
		}
		var required, level luaValue
		if entry.kind == luaTable {
			required, level = entry.table[1], entry.table[2]
		}
		if required.kind != luaNumber || required.num <= 0 || required.num > 0xffff {
			where := "_NeedSkillList"
			if job >= 0 {
				where = fmt.Sprintf("NeedSkillList[%d]", job)
			}
			m.skillInfoWarnings = append(m.skillInfoWarnings,
				fmt.Sprintf("skill %s: %s entry %d names an unknown skill", m.skillLabel(skillID), where, i))
			continue
		}
		lv := 1
		if level.kind == luaNumber && level.num > 0 {
			lv = int(level.num)
		}
		out = append(out, db.SkillRequirement{SkillID: uint16(required.num), Level: lv})
	}
	return out
}

// skillIDGlobals seeds a Lua environment with SKID from the loaded skill IDs
// and JOBID from jobidentity.lub, which skillinfolist and skilltreeview key on.
func (m *Manager) skillIDGlobals() map[string]luaValue {
	m.loadSkillResourceNames()
	skillIDs := make(map[interface{}]luaValue, len(m.skillIDNames))
	for name, id := range m.skillIDNames {
		if id > 0 && name != "" {
			skillIDs[name] = luaValue{kind: luaNumber, num: float64(id)}
		}
	}
	// roBrowserLegacy seeds WS_ARMORREFINE, which kRO skillid tables lack.
	for name, id := range compatSkillIDs {
		if _, ok := skillIDs[name]; !ok {
			skillIDs[name] = luaValue{kind: luaNumber, num: float64(id)}
		}
	}
	globals := map[string]luaValue{"SKID": {kind: luaTable, table: skillIDs}}
	if jobs := m.jobIDTable(); jobs.kind == luaTable {
		globals["JOBID"] = jobs
	}
	return globals
}

var compatSkillIDs = map[string]int{"WS_ARMORREFINE": 709}

// compatJobIDs are mounted-job constants skill tables reference (NeedSkillList
// for Crusader on a Grand Peco) that newer jobidentity tables dropped.
// roBrowserLegacy derives them from its own job list.
var compatJobIDs = map[string]int{
	"JT_KNIGHT2": 13, "JT_KNIGHT2_H": 4014, "JT_KNIGHT2_B": 4036,
	"JT_CRUSADER2": 21, "JT_CRUSADER2_H": 4022, "JT_CRUSADER2_B": 4044,
}

// jobIDTable returns the job constants from jobidentity.lub. Older clients
// name the table JTtbl; skill tables expect JOBID.
func (m *Manager) jobIDTable() luaValue {
	if m.jobIDLoaded {
		return m.jobIDs
	}
	m.jobIDLoaded = true
	globals := make(map[string]luaValue)
	_, jobData, ok := m.ReadFirst(jobIdentityLuaCandidates)
	if !ok || executeLuaChunk(jobData, globals) != nil {
		return m.jobIDs
	}
	m.jobIDs = globals["JOBID"]
	if m.jobIDs.kind != luaTable {
		m.jobIDs = globals["JTtbl"]
	}
	if m.jobIDs.kind == luaTable {
		for name, id := range compatJobIDs {
			if _, ok := m.jobIDs.table[name]; !ok {
				m.jobIDs.table[name] = luaValue{kind: luaNumber, num: float64(id)}
			}
		}
	}
	return m.jobIDs
}

// installJobSkillTab stubs the client's JobSkillTab API: renewal
// skilltreeview.lub calls JobSkillTab.ChangeSkillTabName after the table to
// label the 1st/2nd/3rd/4th tabs, which goro does not use yet.
func installJobSkillTab(state *glua.LState) {
	tab := state.NewTable()
	tab.RawSetString("ChangeSkillTabName", state.NewFunction(func(*glua.LState) int { return 0 }))
	state.SetGlobal("JobSkillTab", tab)
}

func (m *Manager) loadSkillMaxLevels() {
	if m.skillMaxLevelsLoaded {
		return
	}
	m.skillMaxLevelsLoaded = true
	m.loadSkillResourceNames()
	m.skillMaxLevels = make(map[int]int)
	nameToID := m.skillNameToID()
	if _, data, ok := m.ReadFirst(skillSPAmountCandidates); ok {
		for id, level := range parseSkillSPAmountMaxLevels(data, nameToID) {
			m.skillMaxLevels[id] = level
		}
	}
	m.loadSkillInfoList()
	for id, level := range m.skillInfoMaxLevels {
		m.skillMaxLevels[id] = level
	}
}

func (m *Manager) loadSkillTreePositions() {
	if m.skillTreePositionsLoaded {
		return
	}
	m.skillTreePositionsLoaded = true
	m.skillTreePositions = make(map[int]map[int]int)
	m.skillTreeViewSource = nil
	m.loadSkillResourceNames()
	if _, data, ok := m.ReadFirst(skillDataTableCandidates("skilltreeview.txt")); ok {
		m.skillTreeTextPositions = parseSkillTreeTextPositions(data, m.skillNameToID())
	}

	globals := m.skillIDGlobals()
	if globals["JOBID"].kind != luaTable {
		return
	}
	_, treeData, ok := m.ReadFirst(skillTreeViewLuaCandidates)
	if !ok || executeLuaChunkWithSetup(treeData, globals, installJobSkillTab) != nil {
		return
	}
	m.skillTreePositions = parseSkillTreePositions(globals["SKILL_TREEVIEW_FOR_JOB"])
	m.skillTreeViewSource = treeData
}

// Text clients use JOB { SKILL#position#prerequisite positions#max level@ ... }.
// Job keys are names in 2008 and numeric IDs in later text tables.
func parseSkillTreeTextPositions(data []byte, nameToID map[string]int) map[string]map[int]int {
	lines := strings.Split(strings.TrimPrefix(string(data), "\ufeff"), "\n")
	for i, line := range lines {
		lines[i], _, _ = strings.Cut(line, "//")
	}
	out := make(map[string]map[int]int)
	for _, block := range strings.Split(strings.Join(lines, "\n"), "}") {
		job, body, ok := strings.Cut(block, "{")
		job = strings.ToUpper(strings.TrimSpace(job))
		if !ok || job == "" {
			continue
		}
		entries := strings.Split(body, "@")
		for _, entry := range entries[:len(entries)-1] {
			name, fields, ok := strings.Cut(strings.TrimSpace(entry), "#")
			if !ok {
				continue
			}
			id := nameToID[strings.ToLower(strings.TrimSpace(name))]
			positionText, _, _ := strings.Cut(fields, "#")
			position, err := strconv.Atoi(strings.TrimSpace(positionText))
			if id <= 0 || err != nil || position < 0 {
				continue
			}
			if out[job] == nil {
				out[job] = make(map[int]int)
			}
			out[job][id] = position
		}
	}
	return out
}

func parseSkillTreePositions(table luaValue) map[int]map[int]int {
	out := make(map[int]map[int]int)
	if table.kind != luaTable {
		return out
	}
	for rawJob, rawPositions := range table.table {
		job, ok := rawJob.(int)
		if !ok || job < 0 || rawPositions.kind != luaTable {
			continue
		}
		positions := make(map[int]int)
		for rawPosition, rawSkill := range rawPositions.table {
			position, ok := rawPosition.(int)
			if !ok || position < 0 || rawSkill.kind != luaNumber {
				continue
			}
			skillID := int(rawSkill.num)
			if skillID > 0 {
				positions[skillID] = position
			}
		}
		if len(positions) > 0 {
			out[job] = positions
		}
	}
	return out
}

func (m *Manager) skillNameToID() map[string]int {
	out := make(map[string]int, len(m.skillIDNames))
	for name, id := range m.skillIDNames {
		if name == "" {
			continue
		}
		out[strings.ToLower(strings.TrimSpace(name))] = id
	}
	return out
}

func skillDataTableCandidates(fileName string) []string {
	return []string{
		fileName,
		"data\\" + fileName,
		"data/" + fileName,
		"data\\luafiles514\\lua files\\skillinfoz\\" + fileName,
		"data/luafiles514/lua files/skillinfoz/" + fileName,
		"data\\lua files\\skillinfoz\\" + fileName,
		"data/lua files/skillinfoz/" + fileName,
		"lua files\\skillinfoz\\" + fileName,
		"lua files/skillinfoz/" + fileName,
		"data\\luafiles514\\lua files\\skillinfo\\" + fileName,
		"data/luafiles514/lua files/skillinfo/" + fileName,
		"data\\lua files\\skillinfo\\" + fileName,
		"data/lua files/skillinfo/" + fileName,
		"lua files\\skillinfo\\" + fileName,
		"lua files/skillinfo/" + fileName,
	}
}

func parseSkillNameTable(data []byte, nameToID map[string]int) map[int]string {
	out := make(map[int]string)
	for _, rawLine := range strings.Split(string(data), "\n") {
		line := strings.TrimSpace(strings.TrimRight(rawLine, "\r\n"))
		if line == "" || strings.HasPrefix(line, "/") || strings.HasPrefix(line, "#") {
			continue
		}
		tokens := strings.Split(line, "#")
		if len(tokens) < 2 {
			continue
		}
		id, ok := nameToID[strings.ToLower(strings.TrimSpace(tokens[0]))]
		if !ok || id <= 0 {
			continue
		}
		name := normalizeSkillDisplayToken(tokens[1])
		if name != "" {
			out[id] = name
		}
	}
	return out
}

func parseSkillDescriptionTable(data []byte, nameToID map[string]int) (map[int]string, map[int][]string) {
	names := make(map[int]string)
	descriptions := make(map[int][]string)
	currentID := 0
	expectTitle := false
	for _, rawLine := range strings.Split(string(data), "\n") {
		line := strings.TrimRight(rawLine, "\r\n")
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if trimmed == "#" {
			currentID = 0
			expectTitle = false
			continue
		}
		if strings.HasSuffix(trimmed, "#") && len(trimmed) > 1 {
			key := strings.TrimSpace(strings.TrimSuffix(trimmed, "#"))
			currentID = nameToID[strings.ToLower(key)]
			expectTitle = currentID > 0
			continue
		}
		if currentID <= 0 {
			continue
		}
		if expectTitle {
			name := normalizeSkillDisplayToken(trimmed)
			if name != "" {
				names[currentID] = name
			}
			expectTitle = false
			continue
		}
		descriptions[currentID] = append(descriptions[currentID], line)
	}
	return names, descriptions
}

func parseSkillSPAmountMaxLevels(data []byte, nameToID map[string]int) map[int]int {
	out := make(map[int]int)
	currentID := 0
	levelCount := 0
	flush := func() {
		if currentID > 0 && levelCount > 0 {
			out[currentID] = levelCount
		}
		currentID = 0
		levelCount = 0
	}
	for _, rawLine := range strings.Split(string(data), "\n") {
		line := strings.TrimSpace(strings.TrimRight(rawLine, "\r\n"))
		if line == "" || strings.HasPrefix(line, "/") {
			continue
		}
		if line == "@" {
			flush()
			continue
		}
		if strings.HasSuffix(line, "#") {
			token := strings.TrimSpace(strings.TrimSuffix(line, "#"))
			if id, ok := nameToID[strings.ToLower(token)]; ok && id > 0 {
				flush()
				currentID = id
				continue
			}
			if currentID > 0 {
				levelCount++
			}
		}
	}
	flush()
	return out
}

func normalizeSkillDisplayToken(value string) string {
	return strings.ReplaceAll(strings.TrimSpace(value), "_", " ")
}

func SkillIconTextureCandidates(resource string, skillID int) []string {
	resource = strings.TrimSpace(strings.TrimSuffix(resource, ".bmp"))
	seen := make(map[string]struct{})
	out := make([]string, 0, 16)
	add := func(candidate string) {
		if candidate == "" {
			return
		}
		if _, ok := seen[candidate]; ok {
			return
		}
		seen[candidate] = struct{}{}
		out = append(out, candidate)
	}
	addStem := func(stem string) {
		if stem == "" {
			return
		}
		stem = strings.ReplaceAll(stem, "/", "\\")
		lower := strings.ToLower(stem)
		for _, candidateStem := range []string{lower, stem} {
			const uiKorPrefix = "data\\texture\\유저인터페이스\\item\\"
			add(uiKorPrefix + candidateStem + ".bmp")
			add(strings.ReplaceAll(uiKorPrefix, "\\", "/") + candidateStem + ".bmp")
			add("texture\\유저인터페이스\\item\\" + candidateStem + ".bmp")
			add("texture/item/" + candidateStem + ".bmp")
			add("data/texture/item/" + candidateStem + ".bmp")
			add(candidateStem + ".bmp")
		}
	}
	addStem(resource)
	if skillID > 0 {
		addStem(fmt.Sprintf("%d", skillID))
	}
	return out
}
