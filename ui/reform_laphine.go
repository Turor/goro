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

type ReformWindow struct {
	Window
	itemID  int32
	index   string
	message string
}

func (w *ReformWindow) Open(ctx Context, itemID int32) {
	w.itemID = itemID
	w.message = fmt.Sprintf("Item #%d", itemID)
	w.EnsureWindow(leftoverWindowW, leftoverWindowH)
	w.Window.Open(ctx, w.widgetTree(ctx))
	w.OnLayoutReload(func() {
		w.SetSize(leftoverWindowW, leftoverWindowH)
		w.SetContent(w.widgetTree(w.ctx))
	})
	w.Publish(ctx)
}

func (w *ReformWindow) SetAck(ctx Context, ack network.ReformAck) {
	if ack.Result == 0 {
		w.message = "Reformed."
	} else {
		w.message = "Reform failed."
	}
	if w.IsOpen() {
		w.SetContent(w.widgetTree(ctx))
		w.Publish(ctx)
	}
}

func (w *ReformWindow) Update(ctx Context) bool {
	w.EnsureWindow(leftoverWindowW, leftoverWindowH)
	if !w.IsOpen() {
		return false
	}
	consumed := w.Window.Update(ctx)
	w.Publish(ctx)
	return consumed
}

func (w *ReformWindow) widgetTree(ctx Context) widget.Widget {
	field := rotheme.TextField(w.index, textfield.TypeText, func(v string) { w.index = filterDigits(v, 5) }, func(string) {}, textfield.MaxLength(5))
	return Win(
		Title("Item Reform"),
		CloseButton(true),
		OnClose(func() {
			if ctx.Network != nil {
				_ = ctx.Network.SendReformClose()
			}
			w.Close()
			w.Publish(ctx)
		}),
		Size(float32(leftoverWindowW), float32(leftoverWindowH)),
		Content(primitives.Box(
			rotheme.Label(w.message),
			primitives.HBox(primitives.Box(rotheme.Label("Index")).Width(float32(leftoverLabelW)), primitives.Expanded(field)).Height(float32(leftoverRowH)),
		).Padding(float32(leftoverPad)).Gap(float32(leftoverGap))),
		Footer(rotheme.Button("Reform", func() {
			n, _ := strconv.ParseUint(w.index, 10, 16)
			if ctx.Network != nil && n > 0 {
				_ = ctx.Network.SendReformItem(w.itemID, uint16(n))
			}
		})),
	)
}

type LaphineWindow struct {
	Window
	upgrade bool
	itemID  uint32
	index   string
	message string
}

func (w *LaphineWindow) Open(ctx Context, open network.LaphineOpen) {
	w.upgrade = open.Upgrade
	w.itemID = open.ItemID
	w.message = fmt.Sprintf("Item #%d", open.ItemID)
	w.EnsureWindow(leftoverWindowW, leftoverWindowH)
	w.Window.Open(ctx, w.widgetTree(ctx))
	w.OnLayoutReload(func() {
		w.SetSize(leftoverWindowW, leftoverWindowH)
		w.SetContent(w.widgetTree(w.ctx))
	})
	w.Publish(ctx)
}

func (w *LaphineWindow) SetResult(ctx Context, result int16) {
	if result == 0 {
		w.message = "Done."
	} else {
		w.message = fmt.Sprintf("Failed (%d).", result)
	}
	if w.IsOpen() {
		w.SetContent(w.widgetTree(ctx))
		w.Publish(ctx)
	}
}

func (w *LaphineWindow) Update(ctx Context) bool {
	w.EnsureWindow(leftoverWindowW, leftoverWindowH)
	if !w.IsOpen() {
		return false
	}
	consumed := w.Window.Update(ctx)
	w.Publish(ctx)
	return consumed
}

func (w *LaphineWindow) widgetTree(ctx Context) widget.Widget {
	title := "Laphine"
	if w.upgrade {
		title = "Laphine Upgrade"
	}
	field := rotheme.TextField(w.index, textfield.TypeText, func(v string) { w.index = filterDigits(v, 5) }, func(string) {}, textfield.MaxLength(5))
	return Win(
		Title(title),
		CloseButton(true),
		OnClose(func() {
			if ctx.Network != nil {
				if w.upgrade {
					_ = ctx.Network.SendLaphineUpgClose()
				} else {
					_ = ctx.Network.SendLaphineClose()
				}
			}
			w.Close()
			w.Publish(ctx)
		}),
		Size(float32(leftoverWindowW), float32(leftoverWindowH)),
		Content(primitives.Box(
			rotheme.Label(w.message),
			primitives.HBox(primitives.Box(rotheme.Label("Index")).Width(float32(leftoverLabelW)), primitives.Expanded(field)).Height(float32(leftoverRowH)),
		).Padding(float32(leftoverPad)).Gap(float32(leftoverGap))),
		Footer(rotheme.Button("Make", func() {
			n, _ := strconv.ParseUint(w.index, 10, 16)
			if ctx.Network == nil || n == 0 {
				return
			}
			if w.upgrade {
				_ = ctx.Network.SendLaphineUpg(w.itemID, uint16(n))
			} else {
				_ = ctx.Network.SendLaphineMake(w.itemID, uint16(n), 1)
			}
		})),
	)
}

type StylingWindow struct {
	Window
}

func (w *StylingWindow) Open(ctx Context) {
	w.EnsureWindow(leftoverWindowW, leftoverWindowH)
	w.Window.Open(ctx, w.widgetTree(ctx))
	w.OnLayoutReload(func() {
		w.SetSize(leftoverWindowW, leftoverWindowH)
		w.SetContent(w.widgetTree(w.ctx))
	})
	w.Publish(ctx)
}

func (w *StylingWindow) Update(ctx Context) bool {
	w.EnsureWindow(leftoverWindowW, leftoverWindowH)
	if !w.IsOpen() {
		return false
	}
	consumed := w.Window.Update(ctx)
	w.Publish(ctx)
	return consumed
}

func (w *StylingWindow) widgetTree(ctx Context) widget.Widget {
	return Win(
		Title("Styling Shop"),
		CloseButton(true),
		OnClose(func() {
			if ctx.Network != nil {
				_ = ctx.Network.SendStylingClose()
			}
			w.Close()
			w.Publish(ctx)
		}),
		Size(float32(leftoverWindowW), float32(leftoverWindowH)),
		Content(primitives.Box(rotheme.Label("Close when finished.")).Padding(float32(leftoverPad))),
	)
}
