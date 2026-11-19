package ui

import (
	"fmt"
	"strings"

	"github.com/gogpu/ui/core/scrollview"
	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/state"
	"github.com/gogpu/ui/widget"
	"github.com/kivutar/goro/db"
	"github.com/kivutar/goro/network"
	"github.com/kivutar/goro/session"
	"github.com/kivutar/goro/ui/rotheme"
)

type CompanionTab int

const (
	CompanionTabPet CompanionTab = iota
	CompanionTabHomunculus
	CompanionTabMercenary
	CompanionTabStable
)

func ClampCompanionTab(tab int) CompanionTab {
	if tab < int(CompanionTabPet) || tab > int(CompanionTabStable) {
		return CompanionTabPet
	}
	return CompanionTab(tab)
}

type CompanionWindowActionKind uint8

const (
	CompanionActionNone CompanionWindowActionKind = iota
	CompanionActionRequestLists
	CompanionActionHomunculusStore
	CompanionActionHomunculusSwap
	CompanionActionPetStore
	CompanionActionPetSummon
	CompanionActionPetDepositEgg
	CompanionActionPetToEgg
	CompanionActionPetFeed
	CompanionActionPetUnequip
	CompanionActionPetEvolve
	CompanionActionPetPerformance
	CompanionActionPetClassicFeed
	CompanionActionHomunculusSkill
	CompanionActionHomunculusFeed
	CompanionActionHomunculusDelete
	CompanionActionMercenarySkill
	CompanionActionMercenaryDelete
	CompanionActionPetAutoFeed
	CompanionActionHomunAutoFeed
)

type CompanionWindowAction struct {
	Kind         CompanionWindowActionKind
	HomunculusID uint32
	PetID        uint32
	EvolveEggID  uint32
}

type CompanionWindow struct {
	Window
	tab           CompanionTab
	pet           network.PetProperty
	hasPet        bool
	homunculus    session.Companion
	mercenary     session.Companion
	homunList     network.StableHomunculusList
	petList       network.StablePetList
	owned         map[uint32]uint32
	selectedPetID uint32
	snapshot      string
	action        CompanionWindowAction
	scrollY       state.Signal[float32]
	ctx           Context
}

func (w *CompanionWindow) Tab() CompanionTab { return w.tab }

func (w *CompanionWindow) SetTab(tab CompanionTab) {
	w.tab = ClampCompanionTab(int(tab))
}

func (w *CompanionWindow) Toggle(ctx Context) {
	if w.IsOpen() {
		w.Close()
		w.Publish(ctx)
		return
	}
	w.OpenWindow(ctx)
}

func (w *CompanionWindow) OpenWindow(ctx Context) {
	w.EnsureWindow(companionWindowW, companionWindowH)
	w.ctx = ctx
	w.ensureScroll()
	w.syncFromSession(ctx)
	w.snapshot = w.stateSnapshot()
	w.action = CompanionWindowAction{Kind: CompanionActionRequestLists}
	w.Window.Open(ctx, w.widgetTree(ctx))
	w.Publish(ctx)
}

func (w *CompanionWindow) OpenTab(ctx Context, tab CompanionTab) {
	w.tab = tab
	if w.IsOpen() {
		w.refresh(ctx)
		return
	}
	w.OpenWindow(ctx)
}

func (w *CompanionWindow) Rebind(ctx Context) {
	w.EnsureWindow(companionWindowW, companionWindowH)
	w.ctx = ctx
	w.ensureScroll()
	if !w.IsOpen() {
		return
	}
	w.syncFromSession(ctx)
	w.snapshot = w.stateSnapshot()
	w.SetContent(w.widgetTree(ctx))
	w.Publish(ctx)
}

func (w *CompanionWindow) SetHomunculusList(list network.StableHomunculusList) {
	w.homunList = list
}

func (w *CompanionWindow) SetPetList(list network.StablePetList) {
	w.petList = list
	if w.selectedPetID != 0 {
		found := false
		for _, row := range list.List {
			if row.PetID == w.selectedPetID {
				found = true
				break
			}
		}
		if !found {
			w.selectedPetID = 0
		}
	}
}

