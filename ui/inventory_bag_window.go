package ui

import (
	"fmt"
	"github.com/kivutar/goro/glog"
	"github.com/kivutar/goro/input"
	"image"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gogpu/ui/core/scrollview"
	"github.com/gogpu/ui/core/textfield"
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/state"
	"github.com/gogpu/ui/widget"
	"github.com/kivutar/goro/db"
	"github.com/kivutar/goro/render"
	"github.com/kivutar/goro/res"
	"github.com/kivutar/goro/session"
	"github.com/kivutar/goro/ui/rotheme"
)

const inventoryBagTabOver = 1

const (
	inventoryBagTabItem = iota
	inventoryBagTabEquip
	inventoryBagTabEtc
)

const (
	opinionatedSourceBag = iota
	opinionatedSourceCart
	opinionatedSourceStorage
	opinionatedSourceGuild
)

const (
	opinionatedSortRecent = iota
	opinionatedSortName
	opinionatedSortNameDesc
	opinionatedSortCount
)

const (
	opinionatedBagTabAll = iota
	opinionatedBagTabUsable
	opinionatedBagTabEquip
	opinionatedBagTabAmmo
	opinionatedBagTabEtc
	opinionatedBagTabCard
)

// theme:ok packet option bits
const inventoryBagCartOptionMask uint32 = 0x00000008 | 0x00000080 | 0x00000100 | 0x00000200 | 0x00000400

var inventoryBagTabs = []struct {
	label string
	tab   int
}{
	{label: "Item", tab: inventoryBagTabItem},
	{label: "Equip", tab: inventoryBagTabEquip},
	{label: "Etc", tab: inventoryBagTabEtc},
}

var opinionatedInventoryTabs = []struct {
	label string
	tab   int
}{
	{label: "All", tab: opinionatedBagTabAll},
	{label: "Usable", tab: opinionatedBagTabUsable},
	{label: "Equip", tab: opinionatedBagTabEquip},
	{label: "Ammo", tab: opinionatedBagTabAmmo},
	{label: "Etc", tab: opinionatedBagTabEtc},
	{label: "Cards", tab: opinionatedBagTabCard},
	{label: "Favorites", tab: opinionatedBagTabFavorite},
}

const opinionatedBagTabFavorite = opinionatedBagTabCard + 1

type InventoryBagWindow struct {
	Window
	layoutGen     uint64
	tab           int
	opinionated   bool
	search        string
	sort          int
	source        int
	resources     *res.Manager
	scrollY       state.Signal[float32]
	scrollYSide   state.Signal[float32]
	dragPane      int
	snapshot      string
	itemInfo      *ItemWindows
	lastClickItem uint16
	lastClickAt   time.Time
	dragItem      session.InventoryItem
	dragActive    bool
	dragFrom      time.Time
	amountPrompt  amountPrompt
	pendingCard   uint16
	tooltip       tooltipState
	icons         map[inventoryBagIconKey]image.Image
	iconMiss      map[inventoryBagIconKey]struct{}
	searchCtx     Context
	searchField   *textfield.Widget
	searchBox     *inventorySearchBox
	searchArmed   bool
	bagGrid       *inventoryGridWidget
	sideGrid      *inventoryGridWidget
}

type inventoryBagIconKey struct {
	itemID     uint32
	identified bool
}

func (w *InventoryBagWindow) Toggle(ctx Context) {
	w.syncStyle(ctx.Session)
	width, height := inventoryBagFrameSize(inventoryOpinionatedSession(ctx.Session))
	w.EnsureWindow(width, height)
	if w.IsOpen() {
		w.hideTooltip()
		w.amountPrompt.Close(ctx)
		w.Window.Close()
		w.Publish(ctx)
		return
	}
	w.selectFirstNonEmptyTab(ctx.Session)
	w.ClampScroll(ctx.Session)
	w.snapshot = w.inventorySnapshot(ctx.Session)
	x, y := inventoryBagDefaultPosition(ctx)
	w.SetAutoPosition(x, y)
	w.layoutGen = rotheme.Generation()
	w.SetSize(width, height)
	w.Open(ctx, w.widgetTree(ctx, nil))
	w.bindLayoutReload()
	w.Publish(ctx)
}

func (w *InventoryBagWindow) Update(ctx Context, shortcuts *ShortcutBar, storage *StorageWindow, cart *CartWindow, trade *TradeWindow, equipment *EquipmentWindow, itemInfo *ItemWindows, dropTargets ...InventoryDropTarget) bool {
	w.UpdatePresentation(ctx, itemInfo)
	if !w.IsOpen() || ctx.Input == nil {
		w.hideTooltip()
		return false
	}
	if w.UpdateDrag(ctx, shortcuts, storage, cart, trade, equipment, dropTargets...) {
		return true
	}
	if w.opinionated {
		w.syncOpinionatedDrag()
		w.syncSearchPointer(ctx)
	}
	consumed := w.Window.Update(ctx)
	if w.opinionated {
		w.syncSearchTyping(ctx)
	}
	if !w.IsOpen() {
		w.hideTooltip()
		w.Publish(ctx)
		return consumed
	}
	w.Publish(ctx)
	return consumed
}

// UpdatePresentation follows server changes even when another window consumes input.
func (w *InventoryBagWindow) UpdatePresentation(ctx Context, itemInfo *ItemWindows) {
	if ctx.Resources != nil {
		w.resources = ctx.Resources
	}
	width, height := inventoryBagFrameSize(inventoryOpinionatedSession(ctx.Session))
	w.EnsureWindow(width, height)
	// SetContent releases the window's drag capture. Leave the snapshot stale
	// until the drag ends so the next presentation update applies all changes.
	if !w.IsOpen() || w.dragging {
		return
	}
	w.syncStyle(ctx.Session)
	width, height = inventoryBagFrameSize(w.opinionated)
	w.ClampScroll(ctx.Session)
	snapshot := w.inventorySnapshot(ctx.Session)
	if gen := rotheme.Generation(); gen != w.layoutGen {
		// A reloaded ui-layout.ini may change the cell size or rows.
		w.layoutGen = gen
		w.SetSize(width, height)
		w.snapshot = ""
	}
	if snapshot != w.snapshot || itemInfo != w.itemInfo {
		if w.searchField != nil && w.searchField.IsFocused() && w.bagGrid != nil {
			w.snapshot = snapshot
			w.itemInfo = itemInfo
			w.syncSearchGrids(ctx)
			return
		}
		w.hideTooltip()
		w.snapshot = snapshot
		w.itemInfo = itemInfo
		w.SetSize(width, height)
		w.SetContent(w.widgetTree(ctx, itemInfo))
		w.Publish(ctx)
	}
}

