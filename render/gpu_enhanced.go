package render

import (
	_ "embed"
	"encoding/binary"

	"github.com/gogpu/gputypes"
	"github.com/gogpu/wgpu"
)

//go:embed shadow.wgsl
var shadowShaderWGSL string

//go:embed shadow_sprite.wgsl
var shadowSpriteShaderWGSL string

const (
	modelUniformStride    = 256
	modelUniformSize      = 80
	modelSlotCount        = 64
	shadowMapSize         = 1024
	shadowCamUniformBytes = 64
	shadowUniformBytes    = 80
	// shadowClearDepth is the depth stored before any shadow draw (far plane).
	shadowClearDepth float32 = 1
)

// shadowClearColorFormat is the throwaway color attachment that makes the
// Vulkan backend begin the shadow pass. It is not sampled.
const shadowClearColorFormat = gputypes.TextureFormatRGBA8Unorm

func (r *gpuRenderer) initEnhancedResources() error {
	var err error
	r.modelBGL, err = r.dev.CreateBindGroupLayout(&wgpu.BindGroupLayoutDescriptor{
		Label: "goro-model-bind-layout",
		Entries: []gputypes.BindGroupLayoutEntry{{
			Binding:    0,
			Visibility: wgpu.ShaderStageVertex | wgpu.ShaderStageFragment,
			Buffer: &gputypes.BufferBindingLayout{
				Type:             gputypes.BufferBindingTypeUniform,
				HasDynamicOffset: true,
				MinBindingSize:   modelUniformSize,
			},
		}},
	})
	if err != nil {
		return err
	}
	r.modelBuf, err = r.dev.CreateBuffer(&wgpu.BufferDescriptor{
		Label: "goro-model-uniforms",
		Size:  modelUniformStride * modelSlotCount,
		Usage: wgpu.BufferUsageUniform | wgpu.BufferUsageCopyDst,
	})
	if err != nil {
		return err
	}
	r.modelGroup, err = r.makeModelGroup(r.modelBuf)
	if err != nil {
		return err
	}
	r.lightingBuf, err = r.dev.CreateBuffer(&wgpu.BufferDescriptor{
		Label: "goro-lighting-uniform",
		Size:  LightingUniformBytes,
		Usage: wgpu.BufferUsageUniform | wgpu.BufferUsageCopyDst,
	})
	if err != nil {
		return err
	}
	r.shadowUniform, err = r.dev.CreateBuffer(&wgpu.BufferDescriptor{
		Label: "goro-shadow-uniform",
		Size:  shadowUniformBytes,
		Usage: wgpu.BufferUsageUniform | wgpu.BufferUsageCopyDst,
	})
	if err != nil {
		return err
	}
	shadowTex, err := r.dev.CreateTexture(&wgpu.TextureDescriptor{
		Label:         "goro-shadow-map",
		Size:          wgpu.Extent3D{Width: shadowMapSize, Height: shadowMapSize, DepthOrArrayLayers: 1},
		MipLevelCount: 1,
		SampleCount:   1,
		Dimension:     gputypes.TextureDimension2D,
		Format:        depthFormat,
		Usage:         wgpu.TextureUsageRenderAttachment | wgpu.TextureUsageTextureBinding,
	})
	if err != nil {
		return err
	}
	shadowView, err := r.dev.CreateTextureView(shadowTex, nil)
	if err != nil {
		shadowTex.Release()
		return err
	}
	colorTex, err := r.dev.CreateTexture(&wgpu.TextureDescriptor{
		Label:         "goro-shadow-clear-color",
		Size:          wgpu.Extent3D{Width: shadowMapSize, Height: shadowMapSize, DepthOrArrayLayers: 1},
		MipLevelCount: 1,
		SampleCount:   1,
		Dimension:     gputypes.TextureDimension2D,
		Format:        shadowClearColorFormat,
		Usage:         wgpu.TextureUsageRenderAttachment,
	})
	if err != nil {
		shadowView.Release()
		shadowTex.Release()
		return err
	}
	colorView, err := r.dev.CreateTextureView(colorTex, nil)
	if err != nil {
		colorTex.Release()
		shadowView.Release()
		shadowTex.Release()
		return err
	}
	r.shadowTex = shadowTex
	r.shadowView = shadowView
	r.shadowColorTex = colorTex
	r.shadowColorView = colorView
	if err := r.clearShadowMap(); err != nil {
		return err
	}
	shadowBGL, err := r.dev.CreateBindGroupLayout(&wgpu.BindGroupLayoutDescriptor{
		Label: "goro-shadow-bind-layout",
		Entries: []gputypes.BindGroupLayoutEntry{{
			Binding:    0,
			Visibility: wgpu.ShaderStageVertex,
			Buffer:     &gputypes.BufferBindingLayout{Type: gputypes.BufferBindingTypeUniform, MinBindingSize: shadowCamUniformBytes},
		}},
	})
	if err != nil {
		return err
	}
	r.shadowBGL = shadowBGL
	r.shadowLayout, err = r.dev.CreatePipelineLayout(&wgpu.PipelineLayoutDescriptor{
		Label:            "goro-shadow-pipeline-layout",
		BindGroupLayouts: []*wgpu.BindGroupLayout{r.shadowBGL, r.modelBGL},
	})
	if err != nil {
		return err
	}
	r.shadowGroup, err = r.dev.CreateBindGroup(&wgpu.BindGroupDescriptor{
		Label:  "goro-shadow-bind",
		Layout: r.shadowBGL,
		Entries: []wgpu.BindGroupEntry{{
			Binding: 0, Buffer: r.shadowUniform, Size: shadowCamUniformBytes,
		}},
	})
	if err != nil {
		return err
	}
	shader, err := r.dev.CreateShaderModule(&wgpu.ShaderModuleDescriptor{Label: "goro-shadow", WGSL: shadowShaderWGSL})
	if err != nil {
		// Sun and point lights still shade when the depth-only pipeline cannot be created.
		return nil
	}
	defer shader.Release()
	pipeline, pipeErr := r.dev.CreateRenderPipeline(&wgpu.RenderPipelineDescriptor{
		Label:  "goro-shadow-pipeline",
		Layout: r.shadowLayout,
		Vertex: wgpu.VertexState{
			Module:     shader,
			EntryPoint: "vs_main",
			Buffers: []gputypes.VertexBufferLayout{{
				ArrayStride: worldVertexStride,
				Attributes:  []gputypes.VertexAttribute{{Format: gputypes.VertexFormatFloat32x3, Offset: 0, ShaderLocation: 0}},
			}},
		},
		Primitive: gputypes.PrimitiveState{Topology: gputypes.PrimitiveTopologyTriangleList},
		DepthStencil: &wgpu.DepthStencilState{
			Format:            depthFormat,
			DepthWriteEnabled: true,
			DepthCompare:      gputypes.CompareFunctionLessEqual,
		},
		Multisample: gputypes.MultisampleState{Count: 1, Mask: 0xFFFFFFFF},
		Fragment: &wgpu.FragmentState{
			Module:     shader,
			EntryPoint: "fs_main",
			Targets: []gputypes.ColorTargetState{{
				Format:    shadowClearColorFormat,
				WriteMask: gputypes.ColorWriteMaskNone,
			}},
		},
	})
	if pipeErr != nil {
		return nil
	}
	r.shadowPipeline = pipeline
	r.initShadowSpritePipeline()
	return nil
}

