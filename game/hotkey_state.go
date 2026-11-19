package game

import (
	"github.com/kivutar/goro/client"
	"github.com/kivutar/goro/glog"
	"github.com/kivutar/goro/network"
	"github.com/kivutar/goro/session"
)

func applyHotkeyList(ctx client.Context, list network.HotkeyList) {
	if ctx.Session == nil {
		return
	}
	slots := make([]session.HotkeySlot, len(list.Slots))
	for i, slot := range list.Slots {
		slots[i] = session.HotkeySlot{
			Type:  slot.Type,
			ID:    slot.ID,
			Level: slot.Level,
		}
	}
	if list.Tab != 0 {
		if ctx.Session.Hotkeys.Extra == nil {
			ctx.Session.Hotkeys.Extra = map[uint16][]session.HotkeySlot{}
		}
		ctx.Session.Hotkeys.Extra[list.Tab] = slots
		glog.Debugf("hotkey extra tab=%d slots=%d", list.Tab, len(slots))
		return
	}
	ctx.Session.Hotkeys.Slots = slots
	ctx.Session.Hotkeys.Loaded = true
	ctx.Session.Hotkeys.Version++
	glog.Debugf("hotkey list received slots=%d", len(slots))
}
