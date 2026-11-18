package render

import "fmt"

// DynamicBufferSlot is the upload slot for a frame. Consecutive frames use
// different slots so the GPU can still be reading the previous frame's buffer.
func DynamicBufferSlot(frame uint64) int {
	return int(frame & 1)
}

// DynamicUpload identifies a per-frame dynamic buffer.
type DynamicUpload int

const (
	DynamicBillboards DynamicUpload = iota
	DynamicTriangles
)

// DynamicBufferSet holds two payload slots for each dynamic upload. Tests and
// the GPU backend write the same slots.
type DynamicBufferSet struct {
	billboards [2][]byte
	triangles  [2][]byte
}

// Write copies payload into the slot for frame and returns that slot.
func (s *DynamicBufferSet) Write(frame uint64, kind DynamicUpload, payload []byte) int {
	if s == nil {
		return DynamicBufferSlot(frame)
	}
	slot := DynamicBufferSlot(frame)
	dst := append([]byte(nil), payload...)
	switch kind {
	case DynamicBillboards:
		s.billboards[slot] = dst
	default:
		s.triangles[slot] = dst
	}
	return slot
}

// Bytes returns the payload last written to slot.
func (s *DynamicBufferSet) Bytes(kind DynamicUpload, slot int) []byte {
	if s == nil || slot < 0 || slot > 1 {
		return nil
	}
	if kind == DynamicBillboards {
		return s.billboards[slot]
	}
	return s.triangles[slot]
}

// BillboardSubmitItem is one world billboard the submission planner sees.
type BillboardSubmitItem struct {
	TextureID       string
	Blend           Blend
	DepthTest       bool
	DepthWrite      bool
	AfterBillboards bool
	Filter          Filter
	Address         Address
	DisableFog      bool
}

// MeshSubmitItem is one retained world mesh the submission planner sees.
// ModelKey is empty when the mesh uses the shared identity model, so meshes
// that differ only by texture stay in one draw. A posed mesh keeps its own key.
type MeshSubmitItem struct {
	TextureID       string
	LightmapID      string
	ModelKey        string
	Blend           Blend
	DepthTest       bool
	DepthWrite      bool
	AfterBillboards bool
	Filter          Filter
	Address         Address
	DisableFog      bool
}

// PlannedDraw is one draw the backend submits.
// Kind is "instanced" or "multi-draw" when a texture array collapses textures,
// and "texture" when each texture is bound on its own.
type PlannedDraw struct {
	Kind      string
	Textures  []string
	Instances int
	MultiDraw bool
}

type billboardGroupKey struct {
	texture         string
	blend           Blend
	depthTest       bool
	depthWrite      bool
	afterBillboards bool
	filter          Filter
	address         Address
	disableFog      bool
	extra           string
}

func (item BillboardSubmitItem) groupKey(textureArrays bool) billboardGroupKey {
	key := billboardGroupKey{
		blend:           item.Blend,
		depthTest:       item.DepthTest,
		depthWrite:      item.DepthWrite,
		afterBillboards: item.AfterBillboards,
		filter:          item.Filter,
		address:         item.Address,
		disableFog:      item.DisableFog,
	}
	if !textureArrays {
		key.texture = item.TextureID
	}
	return key
}

// PlanBillboardDraws groups billboards. With a texture array, billboards that
// share blend and pipeline state but not a texture become one instanced draw.
// Without one, a texture change starts a new draw (one bind per texture run).
func PlanBillboardDraws(items []BillboardSubmitItem, textureArrays bool) []PlannedDraw {
	var draws []PlannedDraw
	var keys []billboardGroupKey
	for _, item := range items {
		if item.TextureID == "" {
			continue
		}
		key := item.groupKey(textureArrays)
		if len(draws) == 0 || keys[len(keys)-1] != key {
			kind := "texture"
			if textureArrays {
				kind = "instanced"
			}
			draws = append(draws, PlannedDraw{Kind: kind, Instances: 0})
			keys = append(keys, key)
		}
		draw := &draws[len(draws)-1]
		draw.Instances++
		if !containsString(draw.Textures, item.TextureID) {
			draw.Textures = append(draw.Textures, item.TextureID)
		}
	}
	return draws
}

// PlanRetainedMeshDraws groups retained meshes. With a texture array, meshes
// that share pipeline state become one multi-draw. Otherwise each texture is
// its own draw.
func PlanRetainedMeshDraws(items []MeshSubmitItem, textureArrays bool) []PlannedDraw {
	var draws []PlannedDraw
	index := map[billboardGroupKey]int{}
	for _, item := range items {
		if item.TextureID == "" {
			continue
		}
		key := item.meshGroupKey(textureArrays)
		slot, ok := index[key]
		if !ok {
			slot = len(draws)
			index[key] = slot
			kind := "texture"
			if textureArrays {
				kind = "multi-draw"
			}
			draws = append(draws, PlannedDraw{Kind: kind, MultiDraw: textureArrays, Textures: []string{item.TextureID}})
		} else if !containsString(draws[slot].Textures, item.TextureID) {
			draws[slot].Textures = append(draws[slot].Textures, item.TextureID)
		}
		draws[slot].Instances++
	}
	return draws
}

// FrameSubmission is the draw list and dynamic uploads for one frame.
type FrameSubmission struct {
	BillboardDraws []PlannedDraw
	MeshDraws      []PlannedDraw
	BillboardSlot  int
	TriangleSlot   int
	BillboardBytes []byte
	TriangleBytes  []byte
}

