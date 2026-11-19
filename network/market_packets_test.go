package network

import (
	"encoding/binary"
	"testing"
)

func TestMarketRouletteClanPackets(t *testing.T) {
	if len(BuildEmptyPacket(PacketCZOpenRoulette)) != 2 {
		t.Fatal("open")
	}
	buy := BuildMarketPurchasePacket(1201, 1)
	if ID(buy) != PacketCZMarketPurchase || binary.LittleEndian.Uint16(buy[2:4]) != 12 {
		t.Fatalf("buy %x", buy)
	}
	hot := BuildHotkeyPacketV2(1, 3, HotkeySlot{Type: 1, ID: 28, Level: 4})
	if len(hot) != 13 || ID(hot) != PacketCZHotkeyChangeV2 {
		t.Fatal("hotkey v2")
	}
	if binary.LittleEndian.Uint16(hot[2:4]) != 1 || binary.LittleEndian.Uint16(hot[4:6]) != 3 {
		t.Fatalf("tab/index %x", hot)
	}
	open := make([]byte, 23)
	binary.LittleEndian.PutUint16(open[0:2], PacketZCOpenRoulette)
	binary.LittleEndian.PutUint32(open[11:15], 50)
	got, ok, err := ParseRouletteOpen(Packet{ID: PacketZCOpenRoulette, Data: open})
	if err != nil || !ok || got.Gold != 50 {
		t.Fatalf("open parse %+v ok=%t err=%v", got, ok, err)
	}
}
