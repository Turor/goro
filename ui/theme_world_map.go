package ui

import (
	"github.com/gogpu/ui/widget"
	"github.com/kivutar/goro/ui/rotheme"
)

// World map atlas overlay.
var (
	worldMapHoverFill   = widget.Color{R: 1, G: 1, B: 1, A: .2}
	worldMapLabelFill   = widget.Color{R: 1, G: 1, B: 1, A: .85}
	worldMapPreviewFill = widget.Color{R: 1, G: 1, B: 1, A: .95}

	worldMapHoverStroke    float32 = 1
	worldMapMemberStroke   float32 = 2
	worldMapNameSize       float32 = 9
	worldMapNameH          float32 = 14
	worldMapStarRadius     float32 = 7
	worldMapStarStroke     float32 = 2
	worldMapPreviewW       float32 = 190
	worldMapPreviewEdge    float32 = 8
	worldMapPreviewInset   float32 = 4
	worldMapPreviewRadius  float32 = 4
	worldMapPreviewBorder  float32 = 1
	worldMapPreviewPadX    float32 = 6
	worldMapPreviewLabelY  float32 = 3
	worldMapPreviewLabelH  float32 = 18
	worldMapPreviewImageY  float32 = 22
	worldMapPreviewImageH  float32 = 128
	worldMapPreviewMemberY float32 = 153
	worldMapPreviewRowH    float32 = 16
	worldMapPreviewBaseH   float32 = 156
	worldMapPreviewRowsTop float32 = 164
	worldMapPreviewMax             = 128
	worldMapDotSize                = 4
)

func init() {
	s := rotheme.Section("world_map", "world map atlas")
	s.Color("hover_fill", &worldMapHoverFill, "hovered map highlight")
	s.Color("label_fill", &worldMapLabelFill, "map name label background")
	s.Color("preview_fill", &worldMapPreviewFill, "preview panel background")
	s.Float("hover_stroke", &worldMapHoverStroke, 0, 8, "hovered map outline width")
	s.Float("member_stroke", &worldMapMemberStroke, 0, 8, "party member map outline width")
	s.Float("name_size", &worldMapNameSize, 6, 32, "map name text size")
	s.Float("name_height", &worldMapNameH, 8, 40, "map name label height")
	s.Float("star_radius", &worldMapStarRadius, 2, 32, "current map star radius")
	s.Float("star_stroke", &worldMapStarStroke, 1, 8, "current map star stroke width")
	s.Float("preview_width", &worldMapPreviewW, 40, 600, "preview panel width")
	s.Float("preview_edge", &worldMapPreviewEdge, 0, 64, "preview margin from the atlas edge")
	s.Float("preview_inset", &worldMapPreviewInset, 0, 32, "preview inset from the corner")
	s.Float("preview_radius", &worldMapPreviewRadius, 0, 24, "preview panel corner radius")
	s.Float("preview_border", &worldMapPreviewBorder, 0, 8, "preview panel border width")
	s.Float("preview_pad_x", &worldMapPreviewPadX, 0, 32, "preview text horizontal padding")
	s.Float("preview_label_y", &worldMapPreviewLabelY, 0, 40, "preview title offset from the top")
	s.Float("preview_label_height", &worldMapPreviewLabelH, 8, 48, "preview title height")
	s.Float("preview_image_y", &worldMapPreviewImageY, 0, 80, "preview image offset from the top")
	s.Float("preview_image_height", &worldMapPreviewImageH, 32, 400, "preview image area height")
	s.Float("preview_member_y", &worldMapPreviewMemberY, 0, 400, "party name list offset from the top")
	s.Float("preview_row_height", &worldMapPreviewRowH, 8, 40, "party name row height")
	s.Float("preview_base_height", &worldMapPreviewBaseH, 40, 400, "preview panel height before party names")
	s.Float("preview_rows_top", &worldMapPreviewRowsTop, 0, 400, "atlas height reserved above the party name rows")
	s.Int("preview_max", &worldMapPreviewMax, 32, 512, "longest side of a preview minimap")
	s.Int("dot_size", &worldMapDotSize, 1, 16, "player and party dot size")
}
