package ui

type PartyInviteWindow struct {
	TextPromptWindow
}

func (w *PartyInviteWindow) Open(ctx Context) {
	w.EnsureWindow(partyInviteW, partyInviteHeight())
	w.TextPromptWindow.Open(ctx, "Party Invitation", "Player Name", "Player Name", 23)
	w.claimLayoutReload(func() {
		w.SetSize(partyInviteW, partyInviteHeight())
		w.SetContent(w.widgetTree(w.ctx))
		w.focusInput(w.ctx)
	})
}

func (w *PartyInviteWindow) PopAction() string {
	return w.TextPromptWindow.PopAction().Text
}
