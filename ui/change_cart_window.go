package ui

import (
	"fmt"

	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/widget"
	"github.com/kivutar/goro/client"
	"github.com/kivutar/goro/glog"
	"github.com/kivutar/goro/ui/rotheme"
)

type ChangeCartWindow struct {
	Window
	status string
}

func (w *ChangeCartWindow) Open(ctx client.Context) {
	w.ctx = ctx
	w.EnsureWindow(changeCartWindowWidth, changeCartWindowHeight)
	w.SetSize(changeCartWindowWidth, changeCartWindowHeight)
	w.status = ""
	if !inventoryBagHasCart(ctx) {
		w.status = "You need a cart."
	}
	w.Window.Open(ctx, w.widgetTree(ctx))
	w.OnLayoutReload(func() {
		w.SetSize(changeCartWindowWidth, changeCartWindowHeight)
		w.SetContent(w.widgetTree(ctx))
	})
	w.Publish(ctx)
}

func (w *ChangeCartWindow) Update(ctx client.Context) bool {
	w.ctx = ctx
	if !w.IsOpen() {
		return false
	}
	if w.Window.Update(ctx) {
		if !w.IsOpen() {
			w.Publish(ctx)
			return true
		}
		w.Publish(ctx)
		return true
	}
	w.Publish(ctx)
	return true
}

func (w *ChangeCartWindow) refresh(ctx client.Context) {
	if !w.IsOpen() {
		return
	}
	w.SetContent(w.widgetTree(ctx))
	w.Publish(ctx)
}

func (w *ChangeCartWindow) selectCart(ctx client.Context, cartNum int) {
	if ctx.Network == nil {
		w.status = "Not connected."
		w.refresh(ctx)
		return
	}
	if err := ctx.Network.SendChangeCart(uint16(cartNum)); err != nil {
		glog.Warnf("change cart failed cart=%d: %v", cartNum, err)
		w.status = fmt.Sprintf("Change failed: %v", err)
		w.refresh(ctx)
		return
	}
	w.Window.Close()
	w.Publish(ctx)
}

func (w *ChangeCartWindow) widgetTree(ctx client.Context) widget.Widget {
	return Win(
		Title("Change Cart"),
		CloseButton(true),
		OnClose(func() {
			w.Window.Close()
			w.Publish(ctx)
		}),
		Size(float32(changeCartWindowWidth), float32(changeCartWindowHeight)),
		Content(
			primitives.Box(
				rotheme.Text("Select a cart."),
				w.cartGrid(ctx),
				rotheme.Text(w.status),
			).
				Padding(float32(changeCartPad)).
				Gap(float32(changeCartGap)).
				CrossAlign(primitives.CrossAxisStretch),
		),
	)
}

func (w *ChangeCartWindow) cartGrid(ctx client.Context) widget.Widget {
	allowed := changeCartChoices(baseLevelForChangeCart(ctx))
	rows := make([]widget.Widget, 0, changeCartColumns)
	for i := 0; i < len(allowed); i += changeCartColumns {
		cells := make([]widget.Widget, 0, changeCartColumns)
		for _, cartNum := range allowed[i:minInt(i+changeCartColumns, len(allowed))] {
			num := cartNum
			cells = append(cells, rotheme.Button(fmt.Sprintf("%d", num), func() {
				w.selectCart(ctx, num)
			}).Width(float32(changeCartButtonW)))
		}
		rows = append(rows, primitives.HBox(cells...).Gap(float32(changeCartGap)))
	}
	return primitives.Box(rows...).Gap(float32(changeCartGap))
}

func changeCartChoices(baseLevel int) []int {
	out := []int{1}
	if baseLevel > 40 {
		out = append(out, 2)
	}
	if baseLevel > 65 {
		out = append(out, 3)
	}
	if baseLevel > 80 {
		out = append(out, 4)
	}
	if baseLevel > 90 {
		out = append(out, 5)
	}
	if baseLevel > 100 {
		out = append(out, 6)
	}
	if baseLevel > 110 {
		out = append(out, 7)
	}
	if baseLevel > 120 {
		out = append(out, 8)
	}
	if baseLevel > 130 {
		out = append(out, 9)
	}
	return out
}

func baseLevelForChangeCart(ctx client.Context) int {
	if ctx.Session == nil {
		return 0
	}
	if ctx.Session.Progress.BaseLevel > 0 {
		return ctx.Session.Progress.BaseLevel
	}
	return int(ctx.Session.Selected.Level)
}
