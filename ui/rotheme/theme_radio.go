package rotheme

var (
	RadioSize        float32 = 13
	RadioGap         float32 = 7
	radioDotRadius   float32 = 3
	radioFocusOutset float32 = 2
)

func init() {
	s := Section("radio", "radio button and its label")
	s.Float("size", &RadioSize, 6, 48, "radio circle diameter")
	s.Float("gap", &RadioGap, 0, 32, "gap between the circle and the label")
	s.Float("dot_radius", &radioDotRadius, 0, 16, "selected radio dot radius")
	s.Float("focus_outset", &radioFocusOutset, 0, 12, "focus ring outside the circle")
}
