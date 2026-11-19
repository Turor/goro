package render

import "math"

const (
	// MaxPointLights is the strongest point-light set enhanced shading keeps.
	MaxPointLights = 8
	// LightTileCount is the screen-space tile grid along each axis.
	LightTileCount = 16
	// MaxLightsPerTile is how many lights one tile's shade loop may see.
	MaxLightsPerTile = 8
	// ShadowBias is added to the shadow-map depth before the occlusion test.
	ShadowBias float32 = 0.002
	// ShadowDarken scales the sun term on receivers the shadow map occludes.
	ShadowDarken float32 = 0.45
)

// PointLight is one map point light in world space.
type PointLight struct {
	Position  [3]float32
	Color     [3]float32
	Range     float32
	Intensity float32
}

// SunLight is the map sun.
type SunLight struct {
	Direction [3]float32
	Color     [3]float32
}

// SelectPointLights keeps at most cap of the strongest lights.
// Strength is sqrt(range) * intensity, matching a bounded enhanced-light set.
func SelectPointLights(lights []PointLight, cap int) []PointLight {
	if cap <= 0 || len(lights) == 0 {
		return nil
	}
	type scored struct {
		light PointLight
		score float32
		index int
	}
	ranked := make([]scored, 0, len(lights))
	for i, light := range lights {
		if light.Range <= 0 || light.Intensity <= 0 {
			continue
		}
		ranked = append(ranked, scored{
			light: light,
			score: float32(math.Sqrt(float64(light.Range))) * light.Intensity,
			index: i,
		})
	}
	for i := 1; i < len(ranked); i++ {
		item := ranked[i]
		j := i
		for j > 0 && (ranked[j-1].score < item.score || (ranked[j-1].score == item.score && ranked[j-1].index > item.index)) {
			ranked[j] = ranked[j-1]
			j--
		}
		ranked[j] = item
	}
	if len(ranked) > cap {
		ranked = ranked[:cap]
	}
	out := make([]PointLight, len(ranked))
	for i, item := range ranked {
		out[i] = item.light
	}
	return out
}

// ShaderFramebufferTile is the tile tile_point_term reads for a framebuffer
// pixel. The origin is the top-left, matching pixel.y / screen.y in the
// shader: pixel (0, 0) is tile (0, 0). NDC y +1 is that top row, and NDC y -1
// is the last row.
func ShaderFramebufferTile(pixelX, pixelY, screenW, screenH float32, tilesX, tilesY int) (int, int) {
	return shaderTileAxis(pixelX, screenW, tilesX), shaderTileAxis(pixelY, screenH, tilesY)
}

func shaderTileAxis(pixel, screen float32, tiles int) int {
	if tiles < 1 {
		tiles = 1
	}
	if screen < 1 {
		screen = 1
	}
	idx := int(pixel / screen * float32(tiles))
	if idx < 0 {
		idx = 0
	}
	if idx >= tiles {
		idx = tiles - 1
	}
	return idx
}

// AssignTileLights bins lights into a tilesX by tilesY grid. A light is absent
// from a tile its screen-space sphere misses. The returned lists are light
// indices into lights. Rows match ShaderFramebufferTile, so row 0 is the top
// of the framebuffer.
func AssignTileLights(lights []PointLight, viewProj [16]float32, tilesX, tilesY int) [][]int {
	if tilesX < 1 {
		tilesX = LightTileCount
	}
	if tilesY < 1 {
		tilesY = LightTileCount
	}
	tiles := make([][]int, tilesX*tilesY)
	for lightIndex, light := range lights {
		ndcX, ndcY, clipW, ok := projectClip(viewProj, light.Position)
		if !ok {
			continue
		}
		radius := float32(0)
		if clipW > 0.0001 {
			radius = light.Range / clipW
		}
		minX, maxX, okX := ndcSpanTiles(ndcX, radius, tilesX, false)
		minY, maxY, okY := ndcSpanTiles(ndcY, radius, tilesY, true)
		if !okX || !okY {
			continue
		}
		for y := minY; y <= maxY; y++ {
			for x := minX; x <= maxX; x++ {
				slot := y*tilesX + x
				if len(tiles[slot]) >= MaxLightsPerTile {
					continue
				}
				tiles[slot] = append(tiles[slot], lightIndex)
			}
		}
	}
	return tiles
}

