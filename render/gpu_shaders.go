package render

const screenShaderWGSL = `
struct Uniforms {
	screen: vec2<f32>,
	padding: vec2<f32>,
}

@group(0) @binding(0) var<uniform> uniforms: Uniforms;
@group(0) @binding(1) var tex_sampler: sampler;
@group(0) @binding(2) var tex: texture_2d<f32>;

struct VertexInput {
	@location(0) pos: vec2<f32>,
	@location(1) uv: vec2<f32>,
	@location(2) color: vec4<f32>,
}

struct VertexOutput {
	@builtin(position) clip: vec4<f32>,
	@location(0) uv: vec2<f32>,
	@location(1) color: vec4<f32>,
}

@vertex
fn vs_main(input: VertexInput) -> VertexOutput {
	var out: VertexOutput;
	let x = (input.pos.x / uniforms.screen.x) * 2.0 - 1.0;
	let y = 1.0 - (input.pos.y / uniforms.screen.y) * 2.0;
	out.clip = vec4<f32>(x, y, 0.0, 1.0);
	out.uv = input.uv;
	out.color = input.color;
	return out;
}

@fragment
fn fs_main(input: VertexOutput) -> @location(0) vec4<f32> {
	let color = textureSample(tex, tex_sampler, input.uv) * input.color;
	if (color.a < 0.01) {
		discard;
	}
	return color;
}
`

const worldFogShaderWGSL = `
fn fog_visibility(depth: f32, near: f32, far: f32) -> f32 {
	let visibility = clamp((far - depth) / max(far - near, 0.000001), 0.0, 1.0);
	return floor(visibility * 255.0) / 255.0;
}
`

