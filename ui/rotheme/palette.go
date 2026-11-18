package rotheme

import (
	"fmt"
	"image/color"
	"math"
	"sort"
	"strconv"
	"strings"
)

// The palette is the high-level color theme: a handful of base colors in the
// [palette] section of ui-layout.ini. Every other color setting defaults to a
// reference to one of them, optionally adjusted by modifiers:
//
//	window_border = @accent
//	button_down   = @panel_alt darken(9.8%)
//	panel_border  = @background alpha(0.373)
//
// Changing a palette color therefore re-colors every setting that refers to
// it. An explicit #rrggbb[aa] value still overrides a single setting, and a
// setting may also be given its own reference. With the default palette every
// reference resolves to exactly the built-in color (a test checks each one),
// so nothing changes until the palette does.
//
// Modifiers apply left to right in HSL space, like Sass:
//
//	lighten(N%) / darken(N%)        add or subtract N points of lightness
//	saturate(N%) / desaturate(N%)   add or subtract N points of saturation
//	spin(D)                         rotate the hue by D degrees (-360 to 360)
//	alpha(A)                        set the opacity, 0 to 1 (or 0% to 100%)
//
// Without modifiers, or with only alpha, the base color's bytes are used as
// they are.

const paletteSection = "palette"

const (
	maxColorExprLen  = 160 // theme:ok text length limit
	maxColorExprMods = 8   // theme:ok modifier count limit
)

type colorMod struct {
	op string
	v  float64
}

type colorExpr struct {
	base string
	mods []colorMod
}

// paletteColorDefault and paletteColorCurrent look a palette entry up by name.
func paletteToken(name string) *token {
	t := tokenIndex[paletteSection+"."+normalizeTokenName(name)]
	if t == nil || t.kind != tokenRGBA {
		return nil
	}
	return t
}

// PaletteKeys lists the palette entries in registration order.
func PaletteKeys() []string {
	var out []string
	for _, t := range tokens {
		if t.section == paletteSection {
			out = append(out, t.key)
		}
	}
	return out
}

func isPaletteName(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if !(r >= 'a' && r <= 'z' || r >= '0' && r <= '9' || r == '_') {
			return false
		}
	}
	return true
}

// parseColorExpr parses "@name mod(arg) ..." and checks the palette entry
// exists.
func parseColorExpr(raw string) (colorExpr, error) {
	raw = strings.TrimSpace(raw)
	if len(raw) > maxColorExprLen {
		return colorExpr{}, fmt.Errorf("palette reference is longer than %d bytes", maxColorExprLen)
	}
	if !strings.HasPrefix(raw, "@") {
		return colorExpr{}, fmt.Errorf("palette reference must start with @, got %q", raw)
	}
	rest := raw[1:]
	end := strings.IndexFunc(rest, func(r rune) bool { return r == ' ' || r == '\t' })
	name := rest
	if end >= 0 {
		name, rest = rest[:end], rest[end:]
	} else {
		rest = ""
	}
	name = normalizeTokenName(name)
	if !isPaletteName(name) || paletteToken(name) == nil {
		return colorExpr{}, fmt.Errorf("unknown palette color @%s (palette: %s)", name, strings.Join(PaletteKeys(), ", "))
	}
	e := colorExpr{base: name}
	for {
		rest = strings.TrimSpace(rest)
		if rest == "" {
			break
		}
		open := strings.IndexByte(rest, '(')
		closing := strings.IndexByte(rest, ')')
		if open <= 0 || closing < open {
			return colorExpr{}, fmt.Errorf("modifier must look like name(value), got %q", rest)
		}
		op := strings.ToLower(strings.TrimSpace(rest[:open]))
		arg := strings.TrimSpace(rest[open+1 : closing])
		rest = rest[closing+1:]
		v, err := parseModArg(op, arg)
		if err != nil {
			return colorExpr{}, err
		}
		e.mods = append(e.mods, colorMod{op: op, v: v})
		if len(e.mods) > maxColorExprMods {
			return colorExpr{}, fmt.Errorf("palette reference has more than %d modifiers", maxColorExprMods)
		}
	}
	return e, nil
}

