package ui

import (
	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/widget"
	"github.com/kivutar/goro/client"
	"github.com/kivutar/goro/ui/rotheme"
)

type BasicMenu struct {
	Window
	content   widget.Widget
	callbacks BasicMenuCallbacks
	collapsed bool
}

type BasicMenuCallbacks struct {
	OnStatus      func()
	OnOption      func()
	OnItems       func()
	OnEquip       func()
	OnSkill       func()
	OnMap         func()
	OnComm        func()
	OnFriend      func()
	OnWarpra      func()
	OnBank        func()
	OnMail        func()
	OnAttendance  func()
	OnAchievement func()
	OnSearch      func()
	OnSwitch      func()
}

type basicMenuButton struct {
	key   string
	label string
}

var basicMenuButtons = []basicMenuButton{
	{key: "status", label: "Status"},
	{key: "option", label: "Option"},
	{key: "items", label: "Items"},
	{key: "equip", label: "Equip"},
	{key: "skill", label: "Skill"},
	{key: "map", label: "Map"},
	{key: "comm", label: "Comm"},
	{key: "friend", label: "Friend"},
	{key: "warpra", label: "Warpra"},
	{key: "bank", label: "Bank"},
	{key: "mail", label: "Mail"},
	{key: "attend", label: "Attend"},
	{key: "achieve", label: "Achieve"},
	{key: "search", label: "Search"},
	{key: "switch", label: "Switch"},
}

func (m *BasicMenu) Update(ctx client.Context, callbacks BasicMenuCallbacks) bool {
	m.callbacks = callbacks
	width, height := basicMenuSize(m.collapsed)
	if m.EnsureWindow(width, height) {
		m.titleHeight = 0
		m.CloseOnEsc = false
	}
	if m.width != width || m.baseHeight() != height {
		m.SetSize(width, height)
		m.content = nil
	}
	if !m.IsOpen() {
		m.OpenAt(basicMenuX, basicMenuY(), m.widgetTree())
	} else if m.content == nil {
		m.SetContent(m.widgetTree())
	}
	m.bindLayoutReload()
	consumed := m.Window.Update(ctx)
	m.Publish(ctx)
	return consumed
}

func (m *BasicMenu) Rebind(ctx client.Context, callbacks BasicMenuCallbacks) {
	m.callbacks = callbacks
	width, height := basicMenuSize(m.collapsed)
	if m.EnsureWindow(width, height) {
		m.titleHeight = 0
		m.CloseOnEsc = false
	}
	m.SetSize(width, height)
	m.content = nil
	if !m.IsOpen() {
		return
	}
	m.SetContent(m.widgetTree())
	m.Publish(ctx)
	m.OnLayoutReload(func() { m.rebuildFromLayout() })
}

func (m *BasicMenu) bindLayoutReload() {
	if m.layoutRebuild != nil {
		return
	}
	m.OnLayoutReload(func() { m.rebuildFromLayout() })
}

func (m *BasicMenu) rebuildFromLayout() {
	width, height := basicMenuSize(m.collapsed)
	m.SetSize(width, height)
	m.content = nil
	if !m.IsOpen() {
		return
	}
	m.SetContent(m.widgetTree())
}

// FollowCharacterWindow keeps the menu attached below the character window
// while leaving both as independent overlays for input and redraw purposes.
func (m *BasicMenu) FollowCharacterWindow(ctx client.Context, character *CharacterWindow) {
	if character == nil || !character.IsOpen() {
		return
	}
	width, height := basicMenuSize(m.collapsed)
	if m.EnsureWindow(width, height) {
		m.titleHeight = 0
		m.CloseOnEsc = false
	}
	// The menu owns the attached extent, including when only its toggle is
	// visible. Expanding near the bottom moves the whole group back on screen.
	bottom := basicMenuFollowGap + height
	if bottom > character.dragBottom && !character.dragging {
		_, screenH := ctx.ScreenSize()
		maxY := maxInt(windowScreenMargin, screenH-character.height-bottom-windowScreenMargin)
		if character.y > maxY {
			character.setPosition(ctx, character.x, maxY)
		}
	}
	character.dragBottom = bottom
	x := character.x
	y := character.y + character.height + basicMenuFollowGap
	if character.dragLayer {
		m.followDragPosition(ctx, x, y)
		return
	}
	m.endFollowDrag(ctx)
	if m.positioned && m.x == x && m.y == y {
		return
	}
	m.ctx = ctx
	m.positioned = true
	m.setPosition(ctx, x, y)
	if m.IsOpen() {
		m.Publish(ctx)
	}
}

func (m *BasicMenu) followDragPosition(ctx client.Context, x, y int) {
	m.ctx = ctx
	m.positioned = true
	m.x, m.y = x, y
	if overlay := m.positionedOverlay(); overlay != nil {
		overlay.setFrameQuiet(x, y, m.width, m.height)
		if !overlay.hidden {
			overlay.hidden = true
			damage := overlay.markFrameDirty()
			invalidateWindowRect(ctx, damage)
		}
		return
	}
	m.placed = nil
}

