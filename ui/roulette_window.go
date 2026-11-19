package ui

import (
	"fmt"

	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/widget"
	"github.com/kivutar/goro/network"
	"github.com/kivutar/goro/ui/rotheme"
)

type RouletteWindow struct {
	Window
	gold, silver, bronze int32
	message              string
}

func (w *RouletteWindow) Open(ctx Context) {
	w.EnsureWindow(rouletteWindowW, rouletteWindowH)
	if ctx.Network != nil {
		_ = ctx.Network.SendOpenRoulette()
	}
	w.Window.Open(ctx, w.widgetTree(ctx))
	w.OnLayoutReload(func() {
		w.SetSize(rouletteWindowW, rouletteWindowH)
		w.SetContent(w.widgetTree(w.ctx))
	})
	w.Publish(ctx)
}

func (w *RouletteWindow) Toggle(ctx Context) {
	if w.IsOpen() {
		if ctx.Network != nil {
			_ = ctx.Network.SendCloseRoulette()
		}
		w.Close()
		w.Publish(ctx)
		return
	}
	w.Open(ctx)
}

func (w *RouletteWindow) SetOpen(ctx Context, open network.RouletteOpen) {
	w.gold, w.silver, w.bronze = open.Gold, open.Silver, open.Bronze
	if open.Result != 0 {
		w.message = "Roulette is not available."
	}
	if !w.IsOpen() {
		w.Open(ctx)
		return
	}
	w.SetContent(w.widgetTree(ctx))
	w.Publish(ctx)
}

func (w *RouletteWindow) SetSpin(ctx Context, spin network.RouletteSpin) {
	w.gold, w.silver, w.bronze = spin.Gold, spin.Silver, spin.Bronze
	w.message = fmt.Sprintf("Item #%d", spin.ItemID)
	w.SetContent(w.widgetTree(ctx))
	w.Publish(ctx)
}

func (w *RouletteWindow) Update(ctx Context) bool {
	w.EnsureWindow(rouletteWindowW, rouletteWindowH)
	if !w.IsOpen() {
		return false
	}
	consumed := w.Window.Update(ctx)
	w.Publish(ctx)
	return consumed
}

func (w *RouletteWindow) widgetTree(ctx Context) widget.Widget {
	return Win(
		Title("Roulette"),
		CloseButton(true),
		OnClose(func() {
			if ctx.Network != nil {
				_ = ctx.Network.SendCloseRoulette()
			}
			w.Close()
			w.Publish(ctx)
		}),
		Size(float32(rouletteWindowW), float32(rouletteWindowH)),
		Content(primitives.Box(
			rotheme.Label(fmt.Sprintf("Gold %d  Silver %d  Bronze %d", w.gold, w.silver, w.bronze)),
			rotheme.Label(w.message),
		).Padding(float32(roulettePad)).Gap(float32(rouletteGap))),
		Footer(primitives.HBox(
			rotheme.Button("Spin", func() {
				if ctx.Network != nil {
					_ = ctx.Network.SendSpinRoulette()
				}
			}),
			rotheme.Button("Claim", func() {
				if ctx.Network != nil {
					_ = ctx.Network.SendClaimRoulette()
				}
			}),
		)),
	)
}
