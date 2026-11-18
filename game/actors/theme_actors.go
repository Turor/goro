package actors

import (
	"image/color"

	"github.com/kivutar/goro/ui/rotheme"
)

// Names drawn under actors in the world, and where the bars above go.
var (
	actorNameAdminColor   = color.RGBA{R: 255, G: 255, B: 0, A: 255}
	actorNamePlayerColor  = color.RGBA{R: 255, G: 255, B: 255, A: 255}
	actorNameNPCColor     = color.RGBA{R: 148, G: 189, B: 247, A: 255}
	actorNameMonsterColor = color.RGBA{R: 255, G: 198, B: 198, A: 255}
	actorNameOtherColor   = color.RGBA{R: 248, G: 248, B: 248, A: 255}
	actorNameOutlineColor = color.RGBA{A: 196}
	actorNameLabelOffset  = 13.0
	actorLifeBarGap       = 14.0
	actorCastBarGap       = 4.0
)

// Chat room and vending shop boards over characters.
var (
	boardLabelW               = 158
	boardLabelH               = 36
	boardLabelPadX            = 5
	boardLabelIconGap         = 5
	boardLabelGap             = 4
	boardLabelIcon            = 24
	boardLabelTextH           = 20
	boardLabelTextNudge       = 2
	boardLabelRadius          = 4
	boardLabelOutlineW        = 3
	boardLabelBorderW         = 1
	boardLabelEllipsisReserve = 14
	boardLabelFillColor       = color.RGBA{R: 255, G: 255, B: 255, A: 245}
	boardLabelOutlineColor    = color.RGBA{R: 255, G: 255, B: 255, A: 245}
	boardLabelBorderColor     = color.RGBA{R: 74, G: 138, B: 202, A: 245}
	boardLabelTextColor       = color.RGBA{R: 30, G: 34, B: 40, A: 255}
)

// Chat speech bubbles over actors.
var (
	speechBubbleMaxWidth = 220.0
	actorSpeechBubbleGap = 6.0
)

func init() {
	names := rotheme.Section("actor_names", "names under characters, NPCs and monsters in the world")
	names.RGBA("admin", &actorNameAdminColor, "GM and admin names")
	names.RGBA("player", &actorNamePlayerColor, "your own name")
	names.RGBA("npc", &actorNameNPCColor, "NPC names")
	names.RGBA("monster", &actorNameMonsterColor, "monster names")
	names.RGBA("other", &actorNameOtherColor, "other players and everything else")
	names.RGBA("outline", &actorNameOutlineColor, "text outline")
	names.Float64("label_offset", &actorNameLabelOffset, -100, 200, "name distance below the actor's feet (times the actor scale)")
	names.Float64("life_bar_gap", &actorLifeBarGap, -100, 200, "HP bar distance below the name")
	names.Float64("cast_bar_gap", &actorCastBarGap, -100, 200, "cast bar distance below the top of the sprite")

	boards := rotheme.Section("boards", "chat room and vending shop boards over characters")
	boards.Int("width", &boardLabelW, 40, 600, "board width")
	boards.Int("height", &boardLabelH, 16, 200, "board height")
	boards.Int("padding_x", &boardLabelPadX, 0, 60, "horizontal padding inside the frame")
	boards.Int("icon_gap", &boardLabelIconGap, 0, 60, "gap between the icon and the text")
	boards.Int("gap", &boardLabelGap, 0, 100, "gap above the sprite and between stacked boards")
	boards.Int("icon_size", &boardLabelIcon, 0, 128, "icon size")
	boards.Int("text_height", &boardLabelTextH, 4, 100, "text line height")
	boards.Int("text_nudge_y", &boardLabelTextNudge, -40, 40, "moves the text up by this many pixels")
	boards.Int("radius", &boardLabelRadius, 0, 40, "corner radius")
	boards.Int("outline_width", &boardLabelOutlineW, 0, 20, "outer outline width")
	boards.Int("border_width", &boardLabelBorderW, 0, 20, "inner border width")
	boards.Int("ellipsis_reserve", &boardLabelEllipsisReserve, 0, 100, "width kept free for ... on a long title")
	boards.RGBA("fill", &boardLabelFillColor, "board background")
	boards.RGBA("outline", &boardLabelOutlineColor, "outer outline")
	boards.RGBA("border", &boardLabelBorderColor, "inner border")
	boards.RGBA("text", &boardLabelTextColor, "title text")

	speech := rotheme.Section("speech_bubble", "chat bubbles over actors")
	speech.Float64("max_width", &speechBubbleMaxWidth, 40, 800, "widest bubble before text wraps")
	speech.Float64("gap", &actorSpeechBubbleGap, -100, 200, "gap between the bubble and the top of the sprite")
}
