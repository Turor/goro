package effects

import (
	"reflect"
	"testing"
	"time"

	"github.com/kivutar/goro/db"
	"github.com/kivutar/goro/game/internal/testutil"
	"github.com/kivutar/goro/render"
	"github.com/kivutar/goro/res"
)

func TestDrawsBeforeEntitiesUsesRenderBeforeOrNegativeLayer(t *testing.T) {
	if (Component{}).DrawsBeforeEntities() {
		t.Fatal("default layer 0 must draw after entities")
	}
	if !(Component{RenderBefore: true}).DrawsBeforeEntities() {
		t.Fatal("RenderBefore must draw before entities")
	}
	if !(Component{Layer: -1}).DrawsBeforeEntities() {
		t.Fatal("negative Layer must draw before entities")
	}
	if (Component{Layer: 1}).DrawsBeforeEntities() {
		t.Fatal("positive Layer must draw after entities")
	}
	if !(db.EffectComponent{Layer: -2}).DrawsBeforeEntities() || (db.EffectComponent{}).DrawsBeforeEntities() {
		t.Fatal("db EffectComponent layering must match the runtime component")
	}
}

func TestLoudSpecDrawsAfterEntities(t *testing.T) {
	spec, ok := Lookup(Loud)
	if !ok || len(spec.Components) != 1 {
		t.Fatalf("EF_LOUD spec = %+v ok=%t", spec, ok)
	}
	component := spec.Components[0]
	if component.Kind != ComponentSTR || component.StrFile != "loud" || !component.AttachedEntity {
		t.Fatalf("EF_LOUD component = %+v", component)
	}
	if component.DrawsBeforeEntities() || component.Layer < 0 || component.RenderBefore {
		t.Fatalf("Crazy Uproar must draw after NPCs: %+v", component)
	}
}

func TestLoudSTRCommandsDrawAfterBillboards(t *testing.T) {
	str := &res.STR{
		FPS:    60,
		MaxKey: 8,
		Layers: []res.STRLayer{{
			Textures: []string{"loud.bmp"},
			Animations: []res.STRAnimation{
				{Type: 0, Color: [4]float32{1, 1, 1, 1}, XY: [8]float32{-16, 16, 16, -16, -16, -16, 16, 16}},
				{Type: 1, Frame: 8},
			},
		}},
	}
	m := &System{
		StrEffects: map[string]*res.STR{`__str_data\texture\effect\loud.str|`: str},
		textures:   map[string]*render.Image{"__strtex_loud.bmp": render.NewImage(8, 8)},
	}
	ctx := testutil.MerchantEffectContext()
	start := time.Unix(100, 0)
	if !m.AddAt(ctx, Loud, 2000000, start) {
		t.Fatal("failed to add Crazy Uproar")
	}
	projection := testutil.Projection(800, 600, 10, 20, 0)
	after := render.NewFrame(800, 600)
	m.Draw(after, ctx, projection, start, false)
	commands := reflect.ValueOf(after).Elem().FieldByName("worldCommands")
	if commands.Len() == 0 {
		t.Fatal("Crazy Uproar STR produced no world geometry")
	}
	options := commands.Index(0).FieldByName("Options")
	if !options.FieldByName("AfterBillboards").Bool() {
		t.Fatal("Crazy Uproar STR must draw after NPC billboards")
	}
	if !options.FieldByName("DepthTest").Bool() || options.FieldByName("DepthWrite").Bool() {
		t.Fatal("Crazy Uproar STR must test actor silhouettes without writing transparent depth")
	}
	before := render.NewFrame(800, 600)
	m.Draw(before, ctx, projection, start, true)
	if reflect.ValueOf(before).Elem().FieldByName("worldCommands").Len() != 0 {
		t.Fatal("Crazy Uproar must not draw in the pre-entity pass")
	}
}

func TestNegativeLayerSTRDrawsBeforeBillboards(t *testing.T) {
	const effectID = -61
	db.EffectSpecs[effectID] = db.EffectSpec{Components: []db.EffectComponent{
		{Kind: db.EffectComponentSTR, STRFile: "g61before", Layer: -1, Duration: time.Hour},
	}}
	t.Cleanup(func() { delete(db.EffectSpecs, effectID) })
	str := &res.STR{
		FPS:    60,
		MaxKey: 8,
		Layers: []res.STRLayer{{
			Textures: []string{"g61before.bmp"},
			Animations: []res.STRAnimation{
				{Type: 0, Color: [4]float32{1, 1, 1, 1}, XY: [8]float32{-8, 8, 8, -8, -8, -8, 8, 8}},
				{Type: 1, Frame: 8},
			},
		}},
	}
	m := &System{
		StrEffects: map[string]*res.STR{`__str_data\texture\effect\g61before.str|`: str},
		textures:   map[string]*render.Image{"__strtex_g61before.bmp": render.NewImage(4, 4)},
	}
	ctx := testutil.MerchantEffectContext()
	start := time.Unix(100, 0)
	m.Active = []Instance{{EffectID: effectID, X: 10, Y: 20, Starts: start, Expires: start.Add(time.Hour)}}
	projection := testutil.Projection(800, 600, 10, 20, 0)
	before := render.NewFrame(800, 600)
	m.Draw(before, ctx, projection, start, true)
	commands := reflect.ValueOf(before).Elem().FieldByName("worldCommands")
	if commands.Len() == 0 {
		t.Fatal("negative Layer STR produced no pre-entity geometry")
	}
	if commands.Index(0).FieldByName("Options").FieldByName("AfterBillboards").Bool() {
		t.Fatal("pre-entity STR must draw in the first world alpha pass")
	}
	after := render.NewFrame(800, 600)
	m.Draw(after, ctx, projection, start, false)
	if reflect.ValueOf(after).Elem().FieldByName("worldCommands").Len() != 0 {
		t.Fatal("negative Layer STR must not draw after entities")
	}
}

func TestMissingEffectTextureReturnsCheckerboardOnce(t *testing.T) {
	root := t.TempDir()
	manager := &res.Manager{Root: root}
	m := &System{}
	first := m.EffectFileTexture(manager, "g61/missing.tga")
	if !IsMissingTexture(first) {
		t.Fatal("missing effect texture must be the checkerboard placeholder")
	}
	second := m.EffectFileTexture(manager, "g61/missing.tga")
	if first != second {
		t.Fatal("missing effect texture must be cached")
	}
	if _, missed := m.textureMiss["__effectfile_g61/missing.tga|blackkey=true"]; !missed {
		t.Fatal("missing path must be recorded")
	}
	strTex := m.strEffectTexture(manager, "g61/str-hole.bmp")
	if !IsMissingTexture(strTex) {
		t.Fatal("missing STR texture must be the checkerboard placeholder")
	}
	named := m.effectTexture(manager, "g61-named-hole")
	if !IsMissingTexture(named) {
		t.Fatal("missing named effect texture must be the checkerboard placeholder")
	}
}
