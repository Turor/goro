package ui

import (
	"strings"
	"time"

	"github.com/kivutar/goro/render"
)

// theme:ok poptip show and fade durations
const (
	poptipSolid = 3 * time.Second
	poptipFade  = time.Second
)

type poptipItem struct {
	text    string
	shownAt time.Time
}

// Poptips is the short, server-driven message stack shown near the top of the
// game view. It is deliberately input-transparent.
type Poptips struct {
	items []poptipItem
}

func (p *Poptips) Clear() {
	if p != nil {
		p.items = nil
	}
}

func (p *Poptips) Show(text string, now time.Time) {
	text = strings.TrimSpace(text)
	if text == "" {
		return
	}
	if now.IsZero() {
		now = time.Now()
	}
	items := p.items[:0]
	for _, item := range p.items {
		if item.text != text && now.Sub(item.shownAt) < poptipSolid+poptipFade {
			items = append(items, item)
		}
	}
	p.items = append([]poptipItem{{text: text, shownAt: now}}, items...)
	if len(p.items) > poptipMaxItems {
		p.items = p.items[:poptipMaxItems]
	}
}

func (p *Poptips) Draw(screen *render.Frame, now time.Time) {
	if screen == nil || p == nil {
		return
	}
	if now.IsZero() {
		now = time.Now()
	}
	for index, item := range p.items {
		alpha := poptipAlpha(now.Sub(item.shownAt))
		if alpha == 0 {
			continue
		}
		textW, textH := render.BitmapTextSize(item.text)
		boxW := textW + poptipPadding + poptipPadding
		boxH := textH + poptipPadding + poptipPadding
		x := (screen.Bounds().Dx() - boxW) / 2
		y := poptipBaseY - index*poptipLineH - boxH/2
		background := poptipBackground
		background.A = uint8(float64(poptipBackground.A) * alpha)
		foreground := poptipText
		foreground.A = uint8(float64(poptipText.A) * alpha)
		render.DrawRect(screen, float64(x), float64(y), float64(boxW), float64(boxH), background)
		render.DrawCenteredUITextAt(screen, item.text, float64(screen.Bounds().Dx())/2, float64(y+poptipPadding-poptipTextNudge), foreground)
	}
}

func poptipAlpha(age time.Duration) float64 {
	if age < 0 || age >= poptipSolid+poptipFade {
		return 0
	}
	if age <= poptipSolid {
		return 1
	}
	return 1 - float64(age-poptipSolid)/float64(poptipFade)
}