func (w *CompanionWindow) SetOwnedCounts(list network.StableOwnedList) {
	owned := make(map[uint32]uint32, len(list.List))
	for _, row := range list.List {
		owned[row.ItemID] = row.Owned
	}
	w.owned = owned
}

func (w *CompanionWindow) HomunculusList() network.StableHomunculusList { return w.homunList }

func (w *CompanionWindow) PetList() network.StablePetList { return w.petList }

func (w *CompanionWindow) Update(ctx Context, pet network.PetProperty, hasPet bool) bool {
	w.EnsureWindow(companionWindowW, companionWindowH)
	w.ctx = ctx
	w.pet = pet
	w.hasPet = hasPet
	w.syncFromSession(ctx)
	if !w.IsOpen() {
		return false
	}
	if next := w.stateSnapshot(); next != w.snapshot {
		w.snapshot = next
		w.SetContent(w.widgetTree(ctx))
	}
	consumed := w.Window.Update(ctx)
	hasAction := w.action.Kind != CompanionActionNone
	w.Publish(ctx)
	return consumed || hasAction
}

func (w *CompanionWindow) PopAction() CompanionWindowAction {
	action := w.action
	w.action = CompanionWindowAction{}
	return action
}

func (w *CompanionWindow) syncFromSession(ctx Context) {
	if ctx.Session == nil {
		w.homunculus = session.Companion{}
		w.mercenary = session.Companion{}
		return
	}
	w.homunculus = ctx.Session.Homunculus
	w.mercenary = ctx.Session.Mercenary
}

func (w *CompanionWindow) ensureScroll() {
	if w.scrollY == nil {
		w.scrollY = state.NewSignal[float32](0)
	}
}

func (w *CompanionWindow) refresh(ctx Context) {
	if !w.IsOpen() {
		return
	}
	w.syncFromSession(ctx)
	w.snapshot = w.stateSnapshot()
	w.SetContent(w.widgetTree(ctx))
	w.Publish(ctx)
}

func (w *CompanionWindow) emit(action CompanionWindowAction) {
	w.action = action
}

func (w *CompanionWindow) widgetTree(ctx Context) widget.Widget {
	return Win(
		Title("Homunculus & Pet Stable"),
		CloseButton(true),
		OnClose(func() {
			w.Close()
			w.Publish(ctx)
		}),
		Size(float32(companionWindowW), float32(companionWindowH)),
		Content(w.tabFrame(ctx)),
		Footer(w.tabFooter()...),
	)
}

func (w *CompanionWindow) tabFrame(ctx Context) widget.Widget {
	return primitives.Box(
		w.tabStrip(),
		primitives.Expanded(
			primitives.Box(w.tabContent(ctx)).
				Background(rotheme.Default.Colors.WindowBody).
				Padding(float32(companionWindowPad)).
				CrossAlign(primitives.CrossAxisStretch),
		),
	).
		Gap(float32(companionFrameGap)).
		Background(Color(companionFrameBackground)).
		CrossAlign(primitives.CrossAxisStretch)
}

func (w *CompanionWindow) tabStrip() widget.Widget {
	tabs := make([]widget.Widget, 0, 5)
	for _, def := range []struct {
		tab   CompanionTab
		label string
	}{
		{CompanionTabPet, "Pet"},
		{CompanionTabHomunculus, "Homunculus"},
		{CompanionTabMercenary, "Mercenary"},
		{CompanionTabStable, "Stable"},
	} {
		def := def
		tabs = append(tabs, newTabWidget(tabWidgetConfig{
			label:  def.label,
			active: w.tab == def.tab,
			width:  companionWindowTabW,
			height: companionWindowTabH,
			onClick: func() {
				if w.tab == def.tab {
					return
				}
				w.tab = def.tab
				w.scrollY.Set(0)
				if def.tab == CompanionTabStable {
					w.emit(CompanionWindowAction{Kind: CompanionActionRequestLists})
				}
				w.refresh(w.ctx)
			},
		}))
	}
	tabs = append(tabs, primitives.Expanded(primitives.Box()))
	return primitives.HBox(tabs...).
		Gap(-1).
		CrossAlign(primitives.CrossAxisStretch).
		Background(rotheme.Default.Colors.FooterLine)
}

