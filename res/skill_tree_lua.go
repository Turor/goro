package res

import (
	"fmt"
	"path"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/kivutar/goro/db"
	"github.com/kivutar/goro/glog"
)

var jobInheritLuaCandidates = []string{
	"data\\luafiles514\\lua files\\skillinfoz\\jobinheritlist.lub",
	"data\\lua files\\skillinfoz\\jobinheritlist.lub",
	"lua files\\skillinfoz\\jobinheritlist.lub",
}

// skillTreeRetry spaces out retries when the tables could not be read yet,
// for example while the asset server is unreachable.
const skillTreeRetry = 30 * time.Second

// maxLoggedSkillTreeWarnings keeps a bad data push from flooding the log.
const maxLoggedSkillTreeWarnings = 40

// EnsureSkillTree loads the skill tree from the client Lua tables once and
// installs it with db.SetSkillTreeData, so db.SkillTreeSkillGroups and the
// requirement lookups use it. It reports whether client data is installed;
// without it the built-in Go tables stay in use. Call it from the game thread.
func (m *Manager) EnsureSkillTree() bool {
	if m == nil {
		return false
	}
	if m.skillTreeDataLoaded {
		return m.skillTreeData != nil
	}
	if !m.skillTreeRetryAt.IsZero() && time.Now().Before(m.skillTreeRetryAt) {
		return false
	}
	start := time.Now()
	data, warnings := m.buildSkillTreeData()
	m.skillTreeData, m.skillTreeWarnings = data, warnings
	if data == nil && m.remote != nil {
		// Streamed tables may simply not be reachable yet: try again later.
		m.skillTreeRetryAt = time.Now().Add(skillTreeRetry)
		m.resetSkillDataFlags()
		return false
	}
	m.skillTreeDataLoaded = true
	m.skillTreeRetryAt = time.Time{}
	if data == nil {
		glog.Infof("skill tree: client tables not found, using built-in trees")
		return false
	}
	db.SetSkillTreeData(data)
	glog.Infof("skill tree: loaded %d jobs from client tables in %s, %d warnings",
		len(data.Jobs), time.Since(start).Round(time.Millisecond), len(warnings))
	for i, warning := range warnings {
		if i == maxLoggedSkillTreeWarnings {
			glog.Warnf("skill tree: %d more warnings not shown", len(warnings)-i)
			break
		}
		glog.Warnf("skill tree: %s", warning)
	}
	return true
}

// SkillTreeWarnings returns the validation warnings of the last load.
func (m *Manager) SkillTreeWarnings() []string {
	if m == nil {
		return nil
	}
	return append([]string(nil), m.skillTreeWarnings...)
}

// ReloadSkillData drops every loaded skill table (names, descriptions, max
// levels, requirements, tree positions, job inheritance) and loads them
// again, then replaces the installed tree so open skill windows rebuild.
// Call it from the game thread; other goroutines use QueueDataInvalidation.
func (m *Manager) ReloadSkillData() bool {
	if m == nil {
		return false
	}
	m.resetSkillDataFlags()
	m.skillTreeDataLoaded = false
	m.skillTreeRetryAt = time.Time{}
	if m.EnsureSkillTree() {
		return true
	}
	// The tables are gone or unreadable: fall back to the built-in trees.
	if db.ActiveSkillTreeData() != nil {
		db.SetSkillTreeData(nil)
	}
	return false
}

func (m *Manager) resetSkillDataFlags() {
	m.skillResourceNamesLoaded = false
	m.skillInfoListLoaded = false
	m.skillMetadataLoaded = false
	m.skillMaxLevelsLoaded = false
	m.skillTreePositionsLoaded = false
	m.jobIDLoaded = false
	m.jobIDs = luaValue{}
}

// skillDataStems are the file name stems the skill tables load from,
// including the Turoran overlay names (skillinfolist.turoran13.lub).
var skillDataStems = []string{
	"skillid", "skillinfolist", "skilltreeview", "skilldescript", "jobinheritlist",
	"jobidentity", "skillnametable", "skilldesctable", "leveluseskillspamount",
}

// IsSkillDataFile reports whether a client or server path is one of the
// tables ReloadSkillData rebuilds.
func IsSkillDataFile(name string) bool {
	base := strings.ToLower(path.Base(strings.ReplaceAll(name, "\\", "/")))
	for _, stem := range skillDataStems {
		if base == stem || strings.HasPrefix(base, stem+".") {
			return true
		}
	}
	return false
}

