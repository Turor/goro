package rotheme

import "github.com/gogpu/ui/widget"

// [table] defaults: compact tables, table cells, and table views.
var (
	TableRowHeight     float32 = 14
	TableGap           float32 = 1
	TableCellPadX      float32 = 4
	TableHeaderPadY    float32 = 4
	tableIconInset     float32 = 6
	tableIconTextGap   float32 = 12
	tableCellLineScale float32 = 1.2
	tableViewRowHeight float32 = 32
	tableViewFallbackW float32 = 240
	tableViewRowFill           = widget.RGBA8(246, 249, 253, 255)
)

func init() {
	s := Section("table", "compact tables, cells, and table views")
	s.Float("row_height", &TableRowHeight, 8, 56, "compact table row height")
	s.Float("gap", &TableGap, 0, 8, "gap between compact table cells")
	s.Float("cell_pad_x", &TableCellPadX, 0, 24, "cell text horizontal inset")
	s.Float("header_pad_y", &TableHeaderPadY, 0, 24, "table view header text vertical inset")
	s.Float("icon_inset", &tableIconInset, 0, 24, "cell icon inset from the left")
	s.Float("icon_text_gap", &tableIconTextGap, 0, 48, "gap added after a cell icon before its text")
	s.Float("cell_line_scale", &tableCellLineScale, 0.5, 3, "cell text box height as a multiple of the text size")
	s.Float("view_row_height", &tableViewRowHeight, 12, 128, "table view row height")
	s.Float("view_fallback_width", &tableViewFallbackW, 40, 960, "table view width when the parent does not bound it")
	s.Color("view_row_fill", &tableViewRowFill, "table view even row background")
}
