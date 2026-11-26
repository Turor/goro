package network

import "testing"

func TestStorageMovePacketRoundTrip(t *testing.T) {
	raw := BuildStorageMovePacket(StorageWherePersonal, StorageActionWithdraw, 4, 1)
	if len(raw) != 10 || raw[0] != 0xE2 || raw[1] != 0x0E || raw[2] != StorageWherePersonal || raw[3] != StorageActionWithdraw {
		t.Fatalf("move packet %x", raw)
	}
	result := Packet{ID: PacketZCStorageResult, Data: []byte{
		0xE3, 0x0E, StorageResultOK, StorageWhereGuild, StorageActionDeposit, 5,
		1, 0, 2, 0, 3, 0, 4, 0, 0, 0, 0, 0,
	}}
	got, ok, err := ParseStorageMoveResult(result)
	if err != nil || !ok || got.Result != StorageResultOK || got.Where != StorageWhereGuild || got.StorageIndex != 1 || got.Amount != 2 || got.ItemID != 4 {
		t.Fatalf("result %+v ok=%t err=%v", got, ok, err)
	}
	entry := make([]byte, storageEntrySize)
	entry[0], entry[1] = 3, 0
	entry[2], entry[3] = 0xF5, 0x01
	entry[6], entry[7] = 9, 0
	entry[8] = 4
	entry[9] = 1
	body := make([]byte, storageListHeader+storageEntrySize)
	body[0], body[1] = 0xE4, 0x0E
	body[4] = StorageWherePersonal
	body[5] = storageListReplace
	body[8] = 1
	body[10] = 1
	copy(body[storageListHeader:], entry)
	list, ok, err := ParseStorageList(Packet{ID: PacketZCStorageList, Data: body})
	if err != nil || !ok || !list.Replace || len(list.Items) != 1 || list.Items[0].Index != 3 || list.Items[0].ItemID != 0x1F5 || list.Items[0].Amount != 9 {
		t.Fatalf("list %+v ok=%t err=%v", list, ok, err)
	}
}
