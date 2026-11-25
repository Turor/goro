package ui

import (
	"github.com/gogpu/ui/core/scrollview"
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/state"
	"github.com/gogpu/ui/widget"
	"github.com/kivutar/goro/session"
)

func (w *InventoryBagWindow) opinionatedWindow(ctx Context, itemInfo *ItemWindows) widget.Widget {
	w.normalizeOpinionatedSource(ctx)
	width, height := inventoryBagFrameSize(true)
	var content widget.Widget
	if w.opinionatedSplit() {
		content = w.opinionatedPanes(ctx, itemInfo)
	} else {
		content = primitives.Box(
			w.opinionatedTabRow(ctx),
			w.filterRows(ctx),
			w.searchRow(ctx),
			w.opinionatedPanes(ctx, itemInfo),
		).Gap(0)
	}
	w.syncOpinionatedDrag()
	return Win(
		TitleBar(false),
		CloseButton(false),
		OnClose(func() {
			w.closeFromChrome(ctx)
		}),
		Size(float32(width), float32(height)),
		Content(content),
	)
}

func (w *InventoryBagWindow) closeFromChrome(ctx Context) {
	w.amountPrompt.Close(ctx)
	w.disarmSearch()
	w.Window.Close()
	w.Publish(ctx)
}

func (w *InventoryBagWindow) opinionatedTabRow(ctx Context) widget.Widget {
	return primitives.HBox(
		primitives.Expanded(w.sourceTabs(ctx)),
		newOpinionatedCloseBox(w, ctx),
	).Gap(0)
}

func (w *InventoryBagWindow) sourceTabs(ctx Context) widget.Widget {
	sources := w.opinionatedSources(ctx)
	tabs := make([]*tabWidget, 0, len(sources))
	active := 0
	for i, source := range sources {
		source := source
		if source.source == w.source {
			active = i
		}
		tabs = append(tabs, newTabWidget(tabWidgetConfig{
			label:      source.label,
			active:     source.source == w.source,
			width:      opinionatedEqualTabWidth(i, len(sources)),
			height:     inventoryBagFilterH,
			omitBottom: source.source == w.source,
			activeFill: inventoryBagTabActive,
			onClick: func() {
				w.hideTooltip()
				w.source = source.source
				w.ensureScrollSignal().Set(0)
				w.ensureSideScrollSignal().Set(0)
				w.refresh(ctx, w.itemInfo)
			},
		}))
	}
	return newOpinionatedTabStrip(tabs, active)
}

func (w *InventoryBagWindow) opinionatedPanes(ctx Context, itemInfo *ItemWindows) widget.Widget {
	if !w.opinionatedSplit() {
		return w.opinionatedScroll(ctx, itemInfo, w.tabItems(ctx.Session), opinionatedBagCols, opinionatedContentW(), opinionatedPaneBag, w.ensureScrollSignal())
	}
	return primitives.HBox(
		w.opinionatedScroll(ctx, itemInfo, w.bagItems(ctx.Session), opinionatedSplitCols, opinionatedSplitContentW(), opinionatedPaneBag, w.ensureScrollSignal()),
		w.opinionatedCenterColumn(ctx),
		w.opinionatedScroll(ctx, itemInfo, w.tabItems(ctx.Session), opinionatedSplitCols, opinionatedSplitContentWRight(), opinionatedPaneOther, w.ensureSideScrollSignal()),
	).Gap(0)
}

func (w *InventoryBagWindow) opinionatedScroll(ctx Context, itemInfo *ItemWindows, items []session.InventoryItem, cols, viewW, pane int, scroll state.Signal[float32]) widget.Widget {
	grid := newInventoryGridWidget(inventoryGridConfig{
		items:     items,
		icons:     w.itemIcons(ctx, items),
		amounts:   inventoryGridAmountLabels(items),
		cols:      cols,
		viewWidth: viewW,
		fillWidth: true,
		grid:      true,
		onPress: func(item session.InventoryItem) {
			w.startPaneDrag(ctx, item, pane)
		},
		onHover: func(item session.InventoryItem) { w.showTooltip(ctx, item) },
		onLeave: func() { w.hideTooltip() },
		onRightClickMods: func(item session.InventoryItem, mx, my int, mods event.Modifiers) {
			w.opinionatedRightClick(ctx, itemInfo, item, mx, my, mods, pane)
		},
	})
	if pane == opinionatedPaneBag {
		w.bagGrid = grid
		if !w.opinionatedSplit() {
			w.sideGrid = nil
		}
	} else {
		w.sideGrid = grid
	}
	scrollBox := scrollview.New(
		grid,
		scrollview.DirectionOpt(scrollview.Vertical),
		scrollview.ScrollbarOpt(scrollview.ScrollbarAlways),
		scrollview.ScrollYSignal(scroll),
		scrollview.ScrollStep(float32(inventoryBagCell)),
	)
	width := opinionatedScrollW(viewW)
	if pane == opinionatedPaneBag && !w.opinionatedSplit() {
		width = opinionatedScrollW(opinionatedContentW())
	}
	return primitives.Box(scrollBox).Width(float32(width)).Height(float32(opinionatedPaneH(w.opinionatedSplit())))
}

func (w *InventoryBagWindow) ensureSideScrollSignal() state.Signal[float32] {
	if w.scrollYSide == nil {
		w.scrollYSide = state.NewSignal[float32](0)
	}
	return w.scrollYSide
}