func parseModArg(op, arg string) (float64, error) {
	percent := strings.HasSuffix(arg, "%")
	num := strings.TrimSpace(strings.TrimSuffix(arg, "%"))
	if op == "spin" {
		num = strings.TrimSpace(strings.TrimSuffix(num, "deg"))
	}
	v, err := strconv.ParseFloat(num, 64)
	if err != nil || math.IsNaN(v) || math.IsInf(v, 0) {
		return 0, fmt.Errorf("%s(%s): not a number", op, arg)
	}
	switch op {
	case "lighten", "darken", "saturate", "desaturate":
		if v < 0 || v > 100 { // theme:ok percent range
			return 0, fmt.Errorf("%s(%s): must be 0%% to 100%%", op, arg)
		}
	case "spin":
		if percent || v < -360 || v > 360 { // theme:ok degree range
			return 0, fmt.Errorf("spin(%s): must be -360 to 360 degrees", arg)
		}
	case "alpha":
		if percent {
			v /= 100 // theme:ok percent to fraction
		}
		if v < 0 || v > 1 {
			return 0, fmt.Errorf("alpha(%s): must be 0 to 1 (or 0%% to 100%%)", arg)
		}
	default:
		return 0, fmt.Errorf("unknown modifier %q (use lighten, darken, saturate, desaturate, spin or alpha)", op)
	}
	return v, nil
}

// String is the canonical form written back to ui-layout.ini.
func (e colorExpr) String() string {
	var b strings.Builder
	b.WriteString("@")
	b.WriteString(e.base)
	for _, m := range e.mods {
		v := strconv.FormatFloat(m.v, 'f', -1, 64)
		switch m.op {
		case "spin", "alpha":
			fmt.Fprintf(&b, " %s(%s)", m.op, v)
		default:
			fmt.Fprintf(&b, " %s(%s%%)", m.op, v)
		}
	}
	return b.String()
}

// resolveWith applies the modifiers to the named palette color.
func (e colorExpr) resolveWith(lookup func(name string) color.RGBA) color.RGBA {
	return applyColorMods(lookup(e.base), e.mods)
}

func currentPaletteColor(name string) color.RGBA {
	if t := paletteToken(name); t != nil {
		return *t.rgbaPtr
	}
	return color.RGBA{}
}

func defaultPaletteColor(name string) color.RGBA {
	if t := paletteToken(name); t != nil {
		return t.def.rgba
	}
	return color.RGBA{}
}

func applyColorMods(base color.RGBA, mods []colorMod) color.RGBA {
	h, s, l := rgbToHSL(base)
	a := float64(base.A) / 255 // theme:ok channel scale
	hsl := false
	for _, m := range mods {
		switch m.op {
		case "lighten":
			l, hsl = clampUnit(l+m.v/100), true // theme:ok percent
		case "darken":
			l, hsl = clampUnit(l-m.v/100), true // theme:ok percent
		case "saturate":
			s, hsl = clampUnit(s+m.v/100), true // theme:ok percent
		case "desaturate":
			s, hsl = clampUnit(s-m.v/100), true // theme:ok percent
		case "spin":
			h, hsl = math.Mod(math.Mod(h+m.v, 360)+360, 360), true // theme:ok degrees
		case "alpha":
			a = m.v
		}
	}
	out := base
	if hsl {
		out = hslToRGB(h, s, l)
	}
	out.A = unitByte(a)
	return out
}

func clampUnit(v float64) float64 { return min(max(v, 0), 1) }

func unitByte(v float64) uint8 {
	return uint8(math.Round(clampUnit(v) * 255)) // theme:ok channel scale
}

