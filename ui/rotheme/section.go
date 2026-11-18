package rotheme

import (
	"image/color"

	"github.com/gogpu/ui/widget"
)

// LayoutSection registers the layout tokens of one ui-layout.ini section,
// usually one window or screen ([shop], [trade], [login]). Each window keeps
// its defaults table in a theme_<section>.go file next to it: the compiled
// default of every size, offset, padding, row height and color lives there
// and nowhere else, so ui-layout.ini can change all of it at run time.
type LayoutSection struct {
	name string
}

var sectionDocs = map[string]string{}

// Section returns the registrar for section, with a one-line doc written above
// the section in --dump-ui-layout and shown by the F11 tuner.
func Section(name, doc string) LayoutSection {
	name = normalizeTokenName(name)
	if doc != "" {
		if prev, ok := sectionDocs[name]; ok && prev != doc {
			panic("rotheme: layout section " + name + " documented twice")
		}
		sectionDocs[name] = doc
	}
	return LayoutSection{name: name}
}

// SectionDoc returns the doc given to Section, if any.
func SectionDoc(name string) string {
	return sectionDocs[normalizeTokenName(name)]
}

// Name is the section's ui-layout.ini header name.
func (s LayoutSection) Name() string { return s.name }

// Int registers an integer pixel value or count.
func (s LayoutSection) Int(key string, ptr *int, minValue, maxValue int, doc string) {
	RegisterInt(s.name, key, ptr, minValue, maxValue, doc)
}

// Float registers a float32 size, padding or alpha.
func (s LayoutSection) Float(key string, ptr *float32, minValue, maxValue float64, doc string) {
	RegisterFloat(s.name, key, ptr, minValue, maxValue, doc)
}

// Color registers a gogpu/ui color.
func (s LayoutSection) Color(key string, ptr *widget.Color, doc string) {
	RegisterColor(s.name, key, ptr, doc)
}

// RGBA registers an image/color RGBA color.
func (s LayoutSection) RGBA(key string, ptr *color.RGBA, doc string) {
	RegisterRGBA(s.name, key, ptr, doc)
}

// String registers a short text value.
func (s LayoutSection) String(key string, ptr *string, maxLen int, allowEmpty bool, doc string) {
	RegisterString(s.name, key, ptr, maxLen, allowEmpty, doc)
}

// Sections lists every section name that has tokens, sorted.
func Sections() []string {
	seen := map[string]bool{}
	var out []string
	for _, t := range tokens {
		if !seen[t.section] {
			seen[t.section] = true
			out = append(out, t.section)
		}
	}
	sortStrings(out)
	return out
}
