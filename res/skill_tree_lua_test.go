package res

import (
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/kivutar/goro/db"
)

// skillTreeDataURL checks the real streamed tables:
//
//	go test ./res -run TestLuaSkillTreeRealStreamedData -skilltree-url <asset server root>
//
// The URL is only ever passed on the command line, never committed.
var skillTreeDataURL = flag.String("skilltree-url", "", "asset server root for the real skill tree check")

const skillTreeFixtureDir = "data/luafiles514/lua files/skillinfoz/"

var skillTreeFixture = map[string]string{
	"data/luafiles514/lua files/datainfo/jobidentity.lub": `JOBID = {
		JT_NOVICE = 0, JT_SWORDMAN = 1, JT_KNIGHT = 7, JT_KNIGHT_H = 4008,
		JT_RUNE_KNIGHT = 4054, JT_RUNE_KNIGHT_H = 4060, JT_DRAGON_KNIGHT = 4252,
		JT_SWORDMAN_B = 4024, JT_KNIGHT_B = 4030, JT_NOVICE_B = 4023,
	}`,
	skillTreeFixtureDir + "skillid.lub": `SKID = {
		NV_BASIC = 1, SM_SWORD = 2, SM_BASH = 5, SM_MAGNUM = 7, KN_SPEARMASTERY = 55,
		KN_PIERCE = 56, LK_AURABLADE = 355, RK_ENCHANTBLADE = 2001, DK_SERVANTWEAPON = 5001,
	}`,
	skillTreeFixtureDir + "jobinheritlist.lub": `JOB_INHERIT_LIST = {
		[JOBID.JT_SWORDMAN] = JOBID.JT_NOVICE,
		[JOBID.JT_KNIGHT] = JOBID.JT_SWORDMAN,
		[JOBID.JT_KNIGHT_H] = JOBID.JT_KNIGHT,
		[JOBID.JT_RUNE_KNIGHT] = JOBID.JT_KNIGHT,
		[JOBID.JT_RUNE_KNIGHT_H] = JOBID.JT_KNIGHT_H,
		[JOBID.JT_DRAGON_KNIGHT] = JOBID.JT_RUNE_KNIGHT_H,
		[JOBID.JT_NOVICE_B] = JOBID.JT_NOVICE,
		[JOBID.JT_SWORDMAN_B] = JOBID.JT_NOVICE_B,
		[JOBID.JT_KNIGHT_B] = JOBID.JT_SWORDMAN_B,
	}
	JOB_INHERIT_LIST2 = { [JOBID.JT_RUNE_KNIGHT_H] = JOBID.JT_RUNE_KNIGHT }`,
	skillTreeFixtureDir + "skilltreeview.lub": `SKILL_TREEVIEW_FOR_JOB = {
		[JOBID.JT_NOVICE] = { [0] = SKID.NV_BASIC },
		[JOBID.JT_SWORDMAN] = { [1] = SKID.SM_SWORD, [2] = SKID.SM_BASH, [8] = SKID.SM_MAGNUM },
		[JOBID.JT_KNIGHT] = { [0] = SKID.KN_SPEARMASTERY, [1] = SKID.KN_PIERCE },
		[JOBID.JT_KNIGHT_H] = { [5] = SKID.LK_AURABLADE },
		[JOBID.JT_RUNE_KNIGHT] = { [0] = SKID.RK_ENCHANTBLADE },
		[JOBID.JT_DRAGON_KNIGHT] = { [3] = SKID.DK_SERVANTWEAPON },
	}
	JobSkillTab.ChangeSkillTabName(JOBID.JT_DRAGON_KNIGHT, "1st", "2nd", "3rd", "4th")`,
	skillTreeFixtureDir + "skillinfolist.lub": `SKILL_INFO_LIST = {
		[SKID.NV_BASIC] = { "NV_BASIC", SkillName = "Basic Skill", MaxLv = 9 },
		[SKID.SM_SWORD] = { "SM_SWORD", SkillName = "Sword Mastery", MaxLv = 10 },
		[SKID.SM_BASH] = { "SM_BASH", SkillName = "Bash", MaxLv = 10 },
		[SKID.SM_MAGNUM] = { "SM_MAGNUM", SkillName = "Magnum Break", MaxLv = 10,
			_NeedSkillList = { { SKID.SM_BASH, 5 } } },
		[SKID.KN_SPEARMASTERY] = { "KN_SPEARMASTERY", SkillName = "Spear Mastery", MaxLv = 10 },
		[SKID.KN_PIERCE] = { "KN_PIERCE", SkillName = "Pierce", MaxLv = 10,
			_NeedSkillList = { { SKID.KN_SPEARMASTERY, 1 } },
			NeedSkillList = { [JOBID.JT_RUNE_KNIGHT] = { { SKID.KN_SPEARMASTERY, 3 } } } },
		[SKID.LK_AURABLADE] = { "LK_AURABLADE", SkillName = "Aura Blade", MaxLv = 5,
			_NeedSkillList = { { SKID.SM_MAGNUM, 5 } } },
		[SKID.RK_ENCHANTBLADE] = { "RK_ENCHANTBLADE", SkillName = "Enchant Blade", MaxLv = 5,
			_NeedSkillList = { { SKID.KN_PIERCE, 1 } } },
		[SKID.DK_SERVANTWEAPON] = { "DK_SERVANTWEAPON", SkillName = "Servant Weapon", MaxLv = 5,
			_NeedSkillList = { { SKID.RK_ENCHANTBLADE, 5 } } },
	}`,
}

