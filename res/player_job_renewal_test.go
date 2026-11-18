package res

import (
	"sort"
	"strings"
	"testing"

	"github.com/kivutar/goro/db"
)

func TestRenewalJobsHavePlayerSpriteTokens(t *testing.T) {
	for job, want := range map[int]string{
		db.JobRuneKnight:   "룬나이트",
		4060:               "룬나이트",
		db.JobSummoner:     "summoner",
		db.JobDragonKnight: "dragon_knight",
		4302:               "sky_emperor",
		4307:               "hyper_novice",
	} {
		if !HasPlayerJobToken(job) {
			t.Fatalf("job %d has no player sprite token", job)
		}
		if got := PlayerJobToken(job); got != want {
			t.Fatalf("job %d token = %q, want %q", job, got, want)
		}
	}
	if got := PlayerJobToken(db.JobKnightH); got != "로드나이트" {
		t.Fatalf("classic table entry changed: %q", got)
	}
}

func TestDoramPlayerPathsUseDoramRace(t *testing.T) {
	body := PlayerBodyResourceCandidates(db.JobSummoner, 1, "spr")
	if body[0] != "data\\sprite\\도람족\\몸통\\남\\summoner_남.spr" {
		t.Fatalf("doram body = %q", body[0])
	}
	head := PlayerHeadResourceCandidates(db.JobSummoner, 4, 0, "act")
	if head[0] != "data\\sprite\\도람족\\머리통\\여\\4_여.act" || head[len(head)-1] != "data\\sprite\\도람족\\머리통\\여\\1_여.act" {
		t.Fatalf("doram head = %q", head)
	}
	pal := PlayerHeadPaletteResourceCandidates(db.JobSummoner, 2, 1, 3, "pal")
	if pal[0] != "data\\palette\\도람족\\머리\\머리2_남_3.pal" {
		t.Fatalf("doram head palette = %q", pal)
	}
	if human := PlayerBodyResourceCandidates(db.JobRuneKnight, 0, "spr"); !strings.HasPrefix(human[0], playerHumanSpriteRoot) {
		t.Fatalf("human body = %q", human[0])
	}
}

// Every renewal job body must resolve in the real data set for at least one
// sex; jobs without any sprite are listed so table typos show up.
func TestRenewalJobBodySpritesRealWhenConfigured(t *testing.T) {
	m := realDataManager(t)
	var missing []string
	found := 0
	for job := range db.JobResourceName {
		if job < 4054 {
			continue
		}
		ok := false
		for _, sex := range []byte{0, 1} {
			if m.HasFileExact(PlayerBodyResourceCandidates(job, sex, "spr")[0]) && m.HasFileExact(PlayerBodyResourceCandidates(job, sex, "act")[0]) {
				ok = true
			}
		}
		if ok {
			found++
		} else {
			missing = append(missing, PlayerBodyResourceCandidates(job, 1, "spr")[0])
		}
	}
	sort.Strings(missing)
	t.Logf("renewal job bodies found=%d missing=%d %v", found, len(missing), missing)
	if found < 150 {
		t.Fatalf("only %d renewal job bodies resolve", found)
	}
	for _, sex := range []byte{0, 1} {
		for _, path := range []string{PlayerBodyResourceCandidates(db.JobSummoner, sex, "spr")[0], PlayerHeadResourceCandidates(db.JobSummoner, 1, sex, "spr")[0]} {
			if !m.HasFileExact(path) {
				t.Fatalf("doram sprite %s missing", path)
			}
		}
	}
}