// BuildFrameSubmission plans draws for screen and writes this frame's dynamic
// billboard and triangle bytes into bufs. The GPU upload uses the same slots.
func BuildFrameSubmission(frame uint64, screen *Frame, textureArrays bool, bufs *DynamicBufferSet) FrameSubmission {
	if bufs == nil {
		bufs = &DynamicBufferSet{}
	}
	items, commands := billboardSubmitItems(screen)
	meshes := meshSubmitItems(screen)
	instances := EncodeBillboardInstances(commands)
	triangles := EncodeDynamicTriangles(screen)
	sub := FrameSubmission{
		BillboardDraws: PlanBillboardDraws(items, textureArrays),
		MeshDraws:      PlanRetainedMeshDraws(meshes, textureArrays),
		BillboardBytes: floatBytes(instances),
		TriangleBytes:  triangles,
	}
	sub.BillboardSlot = bufs.Write(frame, DynamicBillboards, sub.BillboardBytes)
	sub.TriangleSlot = bufs.Write(frame, DynamicTriangles, sub.TriangleBytes)
	return sub
}

func billboardSubmitItems(screen *Frame) ([]BillboardSubmitItem, []WorldBillboardCommand) {
	if screen == nil {
		return nil, nil
	}
	items := make([]BillboardSubmitItem, 0, len(screen.worldBillboards))
	commands := make([]WorldBillboardCommand, 0, len(screen.worldBillboards))
	for _, cmd := range screen.worldBillboards {
		if cmd.Texture == nil || cmd.Texture.pix == nil || cmd.Width <= 0 || cmd.Height <= 0 {
			continue
		}
		commands = append(commands, cmd)
		items = append(items, BillboardSubmitItem{
			TextureID:       TextureID(cmd.Texture),
			Blend:           cmd.Options.Blend,
			DepthTest:       cmd.Options.DepthTest,
			DepthWrite:      cmd.Options.DepthWrite,
			AfterBillboards: cmd.Options.AfterBillboards,
			Filter:          cmd.Options.Filter,
			Address:         cmd.Options.Address,
			DisableFog:      cmd.Options.DisableFog,
		})
	}
	return items, commands
}

func meshSubmitItems(screen *Frame) []MeshSubmitItem {
	if screen == nil {
		return nil
	}
	items := make([]MeshSubmitItem, 0, len(screen.worldMeshes))
	for _, cmd := range screen.worldMeshes {
		mesh := cmd.Mesh
		if mesh == nil || mesh.texture == nil || mesh.texture.pix == nil || len(mesh.vertices) == 0 || len(mesh.indices) == 0 {
			continue
		}
		if !mesh.options.DepthWrite {
			continue
		}
		modelKey := ""
		if !mesh.modelIsIdentity() || mesh.uvOffset != [2]float32{} {
			modelKey = fmt.Sprintf("%p", mesh)
		}
		items = append(items, MeshSubmitItem{
			TextureID:       TextureID(mesh.texture),
			LightmapID:      TextureID(mesh.lightTexture),
			ModelKey:        modelKey,
			Blend:           mesh.options.Blend,
			DepthTest:       mesh.options.DepthTest,
			DepthWrite:      mesh.options.DepthWrite,
			AfterBillboards: mesh.options.AfterBillboards,
			Filter:          mesh.options.Filter,
			Address:         mesh.options.Address,
			DisableFog:      mesh.options.DisableFog,
		})
	}
	return items
}

// TextureID is stable for a texture value for the rest of the process.
func TextureID(img *Image) string {
	if img == nil {
		return ""
	}
	return fmt.Sprintf("%p", img)
}

// EncodeBillboardInstances packs billboard instance floats in submission order.
func EncodeBillboardInstances(commands []WorldBillboardCommand) []float32 {
	out := make([]float32, 0, len(commands)*billboardInstanceFloatCount)
	for _, cmd := range commands {
		if cmd.Texture == nil || cmd.Texture.pix == nil || cmd.Width <= 0 || cmd.Height <= 0 {
			continue
		}
		out = appendBillboardInstanceData(out, cmd)
	}
	return out
}

// EncodeDynamicTriangles packs dynamic world-triangle vertices. The bytes change
// when the triangle payload changes and stay out of the retained mesh buffers.
func EncodeDynamicTriangles(screen *Frame) []byte {
	if screen == nil {
		return nil
	}
	var floats []float32
	for _, cmd := range screen.worldCommands {
		if cmd.Texture == nil || cmd.Texture.pix == nil || len(cmd.Vertices) == 0 {
			continue
		}
		for _, v := range cmd.Vertices {
			floats = append(floats,
				v.X, v.Y, v.Z,
				v.SrcX, v.SrcY,
				v.ColorR, v.ColorG, v.ColorB, v.ColorA,
				v.DepthX, v.DepthY, v.DepthZ,
				v.LightSrcX, v.LightSrcY,
			)
		}
	}
	return floatBytes(floats)
}

func (item MeshSubmitItem) meshGroupKey(textureArrays bool) billboardGroupKey {
	key := BillboardSubmitItem{
		TextureID:       item.TextureID,
		Blend:           item.Blend,
		DepthTest:       item.DepthTest,
		DepthWrite:      item.DepthWrite,
		AfterBillboards: item.AfterBillboards,
		Filter:          item.Filter,
		Address:         item.Address,
		DisableFog:      item.DisableFog,
	}.groupKey(textureArrays)
	key.extra = item.ModelKey + "\x00" + item.LightmapID
	return key
}

func containsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
