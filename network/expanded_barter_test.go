package network

import (
	"encoding/binary"
	"testing"
)

func TestParseExpandedBarterList20250716(t *testing.T) {
	// One zeny-only item, then one item paid with a single currency.
	body := make([]byte, 8+expandedBarterItemHead+expandedBarterItemHead+expandedBarterCurrency)
	binary.LittleEndian.PutUint16(body[0:2], PacketZCExpandedBarter)
	binary.LittleEndian.PutUint16(body[2:4], uint16(len(body)))
	binary.LittleEndian.PutUint32(body[4:8], 2)

	off := 8
	binary.LittleEndian.PutUint32(body[off:off+4], 501)
	binary.LittleEndian.PutUint32(body[off+6:off+10], 3)
	binary.LittleEndian.PutUint32(body[off+14:off+18], 7)
	binary.LittleEndian.PutUint32(body[off+18:off+22], 100)
	binary.LittleEndian.PutUint32(body[off+28:off+32], 0)
	off += expandedBarterItemHead
	binary.LittleEndian.PutUint32(body[off:off+4], 502)
	binary.LittleEndian.PutUint32(body[off+6:off+10], 1)
	binary.LittleEndian.PutUint32(body[off+14:off+18], 8)
	binary.LittleEndian.PutUint32(body[off+28:off+32], 1)
	off += expandedBarterItemHead
	binary.LittleEndian.PutUint32(body[off:off+4], 909)
	binary.LittleEndian.PutUint32(body[off+6:off+10], 2)

	items, ok, err := ParseExpandedBarterList(Packet{ID: PacketZCExpandedBarter, Data: body})
	if err != nil || !ok || len(items) != 2 {
		t.Fatalf("parse ok=%v err=%v n=%d", ok, err, len(items))
	}
	if items[0].ItemID != 501 || items[0].Amount != 3 || items[0].Index != 7 || items[0].Cost != 100 || !items[0].Expanded {
		t.Fatalf("zeny item %+v", items[0])
	}
	if items[1].ItemID != 502 || items[1].Currency != 909 || items[1].Cost != 2 || items[1].Index != 8 {
		t.Fatalf("currency item %+v", items[1])
	}
	pkt := BuildExpandedBarterPurchasePacket(502, 8, 1)
	if len(pkt) != 16 || ID(pkt) != PacketCZExpandedBarterPurchase || binary.LittleEndian.Uint32(pkt[8:12]) != 8 {
		t.Fatalf("purchase packet % X", pkt)
	}
}
