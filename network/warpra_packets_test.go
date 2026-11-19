package network

import (
	"encoding/binary"
	"testing"
)

func TestWarpraEntrySizeMatchesHercules(t *testing.T) {
	// turoran_warpra.c: WARPRA_ENTRY_SIZE (1+2+1+2+2+1+16+40+24)
	if WarpraEntrySize != 89 {
		t.Fatalf("WarpraEntrySize = %d, want 89", WarpraEntrySize)
	}
	if WarpraListHeaderLen != 17 {
		t.Fatalf("WarpraListHeaderLen = %d, want 17", WarpraListHeaderLen)
	}
	if WarpraCZReqLen != 8 || WarpraZCResultLen != 8 {
		t.Fatalf("fixed lengths CZ=%d ZC_RESULT=%d", WarpraCZReqLen, WarpraZCResultLen)
	}
}

func TestWarpraRequestMatchesHerculesCZLayout(t *testing.T) {
	packet := BuildWarpraRequestPacket(WarpraActWarp, uint16(WarpraKindDungeon), 7)
	if len(packet) != 8 {
		t.Fatalf("len = %d, want 8", len(packet))
	}
	if ID(packet) != PacketCZWarpraReq {
		t.Fatalf("opcode = 0x%04X, want 0x0EF4", ID(packet))
	}
	if binary.LittleEndian.Uint16(packet[2:4]) != WarpraActWarp {
		t.Fatalf("action = %d", binary.LittleEndian.Uint16(packet[2:4]))
	}
	if binary.LittleEndian.Uint16(packet[4:6]) != uint16(WarpraKindDungeon) {
		t.Fatalf("kind = %d", binary.LittleEndian.Uint16(packet[4:6]))
	}
	if binary.LittleEndian.Uint16(packet[6:8]) != 7 {
		t.Fatalf("id = %d", binary.LittleEndian.Uint16(packet[6:8]))
	}
	list := BuildWarpraRequestPacket(WarpraActList, 0, 0)
	if binary.LittleEndian.Uint16(list[2:4]) != WarpraActList || binary.LittleEndian.Uint32(list[4:]) != 0 {
		t.Fatalf("list request = %x", list)
	}
	buy := BuildWarpraRequestPacket(WarpraActBuy, uint16(WarpraKindTown), 1)
	if binary.LittleEndian.Uint16(buy[2:4]) != WarpraActBuy {
		t.Fatalf("buy action = %d", binary.LittleEndian.Uint16(buy[2:4]))
	}
}

func TestWarpraListRoundTripHerculesLayout(t *testing.T) {
	want := WarpraList{
		Zeny:  12_345_678,
		Price: WarpraDefaultPrice,
		Flags: 0,
		Total: 2,
		Locs: []WarpraLoc{
			{Kind: WarpraKindTown, ID: 0, Unlocked: true, X: 156, Y: 187, Map: "prontera", Name: "Prontera"},
			{Kind: WarpraKindDungeon, ID: 3, Unlocked: false, X: 168, Y: 168, Group: 1, Map: "iz_dun00", Name: "Byalan Dungeon 1", GroupName: "Byalan"},
		},
	}
	data := EncodeWarpraList(want)
	if len(data) != WarpraListHeaderLen+2*WarpraEntrySize {
		t.Fatalf("encoded length %d", len(data))
	}
	if data[12] != 0 {
		t.Fatal("more flag set on a complete list")
	}
	got, ok, err := ParseWarpraList(Packet{ID: PacketZCWarpraList, Data: data})
	if err != nil || !ok {
		t.Fatalf("parse: ok=%t err=%v", ok, err)
	}
	if got.Zeny != want.Zeny || got.Price != want.Price || got.Flags != 0 || got.Total != 2 || got.More() {
		t.Fatalf("header = %+v", got)
	}
	if len(got.Locs) != 2 {
		t.Fatalf("locs = %d", len(got.Locs))
	}
	if got.Locs[0] != want.Locs[0] {
		t.Fatalf("town = %+v, want %+v", got.Locs[0], want.Locs[0])
	}
	if got.Locs[1] != want.Locs[1] {
		t.Fatalf("dungeon = %+v, want %+v", got.Locs[1], want.Locs[1])
	}

	// Offsets match turoran_warpra.c: kind at 0, id at 1, unlocked at 3, x at 4, y at 6, group at 8.
	entry := data[WarpraListHeaderLen:]
	if entry[0] != WarpraKindTown || binary.LittleEndian.Uint16(entry[1:3]) != 0 || entry[3] != 1 {
		t.Fatalf("first entry header bytes = %x", entry[:9])
	}
	if binary.LittleEndian.Uint16(entry[4:6]) != 156 || binary.LittleEndian.Uint16(entry[6:8]) != 187 {
		t.Fatalf("first entry coords = %x", entry[4:8])
	}
	if string(entry[9:9+8]) != "prontera" {
		t.Fatalf("map bytes = %q", entry[9:9+16])
	}
}

