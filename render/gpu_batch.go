package render

import (
	"encoding/binary"
	"fmt"
	"strings"

	"github.com/gogpu/gputypes"
	"github.com/gogpu/wgpu"
)

const (
	billboardArrayFloatCount = billboardInstanceFloatCount + 4
	textureArrayLayerCap     = 64
)

func worldMeshArrayShaderSource() string {
	src := strings.Replace(worldShaderWGSL, "var tex: texture_2d<f32>;", "var tex: texture_2d_array<f32>;", 1)
	src = strings.Replace(src, "@location(5) world_pos: vec3<f32>,\n}", "@location(5) world_pos: vec3<f32>,\n\t@location(6) layer: f32,\n}", 1)
	src = strings.Replace(src, "fn vs_main(input: VertexInput) -> VertexOutput {", "fn vs_main(input: VertexInput, @builtin(instance_index) instance: u32) -> VertexOutput {", 1)
	src = strings.Replace(src, "out.world_pos = world_pos;", "out.world_pos = world_pos;\n\tout.layer = f32(instance);", 1)
	return strings.Replace(src, "textureSample(tex, tex_sampler, input.uv)", "textureSample(tex, tex_sampler, input.uv, i32(input.layer))", 1)
}

func (r *gpuRenderer) frameTextureArrays(screen *Frame) bool {
	if r == nil || !r.textureArrays {
		return false
	}
	return billboardsFitArray(screen) && meshesFitArray(screen)
}

func billboardsFitArray(screen *Frame) bool {
	if screen == nil {
		return true
	}
	seen := map[*Image]struct{}{}
	for _, cmd := range screen.worldBillboards {
		if cmd.Texture == nil || cmd.Texture.pix == nil || cmd.Width <= 0 || cmd.Height <= 0 {
			continue
		}
		if _, ok := seen[cmd.Texture]; ok {
			continue
		}
		if len(seen) >= textureArrayLayerCap {
			return false
		}
		seen[cmd.Texture] = struct{}{}
	}
	return true
}

func meshesFitArray(screen *Frame) bool {
	if screen == nil {
		return true
	}
	seen := map[*Image]struct{}{}
	var width, height int
	for _, cmd := range screen.worldMeshes {
		mesh := cmd.Mesh
		if mesh == nil || !mesh.options.DepthWrite || !mesh.modelIsIdentity() || mesh.uvOffset != [2]float32{} {
			continue
		}
		if mesh.texture == nil || mesh.texture.pix == nil {
			continue
		}
		if _, ok := seen[mesh.texture]; ok {
			continue
		}
		w, h := mesh.texture.Bounds().Dx(), mesh.texture.Bounds().Dy()
		if w <= 0 || h <= 0 {
			return false
		}
		if len(seen) >= textureArrayLayerCap {
			return false
		}
		if width == 0 {
			width, height = w, h
		} else if w != width || h != height {
			return false
		}
		seen[mesh.texture] = struct{}{}
	}
	return true
}

func (r *gpuRenderer) initTextureArraySupport() {
	if r == nil || r.dev == nil {
		return
	}
	features := r.dev.Features()
	if !features.Contains(gputypes.FeatureMultiDrawIndirect) || !features.Contains(gputypes.FeatureIndirectFirstInstance) {
		return
	}
	probe, probeView, err := r.createTextureArray("goro-array-probe", 1, 1, 1)
	if err != nil {
		return
	}
	probeView.Release()
	probe.Release()
	if err := r.createArrayPipelines(); err != nil {
		r.releaseArrayPipelines()
		return
	}
	r.textureArrays = true
}

func (r *gpuRenderer) createTextureArray(label string, width, height, layers int) (*wgpu.Texture, *wgpu.TextureView, error) {
	if width < 1 {
		width = 1
	}
	if height < 1 {
		height = 1
	}
	if layers < 1 {
		layers = 1
	}
	tex, err := r.dev.CreateTexture(&wgpu.TextureDescriptor{
		Label:         label,
		Size:          wgpu.Extent3D{Width: uint32(width), Height: uint32(height), DepthOrArrayLayers: uint32(layers)},
		MipLevelCount: 1,
		SampleCount:   1,
		Dimension:     gputypes.TextureDimension2D,
		Format:        gputypes.TextureFormatRGBA8Unorm,
		Usage:         wgpu.TextureUsageTextureBinding | wgpu.TextureUsageCopyDst,
	})
	if err != nil {
		return nil, nil, err
	}
	view, err := r.dev.CreateTextureView(tex, &wgpu.TextureViewDescriptor{
		Label:           label + "-view",
		Format:          gputypes.TextureFormatRGBA8Unorm,
		Dimension:       gputypes.TextureViewDimension2DArray,
		MipLevelCount:   1,
		ArrayLayerCount: uint32(layers),
	})
	if err != nil {
		tex.Release()
		return nil, nil, err
	}
	return tex, view, nil
}

