package render

import (
	"encoding/binary"
	"math"
)

// LightingUniformBytes is the std140-style size of LightingUniforms in the
// world and billboard shaders. The shade matches Shade when enhanced is on
// and the lightmap term alone when it is off.
const LightingUniformBytes = 9600

// EncodeLightingUniform packs the sun, the capped point-light set, and each
// tile's light list for the shader. enhanced false clears the sun enable flag
// so the shader keeps the lightmap term (or the unlit sprite). Ambient is
// packed into existing holes: sun_color.w and shadow_params.zw. Light count
// stays at byte 112 and tile counts stay at byte 384.
func EncodeLightingUniform(sun SunLight, points []PointLight, viewProj, shadowVP [16]float32, width, height int, enhanced bool) []byte {
	buf := make([]byte, LightingUniformBytes)
	selected := SelectPointLights(points, MaxPointLights)
	tiles := AssignTileLights(selected, viewProj, LightTileCount, LightTileCount)
	enabled := float32(0)
	if enhanced {
		enabled = 1
	}
	putF32(buf, 0, sun.Direction[0])
	putF32(buf, 4, sun.Direction[1])
	putF32(buf, 8, sun.Direction[2])
	putF32(buf, 12, enabled)
	putF32(buf, 16, sun.Color[0])
	putF32(buf, 20, sun.Color[1])
	putF32(buf, 24, sun.Color[2])
	putF32(buf, 28, sun.Ambient[2])
	screenW, screenH := float32(width), float32(height)
	if screenW < 1 {
		screenW = 1
	}
	if screenH < 1 {
		screenH = 1
	}
	putF32(buf, 32, screenW)
	putF32(buf, 36, screenH)
	putF32(buf, 40, LightTileCount)
	putF32(buf, 44, LightTileCount)
	for i, value := range shadowVP {
		putF32(buf, 48+i*4, value)
	}
	putF32(buf, 112, float32(len(selected)))
	putF32(buf, 116, ShadowBias)
	putF32(buf, 120, sun.Ambient[0])
	putF32(buf, 124, sun.Ambient[1])
	for i, light := range selected {
		base := 128 + i*32
		putF32(buf, base, light.Position[0])
		putF32(buf, base+4, light.Position[1])
		putF32(buf, base+8, light.Position[2])
		putF32(buf, base+12, light.Range)
		putF32(buf, base+16, light.Color[0])
		putF32(buf, base+20, light.Color[1])
		putF32(buf, base+24, light.Color[2])
		putF32(buf, base+28, light.Intensity)
	}
	countsAt := 384
	indicesAt := 1408
	for tile := 0; tile < LightTileCount*LightTileCount; tile++ {
		list := tiles[tile]
		count := len(list)
		if count > MaxLightsPerTile {
			count = MaxLightsPerTile
		}
		lane := (tile % 4) * 4
		putU32(buf, countsAt+(tile/4)*16+lane, uint32(count))
		for slot := 0; slot < count; slot++ {
			indexLane := (slot % 4) * 4
			vec := tile*2 + slot/4
			putU32(buf, indicesAt+vec*16+indexLane, uint32(list[slot]))
		}
	}
	return buf
}

func putF32(buf []byte, offset int, value float32) {
	binary.LittleEndian.PutUint32(buf[offset:offset+4], binaryFloat32(value))
}

func putU32(buf []byte, offset int, value uint32) {
	binary.LittleEndian.PutUint32(buf[offset:offset+4], value)
}

func binaryFloat32(value float32) uint32 {
	return math.Float32bits(value)
}
