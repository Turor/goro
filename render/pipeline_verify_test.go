package render

import (
	"encoding/binary"
	"image/color"
	"math"
	"strings"
	"testing"
)

func TestFrameSubmission(t *testing.T) {
	empty := NewFrame(8, 8)
	empty.BeginFrame()
	renderer := &gpuRenderer{textureArrays: true}
	emptySub := renderer.recordFrameSubmission(empty)
	if len(emptySub.BillboardDraws) != 0 || len(emptySub.MeshDraws) != 0 {
		t.Fatalf("empty frame draws billboards=%d meshes=%d", len(emptySub.BillboardDraws), len(emptySub.MeshDraws))
	}

	red := solidImage(2, 2, color.RGBA{R: 255, A: 255})
	blue := solidImage(2, 2, color.RGBA{B: 255, A: 255})
	screen := NewFrame(64, 64)
	fillSubmissionFrame(screen, red, blue, 0)
	arrays := &gpuRenderer{textureArrays: true, sampleCount: 1}
	packed := arrays.recordFrameSubmission(screen)
	if len(packed.BillboardDraws) != 1 || packed.BillboardDraws[0].Kind != "instanced" || len(packed.BillboardDraws[0].Textures) != 2 {
		t.Fatalf("texture arrays billboards = %#v", packed.BillboardDraws)
	}
	if len(packed.MeshDraws) != 1 || !packed.MeshDraws[0].MultiDraw || len(packed.MeshDraws[0].Textures) != 2 {
		t.Fatalf("texture arrays meshes = %#v", packed.MeshDraws)
	}

	splitRenderer := &gpuRenderer{sampleCount: 1}
	split := splitRenderer.recordFrameSubmission(screen)
	if len(split.BillboardDraws) != 2 {
		t.Fatalf("per-texture billboards = %#v", split.BillboardDraws)
	}
	for _, draw := range split.BillboardDraws {
		if draw.Kind != "texture" || len(draw.Textures) != 1 {
			t.Fatalf("per-texture billboard draw = %#v", draw)
		}
	}
	if len(split.MeshDraws) != 2 {
		t.Fatalf("per-texture meshes = %#v", split.MeshDraws)
	}
	t.Logf("empty draws=0 array billboards=%d meshes=%d split billboards=%d meshes=%d", len(packed.BillboardDraws), len(packed.MeshDraws), len(split.BillboardDraws), len(split.MeshDraws))
}

func TestDynamicUpload(t *testing.T) {
	red := solidImage(2, 2, color.RGBA{R: 255, A: 255})
	blue := solidImage(2, 2, color.RGBA{B: 255, A: 255})
	screen := NewFrame(64, 64)
	renderer := &gpuRenderer{textureArrays: true, sampleCount: 1}
	fillSubmissionFrame(screen, red, blue, 0.25)
	first := renderer.recordFrameSubmission(screen)
	firstBillboards := append([]byte(nil), renderer.dynamicUploads.Bytes(DynamicBillboards, first.BillboardSlot)...)
	firstTriangles := append([]byte(nil), renderer.dynamicUploads.Bytes(DynamicTriangles, first.TriangleSlot)...)
	if len(firstBillboards) == 0 || len(firstTriangles) == 0 {
		t.Fatalf("first frame dynamic bytes billboards=%d triangles=%d", len(firstBillboards), len(firstTriangles))
	}

	renderer.frameSerial = 1
	fillSubmissionFrame(screen, red, blue, 0.75)
	second := renderer.recordFrameSubmission(screen)
	if first.BillboardSlot == second.BillboardSlot || first.TriangleSlot == second.TriangleSlot {
		t.Fatalf("slots did not rotate billboard %d->%d triangle %d->%d", first.BillboardSlot, second.BillboardSlot, first.TriangleSlot, second.TriangleSlot)
	}
	keptBillboards := renderer.dynamicUploads.Bytes(DynamicBillboards, first.BillboardSlot)
	keptTriangles := renderer.dynamicUploads.Bytes(DynamicTriangles, first.TriangleSlot)
	nextBillboards := renderer.dynamicUploads.Bytes(DynamicBillboards, second.BillboardSlot)
	nextTriangles := renderer.dynamicUploads.Bytes(DynamicTriangles, second.TriangleSlot)
	if &keptBillboards[0] == &nextBillboards[0] || &keptTriangles[0] == &nextTriangles[0] {
		t.Fatal("second frame reused the buffer written by the first frame")
	}
	if string(keptBillboards) != string(firstBillboards) || string(keptTriangles) != string(firstTriangles) {
		t.Fatal("second frame overwrote the first frame slot")
	}
	if string(nextTriangles) == string(firstTriangles) {
		t.Fatal("second frame triangle payload did not change")
	}
	t.Logf("frame0 slots billboard=%d triangle=%d frame1 slots billboard=%d triangle=%d", first.BillboardSlot, first.TriangleSlot, second.BillboardSlot, second.TriangleSlot)
}

