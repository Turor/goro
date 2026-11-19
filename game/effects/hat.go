package effects

import (
	"time"

	"github.com/kivutar/goro/client"
	"github.com/kivutar/goro/game/scene"
)

type hatAttachKey struct {
	actor uint32
	hat   int
}

// SyncHatEffects plays stored hat-effect ids that map to an effect-table id.
// Resource-only cosmetics stay stored until a sprite player exists.
func (m *System) SyncHatEffects(ctx client.Context, now time.Time) {
	if m == nil || ctx.World == nil || ctx.Resources == nil {
		return
	}
	if m.hatAttached == nil {
		m.hatAttached = map[hatAttachKey]int{}
	}
	wanted := map[hatAttachKey]int{}
	add := func(actorID uint32, ids []uint16) {
		if actorID == 0 {
			return
		}
		for _, id := range ids {
			effectID, ok := ctx.Resources.HatEffectPlaybackID(int(id))
			if !ok {
				continue
			}
			key := hatAttachKey{actor: actorID, hat: int(id)}
			wanted[key] = effectID
			if !m.HasActive(effectID, actorID, now) {
				m.AddPersistentBetween(ctx, effectID, actorID, 0, now)
			}
		}
	}
	add(scene.LocalSkillTarget(ctx), ctx.World.Player.HatEffects)
	for _, actor := range ctx.World.Actors {
		add(actor.ID, actor.HatEffects)
	}
	for key, effectID := range m.hatAttached {
		if wanted[key] == effectID {
			continue
		}
		m.Remove(effectID, key.actor)
		delete(m.hatAttached, key)
	}
	for key, effectID := range wanted {
		m.hatAttached[key] = effectID
	}
}
