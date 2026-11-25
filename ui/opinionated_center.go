package ui

import (
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"

	"github.com/kivutar/goro/ui/rotheme"
)

// The 7th column is one cell. The divider is a thin line in that column.
// Tabs and search stay in the rows above so they are not crushed into the tile.
type opinionatedCenterColumn struct {
	widget.WidgetBase
	scale float32
}

func newOpinionatedDividerColumn() *opinionatedCenterColumn {
	col := &opinionatedCenterColumn{}
	col.SetVisible(true)
	col.SetEnabled(true)
	return col
}

func (c *opinionatedCenterColumn) Layout(_ widget.Context, constraints geometry.Constraints) geometry.Size {
	scale := rotheme.ContentScale()
	if scale <= 0 {
		scale = 1
	}
	c.scale = scale
	size := constraints.Constrain(geometry.Sz(float32(opinionatedMiddleW())*scale, float32(inventoryBagViewH)*scale))
	c.SetBounds(geometry.FromPointSize(c.Position(), size))
	return size
}

func (c *opinionatedCenterColumn) Draw(_ widget.Context, canvas widget.Canvas) {
	if !c.IsVisible() {
		return
	}
	bounds := c.Bounds()
	line := float32(opinionatedDividerW()) * c.scale
	if line <= 0 {
		line = c.scale
	}
	if line > bounds.Width() {
		line = bounds.Width()
	}
	x := (bounds.Width() - line) / 2
	canvas.DrawRect(geometry.NewRect(bounds.Min.X+x, bounds.Min.Y, line, bounds.Height()), rotheme.Default.Colors.WindowBorder)
}

func (c *opinionatedCenterColumn) Event(widget.Context, event.Event) bool { return false }

func (c *opinionatedCenterColumn) Children() []widget.Widget { return nil }
