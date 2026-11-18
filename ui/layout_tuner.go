package ui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/gogpu/ui/core/textfield"
	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/widget"
	"github.com/kivutar/goro/ui/rotheme"
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
	// section is the index into tunerSections (0 is every section); query
	// filters keys, docs and section names by substring.
	section int
	query   string
}

func layoutTunerContentH() int {
	return layoutTunerHeaderH + layoutTunerRows*(layoutTunerRowH+layoutTunerRowGap) + layoutTunerContentPad*2
}

// tunerSections is "all" followed by every section with tokens.
func tunerSections() []string {
	return append([]string{"all"}, rotheme.Sections()...)
}

// filtered lists the tokens in the chosen section that match the query.
func (t *LayoutTuner) filtered() []rotheme.TokenInfo {
	sections := tunerSections()
	t.section = (t.section%len(sections) + len(sections)) % len(sections)
	section := sections[t.section]
	query := strings.ToLower(strings.TrimSpace(t.query))
	all := rotheme.Tokens()
	out := make([]rotheme.TokenInfo, 0, len(all))
	for _, info := range all {
		if t.section != 0 && info.Section != section {
			continue
		}
		if query != "" && !strings.Contains(info.Section+"."+info.Key, query) && !strings.Contains(strings.ToLower(info.Doc), query) {
			continue
		}
		out = append(out, info)
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Section < out[j].Section })
	return out
}

func (t *LayoutTuner) chooseSection(ctx Context, delta int) {
	t.section += delta
	t.page = 0
	t.rebuild(ctx)
}

func (t *LayoutTuner) search(ctx Context, query string) {
	t.query = query
	t.page = 0
	t.rebuild(ctx)
}

func (t *LayoutTuner) rebuild(ctx Context) {
	t.SetContent(t.widgetTree(ctx))
	t.Publish(ctx)
}

// NewLayoutTuner edits the layout file at path.
func NewLayoutTuner(path string) *LayoutTuner {
	return &LayoutTuner{path: path, pending: map[string]string{}}
}

func (t *LayoutTuner) Toggle(ctx Context) {
	t.EnsureWindow(layoutTunerW, ROWindowTitleHeight+layoutTunerContentH()+ROWindowFooterHeight)
	if t.IsOpen() {
		t.close(ctx)
		return
	}
	screenW, _ := ctx.ScreenSize()
	t.SetAutoPosition(maxInt(windowScreenMargin, screenW-layoutTunerW-layoutTunerScreenRight), layoutTunerScreenTop)
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
		t.SetSize(layoutTunerW, ROWindowTitleHeight+layoutTunerContentH()+ROWindowFooterHeight)
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
	pages := t.pages(len(t.filtered()))
	t.page = (t.page + delta + pages) % pages
	t.rebuild(ctx)
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
	shownAll := t.filtered()
	pages := t.pages(len(shownAll))
	t.page = maxInt(0, minInt(t.page, pages-1))
	start := t.page * layoutTunerRows
	shown := shownAll[minInt(start, len(shownAll)):minInt(len(shownAll), start+layoutTunerRows)]
	section := tunerSections()[t.section]
	title := fmt.Sprintf("[%s] %d keys", section, len(shownAll))
	if doc := rotheme.SectionDoc(section); doc != "" {
		title += " - " + doc
	}
	status := fmt.Sprintf("page %d/%d", t.page+1, pages)
	if t.status != "" {
		status += "  " + t.status
	}
	search := rotheme.TextField(t.query, textfield.TypeText, func(v string) { t.query = v }, func(v string) { t.search(ctx, v) })
	controls := primitives.HBox(
		rotheme.Button("<", func() { t.chooseSection(ctx, -1) }),
		primitives.Expanded(primitives.Box(rotheme.Label(title)).Height(float32(layoutTunerRowH))),
		rotheme.Button(">", func() { t.chooseSection(ctx, 1) }),
	).Gap(float32(layoutTunerRowGap * 2)).CrossAlign(primitives.CrossAxisCenter).Height(float32(layoutTunerRowH))
	searchRow := primitives.HBox(
		primitives.Box(rotheme.Text("search")).Width(float32(layoutTunerSearchLabelW)),
		primitives.Expanded(primitives.Box(search).Height(float32(layoutTunerRowH))),
		rotheme.Button("Find", func() { t.search(ctx, t.query) }),
		primitives.Box(rotheme.Text(status)).Width(float32(layoutTunerStatusW)),
	).Gap(float32(layoutTunerRowGap * 2)).CrossAlign(primitives.CrossAxisCenter).Height(float32(layoutTunerRowH))
	rows := make([]widget.Widget, 0, len(shown)+2) // theme:ok slice capacity
	rows = append(rows, controls, searchRow)
	for _, info := range shown {
		rows = append(rows, t.row(info, t.section == 0))
	}
	return Win(
		Title("UI Tuner (F11)"),
		CloseButton(true),
		OnClose(func() { t.close(ctx) }),
		Size(float32(layoutTunerW), float32(ROWindowTitleHeight+layoutTunerContentH()+ROWindowFooterHeight)),
		Content(primitives.Box(rows...).Padding(float32(layoutTunerContentPad)).Gap(float32(layoutTunerRowGap))),
		Footer(
			rotheme.Button("Prev", func() { t.turn(ctx, -1) }),
			rotheme.Button("Next", func() { t.turn(ctx, 1) }),
			primitives.Expanded(primitives.Box()),
			rotheme.Button("Apply", t.apply),
			rotheme.Button("Save", t.save),
			rotheme.Button("Revert", t.revert),
		),
	)
}

func (t *LayoutTuner) row(info rotheme.TokenInfo, withSection bool) widget.Widget {
	name := info.Key
	if withSection {
		name = info.Section + "." + info.Key
	}
	label := rotheme.Text(name)
	if info.Value != info.Default {
		label.Color(rotheme.Default.Colors.TitleText)
	}
	key := primitives.Box(label).Width(float32(layoutTunerKeyW))
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
	children := append([]widget.Widget{key, primitives.Expanded(primitives.Box(value).Height(float32(layoutTunerRowH)))}, controls...)
	return primitives.HBox(children...).
		Gap(float32(layoutTunerRowGap * 2)).
		CrossAlign(primitives.CrossAxisCenter).
		Height(float32(layoutTunerRowH))
}
