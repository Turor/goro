package ui

import (
	"unicode"

	"github.com/gogpu/ui/core/textfield"
	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/widget"
	"github.com/kivutar/goro/ui/rotheme"
)

// theme:ok PIN is always four digits
const pincodeLength = 4

type PincodeMode uint8

const (
	PincodeModeAsk PincodeMode = iota
	PincodeModeSet
	PincodeModeChange
	PincodeModeLocked
)

type PincodeActionKind uint8

const (
	PincodeActionNone PincodeActionKind = iota
	PincodeActionCheck
	PincodeActionSet
	PincodeActionChange
)

type PincodeAction struct {
	Kind PincodeActionKind
	PIN  string
	Old  string
}

type PincodeWindow struct {
	Window
	mode    PincodeMode
	message string
	pin     string
	confirm string
	oldPIN  string
	waiting bool
	action  PincodeAction
}

func (w *PincodeWindow) OpenPrompt(ctx Context, mode PincodeMode, message string) {
	w.EnsureWindow(pincodeWindowW, pincodeWindowH)
	w.mode = mode
	w.message = message
	w.pin = ""
	w.confirm = ""
	w.oldPIN = ""
	w.waiting = false
	w.action = PincodeAction{}
	w.CloseOnEsc = false
	w.Window.Open(ctx, w.widgetTree(ctx))
	w.OnLayoutReload(func() {
		w.SetSize(pincodeWindowW, pincodeWindowH)
		w.SetContent(w.widgetTree(w.ctx))
	})
	w.Publish(ctx)
}

func (w *PincodeWindow) Update(ctx Context) bool {
	w.EnsureWindow(pincodeWindowW, pincodeWindowH)
	if !w.IsOpen() {
		return false
	}
	consumed := w.Window.Update(ctx)
	w.Publish(ctx)
	return consumed
}

func (w *PincodeWindow) PopAction() PincodeAction {
	action := w.action
	w.action = PincodeAction{}
	return action
}

func (w *PincodeWindow) ShowMessage(ctx Context, message string) {
	w.message = message
	w.waiting = false
	w.pin = ""
	w.confirm = ""
	w.oldPIN = ""
	if w.IsOpen() {
		w.SetContent(w.widgetTree(ctx))
		w.Publish(ctx)
	}
}

func (w *PincodeWindow) CloseUnlocked(ctx Context) {
	w.action = PincodeAction{}
	w.waiting = false
	w.Close()
	w.Publish(ctx)
}

func (w *PincodeWindow) widgetTree(ctx Context) widget.Widget {
	return Win(
		Title("PIN Code"),
		CloseButton(false),
		Size(float32(pincodeWindowW), float32(pincodeWindowH)),
		Content(w.content(ctx)),
		Footer(w.footer(ctx)...),
	)
}

func (w *PincodeWindow) content(ctx Context) widget.Widget {
	rows := []widget.Widget{
		rotheme.Label(w.message),
		w.fieldRow("PIN", w.passwordField(ctx, &w.pin)),
	}
	if w.mode == PincodeModeChange {
		rows = append([]widget.Widget{w.fieldRow("Current", w.passwordField(ctx, &w.oldPIN))}, rows...)
	}
	if w.mode == PincodeModeSet || w.mode == PincodeModeChange {
		rows = append(rows, w.fieldRow("Confirm", w.passwordField(ctx, &w.confirm)))
	}
	return primitives.Box(rows...).Padding(float32(pincodePad)).Gap(float32(pincodeGap))
}

func (w *PincodeWindow) footer(ctx Context) []widget.Widget {
	if w.mode == PincodeModeLocked {
		return []widget.Widget{rotheme.Label("PIN locked")}
	}
	return []widget.Widget{
		primitives.Expanded(primitives.Box()),
		rotheme.Button("OK", func() { w.submit(ctx) }),
	}
}

func (w *PincodeWindow) fieldRow(label string, field *textfield.Widget) widget.Widget {
	return primitives.HBox(
		primitives.Box(rotheme.Label(label)).Width(float32(pincodeLabelW)).Height(float32(pincodeFieldH)),
		primitives.Expanded(primitives.Box(field).Height(float32(pincodeFieldH))),
	).Gap(float32(pincodeGap)).Height(float32(pincodeFieldH))
}

func (w *PincodeWindow) passwordField(ctx Context, dest *string) *textfield.Widget {
	return rotheme.TextField(
		*dest,
		textfield.TypePassword,
		func(value string) { *dest = filterDigits(value, pincodeLength) },
		func(string) { w.submit(ctx) },
		textfield.MaxLength(pincodeLength),
	)
}

func (w *PincodeWindow) submit(ctx Context) {
	if w.waiting || w.mode == PincodeModeLocked {
		return
	}
	pin := w.pin
	if len(pin) != pincodeLength {
		w.message = "Enter a 4-digit PIN."
		w.SetContent(w.widgetTree(ctx))
		return
	}
	switch w.mode {
	case PincodeModeAsk:
		w.waiting = true
		w.action = PincodeAction{Kind: PincodeActionCheck, PIN: pin}
	case PincodeModeSet:
		if pin != w.confirm {
			w.message = "PIN confirmation does not match."
			w.SetContent(w.widgetTree(ctx))
			return
		}
		w.waiting = true
		w.action = PincodeAction{Kind: PincodeActionSet, PIN: pin}
	case PincodeModeChange:
		if len(w.oldPIN) != pincodeLength {
			w.message = "Enter the current PIN."
			w.SetContent(w.widgetTree(ctx))
			return
		}
		if pin != w.confirm {
			w.message = "PIN confirmation does not match."
			w.SetContent(w.widgetTree(ctx))
			return
		}
		w.waiting = true
		w.action = PincodeAction{Kind: PincodeActionChange, PIN: pin, Old: w.oldPIN}
	}
}

func filterDigits(s string, max int) string {
	out := make([]rune, 0, max)
	for _, r := range s {
		if unicode.IsDigit(r) {
			out = append(out, r)
			if len(out) == max {
				break
			}
		}
	}
	return string(out)
}
