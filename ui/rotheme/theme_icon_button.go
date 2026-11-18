package rotheme

var (
	IconButtonSize                 float32 = 17
	iconButtonGlyphYOffset         float32 = 0.25
	iconButtonTriangleSize         float32 = 6
	iconButtonMinGlyph                     = 6
	iconButtonSearchHandle         float32 = 2
	directionalArrowCornerFraction float32 = 0.24
	directionalArrowBorderInset    float32 = 1.25
	iconArrowLROuterX              float32 = 12.95
	iconArrowLRTipX                float32 = 3.05
	iconArrowLRTopY                float32 = 1.65
	iconArrowLRMidY                float32 = 8.25
	iconArrowLRBottomY             float32 = 14.85
	iconArrowUDLeftX               float32 = 1.4
	iconArrowUDRightX              float32 = 14.6
	iconArrowUDCenterX             float32 = 8
	iconArrowUDBaseY               float32 = 13.2
	iconArrowUDTipY                float32 = 3.3
)

func init() {
	s := Section("icon_button", "icon buttons and their glyphs")
	s.Float("size", &IconButtonSize, 8, 64, "icon button width and height")
	s.Float("glyph_y_offset", &iconButtonGlyphYOffset, -8, 8, "glyph vertical nudge")
	s.Float("triangle_size", &iconButtonTriangleSize, 2, 32, "collapse and expand triangle size")
	s.Int("min_glyph", &iconButtonMinGlyph, 2, 32, "minimum cross and search glyph length")
	s.Float("search_handle", &iconButtonSearchHandle, 0, 16, "search icon handle length")
	s.Float("arrow_corner_fraction", &directionalArrowCornerFraction, 0, 1, "directional arrow corner rounding")
	s.Float("arrow_border_inset", &directionalArrowBorderInset, 0, 8, "directional arrow border thickness")
	s.Float("arrow_lr_outer_x", &iconArrowLROuterX, 0, 48, "left and right arrow outer x")
	s.Float("arrow_lr_tip_x", &iconArrowLRTipX, 0, 48, "left and right arrow tip x")
	s.Float("arrow_lr_top_y", &iconArrowLRTopY, 0, 48, "left and right arrow top y")
	s.Float("arrow_lr_mid_y", &iconArrowLRMidY, 0, 48, "left and right arrow middle y")
	s.Float("arrow_lr_bottom_y", &iconArrowLRBottomY, 0, 48, "left and right arrow bottom y")
	s.Float("arrow_ud_left_x", &iconArrowUDLeftX, 0, 48, "up and down arrow left x")
	s.Float("arrow_ud_right_x", &iconArrowUDRightX, 0, 48, "up and down arrow right x")
	s.Float("arrow_ud_center_x", &iconArrowUDCenterX, 0, 48, "up and down arrow center x")
	s.Float("arrow_ud_base_y", &iconArrowUDBaseY, 0, 48, "up and down arrow base y")
	s.Float("arrow_ud_tip_y", &iconArrowUDTipY, 0, 48, "up and down arrow tip y")
}
