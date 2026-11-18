package network

import (
	"encoding/binary"
	"testing"
)

// Layouts follow turoran_stable.c: STABLE_HOMUN_ENTRY 42, STABLE_PET_ENTRY 191,
// STABLE_PET_BUFF_LEN 96, NAME_LENGTH 24. Headers are opcode+len+id+count.

func TestStablePacketEntrySizesMatchHercules(t *testing.T) {
	if StableHomunculusEntryLen != 42 {
		t.Fatalf("homunculus entry %d, want 42", StableHomunculusEntryLen)
	}
	if StablePetEntryLen != 191 {
		t.Fatalf("pet entry %d, want 191", StablePetEntryLen)
	}
	if StablePetBuffLen != 96 || StableNameLen != 24 {
		t.Fatalf("buff %d name %d", StablePetBuffLen, StableNameLen)
	}
	if homunPackedSize() != StableHomunculusEntryLen {
		t.Fatalf("packed homunculus fields %d, want %d", homunPackedSize(), StableHomunculusEntryLen)
	}
	if petPackedSize() != StablePetEntryLen {
		t.Fatalf("packed pet fields %d, want %d", petPackedSize(), StablePetEntryLen)
	}
}

func homunPackedSize() int {
	return 4 + 2 + 2 + 2 + 4 + 2 + 1 + 1 + StableNameLen
}

func petPackedSize() int {
	return 4 + 4 + 2 + 2 + 2 + 4 + 4 + 4 + 1 + 1 + 1 + StableNameLen + StableNameLen + 2 + 2 + 2 + 4 + 4 + 4 + StablePetBuffLen
}

func TestBuildStableCZPacketsMatchHercules(t *testing.T) {
	open := BuildStableOpenPacket()
	if ID(open) != PacketCZTuroranStableOpen || len(open) != 2 {
		t.Fatalf("open = %x", open)
	}

	homun := BuildStableHomunculusPacket(0x11223344)
	if ID(homun) != PacketCZTuroranStableHomun || len(homun) != 6 {
		t.Fatalf("homun id/len = %x", homun)
	}
	if binary.LittleEndian.Uint32(homun[2:6]) != 0x11223344 {
		t.Fatalf("homun payload = %x", homun)
	}
	store := BuildStableHomunculusPacket(0)
	if len(store) != 6 || binary.LittleEndian.Uint32(store[2:6]) != 0 {
		t.Fatalf("store homun = %x", store)
	}

	pet := BuildStablePetPacket(StablePetSummon, 0x55667788)
	if ID(pet) != PacketCZTuroranStablePet || len(pet) != 8 {
		t.Fatalf("pet id/len = %x", pet)
	}
	if binary.LittleEndian.Uint16(pet[2:4]) != StablePetSummon || binary.LittleEndian.Uint32(pet[4:8]) != 0x55667788 {
		t.Fatalf("pet payload = %x", pet)
	}
}

func TestParseStableHomunculusListMatchesHerculesLayout(t *testing.T) {
	data := encodeStableHomunculusList(0x00AABBCC, []StableHomunculusEntry{
		{
			HomunID: 1501, ClassID: 6001, PrevClass: 0, Level: 45,
			Intimacy: 91000, Hunger: 62, Vaporize: 0, RenameFlag: 1, Name: "Lif",
		},
		{
			HomunID: 1502, ClassID: 6048, PrevClass: 6002, Level: 99,
			Intimacy: 25000, Hunger: 80, Vaporize: 1, RenameFlag: 0, Name: "Amistr",
		},
	})
	if len(data) != StableHomunculusHeaderLen+2*StableHomunculusEntryLen {
		t.Fatalf("encoded length %d", len(data))
	}
	list, ok, err := ParseStableHomunculusList(Packet{ID: PacketZCTuroranStableHomun, Data: data})
	if err != nil || !ok {
		t.Fatalf("parse ok=%t err=%v", ok, err)
	}
	if list.ActiveHomunID != 0x00AABBCC || len(list.List) != 2 {
		t.Fatalf("header = %+v", list)
	}
	a, b := list.List[0], list.List[1]
	if a.HomunID != 1501 || a.ClassID != 6001 || a.Level != 45 || a.Intimacy != 91000 || a.Hunger != 62 || a.RenameFlag != 1 || a.Name != "Lif" {
		t.Fatalf("entry 0 = %+v", a)
	}
	if b.HomunID != 1502 || b.ClassID != 6048 || b.PrevClass != 6002 || b.Vaporize != 1 || b.Name != "Amistr" {
		t.Fatalf("entry 1 = %+v", b)
	}

	empty := encodeStableHomunculusList(0, nil)
	list, ok, err = ParseStableHomunculusList(Packet{ID: PacketZCTuroranStableHomun, Data: empty})
	if err != nil || !ok || list.ActiveHomunID != 0 || len(list.List) != 0 {
		t.Fatalf("empty = %+v ok=%t err=%v", list, ok, err)
	}
}

