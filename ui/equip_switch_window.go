package ui

import (
	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/widget"
	"github.com/kivutar/goro/ui/rotheme"
)

type EquipSwitchWindow struct {
	Window
}

func (w *EquipSwitchWindow) Toggle(ctx Context) {
	w.EnsureWindow(equipSwitchWindowW, equipSwitchWindowH)
	if w.IsOpen() {
		w.Close()
		w.Publish(ctx)
		return
	}
	w.Window.Open(ctx, w.widgetTree(ctx))
	w.OnLayoutReload(func() {
		w.SetSize(equipSwitchWindowW, equipSwitchWindowH)
		w.SetContent(w.widgetTree(w.ctx))
	})
	w.Publish(ctx)
}

func (w *EquipSwitchWindow) Update(ctx Context) bool {
	w.EnsureWindow(equipSwitchWindowW, equipSwitchWindowH)
	if !w.IsOpen() {
		return false
	}
	consumed := w.Window.Update(ctx)
	w.Publish(ctx)
	return consumed
}

func (w *EquipSwitchWindow) widgetTree(ctx Context) widget.Widget {
	return Win(
		Title("Equip Switch"),
		CloseButton(true),
		OnClose(func() { w.Close(); w.Publish(ctx) }),
		Size(float32(equipSwitchWindowW), float32(equipSwitchWindowH)),
		Content(primitives.Box(rotheme.Label("Swap the alternate equipment set.")).Padding(float32(equipSwitchPad))),
		Footer(rotheme.Button("Switch", func() {
			if ctx.Network != nil {
				_ = ctx.Network.SendEquipSwitchAll()
			}
		})),
	)
}
