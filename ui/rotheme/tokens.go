package rotheme

import (
	"bufio"
	"errors"
	"fmt"
	"image/color"
	"io"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync/atomic"
	"unicode"
	"unicode/utf8"

	"github.com/gogpu/ui/widget"
)

// Layout tokens are the UI values that can be overridden from a layout file
// (ui-layout.ini) without recompiling: theme colors, text size, window frame
// geometry, and per-window sizes. Each token is a pointer to the package
// variable the UI already reads, plus the compiled default. With no override
// file nothing is parsed and every token keeps its compiled default, so release
// players see exactly the built-in UI.
//
// A file is applied atomically: it is parsed and validated in full first, and
// only a file without errors replaces the current values. Keys the file omits
// return to their compiled defaults, so deleting a line undoes it.

type tokenKind int

const (
	tokenInt tokenKind = iota
	tokenFloat
	tokenColor
	tokenRGBA
	tokenString
	tokenFloat64
)

type token struct {
	section string
	key     string
	doc     string
	kind    tokenKind
	min     float64
	max     float64

	intPtr   *int
	floatPtr *float32
	f64Ptr   *float64
	colorPtr *widget.Color
	rgbaPtr  *color.RGBA
	strPtr   *string
	// allowEmpty lets a string token be blank (for example "no font file").
	allowEmpty bool

	def tokenValue
	// src is the value in effect as written (a palette reference keeps its
	// expression); applied is what the last Apply stored in the variable.
	src     tokenValue
	applied tokenValue
}

type tokenValue struct {
	i    int
	f    float32
	d    float64
	c    widget.Color
	rgba color.RGBA
	s    string
	// expr is a palette reference ("@panel darken(5%)") for color tokens;
	// the color fields are then resolved by Apply.
	expr string
}

var (
	tokens     []*token
	tokenIndex = map[string]*token{}
	derived    []func()
	generation atomic.Uint64
)

// OnLayoutApplied registers fn to recompute values derived from tokens (for
// example a window height that sums several tokens). It runs after every
// apply, on the UI thread.
func OnLayoutApplied(fn func()) {
	derived = append(derived, fn)
}

func tokenID(section, key string) string {
	return normalizeTokenName(section) + "." + normalizeTokenName(key)
}

func sortStrings(s []string) { sort.Strings(s) }

