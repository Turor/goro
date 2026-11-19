struct ShadowCam {
	row0: vec4<f32>,
	row1: vec4<f32>,
	row2: vec4<f32>,
	row3: vec4<f32>,
	sun_dir: vec4<f32>,
}

struct ShadowAxes {
	right: vec3<f32>,
	up: vec3<f32>,
}

@group(0) @binding(0) var<uniform> cam: ShadowCam;
@group(0) @binding(1) var sprite_sampler: sampler;
@group(0) @binding(2) var sprite_tex: texture_2d<f32>;

struct VertexInput {
	@location(0) local: vec2<f32>,
	@location(1) uv: vec2<f32>,
	@location(2) center: vec3<f32>,
	@location(3) right_axis: vec3<f32>,
	@location(4) up_axis: vec3<f32>,
	@location(6) params: vec4<f32>,
}

struct VertexOutput {
	@builtin(position) clip: vec4<f32>,
	@location(0) uv: vec2<f32>,
}

// light_facing matches render.LightFacingAxes. The stored sun vector is the
// downward ray. A zero ray keeps the camera axes so the quad still has area.
fn light_facing(axis_scale: f32, camera_right: vec3<f32>, camera_up: vec3<f32>) -> ShadowAxes {
	var axes: ShadowAxes;
	let ray = cam.sun_dir.xyz;
	let rlen = length(ray);
	if (rlen < 0.000001 || axis_scale == 0.0) {
		axes.right = camera_right;
		axes.up = camera_up;
		return axes;
	}
	let nray = ray / rlen;
	var world_up = vec3<f32>(0.0, 1.0, 0.0);
	if (abs(dot(nray, world_up)) > 0.95) {
		world_up = vec3<f32>(0.0, 0.0, 1.0);
	}
	let side_raw = cross(world_up, nray);
	if (length(side_raw) < 0.000001) {
		axes.right = camera_right;
		axes.up = camera_up;
		return axes;
	}
	let side = normalize(side_raw);
	let lift = normalize(cross(nray, side));
	// side is -cameraRight and lift is +cameraUp. Texture y=0 is a negative
	// pixel (the head), so negate both to match color billboards.
	axes.right = -side * axis_scale;
	axes.up = -lift * axis_scale;
	return axes;
}

@vertex
fn vs_main(input: VertexInput) -> VertexOutput {
	var out: VertexOutput;
	let width = input.params[0];
	let height = input.params[1];
	let anchor = vec2<f32>(input.params[2], input.params[3]);
	let pixel = vec2<f32>(input.local[0] * width, input.local[1] * height) - anchor;
	let axes = light_facing(length(input.right_axis), input.right_axis, input.up_axis);
	let pos = input.center + axes.right * pixel[0] + axes.up * pixel[1];
	out.clip = cam.row0 * pos.x + cam.row1 * pos.y + cam.row2 * pos.z + cam.row3;
	out.uv = input.uv;
	return out;
}

@fragment
fn fs_main(input: VertexOutput) -> @location(0) vec4<f32> {
	let color = textureSample(sprite_tex, sprite_sampler, input.uv);
	if (color.a < 0.5) {
		discard;
	}
	return vec4<f32>(0.0, 0.0, 0.0, 0.0);
}
