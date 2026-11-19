package res

import "testing"

func TestParseHatEffectPlaybackSkipsResourceOnly(t *testing.T) {
	src := []byte(`
hatEffectTable = {}
hatEffectTable[4] = { hatEffectID = 120, resourceFileName = "effect\\a.str" }
hatEffectTable[5] = { hatEffectID = 0, resourceFileName = "effect\\b.str" }
`)
	got, err := ParseHatEffectPlayback(src)
	if err != nil {
		t.Fatal(err)
	}
	if got[4] != 120 {
		t.Fatalf("id 4 = %d", got[4])
	}
	if _, ok := got[5]; ok {
		t.Fatal("resource-only entry was treated as an effect")
	}
}