// InvalidateData is the entry point for changed client data, such as a data
// version delta from the asset server or a local edit. It evicts each path
// from the streaming cache, so the next read revalidates it, and live-reloads
// the tables built from those files. It returns whether anything reloaded.
// Call it from the game thread.
func (m *Manager) InvalidateData(paths []string) bool {
	if m == nil || len(paths) == 0 {
		return false
	}
	skill := false
	for _, name := range paths {
		if m.remote != nil {
			m.remote.evict(name)
		}
		if IsSkillDataFile(name) {
			skill = true
		}
	}
	if !skill {
		return false
	}
	glog.Infof("skill tree: reloading after %d changed files", len(paths))
	m.ReloadSkillData()
	return true
}

// dataInvalidationQueue hands changed paths from background goroutines (the
// streaming refresh, a future version poller) to the game thread.
type dataInvalidationQueue struct {
	mu    sync.Mutex
	paths []string
}

// QueueDataInvalidation records changed paths from any goroutine; the game
// thread applies them with ApplyDataInvalidations.
func (m *Manager) QueueDataInvalidation(paths ...string) {
	if m == nil || len(paths) == 0 {
		return
	}
	m.invalidations.mu.Lock()
	m.invalidations.paths = append(m.invalidations.paths, paths...)
	m.invalidations.mu.Unlock()
}

// ApplyDataInvalidations runs queued invalidations on the game thread. It is
// cheap when nothing is queued.
func (m *Manager) ApplyDataInvalidations() bool {
	if m == nil {
		return false
	}
	m.invalidations.mu.Lock()
	paths := m.invalidations.paths
	m.invalidations.paths = nil
	m.invalidations.mu.Unlock()
	return m.InvalidateData(paths)
}

// SkillDataLocalFiles returns the loose files on disk the skill tables come
// from, for a development watcher. Archive and streamed files are skipped.
func (m *Manager) SkillDataLocalFiles() []string {
	if m == nil {
		return nil
	}
	var out []string
	for _, candidates := range [][]string{
		skillIDLuaCandidates, skillInfoListLuaCandidates, skillTreeViewLuaCandidates,
		skillDescriptLuaCandidates, jobInheritLuaCandidates, jobIdentityLuaCandidates,
	} {
		if found, ok := m.FindFirst(candidates); ok {
			out = append(out, found)
		}
	}
	return out
}

func (m *Manager) buildSkillTreeData() (*db.SkillTreeData, []string) {
	m.loadSkillTreePositions()
	if len(m.skillTreePositions) == 0 {
		return nil, nil
	}
	m.loadSkillInfoList()
	globals := m.skillIDGlobals()
	inherit, same := luaValue{}, luaValue{}
	var inheritSource []byte
	if _, data, ok := m.ReadFirst(jobInheritLuaCandidates); ok {
		inheritSource = data
		if err := executeLuaChunk(data, globals); err == nil {
			inherit, same = globals["JOB_INHERIT_LIST"], globals["JOB_INHERIT_LIST2"]
		}
	}
	data := &db.SkillTreeData{
		Jobs:            make(map[int]*db.SkillTreeJob),
		SameTree:        make(map[int]int),
		Requirements:    make(map[uint16][]db.SkillRequirement),
		JobRequirements: make(map[uint16]map[int][]db.SkillRequirement),
		HasRequirements: make(map[uint16]bool),
		MaxLevels:       make(map[uint16]int),
	}
	job := func(id int) *db.SkillTreeJob {
		entry := data.Jobs[id]
		if entry == nil {
			entry = &db.SkillTreeJob{Parent: -1}
			data.Jobs[id] = entry
		}
		return entry
	}
	job(db.JobNovice)
	if inherit.kind == luaTable {
		for rawJob, parent := range inherit.table {
			id, ok := rawJob.(int)
			if !ok || id < 0 || parent.kind != luaNumber {
				continue
			}
			if p := int(parent.num); p != id {
				job(id).Parent = p
			}
		}
	}
	if same.kind == luaTable {
		for rawJob, target := range same.table {
			if id, ok := rawJob.(int); ok && id >= 0 && target.kind == luaNumber {
				data.SameTree[id] = int(target.num)
			}
		}
	}
	for id, positions := range m.skillTreePositions {
		entry := job(id)
		entry.Positions = make(map[uint16]int, len(positions))
		for skillID, position := range positions {
			if skillID > 0 && skillID <= 0xffff {
				entry.Positions[uint16(skillID)] = position
			}
		}
	}
	for id, reqs := range m.skillInfoReqs {
		if id <= 0 || id > 0xffff {
			continue
		}
		data.HasRequirements[uint16(id)] = true
		data.Requirements[uint16(id)] = reqs
	}
	for id, byJob := range m.skillInfoJobReqs {
		if id > 0 && id <= 0xffff {
			data.JobRequirements[uint16(id)] = byJob
		}
	}
	for id, level := range m.skillInfoMaxLevels {
		if id > 0 && id <= 0xffff {
			data.MaxLevels[uint16(id)] = level
		}
	}
	warnings := append([]string(nil), m.skillInfoWarnings...)
	warnings = append(warnings, m.unknownLuaNames("skilltreeview", m.skillTreeViewSource)...)
	warnings = append(warnings, m.unknownLuaNames("skillinfolist", m.skillInfoSource)...)
	warnings = append(warnings, m.unknownLuaNames("jobinheritlist", inheritSource)...)
	warnings = append(warnings, m.validateSkillTree(data)...)
	return data, warnings
}

