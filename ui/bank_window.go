package ui

import (
	"fmt"
	"strconv"

	"github.com/gogpu/ui/core/textfield"
	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/widget"
	"github.com/kivutar/goro/ui/rotheme"
)

type BankActionKind uint8

const (
	BankActionNone BankActionKind = iota
	BankActionDeposit
	BankActionWithdraw
	BankActionClose
)

type BankAction struct {
	Kind   BankActionKind
	Amount int32
}

type BankWindow struct {
	Window
	vault   int64
	zeny    int32
	amount  string
	message string
	action  BankAction
}

func (w *BankWindow) Open(ctx Context, vault int64, zeny int32) {
	w.EnsureWindow(bankWindowW, bankWindowH)
	w.vault = vault
	w.zeny = zeny
	w.amount = ""
	w.message = ""
	w.action = BankAction{}
	w.Window.Open(ctx, w.widgetTree(ctx))
	w.OnLayoutReload(func() {
		w.SetSize(bankWindowW, bankWindowH)
		w.SetContent(w.widgetTree(w.ctx))
	})
	w.Publish(ctx)
}

func (w *BankWindow) SetBalances(ctx Context, vault int64, zeny int32, message string) {
	w.vault = vault
	w.zeny = zeny
	w.message = message
	if w.IsOpen() {
		w.SetContent(w.widgetTree(ctx))
		w.Publish(ctx)
	}
}

func (w *BankWindow) Update(ctx Context) bool {
	w.EnsureWindow(bankWindowW, bankWindowH)
	if !w.IsOpen() {
		return false
	}
	consumed := w.Window.Update(ctx)
	w.Publish(ctx)
	return consumed
}

func (w *BankWindow) PopAction() BankAction {
	action := w.action
	w.action = BankAction{}
	return action
}

func (w *BankWindow) widgetTree(ctx Context) widget.Widget {
	return Win(
		Title("Bank"),
		CloseButton(true),
		OnClose(func() {
			w.action = BankAction{Kind: BankActionClose}
			w.Close()
			w.Publish(ctx)
		}),
		Size(float32(bankWindowW), float32(bankWindowH)),
		Content(w.content()),
		Footer(
			rotheme.Button("Deposit", func() { w.submit(BankActionDeposit) }),
			rotheme.Button("Withdraw", func() { w.submit(BankActionWithdraw) }),
		),
	)
}

func (w *BankWindow) content() widget.Widget {
	field := rotheme.TextField(
		w.amount,
		textfield.TypeText,
		func(value string) { w.amount = filterDigits(value, 10) },
		func(string) {},
		textfield.MaxLength(10),
	)
	rows := []widget.Widget{
		rotheme.Label(fmt.Sprintf("Vault: %d", w.vault)),
		rotheme.Label(fmt.Sprintf("Zeny: %d", w.zeny)),
		primitives.HBox(
			primitives.Box(rotheme.Label("Amount")).Width(float32(bankLabelW)).Height(float32(bankFieldH)),
			primitives.Expanded(primitives.Box(field).Height(float32(bankFieldH))),
		).Height(float32(bankFieldH)).Gap(float32(bankGap)),
	}
	if w.message != "" {
		rows = append(rows, rotheme.Label(w.message))
	}
	return primitives.Box(rows...).Padding(float32(bankPad)).Gap(float32(bankGap))
}

func (w *BankWindow) submit(kind BankActionKind) {
	n, err := strconv.ParseInt(w.amount, 10, 32)
	if err != nil || n <= 0 {
		w.message = "Enter an amount."
		return
	}
	w.action = BankAction{Kind: kind, Amount: int32(n)}
}