type InventoryDropTarget interface {
	AcceptInventoryDrop(Context, session.InventoryItem, int, int) bool
}

func (w *InventoryBagWindow) UpdateDrag(ctx Context, shortcuts *ShortcutBar, storage *StorageWindow, cart *CartWindow, trade *TradeWindow, equipment *EquipmentWindow, dropTargets ...InventoryDropTarget) bool {
	if w.UpdateDropPrompt(ctx) {
		return true
	}
	if !w.dragActive || ctx.Input == nil {
		return false
	}
	if ctx.Input.MouseJustReleased(input.MouseButtonLeft) || !ctx.Input.MousePressed(input.MouseButtonLeft) {
		item := w.dragItem
		w.dragActive = false
		w.dragItem = session.InventoryItem{}
		for _, target := range dropTargets {
			if target != nil && target.AcceptInventoryDrop(ctx, item, ctx.Input.MouseX, ctx.Input.MouseY) {
				return true
			}
		}
		if storage != nil && storage.AcceptInventoryDrop(ctx, item, ctx.Input.MouseX, ctx.Input.MouseY) {
			return true
		}
		if cart != nil && cart.AcceptInventoryDrop(ctx, item, ctx.Input.MouseX, ctx.Input.MouseY) {
			return true
		}
		if trade != nil && trade.AcceptInventoryDrop(ctx, item, ctx.Input.MouseX, ctx.Input.MouseY) {
			return true
		}
		if equipment != nil && equipment.AcceptInventoryDrop(ctx, item, ctx.Input.MouseX, ctx.Input.MouseY) {
			return true
		}
		if shortcuts != nil && shortcuts.AcceptItemDrop(ctx, item, ctx.Input.MouseX, ctx.Input.MouseY) {
			return true
		}
		if manager, ok := ctx.UIManager.(interface{ PointerBlocked(int, int) bool }); ok && manager.PointerBlocked(ctx.Input.MouseX, ctx.Input.MouseY) {
			return true
		}
		if w.tryOpinionatedTransfer(ctx, item, ctx.Input.MouseX, ctx.Input.MouseY) {
			return true
		}
		if !w.pointInside(ctx.Input.MouseX, ctx.Input.MouseY) {
			if !inventoryOpinionated(ctx) || w.dragPane == opinionatedPaneBag {
				w.requestDrop(ctx, item)
			}
			return true
		}
		return true
	}
	return true
}

func (w *InventoryBagWindow) UpdateDropPrompt(ctx Context) bool {
	return w != nil && w.amountPrompt.Update(ctx)
}

func (w *InventoryBagWindow) requestDrop(ctx Context, item session.InventoryItem) {
	maxAmount := inventoryDropMaxAmount(item)
	if inventoryItemTypeStackable(item.Type) && maxAmount > 1 {
		w.amountPrompt.Open(ctx, "Enter amount to drop", maxAmount, maxAmount, func(amount uint16) {
			w.sendDrop(ctx, item, amount)
		})
		return
	}
	w.sendDrop(ctx, item, 1)
}

func (w *InventoryBagWindow) sendDrop(ctx Context, item session.InventoryItem, requested uint16) {
	current, ok := currentInventoryDropItem(ctx.Session, item)
	if !ok {
		glog.Warnf("inventory drop cancelled: item no longer available index=%d item=%d", item.Index, item.ItemID)
		return
	}
	item = current
	amount := inventoryDropAmount(item, requested)
	if err := dropInventoryItem(ctx, item, amount); err != nil {
		glog.Warnf("inventory drop failed: %v", err)
		return
	}
	glog.Debugf("inventory drop requested index=%d item=%d amount=%d", item.Index, item.ItemID, amount)
}

func (w *InventoryBagWindow) KeyboardShortcutsBlocked() bool {
	if w == nil {
		return false
	}
	if w.amountPrompt.IsOpen() {
		return true
	}
	if w.searchArmed {
		return true
	}
	return w.searchField != nil && w.searchField.IsFocused()
}

func (w *InventoryBagWindow) DrawTooltip(ctx Context, screen *render.Frame) {
	if w.dragActive {
		return
	}
	w.tooltip.Draw(ctx, screen)
}

func (w *InventoryBagWindow) DrawDragGhost(screen *render.Frame, ctx Context, assets AssetProvider) {
	if !w.dragActive || screen == nil || ctx.Input == nil || assets == nil {
		return
	}
	if time.Since(w.dragFrom) <= 80*time.Millisecond {
		return
	}
	assets.DrawInventoryItemIcon(screen, ctx.Resources, w.dragItem, ctx.Input.MouseX-inventoryIconSize/2, ctx.Input.MouseY-inventoryIconSize/2)
}

func (w *InventoryBagWindow) Rebind(ctx Context, itemInfo *ItemWindows) {
	w.EnsureWindow(inventoryBagWidth, inventoryBagHeight)
	if !w.IsOpen() {
		return
	}
	w.layoutGen = rotheme.Generation()
	w.SetSize(inventoryBagFrameSize(w.opinionated))
	w.refresh(ctx, itemInfo)
	w.bindLayoutReload()
}

func (w *InventoryBagWindow) bindLayoutReload() {
	w.OnLayoutReload(func() {
		if !w.IsOpen() {
			return
		}
		w.layoutGen = rotheme.Generation()
		w.SetSize(inventoryBagFrameSize(w.opinionated))
		w.refresh(w.ctx, w.itemInfo)
	})
}

func (w *InventoryBagWindow) PendingCardIndex() uint16 {
	return w.pendingCard
}