var luaConstantRef = regexp.MustCompile(`\b(SKID|JOBID)\.([A-Za-z0-9_]+)`)

// unknownLuaNames lists SKID.X and JOBID.X names a source table uses that
// skillid.lub or jobidentity.lub do not define; those entries are dropped.
// Compiled tables are skipped: their names are not readable as text.
func (m *Manager) unknownLuaNames(file string, source []byte) []string {
	if len(source) == 0 || isLua51Bytecode(source) {
		return nil
	}
	globals := m.skillIDGlobals()
	counts := make(map[string]int)
	for _, match := range luaConstantRef.FindAllSubmatch(source, -1) {
		table, name := string(match[1]), string(match[2])
		if t := globals[table]; t.kind == luaTable {
			if v := t.table[name]; v.kind == luaNumber {
				continue
			}
		}
		counts[table+"."+name]++
	}
	out := make([]string, 0, len(counts))
	for name, n := range counts {
		out = append(out, fmt.Sprintf("%s: unknown %s (%d references dropped)", file, name, n))
	}
	sort.Strings(out)
	return out
}

// validateSkillTree checks the loaded tree: skills without SKILL_INFO_LIST
// entries, parents the inheritance list does not define, inheritance cycles,
// two skills in one grid cell of a tab, prerequisites outside the job's tree
// or above their max level, and prerequisite cycles.
func (m *Manager) validateSkillTree(data *db.SkillTreeData) []string {
	var out []string
	jobs := make([]int, 0, len(data.Jobs))
	for id := range data.Jobs {
		jobs = append(jobs, id)
	}
	sort.Ints(jobs)
	jobLabel := m.jobLabels()
	label := func(id int) string {
		if name, ok := jobLabel[id]; ok {
			return fmt.Sprintf("%s(%d)", name, id)
		}
		return fmt.Sprintf("job %d", id)
	}

	// Inheritance: unknown parents and cycles.
	for _, id := range jobs {
		entry := data.Jobs[id]
		if entry.Parent >= 0 {
			if _, ok := data.Jobs[entry.Parent]; !ok {
				out = append(out, fmt.Sprintf("%s inherits from %s, which has no tree entry", label(id), label(entry.Parent)))
			}
		}
		seen := map[int]bool{}
		for cur := id; cur >= 0; {
			if seen[cur] {
				out = append(out, fmt.Sprintf("%s: inheritance cycle through %s", label(id), label(cur)))
				break
			}
			seen[cur] = true
			next := data.Jobs[cur]
			if next == nil {
				break
			}
			cur = next.Parent
		}
	}

	// Skills in tree views without skill info.
	noInfo := map[uint16][]int{}
	for _, id := range jobs {
		for skillID := range data.Jobs[id].Positions {
			if !data.HasRequirements[skillID] {
				noInfo[skillID] = append(noInfo[skillID], id)
			}
		}
	}
	for _, skillID := range sortedSkillIDs(noInfo) {
		out = append(out, fmt.Sprintf("skill %s is in the tree of %s but has no SKILL_INFO_LIST entry",
			m.skillLabel(int(skillID)), label(noInfo[skillID][0])))
	}

	reqCycles := map[string]bool{}
	dupCells := map[string]bool{}
	for _, id := range jobs {
		if len(data.PositionsFor(id)) == 0 {
			continue
		}
		inTree := map[uint16]bool{}
		for tab := 1; tab <= 4; tab++ {
			type cell struct {
				skill uint16
				job   int
			}
			cells := map[int]cell{}
			for _, lineageJob := range data.Lineage(id) {
				if db.SkillTreeTab(lineageJob) != tab {
					continue
				}
				positions := data.PositionsFor(lineageJob)
				for _, skillID := range sortedSkillIDs(positions) {
					position := positions[skillID]
					inTree[skillID] = true
					other, taken := cells[position]
					if !taken {
						cells[position] = cell{skill: skillID, job: lineageJob}
						continue
					}
					key := fmt.Sprintf("%d/%d/%d/%d", other.job, lineageJob, position, skillID)
					if other.skill != skillID && !dupCells[key] {
						dupCells[key] = true
						out = append(out, fmt.Sprintf("tab %d cell %d: %s of %s overlaps %s of %s (seen in the %s tree)",
							tab, position, m.skillLabel(int(skillID)), label(lineageJob),
							m.skillLabel(int(other.skill)), label(other.job), label(id)))
					}
				}
			}
		}
		for _, skillID := range sortedSkillSet(inTree) {
			reqs, _ := data.RequirementsFor(id, skillID)
			for _, req := range reqs {
				if !inTree[req.SkillID] {
					out = append(out, fmt.Sprintf("%s: %s requires %s, which is not in its tree",
						label(id), m.skillLabel(int(skillID)), m.skillLabel(int(req.SkillID))))
				} else if maxLevel := data.MaxLevels[req.SkillID]; maxLevel > 0 && req.Level > maxLevel {
					out = append(out, fmt.Sprintf("%s: %s requires %s level %d, above its max level %d",
						label(id), m.skillLabel(int(skillID)), m.skillLabel(int(req.SkillID)), req.Level, maxLevel))
				}
			}
		}
		for _, cycle := range requirementCycles(data, id, inTree) {
			names := make([]string, len(cycle))
			for i, skillID := range cycle {
				names[i] = m.skillLabel(int(skillID))
			}
			key := strings.Join(names, " -> ")
			if !reqCycles[key] {
				reqCycles[key] = true
				out = append(out, fmt.Sprintf("%s: prerequisite cycle %s", label(id), key))
			}
		}
	}
	return out
}