func normalizeTokenName(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

func register(t *token) {
	t.section = normalizeTokenName(t.section)
	t.key = normalizeTokenName(t.key)
	id := t.section + "." + t.key
	if _, dup := tokenIndex[id]; dup {
		panic("rotheme: duplicate layout token " + id)
	}
	t.def = t.current()
	if isColorToken(t) && t.section != paletteSection {
		t.def.expr = paletteRefs[id]
	}
	t.src, t.applied = t.def, t.current()
	tokenIndex[id] = t
	tokens = append(tokens, t)
}

// RegisterInt exposes an integer pixel value. min and max bound valid edits.
func RegisterInt(section, key string, ptr *int, minValue, maxValue int, doc string) {
	register(&token{section: section, key: key, doc: doc, kind: tokenInt, intPtr: ptr, min: float64(minValue), max: float64(maxValue)})
}

// RegisterFloat exposes a float32 value such as a size, padding or alpha.
func RegisterFloat(section, key string, ptr *float32, minValue, maxValue float64, doc string) {
	register(&token{section: section, key: key, doc: doc, kind: tokenFloat, floatPtr: ptr, min: minValue, max: maxValue})
}

// RegisterFloat64 exposes a float64 value (a scale or ratio the code uses as
// float64), so its compiled default stays exact instead of passing through
// float32.
func RegisterFloat64(section, key string, ptr *float64, minValue, maxValue float64, doc string) {
	register(&token{section: section, key: key, doc: doc, kind: tokenFloat64, f64Ptr: ptr, min: minValue, max: maxValue})
}

// RegisterColor exposes a gogpu/ui color, written as #rrggbb or #rrggbbaa.
func RegisterColor(section, key string, ptr *widget.Color, doc string) {
	register(&token{section: section, key: key, doc: doc, kind: tokenColor, colorPtr: ptr})
}

// RegisterRGBA exposes an image/color RGBA, written as #rrggbb or #rrggbbaa.
func RegisterRGBA(section, key string, ptr *color.RGBA, doc string) {
	register(&token{section: section, key: key, doc: doc, kind: tokenRGBA, rgbaPtr: ptr})
}

// RegisterString exposes a short text value such as a font family name. It is
// at most maxLen bytes, has no control characters, and is blank only if
// allowEmpty.
func RegisterString(section, key string, ptr *string, maxLen int, allowEmpty bool, doc string) {
	register(&token{section: section, key: key, doc: doc, kind: tokenString, strPtr: ptr, max: float64(maxLen), allowEmpty: allowEmpty})
}

// Generation increases every time layout values are replaced. Widgets that
// copy token values at construction compare it to know when to refresh.
func Generation() uint64 {
	return generation.Load()
}

func (t *token) current() tokenValue {
	switch t.kind {
	case tokenInt:
		return tokenValue{i: *t.intPtr}
	case tokenFloat:
		return tokenValue{f: *t.floatPtr}
	case tokenFloat64:
		return tokenValue{d: *t.f64Ptr}
	case tokenColor:
		return tokenValue{c: *t.colorPtr}
	case tokenString:
		return tokenValue{s: *t.strPtr}
	default:
		return tokenValue{rgba: *t.rgbaPtr}
	}
}

func (t *token) set(v tokenValue) {
	switch t.kind {
	case tokenInt:
		*t.intPtr = v.i
	case tokenFloat:
		*t.floatPtr = v.f
	case tokenFloat64:
		*t.f64Ptr = v.d
	case tokenColor:
		*t.colorPtr = v.c
	case tokenString:
		*t.strPtr = v.s
	default:
		*t.rgbaPtr = v.rgba
	}
}

func (t *token) format(v tokenValue) string {
	if v.expr != "" {
		return v.expr
	}
	switch t.kind {
	case tokenInt:
		return strconv.Itoa(v.i)
	case tokenFloat:
		return strconv.FormatFloat(float64(v.f), 'f', -1, 32)
	case tokenFloat64:
		return strconv.FormatFloat(v.d, 'f', -1, 64)
	case tokenColor:
		r, g, b, a := colorBytes(v.c)
		return formatHex(r, g, b, a)
	case tokenString:
		return v.s
	default:
		return formatHex(v.rgba.R, v.rgba.G, v.rgba.B, v.rgba.A)
	}
}

func (t *token) parse(raw string) (tokenValue, error) {
	switch t.kind {
	case tokenInt:
		n, err := strconv.Atoi(raw)
		if err != nil {
			return tokenValue{}, fmt.Errorf("%s.%s must be a whole number, got %q", t.section, t.key, raw)
		}
		if float64(n) < t.min || float64(n) > t.max {
			return tokenValue{}, fmt.Errorf("%s.%s must be between %g and %g, got %d", t.section, t.key, t.min, t.max, n)
		}
		return tokenValue{i: n}, nil
	case tokenFloat, tokenFloat64:
		bits := 32
		if t.kind == tokenFloat64 {
			bits = 64
		}
		f, err := strconv.ParseFloat(raw, bits)
		if err != nil || math.IsNaN(f) || math.IsInf(f, 0) {
			return tokenValue{}, fmt.Errorf("%s.%s must be a finite number, got %q", t.section, t.key, raw)
		}
		if f < t.min || f > t.max {
			return tokenValue{}, fmt.Errorf("%s.%s must be between %g and %g, got %s", t.section, t.key, t.min, t.max, raw)
		}
		if t.kind == tokenFloat64 {
			return tokenValue{d: f}, nil
		}
		return tokenValue{f: float32(f)}, nil
	case tokenString:
		return t.parseString(raw)
	default:
		if strings.HasPrefix(raw, "@") {
			if t.section == paletteSection {
				return tokenValue{}, fmt.Errorf("%s.%s: palette colors must be #rrggbb or #rrggbbaa", t.section, t.key)
			}
			e, err := parseColorExpr(raw)
			if err != nil {
				return tokenValue{}, fmt.Errorf("%s.%s: %w", t.section, t.key, err)
			}
			return tokenValue{expr: e.String()}, nil
		}
		r, g, b, a, err := parseHexColor(raw)
		if err != nil {
			return tokenValue{}, fmt.Errorf("%s.%s: %w", t.section, t.key, err)
		}
		if t.kind == tokenColor {
			return tokenValue{c: widget.RGBA8(r, g, b, a)}, nil // theme:ok parsed channel bytes
		}
		return tokenValue{rgba: color.RGBA{R: r, G: g, B: b, A: a}}, nil // theme:ok parsed channel bytes
	}
}

func (t *token) parseString(raw string) (tokenValue, error) {
	if raw == "" && !t.allowEmpty {
		return tokenValue{}, fmt.Errorf("%s.%s must not be empty", t.section, t.key)
	}
	if len(raw) > int(t.max) {
		return tokenValue{}, fmt.Errorf("%s.%s is longer than %d bytes", t.section, t.key, int(t.max))
	}
	if !utf8.ValidString(raw) || strings.IndexFunc(raw, unicode.IsControl) >= 0 {
		return tokenValue{}, fmt.Errorf("%s.%s has invalid characters", t.section, t.key)
	}
	return tokenValue{s: raw}, nil
}

func parseHexColor(raw string) (r, g, b, a uint8, err error) {
	s := strings.TrimPrefix(raw, "#")
	if len(s) != 6 && len(s) != 8 {
		return 0, 0, 0, 0, fmt.Errorf("color must be #rrggbb or #rrggbbaa, got %q", raw)
	}
	n, perr := strconv.ParseUint(s, 16, 32)
	if perr != nil {
		return 0, 0, 0, 0, fmt.Errorf("color must be #rrggbb or #rrggbbaa, got %q", raw)
	}
	if len(s) == 6 {
		return uint8(n >> 16), uint8(n >> 8), uint8(n), 255, nil
	}
	return uint8(n >> 24), uint8(n >> 16), uint8(n >> 8), uint8(n), nil
}

func colorBytes(c widget.Color) (r, g, b, a uint8) {
	to8 := func(v float32) uint8 {
		return uint8(math.Round(float64(min(max(v, 0), 1)) * 255))
	}
	return to8(c.R), to8(c.G), to8(c.B), to8(c.A)
}

func formatHex(r, g, b, a uint8) string {
	if a == 255 {
		return fmt.Sprintf("#%02x%02x%02x", r, g, b)
	}
	return fmt.Sprintf("#%02x%02x%02x%02x", r, g, b, a)
}

const (
	// MaxLayoutFileSize bounds what a layout file may cost at startup. The
	// full default dump (every window and screen) is well under half of it.
	MaxLayoutFileSize = 256 << 10 // theme:ok file size limit
	// MaxLayoutWarnings bounds unknown-key warnings; the rest are counted.
	MaxLayoutWarnings = 20 // theme:ok warning count limit
)

// Layout is a validated set of overrides, keyed by token, ready to apply.
type Layout struct {
	values map[*token]tokenValue
}

// LayoutError is a rejected layout file. The current values are unchanged.
type LayoutError struct {
	Path string
	Line int
	Msg  string
}

func (e *LayoutError) Error() string {
	where := e.Path
	if where == "" {
		where = "ui layout"
	}
	if e.Line > 0 {
		return fmt.Sprintf("%s:%d: %s", where, e.Line, e.Msg)
	}
	return where + ": " + e.Msg
}

// ParseLayout reads a layout file. Unknown keys are returned as warnings so a
// file written for a newer or older build still applies; malformed lines and
// out-of-range values are errors and nothing is applied.
func ParseLayout(r io.Reader, path string) (Layout, []string, error) {
	layout := Layout{values: map[*token]tokenValue{}}
	var warnings []string
	unknown := 0
	scanner := bufio.NewScanner(r)
	section := ""
	lineNo := 0
	for scanner.Scan() {
		lineNo++
		line := strings.TrimSpace(scanner.Text())
		if lineNo == 1 {
			line = strings.TrimPrefix(line, "\ufeff")
		}
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}
		if strings.HasPrefix(line, "[") {
			if !strings.HasSuffix(line, "]") {
				return Layout{}, warnings, &LayoutError{Path: path, Line: lineNo, Msg: "section header must look like [name]"}
			}
			section = normalizeTokenName(line[1 : len(line)-1])
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			return Layout{}, warnings, &LayoutError{Path: path, Line: lineNo, Msg: "expected key = value"}
		}
		value = strings.TrimSpace(value)
		if i := strings.Index(value, " #"); i >= 0 {
			value = strings.TrimSpace(value[:i])
		}
		value = strings.Trim(value, `"'`)
		t := tokenIndex[tokenID(section, key)]
		if t == nil {
			unknown++
			if unknown > MaxLayoutWarnings {
				continue
			}
			warnings = append(warnings, (&LayoutError{Path: path, Line: lineNo, Msg: fmt.Sprintf("unknown key [%s] %s (ignored)", section, strings.TrimSpace(key))}).Error())
			continue
		}
		v, err := t.parse(value)
		if err != nil {
			return Layout{}, warnings, &LayoutError{Path: path, Line: lineNo, Msg: err.Error()}
		}
		layout.values[t] = v
	}
	if unknown > MaxLayoutWarnings {
		warnings = append(warnings, fmt.Sprintf("%d unknown keys in total; only the first %d are listed", unknown, MaxLayoutWarnings))
	}
	if err := scanner.Err(); err != nil {
		return Layout{}, warnings, &LayoutError{Path: path, Msg: err.Error()}
	}
	return layout, warnings, nil
}