func (r *gpuRenderer) createArrayPipelines() error {
	meshBGL, err := r.dev.CreateBindGroupLayout(&wgpu.BindGroupLayoutDescriptor{
		Label:   "goro-world-array-layout",
		Entries: arrayBindEntries(true),
	})
	if err != nil {
		return err
	}
	r.arrayBGL = meshBGL
	r.arrayLayout, err = r.dev.CreatePipelineLayout(&wgpu.PipelineLayoutDescriptor{
		Label:            "goro-world-array-pipeline-layout",
		BindGroupLayouts: []*wgpu.BindGroupLayout{r.arrayBGL, r.modelBGL},
	})
	if err != nil {
		return err
	}
	shader, err := r.dev.CreateShaderModule(&wgpu.ShaderModuleDescriptor{Label: "goro-world-array", WGSL: worldMeshArrayShaderSource()})
	if err != nil {
		return err
	}
	defer shader.Release()
	pipeline, err := r.dev.CreateRenderPipeline(r.arrayPipelineDescriptor(shader, gputypes.BlendStateAlpha(), true, true, "goro-world-array-pipeline"))
	if err != nil {
		return err
	}
	r.arrayPipeline = pipeline
	r.arrayPipelineKey = worldPipelineKey{BlendSourceOver, true, true}

	billboardShader, err := r.dev.CreateShaderModule(&wgpu.ShaderModuleDescriptor{Label: "goro-billboard-array", WGSL: worldBillboardArrayShaderWGSL})
	if err != nil {
		return err
	}
	defer billboardShader.Release()
	r.billboardArrayBGL, err = r.dev.CreateBindGroupLayout(&wgpu.BindGroupLayoutDescriptor{
		Label:   "goro-billboard-array-layout",
		Entries: arrayBindEntries(false),
	})
	if err != nil {
		return err
	}
	r.billboardArrayLayout, err = r.dev.CreatePipelineLayout(&wgpu.PipelineLayoutDescriptor{
		Label:            "goro-billboard-array-pipeline-layout",
		BindGroupLayouts: []*wgpu.BindGroupLayout{r.billboardArrayBGL},
	})
	if err != nil {
		return err
	}
	r.billboardArrayPipeline, err = r.dev.CreateRenderPipeline(r.billboardArrayPipelineDescriptor(billboardShader, gputypes.BlendStateAlpha(), true, false, "goro-billboard-array"))
	if err != nil {
		return err
	}
	r.billboardArrayDepth, err = r.dev.CreateRenderPipeline(r.billboardArrayPipelineDescriptor(billboardShader, gputypes.BlendStateAlpha(), true, true, "goro-billboard-array-depth"))
	if err != nil {
		return err
	}
	if r.arrayPipelines == nil {
		r.arrayPipelines = map[worldPipelineKey]*wgpu.RenderPipeline{}
	}
	if r.billboardArrayPipelines == nil {
		r.billboardArrayPipelines = map[worldPipelineKey]*wgpu.RenderPipeline{}
	}
	add := gputypes.BlendState{
		Color: gputypes.BlendComponent{SrcFactor: gputypes.BlendFactorSrcAlpha, DstFactor: gputypes.BlendFactorOne, Operation: gputypes.BlendOperationAdd},
		Alpha: gputypes.BlendComponent{SrcFactor: gputypes.BlendFactorOne, DstFactor: gputypes.BlendFactorOne, Operation: gputypes.BlendOperationAdd},
	}
	srcDst := gputypes.BlendState{
		Color: gputypes.BlendComponent{SrcFactor: gputypes.BlendFactorSrcAlpha, DstFactor: gputypes.BlendFactorDstAlpha, Operation: gputypes.BlendOperationAdd},
		Alpha: gputypes.BlendComponent{SrcFactor: gputypes.BlendFactorSrcAlpha, DstFactor: gputypes.BlendFactorDstAlpha, Operation: gputypes.BlendOperationAdd},
	}
	blends := []struct {
		blend Blend
		state gputypes.BlendState
	}{
		{BlendSourceOver, gputypes.BlendStateAlpha()},
		{BlendLighter, add},
		{BlendSrcAlphaDstAlpha, srcDst},
	}
	for _, item := range blends {
		for _, depthTest := range []bool{false, true} {
			for _, depthWrite := range []bool{false, true} {
				key := worldPipelineKey{item.blend, depthTest, depthWrite}
				if key == r.arrayPipelineKey {
					continue
				}
				extra, err := r.dev.CreateRenderPipeline(r.arrayPipelineDescriptor(shader, item.state, depthTest, depthWrite, "goro-world-array-extra"))
				if err != nil {
					return err
				}
				r.arrayPipelines[key] = extra
			}
		}
		for _, depthTest := range []bool{false, true} {
			key := worldPipelineKey{item.blend, depthTest, false}
			pipeline, err := r.dev.CreateRenderPipeline(r.billboardArrayPipelineDescriptor(billboardShader, item.state, depthTest, false, "goro-billboard-array-color"))
			if err != nil {
				return err
			}
			if item.blend == BlendSourceOver && depthTest {
				pipeline.Release()
				r.billboardArrayPipelines[key] = r.billboardArrayPipeline
				continue
			}
			r.billboardArrayPipelines[key] = pipeline
		}
	}
	return nil
}

