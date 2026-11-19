package ui

import (
	"fmt"
	"strconv"

	"github.com/gogpu/ui/core/textfield"
	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/widget"
	"github.com/kivutar/goro/session"
	"github.com/kivutar/goro/ui/rotheme"
)

type EnchantUIWindow struct {
	Window
	group    int64
	items    []session.InventoryItem
	selected session.InventoryItem
	material string
	slot     string
	message  string
}

func (w *EnchantUIWindow) Open(ctx Context, group int64) {
	w.EnsureWindow(enchantUIWindowW, enchantUIWindowH)
	w.group = group
	w.selected = session.InventoryItem{}
	w.material = ""
	w.slot = "0"
	w.message = fmt.Sprintf("Enchant group %d", group)
	w.refreshItems(ctx)
	w.Window.Open(ctx, w.widgetTree(ctx))
	w.OnLayoutReload(func() {
		w.SetSize(enchantUIWindowW, enchantUIWindowH)
		w.SetContent(w.widgetTree(w.ctx))
	})
	w.Publish(ctx)
}

func (w *EnchantUIWindow) SetMessage(ctx Context, message string) {
	w.message = message
	if w.IsOpen() {
		w.SetContent(w.widgetTree(ctx))
		w.Publish(ctx)
	}
}

func (w *EnchantUIWindow) Update(ctx Context) bool {
	w.EnsureWindow(enchantUIWindowW, enchantUIWindowH)
	if !w.IsOpen() {
		return false
	}
	w.refreshItems(ctx)
	consumed := w.Window.Update(ctx)
	w.Publish(ctx)
	return consumed
}

func (w *EnchantUIWindow) refreshItems(ctx Context) {
	w.items = nil
	if ctx.Session == nil {
		return
	}
	w.items = append(w.items, ctx.Session.Inventory.Items...)
}

func (w *EnchantUIWindow) widgetTree(ctx Context) widget.Widget {
	return Win(
		Title("Enchant"),
		CloseButton(true),
		OnClose(func() {
			if ctx.Network != nil {
				_ = ctx.Network.SendCloseEnchantUI()
			}
			w.Close()
			w.Publish(ctx)
		}),
		Size(float32(enchantUIWindowW), float32(enchantUIWindowH)),
		Content(w.content(ctx)),
		Footer(w.footer(ctx)...),
	)
}

func (w *EnchantUIWindow) content(ctx Context) widget.Widget {
	rows := []widget.Widget{rotheme.Label(w.message)}
	for _, item := range w.items {
		item := item
		name := inventoryItemDisplayName(ctx.Resources, item)
		rows = append(rows, rotheme.Button(name, func() {
			w.selected = item
			w.message = name
			w.SetContent(w.widgetTree(ctx))
		}).Height(float32(enchantUIRowH)))
	}
	mat := rotheme.TextField(w.material, textfield.TypeText, func(value string) { w.material = filterDigits(value, 10) }, func(string) {}, textfield.MaxLength(10))
	slot := rotheme.TextField(w.slot, textfield.TypeText, func(value string) { w.slot = filterDigits(value, 2) }, func(string) {}, textfield.MaxLength(2))
	rows = append(rows,
		primitives.HBox(primitives.Box(rotheme.Label("Material")).Width(float32(enchantUILabelW)), primitives.Expanded(mat)).Height(float32(enchantUIRowH)),
		primitives.HBox(primitives.Box(rotheme.Label("Slot")).Width(float32(enchantUILabelW)), primitives.Expanded(slot)).Height(float32(enchantUIRowH)),
	)
	return primitives.Box(rows...).Padding(float32(enchantUIPad)).Gap(float32(enchantUIGap))
}

func (w *EnchantUIWindow) footer(ctx Context) []widget.Widget {
	return []widget.Widget{
		rotheme.Button("Random", func() {
			if ctx.Network != nil && w.selected.Index != 0 {
				_ = ctx.Network.SendRandomEnchant(w.group, w.selected.Index)
			}
		}),
		rotheme.Button("Perfect", func() {
			id, _ := strconv.Atoi(w.material)
			if ctx.Network != nil && w.selected.Index != 0 {
				_ = ctx.Network.SendPerfectEnchant(w.group, w.selected.Index, int32(id))
			}
		}),
		rotheme.Button("Upgrade", func() {
			slot, _ := strconv.Atoi(w.slot)
			if ctx.Network != nil && w.selected.Index != 0 {
				_ = ctx.Network.SendUpgradeEnchant(w.group, w.selected.Index, uint16(slot))
			}
		}),
		rotheme.Button("Reset", func() {
			if ctx.Network != nil && w.selected.Index != 0 {
				_ = ctx.Network.SendResetEnchant(w.group, w.selected.Index)
			}
		}),
	}
}
