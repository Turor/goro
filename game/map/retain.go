package maps

import (
	"image/color"
	"time"

	"github.com/kivutar/goro/game/scene"
	"github.com/kivutar/goro/render"
	"github.com/kivutar/goro/res"
)

// WaterAnim is the per-frame water animation. Vertices stay fixed.
type WaterAnim struct {
	Frame    int
	UVOffset float64
}

// WaterChunk is one retained water mesh plus the animation for this frame.
type WaterChunk struct {
	Vertices []render.Vertex3D
	Indices  []uint16
	Anim     WaterAnim
}

type waterRetainCache struct {
	gnd            *res.GND
	x0, x1, y0, y1 int
	level          float32
	vertices       []render.Vertex3D
	indices        []uint16
	mesh           *render.WorldMesh
}

// BuildWaterChunks returns retained water for the visible cells. A later frame
// of the same map reuses the vertex payload and changes only WaterAnim.
func (m *System) BuildWaterChunks(gnd *res.GND, rsw *res.RSW, projection scene.Projection, now time.Time, screenW, screenH int) []WaterChunk {
	if m == nil {
		return nil
	}
	water, ok := mapWater(gnd, rsw)
	if !ok || gnd == nil {
		return nil
	}
	x0, x1, y0, y1, ok := gndDrawBoundsWithMargin(gnd, projection, screenW, screenH, 4)
	if !ok {
		return nil
	}
	anim := WaterAnim{Frame: waterFrameForTime(water, now), UVOffset: waterOffsetForTime(water, now)}
	cache := m.waterRetain
	if cache == nil || cache.gnd != gnd || cache.x0 != x0 || cache.x1 != x1 || cache.y0 != y0 || cache.y1 != y1 || cache.level != water.Level {
		cache = m.buildWaterCache(gnd, water, x0, x1, y0, y1, rsw)
		m.waterRetain = cache
	}
	if cache == nil || len(cache.vertices) == 0 {
		return nil
	}
	return []WaterChunk{{
		Vertices: cache.vertices,
		Indices:  cache.indices,
		Anim:     anim,
	}}
}

func (m *System) buildWaterCache(gnd *res.GND, water res.RSWWater, x0, x1, y0, y1 int, rsw *res.RSW) *waterRetainCache {
	tint := waterTint(water, rsw)
	var vertices []render.Vertex3D
	var indices []uint16
	for y := y0; y <= y1; y++ {
		for x := x0; x <= x1; x++ {
			cell, ok := gnd.Cell(x, y)
			if !ok || cell.Top < 0 {
				continue
			}
			// Offset 0 keeps the wave phase out of the vertex buffer.
			draw, ok := newGNDWaterDraw(x, y, cell, water, tint, 0)
			if !ok {
				continue
			}
			base := uint16(len(vertices))
			uvs := draw.uvs
			for i, vert := range draw.verts {
				vertices = append(vertices, waterRetainVertex(vert, uvs[i], tint))
			}
			for _, index := range draw.indices {
				indices = append(indices, base+index)
			}
		}
	}
	if len(vertices) == 0 {
		return &waterRetainCache{gnd: gnd, x0: x0, x1: x1, y0: y0, y1: y1, level: water.Level}
	}
	return &waterRetainCache{
		gnd:      gnd,
		x0:       x0,
		x1:       x1,
		y0:       y0,
		y1:       y1,
		level:    water.Level,
		vertices: vertices,
		indices:  indices,
	}
}

func waterRetainVertex(vert scene.Point3, uv scene.UV, tint color.RGBA) render.Vertex3D {
	return render.Vertex3D{
		X:      float32(vert.X),
		Y:      float32(vert.Y),
		Z:      float32(vert.Z),
		SrcX:   uv.U,
		SrcY:   uv.V,
		ColorR: float32(tint.R) / 255,
		ColorG: float32(tint.G) / 255,
		ColorB: float32(tint.B) / 255,
		ColorA: float32(tint.A) / 255,
		DepthX: float32(vert.X),
		DepthY: float32(vert.Y),
		DepthZ: float32(vert.Z),
	}
}

// AnimatedNode is one map-model node whose vertices stay in node space.
type AnimatedNode struct {
	Name      string
	Vertices  []render.Vertex3D
	Indices   []uint16
	Transform scene.Mat4
}

