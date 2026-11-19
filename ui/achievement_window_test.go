package ui

import (
	"strings"
	"testing"

	"github.com/kivutar/goro/network"
	"github.com/kivutar/goro/res"
)

func TestAchievementTitleUsesEnglishMetadata(t *testing.T) {
	title := achievementTitle(Context{Resources: &res.Manager{}}, network.Achievement{ID: 7})
	if !strings.Contains(title, "7") {
		t.Fatalf("fallback title = %q", title)
	}
}

func TestAttendanceStatusEnglish(t *testing.T) {
	w := &AttendanceWindow{}
	w.applyData(30) // 3 days claimed, not yet requested today
	if w.day != 3 || w.claimed {
		t.Fatalf("day=%d claimed=%t", w.day, w.claimed)
	}
	w.applyData(31)
	if !w.claimed {
		t.Fatal("expected claimed")
	}
}

func TestAttendanceClaimAckIsRawCount(t *testing.T) {
	w := &AttendanceWindow{}
	w.SetProgress(Context{}, 4)
	if w.day != 4 || !w.claimed {
		t.Fatalf("claim ack decoded as open encoding: day=%d claimed=%t", w.day, w.claimed)
	}
}

func TestAchievementCategoriesMatchMajors(t *testing.T) {
	cats := achievementCategories(Context{})
	if len(cats) != 7 {
		t.Fatalf("categories = %d, want majors 0..6", len(cats))
	}
	w := &AchievementWindow{category: 0, List: network.AchievementList{Items: []network.Achievement{
		{ID: 1},
		{ID: 2},
	}}}
	// Without metadata, major is 0, so General keeps only major 0.
	rows := w.filteredRows(Context{})
	if len(rows) != 2 {
		t.Fatalf("general filter len=%d", len(rows))
	}
	w.category = 5
	if len(w.filteredRows(Context{})) != 0 {
		t.Fatal("quest tab included major 0 rows")
	}
}