func (r *gpuRenderer) makeModelGroup(buf *wgpu.Buffer) (*wgpu.BindGroup, error) {
	return r.dev.CreateBindGroup(&wgpu.BindGroupDescriptor{
		Label:  "goro-model-bind",
		Layout: r.modelBGL,
		Entries: []wgpu.BindGroupEntry{{
			Binding: 0, Buffer: buf, Size: modelUniformSize,
		}},
	})
}

// shadowMapPass is the pass renderer init uses to clear the shadow map, and
// the pass the directional shadow draw begins. This Vulkan backend returns
// from BeginRenderPass without vkCmdBeginRenderPass when ColorAttachments is
// empty, then End still calls vkCmdEndRenderPass and the driver faults on a
// null render pass at 0x108. The color attachment is what makes Begin start
// the pass. DepthLoadOpClear stores shadowClearDepth for later sampling.
func (r *gpuRenderer) shadowMapPass() *wgpu.RenderPassDescriptor {
	return shadowMapClearPass(r.shadowView, r.shadowColorView)
}

func shadowMapClearPass(depth, color *wgpu.TextureView) *wgpu.RenderPassDescriptor {
	if depth == nil || color == nil {
		return nil
	}
	return &wgpu.RenderPassDescriptor{
		ColorAttachments: []wgpu.RenderPassColorAttachment{{
			View:       color,
			LoadOp:     gputypes.LoadOpClear,
			StoreOp:    gputypes.StoreOpDiscard,
			ClearValue: gputypes.Color{},
		}},
		DepthStencilAttachment: &wgpu.RenderPassDepthStencilAttachment{
			View:            depth,
			DepthLoadOp:     gputypes.LoadOpClear,
			DepthStoreOp:    gputypes.StoreOpStore,
			DepthClearValue: shadowClearDepth,
		},
	}
}

