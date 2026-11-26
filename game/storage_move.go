package game

import (
	"github.com/kivutar/goro/client"
	"github.com/kivutar/goro/glog"
	"github.com/kivutar/goro/network"
	"github.com/kivutar/goro/session"
)

func (m *WorldMode) handleQuietStorage(ctx client.Context, pkt network.Packet) bool {
	if result, ok, err := network.ParseStorageMoveResult(pkt); err != nil {
		glog.Errorf("parse storage move result 0x%04X: %v", pkt.ID, err)
		return true
	} else if ok {
		if result.Result != network.StorageResultOK {
			glog.Warnf("storage move result=%d where=%d action=%d", result.Result, result.Where, result.Action)
		}
		return true
	}
	if list, ok, err := network.ParseStorageList(pkt); err != nil {
		glog.Errorf("parse storage list 0x%04X: %v", pkt.ID, err)
		return true
	} else if ok {
		applyQuietStorageList(ctx, list)
		return true
	}
	return false
}

func applyQuietStorageList(ctx client.Context, list network.StorageList) {
	if ctx.Session == nil {
		return
	}
	items := &ctx.Session.Storage.Items
	if list.Where == network.StorageWhereGuild {
		items = &ctx.Session.GuildStorage.Items
	}
	if list.Replace {
		*items = nil
	}
	for _, item := range list.Items {
		addOrReplaceStorageList(items, sessionItemFromNetwork(item))
	}
}

func quietStorageItems(s *session.Session, where byte) []session.InventoryItem {
	if s == nil {
		return nil
	}
	if where == network.StorageWhereGuild {
		return s.GuildStorage.Items
	}
	return s.Storage.Items
}
