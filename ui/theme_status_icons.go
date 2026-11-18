package ui

import (
	"image/color"

	"github.com/gogpu/ui/widget"
	"github.com/kivutar/goro/ui/rotheme"
)

// [status_icons] defaults: the buff icon strip under the minimap.
var (
	statusIconSize                     = 32
	statusIconSpacing                  = 36
	statusIconGap                      = 8
	statusIconFallbackDY               = 7
	statusIconFallbackTextH            = 16
	statusIconFontSize         float32 = 11
	statusIconDurationH                = 4
	statusIconDurationTrack            = widget.RGBA8(18, 24, 34, 180)
	statusIconDurationFill             = widget.RGBA8(244, 228, 130, 230)
	statusIconDrawBottom               = 16
	statusIconTooltipLineH             = 16
	statusIconTooltipExtra             = 6
	statusIconTooltipCursor            = 12
	statusIconTooltipMargin            = 4
	statusIconTooltipTextX             = 5
	statusIconTooltipTextY             = 3
	statusIconTooltipTextInset         = 10
	statusIconTooltipTextH             = 14
	statusIconTooltipCharW             = 7
	statusIconTooltipMinW              = 42
	statusIconTooltipMaxW              = 300
	statusIconTooltipBg                = widget.RGBA8(0, 0, 0, 128)
	statusIconTooltipBorder            = widget.RGBA8(198, 198, 198, 255)
	statusIconTooltipText              = color.RGBA{R: 255, G: 255, B: 255, A: 255}
)

func init() {
	s := rotheme.Section("status_icons", "buff icon strip under the minimap")
	s.Int("size", &statusIconSize, 8, 128, "status icon size")
	s.Int("spacing", &statusIconSpacing, 8, 144, "status icon row stride")
	s.Int("gap", &statusIconGap, 0, 48, "gap between status icon columns and below the minimap")
	s.Int("fallback_text_dy", &statusIconFallbackDY, 0, 32, "missing-icon question mark offset")
	s.Int("fallback_text_height", &statusIconFallbackTextH, 4, 48, "missing-icon question mark height")
	s.Float("font_size", &statusIconFontSize, 6, 32, "status icon text size")
	s.Int("duration_height", &statusIconDurationH, 1, 16, "duration bar height")
	s.Color("duration_track", &statusIconDurationTrack, "duration bar track")
	s.Color("duration_fill", &statusIconDurationFill, "duration bar fill")
	s.Int("draw_bottom", &statusIconDrawBottom, 0, 64, "space kept below the icon strip when counting rows")
	s.Int("tooltip_line_height", &statusIconTooltipLineH, 8, 48, "status tooltip line height")
	s.Int("tooltip_extra", &statusIconTooltipExtra, 0, 32, "status tooltip extra height")
	s.Int("tooltip_cursor", &statusIconTooltipCursor, 0, 48, "status tooltip offset from the cursor")
	s.Int("tooltip_margin", &statusIconTooltipMargin, 0, 32, "status tooltip screen margin")
	s.Int("tooltip_text_x", &statusIconTooltipTextX, 0, 32, "status tooltip text left inset")
	s.Int("tooltip_text_y", &statusIconTooltipTextY, 0, 32, "status tooltip text top inset")
	s.Int("tooltip_text_inset", &statusIconTooltipTextInset, 0, 48, "status tooltip horizontal text inset")
	s.Int("tooltip_text_height", &statusIconTooltipTextH, 6, 40, "status tooltip text line height")
	s.Int("tooltip_char_width", &statusIconTooltipCharW, 1, 32, "status tooltip estimated character width")
	s.Int("tooltip_min_width", &statusIconTooltipMinW, 8, 200, "status tooltip minimum width")
	s.Int("tooltip_max_width", &statusIconTooltipMaxW, 40, 800, "status tooltip maximum width")
	s.Color("tooltip_background", &statusIconTooltipBg, "status tooltip background")
	s.Color("tooltip_border", &statusIconTooltipBorder, "status tooltip border")
	s.RGBA("tooltip_text", &statusIconTooltipText, "status tooltip text")
}
