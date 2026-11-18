package ui

import "github.com/kivutar/goro/ui/rotheme"

// Half of the 35px cell ROBrowser's 2D SpriteRenderer subtracts from canvas Y.
var uiSpriteRendererMidCellY float32 = 17.5

func init() {
	s := rotheme.Section("sprite", "UI sprite layer anchor")
	s.Float("mid_cell_y", &uiSpriteRendererMidCellY, 0, 70, "sprite anchor offset, half of a 35px cell")
}
