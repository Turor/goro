package ui

import (
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
	"github.com/kivutar/goro/input"
	"github.com/kivutar/goro/ui/rotheme"
)

type opinionatedCloseBox struct {
	widget.WidgetBase
	window *InventoryBagWindow
	ctx    Context
}

func newOpinionatedCloseBox(window *InventoryBagWindow, ctx Context) *opinionatedCloseBox {
	box := &opinionatedCloseBox{window: window, ctx: ctx}
	box.SetVisible(true)
	box.SetEnabled(true)
	return box
}

func (b *opinionatedCloseBox) Layout(_ widget.Context, constraints geometry.Constraints) geometry.Size {
	scale := rotheme.ContentScale()
	if scale <= 0 {
		scale = 1
	}
	side := float32(inventoryBagFilterH) * scale
	size := constraints.Constrain(geometry.Sz(side, side))
	b.SetBounds(geometry.FromPointSize(b.Position(), size))
	return size
}

func (b *opinionatedCloseBox) Draw(_ widget.Context, canvas widget.Canvas) {
	if !b.IsVisible() {
		return
	}
	bounds := b.Bounds()
	canvas.DrawRect(bounds, rotheme.Default.Colors.PanelBody)
	canvas.StrokeRect(bounds, rotheme.Default.Colors.WindowBorder, windowBorderWidth)
	inset := windowBorderWidth
	canvas.DrawLine(
		geometry.Pt(bounds.Min.X+inset, bounds.Min.Y+inset),
		geometry.Pt(bounds.Max.X-inset, bounds.Max.Y-inset),
		rotheme.Default.Colors.Text,
		windowBorderWidth,
	)
	canvas.DrawLine(
		geometry.Pt(bounds.Max.X-inset, bounds.Min.Y+inset),
		geometry.Pt(bounds.Min.X+inset, bounds.Max.Y-inset),
		rotheme.Default.Colors.Text,
		windowBorderWidth,
	)
}

func (b *opinionatedCloseBox) Event(_ widget.Context, e event.Event) bool {
	mouse, ok := e.(*event.MouseEvent)
	if !ok || mouse.MouseType != event.MousePress || mouse.Button != event.ButtonLeft || b.window == nil {
		return false
	}
	b.window.closeFromChrome(b.ctx)
	return true
}

type opinionatedSortBox struct {
	widget.WidgetBase
	window *InventoryBagWindow
	ctx    Context
}

func newOpinionatedSortBox(window *InventoryBagWindow, ctx Context) *opinionatedSortBox {
	box := &opinionatedSortBox{window: window, ctx: ctx}
	box.SetVisible(true)
	box.SetEnabled(true)
	return box
}

func (b *opinionatedSortBox) Layout(_ widget.Context, constraints geometry.Constraints) geometry.Size {
	scale := rotheme.ContentScale()
	if scale <= 0 {
		scale = 1
	}
	size := constraints.Constrain(geometry.Sz(float32(inventoryBagSortW)*scale, float32(inventoryBagSearchH)*scale))
	b.SetBounds(geometry.FromPointSize(b.Position(), size))
	return size
}

func (b *opinionatedSortBox) Draw(_ widget.Context, canvas widget.Canvas) {
	if !b.IsVisible() || b.window == nil {
		return
	}
	bounds := b.Bounds()
	canvas.DrawRect(bounds, rotheme.Default.Colors.PanelBody)
	canvas.StrokeRect(bounds, rotheme.Default.Colors.WindowBorder, windowBorderWidth)
	rotheme.DrawText(canvas, b.window.nameSortLabel(), bounds, rotheme.Default.Typography.TextSize, rotheme.Default.Colors.Text, false, widget.TextAlignCenter)
}

func (b *opinionatedSortBox) Event(_ widget.Context, e event.Event) bool {
	mouse, ok := e.(*event.MouseEvent)
	if !ok || mouse.MouseType != event.MousePress || mouse.Button != event.ButtonLeft || b.window == nil {
		return false
	}
	b.window.toggleNameSort(b.ctx)
	b.SetNeedsRedraw(true)
	return true
}

func (w *InventoryBagWindow) nameSortLabel() string {
	if w != nil && w.sort == opinionatedSortNameDesc {
		return "Z-A"
	}
	return "A-Z"
}

func (w *InventoryBagWindow) toggleNameSort(ctx Context) {
	if w == nil {
		return
	}
	if w.sort == opinionatedSortName {
		w.sort = opinionatedSortNameDesc
	} else {
		w.sort = opinionatedSortName
	}
	w.applySearch(ctx, w.search)
}

func (w *InventoryBagWindow) syncOpinionatedDrag() {
	if w == nil {
		return
	}
	scale := w.elementScale()
	if scale <= 0 {
		scale = 1
	}
	w.titleHeight = 0
	if w.opinionatedSplit() {
		slot := float32(opinionatedCenterItemH()) * scale
		sources := w.centerSources
		if sources < 1 {
			sources = 1
		}
		w.dragBounds = geometry.NewRect(
			float32(opinionatedSideViewW())*scale,
			slot,
			float32(opinionatedMiddleW())*scale,
			float32(sources)*slot,
		)
		return
	}
	rowH := float32(inventoryBagFilterH) * scale
	closeW := float32(inventoryBagFilterH) * scale
	width := float32(w.width) - closeW
	if width < rowH {
		width = rowH
	}
	w.dragBounds = geometry.NewRect(0, 0, width, rowH)
}

