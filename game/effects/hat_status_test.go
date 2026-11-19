package effects

import (
	"testing"
	"time"

	"github.com/kivutar/goro/game/internal/testutil"
	worldstate "github.com/kivutar/goro/world"
)

func TestHatEffectDoesNotStandInForStatus(t *testing.T) {
	ctx := testutil.Level99AuraTestContext(1, false)
	ctx.World.Actors[300] = worldstate.Actor{ID: 300, X: 12, Y: 20}
	now := time.Unix(10, 0)
	m := &System{}
	if !m.addHatOwned(ctx, Level99, 300, now) {
		t.Fatal("hat add failed")
	}
	if m.HasActive(Level99, 300, now) {
		t.Fatal("status sync would treat the hat as the status effect")
	}
	if !m.AddPersistentBetween(ctx, Level99, 300, 0, now) {
		t.Fatal("status effect was not added beside the hat")
	}
	m.Remove(Level99, 300)
	if !m.hasHatOwned(Level99, 300, now) {
		t.Fatal("status remove deleted the hat")
	}
	if m.HasActive(Level99, 300, now) {
		t.Fatal("status remove left the status effect")
	}
}