func (w *InventoryBagWindow) widgetTree(ctx Context, itemInfo *ItemWindows) widget.Widget {
	items := w.tabItems(ctx.Session)
	viewW := inventoryBagViewW
	if inventoryOpinionated(ctx) {
		viewW = inventoryOpinionatedViewW()
	}
	grid := newInventoryGridWidget(inventoryGridConfig{
		items:     items,
		icons:     w.itemIcons(ctx, items),
		amounts:   inventoryGridAmountLabels(items),
		viewWidth: viewW, onPress: func(item session.InventoryItem) { w.startItemDragOrActivate(ctx, item) },
		onHover: func(item session.InventoryItem) { w.showTooltip(ctx, item) },
		onLeave: func() { w.hideTooltip() },
		onRightClick: func(item session.InventoryItem, mx, my int) {
			w.hideTooltip()
			w.dragActive = false
			w.dragItem = session.InventoryItem{}
			if itemInfo != nil {
				itemInfo.openItem(ctx, item, mx, my)
			}
		},
	})
	scroll := scrollview.New(
		grid,
		scrollview.DirectionOpt(scrollview.Vertical),
		scrollview.ScrollbarOpt(scrollview.ScrollbarAuto),
		scrollview.ScrollYSignal(w.ensureScrollSignal()),
		scrollview.ScrollStep(float32(inventoryBagCell)),
	)
	width, height := inventoryBagFrameSize(inventoryOpinionated(ctx))
	body := verticalTabFrame(
		w.tabColumn(ctx),
		primitives.Box(scroll).
			Width(float32(viewW)).
			Height(float32(inventoryBagViewH)),
	)
	if inventoryOpinionated(ctx) {
		return w.opinionatedWindow(ctx, itemInfo)
	}
	return Win(
		Title("Inventory"),
		CloseButton(true),
		OnClose(func() {
			w.amountPrompt.Close(ctx)
			w.Window.Close()
			w.Publish(ctx)
		}),
		Size(float32(width), float32(height)),
		Content(body),
	)
}

func (w *InventoryBagWindow) tabColumn(ctx Context) widget.Widget {
	tabs := make([]widget.Widget, 0, len(inventoryBagTabs))
	for _, tab := range inventoryBagTabs {
		tab := tab
		tabs = append(tabs, newTabWidget(tabWidgetConfig{
			label:         tab.label,
			labelRotation: rotheme.TextRotationCounterClockwise,
			active:        tab.tab == w.tab,
			width:         inventoryBagTabW + inventoryBagTabOver*2,
			height:        inventoryBagTabH,
			onClick: func() {
				w.hideTooltip()
				w.tab = tab.tab
				w.ensureScrollSignal().Set(0)
				w.lastClickItem = 0
				w.refresh(ctx, w.itemInfo)
			},
		}))
	}
	return primitives.Box(tabs...).
		Width(float32(inventoryBagTabRail)).
		Height(float32(inventoryBagViewH)).
		Gap(-inventoryBagTabOver)
}

func (w *InventoryBagWindow) moveRow(ctx Context) widget.Widget {
	storageLabel := "To storage"
	storageDest := opinionatedSourceStorage
	if ctx.Session != nil && ctx.Session.GuildStorage.Open {
		storageLabel = "To guild"
		storageDest = opinionatedSourceGuild
	}
	return primitives.HBox(
		rotheme.Button("To cart", func() {
			w.depositSelected(ctx, opinionatedSourceCart)
		}),
		rotheme.Button(storageLabel, func() {
			w.depositSelected(ctx, storageDest)
		}),
		rotheme.Button("Favorite", func() {
			w.toggleFavorite(ctx)
		}),
	).Gap(float32(inventoryBagMenuGap)).Height(float32(inventoryBagFilterH))
}

func (w *InventoryBagWindow) toggleFavorite(ctx Context) {
	if !inventoryOpinionated(ctx) || w.source != opinionatedSourceBag || ctx.Session == nil {
		return
	}
	for i := range ctx.Session.Inventory.Items {
		item := &ctx.Session.Inventory.Items[i]
		if item.Index != w.lastClickItem {
			continue
		}
		next := !item.Favorite
		if ctx.Network != nil {
			if err := ctx.Network.SendItemFavorite(item.Index, next); err != nil {
				glog.Warnf("favorite toggle failed: %v", err)
				return
			}
		}
		item.Favorite = next
		w.refresh(ctx, w.itemInfo)
		return
	}
}

func (w *InventoryBagWindow) depositSelected(ctx Context, dest int) {
	if !inventoryOpinionated(ctx) || w.source != opinionatedSourceBag {
		return
	}
	item, ok := session.FindInventoryItem(ctx.Session, w.lastClickItem)
	if !ok {
		return
	}
	opinionatedDeposit(ctx, dest, item)
}

func (w *InventoryBagWindow) searchRow(ctx Context) widget.Widget {
	w.searchCtx = ctx
	box := newInventorySearchBox(w)
	w.searchBox = box
	return newOpinionatedSearchRow(box, newOpinionatedSortBox(w, ctx))
}

func opinionatedMoveAmount(item session.InventoryItem) uint32 {
	amount := uint32(item.Amount)
	if amount == 0 {
		return 1
	}
	return amount
}

func opinionatedWithdraw(ctx Context, source int, item session.InventoryItem) {
	opinionatedWithdrawAmount(ctx, source, item, opinionatedMoveAmount(item))
}

func opinionatedWithdrawAmount(ctx Context, source int, item session.InventoryItem, amount uint32) {
	if ctx.Network == nil {
		glog.Warnf("opinionated withdraw failed: not connected")
		return
	}
	if amount == 0 {
		amount = opinionatedMoveAmount(item)
	}
	var err error
	switch source {
	case opinionatedSourceCart:
		err = ctx.Network.SendMoveFromCart(item.Index, amount)
	case opinionatedSourceStorage:
		if ctx.Session == nil || !ctx.Session.Storage.Open || ctx.Session.GuildStorage.Open {
			return
		}
		err = ctx.Network.SendMoveFromStorage(item.Index, amount)
	case opinionatedSourceGuild:
		if ctx.Session == nil || !ctx.Session.GuildStorage.Open {
			return
		}
		err = ctx.Network.SendMoveFromStorage(item.Index, amount)
	default:
		return
	}
	if err != nil {
		glog.Warnf("opinionated withdraw failed: %v", err)
		return
	}
	glog.Debugf("opinionated withdraw requested source=%d index=%d item=%d amount=%d", source, item.Index, item.ItemID, amount)
}

func opinionatedDeposit(ctx Context, dest int, item session.InventoryItem) {
	opinionatedDepositAmount(ctx, dest, item, opinionatedMoveAmount(item))
}