func writeSkillTreeFixture(t *testing.T, overrides map[string]string) string {
	t.Helper()
	root := t.TempDir()
	files := make(map[string]string, len(skillTreeFixture))
	for name, body := range skillTreeFixture {
		files[name] = body
	}
	for name, body := range overrides {
		files[name] = body
	}
	for name, body := range files {
		path := filepath.Join(root, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

func loadSkillTreeFixture(t *testing.T, overrides map[string]string) *Manager {
	t.Helper()
	m, err := NewManager(writeSkillTreeFixture(t, overrides))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.SetSkillTreeData(nil) })
	if !m.EnsureSkillTree() {
		t.Fatal("skill tree not loaded from fixture tables")
	}
	return m
}

func groupIDs(groups []db.SkillTreeGroup) map[int][]uint16 {
	out := map[int][]uint16{}
	for _, group := range groups {
		out[group.ClassLevel] = group.SkillIDs
	}
	return out
}

func sameIDs(a, b []uint16) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestLuaSkillTreeFourthJobTabsFollowInheritance(t *testing.T) {
	m := loadSkillTreeFixture(t, nil)
	if warnings := m.SkillTreeWarnings(); len(warnings) != 0 {
		t.Fatalf("clean fixture warned: %v", warnings)
	}
	tabs := groupIDs(db.SkillTreeSkillGroups(4252))
	want := map[int][]uint16{
		1: {1, 2, 5, 7},  // novice and swordman, by grid cell
		2: {55, 56, 355}, // knight and lord knight share the second tab
		3: {2001},        // the transcendent rune knight shows the rune knight tree
		4: {5001},        // dragon knight
	}
	for tab, ids := range want {
		if !sameIDs(tabs[tab], ids) {
			t.Errorf("dragon knight tab %d = %v, want %v", tab, tabs[tab], ids)
		}
	}
	if layout, ok := db.SkillTreeLayout(4252, 1); !ok || layout[1] != 0 || layout[7] != 8 || layout[2] != 1 {
		t.Errorf("tab 1 layout = %v, %v", layout, ok)
	}
	if level, ok := db.SkillMaxLevel(1); !ok || level != 9 {
		t.Errorf("max level from SKILL_INFO_LIST = %d, %v", level, ok)
	}
	// A mounted knight shares the knight tree; a baby knight shows the
	// knight's positions through its own inheritance chain.
	if got, want := groupIDs(db.SkillTreeSkillGroups(db.JobKnight2)), groupIDs(db.SkillTreeSkillGroups(db.JobKnight)); !sameIDs(got[2], want[2]) {
		t.Errorf("mounted knight tab 2 = %v, want %v", got[2], want[2])
	}
	if baby := groupIDs(db.SkillTreeSkillGroups(4030)); !sameIDs(baby[2], []uint16{55, 56}) {
		t.Errorf("baby knight tab 2 = %v", baby[2])
	}
}

func TestLuaSkillTreeRequirementsPreferJobLists(t *testing.T) {
	loadSkillTreeFixture(t, nil)
	if reqs := db.SkillRequirementsForJob(db.JobKnight, 56); len(reqs) != 1 || reqs[0].Level != 1 {
		t.Errorf("knight pierce requirements = %v", reqs)
	}
	// NeedSkillList[JT_RUNE_KNIGHT] applies to the rune knight and every job
	// that inherits from it.
	for _, job := range []int{4054, 4252} {
		if reqs := db.SkillRequirementsForJob(job, 56); len(reqs) != 1 || reqs[0].Level != 3 {
			t.Errorf("job %d pierce requirements = %v", job, reqs)
		}
	}
	if reqs := db.SkillRequirementsForJob(4252, 5001); len(reqs) != 1 || reqs[0].SkillID != 2001 {
		t.Errorf("servant weapon requirements = %v", reqs)
	}
}

func TestLuaSkillTreeValidationWarnings(t *testing.T) {
	m := loadSkillTreeFixture(t, map[string]string{
		skillTreeFixtureDir + "skilltreeview.lub": `SKILL_TREEVIEW_FOR_JOB = {
			[JOBID.JT_NOVICE] = { [0] = SKID.NV_BASIC },
			[JOBID.JT_SWORDMAN] = { [1] = SKID.SM_SWORD, [2] = SKID.SM_BASH, [3] = SKID.SM_TYPO },
			[JOBID.JT_KNIGHT] = { [0] = SKID.KN_SPEARMASTERY, [1] = SKID.KN_PIERCE },
			[JOBID.JT_KNIGHT_H] = { [1] = SKID.LK_AURABLADE },
			[JOBID.JT_RUNE_KNIGHT] = { [0] = SKID.RK_ENCHANTBLADE },
			[JOBID.JT_DRAGON_KNIGHT] = { [3] = SKID.DK_SERVANTWEAPON },
			[4999] = { [0] = SKID.SM_SWORD },
		}`,
		skillTreeFixtureDir + "skillinfolist.lub": `SKILL_INFO_LIST = {
			[SKID.NV_BASIC] = { "NV_BASIC", MaxLv = 9 },
			[SKID.SM_SWORD] = { "SM_SWORD", MaxLv = 10, _NeedSkillList = { { SKID.SM_BASH, 1 } } },
			[SKID.SM_BASH] = { "SM_BASH", MaxLv = 10, _NeedSkillList = { { SKID.SM_SWORD, 1 } } },
			[SKID.KN_SPEARMASTERY] = { "KN_SPEARMASTERY", MaxLv = 10, _NeedSkillList = { { SKID.SM_GONE, 1 } } },
			[SKID.KN_PIERCE] = { "KN_PIERCE", MaxLv = 10, _NeedSkillList = { { SKID.KN_SPEARMASTERY, 11 } } },
			[SKID.LK_AURABLADE] = { "LK_AURABLADE", MaxLv = 5, _NeedSkillList = { { SKID.SM_MAGNUM, 5 } } },
			[SKID.RK_ENCHANTBLADE] = { "RK_ENCHANTBLADE", MaxLv = 5 },
			[SKID.DK_SERVANTWEAPON] = { "DK_SERVANTWEAPON", MaxLv = 5 },
		}`,
		skillTreeFixtureDir + "jobinheritlist.lub": skillTreeFixture[skillTreeFixtureDir+"jobinheritlist.lub"] +
			"\nJOB_INHERIT_LIST[4999] = 4998",
	})
	joined := strings.Join(m.SkillTreeWarnings(), "\n")
	for _, want := range []string{
		"skilltreeview: unknown SKID.SM_TYPO", // unknown skill name
		"skillinfolist: unknown SKID.SM_GONE", // unknown prerequisite name
		"KN_SPEARMASTERY(55): _NeedSkillList entry 1 names an unknown skill",
		"LK_AURABLADE(355) requires SM_MAGNUM(7), which is not in its tree", // missing prerequisite
		"requires KN_SPEARMASTERY(55) level 11, above its max level 10",
		"tab 2 cell 1: KN_PIERCE(56) of JT_KNIGHT(7) overlaps LK_AURABLADE(355) of JT_KNIGHT_H(4008)", // duplicate cell
		"prerequisite cycle SM_SWORD(2) -> SM_BASH(5)",                                                // cycle
		"job 4999 inherits from job 4998, which has no tree entry",                                    // unknown parent
	} {
		if !strings.Contains(joined, want) {
			t.Errorf("missing warning %q in:\n%s", want, joined)
		}
	}
}

func TestLuaSkillTreeReloadPicksUpEdits(t *testing.T) {
	m := loadSkillTreeFixture(t, nil)
	before := db.SkillTreeGeneration()
	path := filepath.Join(m.Root, filepath.FromSlash(skillTreeFixtureDir+"skilltreeview.lub"))
	edited := strings.Replace(skillTreeFixture[skillTreeFixtureDir+"skilltreeview.lub"],
		"[JOBID.JT_DRAGON_KNIGHT] = { [3] = SKID.DK_SERVANTWEAPON }",
		"[JOBID.JT_DRAGON_KNIGHT] = { [3] = SKID.DK_SERVANTWEAPON, [4] = SKID.NV_BASIC }", 1)
	if err := os.WriteFile(path, []byte(edited), 0o644); err != nil {
		t.Fatal(err)
	}
	// A path that is not a skill table changes nothing.
	if m.InvalidateData([]string{"data/texture/x.bmp"}) {
		t.Fatal("non-skill invalidation reloaded the skill tree")
	}
	m.QueueDataInvalidation("data/luafiles514/lua files/skillinfoz/skilltreeview.fourth.lub")
	if !m.ApplyDataInvalidations() {
		t.Fatal("queued invalidation did not reload")
	}
	if db.SkillTreeGeneration() == before {
		t.Fatal("generation unchanged after reload")
	}
	if tabs := groupIDs(db.SkillTreeSkillGroups(4252)); !sameIDs(tabs[4], []uint16{5001, 1}) {
		t.Errorf("tab 4 after reload = %v", tabs[4])
	}
	if m.ApplyDataInvalidations() {
		t.Error("empty queue reloaded again")
	}
}

func TestLuaSkillTreeMissingTablesKeepBuiltinTrees(t *testing.T) {
	m, err := NewManager(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.SetSkillTreeData(nil) })
	if m.EnsureSkillTree() {
		t.Fatal("tree installed without tables")
	}
	if db.ActiveSkillTreeData() != nil {
		t.Fatal("data installed")
	}
	if groups := db.SkillTreeSkillGroups(db.JobKnight); len(groups) != 2 {
		t.Fatalf("built-in knight groups = %d", len(groups))
	}
}

