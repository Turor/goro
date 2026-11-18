package ui

import (
	"fmt"
	"strings"

	"github.com/gogpu/ui/core/scrollview"
	"github.com/gogpu/ui/core/textfield"
	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/state"
	"github.com/gogpu/ui/widget"
	"github.com/kivutar/goro/input"
	"github.com/kivutar/goro/ui/rotheme"
)

type ChatRoomWindowAction struct {
	Message   string
	Leave     bool
	Settings  bool
	MakeOwner string
	Kick      string
}

type chatRoomLine struct {
	text  string
	color widget.Color
}

type ChatRoomWindow struct {
	Window
	RoomID         uint32
	title          string
	limit          uint16
	public         bool
	count          uint16
	owner          string
	localName      string
	selectedMember string
	memberScroll   state.Signal[float32]
	members        []string
	input          string
	inputField     *textfield.Widget
	lines          []chatRoomLine
	scrollY        state.Signal[float32]
	action         ChatRoomWindowAction
}

func (w *ChatRoomWindow) Open(ctx Context, title string, limit uint16, public bool, members []string) {
	title = strings.TrimSpace(title)
	if title == "" {
		return
	}
	w.EnsureWindow(chatRoomWindowW, chatRoomWindowH())
	w.SetSize(chatRoomWindowW, chatRoomWindowH())
	w.SetBackground(chatRoomWindowBg)
	w.SetFullRedraw(true)
	w.ctx = ctx
	w.RoomID = 0
	w.title = title
	w.limit = limit
	w.public = public
	w.members = compactChatRoomMembers(members)
	w.count = uint16(len(w.members))
	w.owner = firstChatRoomMember(w.members)
	w.localName = ""
	if ctx.Session != nil {
		w.localName = ctx.Session.SelectedCharacter().Name
	}
	w.selectedMember = ""
	w.memberScroll = state.NewSignal[float32](0)
	w.input = ""
	w.inputField = nil
	w.lines = nil
	w.scrollY = nil
	w.action = ChatRoomWindowAction{}
	w.Window.Open(ctx, nil)
	w.SetContent(w.widgetTree(ctx))
	w.OnLayoutReload(w.reloadLayout)
	w.Publish(ctx)
	w.focusInput()
}

func (w *ChatRoomWindow) reloadLayout() {
	focused := w.inputField != nil && w.inputField.IsFocused()
	w.SetSize(chatRoomWindowW, chatRoomWindowH())
	w.SetBackground(chatRoomWindowBg)
	w.SetContent(w.widgetTree(w.ctx))
	if focused {
		w.focusInput()
	}
}

