package ui

import (
	"github.com/gogpu/ui/core/listview"
	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/state"
	"github.com/gogpu/ui/widget"
	"github.com/kivutar/goro/network"
	"github.com/kivutar/goro/session"
	"github.com/kivutar/goro/ui/rotheme"
)

const (
	warpraWindowW    = 420
	warpraWindowH    = 340
	warpraTabH       = 24
	warpraRowH       = 22
	warpraListRows   = 10
	warpraTabTown    = 0
	warpraTabField   = 1
	warpraTabDungeon = 2
	warpraTabSpecial = 3
)

// WarpraWindow is the character-menu destination list (towns, fields by
// region, dungeons, special) for players who do not open the world map.
type WarpraWindow struct {
	Window
	Confirm      ConfirmModal
	catalog      session.WarpraCatalog
	tab          int
	fieldGroup   *uint8
	dungeonGroup *uint8
	rows         []warpraListRow
	scrollY      state.Signal[float32]
	onOpen       func()
	onWarp       func(network.WarpraLoc)
	onBuy        func(network.WarpraLoc)
	pending      network.WarpraLoc
	hasPending   bool
	warpAfter    bool
}

type warpraListRow struct {
	label string
	loc   network.WarpraLoc
	group *session.WarpraGroup
	drill bool
}

func (w *WarpraWindow) Toggle(ctx Context) {
	if w.IsOpen() {
		w.Close()
		return
	}
	w.OpenWindow(ctx)
}

func (w *WarpraWindow) OpenWindow(ctx Context) {
	w.EnsureWindow(warpraWindowW, warpraWindowH)
	if w.scrollY == nil {
		w.scrollY = state.NewSignal[float32](0)
	}
	if ctx.Session != nil {
		w.catalog = ctx.Session.Warpra
	}
	w.refreshRows()
	w.Window.Open(ctx, w.widgetTree(ctx))
	w.Publish(ctx)
	if w.onOpen != nil {
		w.onOpen()
	}
}

func (w *WarpraWindow) Rebind(ctx Context, onOpen func(), onWarp, onBuy func(network.WarpraLoc)) {
	w.onOpen = onOpen
	w.onWarp = onWarp
	w.onBuy = onBuy
	if w.scrollY == nil {
		w.scrollY = state.NewSignal[float32](0)
	}
	if ctx.Session != nil {
		w.catalog = ctx.Session.Warpra
	}
	if w.IsOpen() {
		w.refreshRows()
		w.RebindContent(ctx, w.widgetTree(ctx))
	}
}

func (w *WarpraWindow) SetCatalog(ctx Context, catalog session.WarpraCatalog) {
	w.catalog = catalog
	if w.hasPending {
		if next, ok := catalog.Find(w.pending.Kind, w.pending.ID); ok && next.Unlocked {
			w.hasPending = false
			w.Confirm.Close(ctx)
			if w.pending.Kind == network.WarpraKindDungeon {
				g := w.pending.Group
				w.dungeonGroup = &g
			}
		}
	}
	if w.IsOpen() {
		w.refreshRows()
		w.SetContent(w.widgetTree(ctx))
		w.Publish(ctx)
	}
}

func (w *WarpraWindow) Close() {
	w.Confirm.Close(w.ctx)
	w.hasPending = false
	w.warpAfter = false
	w.Window.Close()
}

func (w *WarpraWindow) Update(ctx Context) bool {
	if w.Confirm.Update(ctx) {
		return true
	}
	if !w.IsOpen() {
		return false
	}
	consumed := w.Window.Update(ctx)
	w.Publish(ctx)
	return consumed
}

