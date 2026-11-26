package game

import (
	"testing"

	"github.com/kivutar/goro/client"
	"github.com/kivutar/goro/network"
	"github.com/kivutar/goro/session"
)

func TestQuietStorageListDoesNotOpenStorage(t *testing.T) {
	sessionState := &session.Session{}
	ctx := client.Context{Session: sessionState}
	applyQuietStorageList(ctx, network.StorageList{
		Where:   network.StorageWherePersonal,
		Replace: true,
		Items: []network.InventoryItem{{
			Index: 4, ItemID: 501, Amount: 2, Type: 0, Identified: true,
		}},
	})
	if sessionState.Storage.Open || sessionState.GuildStorage.Open {
		t.Fatal("quiet storage list opened the storage interface")
	}
	if len(sessionState.Storage.Items) != 1 || sessionState.Storage.Items[0].ItemID != 501 || sessionState.Storage.Items[0].Amount != 2 {
		t.Fatalf("storage items = %+v", sessionState.Storage.Items)
	}
	applyQuietStorageList(ctx, network.StorageList{
		Where:   network.StorageWhereGuild,
		Replace: true,
		Items:   []network.InventoryItem{{Index: 2, ItemID: 909, Amount: 1, Identified: true}},
	})
	if sessionState.GuildStorage.Open || len(sessionState.Storage.Items) != 1 || len(sessionState.GuildStorage.Items) != 1 {
		t.Fatalf("guild list changed the open flag or the personal list: %+v %+v", sessionState.Storage, sessionState.GuildStorage)
	}
}
