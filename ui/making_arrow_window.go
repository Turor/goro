package ui

import (
	"image"
	"time"

	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/state"
	"github.com/gogpu/ui/widget"
	"github.com/kivutar/goro/glog"
	"github.com/kivutar/goro/input"
	"github.com/kivutar/goro/network"
	"github.com/kivutar/goro/res"
	"github.com/kivutar/goro/session"
	"github.com/kivutar/goro/ui/rotheme"
)

type MakingArrowWindow struct {
	Window
	scrollY      state.Signal[float32]
	selectedRow  int
	itemIDs      []uint32
	lastClickAt  time.Time
	lastClickRow int
	icons        map[identifyItemIconKey]image.Image
	iconMiss     map[identifyItemIconKey]struct{}
}

func (w *MakingArrowWindow) OpenList(ctx Context, list network.MakingArrowList) {
	w.EnsureWindow(makingArrowWindowWidth, makingArrowWindowHeight)
	w.SetSize(makingArrowWindowWidth, makingArrowWindowHeight)
	w.itemIDs = append(w.itemIDs[:0], list.ItemIDs...)
	w.selectedRow = 0
	w.lastClickRow = -1
	w.lastClickAt = time.Time{}
	w.ensureScrollSignal().Set(0)
	w.clampScroll()
	if len(w.itemIDs) == 0 {
		w.Close()
		w.Publish(ctx)
		return
	}
	w.Open(ctx, w.widgetTree(ctx))
	w.OnLayoutReload(func() {
		w.SetSize(makingArrowWindowWidth, makingArrowWindowHeight)
		w.SetContent(w.widgetTree(ctx))
	})
	w.Publish(ctx)
}

func (w *MakingArrowWindow) Update(ctx Context) bool {
	w.EnsureWindow(makingArrowWindowWidth, makingArrowWindowHeight)
	if !w.IsOpen() {
		return false
	}
	w.clampScroll()
	if ctx.Input != nil && ctx.Input.JustPressed(input.KeyEscape) {
		w.cancel(ctx)
		w.Publish(ctx)
		return true
	}
	consumed := w.Window.Update(ctx)
	keyboardConsumed := false
	if w.IsOpen() {
		keyboardConsumed = updateSelectionTableKeyboard(ctx, &w.selectedRow, len(w.itemIDs), makingArrowRows, float32(makingArrowRowH), w.ensureScrollSignal(), func() {
			w.confirm(ctx)
		})
		if keyboardConsumed && w.IsOpen() {
			w.RebindContent(ctx, w.widgetTree(ctx))
		}
	}
	if w.IsOpen() {
		w.updateDoubleClick(ctx)
	}
	w.Publish(ctx)
	return consumed || keyboardConsumed
}

func (w *MakingArrowWindow) widgetTree(ctx Context) widget.Widget {
	return Win(
		Title("Arrow Crafting"),
		CloseButton(true),
		OnClose(func() {
			w.cancel(ctx)
			w.Publish(ctx)
		}),
		Size(float32(makingArrowWindowWidth), float32(makingArrowWindowHeight)),
		Content(
			primitives.Box(w.tableWidget(ctx)).
				Height(makingArrowTableHeight()).
				Background(rotheme.Default.Colors.PanelBody),
		),
		Footer(
			primitives.Expanded(primitives.Box()),
			rotheme.Button("Cancel", func() {
				w.cancel(ctx)
				w.Publish(ctx)
			}),
			rotheme.Button("OK", func() {
				w.confirm(ctx)
				w.Publish(ctx)
			}),
		),
	)
}

func (w *MakingArrowWindow) tableWidget(ctx Context) *rotheme.TableViewWidget {
	ids := append([]uint32(nil), w.itemIDs...)
	return itemTableView(
		w.itemTableRows(ctx, ids),
		"Item",
		float32(makingArrowRowH),
		float32(makingArrowTableHeaderH),
		"No arrow materials",
		w.ensureScrollSignal(),
		w.selectedRow,
		func(row int) {
			w.selectedRow = row
		},
	)
}