// lightingShaderWGSL is the enhanced-light uniform shared by world and sprite
// shaders. sun_dir.w is 0 when enhanced lighting is off, and the shade then
// matches the lightmap-only (or unlit sprite) path. When it is on, fragments
// multiply by saturate(ambient + sun·N·visibility + points) and ignore the
// baked lightmap. Sprites pass receive_shadow false so they are not darkened
// by the billboard depth they cast. Ambient is shadow_params.zw and sun_color.w.
const lightingShaderWGSL = `
struct PointLightGPU {
	pos_range: vec4<f32>,
	color_intensity: vec4<f32>,
}

struct LightingUniforms {
	sun_dir: vec4<f32>,
	sun_color: vec4<f32>,
	screen_tile: vec4<f32>,
	shadow_row0: vec4<f32>,
	shadow_row1: vec4<f32>,
	shadow_row2: vec4<f32>,
	shadow_row3: vec4<f32>,
	shadow_params: vec4<f32>,
	lights: array<PointLightGPU, 8>,
	tile_counts: array<vec4<u32>, 64>,
	tile_indices: array<vec4<u32>, 512>,
}

@group(0) @binding(5) var<uniform> lighting: LightingUniforms;
@group(0) @binding(6) var shadow_map: texture_depth_2d;

fn light_at(index: u32) -> PointLightGPU {
	switch index {
		case 0u: { return lighting.lights[0]; }
		case 1u: { return lighting.lights[1]; }
		case 2u: { return lighting.lights[2]; }
		case 3u: { return lighting.lights[3]; }
		case 4u: { return lighting.lights[4]; }
		case 5u: { return lighting.lights[5]; }
		case 6u: { return lighting.lights[6]; }
		default: { return lighting.lights[7]; }
	}
}

fn tile_count_at(tile: u32) -> u32 {
	let packed = lighting.tile_counts[tile / 4u];
	let lane = tile % 4u;
	if (lane == 0u) { return packed.x; }
	if (lane == 1u) { return packed.y; }
	if (lane == 2u) { return packed.z; }
	return packed.w;
}

fn tile_index_at(tile: u32, slot: u32) -> u32 {
	let packed = lighting.tile_indices[tile * 2u + slot / 4u];
	let lane = slot % 4u;
	if (lane == 0u) { return packed.x; }
	if (lane == 1u) { return packed.y; }
	if (lane == 2u) { return packed.z; }
	return packed.w;
}

fn point_term(light: PointLightGPU, world_pos: vec3<f32>, normal: vec3<f32>) -> vec3<f32> {
	let range = light.pos_range.w;
	let intensity = light.color_intensity.w;
	if (range <= 0.0 || intensity <= 0.0) {
		return vec3<f32>(0.0);
	}
	let to_light = light.pos_range.xyz - world_pos;
	let dist = length(to_light);
	if (dist >= range) {
		return vec3<f32>(0.0);
	}
	let atten = 1.0 - dist / range;
	let nlen = length(normal);
	var ndotl = 0.0;
	if (nlen > 0.000001 && dist > 0.000001) {
		ndotl = max(dot(normal / nlen, to_light / dist), 0.0);
	}
	return light.color_intensity.xyz * intensity * atten * atten * ndotl;
}

fn tile_point_term(world_pos: vec3<f32>, pixel: vec2<f32>, normal: vec3<f32>) -> vec3<f32> {
	let tiles = max(lighting.screen_tile.zw, vec2<f32>(1.0));
	let screen = max(lighting.screen_tile.xy, vec2<f32>(1.0));
	let tx = min(u32(pixel.x / screen.x * tiles.x), u32(tiles.x) - 1u);
	let ty = min(u32(pixel.y / screen.y * tiles.y), u32(tiles.y) - 1u);
	let tile = ty * u32(tiles.x) + tx;
	let count = min(tile_count_at(tile), 8u);
	var sum = vec3<f32>(0.0);
	for (var slot = 0u; slot < 8u; slot++) {
		if (slot >= count) {
			break;
		}
		let index = tile_index_at(tile, slot);
		if (index < 8u) {
			sum = sum + point_term(light_at(index), world_pos, normal);
		}
	}
	return sum;
}

fn shadow_occluded(world_pos: vec3<f32>) -> bool {
	let lp = lighting.shadow_row0 * world_pos.x + lighting.shadow_row1 * world_pos.y + lighting.shadow_row2 * world_pos.z + lighting.shadow_row3;
	let size = vec2<i32>(textureDimensions(shadow_map));
	let sx = i32((lp.x / max(lp.w, 0.000001) * 0.5 + 0.5) * f32(size.x));
	let sy = i32((lp.y / max(lp.w, 0.000001) * 0.5 + 0.5) * f32(size.y));
	if (sx < 0 || sy < 0 || sx >= size.x || sy >= size.y) {
		return false;
	}
	let map_depth = textureLoad(shadow_map, vec2<i32>(sx, sy), 0);
	let receiver = clamp(lp.z / max(lp.w, 0.000001) * 0.5 + 0.5, 0.0, 1.0);
	return receiver > map_depth + lighting.shadow_params.y;
}

fn sun_ndotl(normal: vec3<f32>) -> f32 {
	let nlen = length(normal);
	let ray = lighting.sun_dir.xyz;
	let rlen = length(ray);
	if (nlen < 0.000001 || rlen < 0.000001) {
		return 0.0;
	}
	// sun_dir is the downward ray. Face the sun.
	return max(dot(normal / nlen, -ray / rlen), 0.0);
}

fn enhanced_light(base: vec3<f32>, normal: vec3<f32>, world_pos: vec3<f32>, pixel: vec2<f32>, receive_shadow: bool) -> vec3<f32> {
	let ambient = vec3<f32>(lighting.shadow_params.z, lighting.shadow_params.w, lighting.sun_color.w);
	var visibility = 1.0;
	if (receive_shadow) {
		if (shadow_occluded(world_pos)) {
			visibility = 0.0;
		}
	}
	let sum = ambient + lighting.sun_color.rgb * sun_ndotl(normal) * visibility + tile_point_term(world_pos, pixel, normal);
	return base * clamp(sum, vec3<f32>(0.0), vec3<f32>(1.0));
}
`

const modelShaderWGSL = `
struct ModelUniforms {
	model0: vec4<f32>,
	model1: vec4<f32>,
	model2: vec4<f32>,
	model3: vec4<f32>,
	uv_offset: vec4<f32>,
}

@group(1) @binding(0) var<uniform> model_uniforms: ModelUniforms;

fn model_point(pos: vec3<f32>) -> vec3<f32> {
	let p = model_uniforms.model0 * pos.x + model_uniforms.model1 * pos.y + model_uniforms.model2 * pos.z + model_uniforms.model3;
	return p.xyz;
}
`

