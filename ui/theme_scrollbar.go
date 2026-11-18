package ui

import "github.com/kivutar/goro/ui/rotheme"

// ROScrollbarGutter is the width reserved beside scrollable lists. Windows
// that add it to a package-level width (cart, skill grid) recompute those
// widths in their own rotheme.OnLayoutApplied hooks.
var ROScrollbarGutter = 12

func init() {
	s := rotheme.Section("scrollbar", "scrollbar gutter reserved beside scrollable lists")
	s.Int("gutter", &ROScrollbarGutter, 0, 48, "scrollbar gutter width")
}
