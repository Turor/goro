package ui

import (
	"fmt"
	"strconv"

	"github.com/gogpu/ui/core/textfield"
	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/widget"
	"github.com/kivutar/goro/network"
	"github.com/kivutar/goro/ui/rotheme"
)

type SearchStoreWindow struct {
	Window
	query   string
	message string
	hits    []network.SearchStoreHit
}

func (w *SearchStoreWindow) Open(ctx Context) {
	w.EnsureWindow(searchStoreWindowW, searchStoreWindowH)
	w.Window.Open(ctx, w.widgetTree(ctx))
	w.OnLayoutReload(func() {
		w.SetSize(searchStoreWindowW, searchStoreWindowH)
		w.SetContent(w.widgetTree(w.ctx))
	})
	w.Publish(ctx)
}

func (w *SearchStoreWindow) Toggle(ctx Context) {
	if w.IsOpen() {
		w.Close()
		w.Publish(ctx)
		return
	}
	w.Open(ctx)
}

func (w *SearchStoreWindow) SetAck(ctx Context, ack network.SearchStoreAck) {
	w.hits = ack.Hits
	w.message = fmt.Sprintf("%d results", len(ack.Hits))
	if !w.IsOpen() {
		w.Open(ctx)
		return
	}
	w.SetContent(w.widgetTree(ctx))
	w.Publish(ctx)
}

func (w *SearchStoreWindow) Update(ctx Context) bool {
	w.EnsureWindow(searchStoreWindowW, searchStoreWindowH)
	if !w.IsOpen() {
		return false
	}
	consumed := w.Window.Update(ctx)
	w.Publish(ctx)
	return consumed
}

func (w *SearchStoreWindow) widgetTree(ctx Context) widget.Widget {
	field := rotheme.TextField(w.query, textfield.TypeText, func(v string) { w.query = v }, func(string) { w.search(ctx) }, textfield.MaxLength(12))
	rows := []widget.Widget{
		primitives.HBox(primitives.Expanded(field), rotheme.Button("Search", func() { w.search(ctx) })).Height(float32(searchStoreRowH)),
		rotheme.Label(w.message),
	}
	for _, hit := range w.hits {
		rows = append(rows, primitives.Box(rotheme.Label(fmt.Sprintf("%s  #%d  %d z  x%d", hit.ShopName, hit.ItemID, hit.Price, hit.Amount))).Height(float32(searchStoreRowH)))
	}
	return Win(
		Title("Search Store"),
		CloseButton(true),
		OnClose(func() { w.Close(); w.Publish(ctx) }),
		Size(float32(searchStoreWindowW), float32(searchStoreWindowH)),
		Content(primitives.Box(rows...).Padding(float32(searchStorePad)).Gap(float32(searchStoreGap))),
	)
}

func (w *SearchStoreWindow) search(ctx Context) {
	id, err := strconv.ParseUint(w.query, 10, 32)
	if err != nil || ctx.Network == nil {
		w.message = "Enter an item ID."
		w.SetContent(w.widgetTree(ctx))
		return
	}
	_ = ctx.Network.SendSearchStore(uint32(id))
}
