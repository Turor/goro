package ui

import (
	"github.com/kivutar/goro/input"
	"github.com/kivutar/goro/session"
)

func (w *InventoryBagWindow) frameSize() (int, int) {
	if w == nil || !w.opinionated {
		return inventoryBagWidth, inventoryBagHeight
	}
	height := opinionatedBodyH()
	if w.opinionatedSplit() {
		height += windowFooterHeight
	}
	return opinionatedFrameW(), height
}

func (w *InventoryBagWindow) noteTabPress(ctx Context) {
	if w == nil || !w.opinionated || ctx.Input == nil || !ctx.Input.MouseJustPressed(input.MouseButtonLeft) {
		return
	}
	scale := w.elementScale()
	if scale <= 0 {
		scale = 1
	}
	rowH := scaleInt(inventoryBagFilterH, scale)
	localX := ctx.Input.MouseX - w.x
	localY := ctx.Input.MouseY - w.y
	if !pointInRect(localX, localY, 0, 0, w.width-rowH, rowH) {
		return
	}
	w.tabPressX = ctx.Input.MouseX
	w.tabPressY = ctx.Input.MouseY
	w.tabPressed = true
}

func (w *InventoryBagWindow) tabClickAllowed(ctx Context) bool {
	if w == nil || !w.tabPressed || ctx.Input == nil {
		return true
	}
	return pointerWithinDragGrace(w.tabPressX, w.tabPressY, ctx.Input.MouseX, ctx.Input.MouseY)
}

func (w *InventoryBagWindow) gridForPane(pane int) *inventoryGridWidget {
	if w == nil {
		return nil
	}
	if pane == opinionatedPaneOther {
		return w.sideGrid
	}
	if pane == opinionatedPaneBag {
		return w.bagGrid
	}
	return nil
}

func (w *InventoryBagWindow) showSubmittedAmount(item session.InventoryItem, from, to int, amount uint16) {
	if w == nil || amount == 0 {
		return
	}
	if grid := w.gridForPane(from); grid != nil {
		grid.takeAmount(item.Index, int(amount))
	}
	if grid := w.gridForPane(to); grid != nil {
		grid.giveAmount(item, int(amount))
	}
}

func (g *inventoryGridWidget) takeAmount(index uint16, amount int) {
	if g == nil || amount <= 0 {
		return
	}
	for i := range g.cfg.items {
		if g.cfg.items[i].Index != index {
			continue
		}
		g.cfg.items[i].Amount -= amount
		if g.cfg.items[i].Amount <= 0 {
			g.cfg.items = append(g.cfg.items[:i], g.cfg.items[i+1:]...)
			if i < len(g.cfg.icons) {
				g.cfg.icons = append(g.cfg.icons[:i], g.cfg.icons[i+1:]...)
			}
		}
		g.cfg.amounts = inventoryGridAmountLabels(g.cfg.items)
		g.SetNeedsRedraw(true)
		return
	}
}

func (g *inventoryGridWidget) giveAmount(item session.InventoryItem, amount int) {
	if g == nil || amount <= 0 {
		return
	}
	for i := range g.cfg.items {
		if g.cfg.items[i].ItemID != item.ItemID {
			continue
		}
		g.cfg.items[i].Amount += amount
		g.cfg.amounts = inventoryGridAmountLabels(g.cfg.items)
		g.SetNeedsRedraw(true)
		return
	}
	copy := item
	copy.Amount = amount
	g.cfg.items = append(g.cfg.items, copy)
	g.cfg.amounts = inventoryGridAmountLabels(g.cfg.items)
	g.SetNeedsRedraw(true)
}

func (w *InventoryBagWindow) showItemHover(ctx Context, item session.InventoryItem) {
	if w == nil || item.ItemID == 0 {
		return
	}
	w.hideTooltip()
	x, y := 0, 0
	if ctx.Input != nil {
		x = ctx.Input.MouseX
		y = ctx.Input.MouseY
	}
	if w.hoverInfo != nil && w.hoverInfo.IsOpen() && w.hoverInfo.item.Index == item.Index && w.hoverInfo.item.ItemID == item.ItemID {
		return
	}
	if w.hoverInfo == nil {
		w.hoverInfo = &ItemInfoWindow{}
	}
	w.hoverInfo.openItem(ctx, item, x, y)
}

func (w *InventoryBagWindow) hideItemHover(ctx Context) {
	if w == nil || w.hoverInfo == nil || !w.hoverInfo.IsOpen() {
		return
	}
	w.hoverInfo.Close()
	w.hoverInfo.Publish(ctx)
}

func (w *InventoryBagWindow) pinItemHover(ctx Context, itemInfo *ItemWindows, item session.InventoryItem, mx, my int) {
	if itemInfo == nil {
		return
	}
	if w != nil && w.hoverInfo != nil && w.hoverInfo.IsOpen() && w.hoverInfo.item.ItemID == item.ItemID {
		itemInfo.adopt(w.hoverInfo)
		w.hoverInfo = nil
		return
	}
	itemInfo.openItem(ctx, item, mx, my)
}

func (w *ItemWindows) adopt(info *ItemInfoWindow) {
	if w == nil || info == nil {
		return
	}
	w.descriptions = append(w.descriptions, info)
}

func (w *InventoryBagWindow) updateItemHover(ctx Context) bool {
	if w == nil || w.hoverInfo == nil || !w.hoverInfo.IsOpen() {
		return false
	}
	return w.hoverInfo.Update(ctx, nil)
}
