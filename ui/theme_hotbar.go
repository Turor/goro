package ui

import "github.com/kivutar/goro/ui/rotheme"

// [hotbar] defaults. slot_size, slot_gap, row_gap and padding are already
// registered in layout_tokens.go; this file adds the rest of the bar.
var (
	shortcutSlot                    = 34
	shortcutGap                     = 2
	shortcutRowGap                  = 2
	shortcutPad                     = 3
	shortcutCols                    = 9
	shortcutMinRows                 = 1
	shortcutControlGap              = 0
	shortcutControlPad              = 2
	shortcutGripWidth               = 10
	shortcutGripMarkW               = 2
	shortcutGripMarkH               = 24
	shortcutGripMarkRadius  float32 = 1
	shortcutIconInset               = 5
	shortcutLabelH                  = 12
	shortcutAmountBottom            = 15
	shortcutAmountInset             = 3
	shortcutSkillLabelX             = 3
	shortcutSkillLabelInset         = 6
	shortcutTooltipGap              = 2
)

func init() {
	rotheme.OnLayoutApplied(recomputeHotbarRows)
	s := rotheme.Section("hotbar", "skill and item hotbar")
	s.Int("columns", &shortcutCols, 1, 27, "hotbar columns")
	s.Int("min_rows", &shortcutMinRows, 1, 9, "minimum visible hotbar rows")
	s.Int("control_gap", &shortcutControlGap, 0, 32, "gap before the row buttons")
	s.Int("control_padding", &shortcutControlPad, 0, 32, "padding after the row buttons")
	s.Int("grip_width", &shortcutGripWidth, 0, 48, "drag grip width")
	s.Int("grip_mark_width", &shortcutGripMarkW, 0, 24, "grip mark width")
	s.Int("grip_mark_height", &shortcutGripMarkH, 4, 96, "grip mark height")
	s.Float("grip_mark_radius", &shortcutGripMarkRadius, 0, 12, "grip mark corner radius")
	s.Int("icon_inset", &shortcutIconInset, 0, 32, "slot icon inset")
	s.Int("label_height", &shortcutLabelH, 6, 40, "slot label height")
	s.Int("amount_bottom", &shortcutAmountBottom, 0, 48, "item amount label offset from the slot bottom")
	s.Int("amount_inset", &shortcutAmountInset, 0, 24, "item amount label inset from the slot edges")
	s.Int("skill_label_x", &shortcutSkillLabelX, 0, 32, "skill level label inset from the slot left")
	s.Int("skill_label_inset", &shortcutSkillLabelInset, 0, 32, "skill level label inset from the slot edges")
	s.Int("tooltip_gap", &shortcutTooltipGap, 0, 32, "gap between the hotbar and its tooltip")
}
