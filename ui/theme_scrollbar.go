package ui

import "github.com/kivutar/goro/ui/rotheme"

var ROScrollbarGutter = 12

func init() {
	s := rotheme.Section("scrollbar", "scrollbar gutter reserved beside scrollable lists")
	s.Int("gutter", &ROScrollbarGutter, 0, 48, "scrollbar gutter width")
	rotheme.OnLayoutApplied(recomputeScrollbarGutterCopies)
}

// Grid widths that add the gutter are package vars in other windows. Recompute
// them so a reloaded gutter reaches those windows.
func recomputeScrollbarGutterCopies() {
	cartGridViewW = cartGridW + ROScrollbarGutter
	cartWindowWidth = cartGridLeftPad + cartGridViewW
	skillGridViewW = skillGridColumns*skillGridCellW + ROScrollbarGutter
	skillGridWindowWidth = skillTabRailW + verticalTabDividerW + skillGridViewW
}