func opinionatedDepositAmount(ctx Context, dest int, item session.InventoryItem, amount uint32) {
	if ctx.Network == nil {
		glog.Warnf("opinionated deposit failed: not connected")
		return
	}
	if amount == 0 {
		amount = opinionatedMoveAmount(item)
	}
	var err error
	switch dest {
	case opinionatedSourceCart:
		err = ctx.Network.SendMoveToCart(item.Index, amount)
	case opinionatedSourceStorage:
		if ctx.Session == nil || !ctx.Session.Storage.Open || ctx.Session.GuildStorage.Open {
			return
		}
		err = ctx.Network.SendMoveToStorage(item.Index, amount)
	case opinionatedSourceGuild:
		if ctx.Session == nil || !ctx.Session.GuildStorage.Open {
			return
		}
		err = ctx.Network.SendMoveToStorage(item.Index, amount)
	default:
		return
	}
	if err != nil {
		glog.Warnf("opinionated deposit failed: %v", err)
		return
	}
	glog.Debugf("opinionated deposit requested dest=%d index=%d item=%d amount=%d", dest, item.Index, item.ItemID, amount)
}

func opinionatedSortLabel(sort int) string {
	switch sort {
	case opinionatedSortName:
		return "A-Z"
	case opinionatedSortNameDesc:
		return "Z-A"
	default:
		return "Recent"
	}
}

func (w *InventoryBagWindow) sourceRow(ctx Context) widget.Widget {
	tabs := []struct {
		label  string
		source int
	}{
		{label: "Bag", source: opinionatedSourceBag},
		{label: "Cart", source: opinionatedSourceCart},
		{label: "Storage", source: opinionatedSourceStorage},
		{label: "Guild", source: opinionatedSourceGuild},
	}
	cells := make([]widget.Widget, 0, len(tabs))
	for i, tab := range tabs {
		tab := tab
		cells = append(cells, newTabWidget(tabWidgetConfig{
			label:  tab.label,
			active: tab.source == w.source,
			width:  inventoryBagFilterTabWidth(i, len(tabs)),
			height: inventoryBagFilterH,
			onClick: func() {
				w.hideTooltip()
				w.source = tab.source
				w.ensureScrollSignal().Set(0)
				w.refresh(ctx, w.itemInfo)
			},
		}))
	}
	return primitives.HBox(cells...).Gap(0).Height(float32(inventoryBagFilterH))
}

func (w *InventoryBagWindow) filterRows(ctx Context) widget.Widget {
	return w.categoryStrip(ctx)
}

func (w *InventoryBagWindow) categoryStrip(ctx Context) widget.Widget {
	tabs := make([]*tabWidget, 0, len(opinionatedInventoryTabs))
	active := 0
	for i, tab := range opinionatedInventoryTabs {
		tab := tab
		if tab.tab == w.tab {
			active = i
		}
		tabs = append(tabs, newTabWidget(tabWidgetConfig{
			label:      tab.label,
			active:     tab.tab == w.tab,
			height:     inventoryBagFilterH,
			omitBottom: tab.tab == w.tab,
			activeFill: inventoryBagTabActive,
			onClick: func() {
				w.hideTooltip()
				w.tab = tab.tab
				w.ensureScrollSignal().Set(0)
				w.lastClickItem = 0
				w.refresh(ctx, w.itemInfo)
			},
		}))
	}
	return newOpinionatedTabStrip(tabs, active)
}

func (w *InventoryBagWindow) refresh(ctx Context, itemInfo *ItemWindows) {
	w.hideTooltip()
	w.ClampScroll(ctx.Session)
	w.snapshot = w.inventorySnapshot(ctx.Session)
	w.itemInfo = itemInfo
	w.SetContent(w.widgetTree(ctx, itemInfo))
	w.Publish(ctx)
}

func inventoryBagHasCart(ctx Context) bool {
	if ctx.World != nil && ctx.World.Player.HasCartState {
		return ctx.World.Player.HasCart
	}
	if selectedCharacter(ctx.Session).Option&inventoryBagCartOptionMask != 0 {
		return true
	}
	return ctx.Session != nil && (ctx.Session.Cart.MaxAmount > 0 || len(ctx.Session.Cart.Items) > 0)
}

func (w *InventoryBagWindow) startItemDragOrActivate(ctx Context, item session.InventoryItem) {
	w.hideTooltip()
	now := time.Now()
	if inventoryOpinionated(ctx) && w.source != opinionatedSourceBag {
		if w.lastClickItem == item.Index && now.Sub(w.lastClickAt) <= 360*time.Millisecond {
			opinionatedWithdraw(ctx, w.source, item)
			w.lastClickItem = 0
			w.refresh(ctx, w.itemInfo)
			return
		}
		w.dragActive = false
		w.dragItem = session.InventoryItem{}
		w.lastClickItem = item.Index
		w.lastClickAt = now
		return
	}
	if w.lastClickItem == item.Index && now.Sub(w.lastClickAt) <= 360*time.Millisecond {
		w.dragActive = false
		w.dragItem = session.InventoryItem{}
		w.activateItem(ctx, item)
		w.lastClickItem = 0
		w.refresh(ctx, w.itemInfo)
		return
	}
	w.dragItem = item
	w.dragActive = true
	w.dragFrom = now
	w.lastClickItem = item.Index
	w.lastClickAt = now
}

func (w *InventoryBagWindow) showTooltip(ctx Context, item session.InventoryItem) {
	if item.ItemID == 0 || w.dragActive || ctx.Input == nil {
		w.hideTooltip()
		return
	}
	text := inventoryItemDisplayName(ctx.Resources, item)
	w.tooltip.Show(ctx, text, ctx.Input.MouseX, ctx.Input.MouseY+inventoryBagTooltipBelow, ctx.Input.MouseY-inventoryBagTooltipAbove)
}

func (w *InventoryBagWindow) hideTooltip() {
	w.tooltip.Hide()
}

func (w *InventoryBagWindow) pointInside(x, y int) bool {
	width, height := w.currentSize(inventoryBagWidth, inventoryBagHeight)
	return pointInRect(x, y, w.x, w.y, width, height)
}

func (w *InventoryBagWindow) AcceptStorageDrop(ctx Context, item session.InventoryItem, mx, my int) bool {
	w.EnsureWindow(inventoryBagWidth, inventoryBagHeight)
	return w.IsOpen() && w.pointInside(mx, my)
}

