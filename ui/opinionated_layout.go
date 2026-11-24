package ui

import (
	"time"

	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
	"github.com/kivutar/goro/session"
	"github.com/kivutar/goro/ui/rotheme"
)

const (
	opinionatedPaneBag = iota
	opinionatedPaneOther
)

type opinionatedSourceTab struct {
	label  string
	source int
}

func (w *InventoryBagWindow) opinionatedSources(ctx Context) []opinionatedSourceTab {
	tabs := []opinionatedSourceTab{
		{label: "Inventory", source: opinionatedSourceBag},
		{label: "Storage", source: opinionatedSourceStorage},
	}
	if inventoryBagHasCart(ctx) {
		tabs = append(tabs, opinionatedSourceTab{label: "Cart", source: opinionatedSourceCart})
	}
	if opinionatedGuildAvailable(ctx) {
		tabs = append(tabs, opinionatedSourceTab{label: "Guild", source: opinionatedSourceGuild})
	}
	return tabs
}

func opinionatedGuildAvailable(ctx Context) bool {
	if ctx.Session == nil {
		return false
	}
	return ctx.Session.GuildID != 0 || ctx.Session.GuildStorage.Open || len(ctx.Session.GuildStorage.Items) > 0
}

func (w *InventoryBagWindow) normalizeOpinionatedSource(ctx Context) {
	for _, tab := range w.opinionatedSources(ctx) {
		if tab.source == w.source {
			return
		}
	}
	w.source = opinionatedSourceBag
}

func (w *InventoryBagWindow) opinionatedSplit() bool {
	return w != nil && w.source != opinionatedSourceBag
}

type opinionatedTabStrip struct {
	widget.WidgetBase
	tabs   []*tabWidget
	active int
}

func newOpinionatedTabStrip(tabs []*tabWidget, active int) *opinionatedTabStrip {
	if active < 0 || active >= len(tabs) {
		active = 0
	}
	strip := &opinionatedTabStrip{tabs: tabs, active: active}
	strip.SetVisible(true)
	strip.SetEnabled(true)
	return strip
}

func (s *opinionatedTabStrip) Layout(ctx widget.Context, constraints geometry.Constraints) geometry.Size {
	count := len(s.tabs)
	width := constraints.MaxWidth
	if !opinionatedWidthBounded(width) {
		width = float32(opinionatedFrameW()) * rotheme.ContentScale()
	}
	height := float32(inventoryBagFilterH) * rotheme.ContentScale()
	if height <= 0 {
		height = float32(inventoryBagFilterH)
	}
	if count == 0 {
		size := constraints.Constrain(geometry.Sz(width, height))
		s.SetBounds(geometry.FromPointSize(s.Position(), size))
		return size
	}
	used := float32(0)
	for i, tab := range s.tabs {
		share := width / float32(count)
		if i == count-1 {
			share = width - used
		}
		tab.Layout(ctx, geometry.Tight(geometry.Sz(share, height)))
		tab.SetBounds(geometry.NewRect(used, 0, share, height))
		used += share
	}
	size := constraints.Constrain(geometry.Sz(width, height))
	s.SetBounds(geometry.FromPointSize(s.Position(), size))
	return size
}

func (s *opinionatedTabStrip) Draw(ctx widget.Context, canvas widget.Canvas) {
	if !s.IsVisible() {
		return
	}
	bounds := s.Bounds()
	canvas.PushTransform(bounds.Min)
	for i, tab := range s.tabs {
		if i == s.active {
			continue
		}
		widget.DrawChild(tab, ctx, canvas)
	}
	if s.active >= 0 && s.active < len(s.tabs) {
		widget.DrawChild(s.tabs[s.active], ctx, canvas)
	}
	canvas.PopTransform()
}

func (s *opinionatedTabStrip) Event(ctx widget.Context, e event.Event) bool {
	mouse, ok := e.(*event.MouseEvent)
	if !ok || !s.IsVisible() || !s.IsEnabled() {
		return false
	}
	local := *mouse
	local.Position = mouse.Position.Sub(s.Bounds().Min)
	order := make([]int, 0, len(s.tabs))
	if s.active >= 0 && s.active < len(s.tabs) {
		order = append(order, s.active)
	}
	for i := range s.tabs {
		if i != s.active {
			order = append(order, i)
		}
	}
	for _, i := range order {
		if s.tabs[i].Bounds().Contains(local.Position) {
			return s.tabs[i].Event(ctx, &local)
		}
	}
	return false
}

func (s *opinionatedTabStrip) Children() []widget.Widget {
	children := make([]widget.Widget, len(s.tabs))
	for i, tab := range s.tabs {
		children[i] = tab
	}
	return children
}