func arrayBindEntries(lightSampler bool) []gputypes.BindGroupLayoutEntry {
	entries := []gputypes.BindGroupLayoutEntry{
		{Binding: 0, Visibility: wgpu.ShaderStageVertex | wgpu.ShaderStageFragment, Buffer: &gputypes.BufferBindingLayout{Type: gputypes.BufferBindingTypeUniform, MinBindingSize: 96}},
		{Binding: 1, Visibility: wgpu.ShaderStageFragment, Sampler: &gputypes.SamplerBindingLayout{Type: gputypes.SamplerBindingTypeFiltering}},
		{Binding: 2, Visibility: wgpu.ShaderStageFragment, Texture: &gputypes.TextureBindingLayout{SampleType: gputypes.TextureSampleTypeFloat, ViewDimension: gputypes.TextureViewDimension2DArray}},
		{Binding: 3, Visibility: wgpu.ShaderStageFragment, Texture: &gputypes.TextureBindingLayout{SampleType: gputypes.TextureSampleTypeFloat, ViewDimension: gputypes.TextureViewDimension2D}},
		{Binding: 5, Visibility: wgpu.ShaderStageVertex | wgpu.ShaderStageFragment, Buffer: &gputypes.BufferBindingLayout{Type: gputypes.BufferBindingTypeUniform, MinBindingSize: LightingUniformBytes}},
		{Binding: 6, Visibility: wgpu.ShaderStageFragment, Texture: &gputypes.TextureBindingLayout{SampleType: gputypes.TextureSampleTypeDepth, ViewDimension: gputypes.TextureViewDimension2D}},
	}
	if lightSampler {
		entries = append(entries, gputypes.BindGroupLayoutEntry{
			Binding: 4, Visibility: wgpu.ShaderStageFragment, Sampler: &gputypes.SamplerBindingLayout{Type: gputypes.SamplerBindingTypeFiltering},
		})
	}
	return entries
}

func (r *gpuRenderer) arrayPipelineDescriptor(shader *wgpu.ShaderModule, blend gputypes.BlendState, depthTest, depthWrite bool, label string) *wgpu.RenderPipelineDescriptor {
	desc := r.worldPipelineDescriptor(shader, blend, depthTest, depthWrite, label)
	desc.Layout = r.arrayLayout
	return desc
}

