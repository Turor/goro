package rotheme

import "github.com/gogpu/ui/widget"

var (
	SliderTrackHeight float32 = 6
	SliderThumbSize   float32 = 16
	sliderTrackColor          = widget.RGBA8(224, 232, 242, 255)
)

func init() {
	s := Section("slider", "slider track and thumb")
	s.Float("track_height", &SliderTrackHeight, 1, 24, "slider track thickness")
	s.Float("thumb_size", &SliderThumbSize, 4, 64, "slider thumb size")
	s.Color("track", &sliderTrackColor, "slider track background")
}
