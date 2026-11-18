package game

import (
	"github.com/kivutar/goro/client"
	"github.com/kivutar/goro/glog"
	"github.com/kivutar/goro/network"
	"github.com/kivutar/goro/session"
)

func (m *WorldMode) bindWarpra(ctx client.Context) {
	request := func() { m.requestWarpraList(ctx, false) }
	needList := func() { m.requestWarpraList(ctx, true) }
	warp := func(loc network.WarpraLoc) { m.sendWarpraWarp(ctx, loc) }
	buy := func(loc network.WarpraLoc) { m.sendWarpraBuy(ctx, loc) }
	m.ui.worldMap.BindWarpra(request, needList, warp, buy)
	m.ui.warpra.Rebind(ctx, request, warp, buy)
}

func (m *WorldMode) handleWarpraPacket(ctx client.Context, pkt network.Packet) bool {
	switch pkt.ID {
	case network.PacketZCWarpraList:
		list, ok, err := network.ParseWarpraList(pkt)
		if err != nil {
			glog.Errorf("parse warpra list 0x%04X: %v", pkt.ID, err)
			return true
		}
		if !ok {
			return false
		}
		if ctx.Session == nil {
			return true
		}
		complete := ctx.Session.Warpra.ApplyList(list)
		if complete {
			m.maybeWarpAfterBuy(ctx)
		}
		m.syncWarpraWindows(ctx)
		return true
	case network.PacketZCWarpraResult:
		result, ok, err := network.ParseWarpraResult(pkt)
		if err != nil {
			glog.Errorf("parse warpra result 0x%04X: %v", pkt.ID, err)
			return true
		}
		if !ok {
			return false
		}
		m.applyWarpraResult(ctx, result)
		return true
	default:
		return false
	}
}

func (m *WorldMode) syncWarpraWindows(ctx client.Context) {
	if ctx.Session == nil {
		return
	}
	catalog := ctx.Session.Warpra
	m.ui.worldMap.SetCatalog(ctx, catalog)
	m.ui.warpra.SetCatalog(ctx, catalog)
}

func (m *WorldMode) applyWarpraResult(ctx client.Context, result network.WarpraResult) {
	loc, ok := network.WarpraLoc{}, false
	if ctx.Session != nil {
		loc, ok = ctx.Session.Warpra.Find(uint8(result.Kind), result.ID)
	}
	name := "that location"
	if ok {
		name = loc.Name
		if name == "" {
			name = loc.Map
		}
	}
	price := network.WarpraDefaultPrice
	if ctx.Session != nil {
		price = ctx.Session.Warpra.UnlockPrice()
	}
	switch result.Result {
	case network.WarpraOK:
		pending, warpAfter, pendingOK := m.pendingWarpraBuy()
		if pendingOK && pending.Kind == uint8(result.Kind) && pending.ID == result.ID {
			m.ui.console.AddSystemMessage("Unlocked %s.", name)
			m.clearWarpraPending()
			m.ui.worldMap.Confirm.Close(ctx)
			m.ui.warpra.Confirm.Close(ctx)
			if warpAfter {
				m.sendWarpraWarp(ctx, pending)
			}
			return
		}
		m.ui.worldMap.CloseAfterWarp()
		m.ui.warpra.Close()
		m.clearWarpraPending()
	case network.WarpraLocked:
		m.ui.console.AddErrorMessage("%s is locked.", name)
		if ok {
			m.ui.worldMap.UnlockOrWarp(ctx, loc, true)
		}
	case network.WarpraNoZeny:
		m.ui.console.AddErrorMessage("Need %s zeny to unlock %s.", formatWarpraConsoleZeny(price), name)
	case network.WarpraBusy:
		m.ui.console.AddErrorMessage("You cannot warp right now.")
	case network.WarpraLevel:
		m.clearWarpraPending()
		m.ui.worldMap.Confirm.Close(ctx)
		m.ui.warpra.Confirm.Close(ctx)
		if ok {
			if msg := session.WarpraLevelGateMessage(loc); msg != "" {
				m.ui.console.AddErrorMessage("%s", msg)
			}
		}
	case network.WarpraInvalid, network.WarpraErr:
		m.ui.console.AddErrorMessage("Warp failed.")
	}
}

