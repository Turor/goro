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
