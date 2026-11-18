package rotheme

var (
	CheckboxSize       float32 = 17
	CheckboxRadius     float32 = 6
	CheckboxGap        float32 = 8
	checkboxMarkInset  float32 = 4
	checkboxMarkStartY float32 = 8
	checkboxMarkMidX   float32 = 7
	checkboxMarkMidY   float32 = 11
	checkboxMarkEndX   float32 = 14
	checkboxMarkEndY   float32 = 4
)

func init() {
	s := Section("checkbox", "checkbox box and check mark")
	s.Float("size", &CheckboxSize, 8, 64, "checkbox box size")
	s.Float("radius", &CheckboxRadius, 0, 24, "checkbox corner radius")
	s.Float("gap", &CheckboxGap, 0, 32, "gap between the box and the label")
	s.Float("mark_inset", &checkboxMarkInset, 0, 32, "check mark and dash inset")
	s.Float("mark_start_y", &checkboxMarkStartY, 0, 32, "check mark start y")
	s.Float("mark_mid_x", &checkboxMarkMidX, 0, 32, "check mark corner x")
	s.Float("mark_mid_y", &checkboxMarkMidY, 0, 32, "check mark corner y")
	s.Float("mark_end_x", &checkboxMarkEndX, 0, 32, "check mark end x")
	s.Float("mark_end_y", &checkboxMarkEndY, 0, 32, "check mark end y")
}