// rgbToHSL returns hue in degrees and saturation and lightness in 0..1.
func rgbToHSL(c color.RGBA) (h, s, l float64) {
	r, g, b := float64(c.R)/255, float64(c.G)/255, float64(c.B)/255 // theme:ok channel scale
	hi, lo := max(r, g, b), min(r, g, b)
	l = (hi + lo) / 2
	if hi == lo {
		return 0, 0, l
	}
	d := hi - lo
	if l > 0.5 { // theme:ok HSL formula
		s = d / (2 - hi - lo) // theme:ok HSL formula
	} else {
		s = d / (hi + lo)
	}
	switch hi {
	case r:
		h = (g - b) / d
		if g < b {
			h += 6 // theme:ok HSL formula
		}
	case g:
		h = (b-r)/d + 2 // theme:ok HSL formula
	default:
		h = (r-g)/d + 4 // theme:ok HSL formula
	}
	return h * 60, s, l // theme:ok degrees per sector
}

func hslToRGB(h, s, l float64) color.RGBA {
	if s == 0 {
		v := unitByte(l)
		return color.RGBA{R: v, G: v, B: v, A: 255} // theme:ok gray from lightness
	}
	var q float64
	if l < 0.5 { // theme:ok HSL formula
		q = l * (1 + s)
	} else {
		q = l + s - l*s
	}
	p := 2*l - q       // theme:ok HSL formula
	hk := h / 360      // theme:ok degrees
	return color.RGBA{ // theme:ok computed channels
		R: unitByte(hueToRGB(p, q, hk+1.0/3)), // theme:ok HSL formula
		G: unitByte(hueToRGB(p, q, hk)),
		B: unitByte(hueToRGB(p, q, hk-1.0/3)), // theme:ok HSL formula
		A: 255,                                // theme:ok opaque
	}
}

func hueToRGB(p, q, t float64) float64 {
	if t < 0 {
		t++
	}
	if t > 1 {
		t--
	}
	switch {
	case t < 1.0/6: // theme:ok HSL formula
		return p + (q-p)*6*t // theme:ok HSL formula
	case t < 0.5: // theme:ok HSL formula
		return q
	case t < 2.0/3: // theme:ok HSL formula
		return p + (q-p)*(2.0/3-t)*6 // theme:ok HSL formula
	}
	return p
}

// tokenRaw returns a color token value as 8-bit RGBA.
func tokenRawRGBA(t *token, v tokenValue) color.RGBA {
	if t.kind == tokenColor {
		r, g, b, a := colorBytes(v.c)
		return color.RGBA{R: r, G: g, B: b, A: a} // theme:ok converted channel bytes
	}
	return v.rgba
}

func isColorToken(t *token) bool { return t.kind == tokenColor || t.kind == tokenRGBA }

// SuggestPaletteRef returns the shortest reference to palette color base that
// reproduces target exactly with the default palette, trying fewer decimals
// first. ok is false if no exact reference was found.
func SuggestPaletteRef(base string, target color.RGBA) (string, bool) {
	t := paletteToken(base)
	if t == nil {
		return "", false
	}
	from := t.def.rgba
	if from.R == target.R && from.G == target.G && from.B == target.B {
		e := colorExpr{base: t.key}
		if target.A != from.A {
			for decimals := 2; decimals <= 6; decimals++ { // theme:ok decimal search
				e.mods = []colorMod{{op: "alpha", v: roundTo(float64(target.A)/255, decimals)}} // theme:ok channel scale
				if applyColorMods(from, e.mods) == target {
					break
				}
			}
		}
		return e.String(), applyColorMods(from, e.mods) == target
	}
	h0, s0, l0 := rgbToHSL(from)
	h1, s1, l1 := rgbToHSL(target)
	dh := math.Mod(h1-h0+540, 360) - 180           // theme:ok shortest hue turn
	for decimals := 0; decimals <= 6; decimals++ { // theme:ok decimal search
		var mods []colorMod
		if s1 > 0 && l1 > 0 && l1 < 1 {
			if v := roundTo(dh, decimals); v != 0 {
				mods = append(mods, colorMod{op: "spin", v: v})
			}
		}
		switch ds := roundTo((s1-s0)*100, decimals); { // theme:ok percent
		case s1 == 0 && s0 > 0:
			mods = append(mods, colorMod{op: "desaturate", v: 100}) // theme:ok full desaturation
		case ds > 0:
			mods = append(mods, colorMod{op: "saturate", v: ds})
		case ds < 0:
			mods = append(mods, colorMod{op: "desaturate", v: -ds})
		}
		switch dl := roundTo((l1-l0)*100, decimals); { // theme:ok percent
		case dl > 0:
			mods = append(mods, colorMod{op: "lighten", v: dl})
		case dl < 0:
			mods = append(mods, colorMod{op: "darken", v: -dl})
		}
		if target.A != from.A {
			mods = append(mods, colorMod{op: "alpha", v: roundTo(float64(target.A)/255, max(decimals, 3))}) // theme:ok channel scale
		}
		if applyColorMods(from, mods) == target {
			return colorExpr{base: t.key, mods: mods}.String(), true
		}
	}
	return "", false
}