// vulkanPassEndable reports whether this Vulkan backend has begun a pass End
// can finish. BeginRenderPass bails out when there is no color attachment.
func vulkanPassEndable(desc *wgpu.RenderPassDescriptor) bool {
	if desc == nil || len(desc.ColorAttachments) == 0 {
		return false
	}
	for _, attachment := range desc.ColorAttachments {
		if attachment.View == nil {
			return false
		}
	}
	if desc.DepthStencilAttachment != nil && desc.DepthStencilAttachment.View == nil {
		return false
	}
	return true
}

// sampledShadowDepthBeforeDraw is the depth later frames sample after the
// init clear and before any shadow draw. An unendable pass does not store it.
func sampledShadowDepthBeforeDraw(desc *wgpu.RenderPassDescriptor) (float32, bool) {
	if !vulkanPassEndable(desc) || desc.DepthStencilAttachment == nil {
		return 0, false
	}
	depth := desc.DepthStencilAttachment
	if depth.DepthLoadOp != gputypes.LoadOpClear || depth.DepthStoreOp != gputypes.StoreOpStore {
		return 0, false
	}
	return depth.DepthClearValue, true
}

func (r *gpuRenderer) clearShadowMap() error {
	desc := r.shadowMapPass()
	if !vulkanPassEndable(desc) {
		// Do not call End. The backend never began this pass.
		return nil
	}
	enc, err := r.dev.CreateCommandEncoder(&wgpu.CommandEncoderDescriptor{Label: "goro-shadow-clear"})
	if err != nil {
		return err
	}
	defer enc.DiscardEncoding()
	pass, err := enc.BeginRenderPass(desc)
	if err != nil {
		return err
	}
	if err := pass.End(); err != nil {
		return err
	}
	enc.TransitionTextures([]wgpu.TextureBarrier{{
		Texture: r.shadowTex,
		Usage:   wgpu.TextureUsageTransition{OldUsage: wgpu.TextureUsageRenderAttachment, NewUsage: wgpu.TextureUsageTextureBinding},
	}})
	cmd, err := enc.Finish()
	if err != nil {
		return err
	}
	_, err = r.queue.Submit(cmd)
	return err
}

func (r *gpuRenderer) releaseEnhanced() {
	r.releaseBillboardArrayGroups()
	for i := range r.worldVertexSlots {
		releaseDynamic(&r.worldVertexSlots[i])
		releaseDynamic(&r.worldIndexSlots[i])
		releaseDynamic(&r.billboardInstanceSlots[i])
	}
	releaseDynamic(&r.arrayVertex)
	releaseDynamic(&r.arrayIndex)
	releaseDynamic(&r.arrayIndirect)
	if r.arrayGroup != nil {
		r.arrayGroup.Release()
		r.arrayGroup = nil
	}
	if r.arrayTex != nil {
		r.arrayView.Release()
		r.arrayTex.Release()
		r.arrayTex = nil
		r.arrayView = nil
	}
	if r.arrayPipeline != nil {
		r.arrayPipeline.Release()
	}
	if r.arrayLayout != nil {
		r.arrayLayout.Release()
	}
	if r.arrayBGL != nil {
		r.arrayBGL.Release()
	}
	if r.billboardArrayPipeline != nil {
		r.billboardArrayPipeline.Release()
	}
	for key, pipeline := range r.billboardArrayPipelines {
		if pipeline != nil && pipeline != r.billboardArrayPipeline {
			pipeline.Release()
		}
		delete(r.billboardArrayPipelines, key)
	}
	for key, pipeline := range r.arrayPipelines {
		if pipeline != nil && pipeline != r.arrayPipeline {
			pipeline.Release()
		}
		delete(r.arrayPipelines, key)
	}
	if r.billboardArrayLayout != nil {
		r.billboardArrayLayout.Release()
	}
	if r.billboardArrayDepth != nil {
		r.billboardArrayDepth.Release()
		r.billboardArrayDepth = nil
	}
	if r.billboardArrayGroup != nil {
		r.billboardArrayGroup.Release()
		r.billboardArrayGroup = nil
	}
	if r.billboardArrayView != nil {
		r.billboardArrayView.Release()
		r.billboardArrayView = nil
	}
	if r.billboardArrayTex != nil {
		r.billboardArrayTex.Release()
		r.billboardArrayTex = nil
	}
	if r.billboardArrayBGL != nil {
		r.billboardArrayBGL.Release()
	}
	r.releaseShadowSpriteGroups()
	releaseDynamic(&r.shadowSpriteSlot)
	if r.shadowSpritePipeline != nil {
		r.shadowSpritePipeline.Release()
	}
	if r.shadowSpriteLayout != nil {
		r.shadowSpriteLayout.Release()
	}
	if r.shadowSpriteBGL != nil {
		r.shadowSpriteBGL.Release()
	}
	if r.shadowPipeline != nil {
		r.shadowPipeline.Release()
	}
	if r.shadowGroup != nil {
		r.shadowGroup.Release()
	}
	if r.shadowLayout != nil {
		r.shadowLayout.Release()
	}
	if r.shadowBGL != nil {
		r.shadowBGL.Release()
	}
	if r.shadowView != nil {
		r.shadowView.Release()
	}
	if r.shadowTex != nil {
		r.shadowTex.Release()
	}
	if r.shadowColorView != nil {
		r.shadowColorView.Release()
	}
	if r.shadowColorTex != nil {
		r.shadowColorTex.Release()
	}
	if r.shadowUniform != nil {
		r.shadowUniform.Release()
	}
	if r.lightingBuf != nil {
		r.lightingBuf.Release()
	}
	if r.modelGroup != nil {
		r.modelGroup.Release()
	}
	if r.modelBuf != nil {
		r.modelBuf.Release()
	}
	if r.modelBGL != nil {
		r.modelBGL.Release()
	}
	if r.billboardLayout != nil {
		r.billboardLayout.Release()
	}
	if r.fxaa != nil {
		r.fxaa.release()
	}
}

