package rotheme

import (
	"image/color"
)

// Map loading screen progress bar (drawn by game/map). Colors fade with the screen: their alpha is
// multiplied by the fade alpha.
var (
	LoadingBarMaxWidth   = 240.0
	LoadingBarSideMargin = 16.0
	LoadingBarHeight     = 15.0
	LoadingBarYRatio     = 0.75
	LoadingBarBorder     = 1.0
	LoadingBarPadding    = 2.0
	LoadingLabelGap      = 4.0
	LoadingLabelSize     = float32(11)

	LoadingBackground     = color.RGBA{A: 255}
	LoadingBarBorderColor = color.RGBA{G: 255, B: 255, A: 255}
	LoadingBarTrack       = color.RGBA{R: 140, G: 140, B: 140, A: 255}
	LoadingBarFill        = color.RGBA{R: 66, G: 99, B: 165, A: 255}
	LoadingLabelColor     = color.RGBA{R: 255, G: 255, A: 255}
)

func init() {
	s := Section("loading", "map loading screen progress bar")
	s.Float64("bar_max_width", &LoadingBarMaxWidth, 40, 1200, "progress bar width on a wide screen")
	s.Float64("bar_side_margin", &LoadingBarSideMargin, 0, 200, "minimum gap between the bar and the screen sides")
	s.Float64("bar_height", &LoadingBarHeight, 4, 80, "progress bar height")
	s.Float64("bar_y_ratio", &LoadingBarYRatio, 0, 1, "bar position as a fraction of the screen height")
	s.Float64("bar_border", &LoadingBarBorder, 0, 8, "progress bar border thickness")
	s.Float64("bar_padding", &LoadingBarPadding, 0, 16, "inset of the fill inside the bar")
	s.Float64("label_gap", &LoadingLabelGap, 0, 40, "gap between the percent label and the bar")
	s.Float("label_size", &LoadingLabelSize, 6, 48, "percent label text size")
	s.RGBA("background", &LoadingBackground, "backdrop behind the loading image")
	s.RGBA("bar_border_color", &LoadingBarBorderColor, "progress bar border")
	s.RGBA("bar_track", &LoadingBarTrack, "progress bar empty track")
	s.RGBA("bar_fill", &LoadingBarFill, "progress bar fill")
	s.RGBA("label_color", &LoadingLabelColor, "percent label")
}
