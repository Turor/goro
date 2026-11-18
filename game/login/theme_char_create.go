package login

import "github.com/kivutar/goro/ui/rotheme"

// Character create preview sprite. The create window tokens share this section.
var charCreatePreviewScale = 1.08

func init() {
	s := rotheme.Section("char_create", "character create screen")
	s.Float64("preview_scale", &charCreatePreviewScale, 0.2, 3, "create preview sprite scale")
}