const worldShaderWGSL = worldFogShaderWGSL + modelShaderWGSL + `
struct Uniforms {
	c0: vec4<f32>,
	c1: vec4<f32>,
	c2: vec4<f32>,
	c3: vec4<f32>,
	fog: vec4<f32>,
	fog_color: vec4<f32>,
}

@group(0) @binding(0) var<uniform> uniforms: Uniforms;
@group(0) @binding(1) var tex_sampler: sampler;
@group(0) @binding(2) var tex: texture_2d<f32>;
@group(0) @binding(3) var light_tex: texture_2d<f32>;
@group(0) @binding(4) var light_sampler: sampler;
` + lightingShaderWGSL + `
struct VertexInput {
	@location(0) pos: vec3<f32>,
	@location(1) uv: vec2<f32>,
	@location(2) color: vec4<f32>,
	@location(3) depth_pos: vec3<f32>,
	@location(4) depth_bias: f32,
	@location(5) fog_enabled: f32,
	@location(6) light_uv: vec2<f32>,
}

struct VertexOutput {
	@builtin(position) clip: vec4<f32>,
	@location(0) uv: vec2<f32>,
	@location(1) color: vec4<f32>,
	@location(2) fog_enabled: f32,
	@location(3) light_uv: vec2<f32>,
	@location(4) fog_visibility: vec2<f32>,
	@location(5) world_pos: vec3<f32>,
}

@vertex
fn vs_main(input: VertexInput) -> VertexOutput {
	var out: VertexOutput;
	let world_pos = model_point(input.pos);
	let depth_world = model_point(input.depth_pos);
	let clip = uniforms.c0 * world_pos[0] + uniforms.c1 * world_pos[1] + uniforms.c2 * world_pos[2] + uniforms.c3;
	let depth_clip = uniforms.c0 * depth_world[0] + uniforms.c1 * depth_world[1] + uniforms.c2 * depth_world[2] + uniforms.c3;
	let clip_z = clip[2] * 0.5 + clip[3] * 0.5;
	let depth_z = depth_clip[2] * 0.5 + depth_clip[3] * 0.5;
	let occlusion_z = min(clip_z, depth_z * clip[3] / max(depth_clip[3], 0.000001));
	let final_z = max(0.0, occlusion_z - input.depth_bias * clip[3]);
	out.clip = vec4<f32>(clip[0], clip[1], final_z, clip[3]);
	out.uv = input.uv + model_uniforms.uv_offset.xy;
	out.color = input.color;
	out.fog_enabled = input.fog_enabled;
	out.light_uv = input.light_uv;
	// Dividing the interpolated (visibility * w, w) pair gives screen-linear fog.
	// Avoids GLES noperspective and backend differences in fragment position.w.
	out.fog_visibility = vec2<f32>(fog_visibility(clip[3], uniforms.fog[0], uniforms.fog[1]), 1.0) * clip[3];
	out.world_pos = world_pos;
	return out;
}

@fragment
fn fs_main(input: VertexOutput) -> @location(0) vec4<f32> {
	var color = textureSample(tex, tex_sampler, input.uv) * input.color;
	var lit: vec3<f32>;
	if (lighting.sun_dir.w != 0.0) {
		let n = cross(dpdx(input.world_pos), dpdy(input.world_pos));
		lit = enhanced_light(color.rgb, n, input.world_pos, input.clip.xy, true);
	} else {
		let lightmap = textureSample(light_tex, light_sampler, input.light_uv);
		lit = color.rgb * lightmap.a + clamp(lightmap.rgb, vec3<f32>(0.0), vec3<f32>(1.0));
	}
	color = vec4<f32>(lit, color.a);
	if (color[3] < 0.01) {
		discard;
	}
	let fog = clamp((1.0 - input.fog_visibility.x / input.fog_visibility.y) * uniforms.fog[2] * uniforms.fog[3] * input.fog_enabled, 0.0, 1.0);
	return vec4<f32>(
		color[0] * (1.0 - fog) + uniforms.fog_color[0] * fog,
		color[1] * (1.0 - fog) + uniforms.fog_color[1] * fog,
		color[2] * (1.0 - fog) + uniforms.fog_color[2] * fog,
		color[3],
	);
}
`

