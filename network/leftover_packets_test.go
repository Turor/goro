package network

import (
	"encoding/binary"
	"testing"
)

func TestReformLaphineEdenPackets(t *testing.T) {
	ref := BuildReformItemPacket(1234, 7)
	if len(ref) != 8 || ID(ref) != PacketCZReformItem || binary.LittleEndian.Uint16(ref[6:8]) != 7 {
		t.Fatalf("reform %x", ref)
	}
	makePkt := BuildLaphineMakePacket(501, 3, 1)
	if ID(makePkt) != PacketCZLaphineMake || binary.LittleEndian.Uint16(makePkt[2:4]) != 12 {
		t.Fatal("laphine make")
	}
	eden := BuildEdenReqPacket(EdenActAccept, 99)
	if len(eden) != 8 || binary.LittleEndian.Uint32(eden[4:8]) != 99 {
		t.Fatal("eden")
	}
	open := make([]byte, 6)
	binary.LittleEndian.PutUint16(open[0:2], PacketZCReformOpen)
	binary.LittleEndian.PutUint32(open[2:6], 26007)
	got, ok, err := ParseReformOpen(Packet{ID: PacketZCReformOpen, Data: open})
	if err != nil || !ok || got.ItemID != 26007 {
		t.Fatalf("open %+v ok=%t err=%v", got, ok, err)
	}
}
