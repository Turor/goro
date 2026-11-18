package rotheme

// [dropdown] defaults: closed trigger and open menu.
var (
	dropdownTriggerPadLeft  float32 = 6
	dropdownTriggerPadRight float32 = 18
	dropdownItemPadX        float32 = 6
	dropdownArrowInset      float32 = 10
	dropdownArrowHalfW      float32 = 4
	dropdownArrowHalfH      float32 = 2
)

func init() {
	s := Section("dropdown", "dropdown trigger and menu")
	s.Float("trigger_pad_left", &dropdownTriggerPadLeft, 0, 24, "trigger text inset from the left")
	s.Float("trigger_pad_right", &dropdownTriggerPadRight, 0, 72, "trigger text inset from the right, leaving room for the arrow")
	s.Float("item_pad_x", &dropdownItemPadX, 0, 24, "menu item text horizontal inset")
	s.Float("arrow_inset", &dropdownArrowInset, 0, 40, "arrow center inset from the right edge")
	s.Float("arrow_half_width", &dropdownArrowHalfW, 0, 16, "arrow half width")
	s.Float("arrow_half_height", &dropdownArrowHalfH, 0, 8, "arrow half height")
}
