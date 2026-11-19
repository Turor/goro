package effects

import (
	"testing"
	"time"

	"github.com/kivutar/goro/game/internal/testutil"
	"github.com/kivutar/goro/res"
	worldstate "github.com/kivutar/goro/world"
)

func TestSyncHatEffectsDoesNotRemoveSharedSkillOrHat(t *testing.T) {
	ctx := testutil.Level99AuraTestContext(1, false)
	ctx.Resources = &res.Manager{}
	ctx.Resources.SetHatEffectPlaybackForTest(map[int]int{7: Level99, 8: Level99})
	ctx.World.Player.HatEffects = []uint16{7, 8}
	ctx.World.Actors[300] = worldstate.Actor{ID: 300, X: 12, Y: 20, HatEffects: []uint16{7}}
	now := time.Unix(10, 0)
	m := &System{}
	m.AddCellLifetime(ctx, Level99, 300, 12, 20, now, time.Minute, false)
	m.SyncHatEffects(ctx, now)
	player := uint32(2000000)
	if !m.hasHatOwned(Level99, player, now) || !m.hasHatOwned(Level99, 300, now) {
		t.Fatal("hat-owned effect was not attached")
	}
	skillLeft := 0
	for _, effect := range m.Active {
		if effect.EffectID == Level99 && effect.ActorID == 300 && !effect.hatOwned {
			skillLeft++
		}
	}
	if skillLeft != 1 {
		t.Fatalf("skill copies = %d, want 1", skillLeft)
	}
	ctx.World.Player.HatEffects = []uint16{8}
	ctx.World.Actors[300] = worldstate.Actor{ID: 300, X: 12, Y: 20}
	m.SyncHatEffects(ctx, now)
	if !m.hasHatOwned(Level99, player, now) {
		t.Fatal("removing one shared hat removed the effect the other hat still needs")
	}
	if m.hasHatOwned(Level99, 300, now) {
		t.Fatal("actor hat stayed after it was cleared")
	}
	skillLeft = 0
	for _, effect := range m.Active {
		if effect.EffectID == Level99 && effect.ActorID == 300 && !effect.hatOwned {
			skillLeft++
		}
	}
	if skillLeft != 1 {
		t.Fatal("unequipping the hat removed the skill effect")
	}
}
