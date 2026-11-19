package render

type WorldMesh struct {
	vertices     []Vertex3D
	indices      []uint16
	texture      *Image
	lightTexture *Image
	options      DrawTrianglesOptions
	version      uint64
	model        [16]float32
	uvOffset     [2]float32
}

type WorldMeshCommand struct {
	Mesh *WorldMesh
}

func identityModelMatrix() [16]float32 {
	return [16]float32{1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1}
}

func NewWorldMeshWithLightmap(vertices []Vertex3D, indices []uint16, texture, lightTexture *Image, opts *DrawTrianglesOptions) *WorldMesh {
	var o DrawTrianglesOptions
	if opts != nil {
		o = *opts
	}
	return &WorldMesh{
		vertices:     append([]Vertex3D(nil), vertices...),
		indices:      append([]uint16(nil), indices...),
		texture:      texture,
		lightTexture: lightTexture,
		options:      o,
		version:      1,
		model:        identityModelMatrix(),
	}
}

func (m *WorldMesh) VertexCount() int {
	if m == nil {
		return 0
	}
	return len(m.vertices)
}

func (m *WorldMesh) IndexCount() int {
	if m == nil {
		return 0
	}
	return len(m.indices)
}

func (m *WorldMesh) modelIsIdentity() bool {
	if m == nil {
		return true
	}
	if m.model == ([16]float32{}) || m.model == identityModelMatrix() {
		return true
	}
	return false
}

func (m *WorldMesh) ModelMatrix() [16]float32 {
	if m == nil || m.modelIsIdentity() {
		return identityModelMatrix()
	}
	return m.model
}

func (m *WorldMesh) UVOffset() [2]float32 {
	if m == nil {
		return [2]float32{}
	}
	return m.uvOffset
}

// SetTexture swaps the frame texture without rebuilding vertices. A size change
// bumps the GPU version so baked UVs match the new texture.
func (m *WorldMesh) SetTexture(texture *Image) {
	if m == nil || texture == nil || m.texture == texture {
		return
	}
	oldW, oldH := 0, 0
	if m.texture != nil {
		oldW, oldH = m.texture.Bounds().Dx(), m.texture.Bounds().Dy()
	}
	m.texture = texture
	w, h := texture.Bounds().Dx(), texture.Bounds().Dy()
	if w != oldW || h != oldH {
		m.version++
	}
}

func (m *WorldMesh) SetUVOffset(offset [2]float32) {
	if m == nil {
		return
	}
	m.uvOffset = offset
}

func (m *WorldMesh) SetModel(model [16]float32) {
	if m == nil {
		return
	}
	m.model = model
}
