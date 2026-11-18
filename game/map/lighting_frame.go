package maps

import (
	"github.com/kivutar/goro/game/scene"
	"github.com/kivutar/goro/render"
	"github.com/kivutar/goro/res"
)

// SubmitMapLighting publishes the map sun, the capped point-light set, and the
// directional shadow matrix for this frame. enhanced false keeps lightmap-only shading.
func (m *System) SubmitMapLighting(screen *render.Frame, gnd *res.GND, rsw *res.RSW, projection scene.Projection, enhanced bool) {
	if m == nil || screen == nil {
		return
	}
	lighting := m.Lighting(rsw)
	dir := lighting.Direction()
	sun := render.SunLight{
		Direction: [3]float32{float32(dir.X), float32(dir.Y), float32(dir.Z)},
		Color:     [3]float32{float32(lighting.Diffuse.X), float32(lighting.Diffuse.Y), float32(lighting.Diffuse.Z)},
	}
	var raw []render.PointLight
	if rsw != nil && gnd != nil {
		raw = make([]render.PointLight, 0, len(rsw.Lights))
		for _, light := range rsw.Lights {
			raw = append(raw, render.PointLight{
				Position: [3]float32{
					light.Position.X + float32(gnd.Width),
					light.Position.Y,
					light.Position.Z + float32(gnd.Height),
				},
				Color:     [3]float32{float32(light.Color.R) / 255, float32(light.Color.G) / 255, float32(light.Color.B) / 255},
				Range:     light.Range,
				Intensity: 10,
			})
		}
	}
	points := render.SelectPointLights(raw, render.MaxPointLights)
	var viewProj [16]float32
	for i, value := range projection.ViewProjection {
		viewProj[i] = float32(value)
	}
	center := [3]float32{float32(projection.PlayerX), float32(projection.PlayerZ), float32(projection.PlayerY)}
	shadow := render.DirectionalShadowMatrix(center, sun.Direction, 64)
	screen.SetMapLighting(sun, points, viewProj, shadow, enhanced)
}
