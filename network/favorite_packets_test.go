package network

import (
	"testing"
)

func TestItemFavoritePacketRoundTrip(t *testing.T) {
	set := BuildItemFavoritePacket(7, true)
	if len(set) != 5 || set[0] != 0x07 || set[1] != 0x09 || set[2] != 7 || set[3] != 0 || set[4] != 0 {
		t.Fatalf("set packet = %x", set)
	}
	clear := BuildItemFavoritePacket(7, false)
	if clear[4] != 1 {
		t.Fatalf("clear flag = %d, want 1", clear[4])
	}
	got, ok, err := ParseItemFavorite(Packet{ID: 0x0908, Data: []byte{0x08, 0x09, 7, 0, 0}})
	if err != nil || !ok || got.Index != 7 || !got.Favorite {
		t.Fatalf("parse favorite = %+v ok=%t err=%v", got, ok, err)
	}
	cleared, ok, err := ParseItemFavorite(Packet{ID: 0x0908, Data: []byte{0x08, 0x09, 7, 0, 1}})
	if err != nil || !ok || cleared.Favorite {
		t.Fatalf("parse clear = %+v ok=%t err=%v", cleared, ok, err)
	}
	if _, ok, err := ParseItemFavorite(Packet{ID: 0x013C, Data: []byte{0x3c, 0x01, 1, 0}}); ok || err != nil {
		t.Fatal("unrelated packet was treated as a favorite")
	}
}
