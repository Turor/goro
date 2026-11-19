package effects

import (
	"image/color"
	"testing"
	"time"

	"github.com/kivutar/goro/client"
	"github.com/kivutar/goro/db"
	"github.com/kivutar/goro/game/internal/testutil"
	"github.com/kivutar/goro/game/scene"
	"github.com/kivutar/goro/render"
	"github.com/kivutar/goro/world"
)

func TestDeformedBodyCastsShadowWithoutEffectCopies(t *testing.T) {
	image := render.NewImage(8, 16)
	image.Fill(color.White)
	billboard := &scene.Billboard{Image: image, AnchorX: 0.5, AnchorY: 1}
	projection := testutil.Projection(800, 600, 0, 0, 0)
	now := time.Unix(1_700_000_000, 0)

	t.Run("idle", func(t *testing.T) {
		screen := render.NewFrame(64, 64)
		screen.BeginFrame()
		entry := scene.ActorDrawEntry{Actor: world.Actor{ID: 7}, WorldX: 10, WorldY: 12, WorldZ: 1}
		(&System{}).DrawActorSpriteBodyEffects3D(client.Context{}, screen, projection, entry, billboard, 1, 1, 1, color.RGBA{255, 255, 255, 255}, render.BlendSourceOver, 0, now)
		assertOneShadowCaster(t, screen)
	})

	t.Run("kaite", func(t *testing.T) {
		screen := render.NewFrame(64, 64)
		screen.BeginFrame()
		entry := scene.ActorDrawEntry{Actor: world.Actor{ID: 7, Opt3State: db.Opt3Kaite}, WorldX: 10, WorldY: 12, WorldZ: 1}
		(&System{}).DrawActorSpriteBodyEffects3D(client.Context{}, screen, projection, entry, billboard, 1, 1, 1, color.RGBA{255, 255, 255, 255}, render.BlendSourceOver, 0, now)
		cmds := screen.WorldBillboards()
		if len(cmds) < 2 {
			t.Fatalf("kaite billboards = %d, want the body plus expansions", len(cmds))
		}
		casters := 0
		for _, cmd := range cmds {
			if cmd.Options.CastShadow {
				casters++
				if !cmd.Options.DepthWrite {
					t.Fatal("shadow caster does not depth-write")
				}
			}
		}
		if casters != 1 {
			t.Fatalf("kaite shadow casters = %d, want only the solid body", casters)
		}
	})

	t.Run("jump", func(t *testing.T) {
		screen := render.NewFrame(64, 64)
		screen.BeginFrame()
		entry := scene.ActorDrawEntry{Actor: world.Actor{ID: 7}, WorldX: 10, WorldY: 12, WorldZ: 1}
		system := &System{Active: []Instance{{
			EffectID: JumpBody,
			ActorID:  7,
			Starts:   now.Add(-50 * db.EffectFrameDuration),
		}}}
		system.DrawActorSpriteBodyEffects3D(client.Context{}, screen, projection, entry, billboard, 1, 1, 1, color.RGBA{255, 255, 255, 255}, render.BlendSourceOver, 0, now)
		assertOneShadowCaster(t, screen)
	})
}

func assertOneShadowCaster(t *testing.T, screen *render.Frame) {
	t.Helper()
	cmds := screen.WorldBillboards()
	if len(cmds) != 1 {
		t.Fatalf("billboards = %d, want 1", len(cmds))
	}
	if !cmds[0].Options.CastShadow {
		t.Fatal("body billboard does not cast a shadow")
	}
}