func TestIsSkillDataFile(t *testing.T) {
	for name, want := range map[string]bool{
		"data/luafiles514/lua files/skillinfoz/skillinfolist.turoran13.lub": true,
		`data\luafiles514\lua files\skillinfoz\SkillTreeView.lub`:           true,
		"System/skilldescript.turoran.re2.lub":                              true,
		"data/luafiles514/lua files/datainfo/jobidentity.lub":               true,
		"data/luafiles514/lua files/skillinfoz/skillidle.lub":               false,
		"System/itemInfo.lua": false,
	} {
		if got := IsSkillDataFile(name); got != want {
			t.Errorf("IsSkillDataFile(%q) = %v", name, got)
		}
	}
}

func TestRemoteEvictRevalidatesWithETag(t *testing.T) {
	srv := newFakeAssetServer(t)
	srv.put("data/table.txt", []byte("v1"), `"1"`)
	m := streamingManager(t, t.TempDir(), t.TempDir(), srv.url(), nil)
	m.remote.refreshed = func(string) {}
	if data, err := m.ReadFile("data\\table.txt"); err != nil || string(data) != "v1" {
		t.Fatalf("first read: %q %v", data, err)
	}
	hits := srv.hitCount("data/table.txt")
	srv.put("data/table.txt", []byte("v2"), `"2"`)
	if data, _ := m.ReadFile("data\\table.txt"); string(data) != "v1" {
		t.Fatalf("validated file was refetched: %q", data)
	}
	m.InvalidateData([]string{"data\\table.txt"})
	if data, err := m.ReadFile("data\\table.txt"); err != nil || string(data) != "v2" {
		t.Fatalf("after evict: %q %v", data, err)
	}
	if got := srv.hitCount("data/table.txt"); got != hits+1 {
		t.Fatalf("hits after evict = %d, want %d", got, hits+1)
	}
	// Unchanged after a second eviction: a conditional request, cache kept.
	m.InvalidateData([]string{"data\\table.txt"})
	if data, _ := m.ReadFile("data\\table.txt"); string(data) != "v2" {
		t.Fatalf("after 304: %q", data)
	}
}

