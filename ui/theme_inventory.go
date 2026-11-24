package ui

import (
	"github.com/gogpu/ui/widget"
	"github.com/kivutar/goro/ui/rotheme"
)

// [inventory] defaults: the inventory bag. Cell size, columns and visible
// rows were the first tokens; the window size follows them.
var (
	inventoryBagTabW                     = 32
	inventoryBagTabH                     = 44
	inventoryBagFilterH                  = 22
	inventoryBagFilterCols               = 3
	inventoryBagSearchH                  = 24
	inventoryBagSortW                    = 64
	inventoryBagCell                     = 32
	inventoryBagCols                     = 8
	opinionatedBagCols                   = 10
	opinionatedSplitCols                 = 6
	inventoryBagSplitFrame               = 4
	inventoryBagSearchClear              = 16
	inventoryBagSearchInset              = 6
	inventoryBagRows                     = 5
	inventoryBagCellShadowInset  float32 = 3
	inventoryBagCellShadowRadius float32 = 3
	inventoryBagCellShadowAlpha  float32 = 0.4
	inventoryBagCellHoverAlpha   float32 = 0.75
	inventoryBagIconPad          float32 = 4
	inventoryBagAmountInsetX     float32 = 18
	inventoryBagAmountInsetY     float32 = 15
	inventoryBagAmountW          float32 = 16
	inventoryBagAmountH          float32 = 12
	inventoryBagEquipInset       float32 = 2
	inventoryBagEquipSize        float32 = 12
	inventoryBagEquippedColor            = widget.RGBA8(54, 128, 76, 255)
	inventoryBagTabActive                = widget.RGBA8(118, 160, 206, 255)
	inventoryBagMenuGap                  = 8
	inventoryBagTooltipBelow             = 18
	inventoryBagTooltipAbove             = 6

	inventoryBagTabRail int
	inventoryBagWidth   int
	inventoryBagViewW   int
	inventoryBagGridW   int
	inventoryBagViewH   int
	inventoryBagHeight  int
)

func recomputeInventoryBagMetrics() {
	inventoryBagTabRail = inventoryBagTabW + inventoryBagTabOver*2
	inventoryBagWidth = characterWindowWidth
	inventoryBagViewW = inventoryBagWidth - inventoryBagTabRail - verticalTabDividerW
	inventoryBagGridW = inventoryBagCols * inventoryBagCell
	inventoryBagViewH = inventoryBagRows * inventoryBagCell
	inventoryBagHeight = ROWindowTitleHeight + inventoryBagViewH
}

func init() {
	s := rotheme.Section("inventory", "inventory bag")
	s.Int("cell_size", &inventoryBagCell, 16, 96, "inventory cell size")
	s.Int("cols", &inventoryBagCols, 1, 16, "inventory columns")
	s.Int("opinionated_cols", &opinionatedBagCols, 4, 16, "opinionated inventory columns")
	s.Int("split_cols", &opinionatedSplitCols, 1, 8, "columns in each split inventory pane")
	s.Int("split_frame", &inventoryBagSplitFrame, 1, 16, "thin frame between split inventory panes")
	s.Int("search_clear", &inventoryBagSearchClear, 8, 32, "opinionated search clear mark width")
	s.Int("search_inset", &inventoryBagSearchInset, 0, 24, "opinionated search bar inset")
	s.Int("rows", &inventoryBagRows, 1, 20, "visible inventory rows")
	s.Int("tab_width", &inventoryBagTabW, 16, 96, "inventory tab width")
	s.Int("tab_height", &inventoryBagTabH, 20, 120, "inventory tab height")
	s.Int("filter_height", &inventoryBagFilterH, 16, 48, "inventory filter row height")
	s.Int("filter_cols", &inventoryBagFilterCols, 1, 8, "inventory filter tabs per row")
	s.Int("search_height", &inventoryBagSearchH, 16, 48, "opinionated inventory search row height")
	s.Int("sort_width", &inventoryBagSortW, 40, 160, "opinionated inventory sort button width")
	s.Float("shadow_inset", &inventoryBagCellShadowInset, 0, 16, "inventory cell shadow inset")
	s.Float("shadow_radius", &inventoryBagCellShadowRadius, 0, 16, "inventory cell shadow radius")
	s.Float("shadow_alpha", &inventoryBagCellShadowAlpha, 0, 1, "inventory cell shadow opacity")
	s.Float("hover_alpha", &inventoryBagCellHoverAlpha, 0, 1, "inventory cell hover opacity")
	s.Float("icon_pad", &inventoryBagIconPad, 0, 16, "inventory item icon inset")
	s.Float("amount_inset_x", &inventoryBagAmountInsetX, 0, 48, "inventory amount label inset from the right")
	s.Float("amount_inset_y", &inventoryBagAmountInsetY, 0, 48, "inventory amount label inset from the bottom")
	s.Float("amount_width", &inventoryBagAmountW, 8, 48, "inventory amount label width")
	s.Float("amount_height", &inventoryBagAmountH, 8, 40, "inventory amount label height")
	s.Float("equipped_inset", &inventoryBagEquipInset, 0, 16, "equipped marker inset")
	s.Float("equipped_size", &inventoryBagEquipSize, 8, 40, "equipped marker size")
	s.Color("equipped_color", &inventoryBagEquippedColor, "equipped marker color")
	s.Color("tab_active", &inventoryBagTabActive, "selected inventory tab highlight")
	s.Int("menu_gap", &inventoryBagMenuGap, 0, 32, "gap between the basic menu and the inventory")
	s.Int("tooltip_below", &inventoryBagTooltipBelow, 0, 64, "tooltip offset below the cursor")
	s.Int("tooltip_above", &inventoryBagTooltipAbove, 0, 48, "tooltip offset above the cursor")
	rotheme.OnLayoutApplied(recomputeInventoryBagMetrics)
	recomputeInventoryBagMetrics()
}