func (m *WorldMode) pendingWarpraBuy() (network.WarpraLoc, bool, bool) {
	if loc, warpAfter, ok := m.ui.worldMap.ConsumePendingBuy(); ok {
		return loc, warpAfter, true
	}
	if loc, warpAfter, ok := m.ui.warpra.ConsumePendingBuy(); ok {
		return loc, warpAfter, true
	}
	return network.WarpraLoc{}, false, false
}

func (m *WorldMode) clearWarpraPending() {
	m.ui.worldMap.ClearPending()
	m.ui.warpra.ClearPending()
}

func (m *WorldMode) maybeWarpAfterBuy(ctx client.Context) {
	loc, warpAfter, ok := m.pendingWarpraBuy()
	if !ok || !warpAfter || ctx.Session == nil {
		return
	}
	next, found := ctx.Session.Warpra.Find(loc.Kind, loc.ID)
	if !found || !next.Unlocked {
		return
	}
	m.clearWarpraPending()
	m.sendWarpraWarp(ctx, next)
}

func (m *WorldMode) requestWarpraList(ctx client.Context, announce bool) {
	if ctx.Session != nil {
		ctx.Session.Warpra.BeginList()
	}
	if announce && (ctx.Session == nil || !ctx.Session.Warpra.Ready()) {
		m.ui.console.AddSystemMessage("Loading Warpra destinations…")
	}
	m.sendWarpraRequest(ctx, network.WarpraActList, 0, 0)
}

func (m *WorldMode) sendWarpraWarp(ctx client.Context, loc network.WarpraLoc) {
	if !loc.Unlocked {
		if lv := session.WarpraLevelGate(loc); lv > 0 {
			if msg := session.WarpraLevelGateMessage(loc); msg != "" {
				m.ui.console.AddErrorMessage("%s", msg)
			}
			return
		}
	}
	m.sendWarpraRequest(ctx, network.WarpraActWarp, uint16(loc.Kind), loc.ID)
}

func (m *WorldMode) sendWarpraBuy(ctx client.Context, loc network.WarpraLoc) {
	if session.WarpraLevelGate(loc) > 0 {
		if msg := session.WarpraLevelGateMessage(loc); msg != "" {
			m.ui.console.AddErrorMessage("%s", msg)
		}
		return
	}
	m.sendWarpraRequest(ctx, network.WarpraActBuy, uint16(loc.Kind), loc.ID)
}

func (m *WorldMode) sendWarpraRequest(ctx client.Context, action, kind, id uint16) {
	if ctx.Network == nil {
		m.ui.console.AddErrorMessage("Warpra request failed: not connected.")
		return
	}
	if err := ctx.Network.SendWarpraRequest(action, kind, id); err != nil {
		m.ui.console.AddErrorMessage("Warpra request failed.")
		glog.Warnf("warpra request failed action=%d kind=%d id=%d: %v", action, kind, id, err)
	}
}

func formatWarpraConsoleZeny(n uint32) string {
	return formatWarpraZenyDigits(n)
}

func formatWarpraZenyDigits(n uint32) string {
	// Keep the helper in game so packet tests do not import ui.
	s := itoaU32(n)
	if len(s) <= 3 {
		return s
	}
	buf := make([]byte, 0, len(s)+len(s)/3)
	lead := len(s) % 3
	if lead == 0 {
		lead = 3
	}
	buf = append(buf, s[:lead]...)
	for i := lead; i < len(s); i += 3 {
		buf = append(buf, ',')
		buf = append(buf, s[i:i+3]...)
	}
	return string(buf)
}

func itoaU32(n uint32) string {
	if n == 0 {
		return "0"
	}
	var buf [10]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}