func TestSSAOPass(t *testing.T) {
	screen := NewFrame(64, 64)
	screen.BeginFrame()
	screen.SetCamera3D(Camera3D{Enabled: true})
	renderer := &gpuRenderer{ssaoEnabled: true, sampleCount: 1}
	renderer.recordFrameSubmission(screen)
	if OpaqueGeometrySubmissions(renderer.framePasses) != 1 {
		t.Fatalf("opaque submissions = %d, want 1", OpaqueGeometrySubmissions(renderer.framePasses))
	}
	estimate, ok := PassByName(renderer.framePasses, "ssao-estimate")
	if !ok || estimate.DepthInput != "opaque-depth" || estimate.SubmitsOpaqueGeometry {
		t.Fatalf("ssao estimate = %#v ok=%t", estimate, ok)
	}
	if _, ok := PassByName(renderer.framePasses, "ssao-composite"); !ok {
		t.Fatal("ssao composite pass missing")
	}
	t.Logf("passes=%v opaque=%d", passNames(renderer.framePasses), OpaqueGeometrySubmissions(renderer.framePasses))
}

func TestFXAAPass(t *testing.T) {
	screen := NewFrame(64, 64)
	screen.BeginFrame()
	screen.SetCamera3D(Camera3D{Enabled: true})
	renderer := &gpuRenderer{fxaaEnabled: true, sampleCount: 1}
	renderer.recordFrameSubmission(screen)
	if !FXAAFollowsResolve(renderer.framePasses) {
		t.Fatalf("fxaa on passes = %v", passNames(renderer.framePasses))
	}
	renderer.fxaaEnabled = false
	renderer.recordFrameSubmission(screen)
	if FXAAFollowsResolve(renderer.framePasses) {
		t.Fatalf("fxaa off passes = %v", passNames(renderer.framePasses))
	}
	if _, ok := PassByName(renderer.framePasses, "fxaa"); ok {
		t.Fatal("fxaa pass present while disabled")
	}
	t.Logf("fxaa off passes=%v", passNames(renderer.framePasses))
}

