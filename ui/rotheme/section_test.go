package rotheme

import (
	"strings"
	"testing"
)

func TestSectionRegistersTokensWithDoc(t *testing.T) {
	saveTokens := tokens
	saveIndex := tokenIndex
	saveDocs := sectionDocs
	t.Cleanup(func() { tokens, tokenIndex, sectionDocs = saveTokens, saveIndex, saveDocs })
	tokens, tokenIndex, sectionDocs = nil, map[string]*token{}, map[string]string{}

	width := 420
	s := Section("Shop", "NPC shop windows")
	s.Int("width", &width, 100, 900, "shop width")
	if got := SectionDoc("shop"); got != "NPC shop windows" {
		t.Fatalf("section doc = %q", got)
	}
	if got := Sections(); len(got) != 1 || got[0] != "shop" {
		t.Fatalf("sections = %v", got)
	}
	var b strings.Builder
	if err := WriteLayout(&b); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(b.String(), "[shop]\n# NPC shop windows\n\n# shop width (100 to 900)\nwidth = 420\n") {
		t.Fatalf("dump missing section doc:\n%s", b.String())
	}
}
