package ui

import (
	"github.com/kivutar/goro/glog"
	"github.com/kivutar/goro/network"
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
	where, ok := quietStorageWhere(source)
	if !ok {
		return
	}
	if err := ctx.Network.SendStorageMove(where, network.StorageActionList, 0, 1); err != nil {
		glog.Warnf("storage list failed: %v", err)
	}
}

func quietStorageWhere(source int) (byte, bool) {
	switch source {
	case opinionatedSourceStorage:
		return network.StorageWherePersonal, true
	case opinionatedSourceGuild:
		return network.StorageWhereGuild, true
	default:
		return 0, false
	}
}

func (w *InventoryBagWindow) sendQuietStorageMove(ctx Context, item session.InventoryItem, from, to int, amount uint16) bool {
	if w == nil || !w.storageSource() || ctx.Network == nil {
		return false
	}
	deposit := from == opinionatedPaneBag && to == opinionatedPaneOther
	withdraw := from == opinionatedPaneOther && to == opinionatedPaneBag
	if !deposit && !withdraw {
		return false
	}
	where, ok := quietStorageWhere(w.source)
	if !ok {
		return false
	}
	action := network.StorageActionDeposit
	if withdraw {
		action = network.StorageActionWithdraw
	}
	if err := ctx.Network.SendStorageMove(where, action, item.Index, uint32(amount)); err != nil {
		glog.Warnf("storage move failed: %v", err)
	}
	return true
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
