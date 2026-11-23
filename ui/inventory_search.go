package ui

import (
	"github.com/gogpu/ui/core/textfield"
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
	"github.com/kivutar/goro/ui/rotheme"
)

// inventorySearchBox is the opinionated search field. The clear mark is drawn
// on top of the field so showing it does not change the row size or move the
// tiles under it.
type inventorySearchBox struct {
	widget.WidgetBase
	window *InventoryBagWindow
	field  *textfield.Widget
}

func newInventorySearchBox(window *InventoryBagWindow) *inventorySearchBox {
	box := &inventorySearchBox{window: window}
	if window != nil {
		box.field = window.ensureSearchField()
	}
	box.SetVisible(true)
	box.SetEnabled(true)
	return box
}

func (w *InventoryBagWindow) ensureSearchField() *textfield.Widget {
	if w == nil {
		return nil
	}
	if w.searchField != nil {
		return w.searchField
	}
	field := rotheme.TextField(w.search, textfield.TypeText, func(value string) {
		w.applySearch(w.searchCtx, value)
	}, func(string) {})
	field.Padding(float32(inventoryBagSearchClear))
	w.searchField = field
	return field
}

func (w *InventoryBagWindow) applySearch(ctx Context, value string) {
	if w == nil {
		return
	}
	w.search = value
	w.syncSearchGrids(ctx)
	if ctx.Session != nil {
		w.snapshot = w.inventorySnapshot(ctx.Session)
	}
}

func (w *InventoryBagWindow) syncSearchGrids(ctx Context) {
	if w == nil {
		return
	}
	if w.bagGrid != nil {
		items := w.tabItems(ctx.Session)
		if w.opinionatedSplit() {
			items = w.bagItems(ctx.Session)
		}
		w.bagGrid.setItems(items, w.itemIcons(ctx, items), inventoryGridAmountLabels(items))
	}
	if w.sideGrid != nil && w.opinionatedSplit() {
		items := w.tabItems(ctx.Session)
		w.sideGrid.setItems(items, w.itemIcons(ctx, items), inventoryGridAmountLabels(items))
	}
}

func (b *inventorySearchBox) Layout(ctx widget.Context, constraints geometry.Constraints) geometry.Size {
	scale := rotheme.ContentScale()
	if scale <= 0 {
		scale = 1
	}
	width := constraints.MaxWidth
	if !opinionatedWidthBounded(width) {
		width = float32(opinionatedFrameW()) * scale
	}
	height := float32(inventoryBagSearchH) * scale
	size := constraints.Constrain(geometry.Sz(width, height))
	if b.field != nil {
		b.field.Layout(ctx, geometry.Tight(size))
		var origin geometry.Point
		b.field.SetBounds(geometry.FromPointSize(origin, size))
	}
	b.SetBounds(geometry.FromPointSize(b.Position(), size))
	return size
}

func (b *inventorySearchBox) Draw(ctx widget.Context, canvas widget.Canvas) {
	if !b.IsVisible() {
		return
	}
	bounds := b.Bounds()
	canvas.PushTransform(bounds.Min)
	if b.field != nil {
		widget.DrawChild(b.field, ctx, canvas)
	}
	if b.window != nil && b.window.search != "" {
		rotheme.DrawText(canvas, "x", b.clearRect(), rotheme.Default.Typography.TextSize, rotheme.Default.Colors.Text, false, widget.TextAlignCenter)
	}
	canvas.PopTransform()
}

func (b *inventorySearchBox) clearRect() geometry.Rect {
	scale := rotheme.ContentScale()
	if scale <= 0 {
		scale = 1
	}
	size := b.Bounds().Size()
	clearW := float32(inventoryBagSearchClear) * scale
	if clearW > size.Width {
		clearW = size.Width
	}
	var origin geometry.Point
	origin.X = size.Width - clearW
	return geometry.FromPointSize(origin, geometry.Sz(clearW, size.Height))
}

func (b *inventorySearchBox) Event(ctx widget.Context, e event.Event) bool {
	if b.field == nil || !b.IsVisible() || !b.IsEnabled() {
		return false
	}
	if mouse, ok := e.(*event.MouseEvent); ok && b.window != nil && b.window.search != "" && mouse.MouseType == event.MousePress && mouse.Button == event.ButtonLeft && b.clearRect().Contains(mouse.Position) {
		b.field.SetText("")
		b.window.applySearch(b.window.searchCtx, "")
		b.SetNeedsRedraw(true)
		return true
	}
	return b.field.Event(ctx, e)
}

func (b *inventorySearchBox) Children() []widget.Widget {
	if b.field == nil {
		return nil
	}
	return []widget.Widget{b.field}
}