func inventoryBagDefaultPosition(ctx Context) (int, int) {
	width, height := ctx.ScreenSize()
	menuX, menuY, _, menuH := basicMenuBounds()
	x := clampWindowInt(menuX, windowScreenMargin, maxInt(windowScreenMargin, width-inventoryBagWidth-windowScreenMargin))
	y := clampWindowInt(menuY+menuH+inventoryBagMenuGap, windowScreenMargin, maxInt(windowScreenMargin, height-inventoryBagHeight-windowScreenMargin))
	return x, y
}

func (w *InventoryBagWindow) itemIconImage(manager *res.Manager, item session.InventoryItem) image.Image {
	if manager == nil || item.ItemID == 0 {
		return nil
	}
	key := inventoryBagIconKey{itemID: item.ItemID, identified: item.Identified}
	if w.icons != nil {
		if img := w.icons[key]; img != nil {
			return img
		}
	}
	if _, ok := w.iconMiss[key]; ok {
		return nil
	}
	resourceName, ok := manager.ItemResourceName(int(item.ItemID), item.Identified)
	if !ok {
		w.markIconMiss(key)
		return nil
	}
	img, _, err := res.LoadImage(manager, res.ItemIconTextureCandidates(resourceName))
	if err != nil {
		w.markIconMiss(key)
		return nil
	}
	if w.icons == nil {
		w.icons = make(map[inventoryBagIconKey]image.Image)
	}
	w.icons[key] = img
	return img
}

func (w *InventoryBagWindow) itemIcons(ctx Context, items []session.InventoryItem) []image.Image {
	icons := make([]image.Image, len(items))
	for i, item := range items {
		icons[i] = w.itemIconImage(ctx.Resources, item)
	}
	return icons
}

func (w *InventoryBagWindow) markIconMiss(key inventoryBagIconKey) {
	if w.iconMiss == nil {
		w.iconMiss = make(map[inventoryBagIconKey]struct{})
	}
	w.iconMiss[key] = struct{}{}
}

func (w *InventoryBagWindow) activateItem(ctx Context, item session.InventoryItem) {
	if item.Type == db.ItemTypeCard {
		if ctx.Network == nil {
			glog.Warnf("card composition failed: not connected")
			return
		}
		candidates := cardCompositionCandidateItems(ctx, item)
		if len(candidates) == 0 {
			glog.Debugf("card composition has no local compatible equipment card_index=%d item=%d location=0x%04X", item.Index, item.ItemID, item.Location)
			logCardCompositionRejectedItems(ctx, item)
		} else {
			glog.Debugf("card composition local compatible equipment card_index=%d item=%d indexes=%v", item.Index, item.ItemID, inventoryItemIndexes(candidates))
		}
		w.pendingCard = item.Index
		if err := ctx.Network.SendItemCompositionList(item.Index); err != nil {
			glog.Warnf("card composition list failed: %v", err)
			return
		}
		glog.Debugf("card composition list requested index=%d item=%d", item.Index, item.ItemID)
		return
	}
	if inventoryItemCanEquip(item) {
		equipInventoryItem(ctx, item)
		return
	}
	if !db.ItemTypeIsUsable(item.Type) {
		glog.Debugf("inventory use skipped: item cannot be used index=%d item=%d type=%d", item.Index, item.ItemID, item.Type)
		return
	}
	if err := UseInventoryItem(ctx, item); err != nil {
		glog.Warnf("inventory use failed: %v", err)
		return
	}
	glog.Debugf("inventory use requested index=%d item=%d type=%d", item.Index, item.ItemID, item.Type)
}

func (w *InventoryBagWindow) ClampScroll(s *session.Session) {
	scroll := w.ensureScrollSignal()
	contentHeight := inventoryBagTotalRows(len(w.tabItems(s))) * inventoryBagCell
	maxScroll := scaledScrollMax(contentHeight, inventoryBagViewH, w.elementScale())
	value := scroll.Get()
	switch {
	case value < 0:
		scroll.Set(0)
	case value > maxScroll:
		scroll.Set(maxScroll)
	}
}

func (w *InventoryBagWindow) selectFirstNonEmptyTab(s *session.Session) {
	if len(w.tabItems(s)) > 0 {
		return
	}
	original := w.tab
	for _, tab := range w.categoryTabs(s) {
		if tab.tab == w.tab {
			continue
		}
		w.tab = tab.tab
		if len(w.tabItems(s)) > 0 {
			w.ensureScrollSignal().Set(0)
			return
		}
	}
	w.tab = original
}

func (w *InventoryBagWindow) tabItems(s *session.Session) []session.InventoryItem {
	items := w.sourceItems(s)
	if len(items) == 0 {
		return nil
	}
	if !inventoryOpinionatedSession(s) {
		filtered := make([]session.InventoryItem, 0, len(items))
		for _, item := range items {
			if inventoryItemTab(item) == w.tab {
				filtered = append(filtered, item)
			}
		}
		sort.SliceStable(filtered, func(i, j int) bool {
			return filtered[i].Index < filtered[j].Index
		})
		return filtered
	}
	if w.tab == opinionatedBagTabFavorite {
		filtered := make([]session.InventoryItem, 0, len(items))
		for _, item := range items {
			if item.Favorite {
				filtered = append(filtered, item)
			}
		}
		items = filtered
	} else if w.tab != opinionatedBagTabAll {
		filtered := make([]session.InventoryItem, 0, len(items))
		for _, item := range items {
			if opinionatedInventoryTab(item) == w.tab {
				filtered = append(filtered, item)
			}
		}
		items = filtered
	}
	items = w.matchSearch(items)
	w.sortOpinionated(items)
	return items
}

func (w *InventoryBagWindow) sourceItems(s *session.Session) []session.InventoryItem {
	if s == nil || !s.OpinionatedUI || w.source == opinionatedSourceBag {
		return sortedInventoryItems(s)
	}
	var src []session.InventoryItem
	switch w.source {
	case opinionatedSourceCart:
		src = s.Cart.Items
	case opinionatedSourceStorage:
		src = s.Storage.Items
	case opinionatedSourceGuild:
		src = s.GuildStorage.Items
	default:
		src = s.Inventory.Items
	}
	items := append([]session.InventoryItem(nil), src...)
	sort.SliceStable(items, func(i, j int) bool {
		return items[i].Index < items[j].Index
	})
	return items
}

