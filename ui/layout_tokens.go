package ui

import "github.com/kivutar/goro/ui/rotheme"

// Layout tokens owned by the ui package. They are read when a widget is laid
// out or built, so a reloaded ui-layout.ini reaches open windows on their next
// layout (see Window.refreshLayoutTokens) and every window opened afterwards.
func init() {
	rotheme.RegisterFloat("window", "radius", &WindowRadius, 0, 24, "window corner radius")
	rotheme.RegisterFloat("window", "border_width", &windowBorderWidth, 0, 6, "window border width")
	rotheme.RegisterFloat("window", "title_padding_left", &windowTitlePadL, 0, 64, "space left of the window title")
	rotheme.RegisterInt("window", "title_padding_right", &windowTitleButtonPadR, 0, 64, "space right of the title bar buttons")
	rotheme.RegisterInt("window", "title_button_gap", &windowTitleButtonGap, 0, 32, "gap between title bar buttons")
	rotheme.RegisterInt("window", "title_height", &windowTitleHeight, 16, 64, "title bar height; windows grow or shrink to keep their content area")
	rotheme.RegisterInt("window", "footer_height", &windowFooterHeight, 24, 96, "footer height; windows grow or shrink to keep their content area")
	rotheme.RegisterInt("window", "footer_padding", &ROWindowFooterPadding, 0, 64, "footer left and right padding")
	rotheme.RegisterInt("window", "footer_gap", &ROWindowFooterGap, 0, 64, "gap between footer buttons")

	rotheme.RegisterInt("hotbar", "slot_size", &shortcutSlot, 16, 96, "hotbar slot size")
	rotheme.RegisterInt("hotbar", "slot_gap", &shortcutGap, 0, 32, "gap between hotbar slots")
	rotheme.RegisterInt("hotbar", "row_gap", &shortcutRowGap, 0, 32, "gap between hotbar rows")
	rotheme.RegisterInt("hotbar", "padding", &shortcutPad, 0, 32, "padding around the hotbar grid")

	rotheme.RegisterRGBA("theme", "hp_bar", &PlayerHPBarColor, "player HP bar")
	rotheme.RegisterRGBA("theme", "sp_bar", &PlayerSPBarColor, "player SP bar")
	rotheme.RegisterRGBA("theme", "hunger_bar", &HungerBarColor, "pet and homunculus hunger bar")
	rotheme.RegisterRGBA("theme", "hunger_bar_low", &HungerBarLowColor, "hunger bar when low")
	rotheme.RegisterRGBA("theme", "good_text", &GoodTextColor, "positive values and success messages")
	rotheme.RegisterRGBA("theme", "error_text", &ErrorTextColor, "errors and negative values")
	rotheme.RegisterRGBA("theme", "panel_alt", &PanelAltColor, "alternate panel rows")
}