func (w *CompanionWindow) tabContent(ctx Context) widget.Widget {
	switch w.tab {
	case CompanionTabPet:
		return w.petTab(ctx)
	case CompanionTabHomunculus:
		return w.homunculusTab(ctx)
	case CompanionTabMercenary:
		return w.mercenaryTab(ctx)
	default:
		return w.stableTab(ctx)
	}
}

func (w *CompanionWindow) tabFooter() []widget.Widget {
	buttons := []widget.Widget{primitives.Expanded(primitives.Box())}
	switch w.tab {
	case CompanionTabPet:
		if w.hasPet {
			buttons = append(buttons,
				rotheme.Button("Feed", func() {
					w.emit(CompanionWindowAction{Kind: CompanionActionPetClassicFeed})
				}).Width(float32(companionButtonW)),
				rotheme.Button("Perform", func() {
					w.emit(CompanionWindowAction{Kind: CompanionActionPetPerformance})
				}).Width(float32(companionButtonW)),
				rotheme.Button("To egg", func() {
					w.emit(CompanionWindowAction{Kind: CompanionActionPetToEgg, PetID: w.petList.ActivePetID})
				}).Width(float32(companionButtonW)),
				rotheme.Button("Auto", func() {
					w.emit(CompanionWindowAction{Kind: CompanionActionPetAutoFeed})
				}).Width(float32(companionButtonW)),
			)
		}
	case CompanionTabHomunculus:
		if w.homunculus.Active {
			buttons = append(buttons,
				rotheme.Button("Skill", func() {
					w.emit(CompanionWindowAction{Kind: CompanionActionHomunculusSkill})
				}).Width(float32(companionButtonW)),
				rotheme.Button("Feed", func() {
					w.emit(CompanionWindowAction{Kind: CompanionActionHomunculusFeed})
				}).Width(float32(companionButtonW)),
				rotheme.Button("Store", func() {
					w.emit(CompanionWindowAction{Kind: CompanionActionHomunculusStore})
				}).Width(float32(companionButtonW)),
				rotheme.Button("Auto", func() {
					w.emit(CompanionWindowAction{Kind: CompanionActionHomunAutoFeed})
				}).Width(float32(companionButtonW)),
			)
		}
	case CompanionTabMercenary:
		if w.mercenary.Active {
			buttons = append(buttons,
				rotheme.Button("Skill", func() {
					w.emit(CompanionWindowAction{Kind: CompanionActionMercenarySkill})
				}).Width(float32(companionButtonW)),
				rotheme.Button("Dismiss", func() {
					w.emit(CompanionWindowAction{Kind: CompanionActionMercenaryDelete})
				}).Width(float32(companionButtonW)),
			)
		}
	case CompanionTabStable:
		buttons = append(buttons,
			rotheme.Button("Store pet", func() {
				w.emit(CompanionWindowAction{Kind: CompanionActionPetStore})
			}).Width(float32(companionStorePetButtonW)),
			rotheme.Button("Store homun", func() {
				w.emit(CompanionWindowAction{Kind: CompanionActionHomunculusStore})
			}).Width(float32(companionWideButtonW)),
			rotheme.Button("Refresh", func() {
				w.emit(CompanionWindowAction{Kind: CompanionActionRequestLists})
			}).Width(float32(companionButtonW)),
		)
	}
	return buttons
}