func (w *InventoryBagWindow) opinionatedPaneAt(mx, my int) int {
	if w == nil || !w.opinionatedSplit() {
		return -1
	}
	scale := w.elementScale()
	if scale <= 0 {
		scale = 1
	}
	top := w.y + scaleInt(inventoryOpinionatedChromeH(), scale)
	height := scaleInt(inventoryBagViewH, scale)
	if my < top || my >= top+height {
		return -1
	}
	relX := mx - w.x
	left := scaleInt(opinionatedSideViewW(), scale)
	divider := scaleInt(opinionatedDividerW(), scale)
	if relX < 0 || relX >= scaleInt(opinionatedFrameW(), scale) {
		return -1
	}
	if relX < left {
		return opinionatedPaneBag
	}
	if relX < left+divider {
		return -1
	}
	return opinionatedPaneOther
}

func (w *InventoryBagWindow) bagItems(s *session.Session) []session.InventoryItem {
	if w == nil {
		return nil
	}
	saved := w.source
	w.source = opinionatedSourceBag
	items := w.tabItems(s)
	w.source = saved
	return items
}

func (w *InventoryBagWindow) startPaneDrag(ctx Context, item session.InventoryItem, pane int) {
	w.hideTooltip()
	now := time.Now()
	if pane == opinionatedPaneOther && w.lastClickItem == item.Index && now.Sub(w.lastClickAt) <= 360*time.Millisecond {
		opinionatedWithdraw(ctx, w.source, item)
		w.lastClickItem = 0
		w.refresh(ctx, w.itemInfo)
		return
	}
	if pane == opinionatedPaneBag && w.lastClickItem == item.Index && now.Sub(w.lastClickAt) <= 360*time.Millisecond {
		w.dragActive = false
		w.dragItem = session.InventoryItem{}
		w.activateItem(ctx, item)
		w.lastClickItem = 0
		w.refresh(ctx, w.itemInfo)
		return
	}
	w.dragItem = item
	w.dragActive = true
	w.dragPane = pane
	w.dragFrom = now
	w.lastClickItem = item.Index
	w.lastClickAt = now
}

func (w *InventoryBagWindow) opinionatedRightClick(ctx Context, itemInfo *ItemWindows, item session.InventoryItem, mx, my int, mods event.Modifiers, pane int) {
	w.hideTooltip()
	w.dragActive = false
	w.dragItem = session.InventoryItem{}
	if w.opinionatedSplit() && mods.Has(event.ModShift) {
		other := opinionatedPaneBag
		if pane == opinionatedPaneBag {
			other = opinionatedPaneOther
		}
		w.requestTransfer(ctx, item, pane, other, true)
		return
	}
	if itemInfo != nil {
		itemInfo.openItem(ctx, item, mx, my)
	}
}

func (w *InventoryBagWindow) tryOpinionatedTransfer(ctx Context, item session.InventoryItem, mx, my int) bool {
	if !inventoryOpinionated(ctx) || !w.opinionatedSplit() {
		return false
	}
	pane := w.opinionatedPaneAt(mx, my)
	if pane < 0 || pane == w.dragPane {
		return false
	}
	w.requestTransfer(ctx, item, w.dragPane, pane, false)
	return true
}

func (w *InventoryBagWindow) requestTransfer(ctx Context, item session.InventoryItem, from, to int, whole bool) {
	if from == to || !w.transferAllowed(ctx) {
		return
	}
	amount, prompt := opinionatedPromptAmount(item)
	if whole || !prompt {
		w.sendTransfer(ctx, item, from, to, amount)
		return
	}
	w.amountPrompt.Open(ctx, "Enter amount to move", amount, amount, func(chosen uint16) {
		w.sendTransfer(ctx, item, from, to, chosen)
	})
}

func (w *InventoryBagWindow) transferAllowed(ctx Context) bool {
	return opinionatedMoveAllowed(ctx, w.source)
}

func opinionatedMoveAllowed(ctx Context, source int) bool {
	switch source {
	case opinionatedSourceCart:
		return true
	case opinionatedSourceStorage:
		return ctx.Session != nil && ctx.Session.Storage.Open && !ctx.Session.GuildStorage.Open
	case opinionatedSourceGuild:
		return ctx.Session != nil && ctx.Session.GuildStorage.Open
	default:
		return false
	}
}

func opinionatedPromptAmount(item session.InventoryItem) (uint16, bool) {
	amount := opinionatedMoveAmount(item)
	if amount > 65535 {
		amount = 65535
	}
	prompt := inventoryItemTypeStackable(item.Type) && item.Amount > 1
	return uint16(amount), prompt
}

func (w *InventoryBagWindow) sendTransfer(ctx Context, item session.InventoryItem, from, to int, amount uint16) {
	if amount == 0 {
		return
	}
	if from == opinionatedPaneOther && to == opinionatedPaneBag {
		opinionatedWithdrawAmount(ctx, w.source, item, uint32(amount))
		return
	}
	if from == opinionatedPaneBag && to == opinionatedPaneOther {
		opinionatedDepositAmount(ctx, w.source, item, uint32(amount))
	}
}
