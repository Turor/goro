package ui

import "github.com/kivutar/goro/ui/rotheme"

// [auto_spell] defaults: the auto-spell skill list.
var (
	autoSpellWindowWidth  = 312
	autoSpellTableHeaderH = 24
	autoSpellRowH         = 32
	autoSpellRows         = 7
	autoSpellWindowHeight int
)

func init() {
	s := rotheme.Section("auto_spell", "auto spell skill list")
	s.Int("width", &autoSpellWindowWidth, 200, 900, "window width")
	s.Int("header_height", &autoSpellTableHeaderH, 12, 80, "table header height")
	s.Int("row_height", &autoSpellRowH, 12, 96, "skill row height")
	s.Int("visible_rows", &autoSpellRows, 1, 24, "visible skill rows")
	rotheme.OnLayoutApplied(recomputeAutoSpellMetrics)
	recomputeAutoSpellMetrics()
}

func recomputeAutoSpellMetrics() {
	autoSpellWindowHeight = ROWindowTitleHeight + autoSpellTableHeaderH + autoSpellRows*autoSpellRowH + ROWindowFooterHeight
}