func (w *CompanionWindow) petTab(ctx Context) widget.Widget {
	if !w.hasPet {
		return companionEmptyText("No pet. Hatch an egg, then Deposit it on the Stable tab.")
	}
	rows := []widget.Widget{
		companionInfoRow("Name", w.pet.Name),
		companionInfoRow("Level", fmt.Sprintf("%d", w.pet.Level)),
		companionInfoRow("Hunger", fmt.Sprintf("%s (%d/100)", petHungerText(w.pet.Fullness), w.pet.Fullness)),
		companionInfoRow("Intimacy", fmt.Sprintf("%s (%d/1000)", petIntimacyText(w.pet.Relationship), w.pet.Relationship)),
		companionInfoRow("Accessory", companionItemName(ctx, uint32(w.pet.AccessoryID))),
	}
	if stable := w.activeStablePet(); stable != nil {
		rows = append(rows,
			companionInfoRow("ATK", fmt.Sprintf("%d ~ %d", stable.Atk, stable.Atk2)),
			companionInfoRow("EXP", companionPetExpText(*stable)),
			companionInfoRow("Bonus", companionPetBuffText(*stable)),
			companionInfoRow("Food", companionItemName(ctx, stable.FoodID)),
		)
		if stable.Intimate >= 900 {
			rows = append(rows, rotheme.Text("Evolution uses the Stable tab once recipes are available.").Color(rotheme.Default.Colors.MutedText))
		}
	}
	return primitives.Box(rows...).Gap(float32(companionStatRowGap))
}

func (w *CompanionWindow) homunculusTab(_ Context) widget.Widget {
	if !w.homunculus.Active {
		return companionEmptyText("No homunculus. Call Homunculus with an Embryo, then Store it on the Stable tab to raise another type.")
	}
	h := HomunculusInfoWindow{companion: w.homunculus}
	info := primitives.HBox(
		h.statsTable(),
		primitives.Box(
			companionInfoRow("Name", displayHomunculusName(w.homunculus)),
			companionInfoRow("Level", fmt.Sprintf("%d", w.homunculus.Level)),
			vitalsRow("HP", w.homunculus.HP, w.homunculus.MaxHP, 220),
			vitalsRow("SP", w.homunculus.SP, w.homunculus.MaxSP, 220),
			h.expBarRow("EXP", w.homunculus.Exp, w.homunculus.MaxExp, Color(homunculusInfoExpColor)),
			h.barRow("Hunger", w.homunculus.Hunger, 100, Color(HungerBarFillColor(w.homunculus.Hunger, 100)), true),
			companionInfoRow("Intimacy", HomunculusIntimacyText(w.homunculus.Intimacy)),
			companionInfoRow("Skill Pt", fmt.Sprintf("%d", w.homunculus.Skills.Points)),
		).Gap(float32(homunculusInfoRowGap)),
	).Gap(float32(homunculusInfoColumnGap)).CrossAlign(primitives.CrossAxisStart)

	children := []widget.Widget{info, rotheme.Text("Lineage swap").Color(rotheme.Default.Colors.MutedText)}
	if len(w.homunList.List) == 0 {
		children = append(children, rotheme.Text("Open the Stable tab to load stored homunculi.").Color(rotheme.Default.Colors.MutedText))
	} else {
		for _, row := range w.homunList.List {
			children = append(children, w.homunculusRow(row))
		}
	}
	return w.scrollBox(children...)
}

func (w *CompanionWindow) mercenaryTab(_ Context) widget.Widget {
	if !w.mercenary.Active {
		return companionEmptyText("No mercenary under contract.")
	}
	h := HomunculusInfoWindow{companion: w.mercenary}
	m := MercenaryInfoWindow{companion: w.mercenary}
	timeFill := Color(mercenaryInfoTimeColor)
	if mercenaryRemainingRatio(w.mercenary.ExpireTick) < 0.25 {
		timeFill = Color(mercenaryInfoTimeLow)
	}
	return primitives.HBox(
		m.statsTable(),
		primitives.Box(
			companionInfoRow("Name", displayMercenaryName(w.mercenary)),
			companionInfoRow("Level", fmt.Sprintf("%d", w.mercenary.Level)),
			vitalsRow("HP", w.mercenary.HP, w.mercenary.MaxHP, 220),
			vitalsRow("SP", w.mercenary.SP, w.mercenary.MaxSP, 220),
			h.barRow("Time", mercenaryRemainingSeconds(w.mercenary.ExpireTick), mercenaryLifetimeSeconds, timeFill, false),
			companionInfoRow("Left", formatMercenaryExpireDate(w.mercenary.ExpireTick)),
			h.barRow("Kills", int(w.mercenary.Kills%50), 50, Color(mercenaryInfoKillsColor), true),
			companionInfoRow("Faith", fmt.Sprintf("%d", w.mercenary.Faith)),
		).Gap(float32(homunculusInfoRowGap)),
	).Gap(float32(homunculusInfoColumnGap)).CrossAlign(primitives.CrossAxisStart)
}

