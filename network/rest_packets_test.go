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