// ndcSpanTiles maps a light's NDC extent onto shader tile indexes. yFromTop
// uses the framebuffer convention: NDC +1 is tile 0 and NDC -1 is the last tile.
func ndcSpanTiles(ndc, radius float32, tiles int, yFromTop bool) (int, int, bool) {
	lo := ndc - radius
	hi := ndc + radius
	if hi < -1 || lo > 1 {
		return 0, 0, false
	}
	screen := float32(tiles)
	var p0, p1 float32
	if yFromTop {
		p0 = (1 - hi) * 0.5 * screen
		p1 = (1 - lo) * 0.5 * screen
	} else {
		p0 = (lo + 1) * 0.5 * screen
		p1 = (hi + 1) * 0.5 * screen
	}
	a := shaderTileAxis(p0, screen, tiles)
	b := shaderTileAxis(p1, screen, tiles)
	if a > b {
		a, b = b, a
	}
	return a, b, true
}

func projectClip(viewProj [16]float32, p [3]float32) (ndcX, ndcY, clipW float32, ok bool) {
	clipX := viewProj[0]*p[0] + viewProj[4]*p[1] + viewProj[8]*p[2] + viewProj[12]
	clipY := viewProj[1]*p[0] + viewProj[5]*p[1] + viewProj[9]*p[2] + viewProj[13]
	clipW = viewProj[3]*p[0] + viewProj[7]*p[1] + viewProj[11]*p[2] + viewProj[15]
	if clipW <= 0.0001 || math.IsNaN(float64(clipW)) || math.IsInf(float64(clipW), 0) {
		return 0, 0, clipW, false
	}
	return clipX / clipW, clipY / clipW, clipW, true
}

// ShadowOccludes reports whether a receiver is past the shadow map depth.
func ShadowOccludes(receiverLightDepth, mapDepth, bias float32) bool {
	if bias <= 0 || math.IsNaN(float64(bias)) {
		bias = ShadowBias
	}
	return receiverLightDepth > mapDepth+bias
}

// Shade applies the lightmap and, when enhanced is on, the sun N·L term, the
// point lights, and shadow darkening. Enhanced off is the lightmap term only.
func Shade(base [3]float32, lightmap [4]float32, normal [3]float32, sun SunLight, points []PointLight, worldPos [3]float32, occluded, enhanced bool) [3]float32 {
	lit := lightmapTerm(base, lightmap)
	if !enhanced {
		return lit
	}
	n := normalize3(normal)
	sunDir := normalize3(sun.Direction)
	ndotl := dot3(n, sunDir)
	if ndotl < 0 {
		ndotl = 0
	}
	sunScale := ndotl
	if occluded {
		sunScale *= ShadowDarken
	}
	lit = add3(lit, scale3(sun.Color, sunScale))
	for _, light := range points {
		lit = add3(lit, pointContribution(light, worldPos))
	}
	return lit
}

func lightmapTerm(base [3]float32, lightmap [4]float32) [3]float32 {
	return [3]float32{
		base[0]*lightmap[3] + clamp01(lightmap[0]),
		base[1]*lightmap[3] + clamp01(lightmap[1]),
		base[2]*lightmap[3] + clamp01(lightmap[2]),
	}
}

