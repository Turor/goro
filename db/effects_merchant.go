package db

import "time"

func mammoniteEffectSpec() EffectSpec {
	spec := strEffectSpecAttachedMin("maemor", "memor_min", "", false)
	spec.Duration = 100 * EffectFrameDuration
	spec.Components = append(spec.Components, EffectComponent{
		Kind: EffectComponentFUNC, FuncName: "MammoniteCoins",
		SpriteFile: "동전", Duration: spec.Duration,
	})
	return spec
}

func cartRevolutionEffectSpec() EffectSpec {
	spec := strEffectSpecAttached("cartrevolution", "", false)
	spec.Duration = 300 * EffectFrameDuration
	spec.SFX = []string{"effect\\ef_magnumbreak.wav", "effect\\ef_magnumbreak.wav"}
	spec.SFXDelays = []time.Duration{7 * EffectFrameDuration, 20 * EffectFrameDuration}
	spec.Components = append(spec.Components, EffectComponent{
		Kind: EffectComponentFUNC, FuncName: "CartRevolution", Duration: 40 * EffectFrameDuration,
		TextureFile: "effect/ring_yellow.tga", TextureFiles: []string{"effect/대폭발.tga"},
	})
	return spec
}

func loudEffectSpec() EffectSpec {
	spec := strEffectSpecAttached("loud", "effect\\고성방가.wav", false)
	// Draw with the actor pass and a camera-plane bias so Crazy Uproar sits
	// in front of NPCs. G61 owns any remaining depth-test visual fix.
	spec.Layer = EffectLayerActor
	spec.ZOrder = 10
	spec.Depth = 0.05
	if len(spec.Components) > 0 {
		spec.Components[0].Layer = EffectLayerActor
		spec.Components[0].ZOrder = 10
		spec.Components[0].Depth = 0.05
		spec.Components[0].RenderBefore = false
	}
	return spec
}
