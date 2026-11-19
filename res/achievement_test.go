package res

import (
	"testing"
	"testing/fstest"
)

func TestAchievementInfoEnglishOverlay(t *testing.T) {
	files := fstest.MapFS{
		"System/achievement_list.lub": &fstest.MapFile{Data: []byte(`
achievement_tbl = {
  [1001] = { title = "Korean title", summary = "Korean summary", contents = "Korean body", major = 1, minor = 2 }
}
`)},
		"SystemEN/achievements.lub": &fstest.MapFile{Data: []byte(`
achievement_tbl = {
  [1001] = { title = "First Hunt", summary = "Kill 10 porings", contents = "Hunt porings in town.", major = 1, minor = 2 }
}
`)},
	}
	manager, err := NewManagerFS(files)
	if err != nil {
		t.Fatal(err)
	}
	info, ok := manager.AchievementInfo(1001)
	if !ok {
		t.Fatal("missing achievement 1001")
	}
	if info.Title != "First Hunt" {
		t.Fatalf("title = %q", info.Title)
	}
	if info.Summary != "Kill 10 porings" {
		t.Fatalf("summary = %q", info.Summary)
	}
	if info.Major != 1 || info.Minor != 2 {
		t.Fatalf("major/minor = %d/%d", info.Major, info.Minor)
	}
}

func TestAttendanceRewardsFromLuaCalls(t *testing.T) {
	files := fstest.MapFS{
		"System/CheckAttendance.lub": &fstest.MapFile{Data: []byte(`
InsertCheckAttendanceReward(1, 501, 5)
InsertCheckAttendanceReward(2, 502, 10)
InsertCheckAttendanceReward(20, 512, 1)
`)},
	}
	manager, err := NewManagerFS(files)
	if err != nil {
		t.Fatal(err)
	}
	rewards := manager.AttendanceRewards()
	if len(rewards) != 3 {
		t.Fatalf("rewards = %#v", rewards)
	}
	if rewards[0] != (AttendanceReward{Day: 1, ItemID: 501, Quantity: 5}) {
		t.Fatalf("day 1 = %#v", rewards[0])
	}
	if rewards[2].Day != 20 || rewards[2].ItemID != 512 {
		t.Fatalf("day 20 = %#v", rewards[2])
	}
}