const worldBillboardShaderWGSL = worldFogShaderWGSL + `
struct Uniforms {
	c0: vec4<f32>,
	c1: vec4<f32>,
	c2: vec4<f32>,
	c3: vec4<f32>,
	fog: vec4<f32>,
	fog_color: vec4<f32>,
}

@group(0) @binding(0) var<uniform> uniforms: Uniforms;
@group(0) @binding(1) var tex_sampler: sampler;
@group(0) @binding(2) var tex: texture_2d<f32>;
@group(0) @binding(3) var light_tex: texture_2d<f32>;
` + lightingShaderWGSL + `
struct VertexInput {
	@location(0) local: vec2<f32>,
	@location(1) uv: vec2<f32>,
	@location(2) center: vec3<f32>,
	@location(3) right_axis: vec3<f32>,
	@location(4) up_axis: vec3<f32>,
	@location(5) depth_up_axis: vec3<f32>,
	@location(6) params: vec4<f32>,
	@location(7) color: vec4<f32>,
	@location(8) depth_bias: f32,
	@location(9) fog_enabled: f32,
}

struct VertexOutput {
	@builtin(position) clip: vec4<f32>,
	@location(0) uv: vec2<f32>,
	@location(1) color: vec4<f32>,
	@location(2) fog_visibility: vec2<f32>,
	@location(3) fog_enabled: f32,
	@location(4) world_pos: vec3<f32>,
	@location(5) normal: vec3<f32>,
}

fn billboard_vertex(input: VertexInput, depth_only: bool) -> VertexOutput {
	var out: VertexOutput;
	let width = input.params[0];
	let height = input.params[1];
	let anchor = vec2<f32>(input.params[2], input.params[3]);
	let pixel = vec2<f32>(input.local[0] * width, input.local[1] * height) - anchor;
	let pos = input.center + input.right_axis * pixel[0] + input.up_axis * pixel[1];
	let depth_pos = input.center + input.right_axis * pixel[0] + input.depth_up_axis * pixel[1];
	let clip = uniforms.c0 * pos[0] + uniforms.c1 * pos[1] + uniforms.c2 * pos[2] + uniforms.c3;
	let depth_clip = uniforms.c0 * depth_pos[0] + uniforms.c1 * depth_pos[1] + uniforms.c2 * depth_pos[2] + uniforms.c3;
	let clip_z = clip[2] * 0.5 + clip[3] * 0.5;
	let depth_z = depth_clip[2] * 0.5 + depth_clip[3] * 0.5;
	let occlusion_z = min(clip_z, depth_z * clip[3] / max(depth_clip[3], 0.000001));
	var final_z = max(0.0, occlusion_z - input.depth_bias * clip[3]);
	if (depth_only) {
		let center = uniforms.c0 * input.center[0] + uniforms.c1 * input.center[1] + uniforms.c2 * input.center[2] + uniforms.c3;
		final_z = max(0.0, center[2] / max(center[3], 0.000001) * 0.5 + 0.5 - input.depth_bias) * clip[3];
	}
	out.clip = vec4<f32>(clip[0], clip[1], final_z, clip[3]);
	out.uv = input.uv;
	out.color = input.color;
	out.fog_visibility = vec2<f32>(fog_visibility(clip[3], uniforms.fog[0], uniforms.fog[1]), 1.0) * clip[3];
	out.fog_enabled = input.fog_enabled;
	out.world_pos = pos;
	out.normal = cross(input.right_axis, input.up_axis);
	return out;
}

@vertex
fn vs_main(input: VertexInput) -> VertexOutput {
	return billboard_vertex(input, false);
}

@vertex
fn vs_depth_main(input: VertexInput) -> VertexOutput {
	return billboard_vertex(input, true);
}

@fragment
fn fs_main(input: VertexOutput) -> @location(0) vec4<f32> {
	var color = textureSample(tex, tex_sampler, input.uv) * input.color;
	if (lighting.sun_dir.w != 0.0) {
		color = vec4<f32>(enhanced_light(color.rgb, input.normal, input.world_pos, input.clip.xy, false), color.a);
	}
	if (color[3] < 0.01) {
		discard;
	}
	let fog = clamp((1.0 - input.fog_visibility.x / input.fog_visibility.y) * uniforms.fog[2] * uniforms.fog[3] * input.fog_enabled, 0.0, 1.0);
	return vec4<f32>(
		color[0] * (1.0 - fog) + uniforms.fog_color[0] * fog,
		color[1] * (1.0 - fog) + uniforms.fog_color[1] * fog,
		color[2] * (1.0 - fog) + uniforms.fog_color[2] * fog,
		color[3],
	);
}
`

