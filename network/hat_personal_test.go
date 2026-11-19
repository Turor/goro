package network

import (
	"encoding/binary"
	"testing"
)

func TestParseHatEffectAndPersonalInfo(t *testing.T) {
	hat := make([]byte, 13)
	binary.LittleEndian.PutUint16(hat[0:2], PacketZCHatEffect)
	binary.LittleEndian.PutUint16(hat[2:4], 13)
	binary.LittleEndian.PutUint32(hat[4:8], 150001)
	hat[8] = 1
	binary.LittleEndian.PutUint16(hat[9:11], 12)
	binary.LittleEndian.PutUint16(hat[11:13], 34)
	got, ok, err := ParseHatEffect(Packet{ID: PacketZCHatEffect, Data: hat})
	if err != nil || !ok || got.GID != 150001 || len(got.IDs) != 2 || got.IDs[1] != 34 {
		t.Fatalf("%+v ok=%t err=%v", got, ok, err)
	}
	info := make([]byte, 16)
	binary.LittleEndian.PutUint16(info[0:2], PacketZCPersonalInformation)
	binary.LittleEndian.PutUint16(info[2:4], 16)
	binary.LittleEndian.PutUint32(info[4:8], 100)
	binary.LittleEndian.PutUint32(info[8:12], 5)
	binary.LittleEndian.PutUint32(info[12:16], 10)
	pi, ok, err := ParsePersonalInformation(Packet{ID: PacketZCPersonalInformation, Data: info})
	if err != nil || !ok || pi.Exp != 100 || pi.Drop != 10 {
		t.Fatalf("%+v", pi)
	}
}
