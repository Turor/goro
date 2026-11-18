package render

import (
	_ "embed"
	"fmt"
	"math"

	"github.com/gogpu/gputypes"
	"github.com/gogpu/wgpu"
)

//go:embed ssao.wgsl
var ssaoShaderWGSL string

//go:embed ssao_depth.wgsl
var ssaoDepthShaderWGSL string

// gpuSSAO keeps its geometry depth separate from the world's MSAA attachment.
// Only opaque geometry participates; composition precedes water and sprites.
type gpuSSAO struct {
	bindings   *wgpu.BindGroupLayout
	layout     *wgpu.PipelineLayout
	uniform    *wgpu.Buffer
	geometry   *wgpu.RenderPipeline
	estimate   *wgpu.RenderPipeline
	composite  *wgpu.RenderPipeline
	occlusion  gpuRenderTarget
	groups     [2]*wgpu.BindGroup
	boundDepth *wgpu.TextureView
}

func (r *gpuRenderer) newSSAO() (_ *gpuSSAO, err error) {
	a := &gpuSSAO{}
	defer func() {
		if err != nil {
			a.release()
		}
	}()
	a.bindings, err = r.dev.CreateBindGroupLayout(&wgpu.BindGroupLayoutDescriptor{
		Label: "goro-ssao-bindings",
		Entries: []gputypes.BindGroupLayoutEntry{
			{Binding: 0, Visibility: wgpu.ShaderStageFragment, Buffer: &gputypes.BufferBindingLayout{Type: gputypes.BufferBindingTypeUniform, MinBindingSize: 80}},
			{Binding: 1, Visibility: wgpu.ShaderStageFragment, Texture: &gputypes.TextureBindingLayout{SampleType: gputypes.TextureSampleTypeDepth, ViewDimension: gputypes.TextureViewDimension2D}},
			{Binding: 2, Visibility: wgpu.ShaderStageFragment, Texture: &gputypes.TextureBindingLayout{SampleType: gputypes.TextureSampleTypeFloat, ViewDimension: gputypes.TextureViewDimension2D}},
		},
	})
	if err != nil {
		return nil, err
	}
	a.layout, err = r.dev.CreatePipelineLayout(&wgpu.PipelineLayoutDescriptor{Label: "goro-ssao-layout", BindGroupLayouts: []*wgpu.BindGroupLayout{a.bindings}})
	if err != nil {
		return nil, err
	}
	a.uniform, err = r.dev.CreateBuffer(&wgpu.BufferDescriptor{Label: "goro-ssao-uniform", Size: 80, Usage: wgpu.BufferUsageUniform | wgpu.BufferUsageCopyDst})
	if err != nil {
		return nil, err
	}
	post, err := r.dev.CreateShaderModule(&wgpu.ShaderModuleDescriptor{Label: "goro-ssao", WGSL: ssaoShaderWGSL})
	if err != nil {
		return nil, err
	}
	defer post.Release()
	desc := &wgpu.RenderPipelineDescriptor{
		Label: "goro-ssao-estimate", Layout: a.layout,
		Vertex:    wgpu.VertexState{Module: post, EntryPoint: "fullscreen"},
		Primitive: gputypes.PrimitiveState{Topology: gputypes.PrimitiveTopologyTriangleList},
		Fragment: &wgpu.FragmentState{Module: post, EntryPoint: "estimate", Targets: []gputypes.ColorTargetState{{
			Format: gputypes.TextureFormatRGBA8Unorm, WriteMask: gputypes.ColorWriteMaskAll,
		}}},
	}
	a.estimate, err = r.dev.CreateRenderPipeline(desc)
	if err != nil {
		return nil, err
	}
	desc.Label = "goro-ssao-composite"
	desc.Fragment.EntryPoint = "composite"
	desc.Fragment.Targets[0].Format = r.format
	desc.Fragment.Targets[0].Blend = &gputypes.BlendState{
		Color: gputypes.BlendComponent{SrcFactor: gputypes.BlendFactorZero, DstFactor: gputypes.BlendFactorSrc, Operation: gputypes.BlendOperationAdd},
		Alpha: gputypes.BlendComponent{SrcFactor: gputypes.BlendFactorZero, DstFactor: gputypes.BlendFactorOne, Operation: gputypes.BlendOperationAdd},
	}
	desc.Multisample = gputypes.MultisampleState{Count: 1, Mask: 0xFFFFFFFF}
	desc.DepthStencil = nil
	a.composite, err = r.dev.CreateRenderPipeline(desc)
	if err != nil {
		return nil, err
	}
	return a, nil
}