func (r *gpuRenderer) billboardArrayPipelineDescriptor(shader *wgpu.ShaderModule, blend gputypes.BlendState, depthTest, depthWrite bool, label string) *wgpu.RenderPipelineDescriptor {
	desc := r.worldBillboardPipelineDescriptor(shader, blend, depthTest, label)
	desc.Layout = r.billboardArrayLayout
	desc.Vertex.Buffers[1].ArrayStride = billboardArrayFloatCount * 4
	desc.Vertex.Buffers[1].Attributes = append(desc.Vertex.Buffers[1].Attributes, gputypes.VertexAttribute{
		Format: gputypes.VertexFormatFloat32x4, Offset: billboardInstanceFloatCount * 4, ShaderLocation: 10,
	})
	if depthWrite {
		desc.Vertex.EntryPoint = "vs_main"
		desc.DepthStencil.DepthWriteEnabled = true
		desc.Fragment.Targets[0].WriteMask = gputypes.ColorWriteMaskNone
	}
	return desc
}

func (r *gpuRenderer) billboardArrayPipelineFor(blend Blend, depthTest bool) *wgpu.RenderPipeline {
	if blend != BlendLighter && blend != BlendSrcAlphaDstAlpha {
		blend = BlendSourceOver
	}
	key := worldPipelineKey{blend, depthTest, false}
	if pipeline := r.billboardArrayPipelines[key]; pipeline != nil {
		return pipeline
	}
	return r.billboardArrayPipeline
}

func (r *gpuRenderer) arrayPipelineFor(options DrawTrianglesOptions) *wgpu.RenderPipeline {
	blend := options.Blend
	if blend != BlendLighter && blend != BlendSrcAlphaDstAlpha {
		blend = BlendSourceOver
	}
	key := worldPipelineKey{blend, options.DepthTest, options.DepthWrite}
	if key == r.arrayPipelineKey && r.arrayPipeline != nil {
		return r.arrayPipeline
	}
	if pipeline := r.arrayPipelines[key]; pipeline != nil {
		return pipeline
	}
	return r.arrayPipeline
}

func (r *gpuRenderer) releaseArrayPipelines() {
	if r.arrayPipeline != nil {
		r.arrayPipeline.Release()
		r.arrayPipeline = nil
	}
	for key, pipeline := range r.arrayPipelines {
		if pipeline != nil {
			pipeline.Release()
		}
		delete(r.arrayPipelines, key)
	}
	if r.arrayLayout != nil {
		r.arrayLayout.Release()
		r.arrayLayout = nil
	}
	if r.arrayBGL != nil {
		r.arrayBGL.Release()
		r.arrayBGL = nil
	}
	for key, pipeline := range r.billboardArrayPipelines {
		if pipeline != nil && pipeline != r.billboardArrayPipeline {
			pipeline.Release()
		}
		delete(r.billboardArrayPipelines, key)
	}
	if r.billboardArrayPipeline != nil {
		r.billboardArrayPipeline.Release()
		r.billboardArrayPipeline = nil
	}
	if r.billboardArrayDepth != nil {
		r.billboardArrayDepth.Release()
		r.billboardArrayDepth = nil
	}
	if r.billboardArrayLayout != nil {
		r.billboardArrayLayout.Release()
		r.billboardArrayLayout = nil
	}
	if r.billboardArrayBGL != nil {
		r.billboardArrayBGL.Release()
		r.billboardArrayBGL = nil
	}
}

type arrayMeshDraw struct {
	options DrawTrianglesOptions
	light   *Image
	texture *Image
	offset  uint64
	count   uint32
}

