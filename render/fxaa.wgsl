@group(0) @binding(0) var scene_tex: texture_2d<f32>;
@group(0) @binding(1) var scene_sampler: sampler;

struct ScreenVertex {
    @builtin(position) position: vec4<f32>,
    @location(0) uv: vec2<f32>,
}

@vertex
fn fullscreen(@builtin(vertex_index) index: u32) -> ScreenVertex {
    let uv = vec2<f32>(f32((index << 1u) & 2u), f32(index & 2u));
    var out: ScreenVertex;
    out.position = vec4<f32>(uv.x * 2.0 - 1.0, 1.0 - uv.y * 2.0, 0.0, 1.0);
    out.uv = uv;
    return out;
}

fn luma(color: vec3<f32>) -> f32 {
    return dot(color, vec3<f32>(0.299, 0.587, 0.114));
}

@fragment
fn fs_main(input: ScreenVertex) -> @location(0) vec4<f32> {
    let size = vec2<f32>(textureDimensions(scene_tex));
    let texel = 1.0 / max(size, vec2<f32>(1.0));
    let center = textureSample(scene_tex, scene_sampler, input.uv);
    let north = textureSample(scene_tex, scene_sampler, input.uv + vec2<f32>(0.0, -texel.y)).rgb;
    let south = textureSample(scene_tex, scene_sampler, input.uv + vec2<f32>(0.0, texel.y)).rgb;
    let west = textureSample(scene_tex, scene_sampler, input.uv + vec2<f32>(-texel.x, 0.0)).rgb;
    let east = textureSample(scene_tex, scene_sampler, input.uv + vec2<f32>(texel.x, 0.0)).rgb;
    let luma_c = luma(center.rgb);
    let luma_n = luma(north);
    let luma_s = luma(south);
    let luma_w = luma(west);
    let luma_e = luma(east);
    let luma_min = min(luma_c, min(min(luma_n, luma_s), min(luma_w, luma_e)));
    let luma_max = max(luma_c, max(max(luma_n, luma_s), max(luma_w, luma_e)));
    let range = luma_max - luma_min;
    if (range < max(0.0312, luma_max * 0.125)) {
        return center;
    }
    let vertical = abs(luma_n + luma_s - 2.0 * luma_c);
    let horizontal = abs(luma_w + luma_e - 2.0 * luma_c);
    var offset = vec2<f32>(texel.x, 0.0);
    if (vertical > horizontal) {
        offset = vec2<f32>(0.0, texel.y);
    }
    let rgb_a = 0.5 * (textureSample(scene_tex, scene_sampler, input.uv + offset).rgb + textureSample(scene_tex, scene_sampler, input.uv - offset).rgb);
    let blend = clamp(range / max(luma_max, 0.000001), 0.0, 0.75);
    return vec4<f32>(mix(center.rgb, rgb_a, blend * 0.5), center.a);
}
