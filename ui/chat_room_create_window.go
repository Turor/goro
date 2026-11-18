package ui

import (
	"strconv"
	"strings"

	"github.com/gogpu/ui/core/radio"
	"github.com/gogpu/ui/core/textfield"
	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/widget"
	"github.com/kivutar/goro/input"
	"github.com/kivutar/goro/ui/rotheme"
)

type ChatRoomCreateWindowAction struct {
	Edit     bool
	Title    string
	Password string
	Limit    uint16
	Public   bool
}

type ChatRoomCreateWindow struct {
	Window
	edit          bool
	title         string
	password      string
	limit         string
	public        bool
	titleField    *textfield.Widget
	passwordField *textfield.Widget
	limitField    *textfield.Widget
	action        ChatRoomCreateWindowAction
}

func (w *ChatRoomCreateWindow) Open(ctx Context) {
	w.openForm(ctx, ChatRoomCreateWindowAction{Limit: 20, Public: true})
}

func (w *ChatRoomCreateWindow) OpenEdit(ctx Context, title string, limit uint16, public bool) {
	w.openForm(ctx, ChatRoomCreateWindowAction{Edit: true, Title: title, Limit: limit, Public: public})
}

func (w *ChatRoomCreateWindow) Editing() bool { return w.IsOpen() && w.edit }

func (w *ChatRoomCreateWindow) openForm(ctx Context, room ChatRoomCreateWindowAction) {
	w.EnsureWindow(chatRoomCreateW, chatRoomCreateWindowH())
	w.SetSize(chatRoomCreateW, chatRoomCreateWindowH())
	w.ctx = ctx
	w.edit = room.Edit
	w.title = room.Title
	w.password = ""
	w.limit = strconv.Itoa(int(room.Limit))
	w.public = room.Public
	w.titleField = nil
	w.passwordField = nil
	w.limitField = nil
	w.action = ChatRoomCreateWindowAction{}
	w.Window.Open(ctx, w.widgetTree(ctx))
	w.OnLayoutReload(w.reloadLayout)
	w.Publish(ctx)
	w.focusTitle()
}

func (w *ChatRoomCreateWindow) reloadLayout() {
	var focused *textfield.Widget
	switch {
	case w.titleField != nil && w.titleField.IsFocused():
		focused = w.titleField
	case w.passwordField != nil && w.passwordField.IsFocused():
		focused = w.passwordField
	case w.limitField != nil && w.limitField.IsFocused():
		focused = w.limitField
	}
	w.SetSize(chatRoomCreateW, chatRoomCreateWindowH())
	w.SetContent(w.widgetTree(w.ctx))
	if focused == nil {
		return
	}
	if wc := windowWidgetContext(w.ctx); wc != nil {
		wc.RequestFocus(focused)
	} else {
		focused.SetFocused(true)
	}
}

func (w *ChatRoomCreateWindow) Update(ctx Context) bool {
	w.EnsureWindow(chatRoomCreateW, chatRoomCreateWindowH())
	w.ctx = ctx
	if !w.IsOpen() {
		return false
	}
	if w.submitFromFocusedEnter(ctx) {
		w.Publish(ctx)
		return true
	}
	consumed := w.Window.Update(ctx)
	w.Publish(ctx)
	return consumed
}

func (w *ChatRoomCreateWindow) Rebind(ctx Context) {
	if !w.IsOpen() {
		return
	}
	w.ctx = ctx
	w.titleField = nil
	w.passwordField = nil
	w.limitField = nil
	w.RebindContent(ctx, w.widgetTree(ctx))
	w.focusTitle()
	w.Publish(ctx)
}

func (w *ChatRoomCreateWindow) PopAction() ChatRoomCreateWindowAction {
	action := w.action
	w.action = ChatRoomCreateWindowAction{}
	return action
}