func TestParseStablePetListMatchesHerculesLayout(t *testing.T) {
	data := encodeStablePetList(77, []StablePetEntry{{
		PetID: 9001, ClassID: 1002, Level: 12, Intimate: 900, Hungry: 80,
		EggID: 9003, Equip: 10013, CharID: 150051, Incubate: 2, RenameFlag: 1,
		Flags: StableFlagOut, Name: "Poring", OwnerName: "kgap01",
		Atk: 10, Atk2: 14, ASPD: 150, Exp: 320, NextExp: 385, FoodID: 909,
		Buff: "Active: +2 STR",
	}})
	if len(data) != StablePetHeaderLen+StablePetEntryLen {
		t.Fatalf("encoded length %d, want %d", len(data), StablePetHeaderLen+StablePetEntryLen)
	}
	list, ok, err := ParseStablePetList(Packet{ID: PacketZCTuroranStablePet, Data: data})
	if err != nil || !ok {
		t.Fatalf("parse ok=%t err=%v", ok, err)
	}
	if list.ActivePetID != 77 || len(list.List) != 1 {
		t.Fatalf("header = %+v", list)
	}
	row := list.List[0]
	if row.PetID != 9001 || row.ClassID != 1002 || row.Level != 12 || row.Intimate != 900 || row.Hungry != 80 {
		t.Fatalf("identity = %+v", row)
	}
	if row.EggID != 9003 || row.Equip != 10013 || row.CharID != 150051 || row.Incubate != 2 || row.Flags != StableFlagOut {
		t.Fatalf("flags = %+v", row)
	}
	if row.Name != "Poring" || row.OwnerName != "kgap01" || row.Atk != 10 || row.Atk2 != 14 || row.ASPD != 150 {
		t.Fatalf("combat = %+v", row)
	}
	if row.Exp != 320 || row.NextExp != 385 || row.FoodID != 909 || row.Buff != "Active: +2 STR" {
		t.Fatalf("exp/buff = %+v", row)
	}
}

func TestParseStableOwnedAndResultMatchHerculesLayout(t *testing.T) {
	owned := encodeStableOwned([]StableOwnedEntry{{ItemID: 909, Owned: 12}, {ItemID: 512, Owned: 3}})
	if len(owned) != StableOwnedHeaderLen+2*StableOwnedEntryLen {
		t.Fatalf("owned length %d", len(owned))
	}
	list, ok, err := ParseStableOwnedList(Packet{ID: PacketZCTuroranStableOwned, Data: owned})
	if err != nil || !ok || len(list.List) != 2 {
		t.Fatalf("owned = %+v ok=%t err=%v", list, ok, err)
	}
	if list.List[0].ItemID != 909 || list.List[0].Owned != 12 || list.List[1].ItemID != 512 || list.List[1].Owned != 3 {
		t.Fatalf("owned rows = %+v", list.List)
	}

	result := make([]byte, 8)
	binary.LittleEndian.PutUint16(result[0:2], PacketZCTuroranStableResult)
	binary.LittleEndian.PutUint16(result[2:4], uint16(StableErrHavePet))
	binary.LittleEndian.PutUint32(result[4:8], 9001)
	got, ok, err := ParseStableResult(Packet{ID: PacketZCTuroranStableResult, Data: result})
	if err != nil || !ok || got.Result != StableErrHavePet || got.ID != 9001 {
		t.Fatalf("result = %+v ok=%t err=%v", got, ok, err)
	}
}

