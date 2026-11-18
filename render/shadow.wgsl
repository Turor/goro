struct ShadowCam {
	row0: vec4<f32>,
	row1: vec4<f32>,
	row2: vec4<f32>,
	row3: vec4<f32>,
}

struct ModelUniforms {
	model0: vec4<f32>,
	model1: vec4<f32>,
	model2: vec4<f32>,
	model3: vec4<f32>,
	uv_offset: vec4<f32>,
}

@group(0) @binding(0) var<uniform> cam: ShadowCam;
@group(1) @binding(0) var<uniform> model_uniforms: ModelUniforms;

struct VertexInput {
	@location(0) pos: vec3<f32>,
}

fn model_point(pos: vec3<f32>) -> vec3<f32> {
	let p = model_uniforms.model0 * pos.x + model_uniforms.model1 * pos.y + model_uniforms.model2 * pos.z + model_uniforms.model3;
	return p.xyz;
}

@vertex
fn vs_main(input: VertexInput) -> @builtin(position) vec4<f32> {
	let world_pos = model_point(input.pos);
	return cam.row0 * world_pos.x + cam.row1 * world_pos.y + cam.row2 * world_pos.z + cam.row3;
}
