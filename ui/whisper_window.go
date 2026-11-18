package ui

import (
	"fmt"
	"github.com/kivutar/goro/input"
	"strings"

	"github.com/gogpu/ui/core/scrollview"
	"github.com/gogpu/ui/core/textfield"
	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/state"
	"github.com/gogpu/ui/widget"
	"github.com/kivutar/goro/ui/rotheme"
)

type WhisperWindowAction struct {
	Target  string
	Message string
}

type whisperWindowLine struct {
	text  string
	color widget.Color
}

type WhisperWindow struct {
	Window
	target     string
	input      string
	inputField *textfield.Widget
	lines      []whisperWindowLine
	scrollY    state.Signal[float32]
	action     WhisperWindowAction
}

func (w *WhisperWindow) Open(ctx Context, target string) {
	w.open(ctx, target)
	w.focusInput()
}

func (w *WhisperWindow) open(ctx Context, target string) {
	target = strings.TrimSpace(target)
	if target == "" {
		return
	}
	w.EnsureWindow(whisperWindowW, whisperWindowH())
	w.SetSize(whisperWindowW, whisperWindowH())
	w.SetBackground(whisperWindowBg)
	w.SetFullRedraw(true)
	w.ctx = ctx
	if !strings.EqualFold(w.target, target) {
		w.target = target
		w.lines = nil
		w.input = ""
		w.inputField = nil
	}
	w.Window.Open(ctx, w.widgetTree(ctx))
	w.OnLayoutReload(w.reloadLayout)
	w.Publish(ctx)
}

func (w *WhisperWindow) reloadLayout() {
	focused := w.inputField != nil && w.inputField.IsFocused()
	w.SetSize(whisperWindowW, whisperWindowH())
	w.SetBackground(whisperWindowBg)
	w.SetContent(w.widgetTree(w.ctx))
	if focused {
		w.focusInput()
	}
}