func TestRemoteChangedRefreshQueuesInvalidation(t *testing.T) {
	srv := newFakeAssetServer(t)
	p := "data/luafiles514/lua files/skillinfoz/skilltreeview.lub"
	srv.put(p, []byte("SKILL_TREEVIEW_FOR_JOB = {}"), `"1"`)
	cache := t.TempDir()
	first := streamingManager(t, t.TempDir(), cache, srv.url(), nil)
	if _, err := first.ReadFile(p); err != nil {
		t.Fatal(err)
	}
	srv.put(p, []byte("SKILL_TREEVIEW_FOR_JOB = { [0] = {} }"), `"2"`)
	second := streamingManager(t, t.TempDir(), cache, srv.url(), nil)
	done := make(chan string, 1)
	second.remote.refreshed = func(p string) { done <- p }
	if _, err := second.ReadFile(p); err != nil {
		t.Fatal(err)
	}
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("no background refresh")
	}
	second.invalidations.mu.Lock()
	queued := append([]string(nil), second.invalidations.paths...)
	second.invalidations.mu.Unlock()
	if len(queued) != 1 || queued[0] != p {
		t.Fatalf("queued invalidations = %v", queued)
	}
}

// TestLuaSkillTreeRealStreamedData loads the real tables through streaming.
// It runs only with -skilltree-url; the gate skips it.
func TestLuaSkillTreeRealStreamedData(t *testing.T) {
	if *skillTreeDataURL == "" {
		t.Skip("pass -skilltree-url to check the real streamed skill tables")
	}
	m := streamingManager(t, t.TempDir(), t.TempDir(), *skillTreeDataURL, TuroranRemoteAliases)
	m.WaitRemoteIndex(30 * time.Second)
	t.Cleanup(func() { db.SetSkillTreeData(nil) })
	start := time.Now()
	if !m.EnsureSkillTree() {
		t.Fatal("real skill tables did not load")
	}
	t.Logf("loaded in %s, %d requests", time.Since(start).Round(time.Millisecond), m.RemoteRequests())
	for _, warning := range m.SkillTreeWarnings() {
		t.Logf("warning: %s", warning)
	}
	data := db.ActiveSkillTreeData()
	covered, full := 0, 0
	for job := range data.Jobs {
		if !data.Covers(job) {
			continue
		}
		covered++
		if groups, _ := data.Groups(job); len(groups) >= 2 {
			full++
		}
	}
	t.Logf("%d jobs covered, %d with more than one tab", covered, full)
	for _, job := range []int{db.JobRuneKnight, 4060, db.JobDragonKnight, 4302, 4307} {
		groups, ok := data.Groups(job)
		if !ok || len(groups) < 3 {
			t.Errorf("job %d: %d tabs, covered %v", job, len(groups), ok)
		}
		counts := make([]int, len(groups))
		for i, group := range groups {
			counts[i] = len(group.SkillIDs)
		}
		t.Logf("job %d tabs %v", job, counts)
	}
}