func TestParseStablePacketsRejectWrongIDAndTruncation(t *testing.T) {
	if _, ok, err := ParseStableHomunculusList(Packet{ID: PacketCZTuroranStableHomun, Data: make([]byte, 6)}); ok || err != nil {
		t.Fatalf("incoming 0x0EF1 must not parse as a stable homunculus list (it is ZC_TURORAN_FEATURES)")
	}
	if _, ok, err := ParseStablePetList(Packet{ID: PacketZCTuroranStableHomun, Data: make([]byte, 9)}); ok || err != nil {
		t.Fatal("homunculus list parsed as pet list")
	}
	shortHomun := make([]byte, 8)
	binary.LittleEndian.PutUint16(shortHomun[0:2], PacketZCTuroranStableHomun)
	if _, ok, err := ParseStableHomunculusList(Packet{ID: PacketZCTuroranStableHomun, Data: shortHomun}); !ok && err == nil {
		t.Fatal("short homunculus list accepted")
	}
	truncated := encodeStableHomunculusList(1, []StableHomunculusEntry{{HomunID: 1, Name: "x"}})
	truncated[8] = 2
	if _, ok, err := ParseStableHomunculusList(Packet{ID: PacketZCTuroranStableHomun, Data: truncated}); !ok && err == nil {
		t.Fatal("truncated homunculus count accepted")
	}
	shortPet := make([]byte, 8)
	binary.LittleEndian.PutUint16(shortPet[0:2], PacketZCTuroranStablePet)
	if _, ok, err := ParseStablePetList(Packet{ID: PacketZCTuroranStablePet, Data: shortPet}); !ok && err == nil {
		t.Fatal("short pet list accepted")
	}
	shortOwned := make([]byte, 5)
	binary.LittleEndian.PutUint16(shortOwned[0:2], PacketZCTuroranStableOwned)
	if _, ok, err := ParseStableOwnedList(Packet{ID: PacketZCTuroranStableOwned, Data: shortOwned}); !ok && err == nil {
		t.Fatal("short owned list accepted")
	}
	shortResult := make([]byte, 7)
	binary.LittleEndian.PutUint16(shortResult[0:2], PacketZCTuroranStableResult)
	if _, ok, err := ParseStableResult(Packet{ID: PacketZCTuroranStableResult, Data: shortResult}); !ok && err == nil {
		t.Fatal("short result accepted")
	}
}

func TestStableCustomLengthsAreFramed(t *testing.T) {
	table := PacketLengths20250716()
	for id, want := range map[uint16]int{
		PacketZCTuroranStableOwned:  -1,
		PacketZCTuroranStableHomun:  -1,
		PacketZCTuroranStablePet:    -1,
		PacketZCTuroranStableResult: 8,
	} {
		if got := table[id]; got != want {
			t.Errorf("length 0x%04X = %d, want %d", id, got, want)
		}
	}
	homun := encodeStableHomunculusList(1, []StableHomunculusEntry{{HomunID: 1, Name: "Lif"}})
	pet := encodeStablePetList(2, nil)
	owned := encodeStableOwned(nil)
	result := make([]byte, 8)
	binary.LittleEndian.PutUint16(result[0:2], PacketZCTuroranStableResult)
	stream := append(append(append(append([]byte{}, homun...), pet...), owned...), result...)
	framer := NewFramer(table)
	var got []Packet
	for _, b := range stream {
		packets, err := framer.Push([]byte{b})
		if err != nil {
			t.Fatalf("frame: %v", err)
		}
		got = append(got, packets...)
	}
	if len(got) != 4 || got[0].ID != PacketZCTuroranStableHomun || got[1].ID != PacketZCTuroranStablePet ||
		got[2].ID != PacketZCTuroranStableOwned || got[3].ID != PacketZCTuroranStableResult {
		t.Fatalf("framed = %v", got)
	}
}