func (w *WhisperWindow) Update(ctx Context) bool {
	w.EnsureWindow(whisperWindowW, whisperWindowH())
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

func (w *WhisperWindow) Rebind(ctx Context) {
	w.EnsureWindow(whisperWindowW, whisperWindowH())
	w.SetBackground(whisperWindowBg)
	w.SetFullRedraw(true)
	if !w.IsOpen() {
		return
	}
	w.ctx = ctx
	focused := w.inputField != nil && w.inputField.IsFocused()
	releaseWindowFocus(ctx, w.content)
	w.inputField = nil
	w.SetContent(w.widgetTree(ctx))
	w.Publish(ctx)
	if focused {
		w.focusInput()
	}
}

func (w *WhisperWindow) PopAction() WhisperWindowAction {
	action := w.action
	w.action = WhisperWindowAction{}
	return action
}

func (w *WhisperWindow) AddIncoming(ctx Context, sender, message string) {
	w.addLine(ctx, fmt.Sprintf("%s : %s", strings.TrimSpace(sender), strings.TrimSpace(message)), whisperIncomingText)
}

func (w *WhisperWindow) AddOutgoing(ctx Context, message string) {
	w.addLine(ctx, fmt.Sprintf("me : %s", strings.TrimSpace(message)), whisperOutgoingText)
}

func (w *WhisperWindow) AddError(ctx Context, message string) {
	w.addLine(ctx, strings.TrimSpace(message), Color(ErrorTextColor))
}

func (w *WhisperWindow) addLine(ctx Context, text string, color widget.Color) {
	text = strings.TrimSpace(text)
	if text == "" {
		return
	}
	w.lines = append(w.lines, whisperWindowLine{text: text, color: color})
	if len(w.lines) > 60 {
		copy(w.lines, w.lines[len(w.lines)-60:]) // theme:ok message history cap
		w.lines = w.lines[:60]
	}
	if w.IsOpen() {
		w.refresh(ctx)
	}
}

func (w *WhisperWindow) widgetTree(ctx Context) widget.Widget {
	return Win(
		Title(w.title()),
		CloseButton(true),
		OnClose(w.Close),
		Size(float32(whisperWindowW), float32(whisperWindowH())),
		Background(whisperWindowBg),
		Content(w.contentTree()),
		Footer(w.footerTree(ctx)...),
	)
}

func (w *WhisperWindow) contentTree() widget.Widget {
	lines := w.visibleLines()
	messageHeight := whisperWindowContentH - whisperMessagePadY*2
	contentHeight := whisperMessageContentHeight(len(lines))
	rows := make([]widget.Widget, 0, len(lines)+1)
	if spacerH := messageHeight - contentHeight; spacerH > 0 {
		rows = append(rows, primitives.Box().Height(float32(spacerH)))
	}
	for _, line := range lines {
		rows = append(rows,
			primitives.Box(
				rotheme.Text(trimRunes(line.text, 46)).
					Color(line.color).
					MaxLines(1).
					Ellipsis(),
			).Height(float32(consoleLineH)),
		)
	}
	messageList := primitives.Box(rows...).
		Gap(1).
		CrossAlign(primitives.CrossAxisStretch)
	w.ensureScrollSignal().Set(consoleBottomScrollY(len(lines), messageHeight))
	return primitives.Box(
		scrollview.New(
			primitives.Box(messageList).
				PaddingRight(float32(ROScrollbarGutter)),
			scrollview.ScrollbarOpt(scrollview.ScrollbarAuto),
			scrollview.ScrollYSignal(w.ensureScrollSignal()),
			scrollview.ScrollStep(float32(consoleLineH*3)),
		),
	).
		PaddingXY(float32(whisperMessagePadX), float32(whisperMessagePadY)).
		Background(whisperPanelBg).
		BorderStyle(1, whisperPanelBorder).
		CrossAlign(primitives.CrossAxisStretch)
}

func (w *WhisperWindow) footerTree(ctx Context) []widget.Widget {
	return []widget.Widget{
		primitives.Expanded(
			primitives.Box(w.inputWidget(ctx)).
				Height(float32(whisperInputH)).
				CrossAlign(primitives.CrossAxisStretch),
		),
		rotheme.Button("Send", func() {
			w.submit(ctx)
		}),
	}
}

func (w *WhisperWindow) inputWidget(ctx Context) *textfield.Widget {
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
	return w.inputField
}

func (w *WhisperWindow) submit(ctx Context) {
	message := strings.TrimSpace(w.input)
	if strings.TrimSpace(w.target) == "" || message == "" {
		return
	}
	w.action = WhisperWindowAction{Target: w.target, Message: message}
	w.input = ""
	if w.inputField != nil {
		w.inputField.SetText("")
	}
	w.focusInput()
	w.refresh(ctx)
}

func (w *WhisperWindow) submitFromFocusedEnter(ctx Context) bool {
	if ctx.Input == nil || w.inputField == nil || !w.inputField.IsFocused() {
		return false
	}
	if !ctx.Input.JustPressed(input.KeyEnter) {
		return false
	}
	w.submit(ctx)
	return true
}

func (w *WhisperWindow) refresh(ctx Context) {
	w.SetContent(w.widgetTree(ctx))
	w.Publish(ctx)
}

func (w *WhisperWindow) focusInput() {
	if w.inputField != nil {
		if ctx := windowWidgetContext(w.ctx); ctx != nil {
			ctx.RequestFocus(w.inputField)
		} else {
			w.inputField.SetFocused(true)
		}
	}
}

func (w *WhisperWindow) ensureScrollSignal() state.Signal[float32] {
	if w.scrollY == nil {
		w.scrollY = state.NewSignal[float32](0)
	}
	return w.scrollY
}

func (w *WhisperWindow) title() string {
	target := strings.TrimSpace(w.target)
	if target == "" {
		return "1:1 Chat"
	}
	return "1:1 " + trimRunes(target, 24)
}

func (w *WhisperWindow) visibleLines() []whisperWindowLine {
	if len(w.lines) == 0 {
		return []whisperWindowLine{{text: "No messages", color: whisperEmptyText}}
	}
	return w.lines
}

func whisperMessageContentHeight(lines int) int {
	if lines <= 0 {
		return 0
	}
	return lines*consoleLineH + maxInt(0, lines-1)
}
