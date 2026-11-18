package rotheme

import "github.com/gogpu/ui/widget"

// [text_area] defaults: the multiline editor.
var (
	textAreaLineHeight float32 = 18
	textAreaWidth      float32 = 300
	textAreaHeight     float32 = 108
	textAreaPadding    float32 = 6
	textAreaCaretInset float32 = 2
	textAreaSelection          = widget.RGBA8(121, 169, 215, 110)
)

func init() {
	s := Section("text_area", "multiline text editor")
	s.Float("line_height", &textAreaLineHeight, 8, 72, "line height")
	s.Float("width", &textAreaWidth, 80, 1200, "preferred width")
	s.Float("height", &textAreaHeight, 36, 432, "preferred height")
	s.Float("padding", &textAreaPadding, 0, 24, "content padding")
	s.Float("caret_inset", &textAreaCaretInset, 0, 12, "caret inset from the top and bottom of the line")
	s.Color("selection", &textAreaSelection, "selection highlight")
}
