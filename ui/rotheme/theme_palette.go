package rotheme

import "image/color"

// The default palette: the built-in light blue Ragnarok theme. Every color
// setting in ui-layout.ini defaults to a reference to one of these (see
// palette.go and palette_refs.go), so changing one here or in [palette]
// re-colors everything derived from it.
var (
	paletteAccent     = color.RGBA{R: 118, G: 160, B: 206, A: 255}
	paletteBackground = color.RGBA{R: 255, G: 255, B: 255, A: 255}
	palettePanel      = color.RGBA{R: 250, G: 252, B: 255, A: 255}
	palettePanelAlt   = color.RGBA{R: 236, G: 244, B: 252, A: 255}
	paletteText       = color.RGBA{R: 38, G: 48, B: 58, A: 255}
	paletteTextDim    = color.RGBA{R: 98, G: 112, B: 126, A: 255}
	paletteBorder     = color.RGBA{R: 138, G: 174, B: 214, A: 255}
	paletteHighlight  = color.RGBA{R: 255, G: 255, B: 0, A: 255}
	paletteGood       = color.RGBA{R: 34, G: 142, B: 72, A: 255}
	paletteBad        = color.RGBA{R: 204, G: 48, B: 48, A: 255}
	paletteWarning    = color.RGBA{R: 255, G: 140, B: 26, A: 255}
	paletteHP         = color.RGBA{R: 16, G: 239, B: 33, A: 255}
	paletteSP         = color.RGBA{R: 24, G: 99, B: 222, A: 255}
	paletteEXP        = color.RGBA{R: 118, G: 160, B: 206, A: 255}
	paletteWorldText  = color.RGBA{R: 255, G: 255, B: 255, A: 255}
	paletteShadow     = color.RGBA{R: 0, G: 0, B: 0, A: 255}
)

func init() {
	s := Section(paletteSection, "base colors of the theme: every color setting below defaults to @name of one of these, optionally with lighten/darken/saturate/desaturate(N%), spin(deg) or alpha(0-1)")
	s.RGBA("accent", &paletteAccent, "accent: window borders, title bars, selection")
	s.RGBA("background", &paletteBackground, "window background")
	s.RGBA("panel", &palettePanel, "inset panels and light fills")
	s.RGBA("panel_alt", &palettePanelAlt, "buttons, alternate rows and secondary fills")
	s.RGBA("text", &paletteText, "body text on windows")
	s.RGBA("text_dim", &paletteTextDim, "secondary, muted and disabled text")
	s.RGBA("border", &paletteBorder, "button, field and control borders")
	s.RGBA("highlight", &paletteHighlight, "highlights and attention: admin names, critical hits, map labels")
	s.RGBA("good", &paletteGood, "positive values, success, healing")
	s.RGBA("bad", &paletteBad, "errors, damage taken, danger")
	s.RGBA("warning", &paletteWarning, "warnings, guild markers, low values")
	s.RGBA("hp", &paletteHP, "HP bars")
	s.RGBA("sp", &paletteSP, "SP bars")
	s.RGBA("exp", &paletteEXP, "experience bars")
	s.RGBA("world_text", &paletteWorldText, "text drawn over the game world: names, damage numbers, dark panels")
	s.RGBA("shadow", &paletteShadow, "outlines, shadows and dark translucent backdrops")
}