func (w *CompanionWindow) stableTab(ctx Context) widget.Widget {
	children := []widget.Widget{
		rotheme.Text("Homunculi").Color(rotheme.Default.Colors.TitleText),
	}
	if len(w.homunList.List) == 0 {
		children = append(children, companionEmptyText("No homunculi yet. Alchemists: Call Homunculus with an Embryo, then Store it here to raise another type."))
	} else {
		for _, row := range w.homunList.List {
			children = append(children, w.homunculusRow(row))
		}
	}
	children = append(children, rotheme.Text("Pets").Color(rotheme.Default.Colors.TitleText))
	if len(w.petList.List) == 0 {
		children = append(children, companionEmptyText("No pets in the account stable. Hatch or pick up an egg, then Deposit it here. Current pets can be Stored from this tab."))
	} else {
		for _, row := range w.petList.List {
			children = append(children, w.petRow(ctx, row))
		}
		if selected := w.selectedStablePet(); selected != nil {
			children = append(children, w.petDetail(ctx, *selected))
		}
	}
	return w.scrollBox(children...)
}

func (w *CompanionWindow) homunculusRow(row network.StableHomunculusEntry) widget.Widget {
	active := row.HomunID == w.homunList.ActiveHomunID
	lineage := HomunculusLineageName(HomunculusLineage(int(row.ClassID), int(row.PrevClass)))
	name := row.Name
	if name == "" {
		name = lineage
	}
	badge := lineage
	if active {
		badge += " · Active"
	}
	meta := fmt.Sprintf("Lv %d · %s · Hunger %d · %s", row.Level, StableHomunculusIntimacyText(int(row.Intimacy)), row.Hunger, CompanionClassName(int(row.ClassID)))
	label := primitives.Box(
		rotheme.Text(fmt.Sprintf("%s  (%s)", name, badge)),
		rotheme.Text(meta).Color(rotheme.Default.Colors.MutedText),
	).Gap(float32(companionRowTextGap))
	rowWidgets := []widget.Widget{primitives.Expanded(label)}
	if !active {
		id := row.HomunID
		rowWidgets = append(rowWidgets, rotheme.Button("Swap in", func() {
			w.emit(CompanionWindowAction{Kind: CompanionActionHomunculusSwap, HomunculusID: id})
		}).Width(float32(companionButtonW)))
	}
	return primitives.HBox(rowWidgets...).Gap(float32(companionRowGap)).CrossAlign(primitives.CrossAxisCenter).Height(float32(companionListRowH))
}

func (w *CompanionWindow) petRow(_ Context, row network.StablePetEntry) widget.Widget {
	active := row.PetID == w.petList.ActivePetID
	selected := row.PetID == w.selectedPetID
	title := row.Name
	if title == "" {
		title = CompanionClassName(int(row.ClassID))
	}
	sub := fmt.Sprintf("Lv %d · %s", row.Level, stablePetStatus(row))
	if active {
		sub += " · Out"
	}
	id := row.PetID
	label := primitives.Box(
		rotheme.Text(title),
		rotheme.Text(sub).Color(rotheme.Default.Colors.MutedText),
	).Gap(float32(companionRowTextGap))
	bg := rotheme.Default.Colors.WindowBody
	if selected {
		bg = rotheme.Default.Colors.ButtonHover
	} else if active {
		bg = Color(PanelAltColor)
	}
	return primitives.HBox(
		primitives.Expanded(label),
		rotheme.Button("Details", func() {
			w.selectedPetID = id
			w.refresh(w.ctx)
		}).Width(float32(companionButtonW)),
	).
		Gap(float32(companionRowGap)).
		CrossAlign(primitives.CrossAxisCenter).
		Height(float32(companionListRowH)).
		Background(bg)
}

