package ui

import (
	"github.com/kivutar/goro/client"
	"github.com/kivutar/goro/glog"
	"github.com/kivutar/goro/session"
)

func (w *InventoryBagWindow) storageSource() bool {
	return w != nil && (w.source == opinionatedSourceStorage || w.source == opinionatedSourceGuild)
}

func (w *InventoryBagWindow) queueStore(item session.InventoryItem, amount uint16) {
	if w == nil {
		return
	}
	w.pendingStore = item
	w.pendingStoreAmount = uint32(amount)
	if w.pendingStoreAmount == 0 {
		w.pendingStoreAmount = opinionatedMoveAmount(item)
	}
	w.pendingStoreSource = w.source
}

func (w *InventoryBagWindow) queueWithdraw(item session.InventoryItem, amount uint16) {
	if w == nil {
		return
	}
	w.pendingWithdraw = item
	w.pendingWithdrawAmount = uint32(amount)
	if w.pendingWithdrawAmount == 0 {
		w.pendingWithdrawAmount = opinionatedMoveAmount(item)
	}
	w.pendingWithdrawSource = w.source
}

func (w *InventoryBagWindow) flushPendingStore(ctx Context) {
	w.flushPendingWithdraw(ctx)
	if w == nil || w.pendingStoreAmount == 0 || w.source != w.pendingStoreSource {
		return
	}
	if !opinionatedMoveAllowed(ctx, w.source) {
		return
	}
	item := w.pendingStore
	amount := w.pendingStoreAmount
	w.pendingStore = session.InventoryItem{}
	w.pendingStoreAmount = 0
	opinionatedDepositAmount(ctx, w.source, item, amount)
}

func (w *InventoryBagWindow) flushPendingWithdraw(ctx Context) {
	if w == nil || w.pendingWithdrawAmount == 0 || w.source != w.pendingWithdrawSource {
		return
	}
	if !opinionatedMoveAllowed(ctx, w.source) {
		return
	}
	item := w.pendingWithdraw
	amount := w.pendingWithdrawAmount
	w.pendingWithdraw = session.InventoryItem{}
	w.pendingWithdrawAmount = 0
	opinionatedWithdrawAmount(ctx, w.source, item, amount)
}

func (w *InventoryBagWindow) requestSourceOpen(ctx Context, source int) {
	if w == nil || ctx.Network == nil {
		return
	}
	w.storageReleaseSent = false
	switch source {
	case opinionatedSourceStorage:
		if ctx.Session != nil && ctx.Session.Storage.Open && !ctx.Session.GuildStorage.Open {
			return
		}
		if ctx.Session != nil && ctx.Session.GuildStorage.Open {
			if err := ctx.Network.SendCloseStorage(); err != nil {
				glog.Warnf("close guild storage before kafra failed: %v", err)
			}
		}
		if err := client.SendChat(ctx, "@storage"); err != nil {
			glog.Warnf("open storage failed: %v", err)
		}
	case opinionatedSourceGuild:
		if ctx.Session != nil && ctx.Session.GuildStorage.Open {
			return
		}
		if ctx.Session != nil && ctx.Session.Storage.Open {
			if err := ctx.Network.SendCloseStorage(); err != nil {
				glog.Warnf("close kafra before guild storage failed: %v", err)
			}
		}
		if err := client.SendChat(ctx, "@gstorage"); err != nil {
			glog.Warnf("open guild storage failed: %v", err)
		}
	}
}

func (w *InventoryBagWindow) releaseServerStorage(ctx Context) {
	if w == nil || !w.HidesClassicStorage() || w.pendingStoreAmount != 0 || w.pendingWithdrawAmount != 0 || ctx.Network == nil || ctx.Session == nil {
		return
	}
	if !ctx.Session.Storage.Open && !ctx.Session.GuildStorage.Open {
		w.storageReleaseSent = false
		return
	}
	if w.storageReleaseSent {
		return
	}
	if err := ctx.Network.SendCloseStorage(); err != nil {
		glog.Warnf("release storage failed: %v", err)
		return
	}
	w.storageReleaseSent = true
}