func releaseDynamic(slot *dynamicGPUBuffer) {
	if slot != nil && slot.buf != nil {
		slot.buf.Release()
		slot.buf = nil
		slot.capacity = 0
	}
}

func (r *gpuRenderer) billboardPipelineLayout() *wgpu.PipelineLayout {
	if r == nil || r.billboardLayout == nil {
		if r == nil {
			return nil
		}
		return r.worldLayout
	}
	return r.billboardLayout
}

func (r *gpuRenderer) bindModelGroup(pass *wgpu.RenderPassEncoder, mesh *WorldMesh) {
	if r == nil || r.modelGroup == nil || pass == nil {
		return
	}
	offset := uint32(0)
	if mesh != nil {
		if off, ok := r.modelOffset[mesh]; ok {
			offset = off
		}
	}
	pass.SetBindGroup(1, r.modelGroup, []uint32{offset})
}

func (r *gpuRenderer) recordFrameSubmission(screen *Frame) FrameSubmission {
	arrays := r.frameTextureArrays(screen)
	r.arrayFrame = arrays
	sub := BuildFrameSubmission(r.frameSerial, screen, arrays, &r.dynamicUploads)
	r.billboardSlot = sub.BillboardSlot
	r.triangleSlot = sub.TriangleSlot
	r.framePasses = PlanFramePasses(PassOptions{
		SSAO:     r.ssaoEnabled && screen != nil && screen.camera.Enabled && r.sampleCount == 1,
		FXAA:     r.fxaaEnabled && screen != nil && screen.camera.Enabled,
		Enhanced: r.enhancedEnabled && screen != nil && screen.enhancedLight,
	})
	return sub
}

func (r *gpuRenderer) prepareFrameResources(screen *Frame, world worldFrame, width, height int) error {
	r.meshesPacked = false
	r.billboardArrayFrame = false
	r.recordFrameSubmission(screen)
	if len(world.floats) > 0 {
		r.dynamicUploads.Write(r.frameSerial, DynamicTriangles, floatBytes(world.floats))
	}
	if err := r.prepareModelUniforms(screen); err != nil {
		return err
	}
	if err := r.uploadFrameLighting(screen, width, height); err != nil {
		return err
	}
	if r.arrayFrame {
		if err := r.prepareMeshArray(screen); err != nil {
			r.arrayFrame = false
			r.meshesPacked = false
		}
	}
	commands := []WorldBillboardCommand(nil)
	if screen != nil {
		commands = screen.worldBillboards
	}
	return r.prepareWorldBillboards(commands)
}

func (r *gpuRenderer) prepareModelUniforms(screen *Frame) error {
	if r.modelBuf == nil {
		return nil
	}
	r.modelOffset = map[*WorldMesh]uint32{}
	records := [][]byte{modelRecord(identityModelMatrix(), [2]float32{})}
	if screen != nil {
		for _, cmd := range screen.worldMeshes {
			mesh := cmd.Mesh
			if mesh == nil || (mesh.modelIsIdentity() && mesh.uvOffset == [2]float32{}) {
				continue
			}
			if len(records) >= modelSlotCount {
				break
			}
			r.modelOffset[mesh] = uint32(len(records) * modelUniformStride)
			records = append(records, modelRecord(mesh.ModelMatrix(), mesh.uvOffset))
		}
	}
	buf := make([]byte, modelUniformStride*len(records))
	for i, record := range records {
		copy(buf[i*modelUniformStride:], record)
	}
	return r.queue.WriteBuffer(r.modelBuf, 0, buf)
}

