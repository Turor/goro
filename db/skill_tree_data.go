package db

import (
	"sort"
	"sync/atomic"
)

// SkillTreeJob is one job's entry in the client skill-tree tables: its parent
// in JOB_INHERIT_LIST and its own SKILL_TREEVIEW_FOR_JOB grid positions.
type SkillTreeJob struct {
	// Parent is the job this one inherits from, or -1 for a root.
	Parent int
	// Positions maps a skill to its zero-based cell in the seven-column grid.
	Positions map[uint16]int
}

// SkillTreeData is the skill tree loaded from the client Lua tables
// (skilltreeview, jobinheritlist, skillinfolist). When it is installed with
// SetSkillTreeData it takes precedence over the built-in Go tables, which stay
// as the fallback for jobs or skills the data does not cover.
type SkillTreeData struct {
	Jobs map[int]*SkillTreeJob
	// SameTree is JOB_INHERIT_LIST2: a job that shows another job's tree view,
	// such as a transcendent third job using the third job's positions.
	SameTree map[int]int
	// Requirements is _NeedSkillList; JobRequirements is NeedSkillList keyed
	// by job. HasRequirements marks skills SKILL_INFO_LIST defines at all.
	Requirements    map[uint16][]SkillRequirement
	JobRequirements map[uint16]map[int][]SkillRequirement
	HasRequirements map[uint16]bool
	MaxLevels       map[uint16]int
}

var (
	activeSkillTree     atomic.Pointer[SkillTreeData]
	skillTreeGeneration atomic.Uint64
)

// SetSkillTreeData installs data loaded from the client tables, or clears it
// with nil. Each call bumps SkillTreeGeneration so open skill windows rebuild.
func SetSkillTreeData(data *SkillTreeData) {
	activeSkillTree.Store(data)
	skillTreeGeneration.Add(1)
}

// ActiveSkillTreeData returns the installed data, or nil.
func ActiveSkillTreeData() *SkillTreeData { return activeSkillTree.Load() }

// SkillTreeGeneration changes whenever the skill tree data is replaced.
func SkillTreeGeneration() uint64 { return skillTreeGeneration.Load() }

// SkillTreeJobAlias maps mounted and riding job variants, which the client
// tables do not list, to the job whose tree they share (roBrowser's
// duplicateEntry and the official client's job-variant handling).
func SkillTreeJobAlias(job int) int {
	if alias, ok := skillTreeJobAliases[job]; ok {
		return alias
	}
	switch {
	case job >= 4265 && job <= 4277: // 4th job riding variants
		return job - 13
	case job >= 4309 && job <= 4315: // expanded 4th riding variants
		return job - 7
	}
	return job
}

var skillTreeJobAliases = map[int]int{
	JobKnight2: JobKnight, JobCrusader2: JobCrusader,
	JobKnight2H: JobKnightH, JobCrusader2H: JobCrusaderH,
	JobKnight2B: JobKnightB, JobCrusader2B: JobCrusaderB,
	JobStar2: JobStar, JobStar2B: JobStarB,
	4080: 4054, 4081: 4060, 4082: 4066, 4083: 4073, // rune knight / royal guard mounts
	4084: 4056, 4085: 4062, 4086: 4058, 4087: 4064, // ranger / mechanic mounts
	4109: 4096, 4110: 4102, 4111: 4098, 4112: 4100, // baby third mounts
	4243: 4239, 4244: 4241, // star emperor union
	4278: 4257, 4279: 4253, 4280: 4252, 4281: 4258, // 4th job mounts
	4316: 4302, // sky emperor union
}

// skillTreeBabyAdult maps a baby job to the adult job whose tree view it
// shows. The client tables give babies an inheritance chain but no positions.
func skillTreeBabyAdult(job int) (int, bool) {
	switch {
	case job >= 4023 && job <= 4043:
		return job - 4023, true
	case job >= 4096 && job <= 4101:
		return job - 42, true
	case job >= 4102 && job <= 4108:
		return job - 36, true
	}
	adult, ok := map[int]int{
		4045: 23, 4191: 4190, 4220: 4218, 4222: 25, 4223: 4211, 4224: 4212,
		4225: 4046, 4226: 4047, 4227: 4049, 4228: 24, 4229: 4215, 4241: 4239, 4242: 4240,
		4352: 4351, 4354: 4353,
	}[job]
	return adult, ok
}

// SkillTreeTab is the 1-based skill window tab (1st, 2nd, 3rd, 4th) a job's
// own skills belong to. It follows roBrowserLegacy's AddSkillTreeView rules,
// so transcendent second jobs share the second tab with their base job and
// expanded jobs (Star Gladiator, Kagerou, Super Novice expanded) are tab 2.
func SkillTreeTab(job int) int {
	switch {
	case job == JobNovice:
		return 1
	case job < 7 || job == 4046 || (job >= 23 && job <= 25) || job == 4218 || job == 4351:
		return 1
	case job < 4001 || job == 4047 || job == 4049 || (job >= 4211 && job <= 4215) ||
		job == 4190 || job == 4308 || job == 4353:
		return 2
	case (job >= 4001 && job <= 4007) || (job >= 4023 && job <= 4029) ||
		job == 4220 || job == 4222 || job == 4225 || job == 4228:
		return 1
	case job < 4054 || (job >= 4030 && job <= 4043) || (job >= 4223 && job <= 4229):
		return 2
	case job < 4252 || job == 4239 || job == 4240 || (job >= 4096 && job <= 4108) ||
		job == 4241 || job == 4242 || job == 4355:
		return 3
	case job <= 4264 || (job >= 4302 && job <= 4307):
		return 4
	}
	return 1
}

