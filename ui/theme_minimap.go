package ui

import (
	"image/color"

	"github.com/kivutar/goro/ui/rotheme"
)

// Minimap defaults. Text colours match the shared text and muted text colours.
var (
	minimapWidth           = 188
	minimapHeight          = 206
	minimapPad             = 10
	minimapInfoBandH       = 22
	minimapMapMin          = 32
	minimapMapTop          = 4
	minimapFooterOff       = 19
	minimapLabelH          = 16
	minimapCoordsPad       = 2
	minimapCoordsExtra     = 4
	minimapCoordsH         = 20
	minimapMarkerSize      = 11
	minimapMarkerExpand    = 4
	minimapPlayerRadius    = 4
	minimapShadowExtra     = 3
	minimapCrossLength     = 8
	minimapCrossThick      = 2
	minimapPartyOuter      = 6
	minimapPartyInner      = 4
	minimapGuildMarkerSize = 3

	minimapTextColor   = color.RGBA{R: 38, G: 48, B: 58, A: 255}
	minimapMutedColor  = color.RGBA{R: 98, G: 112, B: 126, A: 255}
	minimapPlayerColor = color.RGBA{R: 255, G: 232, B: 96, A: 255}
	minimapGuildColor  = color.RGBA{R: 255, G: 140, B: 26, A: 255}
	minimapBossColor   = color.RGBA{R: 255, G: 0, B: 0, A: 255}
	minimapPartyColor  = color.RGBA{R: 255, G: 255, B: 255, A: 255}
	minimapShadowColor = color.RGBA{A: 190}
)

func init() {
	s := rotheme.Section("minimap", "mini map window")
	s.Int("width", &minimapWidth, 80, 640, "minimap width")
	s.Int("height", &minimapHeight, 80, 640, "minimap height")
	s.Int("padding", &minimapPad, 0, 48, "padding around the map")
	s.Int("info_band_height", &minimapInfoBandH, 0, 80, "map name and coordinate band height")
	s.Int("map_min_size", &minimapMapMin, 8, 256, "minimum map image size")
	s.Int("map_top", &minimapMapTop, 0, 48, "gap between the title bar and the map")
	s.Int("footer_offset", &minimapFooterOff, 0, 64, "distance from the bottom to the coordinate text")
	s.Int("label_height", &minimapLabelH, 8, 48, "map name and coordinate text height")
	s.Int("coords_pad", &minimapCoordsPad, 0, 16, "padding around the coordinate dirty rect")
	s.Int("coords_extra", &minimapCoordsExtra, 0, 24, "extra width of the coordinate dirty rect")
	s.Int("coords_height", &minimapCoordsH, 8, 48, "coordinate dirty rect height")
	s.Int("marker_size", &minimapMarkerSize, 4, 48, "player marker dirty size")
	s.Int("marker_expand", &minimapMarkerExpand, 0, 16, "padding around the player marker dirty rect")
	s.Int("player_radius", &minimapPlayerRadius, 1, 16, "fallback player marker radius")
	s.Int("shadow_extra", &minimapShadowExtra, 0, 12, "extra pixels on the marker shadow box")
	s.Int("cross_length", &minimapCrossLength, 2, 32, "compass and boss cross length")
	s.Int("cross_thickness", &minimapCrossThick, 1, 12, "compass and boss cross thickness")
	s.Int("party_outer_size", &minimapPartyOuter, 1, 24, "party marker outer square")
	s.Int("party_inner_size", &minimapPartyInner, 1, 24, "party marker inner square")
	s.Int("guild_marker_size", &minimapGuildMarkerSize, 1, 24, "guild member marker size")
	s.RGBA("text_color", &minimapTextColor, "map name text")
	s.RGBA("muted_color", &minimapMutedColor, "coordinate text")
	s.RGBA("player_color", &minimapPlayerColor, "player marker")
	s.RGBA("guild_color", &minimapGuildColor, "guild member marker")
	s.RGBA("boss_color", &minimapBossColor, "boss marker")
	s.RGBA("party_color", &minimapPartyColor, "party marker outline")
	s.RGBA("shadow_color", &minimapShadowColor, "fallback player marker shadow")
}
