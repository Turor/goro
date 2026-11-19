package effects

import (
	"time"

	"github.com/kivutar/goro/client"
	"github.com/kivutar/goro/game/scene"
)

type actorEffectKey struct {
	actor    uint32
	effectID int
}

// SyncHatEffects plays stored hat-effect ids that map to an effect-table id.
// The instance is marked hat-owned so unequipping it does not delete a skill
// effect, or another hat, that happens to use the same effect-table id.
// Resource-only cosmetics stay stored until a sprite player exists.
func (m *System) SyncHatEffects(ctx client.Context, now time.Time) {
	if m == nil || ctx.World == nil || ctx.Resources == nil {
		return
	}
	wanted := map[actorEffectKey]struct{}{}
	note := func(actorID uint32, ids []uint16) {
		if actorID == 0 {
			return
		}
		for _, id := range ids {
			effectID, ok := ctx.Resources.HatEffectPlaybackID(int(id))
			if !ok {
				continue
			}
			wanted[actorEffectKey{actor: actorID, effectID: effectID}] = struct{}{}
		}
	}
	note(scene.LocalSkillTarget(ctx), ctx.World.Player.HatEffects)
	for _, actor := range ctx.World.Actors {
		note(actor.ID, actor.HatEffects)
	}
	m.removeUnwantedHatEffects(wanted)
	for key := range wanted {
		if m.hasHatOwned(key.effectID, key.actor, now) {
			continue
		}
		m.addHatOwned(ctx, key.effectID, key.actor, now)
	}
}

func (m *System) hasHatOwned(effectID int, actorID uint32, now time.Time) bool {
	for _, effect := range m.Active {
		if effect.hatOwned && effect.EffectID == effectID && effect.ActorID == actorID && !effect.Expired(now) {
			return true
		}
	}
	return false
}

func (m *System) removeUnwantedHatEffects(wanted map[actorEffectKey]struct{}) {
	if len(m.Active) == 0 {
		return
	}
	kept := m.Active[:0]
	for _, effect := range m.Active {
		if effect.hatOwned {
			if _, ok := wanted[actorEffectKey{actor: effect.ActorID, effectID: effect.EffectID}]; !ok {
				continue
			}
		}
		kept = append(kept, effect)
	}
	m.Active = kept
}

func (m *System) addHatOwned(ctx client.Context, effectID int, actorID uint32, starts time.Time) bool {
	if ctx.World == nil || ctx.Config.Headless {
		return false
	}
	if _, ok := Lookup(effectID); !ok {
		return false
	}
	x, y, ok := EffectAnchor(ctx, actorID)
	if !ok {
		return false
	}
	m.Active = append(m.Active, Instance{
		EffectID:   effectID,
		ActorID:    actorID,
		X:          x,
		Y:          y,
		Starts:     starts,
		Persistent: true,
		hatOwned:   true,
	})
	return true
}
