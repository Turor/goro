package render

import (
	_ "embed"

	"github.com/gogpu/gputypes"
	"github.com/gogpu/wgpu"
)

//go:embed fxaa.wgsl
var fxaaShaderWGSL string

// gpuFXAA is one fullscreen pass over the resolved world color.
type gpuFXAA struct {
	bindings *wgpu.BindGroupLayout
	layout   *wgpu.PipelineLayout
	pipeline *wgpu.RenderPipeline
	sampler  *wgpu.Sampler
	source   gpuRenderTarget
	group    *wgpu.BindGroup
	bound    *wgpu.TextureView
}

func (r *gpuRenderer) ensureFXAA(width, height int) error {
	if r.fxaa == nil {
		fx, err := newGPUFXAA(r)
		if err != nil {
			return err
		}
		r.fxaa = fx
	}
	if err := r.fxaa.source.ensure(r.dev, width, height, r.format, 1, wgpu.TextureUsageRenderAttachment|wgpu.TextureUsageTextureBinding); err != nil {
		return err
	}
	if r.fxaa.group != nil && r.fxaa.bound == r.fxaa.source.view {
		return nil
	}
	if r.fxaa.group != nil {
		r.fxaa.group.Release()
		r.fxaa.group = nil
	}
	group, err := r.dev.CreateBindGroup(&wgpu.BindGroupDescriptor{
		Label:  "goro-fxaa-bind",
		Layout: r.fxaa.bindings,
		Entries: []wgpu.BindGroupEntry{
			{Binding: 0, TextureView: r.fxaa.source.view},
			{Binding: 1, Sampler: r.fxaa.sampler},
		},
	})
	if err != nil {
		return err
	}
	r.fxaa.group = group
	r.fxaa.bound = r.fxaa.source.view
	return nil
}

func newGPUFXAA(r *gpuRenderer) (*gpuFXAA, error) {
	fx := &gpuFXAA{}
	var err error
	fx.bindings, err = r.dev.CreateBindGroupLayout(&wgpu.BindGroupLayoutDescriptor{
		Label: "goro-fxaa-layout",
		Entries: []gputypes.BindGroupLayoutEntry{
			{Binding: 0, Visibility: wgpu.ShaderStageFragment, Texture: &gputypes.TextureBindingLayout{SampleType: gputypes.TextureSampleTypeFloat, ViewDimension: gputypes.TextureViewDimension2D}},
			{Binding: 1, Visibility: wgpu.ShaderStageFragment, Sampler: &gputypes.SamplerBindingLayout{Type: gputypes.SamplerBindingTypeFiltering}},
		},
	})
	if err != nil {
		return nil, err
	}
	fx.layout, err = r.dev.CreatePipelineLayout(&wgpu.PipelineLayoutDescriptor{
		Label:            "goro-fxaa-pipeline-layout",
		BindGroupLayouts: []*wgpu.BindGroupLayout{fx.bindings},
	})
	if err != nil {
		return nil, err
	}
	fx.sampler, err = r.dev.CreateSampler(&wgpu.SamplerDescriptor{
		Label:     "goro-fxaa-sampler",
		MagFilter: gputypes.FilterModeLinear,
		MinFilter: gputypes.FilterModeLinear,
	})
	if err != nil {
		return nil, err
	}
	shader, err := r.dev.CreateShaderModule(&wgpu.ShaderModuleDescriptor{Label: "goro-fxaa", WGSL: fxaaShaderWGSL})
	if err != nil {
		return nil, err
	}
	defer shader.Release()
	fx.pipeline, err = r.dev.CreateRenderPipeline(&wgpu.RenderPipelineDescriptor{
		Label:     "goro-fxaa-pipeline",
		Layout:    fx.layout,
		Vertex:    wgpu.VertexState{Module: shader, EntryPoint: "fullscreen"},
		Primitive: gputypes.PrimitiveState{Topology: gputypes.PrimitiveTopologyTriangleList},
		Fragment: &wgpu.FragmentState{
			Module:     shader,
			EntryPoint: "fs_main",
			Targets: []gputypes.ColorTargetState{{
				Format:    r.format,
				WriteMask: gputypes.ColorWriteMaskAll,
			}},
		},
	})
	if err != nil {
		return nil, err
	}
	return fx, nil
}

func (fx *gpuFXAA) draw(enc *wgpu.CommandEncoder, dest *wgpu.TextureView) error {
	if fx == nil || fx.pipeline == nil || fx.source.texture == nil || dest == nil {
		return nil
	}
	enc.TransitionTextures([]wgpu.TextureBarrier{{
		Texture: fx.source.texture,
		Usage:   wgpu.TextureUsageTransition{OldUsage: wgpu.TextureUsageRenderAttachment, NewUsage: wgpu.TextureUsageTextureBinding},
	}})
	pass, err := enc.BeginRenderPass(&wgpu.RenderPassDescriptor{
		ColorAttachments: []wgpu.RenderPassColorAttachment{{
			View: dest, LoadOp: gputypes.LoadOpClear, StoreOp: gputypes.StoreOpStore,
		}},
	})
	if err != nil {
		return err
	}
	pass.SetPipeline(fx.pipeline)
	pass.SetBindGroup(0, fx.group, nil)
	pass.Draw(gputypes.DrawArgs{VertexCount: 3, InstanceCount: 1})
	if err := pass.End(); err != nil {
		return err
	}
	enc.TransitionTextures([]wgpu.TextureBarrier{{
		Texture: fx.source.texture,
		Usage:   wgpu.TextureUsageTransition{OldUsage: wgpu.TextureUsageTextureBinding, NewUsage: wgpu.TextureUsageRenderAttachment},
	}})
	return nil
}

func (fx *gpuFXAA) release() {
	if fx == nil {
		return
	}
	if fx.group != nil {
		fx.group.Release()
	}
	if fx.pipeline != nil {
		fx.pipeline.Release()
	}
	if fx.sampler != nil {
		fx.sampler.Release()
	}
	if fx.layout != nil {
		fx.layout.Release()
	}
	if fx.bindings != nil {
		fx.bindings.Release()
	}
	fx.source.release()
}
