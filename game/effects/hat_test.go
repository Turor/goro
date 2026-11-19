package effects

import (
	"testing"
	"time"

	"github.com/kivutar/goro/game/internal/testutil"
	"github.com/kivutar/goro/res"
	worldstate "github.com/kivutar/goro/world"
)

func TestSyncHatEffectsAddsAndRemovesPlayback(t *testing.T) {
	ctx := testutil.Level99AuraTestContext(1, false)
	ctx.Resources = &res.Manager{}
	ctx.Resources.SetHatEffectPlaybackForTest(map[int]int{7: Level99})
	ctx.World.Player.HatEffects = []uint16{7}
	ctx.World.Actors[300] = worldstate.Actor{ID: 300, X: 12, Y: 20, HatEffects: []uint16{7}}
	now := time.Unix(10, 0)
	m := &System{}
	m.AddCellLifetime(ctx, Heal, 300, 12, 20, now, time.Minute, false)
	m.SyncHatEffects(ctx, now)
	player := uint32(2000000)
	if !m.HasActive(Level99, player, now) || !m.HasActive(Level99, 300, now) {
		t.Fatal("hat effect was not attached")
	}
	if !m.HasActive(Heal, 300, now) {
		t.Fatal("hat sync removed a skill effect")
	}
	ctx.World.Player.HatEffects = nil
	ctx.World.Actors[300] = worldstate.Actor{ID: 300, X: 12, Y: 20}
	m.SyncHatEffects(ctx, now)
	if m.HasActive(Level99, player, now) || m.HasActive(Level99, 300, now) {
		t.Fatal("removed hat effect stayed attached")
	}
	if !m.HasActive(Heal, 300, now) {
		t.Fatal("removing a hat effect removed a skill effect")
	}
}