func TestWarpraListChunksAndZeroPrice(t *testing.T) {
	first := EncodeWarpraList(WarpraList{
		Zeny:  1,
		Flags: WarpraListFlagMore,
		Total: 2,
		Locs:  []WarpraLoc{{Kind: WarpraKindTown, ID: 0, Map: "prontera", Name: "Prontera"}},
	})
	second := EncodeWarpraList(WarpraList{
		Zeny:  1,
		Price: 0, // server omitted; parser fills the 50,000,000 default
		Total: 2,
		Locs:  []WarpraLoc{{Kind: WarpraKindTown, ID: 1, Map: "geffen", Name: "Geffen"}},
	})
	a, ok, err := ParseWarpraList(Packet{ID: PacketZCWarpraList, Data: first})
	if err != nil || !ok || !a.More() || len(a.Locs) != 1 || a.Total != 2 {
		t.Fatalf("first chunk = %+v ok=%t err=%v", a, ok, err)
	}
	b, ok, err := ParseWarpraList(Packet{ID: PacketZCWarpraList, Data: second})
	if err != nil || !ok || b.More() || b.Price != WarpraDefaultPrice || b.Locs[0].Map != "geffen" {
		t.Fatalf("second chunk = %+v ok=%t err=%v", b, ok, err)
	}
}

func TestWarpraResultLayout(t *testing.T) {
	for _, tc := range []struct {
		result int16
		kind   uint16
		id     uint16
	}{
		{WarpraOK, uint16(WarpraKindTown), 0},
		{WarpraLocked, uint16(WarpraKindField), 12},
		{WarpraNoZeny, uint16(WarpraKindDungeon), 3},
		{WarpraLevel, uint16(WarpraKindSpecial), 1},
		{WarpraBusy, 0, 0},
		{WarpraInvalid, uint16(WarpraKindTown), 99},
		{WarpraErr, uint16(WarpraKindField), 2},
	} {
		data := EncodeWarpraResult(WarpraResult{Result: tc.result, Kind: tc.kind, ID: tc.id})
		if len(data) != 8 || ID(data) != PacketZCWarpraResult {
			t.Fatalf("encoded = %x", data)
		}
		got, ok, err := ParseWarpraResult(Packet{ID: PacketZCWarpraResult, Data: data})
		if err != nil || !ok || got.Result != tc.result || got.Kind != tc.kind || got.ID != tc.id {
			t.Fatalf("result = %+v ok=%t err=%v, want %d/%d/%d", got, ok, err, tc.result, tc.kind, tc.id)
		}
	}
}

