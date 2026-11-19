package game

import (
	"testing"

	"github.com/kivutar/goro/network"
)

func TestMergeHatEffectsAddsAndRemovesOneID(t *testing.T) {
	got := mergeHatEffects([]uint16{1, 2}, network.HatEffect{Enabled: true, IDs: []uint16{3}})
	if len(got) != 3 || got[2] != 3 {
		t.Fatalf("enable dropped others: %v", got)
	}
	got = mergeHatEffects(got, network.HatEffect{Enabled: false, IDs: []uint16{2}})
	if len(got) != 2 || got[0] != 1 || got[1] != 3 {
		t.Fatalf("disable cleared the set: %v", got)
	}
}