func (m *BasicMenu) endFollowDrag(ctx client.Context) {
	overlay := m.positionedOverlay()
	if overlay == nil || !overlay.hidden {
		return
	}
	overlay.hidden = false
	damage := overlay.markFrameDirty()
	invalidateWindowRect(ctx, damage)
}

func basicMenuY() int {
	return characterWindowY + characterWindowHeight + basicMenuFollowGap
}

func basicMenuBounds() (int, int, int, int) {
	w, h := basicMenuSize(false)
	return basicMenuX, basicMenuY(), w, h
}

func basicMenuSize(collapsed bool) (int, int) {
	w, h := characterWindowWidth, basicMenuToggleH
	if !collapsed {
		rows := basicMenuRowCount()
		h += basicMenuToggleGap + basicMenuPad*2 + rows*basicMenuButtonH + (rows-1)*basicMenuGapY
	}
	return w, h
}

func basicMenuRowCount() int {
	need := (len(basicMenuButtons) + basicMenuCols - 1) / basicMenuCols
	if basicMenuRows > need {
		return basicMenuRows
	}
	return need
}

func (m *BasicMenu) toggleCollapsed() {
	m.SetCollapsed(!m.collapsed)
}

func (m *BasicMenu) Collapsed() bool { return m.collapsed }

func (m *BasicMenu) SetCollapsed(collapsed bool) {
	if m.collapsed == collapsed {
		return
	}
	m.collapsed = collapsed
	width, height := basicMenuSize(m.collapsed)
	if m.width != 0 {
		m.SetSize(width, height)
	}
	m.content = nil
	if !m.IsOpen() {
		return
	}
	m.SetContent(m.widgetTree())
	m.Publish(m.ctx)
}

func (m *BasicMenu) widgetTree() widget.Widget {
	if m.content != nil {
		return m.content
	}
	width, height := basicMenuSize(m.collapsed)
	kind := rotheme.IconButtonCollapse
	if m.collapsed {
		kind = rotheme.IconButtonExpand
	}
	toggle := rotheme.IconButton(kind, m.toggleCollapsed).
		Width(float32(width)).Height(float32(basicMenuToggleH)).
		CrossAlign(primitives.CrossAxisStretch)
	if m.collapsed {
		m.content = toggle
		return m.content
	}
	rowCount := basicMenuRowCount()
	rows := make([]widget.Widget, 0, rowCount)
	for row := 0; row < rowCount; row++ {
		buttons := make([]widget.Widget, 0, basicMenuCols)
		for col := 0; col < basicMenuCols; col++ {
			index := row*basicMenuCols + col
			if index >= len(basicMenuButtons) {
				buttons = append(buttons, primitives.Box().Width(float32(basicMenuButtonW)).Height(float32(basicMenuButtonH)))
				continue
			}
			button := basicMenuButtons[index]
			key := button.key
			label := button.label
			buttons = append(buttons,
				rotheme.Button(label, func() {
					m.invoke(key)
				}).
					Width(float32(basicMenuButtonW)).
					Height(float32(basicMenuButtonH)),
			)
		}
		if len(buttons) == 0 {
			continue
		}
		rows = append(rows,
			primitives.HBox(buttons...).
				Gap(float32(basicMenuGapX)).
				CrossAlign(primitives.CrossAxisStretch),
		)
	}
	panel := Win(
		TitleBar(false),
		Size(float32(width), float32(height-basicMenuToggleH-basicMenuToggleGap)),
		Content(
			primitives.Box(rows...).
				Padding(float32(basicMenuPad)).
				Gap(float32(basicMenuGapY)).
				CrossAlign(primitives.CrossAxisStretch),
		),
	)
	m.content = primitives.Box(toggle, panel).
		Width(float32(width)).Height(float32(height)).
		Gap(float32(basicMenuToggleGap)).CrossAlign(primitives.CrossAxisStretch)
	return m.content
}

func (m *BasicMenu) invoke(key string) {
	var callback func()
	switch key {
	case "status":
		callback = m.callbacks.OnStatus
	case "option":
		callback = m.callbacks.OnOption
	case "items":
		callback = m.callbacks.OnItems
	case "equip":
		callback = m.callbacks.OnEquip
	case "skill":
		callback = m.callbacks.OnSkill
	case "map":
		callback = m.callbacks.OnMap
	case "comm":
		callback = m.callbacks.OnComm
	case "friend":
		callback = m.callbacks.OnFriend
	case "warpra":
		callback = m.callbacks.OnWarpra
	case "bank":
		callback = m.callbacks.OnBank
	case "mail":
		callback = m.callbacks.OnMail
	case "attend":
		callback = m.callbacks.OnAttendance
	case "achieve":
		callback = m.callbacks.OnAchievement
	case "search":
		callback = m.callbacks.OnSearch
	case "switch":
		callback = m.callbacks.OnSwitch
	}
	if callback != nil {
		callback()
	}
}
