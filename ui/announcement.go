package ui

import (
	"image/color"
	"strings"
	"time"

	"github.com/kivutar/goro/render"
	"github.com/kivutar/goro/ui/rotheme"
)

// theme:ok announcement lifetime
const announcementLife = 20 * time.Second

type AnnouncementStyle struct {
	Y        int
	FontSize int
	Bold     bool
}

type Announcement struct {
	text     string
	color    color.RGBA
	shownAt  time.Time
	y        int
	fontSize int
	bold     bool
	life     time.Duration
}

// ShowFor is Show with a custom lifetime instead of the 20 s announcement.
func (a *Announcement) ShowFor(text string, messageColor color.RGBA, style AnnouncementStyle, now time.Time, life time.Duration) {
	a.Show(text, messageColor, style, now)
	a.life = life
}

func (a *Announcement) Show(text string, messageColor color.RGBA, style AnnouncementStyle, now time.Time) {
	text = strings.TrimSpace(text)
	if text == "" {
		return
	}
	if now.IsZero() {
		now = time.Now()
	}
	if messageColor.A == 0 {
		messageColor.A = 255
	}
	if style.Y <= 0 {
		style.Y = announcementTop
	}
	if style.FontSize <= 0 {
		style.FontSize = int(rotheme.Default.Typography.TextSize)
	} else {
		style.FontSize = maxInt(announcementFontMin, minInt(style.FontSize, announcementFontMax))
	}
	a.text = text
	a.color = messageColor
	a.shownAt = now
	a.y = style.Y
	a.fontSize = style.FontSize
	a.bold = style.Bold
	a.life = announcementLife
}

func (a *Announcement) Visible(now time.Time) bool {
	if a == nil || a.text == "" || a.shownAt.IsZero() {
		return false
	}
	if now.IsZero() {
		now = time.Now()
	}
	life := a.life
	if life <= 0 {
		life = announcementLife
	}
	return now.Sub(a.shownAt) < life
}

func (a *Announcement) Draw(screen *render.Frame, now time.Time) {
	if screen == nil || !a.Visible(now) {
		return
	}
	screenW := screen.Bounds().Dx()
	maxWidth := minInt(announcementMaxWidth+announcementPadding+announcementPadding, screenW)
	render.DrawUITextBanner(screen, a.text, float64(screenW)/2, float64(a.y), float64(maxWidth), a.color, float32(a.fontSize), a.bold)
}
