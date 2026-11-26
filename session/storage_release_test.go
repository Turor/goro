package session

import "testing"

func TestReleaseOpenStorageKeepsTheCachedList(t *testing.T) {
	s := &Session{
		Storage:      Storage{Open: true, Amount: 1, MaxAmount: 300, Items: []InventoryItem{{Index: 4, ItemID: 501, Amount: 2}}},
		GuildStorage: Storage{Open: true, Items: []InventoryItem{{Index: 9, ItemID: 909, Amount: 1}}},
	}
	ReleaseOpenStorage(s)
	if s.Storage.Open || s.GuildStorage.Open {
		t.Fatal("server storage stayed open")
	}
	if len(s.Storage.Items) != 1 || s.Storage.Items[0].ItemID != 501 || len(s.GuildStorage.Items) != 1 {
		t.Fatalf("cached lists were dropped: %+v %+v", s.Storage, s.GuildStorage)
	}
}