func roundTo(v float64, decimals int) float64 {
	p := math.Pow(10, float64(decimals)) // theme:ok decimal places
	return math.Round(v*p) / p
}

// NearestPaletteColor names the palette entry closest to c, for a setting
// that has no reference yet.
func NearestPaletteColor(c color.RGBA) string {
	best, bestD := "", math.Inf(1)
	h1, s1, l1 := rgbToHSL(c)
	for _, key := range PaletteKeys() {
		h0, s0, l0 := rgbToHSL(defaultPaletteColor(key))
		dh := math.Abs(math.Mod(h1-h0+540, 360)-180) / 180 * min(s0, s1) // theme:ok hue weight
		d := dh + math.Abs(s1-s0)/2 + math.Abs(l1-l0)*2                  // theme:ok distance weights
		if d < bestD {
			best, bestD = key, d
		}
	}
	return best
}

// PaletteRef is one color setting's default palette reference.
type PaletteRef struct {
	Section, Key string
	Ref          string
	Color        color.RGBA // compiled default
}

// ColorTokenDefaults lists every non-palette color setting with its compiled
// default and its default reference ("" when it has none), sorted by id.
func ColorTokenDefaults() []PaletteRef {
	var out []PaletteRef
	for _, t := range tokens {
		if !isColorToken(t) || t.section == paletteSection {
			continue
		}
		out = append(out, PaletteRef{Section: t.section, Key: t.key, Ref: t.def.expr, Color: tokenRawRGBA(t, t.def)})
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].Section+"."+out[i].Key < out[j].Section+"."+out[j].Key
	})
	return out
}

// CheckPaletteDefaults reports every color setting whose default reference is
// missing, malformed, or does not reproduce its built-in color exactly with
// the default palette.
func CheckPaletteDefaults() []string {
	var problems []string
	for _, ref := range ColorTokenDefaults() {
		id := ref.Section + "." + ref.Key
		if ref.Ref == "" {
			base := NearestPaletteColor(ref.Color)
			suggestion, _ := SuggestPaletteRef(base, ref.Color)
			problems = append(problems, fmt.Sprintf("%s has no palette reference in paletteRefs (suggested: %q)", id, suggestion))
			continue
		}
		e, err := parseColorExpr(ref.Ref)
		if err != nil {
			problems = append(problems, fmt.Sprintf("%s: %v", id, err))
			continue
		}
		if got := e.resolveWith(defaultPaletteColor); got != ref.Color {
			problems = append(problems, fmt.Sprintf("%s: %s resolves to %s, built-in is %s", id, ref.Ref,
				formatHex(got.R, got.G, got.B, got.A), formatHex(ref.Color.R, ref.Color.G, ref.Color.B, ref.Color.A)))
		}
	}
	return problems
}

// UnusedPaletteRefs lists paletteRefs entries with no registered color
// setting. Only meaningful in a binary that links every token package.
func UnusedPaletteRefs() []string {
	var out []string
	for id := range paletteRefs {
		if t := tokenIndex[id]; t == nil || !isColorToken(t) {
			out = append(out, id)
		}
	}
	sort.Strings(out)
	return out
}

// PaletteBase names the palette color a setting's current value follows, or
// "" for an explicit color.
func PaletteBase(section, key string) string {
	t := tokenIndex[tokenID(section, key)]
	if t == nil {
		return ""
	}
	v := t.value()
	if v.expr == "" {
		return ""
	}
	e, err := parseColorExpr(v.expr)
	if err != nil {
		return ""
	}
	return e.base
}
