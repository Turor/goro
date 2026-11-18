package rotheme

import (
	"bufio"
	"bytes"
	"embed"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Themes are named palette files that sit under the player's ui-layout.ini.
// A theme is an ordinary layout file (usually a [palette] section and a few
// [theme] references): the values it sets replace the compiled defaults, and
// ui-layout.ini, the F11 tuner and live reload still override it key by key.
// The order, lowest first, is
//
//	compiled default  <  active theme  <  ui-layout.ini
//
// The built-in themes are embedded in the binary (themes/*.ini), so every
// release build has them without extra files. Players can add their own as
// <name>.ini in the themes folder next to goro.ini; they appear in the
// settings window's Theme list after the built-in ones. The choice is saved
// in goro.ini as [ui_layout] theme = <name>.

//go:embed themes/*.ini
var builtinThemeFiles embed.FS

// DefaultThemeID is the built-in light blue theme. Its file sets nothing, so
// it is exactly the compiled defaults.
const DefaultThemeID = "default"

// ThemeFileExt is the extension of a theme file.
const ThemeFileExt = ".ini"

// theme:ok name and list limits, not visual sizes
const (
	maxThemeIDLen   = 40
	maxThemeNameLen = 32
	maxUserThemes   = 64
	themeHeadBytes  = 4 << 10
)

// builtinThemes lists the embedded themes in menu order.
var builtinThemes = []string{DefaultThemeID, "dark", "complementary", "analogous", "triadic"}

// ThemeInfo describes one selectable theme.
type ThemeInfo struct {
	// ID is the file name without .ini, in lower case; goro.ini keeps it.
	ID string
	// Name is the label shown in the settings window: a "# name: ..." line
	// in the comment block at the top of the file, or the ID in title case.
	Name string
	// Builtin themes are embedded; user themes have a Path.
	Builtin bool
	Path    string
}

var (
	themesDir  string
	themeID    = DefaultThemeID
	themeBase  Layout
	userLayout Layout
)

// SetThemesDir sets the folder searched for user theme files. Empty means
// built-in themes only.
func SetThemesDir(dir string) { themesDir = dir }

// ThemesDir is the folder searched for user theme files.
func ThemesDir() string { return themesDir }

// ActiveTheme is the ID of the theme in effect.
func ActiveTheme() string { return themeID }

// base is the value a token takes when the user layout does not set it: the
// active theme's value, or the compiled default.
func (t *token) base() tokenValue {
	if v, ok := themeBase.values[t]; ok {
		return v
	}
	return t.def
}

func validThemeID(id string) bool {
	if id == "" || len(id) > maxThemeIDLen {
		return false
	}
	for _, r := range id {
		if !(r >= 'a' && r <= 'z' || r >= '0' && r <= '9' || r == '_' || r == '-') {
			return false
		}
	}
	return true
}

// NormalizeThemeID lower-cases and trims a theme name; empty means the
// default theme.
func NormalizeThemeID(id string) string {
	id = strings.ToLower(strings.TrimSpace(id))
	if id == "" {
		return DefaultThemeID
	}
	return id
}

// ListThemes returns the built-in themes, then the user theme files in the
// themes folder in ID order. A user file with a built-in theme's name,
// an invalid name, or that is not a regular file is skipped.
func ListThemes() []ThemeInfo {
	out := make([]ThemeInfo, 0, len(builtinThemes))
	seen := map[string]bool{}
	for _, id := range builtinThemes {
		data, err := builtinThemeFiles.ReadFile("themes/" + id + ThemeFileExt)
		if err != nil {
			continue
		}
		out = append(out, ThemeInfo{ID: id, Name: themeDisplayName(data, id), Builtin: true})
		seen[id] = true
	}
	if themesDir == "" {
		return out
	}
	entries, err := os.ReadDir(themesDir)
	if err != nil {
		return out
	}
	var users []ThemeInfo
	for _, entry := range entries {
		name := entry.Name()
		ext := filepath.Ext(name)
		if !strings.EqualFold(ext, ThemeFileExt) {
			continue
		}
		id := strings.ToLower(strings.TrimSuffix(name, ext))
		if !validThemeID(id) || seen[id] {
			continue
		}
		path := filepath.Join(themesDir, name)
		info, err := os.Stat(path)
		if err != nil || !info.Mode().IsRegular() || info.Size() > MaxLayoutFileSize {
			continue
		}
		seen[id] = true
		users = append(users, ThemeInfo{ID: id, Path: path})
	}
	sort.Slice(users, func(i, j int) bool { return users[i].ID < users[j].ID })
	if len(users) > maxUserThemes {
		users = users[:maxUserThemes]
	}
	for i := range users {
		users[i].Name = themeDisplayName(readThemeHead(users[i].Path), users[i].ID)
	}
	return append(out, users...)
}

// CheckBuiltinThemes parses every embedded theme and reports errors and
// warnings (such as a key no build registers). Only meaningful in a binary
// that links every token package.
func CheckBuiltinThemes() []string {
	var problems []string
	for _, id := range builtinThemes {
		_, warnings, err := loadTheme(id)
		if err != nil {
			problems = append(problems, err.Error())
		}
		problems = append(problems, warnings...)
	}
	return problems
}

func findTheme(id string) (ThemeInfo, bool) {
	for _, info := range ListThemes() {
		if info.ID == id {
			return info, true
		}
	}
	return ThemeInfo{}, false
}

func readThemeHead(path string) []byte {
	file, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer file.Close()
	head, _ := io.ReadAll(io.LimitReader(file, themeHeadBytes))
	return head
}

// themeDisplayName reads "# name: Dark" from the comment block at the top of
// a theme file, or title-cases the ID.
func themeDisplayName(data []byte, id string) string {
	scanner := bufio.NewScanner(bytes.NewReader(data))
	first := true
	for scanner.Scan() {
		line := scanner.Text()
		if first {
			line = strings.TrimPrefix(line, "\ufeff")
			first = false
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if !strings.HasPrefix(line, "#") && !strings.HasPrefix(line, ";") {
			break
		}
		key, value, ok := strings.Cut(strings.TrimSpace(line[1:]), ":")
		if !ok || !strings.EqualFold(strings.TrimSpace(key), "name") {
			continue
		}
		if name := cleanThemeName(value); name != "" {
			return name
		}
	}
	return titleThemeID(id)
}

func cleanThemeName(s string) string {
	s = strings.TrimSpace(s)
	if !utf8.ValidString(s) || strings.IndexFunc(s, unicode.IsControl) >= 0 {
		return ""
	}
	if utf8.RuneCountInString(s) > maxThemeNameLen {
		s = string([]rune(s)[:maxThemeNameLen])
	}
	return strings.TrimSpace(s)
}

func titleThemeID(id string) string {
	words := strings.FieldsFunc(id, func(r rune) bool { return r == '_' || r == '-' })
	for i, w := range words {
		words[i] = strings.ToUpper(w[:1]) + w[1:]
	}
	return strings.Join(words, " ")
}

// loadTheme parses a theme without applying it.
func loadTheme(id string) (Layout, []string, error) {
	info, ok := findTheme(id)
	if !ok {
		return Layout{}, nil, fmt.Errorf("unknown theme %q", id)
	}
	if info.Builtin {
		data, err := builtinThemeFiles.ReadFile("themes/" + id + ThemeFileExt)
		if err != nil {
			return Layout{}, nil, err
		}
		return ParseLayout(bytes.NewReader(data), "theme "+id)
	}
	layout, exists, warnings, err := readLayoutFile(info.Path)
	if err == nil && !exists {
		err = fmt.Errorf("theme %q: %s is gone", id, info.Path)
	}
	return layout, warnings, err
}

// SetTheme makes id the active theme and re-applies the current user layout
// on top of it, so ui-layout.ini overrides keep winning. An unknown or
// invalid theme changes nothing. It must run on the UI thread.
func SetTheme(id string) (warnings []string, err error) {
	id = NormalizeThemeID(id)
	layout, warnings, err := loadTheme(id)
	if err != nil {
		return warnings, err
	}
	themeID, themeBase = id, layout
	userLayout.Apply()
	return warnings, nil
}
