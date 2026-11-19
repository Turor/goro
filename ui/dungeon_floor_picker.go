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

// DungeonFloorPicker is the world-map overlay for a dungeon tile click.
type DungeonFloorPicker struct {
	Window
	info     session.WarpraDungeon
	price    uint32
	zeny     uint32
	scrollY  state.Signal[float32]
	onSelect func(network.WarpraLoc)
}

func (p *DungeonFloorPicker) OpenInfo(ctx Context, info session.WarpraDungeon, price, zeny uint32, onSelect func(network.WarpraLoc)) {
	if len(info.Floors) == 0 {
		return
	}
	p.info = info
	p.price = price
	p.zeny = zeny
	p.onSelect = onSelect
	if p.scrollY == nil {
		p.scrollY = state.NewSignal[float32](0)
	}
	p.scrollY.Set(0)
	height := p.windowHeight()
	p.EnsureWindow(dungeonPickerW, height)
	p.SetSize(dungeonPickerW, height)
	p.Window.Open(ctx, p.widgetTree(ctx))
	p.Publish(ctx)
}

func (p *DungeonFloorPicker) Refresh(ctx Context, info session.WarpraDungeon, price, zeny uint32) {
	if !p.IsOpen() {
		return
	}
	p.info = info
	p.price = price
	p.zeny = zeny
	p.SetSize(dungeonPickerW, p.windowHeight())
	p.SetContent(p.widgetTree(ctx))
	p.Publish(ctx)
}

func (p *DungeonFloorPicker) Update(ctx Context) bool {
	if !p.IsOpen() {
		return false
	}
	consumed := p.Window.Update(ctx)
	p.Publish(ctx)
	return consumed
}

func (p *DungeonFloorPicker) windowHeight() int {
	rows := minInt(len(p.info.Floors), dungeonPickerMax)
	return ROWindowTitleHeight + dungeonPickerPad*2 + dungeonPickerIntroH + rows*dungeonPickerRowH + ROWindowFooterHeight
}

func (p *DungeonFloorPicker) widgetTree(ctx Context) widget.Widget {
	intro := "Choose a floor to warp to."
	if !p.info.Unlocked {
		intro = "Dungeon locked. Unlock for " + formatWarpraZeny(p.price) + " zeny (you have " + formatWarpraZeny(p.zeny) + "), then warp to any floor."
	}
	floors := append([]network.WarpraLoc(nil), p.info.Floors...)
	lv := listview.New(
		listview.ItemCount(len(floors)),
		listview.FixedItemHeight(float32(dungeonPickerRowH)),
		listview.ScrollYSignal(p.scrollY),
		listview.SelectionModeOpt(listview.SelectionNone),
		listview.PainterOpt(rotheme.SelectListPainter{EmptyText: "No floors listed."}),
		listview.BuildItem(func(item listview.ItemContext) widget.Widget {
			if item.Index < 0 || item.Index >= len(floors) {
				return rotheme.SelectListRow("", false, float32(dungeonPickerRowH))
			}
			floor := floors[item.Index]
			label := warpraFloorLabel(floor, p.price)
			return rotheme.SelectListRow(trimRunes(label, 48), true, float32(dungeonPickerRowH))
		}),
		listview.OnItemClick(func(index int) {
			if index < 0 || index >= len(floors) || p.onSelect == nil {
				return
			}
			p.onSelect(floors[index])
			if floors[index].Unlocked {
				p.Close()
			}
		}),
	)
	title := p.info.Name
	if title == "" {
		title = "Dungeon Floors"
	}
	return Win(Title(title), CloseButton(true), OnClose(p.Close),
		Size(float32(dungeonPickerW), float32(p.windowHeight())),
		Content(
			primitives.Box(
				rotheme.Text(intro).MaxLines(3),
				primitives.Box(lv).Height(float32(minInt(len(floors), dungeonPickerMax)*dungeonPickerRowH)).CrossAlign(primitives.CrossAxisStretch),
			).Padding(float32(dungeonPickerPad)).Gap(float32(dungeonPickerGap)),
		),
		Footer(primitives.Expanded(primitives.Box()), rotheme.Button("Close", p.Close)),
	)
}

func warpraFloorLabel(floor network.WarpraLoc, price uint32) string {
	mark := "✕"
	action := "Unlock " + formatWarpraZeny(price)
	if floor.Unlocked {
		mark = "✓"
		action = "Warp"
	}
	name := floor.Name
	if name == "" {
		name = floor.Map
	}
	return mark + "  " + name + "  " + floor.Map + "  " + action
}
