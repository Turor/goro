package rotheme

func init() {
	c := &Default.Colors
	RegisterColor("theme", "window_body", &c.WindowBody, "window background (alpha makes windows translucent)")
	RegisterColor("theme", "window_title_top", &c.WindowTitleTop, "title bar gradient, top")
	RegisterColor("theme", "window_title", &c.WindowTitle, "title bar gradient, bottom")
	RegisterColor("theme", "window_border", &c.WindowBorder, "window border and title bar underline")
	RegisterColor("theme", "window_footer", &c.WindowFooter, "window footer background")
	RegisterColor("theme", "footer_line", &c.FooterLine, "line above the footer; disabled button border")
	RegisterColor("theme", "panel_body", &c.PanelBody, "inset panel background")
	RegisterColor("theme", "button", &c.Button, "button and hotbar slot fill")
	RegisterColor("theme", "button_hover", &c.ButtonHover, "hovered hotbar slot fill")
	RegisterColor("theme", "button_down", &c.ButtonDown, "pressed button fill")
	RegisterColor("theme", "button_border", &c.ButtonBorder, "button and hotbar slot border")
	RegisterColor("theme", "disabled", &c.Disabled, "disabled button fill")
	RegisterColor("theme", "text", &c.Text, "body text")
	RegisterColor("theme", "title_text", &c.TitleText, "window title text")
	RegisterColor("theme", "label_text", &c.LabelText, "field labels")
	RegisterColor("theme", "muted_text", &c.MutedText, "secondary and disabled text")
	RegisterColor("theme", "input_border", &c.InputBorder, "text field border")
	RegisterColor("theme", "input_focus", &c.InputFocus, "focused text field border")
	RegisterColor("theme", "skill_required", &c.SkillRequired, "skill tree prerequisite highlight")
	RegisterColor("theme", "context_menu_hover", &contextMenuItemHover, "hovered context menu row")

	RegisterFloat("text", "size", &Default.Typography.TextSize, 6, 32, "UI text size in pixels")
	RegisterString("text", "font_family", &Default.Typography.FontFamily, 64, false, "UI font family: the built-in Inter, or the name to give font_file")
	RegisterString("text", "font_file", &FontFile, 256, true, "TTF or OTF file for font_family, relative to this file; empty uses the built-in font")

	RegisterFloat("button", "radius", &ButtonRadius, 0, 16, "button corner radius")
	RegisterFloat("button", "padding_x", &ButtonPaddingX, 0, 32, "button horizontal padding")
	RegisterFloat("button", "padding_y", &ButtonPaddingY, 0, 32, "button vertical padding")
	RegisterFloat("button", "large_padding_y", &LargeButtonPaddingY, 0, 32, "large button vertical padding")
	RegisterFloat("button", "reflect_alpha", &buttonReflectAlpha, 0, 1, "button glossy highlight opacity")

	RegisterFloat("progress_bar", "height", &ProgressBarHeight, 2, 40, "progress bar height")
}
