package rotheme

import "github.com/gogpu/ui/widget"

// [listview] defaults: select-list rows.
var (
	SelectListRowPadX float32 = 6
	selectListRowFill         = widget.RGBA8(246, 249, 253, 255)
)

func init() {
	s := Section("listview", "select list rows")
	s.Float("row_pad_x", &SelectListRowPadX, 0, 24, "row text horizontal inset")
	s.Color("row_fill", &selectListRowFill, "even row background")
}
