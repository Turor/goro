package world

import "testing"

func TestUpsertActorPreservesHatEffects(t *testing.T) {
	w := New()
	w.UpsertActor(Actor{ID: 9, HatEffects: []uint16{4, 5}})
	w.UpsertActor(Actor{ID: 9, X: 3, Y: 4})
	got := w.Actors[9].HatEffects
	if len(got) != 2 || got[0] != 4 || got[1] != 5 {
		t.Fatalf("hats dropped on upsert: %v", got)
	}
}