type animatedRetainKey struct {
	rsm  *res.RSM
	name string
}

type animatedRetainMesh struct {
	vertices []render.Vertex3D
	indices  []uint16
	mesh     *render.WorldMesh
}

// BuildAnimatedNodes returns node-local vertices and the current node transform.
// A later frame reuses the vertex payload and changes only the transform.
func (m *System) BuildAnimatedNodes(rsm *res.RSM, rsw *res.RSW, visible ModelPlacement, frame int) []AnimatedNode {
	if m == nil || rsm == nil {
		return nil
	}
	context, ok := m.rsmPlacementContext(rsm, rsw, visible)
	if !ok {
		return nil
	}
	if m.animatedRetain == nil {
		m.animatedRetain = make(map[animatedRetainKey]*animatedRetainMesh)
	}
	nodeMatrices := m.animatedRSMNodeMatrices(rsm, frame)
	nodes := make([]AnimatedNode, 0, len(context.nodeIndices))
	for _, nodeIndex := range context.nodeIndices {
		if nodeIndex < 0 || nodeIndex >= len(rsm.Nodes) {
			continue
		}
		node := &rsm.Nodes[nodeIndex]
		key := animatedRetainKey{rsm: rsm, name: node.Name}
		cached := m.animatedRetain[key]
		if cached == nil {
			cached = buildAnimatedRetainMesh(rsm, node, context.lighting)
			m.animatedRetain[key] = cached
		}
		if cached == nil || len(cached.vertices) == 0 {
			continue
		}
		transform := rsmNodeModelMatrix(rsm, node, nodeMatrices[node.Name], context.instance)
		nodes = append(nodes, AnimatedNode{
			Name:      node.Name,
			Vertices:  cached.vertices,
			Indices:   cached.indices,
			Transform: transform,
		})
	}
	return nodes
}

func buildAnimatedRetainMesh(rsm *res.RSM, node *res.RSMNode, lighting Lighting) *animatedRetainMesh {
	if node == nil || len(node.Vertices) == 0 {
		return nil
	}
	vertices := make([]render.Vertex3D, len(node.Vertices))
	colors := make([]color.RGBA, len(node.Vertices))
	uvs := make([]scene.UV, len(node.Vertices))
	for i := range colors {
		colors[i] = color.RGBA{R: 255, G: 255, B: 255, A: 255}
	}
	var indices []uint16
	for _, face := range node.Faces {
		ia, ib, ic := int(face.VertexIndices[0]), int(face.VertexIndices[1]), int(face.VertexIndices[2])
		if ia >= len(node.Vertices) || ib >= len(node.Vertices) || ic >= len(node.Vertices) {
			continue
		}
		a := vectorFromRSM(node.Vertices[ia])
		b := vectorFromRSM(node.Vertices[ib])
		c := vectorFromRSM(node.Vertices[ic])
		textureName, faceUVs := rsmFaceTexture(rsm, node, face)
		faceColor := rsmFaceColor(rsm, textureName, a, b, c, lighting)
		for corner, index := range []int{ia, ib, ic} {
			colors[index] = faceColor
			if corner < len(faceUVs) {
				uvs[index] = faceUVs[corner]
			}
		}
		indices = append(indices, face.VertexIndices[0], face.VertexIndices[1], face.VertexIndices[2])
	}
	if len(indices) == 0 {
		return nil
	}
	for i, vertex := range node.Vertices {
		point := vectorFromRSM(vertex)
		tint := colors[i]
		vertices[i] = render.Vertex3D{
			X:      float32(point.X),
			Y:      float32(point.Y),
			Z:      float32(point.Z),
			SrcX:   uvs[i].U,
			SrcY:   uvs[i].V,
			ColorR: float32(tint.R) / 255,
			ColorG: float32(tint.G) / 255,
			ColorB: float32(tint.B) / 255,
			ColorA: float32(tint.A) / 255,
			DepthX: float32(point.X),
			DepthY: float32(point.Y),
			DepthZ: float32(point.Z),
		}
	}
	return &animatedRetainMesh{vertices: vertices, indices: indices}
}

func mat4ToFloat32(matrix scene.Mat4) [16]float32 {
	var out [16]float32
	for i, value := range matrix {
		out[i] = float32(value)
	}
	return out
}
