package ui

import "github.com/kivutar/goro/ui/rotheme"

// [item_info] defaults: the item description pop-up.
var (
	itemInfoWindowWidth       = 340
	itemInfoWindowMaxHeight   = 304
	itemInfoWindowPad         = 10
	itemInfoIllustrationWidth = 75
	itemInfoIllustrationH     = 100
	itemInfoSlotIcon          = 24
	itemInfoLineH             = 14
	itemInfoPanelGap          = 12
	itemInfoInfoGap           = 8
	itemInfoOpenOffsetX       = 14
	itemInfoOpenOffsetY       = 22
	itemInfoTooltipBelow      = 18
	itemInfoTooltipAbove      = 6
	itemInfoDescInset         = 18
	itemInfoRuneW             = 7

	itemInfoDescriptionW    int
	itemInfoDescriptionMaxH int
)

func recomputeItemInfoMetrics() {
	itemInfoDescriptionW = itemInfoWindowWidth - itemInfoWindowPad*2 - itemInfoIllustrationWidth - itemInfoPanelGap
	itemInfoDescriptionMaxH = itemInfoWindowMaxHeight - ROWindowTitleHeight - itemInfoWindowPad*2
}

func init() {
	s := rotheme.Section("item_info", "item description pop-up")
	s.Int("width", &itemInfoWindowWidth, 200, 960, "item info window width")
	s.Int("max_height", &itemInfoWindowMaxHeight, 120, 900, "item info window maximum height")
	s.Int("padding", &itemInfoWindowPad, 0, 40, "item info padding")
	s.Int("illustration_width", &itemInfoIllustrationWidth, 32, 240, "item illustration width")
	s.Int("illustration_height", &itemInfoIllustrationH, 40, 320, "item illustration height")
	s.Int("slot_icon", &itemInfoSlotIcon, 12, 64, "card slot icon size")
	s.Int("line_height", &itemInfoLineH, 10, 32, "item description line height")
	s.Int("panel_gap", &itemInfoPanelGap, 0, 40, "gap between the illustration and the description")
	s.Int("info_gap", &itemInfoInfoGap, 0, 32, "gap inside the description column")
	s.Int("open_offset_x", &itemInfoOpenOffsetX, 0, 64, "pop-up offset to the right of the cursor")
	s.Int("open_offset_y", &itemInfoOpenOffsetY, 0, 80, "pop-up offset above the cursor")
	s.Int("tooltip_below", &itemInfoTooltipBelow, 0, 64, "card tooltip offset below the cursor")
	s.Int("tooltip_above", &itemInfoTooltipAbove, 0, 48, "card tooltip offset above the cursor")
	s.Int("description_inset", &itemInfoDescInset, 0, 64, "description width kept clear of the edges when wrapping")
	s.Int("rune_width", &itemInfoRuneW, 4, 24, "approximate pixels per description character")
	rotheme.OnLayoutApplied(recomputeItemInfoMetrics)
	recomputeItemInfoMetrics()
}