func (w *MakingArrowWindow) confirm(ctx Context) {
	if w.selectedRow < 0 || w.selectedRow >= len(w.itemIDs) {
		return
	}
	w.makeArrow(ctx, w.itemIDs[w.selectedRow])
}

func (w *MakingArrowWindow) updateDoubleClick(ctx Context) {
	if ctx.Input == nil || !ctx.Input.MouseJustPressed(input.MouseButtonLeft) {
		return
	}
	row, ok := w.rowAtMouse(ctx.Input.MouseX, ctx.Input.MouseY)
	if !ok {
		return
	}
	now := time.Now()
	if w.lastClickRow == row && now.Sub(w.lastClickAt) <= 360*time.Millisecond {
		w.selectedRow = row
		w.lastClickRow = -1
		w.lastClickAt = time.Time{}
		w.confirm(ctx)
		return
	}
	w.lastClickRow = row
	w.lastClickAt = now
}

func (w *MakingArrowWindow) rowAtMouse(mouseX, mouseY int) (int, bool) {
	scale := w.elementScale()
	tableX := w.x
	tableY := w.y + scaleInt(ROWindowTitleHeight, scale)
	rowY := tableY + scaleInt(makingArrowTableHeaderH, scale)
	if !pointInRect(mouseX, mouseY, tableX, rowY, scaleInt(scrollbarSafeIntWidth(makingArrowWindowWidth), scale), scaleInt(makingArrowRows*makingArrowRowH, scale)) {
		return 0, false
	}
	row := int((float32(mouseY-rowY) + w.ensureScrollSignal().Get()) / scaleF(float32(makingArrowRowH), scale))
	if row < 0 || row >= len(w.itemIDs) {
		return 0, false
	}
	return row, true
}

func (w *MakingArrowWindow) makeArrow(ctx Context, itemID uint32) {
	if ctx.Network == nil {
		glog.Warnf("making arrow failed: not connected")
		return
	}
	if err := ctx.Network.SendMakingArrow(itemID); err != nil {
		glog.Warnf("making arrow failed: %v", err)
		return
	}
	glog.Debugf("making arrow requested item=%d", itemID)
	w.Close()
}

func (w *MakingArrowWindow) cancel(ctx Context) {
	if ctx.Network != nil {
		if err := ctx.Network.SendMakingArrow(0xFFFF); err != nil {
			glog.Warnf("making arrow cancel failed: %v", err)
		}
	}
	w.Close()
}

func (w *MakingArrowWindow) itemTableRows(ctx Context, itemIDs []uint32) []itemTableRow {
	rows := make([]itemTableRow, len(itemIDs))
	for i, itemID := range itemIDs {
		rows[i] = itemTableRow{
			name: inventoryItemDisplayName(ctx.Resources, session.InventoryItem{ItemID: itemID, Identified: true}),
			icon: w.itemIconImage(ctx.Resources, itemID),
		}
	}
	return rows
}

func (w *MakingArrowWindow) itemIconImage(manager *res.Manager, itemID uint32) image.Image {
	return cachedIdentifiedItemIcon(manager, itemID, &w.icons, &w.iconMiss)
}

func (w *MakingArrowWindow) clampScroll() {
	if w.selectedRow >= len(w.itemIDs) {
		w.selectedRow = -1
	}
	scroll := w.ensureScrollSignal()
	maxScroll := float32(maxInt(0, len(w.itemIDs)-makingArrowRows) * makingArrowRowH)
	switch value := scroll.Get(); {
	case value < 0:
		scroll.Set(0)
	case value > maxScroll:
		scroll.Set(maxScroll)
	}
}

func (w *MakingArrowWindow) ensureScrollSignal() state.Signal[float32] {
	if w.scrollY == nil {
		w.scrollY = state.NewSignal[float32](0)
	}
	return w.scrollY
}

func makingArrowTableHeight() float32 {
	return float32(makingArrowTableHeaderH + makingArrowRows*makingArrowRowH)
}
