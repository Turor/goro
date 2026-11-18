package ui

import (
	"fmt"
	"image"

	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
	"github.com/kivutar/goro/ui/rotheme"
)

type PvPCounter struct {
	widget    *pvpCounterWidget
	assets    pvpRankSpriteSet
	root      widget.Widget
	visible   bool
	x         int
	y         int
	rank      int
	total     int
	layoutGen uint64
}

func (c *PvPCounter) Update(ctx Context) {
	if ctx.UIManager == nil || ctx.World == nil ||
		!ctx.World.MapProperty.PvPRankingEnabled() ||
		!ctx.World.Player.HasPvPRanking ||
		ctx.World.Player.PvPRank <= 0 || ctx.World.Player.PvPTotal <= 0 {
		c.Unpublish(ctx)
		return
	}

	if c.widget == nil {
		c.widget = newPvPCounterWidget()
	}
	layoutChanged := false
	if gen := rotheme.Generation(); gen != c.layoutGen {
		c.layoutGen = gen
		layoutChanged = true
	}
	width, height := ctx.ScreenSize()
	x, y := pvpCounterBounds(width, height)
	positionChanged := c.root == nil || c.x != x || c.y != y
	valueChanged := c.rank != ctx.World.Player.PvPRank || c.total != ctx.World.Player.PvPTotal
	c.rank = ctx.World.Player.PvPRank
	c.total = ctx.World.Player.PvPTotal
	if valueChanged || c.widget.text == "" {
		c.widget.text = fmt.Sprintf("%d/%d", c.rank, c.total)
	}
	rankImage := c.assets.image(ctx.Resources, c.rank, c.total)
	imageChanged := c.widget.image != rankImage
	c.widget.image = rankImage

	if positionChanged {
		c.Unpublish(ctx)
		c.x, c.y = x, y
		c.root = positionedWidget(c.widget, x, y, pvpCounterWidth, pvpCounterHeight)
		if enabled, ok := c.root.(interface{ SetEnabled(bool) }); ok {
			enabled.SetEnabled(false)
		}
		ctx.UIManager.AddOverlay(c.root)
		c.visible = true
		valueChanged = true
	}
	if valueChanged || imageChanged || layoutChanged {
		c.widget.SetNeedsRedraw(true)
		if redraw, ok := c.root.(interface{ SetNeedsRedraw(bool) }); ok {
			redraw.SetNeedsRedraw(true)
		}
	}
}

func (c *PvPCounter) Unpublish(ctx Context) {
	if !c.visible || c.root == nil || ctx.UIManager == nil {
		return
	}
	ctx.UIManager.RemoveOverlay(c.root)
	c.root = nil
	c.visible = false
}

func pvpCounterBounds(width, height int) (int, int) {
	x := width - pvpCounterWidth
	y := height - pvpCounterHeight
	if x < 0 {
		x = 0
	}
	if y < 0 {
		y = 0
	}
	return x, y
}

type pvpCounterWidget struct {
	widget.WidgetBase
	text  string
	image image.Image
}

func newPvPCounterWidget() *pvpCounterWidget {
	w := &pvpCounterWidget{}
	w.SetVisible(true)
	w.SetEnabled(false)
	return w
}

func (w *pvpCounterWidget) Layout(_ widget.Context, constraints geometry.Constraints) geometry.Size {
	size := constraints.BiggestFinite(float32(pvpCounterWidth), float32(pvpCounterHeight))
	w.SetBounds(geometry.FromPointSize(w.Position(), size))
	return size
}

func (w *pvpCounterWidget) Draw(_ widget.Context, canvas widget.Canvas) {
	if w.image != nil {
		canvas.DrawImage(w.image, w.Bounds().Min)
		return
	}
	if w.text == "" {
		return
	}
	bounds := w.Bounds()
	shadow := bounds.Translate(geometry.Pt(pvpCounterShadowDX, pvpCounterShadowDY))
	rotheme.DrawText(canvas, w.text, shadow, pvpCounterFont, pvpCounterShadow, true, widget.TextAlignRight)
	rotheme.DrawText(canvas, w.text, bounds, pvpCounterFont, pvpCounterText, true, widget.TextAlignRight)
}

func (w *pvpCounterWidget) Event(_ widget.Context, _ event.Event) bool {
	return false
}
