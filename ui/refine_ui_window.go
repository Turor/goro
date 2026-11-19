package ui

import (
	"fmt"

	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/widget"
	"github.com/kivutar/goro/db"
	"github.com/kivutar/goro/network"
	"github.com/kivutar/goro/session"
	"github.com/kivutar/goro/ui/rotheme"
)

type RefineUIWindow struct {
	Window
	items     []session.InventoryItem
	selected  session.InventoryItem
	materials []network.RefineMaterial
	blessing  bool
	canBless  bool
	message   string
}

func (w *RefineUIWindow) Open(ctx Context) {
	w.EnsureWindow(refineUIWindowW, refineUIWindowH)
	w.selected = session.InventoryItem{}
	w.materials = nil
	w.blessing = false
	w.canBless = false
	w.message = "Select equipment to refine."
	w.refreshItems(ctx)
	w.Window.Open(ctx, w.widgetTree(ctx))
	w.OnLayoutReload(func() {
		w.SetSize(refineUIWindowW, refineUIWindowH)
		w.SetContent(w.widgetTree(w.ctx))
	})
	w.Publish(ctx)
}

func (w *RefineUIWindow) SetMaterials(ctx Context, list network.RefineMaterialList) {
	w.materials = list.Materials
	w.canBless = list.Blessing
	if item, ok := findInventoryItemByIndex(ctx.Session, list.Index); ok {
		w.selected = item
	}
	w.message = ""
	if w.IsOpen() {
		w.SetContent(w.widgetTree(ctx))
		w.Publish(ctx)
	}
}

func (w *RefineUIWindow) Update(ctx Context) bool {
	w.EnsureWindow(refineUIWindowW, refineUIWindowH)
	if !w.IsOpen() {
		return false
	}
	w.refreshItems(ctx)
	consumed := w.Window.Update(ctx)
	w.Publish(ctx)
	return consumed
}

func (w *RefineUIWindow) refreshItems(ctx Context) {
	w.items = nil
	if ctx.Session == nil {
		return
	}
	for _, item := range ctx.Session.Inventory.Items {
		if item.Type == db.ItemTypeWeapon || item.Type == db.ItemTypeArmor {
			w.items = append(w.items, item)
		}
	}
}

func (w *RefineUIWindow) widgetTree(ctx Context) widget.Widget {
	return Win(
		Title("Refine"),
		CloseButton(true),
		OnClose(func() {
			if ctx.Network != nil {
				_ = ctx.Network.SendCloseRefiningUI()
			}
			w.Close()
			w.Publish(ctx)
		}),
		Size(float32(refineUIWindowW), float32(refineUIWindowH)),
		Content(w.content(ctx)),
	)
}

func (w *RefineUIWindow) content(ctx Context) widget.Widget {
	rows := []widget.Widget{rotheme.Label(w.message)}
	for _, item := range w.items {
		item := item
		name := inventoryItemDisplayName(ctx.Resources, item)
		if item.Refine > 0 {
			name = fmt.Sprintf("+%d %s", item.Refine, name)
		}
		rows = append(rows, rotheme.Button(name, func() {
			w.selected = item
			if ctx.Network != nil {
				_ = ctx.Network.SendRefiningSelectItem(item.Index)
			}
		}).Height(float32(refineUIRowH)))
	}
	for _, mat := range w.materials {
		mat := mat
		label := fmt.Sprintf("Item %d  %d%%  %d zeny", mat.ItemID, mat.Chance, mat.Zeny)
		rows = append(rows, rotheme.Button(label, func() {
			if ctx.Network != nil && w.selected.Index != 0 {
				_ = ctx.Network.SendReqRefining(w.selected.Index, mat.ItemID, w.blessing && w.canBless)
			}
		}).Height(float32(refineUIRowH)))
	}
	return primitives.Box(rows...).Padding(float32(refineUIPad)).Gap(float32(refineUIGap))
}