func (a *gpuSSAO) ensure(dev *wgpu.Device, width, height int, depthView *wgpu.TextureView) error {
	if depthView == nil {
		return fmt.Errorf("SSAO estimate needs the opaque pass depth")
	}
	halfW, halfH := (width+1)/2, (height+1)/2
	if a.groups[1] != nil && a.occlusion.width == halfW && a.occlusion.height == halfH && a.boundDepth == depthView {
		return nil
	}
	a.releaseTargets()
	// The opaque world pass already wrote hardware depth. The estimate reads
	// that attachment and does not draw the geometry again.
	if err := a.occlusion.ensure(dev, halfW, halfH, gputypes.TextureFormatRGBA8Unorm, 1, wgpu.TextureUsageRenderAttachment|wgpu.TextureUsageTextureBinding); err != nil {
		return err
	}
	for i, source := range []*wgpu.TextureView{a.occlusion.view, a.occlusion.view} {
		var err error
		a.groups[i], err = dev.CreateBindGroup(&wgpu.BindGroupDescriptor{
			Label: "goro-ssao-input", Layout: a.bindings,
			Entries: []wgpu.BindGroupEntry{
				{Binding: 0, Buffer: a.uniform, Size: 80},
				{Binding: 1, TextureView: depthView},
				{Binding: 2, TextureView: source},
			},
		})
		if err != nil {
			return err
		}
	}
	// Composite reads the occlusion texture in binding 2.
	group, err := dev.CreateBindGroup(&wgpu.BindGroupDescriptor{
		Label: "goro-ssao-composite", Layout: a.bindings,
		Entries: []wgpu.BindGroupEntry{
			{Binding: 0, Buffer: a.uniform, Size: 80},
			{Binding: 1, TextureView: depthView},
			{Binding: 2, TextureView: a.occlusion.view},
		},
	})
	if err != nil {
		return err
	}
	if a.groups[1] != nil {
		a.groups[1].Release()
	}
	a.groups[1] = group
	a.boundDepth = depthView
	return nil
}

func (a *gpuSSAO) uploadCamera(queue *wgpu.Queue, camera Camera3D) error {
	inverse, ok := inverseViewProjection(camera.ViewProjection)
	if !ok {
		return fmt.Errorf("SSAO camera matrix is not invertible")
	}
	values := [20]float32{}
	copy(values[:16], inverse[:])
	values[16], values[17] = 2, 0.55
	return queue.WriteBuffer(a.uniform, 0, floatBytes(values[:]))
}

// drawSSAO estimates occlusion from the opaque pass depth. It does not submit
// that geometry again; the caller stored the world depth and transitioned it
// to a texture binding.
func (r *gpuRenderer) drawSSAO(enc *wgpu.CommandEncoder, meshes []worldMeshBatch, world worldFrame, vertices, indices *wgpu.Buffer) error {
	_ = meshes
	_ = world
	_ = vertices
	_ = indices
	a := r.ssao
	pass, err := enc.BeginRenderPass(&wgpu.RenderPassDescriptor{
		ColorAttachments: []wgpu.RenderPassColorAttachment{{View: a.occlusion.view, LoadOp: gputypes.LoadOpClear, StoreOp: gputypes.StoreOpStore}},
	})
	if err != nil {
		return err
	}
	pass.SetPipeline(a.estimate)
	pass.SetBindGroup(0, a.groups[0], nil)
	pass.Draw(gputypes.DrawArgs{VertexCount: 3, InstanceCount: 1})
	if err := pass.End(); err != nil {
		return err
	}
	enc.TransitionTextures([]wgpu.TextureBarrier{{Texture: a.occlusion.texture, Usage: wgpu.TextureUsageTransition{OldUsage: wgpu.TextureUsageRenderAttachment, NewUsage: wgpu.TextureUsageTextureBinding}}})
	return nil
}

func (a *gpuSSAO) releaseTargets() {
	for i, group := range a.groups {
		if group != nil {
			group.Release()
			a.groups[i] = nil
		}
	}
	a.occlusion.release()
	a.boundDepth = nil
}

func (a *gpuSSAO) release() {
	a.releaseTargets()
	for _, pipeline := range []*wgpu.RenderPipeline{a.geometry, a.estimate, a.composite} {
		if pipeline != nil {
			pipeline.Release()
		}
	}
	if a.uniform != nil {
		a.uniform.Release()
	}
	if a.layout != nil {
		a.layout.Release()
	}
	if a.bindings != nil {
		a.bindings.Release()
	}
}

// inverseViewProjection inverts column-major camera matrices with pivoting.
func inverseViewProjection(matrix [16]float32) ([16]float32, bool) {
	var rows [4][8]float64
	for row := range 4 {
		for col := range 4 {
			rows[row][col] = float64(matrix[col*4+row])
		}
		rows[row][row+4] = 1
	}
	for col := range 4 {
		pivot := col
		for row := col + 1; row < 4; row++ {
			if math.Abs(rows[row][col]) > math.Abs(rows[pivot][col]) {
				pivot = row
			}
		}
		divisor := rows[pivot][col]
		if divisor == 0 || math.IsNaN(divisor) || math.IsInf(divisor, 0) {
			return [16]float32{}, false
		}
		rows[col], rows[pivot] = rows[pivot], rows[col]
		for i := range 8 {
			rows[col][i] /= divisor
		}
		for row := range 4 {
			if row == col {
				continue
			}
			scale := rows[row][col]
			for i := range 8 {
				rows[row][i] -= rows[col][i] * scale
			}
		}
	}
	var inverse [16]float32
	for row := range 4 {
		for col := range 4 {
			value := rows[row][col+4]
			if math.IsNaN(value) || math.IsInf(value, 0) || math.Abs(value) > math.MaxFloat32 {
				return [16]float32{}, false
			}
			inverse[col*4+row] = float32(value)
		}
	}
	return inverse, true
}