func modelRecord(model [16]float32, uv [2]float32) []byte {
	buf := make([]byte, modelUniformStride)
	for i, value := range model {
		binary.LittleEndian.PutUint32(buf[i*4:], binaryFloat32(value))
	}
	binary.LittleEndian.PutUint32(buf[64:], binaryFloat32(uv[0]))
	binary.LittleEndian.PutUint32(buf[68:], binaryFloat32(uv[1]))
	return buf
}

func (r *gpuRenderer) uploadFrameLighting(screen *Frame, width, height int) error {
	if r.lightingBuf == nil {
		return nil
	}
	var sun SunLight
	var points []PointLight
	vp := IdentityViewProjection()
	shadow := IdentityViewProjection()
	enhanced := r.enhancedEnabled && screen != nil && screen.enhancedLight
	if screen != nil {
		sun = screen.sunLight
		points = screen.pointLights
		if screen.hasMapLight {
			vp = screen.viewProj
			shadow = screen.shadowVP
		} else if screen.camera.Enabled {
			vp = screen.camera.ViewProjection
		}
	}
	if err := r.queue.WriteBuffer(r.lightingBuf, 0, EncodeLightingUniform(sun, points, vp, shadow, width, height, enhanced)); err != nil {
		return err
	}
	if r.shadowUniform == nil {
		return nil
	}
	raw := make([]byte, shadowUniformBytes)
	for i, value := range shadow {
		binary.LittleEndian.PutUint32(raw[i*4:], binaryFloat32(value))
	}
	binary.LittleEndian.PutUint32(raw[64:], binaryFloat32(sun.Direction[0]))
	binary.LittleEndian.PutUint32(raw[68:], binaryFloat32(sun.Direction[1]))
	binary.LittleEndian.PutUint32(raw[72:], binaryFloat32(sun.Direction[2]))
	return r.queue.WriteBuffer(r.shadowUniform, 0, raw)
}

func (r *gpuRenderer) drawDirectionalShadow(enc *wgpu.CommandEncoder, screen *Frame) error {
	r.releaseShadowSpriteGroups()
	if r.shadowPipeline == nil || r.shadowTex == nil || screen == nil {
		return nil
	}
	enc.TransitionTextures([]wgpu.TextureBarrier{{
		Texture: r.shadowTex,
		Usage:   wgpu.TextureUsageTransition{OldUsage: wgpu.TextureUsageTextureBinding, NewUsage: wgpu.TextureUsageRenderAttachment},
	}})
	desc := r.shadowMapPass()
	if !vulkanPassEndable(desc) {
		enc.TransitionTextures([]wgpu.TextureBarrier{{
			Texture: r.shadowTex,
			Usage:   wgpu.TextureUsageTransition{OldUsage: wgpu.TextureUsageRenderAttachment, NewUsage: wgpu.TextureUsageTextureBinding},
		}})
		return nil
	}
	pass, err := enc.BeginRenderPass(desc)
	if err != nil {
		return err
	}
	pass.SetPipeline(r.shadowPipeline)
	pass.SetBindGroup(0, r.shadowGroup, nil)
	for _, cmd := range screen.worldMeshes {
		mesh := cmd.Mesh
		if mesh == nil || !mesh.options.DepthWrite {
			continue
		}
		gpuMesh, err := r.ensureWorldMesh(mesh)
		if err != nil {
			_ = pass.End()
			return err
		}
		r.bindModelGroup(pass, mesh)
		pass.SetVertexBuffer(0, gpuMesh.vertexBuf, 0)
		pass.SetIndexBuffer(gpuMesh.indexBuf, gputypes.IndexFormatUint32, 0)
		pass.DrawIndexed(gputypes.DrawIndexedArgs{IndexCount: gpuMesh.indexCount, InstanceCount: 1})
	}
	if err := r.drawShadowSprites(pass, screen); err != nil {
		_ = pass.End()
		return err
	}
	if err := pass.End(); err != nil {
		return err
	}
	enc.TransitionTextures([]wgpu.TextureBarrier{{
		Texture: r.shadowTex,
		Usage:   wgpu.TextureUsageTransition{OldUsage: wgpu.TextureUsageRenderAttachment, NewUsage: wgpu.TextureUsageTextureBinding},
	}})
	return nil
}

