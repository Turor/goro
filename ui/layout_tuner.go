package ui

import (
	"fmt"

	"github.com/gogpu/ui/core/textfield"
	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/widget"
	"github.com/kivutar/goro/ui/rotheme"
)

const (
	layoutTunerW        = 400
	layoutTunerRows     = 12
	layoutTunerRowH     = 22
	layoutTunerKeyW     = 170
	layoutTunerContentH = 20 + layoutTunerRows*(layoutTunerRowH+2) + 16
)

// LayoutTuner is the --ui-dev panel (F11) that edits ui-layout.ini values
// live. Numbers step with -/+; colors and names are typed, then Apply. Save
// writes the changed values back to the layout file.
type LayoutTuner struct {
	Window
	path    string
	page    int
	gen     uint64
	status  string
	pending map[string]string
}

// NewLayoutTuner edits the layout file at path.
func NewLayoutTuner(path string) *LayoutTuner {
	return &LayoutTuner{path: path, pending: map[string]string{}}
}

func (t *LayoutTuner) Toggle(ctx Context) {
	t.EnsureWindow(layoutTunerW, ROWindowTitleHeight+layoutTunerContentH+ROWindowFooterHeight)
	if t.IsOpen() {
		t.close(ctx)
		return
	}
	screenW, _ := ctx.ScreenSize()
	t.SetAutoPosition(maxInt(windowScreenMargin, screenW-layoutTunerW-16), 60)
	t.gen = rotheme.Generation()
	t.Open(ctx, t.widgetTree(ctx))
	t.Publish(ctx)
}

// Update rebuilds the panel when values change (from it or the file).
func (t *LayoutTuner) Update(ctx Context) bool {
	t.ctx = ctx
	if !t.IsOpen() {
		return false
	}
	if gen := rotheme.Generation(); gen != t.gen {
		t.gen = gen
		t.SetContent(t.widgetTree(ctx))
	}
	consumed := t.Window.Update(ctx)
	t.Publish(ctx)
	return consumed
}

func (t *LayoutTuner) close(ctx Context) {
	t.Close()
	t.Publish(ctx)
}

func (t *LayoutTuner) pages(total int) int {
	return maxInt(1, (total+layoutTunerRows-1)/layoutTunerRows)
}

func (t *LayoutTuner) turn(ctx Context, delta int) {
	pages := t.pages(len(rotheme.Tokens()))
	t.page = (t.page + delta + pages) % pages
	t.SetContent(t.widgetTree(ctx))
	t.Publish(ctx)
}

func (t *LayoutTuner) report(err error, ok string) {
	if err != nil {
		t.status = err.Error()
		return
	}
	t.status = ok
}

// applyPending applies typed values; the first invalid one stops it.
func (t *LayoutTuner) applyPending() error {
	for id, raw := range t.pending {
		info, found := tunerToken(id)
		if !found {
			continue
		}
		if err := rotheme.SetToken(info.Section, info.Key, raw); err != nil {
			return err
		}
		delete(t.pending, id)
	}
	return nil
}

func (t *LayoutTuner) apply() {
	t.report(t.applyPending(), "applied")
}

func (t *LayoutTuner) save() {
	if err := t.applyPending(); err != nil {
		t.report(err, "")
		return
	}
	t.report(rotheme.SaveOverrides(t.path), "saved")
}

func (t *LayoutTuner) revert() {
	clear(t.pending)
	applied, _, err := rotheme.LoadLayoutFile(t.path)
	if err == nil && !applied {
		rotheme.ResetLayout()
	}
	t.report(err, "reverted to the file")
}

func tunerToken(id string) (rotheme.TokenInfo, bool) {
	for _, info := range rotheme.Tokens() {
		if info.Section+"."+info.Key == id {
			return info, true
		}
	}
	return rotheme.TokenInfo{}, false
}

func (t *LayoutTuner) widgetTree(ctx Context) widget.Widget {
	all := rotheme.Tokens()
	pages := t.pages(len(all))
	t.page = minInt(t.page, pages-1)
	start := t.page * layoutTunerRows
	shown := all[start:minInt(len(all), start+layoutTunerRows)]
	header := fmt.Sprintf("page %d/%d  [%s]", t.page+1, pages, shown[0].Section)
	if t.status != "" {
		header += "  " + t.status
	}
	rows := make([]widget.Widget, 0, len(shown)+1)
	rows = append(rows, rotheme.Label(header))
	for _, info := range shown {
		rows = append(rows, t.row(info))
	}
	return Win(
		Title("UI Tuner (F11)"),
		CloseButton(true),
		OnClose(func() { t.close(ctx) }),
		Size(layoutTunerW, float32(ROWindowTitleHeight+layoutTunerContentH+ROWindowFooterHeight)),
		Content(primitives.Box(rows...).Padding(8).Gap(2)),
		Footer(
			rotheme.Button("<", func() { t.turn(ctx, -1) }),
			rotheme.Button(">", func() { t.turn(ctx, 1) }),
			primitives.Expanded(primitives.Box()),
			rotheme.Button("Apply", t.apply),
			rotheme.Button("Save", t.save),
			rotheme.Button("Revert", t.revert),
		),
	)
}

func (t *LayoutTuner) row(info rotheme.TokenInfo) widget.Widget {
	label := rotheme.Text(info.Key)
	if info.Value != info.Default {
		label.Color(rotheme.Default.Colors.TitleText)
	}
	key := primitives.Box(label).Width(layoutTunerKeyW)
	var value widget.Widget
	var controls []widget.Widget
	if info.Numeric {
		value = rotheme.Text(info.Value)
		controls = []widget.Widget{
			rotheme.Button("-", func() { t.report(rotheme.StepToken(info, -1), "") }),
			rotheme.Button("+", func() { t.report(rotheme.StepToken(info, 1), "") }),
		}
	} else {
		id := info.Section + "." + info.Key
		current := info.Value
		if pending, ok := t.pending[id]; ok {
			current = pending
		}
		value = rotheme.TextField(current, textfield.TypeText, func(v string) { t.pending[id] = v }, func(string) {})
	}
	children := append([]widget.Widget{key, primitives.Expanded(primitives.Box(value).Height(layoutTunerRowH))}, controls...)
	return primitives.HBox(children...).
		Gap(4).
		CrossAlign(primitives.CrossAxisCenter).
		Height(layoutTunerRowH)
}