func (w *ChatRoomWindow) Update(ctx Context) bool {
	w.EnsureWindow(chatRoomWindowW, chatRoomWindowH())
	w.SetFullRedraw(true)
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

func (w *ChatRoomWindow) Rebind(ctx Context) {
	w.EnsureWindow(chatRoomWindowW, chatRoomWindowH())
	w.SetBackground(chatRoomWindowBg)
	w.SetFullRedraw(true)
	if !w.IsOpen() {
		return
	}
	w.ctx = ctx
	w.inputField = nil
	w.refresh(ctx)
}

func (w *ChatRoomWindow) PopAction() ChatRoomWindowAction {
	action := w.action
	w.action = ChatRoomWindowAction{}
	return action
}

func (w *ChatRoomWindow) IsOwner() bool {
	return w.IsOpen() && w.localName != "" && strings.EqualFold(w.owner, w.localName)
}

func (w *ChatRoomWindow) CanManageMember(name string) bool {
	return w.IsOwner() && name != "" && !strings.EqualFold(name, w.localName) && chatRoomHasMember(w.members, name)
}

func (w *ChatRoomWindow) OpenSettings(ctx Context, form *ChatRoomCreateWindow) {
	if w.IsOwner() {
		form.OpenEdit(ctx, w.title, w.limit, w.public)
	}
}

func (w *ChatRoomWindow) AddMessage(ctx Context, text string) {
	w.addLine(ctx, strings.TrimSpace(text), chatRoomMessageText)
}

func (w *ChatRoomWindow) AddSystem(ctx Context, text string) {
	w.addLine(ctx, strings.TrimSpace(text), chatRoomSystemText)
}

func (w *ChatRoomWindow) AddError(ctx Context, text string) {
	w.addLine(ctx, strings.TrimSpace(text), Color(ErrorTextColor))
}

func (w *ChatRoomWindow) SetMembers(ctx Context, members []string, owner string) {
	w.members = compactChatRoomMembers(members)
	w.count = uint16(len(w.members))
	w.owner = strings.TrimSpace(owner)
	if w.owner == "" {
		w.owner = firstChatRoomMember(w.members)
	}
	if w.IsOpen() {
		w.refresh(ctx)
	}
}

func (w *ChatRoomWindow) AddMember(ctx Context, name string, count uint16) {
	name = strings.TrimSpace(name)
	if name == "" {
		return
	}
	if !chatRoomHasMember(w.members, name) {
		w.members = append(w.members, name)
	}
	w.count = count
	if w.count == 0 {
		w.count = uint16(len(w.members))
	}
	if w.owner == "" {
		w.owner = firstChatRoomMember(w.members)
	}
	if w.IsOpen() {
		w.refresh(ctx)
	}
}

func (w *ChatRoomWindow) RemoveMember(ctx Context, name string, count uint16) {
	name = strings.TrimSpace(name)
	if name == "" {
		return
	}
	next := w.members[:0]
	for _, member := range w.members {
		if !strings.EqualFold(member, name) {
			next = append(next, member)
		}
	}
	w.members = next
	w.count = count
	if w.count == 0 {
		w.count = uint16(len(w.members))
	}
	if strings.EqualFold(w.owner, name) {
		w.owner = ""
	}
	if w.IsOpen() {
		w.refresh(ctx)
	}
}

func (w *ChatRoomWindow) SetOwner(ctx Context, name string) {
	name = strings.TrimSpace(name)
	if name == "" {
		return
	}
	w.owner = name
	if !chatRoomHasMember(w.members, name) {
		w.members = append([]string{name}, w.members...)
		w.count = uint16(len(w.members))
	}
	if w.IsOpen() {
		w.refresh(ctx)
	}
}

func (w *ChatRoomWindow) UpdateRoom(ctx Context, title string, limit uint16, count uint16, public bool) {
	title = strings.TrimSpace(title)
	if title != "" {
		w.title = title
	}
	w.limit = limit
	w.count = count
	w.public = public
	if w.IsOpen() {
		w.refresh(ctx)
	}
}

func (w *ChatRoomWindow) addLine(ctx Context, text string, color widget.Color) {
	if text == "" {
		return
	}
	w.lines = append(w.lines, chatRoomLine{text: text, color: color})
	if len(w.lines) > 80 {
		copy(w.lines, w.lines[len(w.lines)-80:]) // theme:ok message history cap
		w.lines = w.lines[:80]
	}
	if w.IsOpen() {
		w.refresh(ctx)
	}
}

func (w *ChatRoomWindow) widgetTree(ctx Context) widget.Widget {
	return Win(
		Title(w.windowTitle()),
		CloseButton(true),
		OnClose(func() { w.requestLeave(ctx) }),
		Size(float32(chatRoomWindowW), float32(chatRoomWindowH())),
		Background(chatRoomWindowBg),
		Content(primitives.HBox(
			primitives.Expanded(w.contentTree()),
			w.memberList(),
		).CrossAlign(primitives.CrossAxisStretch)),
		Footer(w.footerTree(ctx)...),
	)
}

func (w *ChatRoomWindow) contentTree() widget.Widget {
	messageHeight := chatRoomWindowContentH - chatRoomSummaryReserve - chatRoomMessagePadY*2
	textWidth := maxInt(1, scrollbarSafeIntWidth(chatRoomWindowW-chatRoomMembersW-2*chatRoomMessagePadX)-chatRoomTextInset)
	lines := wrapChatRoomLines(w.visibleLines(), textWidth)
	contentHeight := whisperMessageContentHeight(len(lines))
	rows := make([]widget.Widget, 0, len(lines)+1)
	if spacerH := messageHeight - contentHeight - chatRoomMessageGap; spacerH > 0 {
		rows = append(rows, primitives.Box().Height(float32(spacerH)))
	}
	for _, line := range lines {
		rows = append(rows,
			primitives.Box(
				rotheme.Text(line.text).
					Color(line.color).
					MaxLines(1),
			).Height(float32(consoleLineH)),
		)
	}
	messageList := primitives.Box(rows...).
		Gap(float32(chatRoomMessageGap)).
		CrossAlign(primitives.CrossAxisStretch)
	w.ensureScrollSignal().Set(consoleBottomScrollY(len(lines), messageHeight))
	return primitives.Box(
		primitives.Box(w.memberSummary()).
			Height(float32(chatRoomSummaryH)).
			PaddingXY(float32(chatRoomMessagePadX), float32(chatRoomSummaryPadY)).
			CrossAlign(primitives.CrossAxisStretch),
		primitives.Box(
			scrollview.New(
				primitives.Box(messageList).
					PaddingRight(ROScrollbarGutter),
				scrollview.ScrollbarOpt(scrollview.ScrollbarAuto),
				scrollview.ScrollYSignal(w.ensureScrollSignal()),
				scrollview.ScrollStep(float32(consoleLineH*3)),
			),
		).
			Height(float32(messageHeight+chatRoomMessagePadY*2)).
			PaddingXY(float32(chatRoomMessagePadX), float32(chatRoomMessagePadY)).
			CrossAlign(primitives.CrossAxisStretch),
	).
		Background(chatRoomPanelBg).
		BorderStyle(1, chatRoomPanelBorder).
		CrossAlign(primitives.CrossAxisStretch)
}

func (w *ChatRoomWindow) memberSummary() widget.Widget {
	count := int(w.count)
	if count <= 0 {
		count = len(w.members)
	}
	limit := int(w.limit)
	if limit <= 0 {
		limit = count
	}
	names := chatRoomMemberListText(w.members, w.owner)
	if names == "" {
		names = "No members"
	}
	text := fmt.Sprintf("%d/%d  %s", count, limit, names)
	return rotheme.Text(trimRunes(text, 58)).
		Color(chatRoomSummaryText).
		MaxLines(1).
		Ellipsis()
}

func (w *ChatRoomWindow) footerTree(ctx Context) []widget.Widget {
	buttons := []widget.Widget{
		primitives.Expanded(
			primitives.Box(w.inputWidget(ctx)).
				Height(float32(chatRoomInputH)).
				CrossAlign(primitives.CrossAxisStretch),
		),
		rotheme.Button("Send", func() {
			w.submit(ctx)
		}),
	}
	if w.IsOwner() {
		buttons = append(buttons, rotheme.Button("Settings", func() {
			if w.IsOwner() {
				w.action = ChatRoomWindowAction{Settings: true}
			}
		}))
	}
	return append(buttons, rotheme.Button("Leave", func() {
		w.requestLeave(ctx)
	}))
}

func (w *ChatRoomWindow) memberList() widget.Widget {
	members := append([]string(nil), w.members...)
	rows := make([]itemTableRow, len(members))
	selected := -1
	for i, name := range members {
		rows[i].name = name
		if strings.EqualFold(name, w.owner) {
			rows[i].name = "* " + name
		}
		if name == w.selectedMember {
			selected = i
		}
	}
	list := itemTableView(rows, "Members", float32(chatRoomMemberRowH), float32(chatRoomMemberHeaderH), "No members", w.memberScroll, selected, func(row int) {
		w.selectedMember = members[row]
	})
	children := []widget.Widget{primitives.Expanded(list)}
	if w.IsOwner() {
		children = append(children, primitives.HBox(
			rotheme.ButtonDisabledFn("Make Owner", func() bool { return !w.CanManageMember(w.selectedMember) }, func() {
				if w.CanManageMember(w.selectedMember) {
					w.action = ChatRoomWindowAction{MakeOwner: w.selectedMember}
				}
			}),
			rotheme.ButtonDisabledFn("Kick", func() bool { return !w.CanManageMember(w.selectedMember) }, func() {
				if w.CanManageMember(w.selectedMember) {
					w.action = ChatRoomWindowAction{Kick: w.selectedMember}
				}
			}),
		).Gap(float32(chatRoomOwnerGap)).Padding(float32(chatRoomOwnerPad)).Height(float32(chatRoomOwnerH)).CrossAlign(primitives.CrossAxisCenter))
	}
	return primitives.Box(children...).Width(float32(chatRoomMembersW)).
		Background(rotheme.Default.Colors.PanelBody).CrossAlign(primitives.CrossAxisStretch)
}

func (w *ChatRoomWindow) inputWidget(ctx Context) *textfield.Widget {
	if w.inputField != nil {
		return w.inputField
	}
	w.inputField = rotheme.TextField(
		w.input,
		textfield.TypeText,
		func(value string) {
			w.input = value
		},
		func(string) {
			w.submit(ctx)
		},
		textfield.MaxLength(100),
		textfield.Placeholder("Message"),
	)
	w.focusInput()
	return w.inputField
}

func (w *ChatRoomWindow) submit(ctx Context) {
	message := strings.TrimSpace(w.input)
	if !w.IsOpen() || message == "" {
		return
	}
	w.action = ChatRoomWindowAction{Message: message}
	w.input = ""
	if w.inputField != nil {
		w.inputField.SetText("")
	}
	w.focusInput()
	w.refresh(ctx)
}

func (w *ChatRoomWindow) requestLeave(ctx Context) {
	if !w.IsOpen() {
		return
	}
	w.action = ChatRoomWindowAction{Leave: true}
	w.Close()
	w.Publish(ctx)
}

func (w *ChatRoomWindow) submitFromFocusedEnter(ctx Context) bool {
	if ctx.Input == nil || w.inputField == nil || !w.inputField.IsFocused() {
		return false
	}
	if !ctx.Input.JustPressed(input.KeyEnter) {
		return false
	}
	w.submit(ctx)
	return true
}

func (w *ChatRoomWindow) refresh(ctx Context) {
	focused := w.inputField != nil && w.inputField.IsFocused()
	w.SetContent(w.widgetTree(ctx))
	w.Publish(ctx)
	if focused {
		w.focusInput()
	}
	if ctx.UIApp != nil {
		ctx.UIApp.Invalidate()
	}
}

func (w *ChatRoomWindow) focusInput() {
	if w.inputField != nil {
		if wc := windowWidgetContext(w.ctx); wc != nil {
			wc.RequestFocus(w.inputField)
		} else {
			w.inputField.SetFocused(true)
		}
	}
}

func (w *ChatRoomWindow) ensureScrollSignal() state.Signal[float32] {
	if w.scrollY == nil {
		w.scrollY = state.NewSignal[float32](0)
	}
	return w.scrollY
}

func (w *ChatRoomWindow) windowTitle() string {
	privacy := "Public"
	if !w.public {
		privacy = "Private"
	}
	return fmt.Sprintf("%s - %s", trimRunes(w.title, 24), privacy)
}

func (w *ChatRoomWindow) visibleLines() []chatRoomLine {
	if len(w.lines) == 0 {
		return []chatRoomLine{{text: "No messages", color: chatRoomEmptyText}}
	}
	return w.lines
}

func compactChatRoomMembers(members []string) []string {
	out := make([]string, 0, len(members))
	for _, member := range members {
		member = strings.TrimSpace(member)
		if member != "" {
			out = append(out, member)
		}
	}
	return out
}

func firstChatRoomMember(members []string) string {
	if len(members) == 0 {
		return ""
	}
	return strings.TrimSpace(members[0])
}

func chatRoomHasMember(members []string, name string) bool {
	for _, member := range members {
		if strings.EqualFold(member, name) {
			return true
		}
	}
	return false
}

func chatRoomMemberListText(members []string, owner string) string {
	owner = strings.TrimSpace(owner)
	labels := make([]string, 0, len(members))
	for _, member := range members {
		member = strings.TrimSpace(member)
		if member == "" {
			continue
		}
		if owner != "" && strings.EqualFold(member, owner) {
			member += " (owner)"
		}
		labels = append(labels, member)
	}
	return strings.Join(labels, ", ")
}
