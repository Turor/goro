package game

import (
	"testing"

	"github.com/kivutar/goro/client"
	"github.com/kivutar/goro/network"
	"github.com/kivutar/goro/session"
)

func TestGuildStorageListStaysSeparateFromKafra(t *testing.T) {
	s := &session.Session{
		Storage: session.Storage{
			Open:  true,
			Items: []session.InventoryItem{{Index: 1, ItemID: 501, Amount: 1}},
		},
	}
	ctx := client.Context{Session: s}
	applyGuildStorageItemList(ctx, []network.InventoryItem{{Index: 4, ItemID: 909, Amount: 1}})
	if s.Storage.Open || len(s.Storage.Items) != 1 || s.Storage.Items[0].ItemID != 501 {
		t.Fatalf("kafra storage changed: %+v", s.Storage)
	}
	if !s.GuildStorage.Open || len(s.GuildStorage.Items) != 1 || s.GuildStorage.Items[0].Index != 4 {
		t.Fatalf("guild storage = %+v", s.GuildStorage)
	}
	applyStorageItemAdded(ctx, network.InventoryItem{Index: 5, ItemID: 910, Amount: 2})
	if len(s.GuildStorage.Items) != 2 || len(s.Storage.Items) != 1 {
		t.Fatalf("add landed in the wrong storage: kafra=%d guild=%d", len(s.Storage.Items), len(s.GuildStorage.Items))
	}
	applyStorageItemRemoved(ctx, network.StorageItemRemoved{Index: 4, Amount: 1})
	if len(s.GuildStorage.Items) != 1 || s.GuildStorage.Items[0].Index != 5 {
		t.Fatalf("remove hit the wrong storage: %+v", s.GuildStorage.Items)
	}
}