// Len reports how many tokens the layout overrides.
func (l Layout) Len() int { return len(l.values) }

// Apply replaces every token: overridden ones take the file value and the rest
// return to the active theme's value, or their compiled default when the theme
// does not set them (see themes.go). The palette is applied first, then every
// palette reference is resolved against it. The layout is remembered as the
// user layout, so switching theme later re-applies it on top of the new theme.
// It must run on the UI thread.
func (l Layout) Apply() {
	userLayout = l
	for _, palette := range []bool{true, false} {
		for _, t := range tokens {
			if (t.section == paletteSection) != palette {
				continue
			}
			src, ok := l.values[t]
			if !ok {
				src = t.base()
			}
			v := t.resolve(src)
			t.set(v)
			t.src, t.applied = src, v
		}
	}
	for _, fn := range derived {
		fn()
	}
	generation.Add(1)
}

// resolve turns a source value into the value stored in the variable: a
// palette reference becomes a color from the current palette. A reference
// that yields the built-in color's bytes stores the built-in value itself.
func (t *token) resolve(src tokenValue) tokenValue {
	if src.expr == "" {
		return src
	}
	e, err := parseColorExpr(src.expr)
	if err != nil {
		// Only validated expressions reach here; keep the default if the
		// palette entry disappeared.
		def := t.def
		def.expr = ""
		return def
	}
	c := e.resolveWith(currentPaletteColor)
	if c == tokenRawRGBA(t, t.def) {
		def := t.def
		def.expr = ""
		return def
	}
	if t.kind == tokenColor {
		return tokenValue{c: fromRGBA(c)}
	}
	return tokenValue{rgba: c}
}

