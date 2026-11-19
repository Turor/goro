package ui

import (
	"fmt"

	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/widget"
	"github.com/kivutar/goro/network"
	"github.com/kivutar/goro/ui/rotheme"
)

type EdenWindow struct {
	Window
	list     network.EdenList
	shop     network.EdenShop
	message  string
	selected uint32
}

func (w *EdenWindow) SetList(ctx Context, list network.EdenList) {
	w.list = list
	w.message = fmt.Sprintf("Lv %d  merit %d", list.Level, list.Merit)
	w.open(ctx)
}

func (w *EdenWindow) SetShop(ctx Context, shop network.EdenShop) {
	w.shop = shop
	w.message = fmt.Sprintf("Shop merit %d", shop.Merit)
	w.open(ctx)
}

func (w *EdenWindow) SetResult(ctx Context, res network.EdenResult) {
	w.message = fmt.Sprintf("Result %d", res.Result)
	if w.IsOpen() {
		w.SetContent(w.widgetTree(ctx))
		w.Publish(ctx)
	}
}

func (w *EdenWindow) open(ctx Context) {
	w.EnsureWindow(edenWindowW, edenWindowH)
	if !w.IsOpen() {
		w.Window.Open(ctx, w.widgetTree(ctx))
		w.OnLayoutReload(func() {
			w.SetSize(edenWindowW, edenWindowH)
			w.SetContent(w.widgetTree(w.ctx))
		})
	} else {
		w.SetContent(w.widgetTree(ctx))
	}
	w.Publish(ctx)
}

func (w *EdenWindow) Update(ctx Context) bool {
	w.EnsureWindow(edenWindowW, edenWindowH)
	if !w.IsOpen() {
		return false
	}
	consumed := w.Window.Update(ctx)
	w.Publish(ctx)
	return consumed
}

func (w *EdenWindow) widgetTree(ctx Context) widget.Widget {
	rows := []widget.Widget{rotheme.Label(w.message)}
	for _, q := range w.list.Quests {
		q := q
		mark := " "
		if q.QuestID == w.selected {
			mark = "*"
		}
		rows = append(rows, primitives.HBox(
			primitives.Expanded(rotheme.Label(fmt.Sprintf("%s #%d %s", mark, q.QuestID, q.Name))),
			rotheme.Button("Pick", func() {
				w.selected = q.QuestID
				w.SetContent(w.widgetTree(ctx))
			}),
		).Height(float32(leftoverRowH)))
	}
	for _, it := range w.shop.Items {
		it := it
		rows = append(rows, primitives.HBox(
			primitives.Expanded(rotheme.Label(fmt.Sprintf("#%d x%d  %d m / %d z", it.ItemID, it.Amount, it.PriceMerit, it.PriceZeny))),
			rotheme.Button("Buy", func() {
				if ctx.Network != nil {
					_ = ctx.Network.SendEdenReq(network.EdenActBuy, uint32(it.Slot))
				}
			}),
		).Height(float32(leftoverRowH)))
	}
	return Win(
		Title("Eden"),
		CloseButton(true),
		OnClose(func() { w.Close(); w.Publish(ctx) }),
		Size(float32(edenWindowW), float32(edenWindowH)),
		Content(primitives.Box(rows...).Padding(float32(leftoverPad)).Gap(float32(leftoverGap))),
		Footer(primitives.HBox(
			rotheme.Button("List", func() {
				if ctx.Network != nil {
					_ = ctx.Network.SendEdenReq(network.EdenActList, 0)
				}
			}),
			rotheme.Button("Accept", func() {
				if ctx.Network != nil && w.selected != 0 {
					_ = ctx.Network.SendEdenReq(network.EdenActAccept, w.selected)
				}
			}),
			rotheme.Button("Done", func() {
				if ctx.Network != nil && w.selected != 0 {
					_ = ctx.Network.SendEdenReq(network.EdenActComplete, w.selected)
				}
			}),
			rotheme.Button("Drop", func() {
				if ctx.Network != nil && w.selected != 0 {
					_ = ctx.Network.SendEdenReq(network.EdenActAbandon, w.selected)
				}
			}),
		)),
	)
}
