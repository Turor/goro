package ui

import "github.com/kivutar/goro/ui/rotheme"

// [cart] defaults: the pushcart window. Width and height follow the grid.
var (
	cartGridCols     = 7
	cartGridRows     = 4
	cartGridCell     = 32
	cartGridLeftPad  = 40
	cartTooltipBelow = 18
	cartTooltipAbove = 6
	cartScreenTop    = 118

	cartGridW        int
	cartGridViewW    int
	cartGridViewH    int
	cartWindowWidth  int
	cartWindowHeight int
)

func init() {
	s := rotheme.Section("cart", "pushcart window")
	s.Int("cols", &cartGridCols, 1, 16, "cart columns")
	s.Int("rows", &cartGridRows, 1, 16, "visible cart rows")
	s.Int("cell_size", &cartGridCell, 16, 96, "cart cell size")
	s.Int("left_padding", &cartGridLeftPad, 0, 120, "padding left of the cart grid")
	s.Int("tooltip_below", &cartTooltipBelow, 0, 64, "tooltip offset below the cursor")
	s.Int("tooltip_above", &cartTooltipAbove, 0, 64, "tooltip offset above the cursor")
	s.Int("screen_top", &cartScreenTop, 0, 400, "distance from the screen top when opened")
	rotheme.OnLayoutApplied(recomputeCartMetrics)
	recomputeCartMetrics()
}

func recomputeCartMetrics() {
	cartGridW = cartGridCols * cartGridCell
	cartGridViewW = cartGridW + ROScrollbarGutter
	cartGridViewH = cartGridRows * cartGridCell
	cartWindowWidth = cartGridLeftPad + cartGridViewW
	cartWindowHeight = ROWindowTitleHeight + cartGridViewH + ROWindowFooterHeight
}