// value is the token's current value for display and saving: its source
// (keeping a palette reference) unless code changed the variable since the
// last Apply.
func (t *token) value() tokenValue {
	if cur := t.current(); cur != t.applied {
		return cur
	}
	return t.src
}

// sameSource reports whether two values are the same setting: the same
// palette reference, or the same explicit value.
func sameSource(a, b tokenValue) bool {
	if a.expr != "" || b.expr != "" {
		return a.expr == b.expr
	}
	return a == b
}

// ResetLayout drops every user override: the active theme and, below it, the
// compiled defaults apply.
func ResetLayout() {
	Layout{}.Apply()
}

// LoadLayoutFile parses path and, if it is valid, applies it. A missing file
// is not an error: it reports applied=false and leaves values untouched.
// Only regular files up to MaxLayoutFileSize are read, so a FIFO or device at
// the path cannot block startup and a huge file cannot exhaust memory.
func LoadLayoutFile(path string) (applied bool, warnings []string, err error) {
	layout, exists, warnings, err := readLayoutFile(path)
	if err != nil || !exists {
		return false, warnings, err
	}
	layout.Apply()
	return true, warnings, nil
}

// readLayoutFile parses path without applying it. A missing file reports
// exists=false and no error.
func readLayoutFile(path string) (layout Layout, exists bool, warnings []string, err error) {
	info, err := os.Stat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Layout{}, false, nil, nil
		}
		return Layout{}, false, nil, &LayoutError{Path: path, Msg: err.Error()}
	}
	if !info.Mode().IsRegular() {
		return Layout{}, false, nil, &LayoutError{Path: path, Msg: "not a regular file"}
	}
	if info.Size() > MaxLayoutFileSize {
		return Layout{}, false, nil, &LayoutError{Path: path, Msg: fmt.Sprintf("file is %d bytes; the limit is %d", info.Size(), MaxLayoutFileSize)}
	}
	file, err := os.Open(path)
	if err != nil {
		return Layout{}, false, nil, &LayoutError{Path: path, Msg: err.Error()}
	}
	defer file.Close()
	// The file may grow between Stat and Open: never read past the limit.
	limited := &io.LimitedReader{R: file, N: MaxLayoutFileSize + 1}
	layout, warnings, err = ParseLayout(limited, path)
	if err != nil {
		return Layout{}, false, warnings, err
	}
	if limited.N <= 0 {
		return Layout{}, false, nil, &LayoutError{Path: path, Msg: fmt.Sprintf("file is over the %d byte limit", MaxLayoutFileSize)}
	}
	return layout, true, warnings, nil
}