func (r *gpuRenderer) prepareMeshArray(screen *Frame) error {
	r.meshesPacked = false
	r.arrayDraws = r.arrayDraws[:0]
	if screen == nil || r.arrayPipeline == nil {
		return nil
	}
	type group struct {
		options DrawTrianglesOptions
		light   *Image
		texture *Image
		meshes  []*WorldMesh
	}
	var groups []*group
	index := map[string]int{}
	layers := map[*Image]uint32{}
	var layerOrder []*Image
	var width, height int
	for _, cmd := range screen.worldMeshes {
		mesh := cmd.Mesh
		if mesh == nil || !mesh.options.DepthWrite || !mesh.modelIsIdentity() || mesh.uvOffset != [2]float32{} {
			continue
		}
		if mesh.texture == nil || mesh.texture.pix == nil || len(mesh.vertices) == 0 || len(mesh.indices) == 0 {
			continue
		}
		w, h := mesh.texture.Bounds().Dx(), mesh.texture.Bounds().Dy()
		if width == 0 {
			width, height = w, h
		}
		if _, ok := layers[mesh.texture]; !ok {
			if len(layerOrder) >= textureArrayLayerCap {
				return fmt.Errorf("texture array layer cap")
			}
			layers[mesh.texture] = uint32(len(layerOrder))
			layerOrder = append(layerOrder, mesh.texture)
		}
		key := fmt.Sprintf("%p|%v", mesh.lightTexture, mesh.options)
		slot, ok := index[key]
		if !ok {
			slot = len(groups)
			index[key] = slot
			groups = append(groups, &group{options: mesh.options, light: mesh.lightTexture, texture: mesh.texture})
		}
		groups[slot].meshes = append(groups[slot].meshes, mesh)
	}
	if len(groups) == 0 {
		return nil
	}
	if err := r.ensureMeshArrayTexture(width, height, len(layerOrder)); err != nil {
		return err
	}
	for layer, texture := range layerOrder {
		if err := r.writeArrayLayer(r.arrayTex, texture, width, height, layer); err != nil {
			return err
		}
	}
	var floats []float32
	var indices []uint32
	var indirect []byte
	draws := make([]arrayMeshDraw, 0, len(groups))
	for _, group := range groups {
		start := len(indirect)
		for _, mesh := range group.meshes {
			baseVertex := uint32(len(floats) / worldVertexFloatCount)
			firstIndex := uint32(len(indices))
			meshFloats, meshIndices := worldMeshGPUData(mesh, width, height)
			floats = append(floats, meshFloats...)
			indices = append(indices, meshIndices...)
			var cmd [20]byte
			binary.LittleEndian.PutUint32(cmd[0:4], uint32(len(meshIndices)))
			binary.LittleEndian.PutUint32(cmd[4:8], 1)
			binary.LittleEndian.PutUint32(cmd[8:12], firstIndex)
			binary.LittleEndian.PutUint32(cmd[12:16], baseVertex)
			binary.LittleEndian.PutUint32(cmd[16:20], layers[mesh.texture])
			indirect = append(indirect, cmd[:]...)
		}
		draws = append(draws, arrayMeshDraw{
			options: group.options,
			light:   group.light,
			texture: group.texture,
			offset:  uint64(start),
			count:   uint32(len(group.meshes)),
		})
	}
	if _, err := r.dynamicBuffer(&r.arrayVertex, "goro-array-vertices", len(floats)*4, wgpu.BufferUsageVertex|wgpu.BufferUsageCopyDst, floatBytes(floats)); err != nil {
		return err
	}
	if _, err := r.dynamicBuffer(&r.arrayIndex, "goro-array-indices", len(indices)*4, wgpu.BufferUsageIndex|wgpu.BufferUsageCopyDst, u32Bytes(indices)); err != nil {
		return err
	}
	if _, err := r.dynamicBuffer(&r.arrayIndirect, "goro-array-indirect", len(indirect), wgpu.BufferUsageIndirect|wgpu.BufferUsageCopyDst, indirect); err != nil {
		return err
	}
	r.arrayDraws = draws
	r.meshesPacked = true
	return nil
}

func (r *gpuRenderer) ensureMeshArrayTexture(width, height, layers int) error {
	if r.arrayTex != nil && r.arrayWidth == width && r.arrayHeight == height && r.arrayLayers == layers {
		return nil
	}
	r.dropArrayBindGroups(r.arrayView)
	if r.arrayView != nil {
		r.arrayView.Release()
		r.arrayView = nil
	}
	if r.arrayTex != nil {
		r.arrayTex.Release()
		r.arrayTex = nil
	}
	tex, view, err := r.createTextureArray("goro-mesh-array", width, height, layers)
	if err != nil {
		return err
	}
	r.arrayTex = tex
	r.arrayView = view
	r.arrayWidth = width
	r.arrayHeight = height
	r.arrayLayers = layers
	return nil
}

func (r *gpuRenderer) dropArrayBindGroups(view *wgpu.TextureView) {
	if view == nil {
		return
	}
	for key, group := range r.bindGroups {
		if key.arrayView == view {
			group.Release()
			delete(r.bindGroups, key)
		}
	}
}

