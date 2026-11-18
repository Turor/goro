package rotheme

import (
	"bufio"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"strconv"
)

// TokenInfo describes one layout token for the in-game tuning panel.
type TokenInfo struct {
	Section, Key, Doc string
	// Numeric tokens step with +/-; the rest (colors, names) are typed.
	Numeric, Integer bool
	Min, Max         float64
	Value, Default   string
}

// Tokens lists every token in registration order with its current value.
func Tokens() []TokenInfo {
	out := make([]TokenInfo, 0, len(tokens))
	for _, t := range tokens {
		out = append(out, TokenInfo{
			Section: t.section, Key: t.key, Doc: t.doc,
			Numeric: t.kind == tokenInt || t.kind == tokenFloat,
			Integer: t.kind == tokenInt,
			Min:     t.min, Max: t.max,
			Value: t.format(t.current()), Default: t.format(t.def),
		})
	}
	return out
}

// currentLayout captures every value that differs from its default.
func currentLayout() Layout {
	layout := Layout{values: map[*token]tokenValue{}}
	for _, t := range tokens {
		if v := t.current(); v != t.def {
			layout.values[t] = v
		}
	}
	return layout
}

// SetToken validates raw for section.key and applies it, keeping every other
// current value. A rejected value changes nothing.
func SetToken(section, key, raw string) error {
	t := tokenIndex[tokenID(section, key)]
	if t == nil {
		return fmt.Errorf("unknown key [%s] %s", section, key)
	}
	v, err := t.parse(raw)
	if err != nil {
		return err
	}
	layout := currentLayout()
	layout.values[t] = v
	layout.Apply()
	return nil
}

// StepToken moves a numeric token by steps of its natural increment (1, or
// 0.05 for 0..1 values such as alpha), clamped to its range.
func StepToken(info TokenInfo, steps int) error {
	if !info.Numeric {
		return fmt.Errorf("[%s] %s is not a number", info.Section, info.Key)
	}
	value, err := strconv.ParseFloat(info.Value, 64)
	if err != nil {
		return err
	}
	step := 1.0
	if !info.Integer && info.Max <= 1 {
		step = 0.05
	}
	value = math.Round((value+float64(steps)*step)*100) / 100
	value = min(max(value, info.Min), info.Max)
	raw := strconv.FormatFloat(value, 'f', -1, 32)
	if info.Integer {
		raw = strconv.Itoa(int(math.Round(value)))
	}
	return SetToken(info.Section, info.Key, raw)
}

// WriteOverrides writes only the values that differ from the defaults, so a
// saved file stays short and later default changes still reach the player.
func WriteOverrides(w io.Writer) error {
	layout := currentLayout()
	bw := bufio.NewWriter(w)
	fmt.Fprintln(bw, "# goro UI layout, saved by the in-game tuner (--ui-dev, F11).")
	fmt.Fprintln(bw, "# Only changed values are listed; see docs/ui-layout.default.ini for every key.")
	section := ""
	for _, t := range tokens {
		v, ok := layout.values[t]
		if !ok {
			continue
		}
		if t.section != section {
			section = t.section
			fmt.Fprintf(bw, "\n[%s]\n", section)
		}
		fmt.Fprintf(bw, "%s = %s\n", t.key, t.format(v))
	}
	return bw.Flush()
}

// SaveOverrides writes WriteOverrides to path through a temporary file and a
// rename, so a reader (the --ui-dev watcher) never sees a half-written file.
func SaveOverrides(path string) error {
	tmp, err := os.CreateTemp(filepath.Dir(path), ".ui-layout-*.tmp")
	if err != nil {
		return err
	}
	name := tmp.Name()
	defer os.Remove(name)
	if err := tmp.Chmod(0o644); err != nil {
		tmp.Close()
		return err
	}
	if err := WriteOverrides(tmp); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return replaceFile(name, path)
}

// replaceFile renames src over dst. Go's os.Rename already replaces an
// existing file on Windows (MoveFileEx with MOVEFILE_REPLACE_EXISTING), but it
// can still fail there while another process holds dst open without
// FILE_SHARE_DELETE; then remove dst and retry once.
func replaceFile(src, dst string) error {
	err := os.Rename(src, dst)
	if err == nil {
		return nil
	}
	if _, statErr := os.Stat(dst); statErr != nil {
		return err
	}
	if removeErr := os.Remove(dst); removeErr != nil {
		return err
	}
	return os.Rename(src, dst)
}