func TestEnhancedShade(t *testing.T) {
	base := [3]float32{0.2, 0.2, 0.2}
	lightmap := [4]float32{0.25, 0.1, 0.05, 1}
	sun := SunLight{Direction: [3]float32{0, -1, 0}, Color: [3]float32{0.8, 0.2, 0.1}, Ambient: [3]float32{0.2, 0.3, 0.4}}
	facing := Shade(base, lightmap, [3]float32{0, 1, 0}, sun, nil, [3]float32{}, false, true)
	away := Shade(base, lightmap, [3]float32{0, -1, 0}, sun, nil, [3]float32{}, false, true)
	if facing[0] <= away[0] {
		t.Fatalf("sun-facing red %.3f is not brighter than away %.3f", facing[0], away[0])
	}
	occluded := Shade(base, lightmap, [3]float32{0, 1, 0}, sun, nil, [3]float32{}, true, true)
	if occluded[0] >= facing[0] {
		t.Fatalf("occluded red %.3f is not darker than lit %.3f", occluded[0], facing[0])
	}
	if !ShadowOccludes(0.5, 0.4, 0) || ShadowOccludes(0.4, 0.4, ShadowBias) {
		t.Fatal("shadow occlusion compare did not match the shader rule")
	}
	points := make([]PointLight, 10)
	for i := range points {
		points[i] = PointLight{Position: [3]float32{float32(i), 0, 0}, Color: [3]float32{1, 1, 1}, Range: 4, Intensity: float32(i + 1)}
	}
	selected := SelectPointLights(points, MaxPointLights)
	if len(selected) != MaxPointLights {
		t.Fatalf("selected %d lights, cap %d", len(selected), MaxPointLights)
	}
	if !pointLightIn(selected, points[len(points)-1]) || pointLightIn(selected, points[0]) {
		t.Fatalf("selection kept a weak light or dropped the strongest: %#v", selected)
	}
	near := []PointLight{{Position: [3]float32{}, Color: [3]float32{1, 1, 1}, Range: 0.05, Intensity: 1}}
	tiles := AssignTileLights(near, IdentityViewProjection(), LightTileCount, LightTileCount)
	if len(tiles[0]) != 0 {
		t.Fatalf("tile 0 contains a light that misses it: %v", tiles[0])
	}
	if !tileHasLight(tiles) {
		t.Fatal("point light was absent from every tile")
	}
	off := Shade(base, lightmap, [3]float32{0, 1, 0}, sun, selected, [3]float32{1, 0, 0}, true, false)
	lightmapOnly := Shade(base, lightmap, [3]float32{}, SunLight{}, nil, [3]float32{}, false, false)
	if off != lightmapOnly {
		t.Fatalf("enhanced off = %#v, lightmap only = %#v", off, lightmapOnly)
	}
	if lightmapOnly[0] <= base[0] {
		t.Fatalf("lightmap did not contribute: %#v", lightmapOnly)
	}
	rawOff := EncodeLightingUniform(sun, points, IdentityViewProjection(), IdentityViewProjection(), 64, 64, false)
	rawOn := EncodeLightingUniform(sun, points, IdentityViewProjection(), IdentityViewProjection(), 64, 64, true)
	if math.Float32frombits(binary.LittleEndian.Uint32(rawOff[12:16])) != 0 {
		t.Fatal("enhanced off still sets the sun enable flag")
	}
	if math.Float32frombits(binary.LittleEndian.Uint32(rawOn[12:16])) != 1 {
		t.Fatal("enhanced on does not set the sun enable flag")
	}
	if math.Float32frombits(binary.LittleEndian.Uint32(rawOn[112:116])) != float32(MaxPointLights) {
		t.Fatal("uniform light count exceeds or misses the cap")
	}
	encodedTiles := AssignTileLights(SelectPointLights(points, MaxPointLights), IdentityViewProjection(), LightTileCount, LightTileCount)
	if binary.LittleEndian.Uint32(rawOn[384:388]) != uint32(len(encodedTiles[0])) {
		t.Fatalf("uniform tile 0 count %d does not match AssignTileLights %d", binary.LittleEndian.Uint32(rawOn[384:388]), len(encodedTiles[0]))
	}
	if math.Float32frombits(binary.LittleEndian.Uint32(rawOn[28:32])) != sun.Ambient[2] {
		t.Fatal("ambient blue is not packed in sun_color.w")
	}
	if math.Float32frombits(binary.LittleEndian.Uint32(rawOn[120:124])) != sun.Ambient[0] || math.Float32frombits(binary.LittleEndian.Uint32(rawOn[124:128])) != sun.Ambient[1] {
		t.Fatal("ambient red/green are not packed in shadow_params.zw")
	}
	ignored := Shade([3]float32{1, 1, 1}, [4]float32{1, 0, 0, 0}, [3]float32{0, 1, 0}, SunLight{}, nil, [3]float32{}, false, true)
	if ignored != ([3]float32{}) {
		t.Fatalf("enhanced on kept the lightmap: %#v", ignored)
	}
	hot := PointLight{Position: [3]float32{0, 0.2, 0}, Color: [3]float32{1, 1, 1}, Range: 8, Intensity: 10}
	white := Shade([3]float32{1, 1, 1}, lightmap, [3]float32{0, 1, 0}, sun, []PointLight{hot}, [3]float32{}, false, true)
	for i, channel := range white {
		if channel > 1 {
			t.Fatalf("enhanced shade channel %d blew out to %.3f", i, channel)
		}
	}
	dim := Shade(base, lightmap, [3]float32{0, 1, 0}, sun, []PointLight{hot}, [3]float32{}, false, true)
	for i, channel := range dim {
		if channel > base[i] {
			t.Fatalf("enhanced shade channel %d %.3f exceeds base %.3f", i, channel, base[i])
		}
	}
	above := Shade(base, [4]float32{}, [3]float32{0, 1, 0}, SunLight{}, []PointLight{{Position: [3]float32{0, 2, 0}, Color: [3]float32{1, 1, 1}, Range: 10, Intensity: 1}}, [3]float32{}, false, true)
	below := Shade(base, [4]float32{}, [3]float32{0, 1, 0}, SunLight{}, []PointLight{{Position: [3]float32{0, -2, 0}, Color: [3]float32{1, 1, 1}, Range: 10, Intensity: 1}}, [3]float32{}, false, true)
	if above[0] <= below[0] || below[0] != 0 {
		t.Fatalf("point N·L above %.3f below %.3f", above[0], below[0])
	}
	t.Logf("facing=%.3f away=%.3f occluded=%.3f selected=%d tile0=%d", facing[0], away[0], occluded[0], len(selected), len(tiles[0]))
}