func (w *InventoryBagWindow) matchSearch(items []session.InventoryItem) []session.InventoryItem {
	query := strings.TrimSpace(strings.ToLower(w.search))
	if query == "" {
		return items
	}
	matched := make([]session.InventoryItem, 0, len(items))
	for _, item := range items {
		if strings.Contains(strings.ToLower(w.itemName(item)), query) {
			matched = append(matched, item)
		}
	}
	return matched
}

func (w *InventoryBagWindow) itemName(item session.InventoryItem) string {
	return ItemDisplayName(w.resources, item)
}
func (w *InventoryBagWindow) sortOpinionated(items []session.InventoryItem) {
	switch w.sort {
	case opinionatedSortName:
		sort.SliceStable(items, func(i, j int) bool {
			return strings.ToLower(w.itemName(items[i])) < strings.ToLower(w.itemName(items[j]))
		})
	case opinionatedSortNameDesc:
		sort.SliceStable(items, func(i, j int) bool {
			return strings.ToLower(w.itemName(items[i])) > strings.ToLower(w.itemName(items[j]))
		})
	default:
		sort.SliceStable(items, func(i, j int) bool {
			return items[i].Index < items[j].Index
		})
	}
}

func (w *InventoryBagWindow) inventorySnapshot(s *session.Session) string {
	var b strings.Builder
	fmt.Fprintf(&b, "tab=%d;opinionated=%t;search=%s;sort=%d;source=%d;", w.tab, s != nil && s.OpinionatedUI, w.search, w.sort, w.source)
	for _, item := range sortedInventoryItems(s) {
		fmt.Fprintf(&b, "%d:%d:%d:%d:%d:%t:%t:%t;", item.Index, item.ItemID, item.Type, item.Amount, item.Location, item.Equip, item.Equipped, item.Favorite)
	}
	if s != nil && s.OpinionatedUI {
		for _, item := range s.Cart.Items {
			fmt.Fprintf(&b, "c%d:%d:%d;", item.Index, item.ItemID, item.Amount)
		}
		for _, item := range s.Storage.Items {
			fmt.Fprintf(&b, "s%d:%d:%d;", item.Index, item.ItemID, item.Amount)
		}
		for _, item := range s.GuildStorage.Items {
			fmt.Fprintf(&b, "g%d:%d:%d:%t;", item.Index, item.ItemID, item.Amount, s.GuildStorage.Open)
		}
	}
	return b.String()
}

func (w *InventoryBagWindow) ensureScrollSignal() state.Signal[float32] {
	if w.scrollY == nil {
		w.scrollY = state.NewSignal[float32](0)
	}
	return w.scrollY
}

func inventoryBagTotalRows(itemCount int) int {
	return inventoryGridTotalRows(itemCount, inventoryBagCols, inventoryBagRows)
}

func inventoryGridTotalRows(itemCount, cols, minRows int) int {
	if cols <= 0 {
		cols = 1
	}
	rows := (itemCount + cols - 1) / cols
	if rows < minRows {
		return minRows
	}
	return rows
}

func inventoryGridAmountLabels(items []session.InventoryItem) []string {
	labels := make([]string, len(items))
	for i, item := range items {
		if item.Amount > 1 {
			labels[i] = strconv.Itoa(item.Amount)
		}
	}
	return labels
}

type inventoryGridConfig struct {
	items            []session.InventoryItem
	icons            []image.Image
	amounts          []string
	cols             int
	minRows          int
	cellSize         int
	viewWidth        int
	grid             bool
	onPress          func(session.InventoryItem)
	onHover          func(session.InventoryItem)
	onLeave          func()
	onRightClick     func(session.InventoryItem, int, int)
	onRightClickMods func(session.InventoryItem, int, int, event.Modifiers)
	fillWidth        bool
}

type inventoryGridWidget struct {
	widget.WidgetBase
	cfg      inventoryGridConfig
	hovered  int
	laidCell float32
}

func newInventoryGridWidget(cfg inventoryGridConfig) *inventoryGridWidget {
	w := &inventoryGridWidget{cfg: cfg, hovered: -1}
	w.SetVisible(true)
	w.SetEnabled(true)
	return w
}

func (w *inventoryGridWidget) setItems(items []session.InventoryItem, icons []image.Image, amounts []string) {
	w.cfg.items = items
	w.cfg.icons = icons
	w.cfg.amounts = amounts
	w.hovered = -1
	w.SetNeedsRedraw(true)
}

func (w *inventoryGridWidget) Layout(ctx widget.Context, constraints geometry.Constraints) geometry.Size {
	scale := rotheme.ContentScale()
	if scale <= 0 {
		scale = 1
	}
	cell := float32(w.cellSize()) * scale
	width := float32(w.viewWidth()) * scale
	if w.cfg.grid {
		width = float32(w.cols()) * cell
	} else if w.cfg.fillWidth && opinionatedWidthBounded(constraints.MaxWidth) {
		width = constraints.MaxWidth
	}
	if !w.cfg.grid {
		if cols := w.cols(); cols > 0 && width > 0 {
			fitted := width / float32(cols)
			if w.cfg.fillWidth || (fitted > 0 && fitted < cell) {
				cell = fitted
			}
		}
	}
	pref := geometry.Sz(width, float32(w.totalRows())*cell)
	size := constraints.Constrain(pref)
	if cell <= 0 {
		cell = float32(w.cellSize())
	}
	w.laidCell = cell
	size = constraints.Constrain(geometry.Sz(float32(w.cols())*cell, float32(w.totalRows())*cell))
	w.SetBounds(geometry.FromPointSize(w.Position(), size))
	return size
}