func (w *CompanionWindow) petDetail(ctx Context, row network.StablePetEntry) widget.Widget {
	children := []widget.Widget{
		rotheme.Text("Pet details").Color(rotheme.Default.Colors.TitleText),
		companionInfoRow("Name", row.Name),
		companionInfoRow("Species", fmt.Sprintf("%s (%d)", CompanionClassName(int(row.ClassID)), row.ClassID)),
		companionInfoRow("Level", fmt.Sprintf("%d", row.Level)),
		companionInfoRow("ATK", fmt.Sprintf("%d ~ %d", row.Atk, row.Atk2)),
		companionInfoRow("ASPD", fmt.Sprintf("%d", HomunculusASPDDisplay(int(row.ASPD)))),
		companionInfoRow("Hunger", fmt.Sprintf("%s (%d/100)", petHungerText(row.Hungry), row.Hungry)),
		companionInfoRow("Intimacy", fmt.Sprintf("%s (%d/1000)", petIntimacyText(row.Intimate), row.Intimate)),
		companionInfoRow("Accessory", companionItemName(ctx, row.Equip)),
		companionInfoRow("Bonus", companionPetBuffText(row)),
		companionInfoRow("Food", companionItemName(ctx, row.FoodID)),
		companionInfoRow("Status", stablePetStatus(row)),
		companionInfoRow("Renamed", yesNo(row.RenameFlag != 0)),
		companionInfoRow("Owner", emptyDash(row.OwnerName)),
		companionInfoRow("EXP", companionPetExpText(row)),
	}
	actions := []widget.Widget{}
	egg := row.Flags&network.StableFlagInvEgg != 0
	active := row.PetID == w.petList.ActivePetID
	if egg {
		id := row.PetID
		actions = append(actions, rotheme.Button("Deposit egg", func() {
			w.emit(CompanionWindowAction{Kind: CompanionActionPetDepositEgg, PetID: id})
		}).Width(float32(companionWideButtonW)))
	} else if !active {
		id := row.PetID
		actions = append(actions, rotheme.Button("Summon", func() {
			w.emit(CompanionWindowAction{Kind: CompanionActionPetSummon, PetID: id})
		}).Width(float32(companionMediumButtonW)))
	}
	if !egg {
		id := row.PetID
		actions = append(actions, rotheme.Button("To egg", func() {
			w.emit(CompanionWindowAction{Kind: CompanionActionPetToEgg, PetID: id})
		}).Width(float32(companionButtonW)))
	}
	if !egg && active {
		id := row.PetID
		actions = append(actions, rotheme.Button("Feed", func() {
			w.emit(CompanionWindowAction{Kind: CompanionActionPetFeed, PetID: id})
		}).Width(float32(companionButtonW)))
		if row.Equip != 0 {
			actions = append(actions, rotheme.Button("Unequip", func() {
				w.emit(CompanionWindowAction{Kind: CompanionActionPetUnequip, PetID: id})
			}).Width(float32(companionMediumButtonW)))
		}
		if row.Intimate >= 900 {
			actions = append(actions, rotheme.Button("Evolve", func() {
				w.emit(CompanionWindowAction{Kind: CompanionActionPetEvolve, PetID: id})
			}).Width(float32(companionButtonW)))
		}
	}
	if len(actions) > 0 {
		children = append(children, primitives.HBox(actions...).Gap(float32(companionActionGap)))
	}
	if len(w.owned) > 0 {
		children = append(children, rotheme.Text("Owned evolution materials (inventory + storage)").Color(rotheme.Default.Colors.MutedText))
		for _, entry := range ownedEntries(w.owned) {
			children = append(children, companionInfoRow(companionItemName(ctx, entry.ItemID), fmt.Sprintf("%d", entry.Owned)))
		}
	}
	return primitives.Box(children...).Gap(float32(companionDetailGap)).Padding(float32(companionDetailPad)).Background(Color(PanelAltColor))
}