func inventoryBagFilterRowCount() int {
	cols := inventoryBagFilterCols
	if cols < 1 {
		cols = 1
	}
	count := len(opinionatedInventoryTabs)
	if count < 1 {
		return 1
	}
	return (count + cols - 1) / cols
}

func inventoryBagFilterBandH() int {
	return inventoryBagFilterRowCount() * inventoryBagFilterH
}

func inventorySearchRowH() int {
	return inventoryBagSearchH + inventoryBagSearchInset*2
}

func inventoryOpinionatedChromeH() int {
	return inventorySearchRowH() + inventoryBagFilterH + inventoryBagFilterH
}

func opinionatedContentW() int {
	width := opinionatedFrameW() - ROScrollbarGutter
	if width < 1 {
		return 1
	}
	return width
}

func opinionatedFrameW() int {
	return opinionatedSplitCols*inventoryBagCell*2 + opinionatedDividerW() + ROScrollbarGutter*2
}

func opinionatedDividerW() int {
	gap := inventoryBagSplitFrame
	if gap < 1 {
		return 1
	}
	return gap
}

func opinionatedSideViewW() int {
	remain := opinionatedFrameW() - opinionatedDividerW()
	if remain < 2 {
		return 1
	}
	return remain / 2
}

func opinionatedSideViewWRight() int {
	return opinionatedFrameW() - opinionatedDividerW() - opinionatedSideViewW()
}

func opinionatedEqualTabWidth(index, count int) int {
	width := opinionatedFrameW()
	if count < 1 {
		return width
	}
	base := width / count
	if index == count-1 {
		return width - base*(count-1)
	}
	if base < 1 {
		return 1
	}
	return base
}

func opinionatedSplitContentW() int {
	width := opinionatedSideViewW() - ROScrollbarGutter
	if width < 1 {
		return 1
	}
	return width
}

func opinionatedSplitContentWRight() int {
	width := opinionatedSideViewWRight() - ROScrollbarGutter
	if width < 1 {
		return 1
	}
	return width
}

func opinionatedScrollW(contentW int) int {
	return contentW + ROScrollbarGutter
}

func opinionatedWidthBounded(width float32) bool {
	return width > 0 && width <= float32(opinionatedFrameW()*opinionatedBagCols)
}

func inventoryBagFrameSize(opinionated bool) (int, int) {
	if opinionated {
		return opinionatedFrameW(), inventoryBagViewH + inventoryOpinionatedChromeH()
	}
	return inventoryBagWidth, inventoryBagHeight
}

func inventoryOpinionatedViewW() int {
	return inventoryBagWidth
}

func inventoryBagFilterTabWidth(indexInRow, countInRow int) int {
	cols := inventoryBagFilterCols
	if cols < 1 {
		cols = 1
	}
	base := inventoryBagWidth / cols
	if countInRow < cols && indexInRow == countInRow-1 {
		return inventoryBagWidth - base*(countInRow-1)
	}
	if indexInRow == cols-1 {
		return inventoryBagWidth - base*(cols-1)
	}
	if base < 1 {
		return 1
	}
	return base
}