func (w *WarpraWindow) widgetTree(ctx Context) widget.Widget {
	tabs := []struct {
		label string
		id    int
	}{
		{"Towns", warpraTabTown},
		{"Fields", warpraTabField},
		{"Dungeons", warpraTabDungeon},
		{"Special", warpraTabSpecial},
	}
	tabRow := make([]widget.Widget, 0, len(tabs))
	for _, tab := range tabs {
		id := tab.id
		tabRow = append(tabRow, newTabWidget(tabWidgetConfig{
			label: tab.label, active: w.tab == id, width: 90, height: warpraTabH,
			onClick: func() {
				w.tab = id
				w.fieldGroup = nil
				w.dungeonGroup = nil
				w.scrollY.Set(0)
				w.refreshRows()
				w.SetContent(w.widgetTree(ctx))
				w.Publish(ctx)
			},
		}))
	}
	crumb := w.crumbBar(ctx)
	lv := listview.New(
		listview.ItemCount(len(w.rows)),
		listview.FixedItemHeight(warpraRowH),
		listview.ScrollYSignal(w.scrollY),
		listview.SelectionModeOpt(listview.SelectionNone),
		listview.PainterOpt(rotheme.SelectListPainter{EmptyText: w.emptyText()}),
		listview.BuildItem(func(item listview.ItemContext) widget.Widget {
			if item.Index < 0 || item.Index >= len(w.rows) {
				return rotheme.SelectListRow("", false, warpraRowH)
			}
			return rotheme.SelectListRow(trimRunes(w.rows[item.Index].label, 52), true, warpraRowH)
		}),
		listview.OnItemClick(func(index int) {
			w.clickRow(ctx, index)
		}),
	)
	zeny := "Zeny " + formatWarpraZeny(w.catalog.Zeny)
	return Win(Title("Warpra"), CloseButton(true), OnClose(w.Close),
		Size(warpraWindowW, warpraWindowH),
		Content(
			primitives.Box(
				primitives.HBox(tabRow...).Gap(2),
				crumb,
				primitives.Box(lv).Height(float32(warpraListRows*warpraRowH)).CrossAlign(primitives.CrossAxisStretch),
			).Padding(10).Gap(6),
		),
		Footer(footerLabel(zeny), primitives.Expanded(primitives.Box()), rotheme.Button("Close", w.Close)),
	)
}

func (w *WarpraWindow) crumbBar(ctx Context) widget.Widget {
	label := ""
	show := false
	if w.tab == warpraTabField && w.fieldGroup != nil {
		show = true
		label = w.groupName(network.WarpraKindField, *w.fieldGroup) + " Fields"
	}
	if w.tab == warpraTabDungeon && w.dungeonGroup != nil {
		show = true
		label = w.groupName(network.WarpraKindDungeon, *w.dungeonGroup)
	}
	if !show {
		return primitives.Box().Height(1)
	}
	return primitives.HBox(
		rotheme.Button("Back", func() {
			if w.dungeonGroup != nil {
				w.dungeonGroup = nil
			} else {
				w.fieldGroup = nil
			}
			w.scrollY.Set(0)
			w.refreshRows()
			w.SetContent(w.widgetTree(ctx))
			w.Publish(ctx)
		}).Width(64).Height(warpraTabH),
		rotheme.Label(label).MaxLines(1),
	).Gap(8).CrossAlign(primitives.CrossAxisCenter)
}

func (w *WarpraWindow) groupName(kind uint8, id uint8) string {
	for _, g := range w.catalog.Groups(kind) {
		if g.ID == id {
			return g.Name
		}
	}
	return ""
}

func (w *WarpraWindow) emptyText() string {
	switch w.tab {
	case warpraTabField:
		return "No field regions loaded."
	case warpraTabDungeon:
		return "No dungeons loaded."
	default:
		return "No locations in this list."
	}
}