func TestLightFacingAxes(t *testing.T) {
	ray := [3]float32{1, 0, 0}
	right, up := LightFacingAxes(2, ray)
	if absDot(right, ray) > 1e-4 || absDot(up, ray) > 1e-4 || absDot(right, up) > 1e-4 {
		t.Fatalf("light basis is not perpendicular to +X: right %v up %v", right, up)
	}
	rightLen := math.Sqrt(float64(dot3(right, right)))
	upLen := math.Sqrt(float64(dot3(up, up)))
	if math.Abs(rightLen-2) > 1e-4 || math.Abs(upLen-2) > 1e-4 {
		t.Fatalf("axis length right %.4f up %.4f, want 2", rightLen, upLen)
	}
	if zr, zu := LightFacingAxes(2, [3]float32{}); zr != ([3]float32{}) || zu != ([3]float32{}) {
		t.Fatalf("zero ray returned %v %v", zr, zu)
	}
}

func absDot(a, b [3]float32) float64 {
	return math.Abs(float64(dot3(a, b)))
}

func TestShadowCasterBillboards(t *testing.T) {
	img := solidImage(2, 2, color.RGBA{R: 255, G: 255, B: 255, A: 255})
	cmds := []WorldBillboardCommand{
		{Texture: img, Width: 4, Height: 8, Options: DrawTrianglesOptions{CastShadow: true, DepthWrite: true}},
		{Texture: img, Width: 4, Height: 8, Options: DrawTrianglesOptions{DepthWrite: true}},
		{Texture: img, Width: 4, Height: 8, Options: DrawTrianglesOptions{CastShadow: true}},
		{Width: 4, Height: 8, Options: DrawTrianglesOptions{CastShadow: true}},
	}
	got := shadowCasterBillboards(cmds)
	if len(got) != 2 {
		t.Fatalf("casters = %d, want the two textured CastShadow quads", len(got))
	}
}

func TestEnhancedLightShader(t *testing.T) {
	wantWorld := "enhanced_light(color.rgb, n, input.world_pos, input.clip.xy, true)"
	wantSprite := "enhanced_light(color.rgb, input.normal, input.world_pos, input.clip.xy, false)"
	if !strings.Contains(worldShaderWGSL, wantWorld) {
		t.Fatal("world shader does not use clamped enhanced light")
	}
	if !strings.Contains(worldBillboardShaderWGSL, wantSprite) || !strings.Contains(worldBillboardArrayShaderWGSL, wantSprite) {
		t.Fatal("sprite shaders still add light or receive billboard shadows")
	}
	if strings.Contains(lightingShaderWGSL, "sun_scale(") {
		t.Fatal("sun_scale is still in the enhanced shader")
	}
	if !strings.Contains(lightingShaderWGSL, "lighting.shadow_params.z, lighting.shadow_params.w, lighting.sun_color.w") {
		t.Fatal("ambient is not read from the packed uniform holes")
	}
	if !strings.Contains(shadowSpriteShaderWGSL, "color.a < 0.5") || !strings.Contains(shadowSpriteShaderWGSL, "fn light_facing(") {
		t.Fatal("sprite shadow shader does not alpha-test a light-facing quad")
	}
}