func (w *CompanionWindow) scrollBox(children ...widget.Widget) widget.Widget {
	w.ensureScroll()
	return scrollview.New(
		primitives.Box(children...).Gap(float32(companionListGap)).CrossAlign(primitives.CrossAxisStretch),
		scrollview.DirectionOpt(scrollview.Vertical),
		scrollview.ScrollbarOpt(scrollview.ScrollbarAuto),
		scrollview.ScrollYSignal(w.scrollY),
		scrollview.ScrollStep(24),
	)
}

func (w *CompanionWindow) activeStablePet() *network.StablePetEntry {
	if w.petList.ActivePetID == 0 {
		return nil
	}
	for i := range w.petList.List {
		if w.petList.List[i].PetID == w.petList.ActivePetID {
			return &w.petList.List[i]
		}
	}
	return nil
}

func (w *CompanionWindow) selectedStablePet() *network.StablePetEntry {
	if w.selectedPetID == 0 {
		return nil
	}
	for i := range w.petList.List {
		if w.petList.List[i].PetID == w.selectedPetID {
			return &w.petList.List[i]
		}
	}
	return nil
}

func (w *CompanionWindow) stateSnapshot() string {
	var b strings.Builder
	fmt.Fprintf(&b, "tab=%d;sel=%d;pet=%t,%s,%d,%d,%d,%d;hom=%t,%s,%d,%d/%d,%d,%d;merc=%t,%s,%d,%d/%d,%d,%d;",
		w.tab, w.selectedPetID,
		w.hasPet, w.pet.Name, w.pet.Level, w.pet.Fullness, w.pet.Relationship, w.pet.AccessoryID,
		w.homunculus.Active, w.homunculus.Name, w.homunculus.Level, w.homunculus.HP, w.homunculus.MaxHP, w.homunculus.Hunger, w.homunculus.Intimacy,
		w.mercenary.Active, w.mercenary.Name, w.mercenary.Level, w.mercenary.HP, w.mercenary.MaxHP, w.mercenary.ExpireTick, w.mercenary.Kills,
	)
	fmt.Fprintf(&b, "hactive=%d;hcount=%d;pactive=%d;pcount=%d;owned=%d",
		w.homunList.ActiveHomunID, len(w.homunList.List), w.petList.ActivePetID, len(w.petList.List), len(w.owned))
	for _, row := range w.homunList.List {
		fmt.Fprintf(&b, ";h=%d,%d,%d,%d,%d", row.HomunID, row.ClassID, row.Level, row.Intimacy, row.Hunger)
	}
	for _, row := range w.petList.List {
		fmt.Fprintf(&b, ";p=%d,%d,%d,%d,%d,%q", row.PetID, row.ClassID, row.Level, row.Flags, row.Intimate, row.Buff)
	}
	return b.String()
}

func companionInfoRow(label, value string) widget.Widget {
	return primitives.HBox(
		primitives.Box(rotheme.Text(label).Color(rotheme.Default.Colors.MutedText)).Width(float32(companionDetailLabelW)),
		rotheme.Text(value),
	).Gap(float32(companionRowGap)).CrossAlign(primitives.CrossAxisCenter).Height(float32(companionWindowRowH))
}

func companionEmptyText(text string) widget.Widget {
	return primitives.Box(rotheme.Text(text).Color(rotheme.Default.Colors.MutedText)).Padding(float32(companionEmptyPad))
}

func CompanionClassName(classID int) string {
	if name := db.MonsterResourceName[classID]; name != "" {
		return name
	}
	return fmt.Sprintf("#%d", classID)
}