func (r *gpuRenderer) drawWorldScene(enc *wgpu.CommandEncoder, screen *Frame, worldTarget *wgpu.TextureView, clear gputypes.Color, meshBatches []worldMeshBatch, world worldFrame, worldVertexBuf, worldIndexBuf *wgpu.Buffer, ssaoActive bool) error {
	if _, ok := PassByName(r.framePasses, "directional-shadow"); ok {
		if err := r.drawDirectionalShadow(enc, screen); err != nil {
			return err
		}
	} else {
		r.releaseShadowSpriteGroups()
	}
	splitAO := ssaoActive && r.sampleCount == 1
	if _, ok := PassByName(r.framePasses, "ssao-estimate"); ok && r.sampleCount != 1 {
		splitAO = false
	}
	if splitAO {
		return r.drawWorldSceneWithAO(enc, screen, worldTarget, clear, meshBatches, world, worldVertexBuf, worldIndexBuf)
	}
	pass, err := enc.BeginRenderPass(r.worldPassDescriptor(worldTarget, clear))
	if err != nil {
		return err
	}
	if err := r.drawWorldContents(pass, screen, meshBatches, world, worldVertexBuf, worldIndexBuf); err != nil {
		_ = pass.End()
		return err
	}
	return pass.End()
}

func (r *gpuRenderer) drawWorldSceneWithAO(enc *wgpu.CommandEncoder, screen *Frame, worldTarget *wgpu.TextureView, clear gputypes.Color, meshBatches []worldMeshBatch, world worldFrame, worldVertexBuf, worldIndexBuf *wgpu.Buffer) error {
	opaque := r.worldPassDescriptor(worldTarget, clear)
	if opaque.DepthStencilAttachment != nil {
		opaque.DepthStencilAttachment.DepthStoreOp = gputypes.StoreOpStore
	}
	pass, err := enc.BeginRenderPass(opaque)
	if err != nil {
		return err
	}
	if err := r.drawWorldOpaque(pass, meshBatches, world, worldVertexBuf, worldIndexBuf, nil); err != nil {
		_ = pass.End()
		return err
	}
	if err := pass.End(); err != nil {
		return err
	}
	enc.TransitionTextures([]wgpu.TextureBarrier{{
		Texture: r.depthTarget.texture,
		Usage:   wgpu.TextureUsageTransition{OldUsage: wgpu.TextureUsageRenderAttachment, NewUsage: wgpu.TextureUsageTextureBinding},
	}})
	if err := r.ssao.ensure(r.dev, r.depthTarget.width, r.depthTarget.height, r.depthTarget.view); err != nil {
		return err
	}
	if err := r.drawSSAO(enc, nil, worldFrame{}, nil, nil); err != nil {
		return err
	}
	if err := r.drawSSAOComposite(enc, worldTarget); err != nil {
		return err
	}
	enc.TransitionTextures([]wgpu.TextureBarrier{{
		Texture: r.depthTarget.texture,
		Usage:   wgpu.TextureUsageTransition{OldUsage: wgpu.TextureUsageTextureBinding, NewUsage: wgpu.TextureUsageRenderAttachment},
	}})
	cont := r.worldContinueDescriptor(worldTarget)
	pass, err = enc.BeginRenderPass(cont)
	if err != nil {
		return err
	}
	if err := r.drawWorldRest(pass, screen, world, worldVertexBuf, worldIndexBuf); err != nil {
		_ = pass.End()
		return err
	}
	return pass.End()
}

func (r *gpuRenderer) worldContinueDescriptor(surface *wgpu.TextureView) *wgpu.RenderPassDescriptor {
	color := wgpu.RenderPassColorAttachment{
		View: surface, LoadOp: gputypes.LoadOpLoad, StoreOp: gputypes.StoreOpStore,
	}
	if r.sampleCount > 1 {
		color.View = r.msaaTarget.view
		color.ResolveTarget = surface
		color.StoreOp = gputypes.StoreOpDiscard
	}
	return &wgpu.RenderPassDescriptor{
		ColorAttachments: []wgpu.RenderPassColorAttachment{color},
		DepthStencilAttachment: &wgpu.RenderPassDepthStencilAttachment{
			View: r.depthTarget.view, DepthLoadOp: gputypes.LoadOpLoad, DepthStoreOp: gputypes.StoreOpDiscard,
		},
	}
}

func (r *gpuRenderer) drawSSAOComposite(enc *wgpu.CommandEncoder, target *wgpu.TextureView) error {
	pass, err := enc.BeginRenderPass(&wgpu.RenderPassDescriptor{
		ColorAttachments: []wgpu.RenderPassColorAttachment{{
			View: target, LoadOp: gputypes.LoadOpLoad, StoreOp: gputypes.StoreOpStore,
		}},
	})
	if err != nil {
		return err
	}
	pass.SetPipeline(r.ssao.composite)
	pass.SetBindGroup(0, r.ssao.groups[1], nil)
	pass.Draw(gputypes.DrawArgs{VertexCount: 3, InstanceCount: 1})
	return pass.End()
}