// Lineage returns job (after aliasing) and its ancestors, nearest first.
// It stops at a cycle or a job the tables do not list.
func (d *SkillTreeData) Lineage(job int) []int {
	if d == nil {
		return nil
	}
	job = SkillTreeJobAlias(job)
	out := make([]int, 0, 6)
	seen := make(map[int]bool, 6)
	for job >= 0 && !seen[job] {
		seen[job] = true
		out = append(out, job)
		entry := d.Jobs[job]
		if entry == nil {
			break
		}
		job = entry.Parent
	}
	return out
}

// Covers reports whether the data defines a tree for job.
func (d *SkillTreeData) Covers(job int) bool {
	if d == nil {
		return false
	}
	job = SkillTreeJobAlias(job)
	if _, ok := d.Jobs[job]; ok {
		return true
	}
	return len(d.PositionsFor(job)) > 0
}

// PositionsFor returns the grid positions shown for job's own skills: its own
// tree view, else the JOB_INHERIT_LIST2 job's, else the adult job's for a baby.
func (d *SkillTreeData) PositionsFor(job int) map[uint16]int {
	if d == nil {
		return nil
	}
	own := func(j int) map[uint16]int {
		if entry := d.Jobs[j]; entry != nil && len(entry.Positions) > 0 {
			return entry.Positions
		}
		if same, ok := d.SameTree[j]; ok {
			if entry := d.Jobs[same]; entry != nil && len(entry.Positions) > 0 {
				return entry.Positions
			}
		}
		return nil
	}
	if positions := own(job); positions != nil {
		return positions
	}
	if adult, ok := skillTreeBabyAdult(job); ok {
		return own(adult)
	}
	return nil
}

// TabLayout merges the positions of every job in the lineage whose skills
// belong to tab (1-based). Nearer jobs win a contested cell.
func (d *SkillTreeData) TabLayout(job, tab int) map[uint16]int {
	out := make(map[uint16]int)
	used := make(map[int]bool)
	for _, lineageJob := range d.Lineage(job) {
		if SkillTreeTab(lineageJob) != tab {
			continue
		}
		for skillID, position := range d.PositionsFor(lineageJob) {
			if _, ok := out[skillID]; ok || used[position] {
				continue
			}
			out[skillID] = position
			used[position] = true
		}
	}
	return out
}

// Groups returns the skills of every tab for job, ordered by grid position,
// or false when the data does not cover job.
func (d *SkillTreeData) Groups(job int) ([]SkillTreeGroup, bool) {
	if !d.Covers(job) {
		return nil, false
	}
	byTab := make(map[int]map[uint16]int)
	maxTab := 1
	for _, lineageJob := range d.Lineage(job) {
		tab := SkillTreeTab(lineageJob)
		if _, done := byTab[tab]; done {
			continue
		}
		layout := d.TabLayout(job, tab)
		if len(layout) == 0 {
			continue
		}
		byTab[tab] = layout
		maxTab = max(maxTab, tab)
	}
	groups := make([]SkillTreeGroup, 0, maxTab)
	for tab := 1; tab <= maxTab; tab++ {
		layout := byTab[tab]
		if len(layout) == 0 && tab != 1 {
			continue
		}
		ids := make([]uint16, 0, len(layout))
		for id := range layout {
			ids = append(ids, id)
		}
		sort.Slice(ids, func(i, j int) bool {
			if layout[ids[i]] != layout[ids[j]] {
				return layout[ids[i]] < layout[ids[j]]
			}
			return ids[i] < ids[j]
		})
		groups = append(groups, SkillTreeGroup{ClassLevel: tab, SkillIDs: ids})
	}
	return groups, true
}

// RequirementsFor resolves a skill's prerequisites for job the way
// roBrowserLegacy does: a NeedSkillList entry for the job or any ancestor
// (or the job it shares a tree with) wins, else _NeedSkillList. ok is false
// when SKILL_INFO_LIST does not define the skill.
func (d *SkillTreeData) RequirementsFor(job int, skillID uint16) ([]SkillRequirement, bool) {
	if d == nil || !d.HasRequirements[skillID] {
		return nil, false
	}
	if byJob := d.JobRequirements[skillID]; len(byJob) > 0 {
		for _, lineageJob := range d.Lineage(job) {
			if reqs, ok := byJob[lineageJob]; ok {
				return reqs, true
			}
			if same, ok := d.SameTree[lineageJob]; ok {
				if reqs, ok := byJob[same]; ok {
					return reqs, true
				}
			}
			if adult, ok := skillTreeBabyAdult(lineageJob); ok {
				if reqs, ok := byJob[adult]; ok {
					return reqs, true
				}
			}
		}
	}
	return d.Requirements[skillID], true
}
