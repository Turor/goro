package effects

import (
	"testing"

	"github.com/kivutar/goro/db"
)

func TestLookupPlumbsLayeringFields(t *testing.T) {
	spec, ok := Lookup(Loud)
	if !ok {
		t.Fatal("loud spec missing")
	}
	if spec.Layer != db.EffectLayerActor || spec.ZOrder != 10 || spec.Depth != 0.05 {
		t.Fatalf("loud spec layering = layer %d z %d depth %v", spec.Layer, spec.ZOrder, spec.Depth)
	}
	if len(spec.Components) == 0 {
		t.Fatal("no components")
	}
	c := spec.Components[0]
	if spec.componentDrawsBeforeEntities(c) {
		t.Fatal("Crazy Uproar would draw in the ground pass, behind NPCs")
	}
	if c.Depth != 0.05 || c.Layer != db.EffectLayerActor {
		t.Fatalf("loud component layering = %+v", c)
	}
}

func TestComponentDrawsBeforeEntitiesHonorsLayer(t *testing.T) {
	spec := Spec{Layer: db.EffectLayerGround}
	if !spec.componentDrawsBeforeEntities(Component{}) {
		t.Fatal("ground layer not before entities")
	}
	spec.Layer = db.EffectLayerActor
	if spec.componentDrawsBeforeEntities(Component{RenderBefore: true}) {
		t.Fatal("actor layer lost to RenderBefore")
	}
	if (Spec{}).componentDrawsBeforeEntities(Component{RenderBefore: true}) != true {
		t.Fatal("auto layer should use RenderBefore")
	}
}

func TestEffect3DSpriteDrawOptionsHonorDepth(t *testing.T) {
	options := Effect3DSpriteDrawOptions(Component{Depth: 0.08})
	if options.DepthBias != 0.08 {
		t.Fatalf("depth bias = %v, want 0.08", options.DepthBias)
	}
}

func TestInvalidateSpecsAfterTableReload(t *testing.T) {
	m := &System{}
	m.specs = map[int]Spec{Loud: {Depth: 1}}
	db.RestoreCompiledEffectTables()
	m.effectGen = db.EffectTablesGeneration()
	if m.effectGen == 0 {
		t.Fatal("restore did not bump generation")
	}
	db.RestoreCompiledEffectTables()
	if m.effectGen == db.EffectTablesGeneration() {
		t.Fatal("second restore did not bump generation")
	}
	m.invalidateSpecsIfStale()
	if m.specs != nil {
		t.Fatal("cached specs survived table generation change")
	}
	if m.effectGen != db.EffectTablesGeneration() {
		t.Fatal("generation not recorded")
	}
}
