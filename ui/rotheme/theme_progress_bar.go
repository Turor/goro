package rotheme

import "github.com/gogpu/ui/widget"

// Progress bar metrics. height stays registered in theme_tokens.go.
var (
	ProgressBarHeight    float32 = 12
	progressBarLowFill           = widget.Hex(0xCE768A)
	progressBarLowBorder         = widget.Hex(0xD68AA2)
)

func init() {
	s := Section("progress_bar", "progress and vital bars")
	s.Color("low_fill", &progressBarLowFill, "vital bar fill below the warning threshold")
	s.Color("low_border", &progressBarLowBorder, "vital bar border below the warning threshold")
}
