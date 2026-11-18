package game

import (
	"image/color"

	"github.com/kivutar/goro/ui/rotheme"
)

// Labels and the fallback marker of items lying on the ground.
var (
	itemNameLabelForeground = color.RGBA{R: 255, G: 239, B: 148, A: 255}
	itemNameLabelOutline    = color.RGBA{A: 196}
	itemNameLabelOffset     = 12.0
	itemMarkerFill          = color.RGBA{R: 245, G: 220, B: 96, A: 255}
	itemMarkerBorder        = color.RGBA{R: 80, G: 48, B: 16, A: 230}
)

func init() {
	s := rotheme.Section("item_labels", "names of items on the ground, and the marker drawn when an item has no sprite")
	s.RGBA("text", &itemNameLabelForeground, "item name text")
	s.RGBA("outline", &itemNameLabelOutline, "item name outline")
	s.Float64("label_offset", &itemNameLabelOffset, -100, 200, "name distance below the item (times the item scale)")
	s.RGBA("marker_fill", &itemMarkerFill, "fallback marker glow (its alpha fades toward the edge)")
	s.RGBA("marker_border", &itemMarkerBorder, "fallback marker box outline")
}