func (r *gpuRenderer) writeArrayLayer(tex *wgpu.Texture, image *Image, width, height, layer int) error {
	if tex == nil || image == nil || image.pix == nil {
		return fmt.Errorf("array layer missing texture")
	}
	bytesPerRow := align256(width * 4)
	payload := make([]byte, bytesPerRow*height)
	srcW := image.pix.Rect.Dx()
	srcH := image.pix.Rect.Dy()
	if srcW > width {
		srcW = width
	}
	if srcH > height {
		srcH = height
	}
	for y := 0; y < srcH; y++ {
		row := image.pix.Pix[y*image.pix.Stride:]
		n := srcW * 4
		if n > len(row) {
			n = len(row)
		}
		copy(payload[y*bytesPerRow:], row[:n])
	}
	return r.queue.WriteTexture(
		&wgpu.ImageCopyTexture{Texture: tex, Aspect: gputypes.TextureAspectAll, Origin: wgpu.Origin3D{Z: uint32(layer)}},
		payload,
		&wgpu.ImageDataLayout{BytesPerRow: uint32(bytesPerRow), RowsPerImage: uint32(height)},
		&wgpu.Extent3D{Width: uint32(width), Height: uint32(height), DepthOrArrayLayers: 1},
	)
}

func align256(n int) int {
	if n < 256 {
		return 256
	}
	return (n + 255) &^ 255
}

func (r *gpuRenderer) bindArrayGroup(layout *wgpu.BindGroupLayout, arrayView *wgpu.TextureView, light *gpuImageTexture, sampler *wgpu.Sampler, lightSampler bool) (*wgpu.BindGroup, error) {
	key := bindGroupKey{lightTexture: light, sampler: sampler, layout: layout, shadow: r.shadowView, arrayView: arrayView}
	if group := r.bindGroups[key]; group != nil {
		return group, nil
	}
	entries := []wgpu.BindGroupEntry{
		{Binding: 0, Buffer: r.worldUniform, Size: 96},
		{Binding: 1, Sampler: sampler},
		{Binding: 2, TextureView: arrayView},
		{Binding: 3, TextureView: light.view},
		{Binding: 5, Buffer: r.lightingBuf, Size: LightingUniformBytes},
		{Binding: 6, TextureView: r.shadowView},
	}
	if lightSampler {
		entries = append(entries, wgpu.BindGroupEntry{Binding: 4, Sampler: r.lightmapSampler})
	}
	group, err := r.dev.CreateBindGroup(&wgpu.BindGroupDescriptor{Label: "goro-array-bind", Layout: layout, Entries: entries})
	if err != nil {
		return nil, err
	}
	r.bindGroups[key] = group
	return group, nil
}

func (r *gpuRenderer) drawArrayMeshes(pass *wgpu.RenderPassEncoder) error {
	if !r.meshesPacked || pass == nil || r.arrayView == nil || len(r.arrayDraws) == 0 {
		return nil
	}
	if r.arrayVertex.buf == nil || r.arrayIndex.buf == nil || r.arrayIndirect.buf == nil {
		return nil
	}
	pass.SetVertexBuffer(0, r.arrayVertex.buf, 0)
	pass.SetIndexBuffer(r.arrayIndex.buf, gputypes.IndexFormatUint32, 0)
	for _, draw := range r.arrayDraws {
		if draw.count == 0 {
			continue
		}
		lightTex, err := r.ensureBatchLightTexture(drawBatchKey{lightTexture: draw.light})
		if err != nil {
			return err
		}
		sampler, err := r.sampler(draw.options, draw.texture)
		if err != nil {
			return err
		}
		group, err := r.bindArrayGroup(r.arrayBGL, r.arrayView, lightTex, sampler, true)
		if err != nil {
			return err
		}
		pass.SetPipeline(r.arrayPipelineFor(draw.options))
		pass.SetBindGroup(0, group, nil)
		r.bindModelGroup(pass, nil)
		pass.MultiDrawIndexedIndirect(r.arrayIndirect.buf, draw.offset, draw.count)
	}
	return nil
}