func (w *inventoryGridWidget) Draw(ctx widget.Context, canvas widget.Canvas) {
	bounds := w.Bounds()
	canvas.DrawRect(bounds, rotheme.Default.Colors.WindowBody)
	startRow, endRow := w.visibleRows(canvas)
	cols := w.cols()
	for row := startRow; row < endRow; row++ {
		for col := 0; col < cols; col++ {
			index := row*cols + col
			cell := w.cellBounds(index)
			if w.cfg.grid {
				drawInventoryGridCell(canvas, cell, index == w.hovered)
			} else {
				drawInventoryGridCellShadow(canvas, cell, index == w.hovered)
			}
		}
	}
	startIndex := startRow * cols
	endIndex := minInt(len(w.cfg.items), endRow*cols)
	for i := startIndex; i < endIndex; i++ {
		item := w.cfg.items[i]
		cell := w.cellBounds(i)
		if i < len(w.cfg.icons) && w.cfg.icons[i] != nil {
			pad := cell.Width() * inventoryBagIconPad / float32(inventoryBagCell)
			drawFittedImage(canvas, w.cfg.icons[i], geometry.NewRect(cell.Min.X+pad, cell.Min.Y+pad, cell.Width()-pad*2, cell.Height()-pad*2))
		}
		if item.Amount > 1 {
			amount := ""
			if i < len(w.cfg.amounts) {
				amount = w.cfg.amounts[i]
			}
			if amount == "" {
				amount = strconv.Itoa(item.Amount)
			}
			amountW := cell.Width() * inventoryBagAmountW / float32(inventoryBagCell)
			amountH := cell.Height() * inventoryBagAmountH / float32(inventoryBagCell)
			amountX := cell.Width() * inventoryBagAmountInsetX / float32(inventoryBagCell)
			amountY := cell.Height() * inventoryBagAmountInsetY / float32(inventoryBagCell)
			rotheme.DrawText(canvas, amount, geometry.NewRect(cell.Max.X-amountX, cell.Max.Y-amountY, amountW, amountH), rotheme.Default.Typography.TextSize, rotheme.Default.Colors.Text, false, widget.TextAlignRight)
		}
		if item.Equipped {
			equip := cell.Width() * inventoryBagEquipSize / float32(inventoryBagCell)
			equipInset := cell.Width() * inventoryBagEquipInset / float32(inventoryBagCell)
			rotheme.DrawText(canvas, "E", geometry.NewRect(cell.Min.X+equipInset, cell.Min.Y+equipInset, equip, equip), rotheme.Default.Typography.TextSize, inventoryBagEquippedColor, false, widget.TextAlignLeft)
		}
	}
}

func drawInventoryGridCell(canvas widget.Canvas, cell geometry.Rect, hovered bool) {
	fill := rotheme.Default.Colors.PanelBody
	if hovered {
		fill = rotheme.Default.Colors.ButtonHover
	}
	canvas.DrawRect(cell, fill)
	canvas.StrokeRect(cell, rotheme.Default.Colors.WindowBorder, windowBorderWidth)
}

func drawInventoryGridCellShadow(canvas widget.Canvas, cell geometry.Rect, hovered bool) {
	width := cell.Width() - inventoryBagCellShadowInset*2
	height := (cell.Height() - inventoryBagCellShadowInset*2) / 2
	if width <= 0 || height <= 0 {
		return
	}
	shadow := geometry.NewRect(
		cell.Min.X+inventoryBagCellShadowInset,
		cell.Max.Y-inventoryBagCellShadowInset-height,
		width,
		height,
	)
	color := rotheme.Default.Colors.ButtonHover
	color.A = inventoryBagCellShadowAlpha
	if hovered {
		color.A = inventoryBagCellHoverAlpha
	}
	canvas.DrawRoundRect(shadow, color, min(inventoryBagCellShadowRadius, height/2))
}

func (w *inventoryGridWidget) visibleRows(canvas widget.Canvas) (int, int) {
	rows := w.totalRows()
	if rows <= 0 {
		return 0, 0
	}
	clip := canvas.ClipBounds()
	if clip.IsEmpty() {
		return 0, 0
	}
	offset := canvas.TransformOffset()
	cell := w.laidCell
	if cell <= 0 {
		cell = float32(w.cellSize())
	}
	localTop := clip.Min.Y - offset.Y
	localBottom := clip.Max.Y - offset.Y
	if localBottom <= 0 || localTop >= float32(rows)*cell {
		return 0, 0
	}
	start := maxInt(0, int(localTop/cell))
	end := minInt(rows, int(localBottom/cell)+1)
	if end < start {
		end = start
	}
	return start, end
}

func (w *inventoryGridWidget) Event(ctx widget.Context, e event.Event) bool {
	switch ev := e.(type) {
	case *event.MouseEvent:
		index := w.indexAt(ev.Position)
		switch ev.MouseType {
		case event.MouseEnter, event.MouseMove:
			w.hovered = index
			if index >= 0 && index < len(w.cfg.items) {
				if w.cfg.onHover != nil {
					w.cfg.onHover(w.cfg.items[index])
				}
				ctx.SetCursor(widget.CursorPointer)
			} else {
				if w.cfg.onLeave != nil {
					w.cfg.onLeave()
				}
				ctx.SetCursor(widget.CursorDefault)
			}
			return true
		case event.MouseLeave:
			w.hovered = -1
			if w.cfg.onLeave != nil {
				w.cfg.onLeave()
			}
			ctx.SetCursor(widget.CursorDefault)
			return false
		case event.MousePress:
			if index < 0 || index >= len(w.cfg.items) {
				return true
			}
			item := w.cfg.items[index]
			switch ev.Button {
			case event.ButtonLeft:
				if w.cfg.onPress != nil {
					w.cfg.onPress(item)
				}
			case event.ButtonRight:
				if w.cfg.onRightClickMods != nil {
					w.cfg.onRightClickMods(item, int(ev.GlobalPosition.X), int(ev.GlobalPosition.Y), ev.Modifiers())
				} else if w.cfg.onRightClick != nil {
					w.cfg.onRightClick(item, int(ev.GlobalPosition.X), int(ev.GlobalPosition.Y))
				}
			}
			return true
		}
		return true
	}
	return false
}

func (w *inventoryGridWidget) cellBounds(index int) geometry.Rect {
	bounds := w.Bounds()
	cell := w.laidCell
	if cell <= 0 {
		cell = float32(w.cellSize())
	}
	col := index % w.cols()
	row := index / w.cols()
	return geometry.NewRect(
		bounds.Min.X+float32(col)*cell,
		bounds.Min.Y+float32(row)*cell,
		cell-1,
		cell-1,
	)
}

func (w *inventoryGridWidget) indexAt(point geometry.Point) int {
	local := point.Sub(w.Bounds().Min)
	cell := w.laidCell
	if cell <= 0 {
		cell = float32(w.cellSize())
	}
	if local.X < 0 || local.Y < 0 || local.X >= float32(w.cols())*cell || local.Y >= float32(w.totalRows())*cell {
		return -1
	}
	col := int(local.X / cell)
	row := int(local.Y / cell)
	index := row*w.cols() + col
	if index < 0 || index >= len(w.cfg.items) {
		return -1
	}
	return index
}