func pointContribution(light PointLight, worldPos [3]float32) [3]float32 {
	if light.Range <= 0 || light.Intensity <= 0 {
		return [3]float32{}
	}
	dx := worldPos[0] - light.Position[0]
	dy := worldPos[1] - light.Position[1]
	dz := worldPos[2] - light.Position[2]
	dist := float32(math.Sqrt(float64(dx*dx + dy*dy + dz*dz)))
	if dist >= light.Range {
		return [3]float32{}
	}
	atten := 1 - dist/light.Range
	atten *= atten
	return scale3(light.Color, light.Intensity*atten)
}

func clamp01(v float32) float32 {
	if v < 0 || math.IsNaN(float64(v)) {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

func dot3(a, b [3]float32) float32 {
	return a[0]*b[0] + a[1]*b[1] + a[2]*b[2]
}

func scale3(v [3]float32, s float32) [3]float32 {
	return [3]float32{v[0] * s, v[1] * s, v[2] * s}
}

func add3(a, b [3]float32) [3]float32 {
	return [3]float32{a[0] + b[0], a[1] + b[1], a[2] + b[2]}
}

func normalize3(v [3]float32) [3]float32 {
	length := float32(math.Sqrt(float64(dot3(v, v))))
	if length <= 0.000001 || math.IsNaN(float64(length)) {
		return [3]float32{}
	}
	return scale3(v, 1/length)
}

// IdentityViewProjection projects world coordinates that are already clip coordinates.
func IdentityViewProjection() [16]float32 {
	return [16]float32{
		1, 0, 0, 0,
		0, 1, 0, 0,
		0, 0, 1, 0,
		0, 0, 0, 1,
	}
}

// DirectionalShadowMatrix is an orthographic light matrix centered on the
// receiver. Direction is the sun ray direction used by Shade.
func DirectionalShadowMatrix(center, direction [3]float32, extent float32) [16]float32 {
	if extent <= 1 {
		extent = 64
	}
	dir := normalize3(direction)
	if dir == ([3]float32{}) {
		dir = normalize3([3]float32{-0.5, -0.7, -0.5})
	}
	eye := add3(center, scale3(dir, -extent))
	up := [3]float32{0, 1, 0}
	if math.Abs(float64(dot3(dir, up))) > 0.95 {
		up = [3]float32{0, 0, 1}
	}
	view := lookAt(eye, center, up)
	projection := ortho(extent, 0.1, extent*2)
	return mulMat4(projection, view)
}

func lookAt(eye, target, up [3]float32) [16]float32 {
	forward := normalize3(sub3(target, eye))
	side := normalize3(cross3(forward, up))
	if side == ([3]float32{}) {
		side = [3]float32{1, 0, 0}
	}
	newUp := cross3(side, forward)
	return [16]float32{
		side[0], newUp[0], -forward[0], 0,
		side[1], newUp[1], -forward[1], 0,
		side[2], newUp[2], -forward[2], 0,
		-dot3(side, eye), -dot3(newUp, eye), dot3(forward, eye), 1,
	}
}

func ortho(extent, near, far float32) [16]float32 {
	if extent <= 0.001 {
		extent = 1
	}
	span := far - near
	if span <= 0.001 {
		span = 1
	}
	return [16]float32{
		1 / extent, 0, 0, 0,
		0, 1 / extent, 0, 0,
		0, 0, 1 / span, 0,
		0, 0, -near / span, 1,
	}
}

func mulMat4(a, b [16]float32) [16]float32 {
	var out [16]float32
	for col := 0; col < 4; col++ {
		for row := 0; row < 4; row++ {
			out[col*4+row] = a[row]*b[col*4] + a[4+row]*b[col*4+1] + a[8+row]*b[col*4+2] + a[12+row]*b[col*4+3]
		}
	}
	return out
}

func sub3(a, b [3]float32) [3]float32 {
	return [3]float32{a[0] - b[0], a[1] - b[1], a[2] - b[2]}
}

func cross3(a, b [3]float32) [3]float32 {
	return [3]float32{
		a[1]*b[2] - a[2]*b[1],
		a[2]*b[0] - a[0]*b[2],
		a[0]*b[1] - a[1]*b[0],
	}
}
