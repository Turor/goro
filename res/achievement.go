package res

import (
	"regexp"
	"strconv"
	"strings"

	glua "github.com/yuin/gopher-lua"
)

type AchievementInfo struct {
	Title    string
	Summary  string
	Contents string
	Group    string
	Major    int
	Minor    int
}

type AttendanceReward struct {
	Day      int
	ItemID   int
	Quantity int
}

var achievementLuaCandidates = []string{
	"System/achievement_list.lub",
	"data/System/achievement_list.lub",
	"System/achievements.lub",
	"SystemEN/achievements.lub",
	"data/SystemEN/achievements.lub",
}

var attendanceLuaCandidates = []string{
	"System/CheckAttendance.turoran1.lub",
	"System/CheckAttendance.lub",
	"data/System/CheckAttendance.turoran1.lub",
	"data/System/CheckAttendance.lub",
}

func (m *Manager) AchievementInfo(id uint32) (AchievementInfo, bool) {
	if m == nil {
		return AchievementInfo{}, false
	}
	m.loadAchievementInfo()
	info, ok := m.achievementInfo[id]
	return info, ok && info.Title != ""
}

func (m *Manager) AttendanceRewards() []AttendanceReward {
	if m == nil {
		return nil
	}
	m.loadAttendanceRewards()
	return m.attendanceRewards
}

func (m *Manager) loadAchievementInfo() {
	if m.achievementInfoLoaded {
		return
	}
	m.achievementInfoLoaded = true
	m.achievementInfo = map[uint32]AchievementInfo{}
	for _, name := range achievementLuaCandidates {
		_, data, ok := m.ReadFirst([]string{name})
		if !ok {
			continue
		}
		globals := map[string]luaValue{}
		if err := executeLuaChunk(data, globals); err != nil {
			continue
		}
		m.applyAchievementTable(globals["achievement_tbl"])
		m.applyAchievementTable(globals["achievements"])
	}
}

func (m *Manager) applyAchievementTable(tbl luaValue) {
	if tbl.kind != luaTable {
		return
	}
	for key, value := range tbl.table {
		id, ok := luaTableIntKey(key)
		if !ok || id <= 0 || value.kind != luaTable {
			continue
		}
		info := m.achievementInfo[uint32(id)]
		if title := luaStringField(value, "title"); title != "" {
			info.Title = title
		}
		if summary := luaStringField(value, "summary"); summary != "" {
			info.Summary = summary
		}
		if contents := luaStringField(value, "contents"); contents != "" {
			info.Contents = contents
		}
		if group := luaStringField(value, "group"); group != "" {
			info.Group = group
		}
		if n := luaNumberField(value, "major"); n != 0 {
			info.Major = n
		}
		if n := luaNumberField(value, "minor"); n != 0 {
			info.Minor = n
		}
		m.achievementInfo[uint32(id)] = info
	}
}

func (m *Manager) loadAttendanceRewards() {
	if m.attendanceRewardsLoaded {
		return
	}
	m.attendanceRewardsLoaded = true
	_, data, ok := m.ReadFirst(attendanceLuaCandidates)
	if !ok {
		return
	}
	rewards := map[int]AttendanceReward{}
	globals := map[string]luaValue{}
	_ = executeLuaChunkWithSetup(data, globals, func(state *glua.LState) {
		state.SetGlobal("InsertCheckAttendanceConfig", state.NewFunction(func(state *glua.LState) int { return 1 }))
		state.SetGlobal("InsertCheckAttendanceReward", state.NewFunction(func(state *glua.LState) int {
			day := state.OptInt(1, 0)
			item := state.OptInt(2, 0)
			qty := state.OptInt(3, 0)
			if day > 0 {
				rewards[day] = AttendanceReward{Day: day, ItemID: item, Quantity: qty}
			}
			return 1
		}))
	})
	if len(rewards) == 0 {
		re := regexp.MustCompile(`InsertCheckAttendanceReward\(\s*(\d+)\s*,\s*(\d+)\s*,\s*(\d+)\s*\)`)
		for _, match := range re.FindAllSubmatch(data, -1) {
			day, _ := strconv.Atoi(string(match[1]))
			item, _ := strconv.Atoi(string(match[2]))
			qty, _ := strconv.Atoi(string(match[3]))
			if day > 0 {
				rewards[day] = AttendanceReward{Day: day, ItemID: item, Quantity: qty}
			}
		}
	}
	for day := 1; day <= 20; day++ {
		if r, ok := rewards[day]; ok {
			m.attendanceRewards = append(m.attendanceRewards, r)
		}
	}
}

func luaStringField(table luaValue, key string) string {
	value := luaField(table, key)
	if value.kind == luaString {
		return strings.TrimSpace(decodeROText([]byte(value.str)))
	}
	return ""
}

func luaNumberField(table luaValue, key string) int {
	value := luaField(table, key)
	if value.kind == luaNumber {
		return int(value.num)
	}
	return 0
}
