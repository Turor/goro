package res

import (
	"strings"
)

func (m *Manager) MapDisplayName(mapName string) (string, bool) {
	if m == nil {
		return "", false
	}
	m.loadMapDisplayNames()
	key := normalizeRswNameForCameraTable(mapName)
	if key == "" {
		return "", false
	}
	name, ok := m.mapDisplayNames[key]
	return name, ok && name != ""
}

func (m *Manager) loadMapDisplayNames() {
	if m.mapDisplayNamesLoaded {
		return
	}
	m.mapDisplayNamesLoaded = true
	m.mapDisplayNames = make(map[string]string)
	_, data, ok := m.ReadFirst([]string{
		"mapnametable.txt",
		"data\\mapnametable.txt",
		"data/mapnametable.txt",
	})
	if !ok {
		return
	}
	for key, name := range parseMapNameTable(data) {
		m.mapDisplayNames[key] = name
	}
}

func parseMapNameTable(data []byte) map[string]string {
	out := make(map[string]string)
	for _, rawLine := range strings.Split(decodeItemTableText(data), "\n") {
		line := strings.TrimSpace(strings.TrimRight(rawLine, "\r"))
		if line == "" || strings.HasPrefix(line, "//") || strings.HasPrefix(line, "/") {
			continue
		}
		firstHash := strings.IndexByte(line, '#')
		if firstHash <= 0 {
			continue
		}
		secondHash := strings.IndexByte(line[firstHash+1:], '#')
		if secondHash < 0 {
			continue
		}
		secondHash += firstHash + 1
		key := normalizeRswNameForCameraTable(line[:firstHash])
		name := strings.TrimSpace(line[firstHash+1 : secondHash])
		if key == "" || name == "" {
			continue
		}
		if _, exists := out[key]; !exists {
			out[key] = name
		}
	}
	return out
}
