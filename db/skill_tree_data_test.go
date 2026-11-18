package db

import "testing"

func TestSkillTreeTabMatchesRoBrowserLists(t *testing.T) {
	for job, want := range map[int]int{
		JobNovice: 1, JobSwordman: 1, JobTaekwon: 1, JobSuperNovice: 1, JobNinja: 1, JobSummoner: 1,
		JobKnight: 2, JobKnightH: 2, JobStar: 2, JobLinker: 2, 4211: 2, 4190: 2, 4308: 2,
		4001: 1, JobKnightB: 2, 4229: 2,
		JobRuneKnight: 3, 4060: 3, 4096: 3, 4239: 3, 4240: 3,
		JobDragonKnight: 4, JobTrouvere: 4, 4302: 4, 4307: 4,
	} {
		if got := SkillTreeTab(job); got != want {
			t.Errorf("SkillTreeTab(%d) = %d, want %d", job, got, want)
		}
	}
}

func TestSkillTreeJobAliasMapsMountsAndRiding(t *testing.T) {
	for job, want := range map[int]int{
		JobKnight2: JobKnight, JobCrusader2H: JobCrusaderH, 4080: JobRuneKnight,
		4265: JobDragonKnight, 4280: JobDragonKnight, 4314: 4307, JobKnight: JobKnight,
	} {
		if got := SkillTreeJobAlias(job); got != want {
			t.Errorf("SkillTreeJobAlias(%d) = %d, want %d", job, got, want)
		}
	}
}

func TestSkillTreeDataLineageStopsAtCycles(t *testing.T) {
	data := &SkillTreeData{Jobs: map[int]*SkillTreeJob{
		1: {Parent: 2}, 2: {Parent: 1, Positions: map[uint16]int{5: 0}},
	}}
	if got := data.Lineage(1); len(got) != 2 {
		t.Fatalf("lineage = %v", got)
	}
	var nilData *SkillTreeData
	if nilData.Covers(1) {
		t.Fatal("nil data covers a job")
	}
	if _, ok := nilData.Groups(1); ok {
		t.Fatal("nil data has groups")
	}
}
