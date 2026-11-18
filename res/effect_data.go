package res

import (
	"path"
	"strings"
	"time"

	"github.com/kivutar/goro/db"
	"github.com/kivutar/goro/glog"
)

var effectTablesJSONCandidates = []string{
	"data\\contentdata\\effectdata\\effects.json",
	"data\\effectdata\\effects.json",
	"data\\effects.json",
	"effects.json",
}

var ez2STREffectBSONCandidates = []string{
	"data\\contentdata\\effectdata\\ez2streffect.bson",
	"data\\effectdata\\ez2streffect.bson",
	"data\\contentdata\\effectdata\\ez2streffect.BSON",
}

const effectDataRetry = 30 * time.Second

// EnsureEffectData loads effects.json through the streaming path (then merges
// ez2streffect.bson when present) and installs the live tables. Missing files
// keep the compiled fallback. Call it from the game thread.
func (m *Manager) EnsureEffectData() bool {
	if m == nil {
		return false
	}
	if m.effectDataLoaded {
		return db.EffectTablesUseDataFile()
	}
	if !m.effectDataRetryAt.IsZero() && time.Now().Before(m.effectDataRetryAt) {
		return db.EffectTablesUseDataFile()
	}
	installed, err := m.loadEffectTables()
	if err != nil {
		glog.Errorf("effects: %v (keeping previous tables)", err)
		m.effectDataLoaded = true
		m.effectDataRetryAt = time.Time{}
		return db.EffectTablesUseDataFile()
	}
	if !installed && m.remote != nil {
		m.effectDataRetryAt = time.Now().Add(effectDataRetry)
		return false
	}
	m.effectDataLoaded = true
	m.effectDataRetryAt = time.Time{}
	return db.EffectTablesUseDataFile()
}

// ReloadEffectData drops the load-once flag and reads the tables again.
func (m *Manager) ReloadEffectData() bool {
	if m == nil {
		return false
	}
	m.effectDataLoaded = false
	m.effectDataRetryAt = time.Time{}
	return m.EnsureEffectData()
}

func (m *Manager) loadEffectTables() (bool, error) {
	overlay := db.EffectTables{}
	fromJSON := false
	if _, data, ok := m.ReadFirst(effectTablesJSONCandidates); ok {
		parsed, err := db.ParseEffectTablesJSON(data)
		if err != nil {
			return false, err
		}
		overlay = parsed
		fromJSON = true
	}

	specs, skills, actions := db.OverlayEffectTables(overlay)
	fromBSON := false
	var names map[string]int
	if _, data, ok := m.ReadFirst(ez2STREffectBSONCandidates); ok {
		table, err := db.ParseEz2STREffectBSON(data)
		if err != nil {
			glog.Errorf("effects: ez2streffect.bson: %v (skipping BSON merge)", err)
		} else if len(table.Entries) > 0 {
			names = db.MergeEz2Effects(specs, skills, table, m.skillIDsForEz2())
			fromBSON = true
			glog.Infof("effects: merged %d ez2streffect entries", len(table.Entries))
		}
	}

	fromData := fromJSON || fromBSON
	if !fromData {
		if db.EffectTablesUseDataFile() {
			db.RestoreCompiledEffectTables()
		}
		glog.Infof("effects: client tables not found, using built-in specs")
		return false, nil
	}
	db.InstallEffectTables(specs, skills, actions, names, true)
	if fromJSON {
		glog.Infof("effects: loaded %d overlay specs, %d skill effects, %d skill actions",
			len(overlay.Effects), len(overlay.SkillEffects), len(overlay.SkillActions))
	}
	return true, nil
}

func (m *Manager) skillIDsForEz2() map[string]uint16 {
	m.loadSkillResourceNames()
	out := make(map[string]uint16, len(m.skillIDNames)+len(db.SkillResourceName))
	for id, name := range db.SkillResourceName {
		if name != "" {
			out[name] = id
		}
	}
	for name, id := range m.skillIDNames {
		if name == "" || id <= 0 || id > 0xffff {
			continue
		}
		out[name] = uint16(id)
	}
	return out
}

func IsEffectDataFile(name string) bool {
	base := strings.ToLower(path.Base(strings.ReplaceAll(name, "\\", "/")))
	switch {
	case base == "effects.json" || strings.HasPrefix(base, "effects.") && strings.HasSuffix(base, ".json"):
		return true
	case base == "ez2streffect.bson" || strings.HasPrefix(base, "ez2streffect."):
		return true
	default:
		return false
	}
}

// EffectDataLocalFiles returns loose overlay files a development watcher can poll.
func (m *Manager) EffectDataLocalFiles() []string {
	if m == nil {
		return nil
	}
	var out []string
	for _, candidates := range [][]string{effectTablesJSONCandidates, ez2STREffectBSONCandidates} {
		if found, ok := m.FindFirst(candidates); ok {
			out = append(out, found)
		}
	}
	return out
}