// worldBillboardArrayShaderWGSL samples a texture array so one instanced draw
// can cover billboards that differ only by texture.
const worldBillboardArrayShaderWGSL = worldFogShaderWGSL + `
struct Uniforms {
	c0: vec4<f32>,
	c1: vec4<f32>,
	c2: vec4<f32>,
	c3: vec4<f32>,
	fog: vec4<f32>,
	fog_color: vec4<f32>,
}

@group(0) @binding(0) var<uniform> uniforms: Uniforms;
@group(0) @binding(1) var tex_sampler: sampler;
@group(0) @binding(2) var tex: texture_2d_array<f32>;
@group(0) @binding(3) var light_tex: texture_2d<f32>;
` + lightingShaderWGSL + `
struct VertexInput {
	@location(0) local: vec2<f32>,
	@location(1) uv: vec2<f32>,
	@location(2) center: vec3<f32>,
	@location(3) right_axis: vec3<f32>,
	@location(4) up_axis: vec3<f32>,
	@location(5) depth_up_axis: vec3<f32>,
	@location(6) params: vec4<f32>,
	@location(7) color: vec4<f32>,
	@location(8) depth_bias: f32,
	@location(9) fog_enabled: f32,
	@location(10) layer_uv: vec4<f32>,
}

struct VertexOutput {
	@builtin(position) clip: vec4<f32>,
	@location(0) uv: vec2<f32>,
	@location(1) color: vec4<f32>,
	@location(2) fog_visibility: vec2<f32>,
	@location(3) fog_enabled: f32,
	@location(4) world_pos: vec3<f32>,
	@location(5) normal: vec3<f32>,
	@location(6) layer_uv: vec4<f32>,
}

@vertex
fn vs_main(input: VertexInput) -> VertexOutput {
	var out: VertexOutput;
	let width = input.params[0];
	let height = input.params[1];
	let anchor = vec2<f32>(input.params[2], input.params[3]);
	let pixel = vec2<f32>(input.local[0] * width, input.local[1] * height) - anchor;
	let pos = input.center + input.right_axis * pixel[0] + input.up_axis * pixel[1];
	let depth_pos = input.center + input.right_axis * pixel[0] + input.depth_up_axis * pixel[1];
	let clip = uniforms.c0 * pos[0] + uniforms.c1 * pos[1] + uniforms.c2 * pos[2] + uniforms.c3;
	let depth_clip = uniforms.c0 * depth_pos[0] + uniforms.c1 * depth_pos[1] + uniforms.c2 * depth_pos[2] + uniforms.c3;
	let clip_z = clip[2] * 0.5 + clip[3] * 0.5;
	let depth_z = depth_clip[2] * 0.5 + depth_clip[3] * 0.5;
	let occlusion_z = min(clip_z, depth_z * clip[3] / max(depth_clip[3], 0.000001));
	let final_z = max(0.0, occlusion_z - input.depth_bias * clip[3]);
	out.clip = vec4<f32>(clip[0], clip[1], final_z, clip[3]);
	out.uv = input.uv;
	out.color = input.color;
	out.fog_visibility = vec2<f32>(fog_visibility(clip[3], uniforms.fog[0], uniforms.fog[1]), 1.0) * clip[3];
	out.fog_enabled = input.fog_enabled;
	out.world_pos = pos;
	out.normal = cross(input.right_axis, input.up_axis);
	out.layer_uv = input.layer_uv;
	return out;
}

@fragment
fn fs_main(input: VertexOutput) -> @location(0) vec4<f32> {
	let uv = input.uv * max(input.layer_uv.yz, vec2<f32>(0.000001));
	var color = textureSample(tex, tex_sampler, uv, i32(input.layer_uv.x)) * input.color;
	if (lighting.sun_dir.w != 0.0) {
		color = vec4<f32>(enhanced_light(color.rgb, input.normal, input.world_pos, input.clip.xy, false), color.a);
	}
	if (color[3] < 0.01) {
		discard;
	}
	let fog = clamp((1.0 - input.fog_visibility.x / input.fog_visibility.y) * uniforms.fog[2] * uniforms.fog[3] * input.fog_enabled, 0.0, 1.0);
	return vec4<f32>(
		color[0] * (1.0 - fog) + uniforms.fog_color[0] * fog,
		color[1] * (1.0 - fog) + uniforms.fog_color[1] * fog,
		color[2] * (1.0 - fog) + uniforms.fog_color[2] * fog,
		color[3],
	);
}
`
