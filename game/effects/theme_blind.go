package effects

import (
	"image/color"

	"github.com/kivutar/goro/ui/rotheme"
)

// The screen darkening of the Blind and Darkness (Devil) statuses. The alpha
// of each tint fades in; only its color is used.
var (
	blindRadius  = 10.0
	blindTint    = color.RGBA{R: 10, G: 10, B: 10, A: 255}
	darknessTint = color.RGBA{R: 30, G: 30, B: 30, A: 255}
)

func init() {
	s := rotheme.Section("blind", "screen darkening of the Blind and Darkness statuses")
	s.Float64("radius", &blindRadius, 1, 100, "clear radius around you while blinded (Darkness uses the original per-level radius)")
	s.RGBA("blind_tint", &blindTint, "Blind darkness color")
	s.RGBA("darkness_tint", &darknessTint, "Darkness (Devil) color")
}