// requirementCycles finds prerequisite cycles among the skills of one job,
// each reported once starting from its smallest skill ID.
func requirementCycles(data *db.SkillTreeData, job int, inTree map[uint16]bool) [][]uint16 {
	const (
		unvisited = iota
		active
		done
	)
	state := map[uint16]int{}
	var stack []uint16
	var cycles [][]uint16
	var visit func(uint16)
	visit = func(skillID uint16) {
		state[skillID] = active
		stack = append(stack, skillID)
		reqs, _ := data.RequirementsFor(job, skillID)
		for _, req := range reqs {
			if !inTree[req.SkillID] {
				continue
			}
			switch state[req.SkillID] {
			case unvisited:
				visit(req.SkillID)
			case active:
				start := len(stack) - 1
				for start > 0 && stack[start] != req.SkillID {
					start--
				}
				cycles = append(cycles, rotateCycle(append([]uint16(nil), stack[start:]...)))
			}
		}
		stack = stack[:len(stack)-1]
		state[skillID] = done
	}
	for _, skillID := range sortedSkillSet(inTree) {
		if state[skillID] == unvisited {
			visit(skillID)
		}
	}
	return cycles
}

func rotateCycle(cycle []uint16) []uint16 {
	least := 0
	for i, id := range cycle {
		if id < cycle[least] {
			least = i
		}
	}
	return append(cycle[least:], cycle[:least]...)
}

func sortedSkillIDs[V any](set map[uint16]V) []uint16 {
	out := make([]uint16, 0, len(set))
	for id := range set {
		out = append(out, id)
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

func sortedSkillSet(set map[uint16]bool) []uint16 { return sortedSkillIDs(set) }

// skillLabel names a skill by its SKID name and ID for warnings.
func (m *Manager) skillLabel(id int) string {
	if name := m.skillResourceNames[id]; name != "" {
		return fmt.Sprintf("%s(%d)", name, id)
	}
	return fmt.Sprintf("skill %d", id)
}

// jobLabels names jobs by their JOBID constants for warnings.
func (m *Manager) jobLabels() map[int]string {
	out := map[int]string{}
	if jobs := m.jobIDTable(); jobs.kind == luaTable {
		for rawName, value := range jobs.table {
			name, ok := rawName.(string)
			if !ok || value.kind != luaNumber {
				continue
			}
			id := int(value.num)
			if prev, ok := out[id]; !ok || name < prev {
				out[id] = name
			}
		}
	}
	return out
}