func (r *gpuRenderer) drawWorldContents(pass *wgpu.RenderPassEncoder, screen *Frame, meshBatches []worldMeshBatch, world worldFrame, worldVertexBuf, worldIndexBuf *wgpu.Buffer) error {
	if err := r.drawWorldOpaque(pass, meshBatches, world, worldVertexBuf, worldIndexBuf, nil); err != nil {
		return err
	}
	return r.drawWorldRest(pass, screen, world, worldVertexBuf, worldIndexBuf)
}

func (r *gpuRenderer) drawWorldRest(pass *wgpu.RenderPassEncoder, screen *Frame, world worldFrame, worldVertexBuf, worldIndexBuf *wgpu.Buffer) error {
	lateGeometry, err := r.drawWorldAlpha(pass, world, screen.worldMeshes, worldVertexBuf, worldIndexBuf, false)
	if err != nil {
		return err
	}
	if err := r.drawWorldBillboards(pass, billboardColorPass); err != nil {
		return err
	}
	if lateGeometry {
		if err := r.drawWorldBillboards(pass, billboardDepthPass); err != nil {
			return err
		}
		if _, err := r.drawWorldAlpha(pass, world, screen.worldMeshes, worldVertexBuf, worldIndexBuf, true); err != nil {
			return err
		}
	}
	return r.drawWorldBillboards(pass, billboardOverlayPass)
}

func (r *gpuRenderer) initShadowSpritePipeline() {
	if r == nil || r.dev == nil || r.shadowUniform == nil {
		return
	}
	bgl, err := r.dev.CreateBindGroupLayout(&wgpu.BindGroupLayoutDescriptor{
		Label: "goro-shadow-sprite-layout",
		Entries: []gputypes.BindGroupLayoutEntry{
			{Binding: 0, Visibility: wgpu.ShaderStageVertex, Buffer: &gputypes.BufferBindingLayout{Type: gputypes.BufferBindingTypeUniform, MinBindingSize: shadowUniformBytes}},
			{Binding: 1, Visibility: wgpu.ShaderStageFragment, Sampler: &gputypes.SamplerBindingLayout{Type: gputypes.SamplerBindingTypeFiltering}},
			{Binding: 2, Visibility: wgpu.ShaderStageFragment, Texture: &gputypes.TextureBindingLayout{SampleType: gputypes.TextureSampleTypeFloat, ViewDimension: gputypes.TextureViewDimension2D}},
		},
	})
	if err != nil {
		return
	}
	layout, err := r.dev.CreatePipelineLayout(&wgpu.PipelineLayoutDescriptor{
		Label:            "goro-shadow-sprite-pipeline-layout",
		BindGroupLayouts: []*wgpu.BindGroupLayout{bgl},
	})
	if err != nil {
		bgl.Release()
		return
	}
	shader, err := r.dev.CreateShaderModule(&wgpu.ShaderModuleDescriptor{Label: "goro-shadow-sprite", WGSL: shadowSpriteShaderWGSL})
	if err != nil {
		layout.Release()
		bgl.Release()
		return
	}
	defer shader.Release()
	pipeline, err := r.dev.CreateRenderPipeline(&wgpu.RenderPipelineDescriptor{
		Label:  "goro-shadow-sprite-pipeline",
		Layout: layout,
		Vertex: wgpu.VertexState{
			Module:     shader,
			EntryPoint: "vs_main",
			Buffers: []gputypes.VertexBufferLayout{
				{
					ArrayStride: billboardVertexStride,
					StepMode:    gputypes.VertexStepModeVertex,
					Attributes: []gputypes.VertexAttribute{
						{Format: gputypes.VertexFormatFloat32x2, Offset: 0, ShaderLocation: 0},
						{Format: gputypes.VertexFormatFloat32x2, Offset: 8, ShaderLocation: 1},
					},
				},
				{
					ArrayStride: billboardInstanceStride,
					StepMode:    gputypes.VertexStepModeInstance,
					Attributes: []gputypes.VertexAttribute{
						{Format: gputypes.VertexFormatFloat32x3, Offset: 0, ShaderLocation: 2},
						{Format: gputypes.VertexFormatFloat32x3, Offset: 12, ShaderLocation: 3},
						{Format: gputypes.VertexFormatFloat32x3, Offset: 24, ShaderLocation: 4},
						{Format: gputypes.VertexFormatFloat32x4, Offset: 48, ShaderLocation: 6},
					},
				},
			},
		},
		Primitive: gputypes.PrimitiveState{
			Topology:  gputypes.PrimitiveTopologyTriangleList,
			FrontFace: gputypes.FrontFaceCCW,
			CullMode:  gputypes.CullModeNone,
		},
		DepthStencil: &wgpu.DepthStencilState{
			Format:            depthFormat,
			DepthWriteEnabled: true,
			DepthCompare:      gputypes.CompareFunctionLessEqual,
		},
		Multisample: gputypes.MultisampleState{Count: 1, Mask: 0xFFFFFFFF},
		Fragment: &wgpu.FragmentState{
			Module:     shader,
			EntryPoint: "fs_main",
			Targets: []gputypes.ColorTargetState{{
				Format:    shadowClearColorFormat,
				WriteMask: gputypes.ColorWriteMaskNone,
			}},
		},
	})
	if err != nil {
		layout.Release()
		bgl.Release()
		return
	}
	r.shadowSpriteBGL = bgl
	r.shadowSpriteLayout = layout
	r.shadowSpritePipeline = pipeline
}

