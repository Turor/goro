package render

import (
	"testing"

	"github.com/gogpu/gputypes"
	"github.com/gogpu/wgpu"
)

// TestShadowMapInitClear drives the shadow pass renderer init clears with.
// A depth-only pass is one this Vulkan backend cannot end: BeginRenderPass
// returns before vkCmdBeginRenderPass, and End still calls vkCmdEndRenderPass.
func TestShadowMapInitClear(t *testing.T) {
	depth := &wgpu.TextureView{}
	color := &wgpu.TextureView{}
	init := &gpuRenderer{shadowView: depth, shadowColorView: color}
	pass := init.shadowMapPass()
	if !vulkanPassEndable(pass) {
		t.Fatal("renderer init shadow clear builds a pass the Vulkan backend cannot end")
	}
	got, ok := sampledShadowDepthBeforeDraw(pass)
	if !ok || got != shadowClearDepth {
		t.Fatalf("sampled shadow depth before any shadow draw = %v ok=%v, want %v", got, ok, shadowClearDepth)
	}
	if pass.DepthStencilAttachment.View != depth {
		t.Fatal("shadow clear does not store the shadow depth image")
	}

	depthOnly := &gpuRenderer{shadowView: depth}
	if vulkanPassEndable(depthOnly.shadowMapPass()) {
		t.Fatal("depth-only clear was treated as endable; End would call vkCmdEndRenderPass on a pass that was not begun")
	}
	if _, stored := sampledShadowDepthBeforeDraw(depthOnly.shadowMapPass()); stored {
		t.Fatal("depth-only clear was treated as a stored depth")
	}
	bare := &wgpu.RenderPassDescriptor{
		DepthStencilAttachment: &wgpu.RenderPassDepthStencilAttachment{
			View:            depth,
			DepthLoadOp:     gputypes.LoadOpClear,
			DepthStoreOp:    gputypes.StoreOpStore,
			DepthClearValue: shadowClearDepth,
		},
	}
	if vulkanPassEndable(bare) {
		t.Fatal("depth-only descriptor is a pass Vulkan End cannot complete")
	}
	t.Logf("init shadow clear endable sampledDepth=%v before draw", got)
}
