package ui

import "github.com/kivutar/goro/ui/rotheme"

// Warpra destination list and the world-map dungeon floor picker.
var (
	warpraWindowW  = 720
	warpraWindowH  = 540
	warpraTabH     = 24
	warpraRowH     = 22
	warpraListRows = 10
	warpraPad      = 10
	warpraGap      = 6
	warpraTabGap   = 2
	warpraBackW    = 64
	warpraBackGap  = 8

	dungeonPickerW      = 360
	dungeonPickerMinH   = 220
	dungeonPickerRowH   = 22
	dungeonPickerPad    = 12
	dungeonPickerMax    = 8
	dungeonPickerIntroH = 36
	dungeonPickerGap    = 8
)

func init() {
	s := rotheme.Section("warpra", "Warpra destination list")
	s.Int("width", &warpraWindowW, 240, 900, "window width")
	s.Int("height", &warpraWindowH, 180, 800, "window height")
	s.Int("tab_height", &warpraTabH, 12, 48, "tab and back button height")
	s.Int("row_height", &warpraRowH, 12, 48, "destination row height")
	s.Int("list_rows", &warpraListRows, 4, 30, "destination rows shown")
	s.Int("padding", &warpraPad, 0, 32, "content padding")
	s.Int("gap", &warpraGap, 0, 24, "gap between tabs, crumb, and list")
	s.Int("tab_gap", &warpraTabGap, 0, 24, "gap between tab buttons")
	s.Int("back_width", &warpraBackW, 32, 160, "back button width")
	s.Int("back_gap", &warpraBackGap, 0, 24, "gap between the back button and the crumb")

	d := rotheme.Section("dungeon_picker", "world map dungeon floor picker")
	d.Int("width", &dungeonPickerW, 200, 800, "picker width")
	d.Int("min_height", &dungeonPickerMinH, 80, 600, "picker height before floors")
	d.Int("row_height", &dungeonPickerRowH, 12, 48, "floor row height")
	d.Int("padding", &dungeonPickerPad, 0, 32, "content padding")
	d.Int("max_rows", &dungeonPickerMax, 1, 24, "most floors shown before scrolling")
	d.Int("intro_height", &dungeonPickerIntroH, 12, 80, "intro text height")
	d.Int("gap", &dungeonPickerGap, 0, 24, "gap between the intro and the floor list")
}