// Renamed skills keep their old SKID name as an alias of the same ID. Both
// names must resolve every time, not whichever one map order kept.
func TestLuaSkillTreeSkillIDAliasesAllResolve(t *testing.T) {
	for i := 0; i < 10; i++ {
		m := loadSkillTreeFixture(t, map[string]string{
			skillTreeFixtureDir + "skillid.lub": strings.Replace(skillTreeFixture[skillTreeFixtureDir+"skillid.lub"],
				"KN_PIERCE = 56,", "KN_PIERCE = 56, KN_PIERCE_OLD = 56, A_PIERCE = 56,", 1),
			skillTreeFixtureDir + "skilltreeview.lub": strings.Replace(skillTreeFixture[skillTreeFixtureDir+"skilltreeview.lub"],
				"[1] = SKID.KN_PIERCE", "[1] = SKID.KN_PIERCE_OLD", 1),
		})
		if warnings := m.SkillTreeWarnings(); len(warnings) != 0 {
			t.Fatalf("run %d warned: %v", i, warnings)
		}
		if tabs := groupIDs(db.SkillTreeSkillGroups(db.JobKnight)); !sameIDs(tabs[2], []uint16{55, 56}) {
			t.Fatalf("run %d knight tab 2 = %v", i, tabs[2])
		}
		if name, _ := m.SkillResourceName(56); name != "A_PIERCE" {
			t.Fatalf("run %d display name for 56 = %q, want the smallest alias", i, name)
		}
	}
}
