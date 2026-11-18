package battle

import (
	"image/color"

	"github.com/kivutar/goro/ui/rotheme"
)

// Bars drawn over actors in the world (HP, SP, hunger and cast bars). Values
// are read on every draw, so ui-layout.ini edits apply at once.
var (
	actorOverlayBarWidth      = 60.0
	actorOverlayBarRowHeight  = 3.0
	actorOverlayBarRowSpacing = 4.0
	actorOverlayBarFrame      = 1.0
	actorNameBelowBarGap      = 3.0
	actorBarLowRatio          = 0.25
	actorBarFrameColor        = color.RGBA{R: 16, G: 24, B: 156, A: 255}
	actorBarTrackColor        = color.RGBA{R: 66, G: 66, B: 66, A: 255}
	actorBarMonsterFill       = color.RGBA{R: 255, G: 0, B: 231, A: 255}
	actorBarMonsterLowFill    = color.RGBA{R: 255, G: 255, B: 0, A: 255}
	actorBarPlayerLowFill     = color.RGBA{R: 255, G: 0, B: 0, A: 255}
	actorCastBarFill          = color.RGBA{R: 0, G: 255, B: 0, A: 255}
)

// Damage numbers and the miss/critical sprites over actors.
var (
	DamageFloaterWhite          = color.RGBA{R: 255, G: 255, B: 255, A: 255}
	damageFloaterMiss           = color.RGBA{R: 255, G: 255, B: 255, A: 255}
	damageFloaterYellow         = color.RGBA{R: 230, G: 230, B: 38, A: 255}
	damageFloaterCombo          = color.RGBA{R: 230, G: 230, B: 38, A: 255}
	damageFloaterRed            = color.RGBA{R: 255, G: 64, B: 64, A: 255}
	damageFloaterHealHP         = color.RGBA{R: 0, G: 255, B: 0, A: 255}
	damageFloaterHealSP         = color.RGBA{R: 0, G: 0, B: 255, A: 255}
	damageCriticalBackdrop      = color.RGBA{R: 168, G: 168, B: 168, A: 255}
	damageNumberDigitGap        = 2
	damageNumberScale           = 1.0
	damageMissScale             = 0.5
	damageCriticalBackdropScale = 0.6
	damageFallbackTextDX        = -8
	damageFallbackTextDY        = -40
)

func init() {
	bars := rotheme.Section("actor_bars", "HP, SP, hunger and cast bars drawn over characters and monsters")
	bars.Float64("width", &actorOverlayBarWidth, 10, 400, "bar width")
	bars.Float64("row_height", &actorOverlayBarRowHeight, 1, 40, "height of one bar row")
	bars.Float64("row_spacing", &actorOverlayBarRowSpacing, 1, 40, "distance from one row (HP, SP, hunger) to the next")
	bars.Float64("frame", &actorOverlayBarFrame, 0, 8, "frame thickness around the bars")
	bars.Float64("name_gap", &actorNameBelowBarGap, 0, 40, "gap between the bars and a name drawn below them")
	bars.Float64("low_ratio", &actorBarLowRatio, 0, 1, "HP fraction below which the low colors are used")
	bars.RGBA("frame_color", &actorBarFrameColor, "bar frame and row separator")
	bars.RGBA("track", &actorBarTrackColor, "empty part of a bar")
	bars.RGBA("monster_fill", &actorBarMonsterFill, "monster HP")
	bars.RGBA("monster_low_fill", &actorBarMonsterLowFill, "monster HP when low")
	bars.RGBA("player_low_fill", &actorBarPlayerLowFill, "player and ally HP when low (normal fill is [theme] hp_bar)")
	bars.RGBA("cast_fill", &actorCastBarFill, "cast bar fill when the server sends no color")

	dmg := rotheme.Section("damage_numbers", "damage, heal and miss numbers over actors")
	dmg.RGBA("normal", &DamageFloaterWhite, "damage dealt")
	dmg.RGBA("miss", &damageFloaterMiss, "miss")
	dmg.RGBA("critical", &damageFloaterYellow, "critical hit")
	dmg.RGBA("combo", &damageFloaterCombo, "multi-hit combo total")
	dmg.RGBA("incoming", &damageFloaterRed, "damage taken by you")
	dmg.RGBA("heal_hp", &damageFloaterHealHP, "HP recovery")
	dmg.RGBA("heal_sp", &damageFloaterHealSP, "SP recovery")
	dmg.RGBA("critical_backdrop", &damageCriticalBackdrop, "burst sprite behind a critical hit")
	dmg.Int("digit_gap", &damageNumberDigitGap, 0, 16, "gap between digits")
	dmg.Float64("scale", &damageNumberScale, 0.1, 4, "size multiplier for every number")
	dmg.Float64("miss_scale", &damageMissScale, 0.05, 4, "size of the miss sprite")
	dmg.Float64("critical_backdrop_scale", &damageCriticalBackdropScale, 0.05, 4, "size of the critical burst relative to the number")
	dmg.Int("fallback_text_dx", &damageFallbackTextDX, -200, 200, "text offset when the digit sprite is missing")
	dmg.Int("fallback_text_dy", &damageFallbackTextDY, -200, 200, "vertical text offset when the digit sprite is missing")
}
