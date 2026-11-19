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

type GradeEnchantWindow struct {
	Window
	items    []session.InventoryItem
	selected session.InventoryItem
	list     network.GradeMaterialList
	message  string
}

func (w *GradeEnchantWindow) Open(ctx Context) {
	w.EnsureWindow(gradeWindowW, gradeWindowH)
	w.selected = session.InventoryItem{}
	w.list = network.GradeMaterialList{}
	w.message = "Select equipment to grade."
	w.refresh(ctx)
	w.Window.Open(ctx, w.widgetTree(ctx))
	w.OnLayoutReload(func() {
		w.SetSize(gradeWindowW, gradeWindowH)
		w.SetContent(w.widgetTree(w.ctx))
	})
	w.Publish(ctx)
}

func (w *GradeEnchantWindow) SetMaterials(ctx Context, list network.GradeMaterialList) {
	w.list = list
	if item, ok := findInventoryItemByIndex(ctx.Session, list.Index); ok {
		w.selected = item
	}
	w.message = fmt.Sprintf("Chance %d%%", list.Chance)
	if w.IsOpen() {
		w.SetContent(w.widgetTree(ctx))
		w.Publish(ctx)
	}
}

func (w *GradeEnchantWindow) SetAck(ctx Context, ack network.GradeAck) {
	w.message = fmt.Sprintf("Grade %d result %d", ack.Grade, ack.Result)
	if w.IsOpen() {
		w.SetContent(w.widgetTree(ctx))
		w.Publish(ctx)
	}
}

func (w *GradeEnchantWindow) Update(ctx Context) bool {
	w.EnsureWindow(gradeWindowW, gradeWindowH)
	if !w.IsOpen() {
		return false
	}
	w.refresh(ctx)
	consumed := w.Window.Update(ctx)
	w.Publish(ctx)
	return consumed
}

func (w *GradeEnchantWindow) refresh(ctx Context) {
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

func (w *GradeEnchantWindow) widgetTree(ctx Context) widget.Widget {
	rows := []widget.Widget{rotheme.Label(w.message)}
	for _, item := range w.items {
		item := item
		name := inventoryItemDisplayName(ctx.Resources, item)
		rows = append(rows, rotheme.Button(name, func() {
			w.selected = item
			if ctx.Network != nil {
				_ = ctx.Network.SendGradeSelect(item.Index)
			}
		}).Height(float32(gradeRowH)))
	}
	for i, mat := range w.list.Materials {
		i, mat := i, mat
		label := fmt.Sprintf("Mat %d x%d  %d zeny", mat.ItemID, mat.Amount, mat.Price)
		rows = append(rows, rotheme.Button(label, func() {
			if ctx.Network != nil && w.selected.Index != 0 {
				_ = ctx.Network.SendGradeRequest(w.selected.Index, int32(i), false, 0)
			}
		}).Height(float32(gradeRowH)))
	}
	return Win(
		Title("Grade Enchant"),
		CloseButton(true),
		OnClose(func() {
			if ctx.Network != nil {
				_ = ctx.Network.SendGradeClose()
			}
			w.Close()
			w.Publish(ctx)
		}),
		Size(float32(gradeWindowW), float32(gradeWindowH)),
		Content(primitives.Box(rows...).Padding(float32(gradePad)).Gap(float32(gradeGap))),
	)
}
