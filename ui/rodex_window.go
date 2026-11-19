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

type RodexWindow struct {
	Window
	unread   bool
	mail     []network.RodexMailHeader
	selected network.RodexMailHeader
	body     string
	zeny     int64
	items    int
	compose  bool
	to       string
	title    string
	text     string
	sendZeny string
	attach   string
	message  string
	sender   string
}

func (w *RodexWindow) SetUnread(show bool) { w.unread = show }

func (w *RodexWindow) Unread() bool { return w.unread }

func (w *RodexWindow) OpenMailbox(ctx Context) {
	w.EnsureWindow(rodexWindowW, rodexWindowH)
	w.compose = false
	w.message = ""
	w.mail = nil
	w.selected = network.RodexMailHeader{}
	w.body = ""
	w.Window.Open(ctx, w.widgetTree(ctx))
	w.OnLayoutReload(func() {
		w.SetSize(rodexWindowW, rodexWindowH)
		w.SetContent(w.widgetTree(w.ctx))
	})
	if ctx.Network != nil {
		_ = ctx.Network.SendRodexOpen()
	}
	w.Publish(ctx)
}

func (w *RodexWindow) Toggle(ctx Context) {
	if w.IsOpen() {
		if ctx.Network != nil {
			_ = ctx.Network.SendRodexClose()
		}
		w.Close()
		w.Publish(ctx)
		return
	}
	w.OpenMailbox(ctx)
}

func (w *RodexWindow) SetList(ctx Context, list network.RodexMailList) {
	w.mail = append(w.mail, list.Mail...)
	if w.IsOpen() {
		w.SetContent(w.widgetTree(ctx))
		w.Publish(ctx)
	}
}

func (w *RodexWindow) SetRead(ctx Context, msg network.RodexRead) {
	w.body = msg.Body
	w.zeny = msg.Zeny
	w.items = msg.Items
	w.compose = false
	if w.IsOpen() {
		w.SetContent(w.widgetTree(ctx))
		w.Publish(ctx)
	}
}

func (w *RodexWindow) SetMessage(ctx Context, message string) {
	w.message = message
	if w.IsOpen() {
		w.SetContent(w.widgetTree(ctx))
		w.Publish(ctx)
	}
}

func (w *RodexWindow) Update(ctx Context) bool {
	w.EnsureWindow(rodexWindowW, rodexWindowH)
	if !w.IsOpen() {
		return false
	}
	if ctx.Session != nil {
		w.sender = ctx.Session.SelectedCharacter().Name
	}
	consumed := w.Window.Update(ctx)
	w.Publish(ctx)
	return consumed
}

func (w *RodexWindow) widgetTree(ctx Context) widget.Widget {
	title := "RoDEX"
	if w.unread {
		title = "RoDEX *"
	}
	return Win(
		Title(title),
		CloseButton(true),
		OnClose(func() {
			if ctx.Network != nil {
				_ = ctx.Network.SendRodexClose()
			}
			w.Close()
			w.Publish(ctx)
		}),
		Size(float32(rodexWindowW), float32(rodexWindowH)),
		Content(w.content(ctx)),
		Footer(w.footer(ctx)...),
	)
}

func (w *RodexWindow) content(ctx Context) widget.Widget {
	if w.compose {
		return w.composeForm()
	}
	rows := []widget.Widget{}
	if w.message != "" {
		rows = append(rows, rotheme.Label(w.message))
	}
	for _, mail := range w.mail {
		mail := mail
		label := fmt.Sprintf("%s — %s", mail.Sender, mail.Title)
		rows = append(rows, rotheme.Button(label, func() {
			w.selected = mail
			if ctx.Network != nil {
				_ = ctx.Network.SendRodexRead(mail.OpenType, mail.MailID)
			}
		}).Height(float32(rodexRowH)))
	}
	if w.selected.MailID != 0 {
		rows = append(rows,
			rotheme.Label(fmt.Sprintf("From %s", w.selected.Sender)),
			rotheme.Label(w.body),
			rotheme.Label(fmt.Sprintf("Zeny %d  Items %d", w.zeny, w.items)),
		)
	}
	return primitives.Box(rows...).Padding(float32(rodexPad)).Gap(float32(rodexGap))
}

func (w *RodexWindow) composeForm() widget.Widget {
	to := rotheme.TextField(w.to, textfield.TypeText, func(v string) { w.to = v }, func(string) {}, textfield.MaxLength(24))
	title := rotheme.TextField(w.title, textfield.TypeText, func(v string) { w.title = v }, func(string) {}, textfield.MaxLength(64))
	body := rotheme.TextField(w.text, textfield.TypeText, func(v string) { w.text = v }, func(string) {})
	zeny := rotheme.TextField(w.sendZeny, textfield.TypeText, func(v string) { w.sendZeny = filterDigits(v, 10) }, func(string) {}, textfield.MaxLength(10))
	attach := rotheme.TextField(w.attach, textfield.TypeText, func(v string) { w.attach = filterDigits(v, 5) }, func(string) {}, textfield.MaxLength(5))
	return primitives.Box(
		primitives.HBox(primitives.Box(rotheme.Label("To")).Width(float32(rodexLabelW)), primitives.Expanded(to)).Height(float32(rodexRowH)),
		primitives.HBox(primitives.Box(rotheme.Label("Title")).Width(float32(rodexLabelW)), primitives.Expanded(title)).Height(float32(rodexRowH)),
		primitives.HBox(primitives.Box(rotheme.Label("Zeny")).Width(float32(rodexLabelW)), primitives.Expanded(zeny)).Height(float32(rodexRowH)),
		primitives.HBox(primitives.Box(rotheme.Label("Item #")).Width(float32(rodexLabelW)), primitives.Expanded(attach)).Height(float32(rodexRowH)),
		primitives.Expanded(body),
	).Padding(float32(rodexPad)).Gap(float32(rodexGap))
}

func (w *RodexWindow) footer(ctx Context) []widget.Widget {
	if w.compose {
		return []widget.Widget{
			rotheme.Button("Cancel", func() {
				w.compose = false
				w.SetContent(w.widgetTree(ctx))
			}),
			rotheme.Button("Check", func() {
				if ctx.Network != nil {
					_ = ctx.Network.SendCheckName(w.to)
				}
			}),
			rotheme.Button("Attach", func() {
				n, _ := strconv.ParseUint(w.attach, 10, 16)
				if ctx.Network != nil && n > 0 {
					_ = ctx.Network.SendRodexAddItem(uint16(n), 1)
				}
			}),
			rotheme.Button("Send", func() {
				n, _ := strconv.ParseInt(w.sendZeny, 10, 64)
				if ctx.Network != nil {
					_ = ctx.Network.SendRodexSend(w.to, w.sender, w.title, w.text, n, 0)
				}
			}),
		}
	}
	return []widget.Widget{
		rotheme.Button("New", func() {
			w.compose = true
			w.SetContent(w.widgetTree(ctx))
		}),
		rotheme.Button("Claim", func() {
			if ctx.Network == nil || w.selected.MailID == 0 {
				return
			}
			_ = ctx.Network.SendRodexGetZeny(w.selected.OpenType, w.selected.MailID)
			_ = ctx.Network.SendRodexGetItems(w.selected.OpenType, w.selected.MailID)
		}),
		rotheme.Button("Delete", func() {
			if ctx.Network != nil && w.selected.MailID != 0 {
				_ = ctx.Network.SendRodexDelete(w.selected.OpenType, w.selected.MailID)
			}
		}),
	}
}