func (w *ChatRoomCreateWindow) widgetTree(ctx Context) widget.Widget {
	title, passwordLabel := "Make a Room", "Password"
	if w.edit {
		title, passwordLabel = "Room Settings", "Password (blank clears it)"
	}
	return Win(
		Title(title),
		CloseButton(true),
		OnClose(w.Close),
		Size(float32(chatRoomCreateW), float32(chatRoomCreateWindowH())),
		Content(
			primitives.Box(
				rotheme.Label("Title"),
				primitives.Box(w.titleInput(ctx)).
					Height(float32(chatRoomCreateFieldH)).
					CrossAlign(primitives.CrossAxisStretch),
				primitives.HBox(
					primitives.Box(
						rotheme.Label("Limit"),
						primitives.Box(w.limitInput(ctx)).
							Height(float32(chatRoomCreateFieldH)).
							CrossAlign(primitives.CrossAxisStretch),
					).
						Width(float32(chatRoomCreateLimitW)).
						Gap(float32(chatRoomCreateLabelGap)),
					primitives.Box(
						rotheme.Label("Type"),
						primitives.Box(
							rotheme.Radio(
								radio.Items(
									radio.ItemDef{Value: "public", Label: "Public"},
									radio.ItemDef{Value: "private", Label: "Private"},
								),
								radio.DirectionOpt(radio.Horizontal),
								radio.Selected(chatRoomTypeValue(w.public)),
								radio.OnChange(func(value string) {
									w.public = value != "private"
								}),
							),
						).Height(float32(chatRoomCreateFieldH)),
					).
						Gap(float32(chatRoomCreateLabelGap)),
				).
					Gap(float32(chatRoomCreateColumnGap)),
				rotheme.Label(passwordLabel),
				primitives.Box(w.passwordInput(ctx)).
					Height(float32(chatRoomCreateFieldH)).
					CrossAlign(primitives.CrossAxisStretch),
			).
				Padding(float32(chatRoomCreatePad)).
				Gap(float32(chatRoomCreateGap)),
		),
		Footer(
			primitives.Expanded(primitives.Box()),
			rotheme.Button("OK", func() {
				w.submit(ctx)
			}),
			rotheme.Button("Cancel", w.Close),
		),
	)
}

func (w *ChatRoomCreateWindow) titleInput(ctx Context) *textfield.Widget {
	if w.titleField != nil {
		return w.titleField
	}
	w.titleField = rotheme.TextField(
		w.title,
		textfield.TypeText,
		func(value string) {
			w.title = value
		},
		func(string) {
			w.submit(ctx)
		},
		textfield.MaxLength(32),
		textfield.Placeholder("Room Title"),
	)
	return w.titleField
}

func (w *ChatRoomCreateWindow) passwordInput(ctx Context) *textfield.Widget {
	if w.passwordField != nil {
		return w.passwordField
	}
	w.passwordField = rotheme.TextField(
		w.password,
		textfield.TypePassword,
		func(value string) {
			w.password = value
		},
		func(string) {
			w.submit(ctx)
		},
		textfield.MaxLength(8),
		textfield.Placeholder("Optional"),
	)
	return w.passwordField
}

func (w *ChatRoomCreateWindow) limitInput(ctx Context) *textfield.Widget {
	if w.limitField != nil {
		return w.limitField
	}
	w.limitField = rotheme.TextField(
		w.limit,
		textfield.TypeNumber,
		func(value string) {
			w.limit = value
		},
		func(string) {
			w.submit(ctx)
		},
		textfield.MaxLength(2),
	)
	return w.limitField
}

func (w *ChatRoomCreateWindow) submit(ctx Context) {
	title := strings.TrimSpace(w.title)
	if !w.IsOpen() || title == "" {
		return
	}
	w.action = ChatRoomCreateWindowAction{
		Edit:     w.edit,
		Title:    title,
		Password: w.password,
		Limit:    chatRoomLimit(w.limit),
		Public:   w.public,
	}
	w.Close()
	w.Publish(ctx)
}

func (w *ChatRoomCreateWindow) submitFromFocusedEnter(ctx Context) bool {
	if ctx.Input == nil {
		return false
	}
	focused := w.titleField != nil && w.titleField.IsFocused() ||
		w.passwordField != nil && w.passwordField.IsFocused() ||
		w.limitField != nil && w.limitField.IsFocused()
	if !focused || !ctx.Input.JustPressed(input.KeyEnter) {
		return false
	}
	w.submit(ctx)
	return true
}

func (w *ChatRoomCreateWindow) focusTitle() {
	if w.titleField != nil {
		if wc := windowWidgetContext(w.ctx); wc != nil {
			wc.RequestFocus(w.titleField)
		} else {
			w.titleField.SetFocused(true)
		}
	}
}

func chatRoomTypeValue(public bool) string {
	if public {
		return "public"
	}
	return "private"
}

func chatRoomLimit(value string) uint16 {
	n, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil {
		return 20
	}
	if n < 2 {
		return 2
	}
	if n > 20 {
		return 20
	}
	return uint16(n)
}