func TestWarpraPacketsRejectMalformedAndIgnoreUnrelated(t *testing.T) {
	if _, ok, err := ParseWarpraList(Packet{ID: 0x0080}); ok || err != nil {
		t.Fatal("list handled unrelated packet")
	}
	if _, ok, err := ParseWarpraResult(Packet{ID: 0x0080}); ok || err != nil {
		t.Fatal("result handled unrelated packet")
	}
	shortList := EncodeWarpraList(WarpraList{Locs: []WarpraLoc{{Map: "prontera"}}})
	shortList = shortList[:len(shortList)-1]
	if _, ok, err := ParseWarpraList(Packet{ID: PacketZCWarpraList, Data: shortList}); err == nil || ok {
		t.Fatal("truncated list accepted")
	}
	badCount := EncodeWarpraList(WarpraList{Locs: []WarpraLoc{{Map: "prontera"}}})
	binary.LittleEndian.PutUint16(badCount[15:17], 9)
	if _, ok, err := ParseWarpraList(Packet{ID: PacketZCWarpraList, Data: badCount}); err == nil || ok {
		t.Fatal("count mismatch accepted")
	}
	if _, ok, err := ParseWarpraResult(Packet{ID: PacketZCWarpraResult, Data: make([]byte, 7)}); err == nil || ok {
		t.Fatal("short result accepted")
	}
	if _, ok, err := ParseWarpraResult(Packet{ID: PacketZCWarpraResult, Data: make([]byte, 9)}); err == nil || ok {
		t.Fatal("long result accepted")
	}
}

func TestWarpraListFramingAcrossPartialReads(t *testing.T) {
	chunk := EncodeWarpraList(WarpraList{
		Zeny:  50,
		Price: WarpraDefaultPrice,
		Total: 1,
		Locs:  []WarpraLoc{{Kind: WarpraKindTown, ID: 0, Unlocked: true, Map: "prontera", Name: "Prontera"}},
	})
	result := EncodeWarpraResult(WarpraResult{Result: WarpraOK, Kind: uint16(WarpraKindTown), ID: 0})
	stream := append(append([]byte{}, chunk...), result...)
	framer := NewFramer(PacketLengths20250716())
	var packets []Packet
	for _, b := range stream {
		part, err := framer.Push([]byte{b})
		if err != nil {
			t.Fatalf("frame error: %v", err)
		}
		packets = append(packets, part...)
	}
	if len(packets) != 2 || packets[0].ID != PacketZCWarpraList || packets[1].ID != PacketZCWarpraResult {
		t.Fatalf("packets = %v", packets)
	}
	list, ok, err := ParseWarpraList(packets[0])
	if err != nil || !ok || list.Locs[0].Map != "prontera" {
		t.Fatalf("framed list = %+v ok=%t err=%v", list, ok, err)
	}
	res, ok, err := ParseWarpraResult(packets[1])
	if err != nil || !ok || res.Result != WarpraOK {
		t.Fatalf("framed result = %+v ok=%t err=%v", res, ok, err)
	}
}

func TestNormalizeWarpraMapID(t *testing.T) {
	for _, tc := range []struct{ in, want string }{
		{"Prontera.rsw", "prontera"},
		{"iz_dun00.gat", "iz_dun00"},
		{"  Geffen  ", "geffen"},
		{"alberta", "alberta"},
	} {
		if got := NormalizeWarpraMapID(tc.in); got != tc.want {
			t.Fatalf("NormalizeWarpraMapID(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestWarpraCustomLengthsOn20250716Profile(t *testing.T) {
	table := PacketLengths20250716()
	if table[PacketZCWarpraList] != -1 {
		t.Fatalf("0x0EFE length = %d, want -1", table[PacketZCWarpraList])
	}
	if table[PacketZCWarpraResult] != 8 {
		t.Fatalf("0x0EFF length = %d, want 8", table[PacketZCWarpraResult])
	}
	if table[PacketCZWarpraReq] != 0 {
		t.Fatalf("0x0EF4 is CZ-only and must stay out of the ZC length table, got %d", table[PacketCZWarpraReq])
	}
	if PacketLengths2008()[PacketZCWarpraList] != 0 {
		t.Fatal("2008 profile gained Warpra lengths")
	}
}