func (w *WarpraWindow) refreshRows() {
	w.rows = w.rows[:0]
	switch w.tab {
	case warpraTabField:
		if w.fieldGroup == nil {
			for _, g := range w.catalog.Groups(network.WarpraKindField) {
				group := g
				w.rows = append(w.rows, warpraListRow{
					label: group.Name + "  —  " + formatInt(len(group.Locs)) + " maps  " + formatInt(group.Unlocked) + "/" + formatInt(len(group.Locs)) + " visited",
					group: &group,
					drill: true,
				})
			}
			return
		}
		w.appendLocRows(w.catalog.OfKind(network.WarpraKindField), *w.fieldGroup)
	case warpraTabDungeon:
		if w.dungeonGroup == nil {
			for _, g := range w.catalog.Groups(network.WarpraKindDungeon) {
				group := g
				parent := group.Parent()
				stamp := "Locked"
				if parent.Unlocked {
					stamp = "Visited"
				}
				w.rows = append(w.rows, warpraListRow{
					label: warpraMark(parent.Unlocked) + "  " + group.Name + "  —  " + formatInt(len(group.Locs)) + " levels  " + stamp,
					loc:   parent,
					group: &group,
					drill: parent.Unlocked,
				})
			}
			return
		}
		w.appendLocRows(w.catalog.OfKind(network.WarpraKindDungeon), *w.dungeonGroup)
	case warpraTabSpecial:
		for _, loc := range w.catalog.OfKind(network.WarpraKindSpecial) {
			w.rows = append(w.rows, warpraLocRow(loc))
		}
	default:
		for _, loc := range w.catalog.OfKind(network.WarpraKindTown) {
			w.rows = append(w.rows, warpraLocRow(loc))
		}
	}
}

func (w *WarpraWindow) appendLocRows(locs []network.WarpraLoc, group uint8) {
	for _, loc := range locs {
		if loc.Group != group && w.tab != warpraTabSpecial {
			continue
		}
		w.rows = append(w.rows, warpraLocRow(loc))
	}
}

func warpraLocRow(loc network.WarpraLoc) warpraListRow {
	stamp := "Locked"
	if loc.Unlocked {
		stamp = "Visited"
	}
	extra := ""
	if loc.GroupName != "" && (loc.Kind == network.WarpraKindSpecial || loc.Kind == network.WarpraKindDungeon) {
		extra = "  " + loc.GroupName
	}
	return warpraListRow{label: warpraMark(loc.Unlocked) + "  " + loc.Name + extra + "  " + stamp, loc: loc}
}

func warpraMark(unlocked bool) string {
	if unlocked {
		return "✓"
	}
	return "✕"
}

func formatInt(n int) string {
	return formatWarpraZeny(uint32(n))
}

func (w *WarpraWindow) clickRow(ctx Context, index int) {
	if index < 0 || index >= len(w.rows) {
		return
	}
	row := w.rows[index]
	if w.tab == warpraTabField && w.fieldGroup == nil && row.group != nil {
		id := row.group.ID
		w.fieldGroup = &id
		w.scrollY.Set(0)
		w.refreshRows()
		w.SetContent(w.widgetTree(ctx))
		w.Publish(ctx)
		return
	}
	if w.tab == warpraTabDungeon && w.dungeonGroup == nil && row.group != nil {
		if !row.drill {
			w.UnlockOrWarp(ctx, row.loc, false)
			return
		}
		id := row.group.ID
		w.dungeonGroup = &id
		w.scrollY.Set(0)
		w.refreshRows()
		w.SetContent(w.widgetTree(ctx))
		w.Publish(ctx)
		return
	}
	w.UnlockOrWarp(ctx, row.loc, false)
}

func (w *WarpraWindow) UnlockOrWarp(ctx Context, loc network.WarpraLoc, autoWarp bool) {
	if loc.Map == "" && loc.Name == "" {
		return
	}
	if loc.Unlocked {
		w.warpAfter = false
		if w.onWarp != nil {
			w.onWarp(loc)
		}
		return
	}
	if session.WarpraLevelGate(loc) > 0 {
		if w.onWarp != nil {
			w.onWarp(loc)
		}
		return
	}
	w.pending = loc
	w.hasPending = true
	w.warpAfter = autoWarp
	title, body := warpraUnlockMessage(loc, w.catalog.UnlockPrice(), w.catalog.Zeny)
	w.Confirm.Open(ctx, title, body, func() {
		if w.onBuy != nil {
			w.onBuy(w.pending)
		}
	}, func() {
		w.warpAfter = false
		w.hasPending = false
	})
}

func (w *WarpraWindow) ConsumePendingBuy() (network.WarpraLoc, bool, bool) {
	if !w.hasPending {
		return network.WarpraLoc{}, false, false
	}
	return w.pending, w.warpAfter, true
}

func (w *WarpraWindow) ClearPending() {
	w.hasPending = false
	w.warpAfter = false
}