func (w *InventoryBagWindow) syncSearchPointer(ctx Context) {
	if w == nil || !w.opinionated || ctx.Input == nil || !ctx.Input.MouseJustPressed(input.MouseButtonLeft) {
		return
	}
	if w.pointInSearchBar(ctx.Input.MouseX, ctx.Input.MouseY) {
		w.searchArmed = true
		if w.searchField != nil {
			w.searchField.SetFocused(true)
		}
		return
	}
	w.searchArmed = false
	if w.searchField != nil {
		w.searchField.SetFocused(false)
	}
}

func (w *InventoryBagWindow) pointInSearchBar(x, y int) bool {
	scale := w.elementScale()
	if scale <= 0 {
		scale = 1
	}
	if w.opinionatedSplit() {
		sources := w.centerSources
		if sources < 1 {
			sources = 1
		}
		return pointInRect(
			x, y,
			w.x+scaleInt(opinionatedSideViewW(), scale),
			w.y+scaleInt(opinionatedCenterSearchY(sources), scale),
			scaleInt(opinionatedMiddleW(), scale),
			scaleInt(opinionatedCenterItemH(), scale),
		)
	}
	top := w.y + scaleInt(inventoryBagFilterH*2, scale)
	height := scaleInt(inventorySearchRowH(), scale)
	width := w.width - scaleInt(inventoryBagSortW, scale)
	return pointInRect(x, y, w.x, top, width, height)
}

func (w *InventoryBagWindow) syncSearchTyping(ctx Context) {
	if w == nil || !w.searchArmed || ctx.Input == nil {
		return
	}
	next := w.search
	if w.searchField != nil && w.searchField.Text() != w.search {
		next = w.searchField.Text()
	} else {
		if text := ctx.Input.TextInput(); text != "" {
			next += text
		}
		if ctx.Input.JustPressed(input.KeyBackspace) && next != "" {
			runes := []rune(next)
			next = string(runes[:len(runes)-1])
		}
	}
	if next == w.search {
		return
	}
	if w.searchField != nil && w.searchField.Text() != next {
		w.searchField.SetText(next)
	}
	w.applySearch(ctx, next)
	if w.searchBox != nil {
		w.searchBox.SetNeedsRedraw(true)
	}
}

type opinionatedSearchRow struct {
	widget.WidgetBase
	search *inventorySearchBox
	sort   *opinionatedSortBox
}

func newOpinionatedSearchRow(search *inventorySearchBox, sort *opinionatedSortBox) *opinionatedSearchRow {
	row := &opinionatedSearchRow{search: search, sort: sort}
	row.SetVisible(true)
	row.SetEnabled(true)
	return row
}

func (r *opinionatedSearchRow) Layout(ctx widget.Context, constraints geometry.Constraints) geometry.Size {
	scale := rotheme.ContentScale()
	if scale <= 0 {
		scale = 1
	}
	width := constraints.MaxWidth
	if !opinionatedWidthBounded(width) {
		width = float32(opinionatedFrameW()) * scale
	}
	height := float32(inventorySearchRowH()) * scale
	sortW := float32(inventoryBagSortW) * scale
	barH := float32(inventoryBagSearchH) * scale
	inset := float32(inventoryBagSearchInset) * scale
	if sortW > width {
		sortW = width
	}
	searchW := width - sortW
	if r.search != nil {
		r.search.Layout(ctx, geometry.Tight(geometry.Sz(searchW, height)))
		var origin geometry.Point
		r.search.SetBounds(geometry.FromPointSize(origin, geometry.Sz(searchW, height)))
	}
	if r.sort != nil {
		r.sort.Layout(ctx, geometry.Tight(geometry.Sz(sortW, barH)))
		origin := geometry.Pt(searchW, inset)
		r.sort.SetBounds(geometry.FromPointSize(origin, geometry.Sz(sortW, barH)))
	}
	size := constraints.Constrain(geometry.Sz(width, height))
	r.SetBounds(geometry.FromPointSize(r.Position(), size))
	return size
}

func (r *opinionatedSearchRow) Draw(ctx widget.Context, canvas widget.Canvas) {
	if !r.IsVisible() {
		return
	}
	bounds := r.Bounds()
	canvas.PushTransform(bounds.Min)
	if r.search != nil {
		widget.DrawChild(r.search, ctx, canvas)
	}
	if r.sort != nil {
		widget.DrawChild(r.sort, ctx, canvas)
	}
	canvas.PopTransform()
}

func (r *opinionatedSearchRow) Event(ctx widget.Context, e event.Event) bool {
	mouse, ok := e.(*event.MouseEvent)
	if !ok {
		if r.search != nil {
			return r.search.Event(ctx, e)
		}
		return false
	}
	if r.sort != nil && r.sort.Bounds().Contains(mouse.Position) {
		return r.sort.Event(ctx, e)
	}
	if r.search != nil {
		return r.search.Event(ctx, e)
	}
	return false
}

func (r *opinionatedSearchRow) Children() []widget.Widget {
	children := make([]widget.Widget, 0, 2)
	if r.search != nil {
		children = append(children, r.search)
	}
	if r.sort != nil {
		children = append(children, r.sort)
	}
	return children
}
