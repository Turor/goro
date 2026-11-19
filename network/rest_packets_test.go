package network

import (
	"encoding/binary"
	"testing"
)

func TestConfigAndSearchAndSwitchPackets(t *testing.T) {
	cfg := BuildConfigPacket(ConfigPetFeed, 1)
	if len(cfg) != 10 || ID(cfg) != PacketCZConfig {
		t.Fatalf("config")
	}
	sw := BuildEquipSwitchAllPacket()
	if len(sw) != 2 || ID(sw) != PacketCZEquipSwitchAll {
		t.Fatalf("switch")
	}
	add := BuildEquipSwitchAddPacket(5, 0x10)
	if binary.LittleEndian.Uint16(add[2:4]) != 5 || binary.LittleEndian.Uint32(add[4:8]) != 0x10 {
		t.Fatalf("add")
	}
	search := BuildSearchStorePacket(1201)
	if ID(search) != PacketCZSearchStore || binary.LittleEndian.Uint16(search[2:4]) != uint16(len(search)) {
		t.Fatalf("search")
	}
	attach := BuildRodexAddItemPacket(3, 1)
	if len(attach) != 6 || binary.LittleEndian.Uint16(attach[2:4]) != 3 {
		t.Fatalf("attach")
	}
}

func TestParseSearchStoreAck2025(t *testing.T) {
	const entry = 157
	data := make([]byte, 7+entry)
	binary.LittleEndian.PutUint16(data[0:2], PacketZCSearchStoreAck)
	binary.LittleEndian.PutUint16(data[2:4], uint16(len(data)))
	data[4], data[5], data[6] = 1, 0, 3
	binary.LittleEndian.PutUint32(data[7:11], 11)
	binary.LittleEndian.PutUint32(data[11:15], 22)
	copy(data[15:95], []byte("Shop"))
	binary.LittleEndian.PutUint32(data[95:99], 1201)
	data[99] = 0
	binary.LittleEndian.PutUint32(data[100:104], 500)
	binary.LittleEndian.PutUint16(data[104:106], 2)
	ack, ok, err := ParseSearchStoreAck(Packet{ID: PacketZCSearchStoreAck, Data: data})
	if err != nil || !ok || len(ack.Hits) != 1 || ack.Hits[0].ItemID != 1201 || ack.Hits[0].Price != 500 {
		t.Fatalf("%+v ok=%t err=%v", ack, ok, err)
	}
}