func encodeStableHomunculusList(active uint32, rows []StableHomunculusEntry) []byte {
	total := StableHomunculusHeaderLen + len(rows)*StableHomunculusEntryLen
	data := make([]byte, total)
	binary.LittleEndian.PutUint16(data[0:2], PacketZCTuroranStableHomun)
	binary.LittleEndian.PutUint16(data[2:4], uint16(total))
	binary.LittleEndian.PutUint32(data[4:8], active)
	data[8] = byte(len(rows))
	off := StableHomunculusHeaderLen
	for _, row := range rows {
		binary.LittleEndian.PutUint32(data[off:off+4], row.HomunID)
		binary.LittleEndian.PutUint16(data[off+4:off+6], row.ClassID)
		binary.LittleEndian.PutUint16(data[off+6:off+8], row.PrevClass)
		binary.LittleEndian.PutUint16(data[off+8:off+10], row.Level)
		binary.LittleEndian.PutUint32(data[off+10:off+14], row.Intimacy)
		binary.LittleEndian.PutUint16(data[off+14:off+16], row.Hunger)
		data[off+16] = row.Vaporize
		data[off+17] = row.RenameFlag
		copy(data[off+18:off+42], encodeROFixedString(row.Name, StableNameLen))
		off += StableHomunculusEntryLen
	}
	return data
}

func encodeStablePetList(active uint32, rows []StablePetEntry) []byte {
	total := StablePetHeaderLen + len(rows)*StablePetEntryLen
	data := make([]byte, total)
	binary.LittleEndian.PutUint16(data[0:2], PacketZCTuroranStablePet)
	binary.LittleEndian.PutUint16(data[2:4], uint16(total))
	binary.LittleEndian.PutUint32(data[4:8], active)
	data[8] = byte(len(rows))
	off := StablePetHeaderLen
	for _, row := range rows {
		binary.LittleEndian.PutUint32(data[off:off+4], row.PetID)
		binary.LittleEndian.PutUint32(data[off+4:off+8], row.ClassID)
		binary.LittleEndian.PutUint16(data[off+8:off+10], row.Level)
		binary.LittleEndian.PutUint16(data[off+10:off+12], row.Intimate)
		binary.LittleEndian.PutUint16(data[off+12:off+14], row.Hungry)
		binary.LittleEndian.PutUint32(data[off+14:off+18], row.EggID)
		binary.LittleEndian.PutUint32(data[off+18:off+22], row.Equip)
		binary.LittleEndian.PutUint32(data[off+22:off+26], row.CharID)
		data[off+26] = row.Incubate
		data[off+27] = row.RenameFlag
		data[off+28] = row.Flags
		copy(data[off+29:off+53], encodeROFixedString(row.Name, StableNameLen))
		copy(data[off+53:off+77], encodeROFixedString(row.OwnerName, StableNameLen))
		binary.LittleEndian.PutUint16(data[off+77:off+79], row.Atk)
		binary.LittleEndian.PutUint16(data[off+79:off+81], row.Atk2)
		binary.LittleEndian.PutUint16(data[off+81:off+83], row.ASPD)
		binary.LittleEndian.PutUint32(data[off+83:off+87], row.Exp)
		binary.LittleEndian.PutUint32(data[off+87:off+91], row.NextExp)
		binary.LittleEndian.PutUint32(data[off+91:off+95], row.FoodID)
		copy(data[off+95:off+191], encodeROFixedString(row.Buff, StablePetBuffLen))
		off += StablePetEntryLen
	}
	return data
}

func encodeStableOwned(rows []StableOwnedEntry) []byte {
	total := StableOwnedHeaderLen + len(rows)*StableOwnedEntryLen
	data := make([]byte, total)
	binary.LittleEndian.PutUint16(data[0:2], PacketZCTuroranStableOwned)
	binary.LittleEndian.PutUint16(data[2:4], uint16(total))
	binary.LittleEndian.PutUint16(data[4:6], uint16(len(rows)))
	off := StableOwnedHeaderLen
	for _, row := range rows {
		binary.LittleEndian.PutUint32(data[off:off+4], row.ItemID)
		binary.LittleEndian.PutUint32(data[off+4:off+8], row.Owned)
		off += StableOwnedEntryLen
	}
	return data
}