func (r *gpuRenderer) releaseShadowSpriteGroups() {
	if r == nil {
		return
	}
	for _, bg := range r.shadowSpriteGroups {
		if bg != nil {
			bg.Release()
		}
	}
	r.shadowSpriteGroups = r.shadowSpriteGroups[:0]
}

// shadowCasterBillboards keeps actor body quads. Effect billboards leave CastShadow false.
func shadowCasterBillboards(commands []WorldBillboardCommand) []WorldBillboardCommand {
	if len(commands) == 0 {
		return nil
	}
	out := make([]WorldBillboardCommand, 0, len(commands))
	for _, cmd := range commands {
		if !cmd.Options.CastShadow || cmd.Texture == nil || cmd.Texture.pix == nil || cmd.Width <= 0 || cmd.Height <= 0 {
			continue
		}
		out = append(out, cmd)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

type shadowSpriteBatch struct {
	tex   *Image
	first uint32
	count uint32
}

func (r *gpuRenderer) drawShadowSprites(pass *wgpu.RenderPassEncoder, screen *Frame) error {
	if r == nil || r.shadowSpritePipeline == nil || r.billboardQuadBuf == nil || r.shadowUniform == nil || screen == nil {
		return nil
	}
	casters := shadowCasterBillboards(screen.worldBillboards)
	if len(casters) == 0 {
		return nil
	}
	floats := make([]float32, 0, len(casters)*billboardInstanceFloatCount)
	batches := make([]shadowSpriteBatch, 0, 4)
	for _, cmd := range casters {
		if len(batches) == 0 || batches[len(batches)-1].tex != cmd.Texture {
			batches = append(batches, shadowSpriteBatch{tex: cmd.Texture, first: uint32(len(floats) / billboardInstanceFloatCount)})
		}
		floats = appendBillboardInstanceData(floats, cmd)
		batches[len(batches)-1].count++
	}
	raw := floatBytes(floats)
	buf, err := r.dynamicBuffer(&r.shadowSpriteSlot, "goro-shadow-sprites", len(raw), wgpu.BufferUsageVertex|wgpu.BufferUsageCopyDst, raw)
	if err != nil {
		return err
	}
	sampler, err := r.sampler(DrawTrianglesOptions{Filter: FilterNearest, Address: AddressClampToEdge}, nil)
	if err != nil {
		return err
	}
	pass.SetPipeline(r.shadowSpritePipeline)
	pass.SetVertexBuffer(0, r.billboardQuadBuf, 0)
	for _, batch := range batches {
		tex, err := r.ensureTexture(batch.tex)
		if err != nil {
			return err
		}
		bg, err := r.dev.CreateBindGroup(&wgpu.BindGroupDescriptor{
			Label:  "goro-shadow-sprite-bind",
			Layout: r.shadowSpriteBGL,
			Entries: []wgpu.BindGroupEntry{
				{Binding: 0, Buffer: r.shadowUniform, Size: shadowUniformBytes},
				{Binding: 1, Sampler: sampler},
				{Binding: 2, TextureView: tex.view},
			},
		})
		if err != nil {
			return err
		}
		r.shadowSpriteGroups = append(r.shadowSpriteGroups, bg)
		pass.SetBindGroup(0, bg, nil)
		pass.SetVertexBuffer(1, buf, uint64(batch.first)*billboardInstanceStride)
		pass.Draw(gputypes.DrawArgs{VertexCount: 6, InstanceCount: batch.count})
	}
	return nil
}
