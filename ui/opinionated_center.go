package ui

import (
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"

	"github.com/kivutar/goro/ui/rotheme"
)

// The split bag is 6 columns, one middle column, and 6 columns. The thin
// divider sits in that middle column. Tabs, categories, search, and sort
// occupy the rest of the same column so they sit between the panels.
type opinionatedCenterColumn struct {
	widget.WidgetBase
	window     *InventoryBagWindow
	close      *opinionatedCloseBox
	sources    []*tabWidget
	categories []*tabWidget
	search     *inventorySearchBox
	sort       *opinionatedSortBox
	scale      float32
}

func (w *InventoryBagWindow) opinionatedCenterColumn(ctx Context) widget.Widget {
	if w == nil {
		return nil
	}
	w.searchCtx = ctx
	sources := w.opinionatedSources(ctx)
	w.centerSources = len(sources)
	col := &opinionatedCenterColumn{
		window:     w,
		close:      newOpinionatedCloseBox(w, ctx),
		sources:    w.centerSourceTabs(ctx, sources),
		categories: w.centerCategoryTabs(ctx),
		search:     newInventorySearchBox(w),
		sort:       newOpinionatedSortBox(w, ctx),
	}
	w.searchBox = col.search
	col.SetVisible(true)
	col.SetEnabled(true)
	return col
}

func (w *InventoryBagWindow) centerSourceTabs(ctx Context, sources []opinionatedSourceTab) []*tabWidget {
	tabs := make([]*tabWidget, 0, len(sources))
	for _, source := range sources {
		source := source
		tabs = append(tabs, newTabWidget(tabWidgetConfig{
			label:         source.label,
			labelRotation: rotheme.TextRotationCounterClockwise,
			active:        source.source == w.source,
			width:         opinionatedMiddleW(),
			height:        opinionatedCenterItemH(),
			omitBottom:    source.source == w.source,
			activeFill:    inventoryBagTabActive,
			onClick: func() {
				w.hideTooltip()
				w.source = source.source
				w.ensureScrollSignal().Set(0)
				w.ensureSideScrollSignal().Set(0)
				w.refresh(ctx, w.itemInfo)
			},
		}))
	}
	return tabs
}

func (w *InventoryBagWindow) centerCategoryTabs(ctx Context) []*tabWidget {
	tabs := make([]*tabWidget, 0, len(opinionatedInventoryTabs))
	for _, tab := range opinionatedInventoryTabs {
		tab := tab
		tabs = append(tabs, newTabWidget(tabWidgetConfig{
			label:         tab.label,
			labelRotation: rotheme.TextRotationCounterClockwise,
			active:        tab.tab == w.tab,
			width:         opinionatedMiddleW(),
			height:        opinionatedCenterItemH(),
			omitBottom:    tab.tab == w.tab,
			activeFill:    inventoryBagTabActive,
			onClick: func() {
				w.hideTooltip()
				w.tab = tab.tab
				w.ensureScrollSignal().Set(0)
				w.lastClickItem = 0
				w.refresh(ctx, w.itemInfo)
			},
		}))
	}
	return tabs
}

func (c *opinionatedCenterColumn) Layout(ctx widget.Context, constraints geometry.Constraints) geometry.Size {
	scale := rotheme.ContentScale()
	if scale <= 0 {
		scale = 1
	}
	c.scale = scale
	width := float32(opinionatedMiddleW()) * scale
	height := float32(opinionatedBodyH()) * scale
	size := constraints.Constrain(geometry.Sz(width, height))
	c.SetBounds(geometry.FromPointSize(c.Position(), size))
	c.place(ctx, scale)
	return size
}

type centerPlaced interface {
	widget.Widget
	SetBounds(geometry.Rect)
	Bounds() geometry.Rect
	IsVisible() bool
}

func (c *opinionatedCenterColumn) place(ctx widget.Context, scale float32) {
	slot := float32(opinionatedCenterItemH()) * scale
	line := float32(opinionatedDividerW()) * scale
	inner := float32(opinionatedMiddleW())*scale - line
	if inner < line {
		inner = line
	}
	tight := geometry.Tight(geometry.Sz(inner, slot))
	y := float32(0)
	place := func(child centerPlaced) {
		if child == nil {
			return
		}
		child.Layout(ctx, tight)
		child.SetBounds(geometry.NewRect(line, y, inner, slot))
		y += slot
	}
	place(c.close)
	for _, tab := range c.sources {
		place(tab)
	}
	for _, tab := range c.categories {
		place(tab)
	}
	place(c.search)
	place(c.sort)
}

func (c *opinionatedCenterColumn) Draw(ctx widget.Context, canvas widget.Canvas) {
	if !c.IsVisible() {
		return
	}
	bounds := c.Bounds()
	canvas.PushTransform(bounds.Min)
	canvas.PushClip(geometry.NewRect(0, 0, bounds.Width(), bounds.Height()))
	line := float32(opinionatedDividerW()) * c.scale
	if line <= 0 {
		line = c.scale
	}
	if line > bounds.Width() {
		line = bounds.Width()
	}
	canvas.DrawRect(geometry.NewRect(0, 0, line, bounds.Height()), rotheme.Default.Colors.WindowBorder)
	for _, child := range c.placed() {
		if child == nil || !child.IsVisible() {
			continue
		}
		childBounds := child.Bounds()
		canvas.PushClip(childBounds)
		widget.DrawChild(child, ctx, canvas)
		canvas.PopClip()
	}
	canvas.PopClip()
	canvas.PopTransform()
}

func (c *opinionatedCenterColumn) Event(ctx widget.Context, e event.Event) bool {
	mouse, ok := e.(*event.MouseEvent)
	if !ok || !c.IsVisible() || !c.IsEnabled() {
		return false
	}
	local := *mouse
	local.Position = mouse.Position.Sub(c.Bounds().Min)
	for _, child := range c.placed() {
		if child == nil || !child.Bounds().Contains(local.Position) {
			continue
		}
		return child.Event(ctx, &local)
	}
	return false
}

func (c *opinionatedCenterColumn) placed() []centerPlaced {
	if c == nil {
		return nil
	}
	children := make([]centerPlaced, 0)
	if c.close != nil {
		children = append(children, c.close)
	}
	for _, tab := range c.sources {
		children = append(children, tab)
	}
	for _, tab := range c.categories {
		children = append(children, tab)
	}
	if c.search != nil {
		children = append(children, c.search)
	}
	if c.sort != nil {
		children = append(children, c.sort)
	}
	return children
}

func (c *opinionatedCenterColumn) Children() []widget.Widget {
	placed := c.placed()
	children := make([]widget.Widget, len(placed))
	for i, child := range placed {
		children[i] = child
	}
	return children
}