func TestTileRowMatchesShader(t *testing.T) {
	const screen = float32(64)
	bottom := PointLight{Position: [3]float32{0, -1, 0}, Color: [3]float32{1, 1, 1}, Range: 0.05, Intensity: 1}
	if !strings.Contains(lightingShaderWGSL, "pixel.y / screen.y * tiles.y") {
		t.Fatal("shader tile row does not read framebuffer y from the top")
	}
	pixelY := (1 - float32(-1)) * screen * 0.5
	if pixelY >= screen {
		pixelY = screen - 0.5
	}
	col, row := ShaderFramebufferTile(screen*0.5, pixelY, screen, screen, LightTileCount, LightTileCount)
	if row == 0 {
		t.Fatal("bottom framebuffer pixel mapped to tile row 0")
	}
	tiles := AssignTileLights([]PointLight{bottom}, IdentityViewProjection(), LightTileCount, LightTileCount)
	if !tileRowHas(tiles, row) {
		t.Fatalf("light at NDC y=-1 is absent from shader row %d", row)
	}
	if tileRowHas(tiles, 0) {
		t.Fatal("light at NDC y=-1 was assigned to the top tile row")
	}
	raw := EncodeLightingUniform(SunLight{}, []PointLight{bottom}, IdentityViewProjection(), IdentityViewProjection(), int(screen), int(screen), true)
	tile := row*LightTileCount + col
	if uniformTileCount(raw, tile) == 0 {
		t.Fatalf("uniform tile %d (row %d) does not carry the bottom light", tile, row)
	}
	if uniformTileCount(raw, col) != 0 {
		t.Fatal("uniform top row carries the bottom light")
	}
	t.Logf("NDC y=-1 shader row %d tile %d count %d", row, tile, uniformTileCount(raw, tile))
}

func tileRowHas(tiles [][]int, row int) bool {
	for x := 0; x < LightTileCount; x++ {
		slot := row*LightTileCount + x
		if slot >= 0 && slot < len(tiles) && len(tiles[slot]) > 0 {
			return true
		}
	}
	return false
}

func uniformTileCount(buf []byte, tile int) uint32 {
	offset := 384 + (tile/4)*16 + (tile%4)*4
	return binary.LittleEndian.Uint32(buf[offset : offset+4])
}

func fillSubmissionFrame(screen *Frame, red, blue *Image, triX float32) {
	screen.BeginFrame()
	screen.SetCamera3D(Camera3D{Enabled: true})
	opts := &DrawTrianglesOptions{DepthTest: true, DepthWrite: true}
	verts := []Vertex3D{
		{X: 0, Y: 0, Z: 0, ColorR: 1, ColorG: 1, ColorB: 1, ColorA: 1},
		{X: 1, Y: 0, Z: 0, SrcX: 1, ColorR: 1, ColorG: 1, ColorB: 1, ColorA: 1},
		{X: 0, Y: 1, Z: 0, SrcY: 1, ColorR: 1, ColorG: 1, ColorB: 1, ColorA: 1},
	}
	for _, img := range []*Image{red, blue} {
		screen.DrawWorldBillboard(WorldBillboardCommand{Texture: img, Width: 4, Height: 4, Options: DrawTrianglesOptions{DepthTest: true}})
		screen.DrawWorldMesh(NewWorldMeshWithLightmap(verts, []uint16{0, 1, 2}, img, nil, opts))
	}
	screen.DrawTriangles3D([]Vertex3D{
		{X: triX, Y: 0, Z: 0, ColorA: 1},
		{X: triX + 1, Y: 0, Z: 0, ColorA: 1},
		{X: triX, Y: 1, Z: 0, ColorA: 1},
	}, []uint16{0, 1, 2}, red, nil)
}

func solidImage(width, height int, c color.RGBA) *Image {
	img := NewImage(width, height)
	img.Fill(c)
	return img
}

func passNames(passes []Pass) []string {
	names := make([]string, len(passes))
	for i, pass := range passes {
		names[i] = pass.Name
	}
	return names
}

func pointLightIn(lights []PointLight, want PointLight) bool {
	for _, light := range lights {
		if light == want {
			return true
		}
	}
	return false
}

func tileHasLight(tiles [][]int) bool {
	for _, lights := range tiles {
		if len(lights) > 0 {
			return true
		}
	}
	return false
}
