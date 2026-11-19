package ui

import (
	"fmt"

	"github.com/gogpu/ui/core/textfield"
	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/widget"
	"github.com/kivutar/goro/network"
	"github.com/kivutar/goro/ui/rotheme"
)

type MarketWindow struct {
	Window
	title string
	items []network.MarketItem
}

func (w *MarketWindow) OpenList(ctx Context, items []network.MarketItem) {
	w.title = "NPC Market"
	w.items = items
	w.EnsureWindow(marketWindowW, marketWindowH)
	w.Window.Open(ctx, w.widgetTree(ctx))
	w.OnLayoutReload(func() {
		w.SetSize(marketWindowW, marketWindowH)
		w.SetContent(w.widgetTree(w.ctx))
	})
	w.Publish(ctx)
}

func (w *MarketWindow) Update(ctx Context) bool {
	w.EnsureWindow(marketWindowW, marketWindowH)
	if !w.IsOpen() {
		return false
	}
	consumed := w.Window.Update(ctx)
	w.Publish(ctx)
	return consumed
}

func (w *MarketWindow) widgetTree(ctx Context) widget.Widget {
	rows := []widget.Widget{}
	for _, item := range w.items {
		it := item
		rows = append(rows, primitives.HBox(
			primitives.Expanded(rotheme.Label(fmt.Sprintf("#%d  %d z  x%d", it.ItemID, it.Price, it.Qty))),
			rotheme.Button("Buy", func() {
				if ctx.Network != nil {
					_ = ctx.Network.SendMarketPurchase(it.ItemID, 1)
				}
			}),
		).Height(float32(marketRowH)))
	}
	return Win(
		Title(w.title),
		CloseButton(true),
		OnClose(func() {
			if ctx.Network != nil {
				_ = ctx.Network.SendMarketClose()
			}
			w.Close()
			w.Publish(ctx)
		}),
		Size(float32(marketWindowW), float32(marketWindowH)),
		Content(primitives.Box(rows...).Padding(float32(marketPad)).Gap(float32(marketGap))),
	)
}

type BarterWindow struct {
	Window
	items []network.BarterItem
}

func (w *BarterWindow) OpenList(ctx Context, items []network.BarterItem) {
	w.items = items
	w.EnsureWindow(marketWindowW, marketWindowH)
	w.Window.Open(ctx, w.widgetTree(ctx))
	w.OnLayoutReload(func() {
		w.SetSize(marketWindowW, marketWindowH)
		w.SetContent(w.widgetTree(w.ctx))
	})
	w.Publish(ctx)
}

func (w *BarterWindow) Update(ctx Context) bool {
	w.EnsureWindow(marketWindowW, marketWindowH)
	if !w.IsOpen() {
		return false
	}
	consumed := w.Window.Update(ctx)
	w.Publish(ctx)
	return consumed
}

func (w *BarterWindow) widgetTree(ctx Context) widget.Widget {
	rows := []widget.Widget{}
	for _, item := range w.items {
		it := item
		rows = append(rows, primitives.HBox(
			primitives.Expanded(rotheme.Label(barterRowLabel(it))),
			rotheme.Button("Trade", func() {
				if ctx.Network != nil {
					inv := uint16(0)
					if ctx.Session != nil {
						for _, item := range ctx.Session.Inventory.Items {
							if item.ItemID == it.Currency {
								inv = item.Index
								break
							}
						}
					}
					if it.Expanded {
						_ = ctx.Network.SendExpandedBarterPurchase(it.ItemID, 1, it.Index)
					} else {
						_ = ctx.Network.SendBarterPurchase(it.ItemID, 1, inv, it.Index)
					}
				}
			}),
		).Height(float32(marketRowH)))
	}
	return Win(
		Title("Barter"),
		CloseButton(true),
		OnClose(func() {
			if ctx.Network != nil {
				if len(w.items) > 0 && w.items[0].Expanded {
					_ = ctx.Network.SendExpandedBarterClose()
				} else {
					_ = ctx.Network.SendBarterClose()
				}
			}
			w.Close()
			w.Publish(ctx)
		}),
		Size(float32(marketWindowW), float32(marketWindowH)),
		Content(primitives.Box(rows...).Padding(float32(marketPad)).Gap(float32(marketGap))),
	)
}

func barterRowLabel(it network.BarterItem) string {
	if it.Currency == 0 {
		return fmt.Sprintf("#%d x%d for %d Z", it.ItemID, it.Amount, it.Cost)
	}
	return fmt.Sprintf("#%d x%d for #%d x%d", it.ItemID, it.Amount, it.Currency, it.Cost)
}

type ClanWindow struct {
	Window
	info network.ClanInfo
	chat string
}

func (w *ClanWindow) SetInfo(ctx Context, info network.ClanInfo) {
	if info.Name != "" {
		w.info.Name, w.info.Master, w.info.Map, w.info.ID = info.Name, info.Master, info.Map, info.ID
	}
	if info.Total != 0 || info.Online != 0 {
		w.info.Online, w.info.Total = info.Online, info.Total
	}
	w.EnsureWindow(clanWindowW, clanWindowH)
	if !w.IsOpen() {
		w.Window.Open(ctx, w.widgetTree(ctx))
		w.OnLayoutReload(func() {
			w.SetSize(clanWindowW, clanWindowH)
			w.SetContent(w.widgetTree(w.ctx))
		})
	} else {
		w.SetContent(w.widgetTree(ctx))
	}
	w.Publish(ctx)
}

func (w *ClanWindow) Update(ctx Context) bool {
	w.EnsureWindow(clanWindowW, clanWindowH)
	if !w.IsOpen() {
		return false
	}
	consumed := w.Window.Update(ctx)
	w.Publish(ctx)
	return consumed
}

func (w *ClanWindow) widgetTree(ctx Context) widget.Widget {
	field := rotheme.TextField(w.chat, textfield.TypeText, func(v string) { w.chat = v }, func(v string) {
		if ctx.Network != nil && v != "" {
			_ = ctx.Network.SendClanChat(v)
			w.chat = ""
		}
	}, textfield.MaxLength(60))
	return Win(
		Title("Clan"),
		CloseButton(true),
		OnClose(func() { w.Close(); w.Publish(ctx) }),
		Size(float32(clanWindowW), float32(clanWindowH)),
		Content(primitives.Box(
			rotheme.Label(w.info.Name),
			rotheme.Label(fmt.Sprintf("%s  %s", w.info.Master, w.info.Map)),
			rotheme.Label(fmt.Sprintf("%d / %d online", w.info.Online, w.info.Total)),
			primitives.Expanded(field),
		).Padding(float32(clanPad)).Gap(float32(clanGap))),
	)
}