func (r *gpuRenderer) prepareArrayBillboards(commands []WorldBillboardCommand) error {
	r.releaseBillboardArrayGroups()
	type layer struct {
		image *Image
	}
	seen := map[*Image]int{}
	var images []*Image
	maxW, maxH := 0, 0
	for _, cmd := range commands {
		if cmd.Texture == nil || cmd.Texture.pix == nil || cmd.Width <= 0 || cmd.Height <= 0 {
			continue
		}
		w, h := cmd.Texture.Bounds().Dx(), cmd.Texture.Bounds().Dy()
		if w > maxW {
			maxW = w
		}
		if h > maxH {
			maxH = h
		}
		if _, ok := seen[cmd.Texture]; ok {
			continue
		}
		seen[cmd.Texture] = len(images)
		images = append(images, cmd.Texture)
	}
	if len(images) == 0 || maxW <= 0 || maxH <= 0 {
		return nil
	}
	if r.billboardArrayTex == nil || r.billboardArrayW != maxW || r.billboardArrayH != maxH || r.billboardArrayLayers != len(images) {
		r.dropArrayBindGroups(r.billboardArrayView)
		if r.billboardArrayView != nil {
			r.billboardArrayView.Release()
			r.billboardArrayView = nil
		}
		if r.billboardArrayTex != nil {
			r.billboardArrayTex.Release()
			r.billboardArrayTex = nil
		}
		tex, view, err := r.createTextureArray("goro-billboard-array", maxW, maxH, len(images))
		if err != nil {
			return err
		}
		r.billboardArrayTex = tex
		r.billboardArrayView = view
		r.billboardArrayW = maxW
		r.billboardArrayH = maxH
		r.billboardArrayLayers = len(images)
	}
	for layer, image := range images {
		if err := r.writeArrayLayer(r.billboardArrayTex, image, maxW, maxH, layer); err != nil {
			return err
		}
	}
	r.worldBillboardFloats = r.worldBillboardFloats[:0]
	r.worldBillboardBatches = r.worldBillboardBatches[:0]
	for _, cmd := range commands {
		if cmd.Texture == nil || cmd.Texture.pix == nil || cmd.Width <= 0 || cmd.Height <= 0 {
			continue
		}
		key := drawBatchKey{options: cmd.Options}
		if len(r.worldBillboardBatches) == 0 || r.worldBillboardBatches[len(r.worldBillboardBatches)-1].key != key {
			r.worldBillboardBatches = append(r.worldBillboardBatches, worldBillboardBatch{
				key:           key,
				firstInstance: uint32(len(r.worldBillboardFloats) / billboardArrayFloatCount),
			})
		}
		w, h := cmd.Texture.Bounds().Dx(), cmd.Texture.Bounds().Dy()
		scaleX, scaleY := float32(w)/float32(maxW), float32(h)/float32(maxH)
		r.worldBillboardFloats = appendBillboardInstanceData(r.worldBillboardFloats, cmd)
		r.worldBillboardFloats = append(r.worldBillboardFloats, float32(seen[cmd.Texture]), scaleX, scaleY, 0)
		r.worldBillboardBatches[len(r.worldBillboardBatches)-1].instanceCount++
	}
	if len(r.worldBillboardFloats) == 0 {
		return nil
	}
	slot := &r.billboardInstanceSlots[r.billboardSlot&1]
	if _, err := r.dynamicBuffer(slot, "goro-world-billboard-instances", len(r.worldBillboardFloats)*4, wgpu.BufferUsageVertex|wgpu.BufferUsageCopyDst, floatBytes(r.worldBillboardFloats)); err != nil {
		return err
	}
	light, err := r.ensureTexture(r.neutralLightmapImage())
	if err != nil {
		return err
	}
	for _, batch := range r.worldBillboardBatches {
		sampler, err := r.sampler(batch.key.options, images[0])
		if err != nil {
			return err
		}
		group, err := r.bindArrayGroup(r.billboardArrayBGL, r.billboardArrayView, light, sampler, false)
		if err != nil {
			return err
		}
		r.billboardArrayGroups = append(r.billboardArrayGroups, group)
	}
	r.activeBillboard = slot
	r.billboardStride = billboardArrayFloatCount * 4
	r.billboardArrayFrame = true
	return nil
}

func (r *gpuRenderer) releaseBillboardArrayGroups() {
	if r == nil {
		return
	}
	r.billboardArrayGroups = r.billboardArrayGroups[:0]
}