func HomunculusLineage(classID, prevClass int) int {
	c := classID
	switch classID {
	case 6048, 6049, 6050, 6051, 6052:
		if prevClass > 0 {
			c = prevClass
		}
	}
	switch c {
	case 6001, 6005, 6009, 6013, 6048:
		return 0
	case 6002, 6006, 6010, 6014, 6049:
		return 1
	case 6003, 6007, 6011, 6015, 6051:
		return 2
	case 6004, 6008, 6012, 6016, 6050:
		return 3
	case 6052:
		return 4
	default:
		return -1
	}
}

func HomunculusLineageName(lineage int) string {
	switch lineage {
	case 0:
		return "Lif"
	case 1:
		return "Amistr"
	case 2:
		return "Filir"
	case 3:
		return "Vanilmirth"
	case 4:
		return "Eleanor"
	default:
		return "Homunculus"
	}
}

func StableHomunculusIntimacyText(value int) string {
	switch {
	case value >= 91000:
		return "Loyal"
	case value >= 75000:
		return "Cordial"
	case value >= 25000:
		return "Neutral"
	case value >= 10000:
		return "Shy"
	default:
		return "Awkward"
	}
}

func stablePetStatus(row network.StablePetEntry) string {
	if row.Flags&network.StableFlagInvEgg != 0 {
		return "Legacy egg on this character"
	}
	if row.Flags&network.StableFlagOut != 0 {
		if row.OwnerName != "" {
			return "Out · " + row.OwnerName
		}
		return "Out"
	}
	return "In account stable"
}

func companionPetBuffText(row network.StablePetEntry) string {
	text := strings.TrimSpace(row.Buff)
	if text == "" {
		return "No owner bonus"
	}
	return text
}

func companionPetExpText(row network.StablePetEntry) string {
	if row.NextExp == 0 {
		return "Max"
	}
	pct := 0
	if row.NextExp > 0 {
		pct = int((uint64(row.Exp) * 100) / uint64(row.NextExp))
		if pct > 100 {
			pct = 100
		}
	}
	return fmt.Sprintf("%d / %d (%d%%)", row.Exp, row.NextExp, pct)
}

func companionItemName(ctx Context, itemID uint32) string {
	if itemID == 0 {
		return "None"
	}
	if ctx.Resources != nil {
		if name, ok := ctx.Resources.ItemDisplayName(int(itemID), true); ok && strings.TrimSpace(name) != "" {
			return name
		}
	}
	return fmt.Sprintf("#%d", itemID)
}

func StableResultMessage(code int16) string {
	switch code {
	case network.StableOK:
		return ""
	case network.StableErrNoHomun:
		return "No homunculus in that slot."
	case network.StableErrType:
		return "You already have that homunculus type."
	case network.StableErrInUse:
		return "That pet is already out on another character."
	case network.StableErrInventory:
		return "Inventory is full."
	case network.StableErrNotYours:
		return "That companion does not belong here."
	case network.StableErrNoPet:
		return "No pet found."
	case network.StableErrHavePet:
		return "Return your current pet to its egg first, or Store it."
	case network.StableErrDead:
		return "Revive your homunculus first."
	case network.StableErrNotActive:
		return "Summon that pet first."
	case network.StableErrNoFood:
		return "No pet food in inventory or storage."
	case network.StableErrNoAcce:
		return "Could not equip or unequip that accessory."
	case network.StableErrEvolve:
		return "Evolution failed (loyalty, materials, or recipe)."
	default:
		return "Stable action failed."
	}
}

func yesNo(v bool) string {
	if v {
		return "Yes"
	}
	return "No"
}

func emptyDash(v string) string {
	if strings.TrimSpace(v) == "" {
		return "—"
	}
	return v
}

func ownedEntries(owned map[uint32]uint32) []network.StableOwnedEntry {
	entries := make([]network.StableOwnedEntry, 0, len(owned))
	for id, count := range owned {
		entries = append(entries, network.StableOwnedEntry{ItemID: id, Owned: count})
	}
	for i := 0; i < len(entries); i++ {
		for j := i + 1; j < len(entries); j++ {
			if entries[j].ItemID < entries[i].ItemID {
				entries[i], entries[j] = entries[j], entries[i]
			}
		}
	}
	return entries
}
