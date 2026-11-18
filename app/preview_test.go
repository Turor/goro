package app

import "testing"

func TestParsePreview(t *testing.T) {
	spec, ok := parsePreview(" prontera , 156,180, 4218", "1002, 1115,")
	if !ok || spec.Map != "prontera" || spec.X != 156 || spec.Y != 180 || spec.Job != 4218 {
		t.Fatalf("spec = %+v ok=%t", spec, ok)
	}
	if len(spec.Actors) != 2 || spec.Actors[0] != 1002 || spec.Actors[1] != 1115 {
		t.Fatalf("actors = %v", spec.Actors)
	}
	if spec, ok := parsePreview("payon", ""); !ok || spec.X != 150 || spec.Y != 150 || spec.Job != 0 {
		t.Fatalf("defaults = %+v ok=%t", spec, ok)
	}
	for _, bad := range []string{"", " ", "prontera,x,1"} {
		if _, ok := parsePreview(bad, ""); ok {
			t.Fatalf("parsePreview(%q) accepted", bad)
		}
	}
	if _, ok := parsePreview("prontera", "1002,abc"); ok {
		t.Fatal("bad actor list accepted")
	}
}
