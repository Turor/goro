package maps

import (
	"testing"
	"time"

	"github.com/kivutar/goro/game/scene"
	"github.com/kivutar/goro/render"
	"github.com/kivutar/goro/res"
)

func TestWaterChunks(t *testing.T) {
	gnd := &res.GND{
		Width:  2,
		Height: 2,
		Water: res.GNDWater{
			Present: true, Level: 5, Type: 1, WaveSpeed: 1, AnimSpeed: 1,
		},
		Cells: make([]res.GNDCell, 4),
	}
	for i := range gnd.Cells {
		gnd.Cells[i].Top = 0
	}
	system := &System{}
	early := time.UnixMilli(0)
	later := time.UnixMilli(1000)
	first := system.BuildWaterChunks(gnd, nil, scene.Projection{}, early, 64, 64)
	if len(first) != 1 || len(first[0].Vertices) == 0 {
		t.Fatalf("first water chunks = %d", len(first))
	}
	vertex := &first[0].Vertices[0]
	screen := newRetainScreen()
	system.DrawGNDWater(screen, nil, gnd, nil, scene.Projection{}, early, scene.SceneFog{})
	if system.waterRetain == nil || system.waterRetain.mesh == nil {
		t.Fatal("water draw did not retain a mesh")
	}
	mesh := system.waterRetain.mesh
	uvEarly := mesh.UVOffset()
	screen.BeginFrame()
	system.DrawGNDWater(screen, nil, gnd, nil, scene.Projection{}, later, scene.SceneFog{})
	second := system.BuildWaterChunks(gnd, nil, scene.Projection{}, later, 64, 64)
	if len(second) != 1 || &second[0].Vertices[0] != vertex {
		t.Fatal("water chunk vertices were rebuilt")
	}
	if second[0].Anim.Frame == first[0].Anim.Frame && second[0].Anim.UVOffset == first[0].Anim.UVOffset {
		t.Fatalf("water animation did not advance: %#v", second[0].Anim)
	}
	if mesh.UVOffset() == uvEarly {
		t.Fatalf("water UV offset stayed %#v", uvEarly)
	}
	if &system.waterRetain.vertices[0] != vertex {
		t.Fatal("retained water vertex storage changed")
	}
	t.Logf("water verts=%d frame %d->%d uv %.3f->%.3f meshUV=%v", len(first[0].Vertices), first[0].Anim.Frame, second[0].Anim.Frame, first[0].Anim.UVOffset, second[0].Anim.UVOffset, mesh.UVOffset())
}

func TestAnimatedNodes(t *testing.T) {
	rsm := &res.RSM{
		Nodes: []res.RSMNode{{
			Name:   "root",
			Matrix: identityRSMMatrix(),
			Scale:  res.RSMVector3{X: 1, Y: 1, Z: 1},
			Vertices: []res.RSMVector3{
				{X: 0, Y: 0, Z: 0},
				{X: 1, Y: 0, Z: 0},
				{X: 0, Y: 1, Z: 0},
			},
			Faces: []res.RSMFace{{VertexIndices: [3]uint16{0, 1, 2}}},
			PositionKeyframes: []res.RSMPositionKeyframe{
				{Frame: 0, Pos: res.RSMVector3{}},
				{Frame: 100, Pos: res.RSMVector3{X: 5}},
			},
		}},
	}
	system := &System{}
	placement := ModelPlacement{Index: 3, Model: res.RSWModel{Scale: res.RSWVector3{X: 1, Y: 1, Z: 1}}}
	rsw := &res.RSW{}
	first := system.BuildAnimatedNodes(rsm, rsw, placement, 0)
	if len(first) != 1 || len(first[0].Vertices) == 0 {
		t.Fatalf("animated nodes = %d", len(first))
	}
	vertex := &first[0].Vertices[0]
	screen := newRetainScreen()
	system.DrawAnimatedRSMPlacement(screen, nil, rsw, rsm, placement, 0)
	key := animatedRetainKey{rsm: rsm, name: "root"}
	cached := system.animatedRetain[key]
	if cached == nil || cached.mesh == nil {
		t.Fatal("animated draw did not retain a mesh")
	}
	modelEarly := cached.mesh.ModelMatrix()
	screen.BeginFrame()
	system.DrawAnimatedRSMPlacement(screen, nil, rsw, rsm, placement, 100)
	second := system.BuildAnimatedNodes(rsm, rsw, placement, 100)
	if len(second) != 1 || &second[0].Vertices[0] != vertex {
		t.Fatal("animated node vertices were rebuilt")
	}
	if second[0].Transform == first[0].Transform {
		t.Fatal("animated node transform did not change")
	}
	if cached.mesh.ModelMatrix() == modelEarly {
		t.Fatal("retained mesh kept the previous node transform")
	}
	if &cached.vertices[0] != vertex {
		t.Fatal("retained animated vertex storage changed")
	}
	t.Logf("animated verts=%d transform changed model changed", len(first[0].Vertices))
}

func TestRSMFrustum(t *testing.T) {
	gnd := &res.GND{Width: 200, Height: 200}
	rsw := &res.RSW{Models: []res.RSWModel{
		{Filename: "front.rsm", Position: res.RSWVector3{}},
		{Filename: "back.rsm", Position: res.RSWVector3{Z: -70}},
		{Filename: "side.rsm", Position: res.RSWVector3{X: 30}},
	}}
	system := &System{}
	projection := scene.NewProjection(800, 600, 200, 200, 0, 0, 230, 80)
	visible := system.visibleRSMPlacements(rsw, gnd, projection)
	front := placementAt(visible, 200, 200)
	back := placementAt(visible, 200, 130)
	side := placementAt(visible, 230, 200)
	if !front || back || side {
		t.Fatalf("frustum visible front=%t back=%t side=%t count=%d", front, back, side, len(visible))
	}
	open := scene.Projection{PlayerX: 200, PlayerY: 200}
	both := system.visibleRSMPlacements(rsw, gnd, open)
	if !placementAt(both, 200, 200) || !placementAt(both, 200, 130) || !placementAt(both, 230, 200) {
		t.Fatalf("zero view kept front=%t back=%t side=%t", placementAt(both, 200, 200), placementAt(both, 200, 130), placementAt(both, 230, 200))
	}
	t.Logf("frustum kept %d of 3, zero view kept %d", len(visible), len(both))
}

func placementAt(visible []ModelPlacement, baseX, baseY float64) bool {
	for _, placement := range visible {
		if placement.BaseX == baseX && placement.BaseY == baseY {
			return true
		}
	}
	return false
}

func newRetainScreen() *render.Frame {
	screen := render.NewFrame(64, 64)
	screen.BeginFrame()
	return screen
}
