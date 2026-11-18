package login

import "github.com/kivutar/goro/ui/rotheme"

// Character slot preview sprite. The select window tokens share this section.
var (
	charSelectPreviewScale       = 0.92
	charSelectPreviewFeetLift    = 10
	charSelectPreviewBottomInset = 15
)

func init() {
	s := rotheme.Section("char_select", "character select screen")
	s.Float64("preview_scale", &charSelectPreviewScale, 0.2, 3, "character preview sprite scale")
	s.Int("preview_feet_lift", &charSelectPreviewFeetLift, 0, 64, "character preview lift above the slot floor")
	s.Int("preview_bottom_inset", &charSelectPreviewBottomInset, 0, 64, "slot floor inset before the feet lift")
}