func (w *inventoryGridWidget) cols() int {
	if w.cfg.cols > 0 {
		return w.cfg.cols
	}
	return inventoryBagCols
}

func (w *inventoryGridWidget) minRows() int {
	if w.cfg.minRows > 0 {
		return w.cfg.minRows
	}
	return inventoryBagRows
}

func (w *inventoryGridWidget) cellSize() int {
	if w.cfg.cellSize > 0 {
		return w.cfg.cellSize
	}
	return inventoryBagCell
}

func (w *inventoryGridWidget) gridWidth() int {
	return w.cols() * w.cellSize()
}

func (w *inventoryGridWidget) viewWidth() int {
	if w.cfg.viewWidth > 0 {
		return w.cfg.viewWidth
	}
	return w.gridWidth() + ROScrollbarGutter
}

func (w *inventoryGridWidget) totalRows() int {
	return inventoryGridTotalRows(len(w.cfg.items), w.cols(), w.minRows())
}

func inventoryItemTab(item session.InventoryItem) int {
	if inventoryItemUsesEquipmentTab(item) {
		return inventoryBagTabEquip
	}
	if db.ItemTypeIsUsable(item.Type) {
		return inventoryBagTabItem
	}
	return inventoryBagTabEtc
}

func opinionatedInventoryTab(item session.InventoryItem) int {
	switch item.Type {
	case db.ItemTypeAmmo:
		return opinionatedBagTabAmmo
	case db.ItemTypeCard:
		return opinionatedBagTabCard
	}
	if inventoryItemUsesEquipmentTab(item) {
		return opinionatedBagTabEquip
	}
	if db.ItemTypeIsUsable(item.Type) {
		return opinionatedBagTabUsable
	}
	return opinionatedBagTabEtc
}

func inventoryOpinionated(ctx Context) bool {
	return inventoryOpinionatedSession(ctx.Session)
}

func inventoryOpinionatedSession(s *session.Session) bool {
	return s != nil && s.OpinionatedUI
}

func (w *InventoryBagWindow) syncStyle(s *session.Session) {
	opinionated := inventoryOpinionatedSession(s)
	if opinionated == w.opinionated {
		return
	}
	w.opinionated = opinionated
	w.tab = 0
}

func (w *InventoryBagWindow) categoryTabs(s *session.Session) []struct {
	label string
	tab   int
} {
	if inventoryOpinionatedSession(s) {
		return opinionatedInventoryTabs
	}
	return inventoryBagTabs
}

func inventoryItemCanEquip(item session.InventoryItem) bool {
	if item.Type == db.ItemTypeCard {
		return false
	}
	return item.Equip || inventoryItemTypeCanEquip(item.Type)
}

func inventoryItemUsesEquipmentTab(item session.InventoryItem) bool {
	switch item.Type {
	case db.ItemTypeCard, db.ItemTypeAmmo:
		return false
	default:
		return item.Equip || inventoryItemTypeUsesEquipmentTab(item.Type)
	}
}

func inventoryItemTypeCanEquip(itemType uint8) bool {
	switch itemType {
	case db.ItemTypeArmor, db.ItemTypeWeapon, db.ItemTypePetEgg, db.ItemTypePetArmor, db.ItemTypeAmmo, db.ItemTypeShadowGear:
		return true
	default:
		return false
	}
}

func inventoryItemTypeUsesEquipmentTab(itemType uint8) bool {
	switch itemType {
	case db.ItemTypeArmor, db.ItemTypeWeapon, db.ItemTypePetEgg, db.ItemTypePetArmor, db.ItemTypeShadowGear:
		return true
	default:
		return false
	}
}

func cardCompositionCandidateItems(ctx Context, card session.InventoryItem) []session.InventoryItem {
	if ctx.Session == nil || ctx.Resources == nil || card.Type != db.ItemTypeCard || card.Location == 0 {
		return nil
	}
	candidates := make([]session.InventoryItem, 0, len(ctx.Session.Inventory.Items))
	for _, item := range ctx.Session.Inventory.Items {
		if ok, _ := cardCompositionCanTarget(ctx, card, item); !ok {
			continue
		}
		candidates = append(candidates, item)
	}
	return candidates
}

func cardCompositionCanTarget(ctx Context, card, item session.InventoryItem) (bool, string) {
	switch item.Type {
	case db.ItemTypeWeapon, db.ItemTypeArmor:
	default:
		return false, "not weapon or armor"
	}
	if !inventoryItemCanShowSlots(item) {
		return false, "special card slots"
	}
	if !item.Identified {
		return false, "unidentified"
	}
	if item.Location&card.Location == 0 {
		return false, "incompatible location"
	}
	slotCount, ok := ctx.Resources.ItemSlotCount(int(item.ItemID))
	if !ok || slotCount <= 0 {
		return false, "no slots in resource db"
	}
	for i := 0; i < slotCount && i < len(item.Cards); i++ {
		if item.Cards[i] == 0 {
			if item.Equipped {
				return false, "equipped"
			}
			return true, ""
		}
	}
	return false, "no empty slot"
}

func logCardCompositionRejectedItems(ctx Context, card session.InventoryItem) {
	if ctx.Session == nil || ctx.Resources == nil {
		return
	}
	for _, item := range ctx.Session.Inventory.Items {
		switch item.Type {
		case db.ItemTypeWeapon, db.ItemTypeArmor:
		default:
			continue
		}
		if ok, reason := cardCompositionCanTarget(ctx, card, item); !ok {
			glog.Warnf("card composition rejected equipment index=%d item=%d type=%d location=0x%04X identified=%v equipped=%v cards=%v reason=%q", item.Index, item.ItemID, item.Type, item.Location, item.Identified, item.Equipped, item.Cards, reason)
		}
	}
}

func inventoryItemIndexes(items []session.InventoryItem) []uint16 {
	indexes := make([]uint16, len(items))
	for i, item := range items {
		indexes[i] = item.Index
	}
	return indexes
}

func inventoryItemEquipLocation(item session.InventoryItem) uint32 {
	if item.Location != 0 {
		return item.Location
	}
	if item.Type == db.ItemTypeAmmo {
		return db.EquipAmmo
	}
	return 0
}
