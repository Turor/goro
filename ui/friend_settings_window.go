package ui

import (
	"github.com/gogpu/ui/core/checkbox"
	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/widget"
	"github.com/kivutar/goro/session"
	"github.com/kivutar/goro/ui/rotheme"
)

type FriendSettingsWindow struct {
	Window
	settings session.WhisperSettings
}

func (w *FriendSettingsWindow) Open(ctx Context) {
	w.EnsureWindow(friendSettingsW, friendSettingsWindowH())
	w.SetSize(friendSettingsW, friendSettingsWindowH())
	w.ctx = ctx
	w.settings = friendSettings(ctx.Session)
	w.Window.Open(ctx, w.widgetTree(ctx))
	w.OnLayoutReload(w.reloadLayout)
	w.Publish(ctx)
}

func (w *FriendSettingsWindow) reloadLayout() {
	w.SetSize(friendSettingsW, friendSettingsWindowH())
	w.SetContent(w.widgetTree(w.ctx))
}

func (w *FriendSettingsWindow) Update(ctx Context) bool {
	w.EnsureWindow(friendSettingsW, friendSettingsWindowH())
	w.ctx = ctx
	if !w.IsOpen() {
		return false
	}
	consumed := w.Window.Update(ctx)
	w.Publish(ctx)
	return consumed
}

func (w *FriendSettingsWindow) Rebind(ctx Context) {
	if !w.IsOpen() {
		return
	}
	w.ctx = ctx
	w.RebindContent(ctx, w.widgetTree(ctx))
}

func (w *FriendSettingsWindow) widgetTree(ctx Context) widget.Widget {
	return Win(
		Title("Friend Setup"),
		CloseButton(true),
		OnClose(w.Close),
		Size(float32(friendSettingsW), float32(friendSettingsWindowH())),
		Content(
			primitives.Box(
				rotheme.Checkbox(
					checkbox.Checked(w.settings.OpenStrangers),
					checkbox.LabelOpt("1:1 Chat from Strangers"),
					checkbox.OnToggle(func(enabled bool) {
						w.settings.OpenStrangers = enabled
						w.settings.Configured = true
					}),
				),
				rotheme.Checkbox(
					checkbox.Checked(w.settings.OpenFriends),
					checkbox.LabelOpt("1:1 Chat from Friends"),
					checkbox.OnToggle(func(enabled bool) {
						w.settings.OpenFriends = enabled
						w.settings.Configured = true
					}),
				),
				rotheme.Checkbox(
					checkbox.Checked(w.settings.Alert),
					checkbox.LabelOpt("1:1 Chat Alert"),
					checkbox.OnToggle(func(enabled bool) {
						w.settings.Alert = enabled
						w.settings.Configured = true
					}),
				),
			).
				Padding(float32(friendSettingsPad)).
				Gap(float32(friendSettingsGap)),
		),
		Footer(
			primitives.Expanded(primitives.Box()),
			rotheme.Button("OK", func() {
				w.apply(ctx)
			}),
			rotheme.Button("Cancel", w.Close),
		),
	)
}

func (w *FriendSettingsWindow) apply(ctx Context) {
	if ctx.Session != nil {
		w.settings.Configured = true
		ctx.Session.Whisper = w.settings
	}
	w.Close()
}

func friendSettings(s *session.Session) session.WhisperSettings {
	if s == nil || !s.Whisper.Configured {
		return session.DefaultWhisperSettings()
	}
	return s.Whisper
}
