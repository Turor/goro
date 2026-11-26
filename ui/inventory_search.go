package ui

import (
	"github.com/gogpu/ui/core/textfield"
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
	"github.com/kivutar/goro/ui/rotheme"
)

// inventorySearchBox draws the opinionated search field itself. The text
// widget keeps the keystrokes, but its own paint uses bounds that the window
// clip throws away, so the letters are drawn here.
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
	height := float32(inventorySearchRowH()) * scale
	size := constraints.Constrain(geometry.Sz(width, height))
	b.SetBounds(geometry.FromPointSize(b.Position(), size))
	if b.field != nil {
		b.field.Layout(ctx, geometry.Tight(size))
		b.field.SetBounds(b.Bounds())
	}
	return size
}

func (b *inventorySearchBox) Draw(ctx widget.Context, canvas widget.Canvas) {
	if !b.IsVisible() {
		return
	}
	field := b.fieldRect()
	border := rotheme.Default.Colors.FooterLine
	if b.window.searchArmed || (b.field != nil && b.field.IsFocused()) {
		border = rotheme.Default.Colors.InputFocus
	}
	canvas.DrawRect(field, rotheme.Default.Colors.PanelBody)
	canvas.StrokeRect(field, border, windowBorderWidth)
	if b.window == nil {
		return
	}
	if b.window.search != "" {
		rotheme.DrawText(canvas, b.window.search, b.textRect(), rotheme.Default.Typography.TextSize, rotheme.Default.Colors.Text, false, widget.TextAlignLeft)
		rotheme.DrawText(canvas, "x", b.clearRect(), rotheme.Default.Typography.TextSize, rotheme.Default.Colors.Text, false, widget.TextAlignCenter)
	}
	if b.window.searchArmed || (b.field != nil && b.field.IsFocused()) {
		text := b.textRect()
		caretX := text.Min.X
		if b.window.search != "" {
			caretX += rotheme.MeasureText(canvas, b.window.search, rotheme.Default.Typography.TextSize, false)
		}
		if caretX > text.Max.X {
			caretX = text.Max.X
		}
		canvas.DrawLine(geometry.Pt(caretX, text.Min.Y), geometry.Pt(caretX, text.Max.Y), rotheme.Default.Colors.Text, windowBorderWidth)
	}
}

func (b *inventorySearchBox) scale() float32 {
	scale := rotheme.ContentScale()
	if scale <= 0 {
		return 1
	}
	return scale
}

func (b *inventorySearchBox) fieldRect() geometry.Rect {
	bounds := b.Bounds()
	inset := float32(inventoryBagSearchInset) * b.scale()
	rect := bounds
	rect.Min.X += inset
	rect.Min.Y += inset
	rect.Max.X -= inset
	rect.Max.Y -= inset
	if rect.Max.X < rect.Min.X {
		rect.Max.X = rect.Min.X
	}
	if rect.Max.Y < rect.Min.Y {
		rect.Max.Y = rect.Min.Y
	}
	return rect
}

func (b *inventorySearchBox) textRect() geometry.Rect {
	field := b.fieldRect()
	pad := float32(inventoryBagIconPad) * b.scale()
	clearW := float32(inventoryBagSearchClear) * b.scale()
	rect := field
	rect.Min.X += pad
	rect.Max.X -= clearW
	if rect.Max.X < rect.Min.X {
		rect.Max.X = rect.Min.X
	}
	return rect
}

func (b *inventorySearchBox) clearRect() geometry.Rect {
	field := b.fieldRect()
	clearW := float32(inventoryBagSearchClear) * b.scale()
	if clearW > field.Width() {
		clearW = field.Width()
	}
	rect := field
	rect.Min.X = field.Max.X - clearW
	return rect
}

func (b *inventorySearchBox) Event(ctx widget.Context, e event.Event) bool {
	if b.field == nil || !b.IsVisible() || !b.IsEnabled() {
		return false
	}
	mouse, ok := e.(*event.MouseEvent)
	if !ok {
		return b.field.Event(ctx, e)
	}
	local := mouse.Position
	if b.Bounds().Min.X != 0 || b.Bounds().Min.Y != 0 {
		local = mouse.Position.Sub(b.Bounds().Min)
	}
	if b.window != nil && b.window.search != "" && mouse.MouseType == event.MousePress && mouse.Button == event.ButtonLeft && (b.clearRect().Contains(mouse.Position) || b.clearRect().Contains(local) || b.clearRect().Contains(mouse.Position.Add(b.Bounds().Min))) {
		b.field.SetText("")
		b.window.applySearch(b.window.searchCtx, "")
		b.SetNeedsRedraw(true)
		return true
	}
	if mouse.MouseType == event.MousePress && mouse.Button == event.ButtonLeft {
		b.window.searchArmed = true
		ctx.RequestFocus(b.field)
		b.field.SetFocused(true)
	}
	return b.field.Event(ctx, e)
}

func (b *inventorySearchBox) Children() []widget.Widget {
	if b.field == nil {
		return nil
	}
	return []widget.Widget{b.field}
}