// WriteLayout writes every token with its current value and documentation, so
// the output is a complete, valid layout file that names every key.
func WriteLayout(w io.Writer) error {
	return writeLayout(w, func(t *token) (tokenValue, tokenValue) { return t.value(), t.current() })
}

// writeLayout writes value(t) for every token; resolved is the color a
// palette reference stands for, noted in its comment.
func writeLayout(w io.Writer, value func(t *token) (src, resolved tokenValue)) error {
	sections := map[string][]*token{}
	var order []string
	for _, t := range tokens {
		if _, ok := sections[t.section]; !ok {
			order = append(order, t.section)
		}
		sections[t.section] = append(sections[t.section], t)
	}
	sort.Slice(order, func(i, j int) bool {
		// The palette comes first: it is what most themes change.
		if (order[i] == paletteSection) != (order[j] == paletteSection) {
			return order[i] == paletteSection
		}
		return order[i] < order[j]
	})
	bw := bufio.NewWriter(w)
	fmt.Fprintln(bw, "# goro UI layout and theme. Generated by goro --dump-ui-layout.")
	fmt.Fprintln(bw, "# Copy the lines you want to change into ui-layout.ini next to goro.ini.")
	fmt.Fprintln(bw, "# Omitted keys keep the built-in default. Colors are #rrggbb or #rrggbbaa, or a")
	fmt.Fprintln(bw, "# [palette] reference such as @panel, @accent darken(15%) or @bad alpha(0.8):")
	fmt.Fprintln(bw, "# change a palette color and every setting that refers to it follows.")
	fmt.Fprintln(bw, "# With --ui-dev (or [ui_layout] watch = true) edits apply while the game runs; F10 reloads.")
	for _, section := range order {
		fmt.Fprintf(bw, "\n[%s]\n", section)
		if doc := sectionDocs[section]; doc != "" {
			fmt.Fprintf(bw, "# %s\n\n", doc)
		}
		for _, t := range sections[section] {
			doc := t.doc
			if t.kind == tokenInt || t.kind == tokenFloat || t.kind == tokenFloat64 {
				doc = fmt.Sprintf("%s (%g to %g)", doc, t.min, t.max)
			}
			src, resolved := value(t)
			if src.expr != "" {
				doc = fmt.Sprintf("%s (= %s)", doc, t.format(tokenValue{c: resolved.c, rgba: resolved.rgba}))
			}
			if doc != "" {
				fmt.Fprintf(bw, "# %s\n", doc)
			}
			if value := t.format(src); value != "" {
				fmt.Fprintf(bw, "%s = %s\n", t.key, value)
			} else {
				fmt.Fprintf(bw, "%s =\n", t.key)
			}
		}
	}
	return bw.Flush()
}

// WriteDefaultLayout writes the compiled defaults, whatever is applied now.
func WriteDefaultLayout(w io.Writer) error {
	return writeLayout(w, func(t *token) (tokenValue, tokenValue) {
		raw := t.def
		raw.expr = ""
		return t.def, raw
	})
}
