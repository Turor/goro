package network

import (
	"bufio"
	"os"
	"strconv"
	"strings"
	"testing"
)

// Fixtures follow Hercules map/packets_struct.h at PACKETVER 20250716
// (main client): NORMALITEM_INFO 34 bytes, EQUIPITEM_INFO 83 bytes,
// SKILLDATA 15 bytes, 32-bit item and card IDs.

func normalItem2025(index uint16, itemID uint32, amount uint16, cards [4]uint32) []byte {
	e := make([]byte, normalItemInfo2025Size)
	put16(e, 0, index)
	put32(e, 2, itemID)
	e[6] = 3 // etc
	put16(e, 7, amount)
	for i, c := range cards {
		put32(e, 13+4*i, c)
	}
	e[33] = 1 // identified
	return e
}

func equipItem2025(index uint16, itemID uint32, location, wear uint32, refine uint8, cards [4]uint32) []byte {
	e := make([]byte, equipItemInfo2025Size)
	put16(e, 0, index)
	put32(e, 2, itemID)
	e[6] = 5 // armor
	put32(e, 7, location)
	put32(e, 11, wear)
	for i, c := range cards {
		put32(e, 15+4*i, c)
	}
	put16(e, 40, 8)
	put16(e, 42, 15)
	e[80] = refine
	e[81] = 2     // grade
	e[82] = 1 | 2 // identified, damaged
	return e
}

func itemList2025(id uint16, invType uint8, entries ...[]byte) Packet {
	data := []byte{0, 0, 0, 0, invType}
	for _, e := range entries {
		data = append(data, e...)
	}
	put16(data, 0, id)
	put16(data, 2, uint16(len(data)))
	return Packet{ID: id, Data: data}
}

func TestParseInventoryLists2025(t *testing.T) {
	normal := itemList2025(0x0B09, InventoryTypeInventory, normalItem2025(2, 1000123, 5, [4]uint32{}))
	items, ok, err := ParseInventoryItemList(normal)
	if err != nil || !ok || len(items) != 1 {
		t.Fatalf("normal list ok=%t err=%v items=%d", ok, err, len(items))
	}
	if it := items[0]; it.Index != 2 || it.ItemID != 1000123 || it.Amount != 5 || it.Type != 3 || !it.Identified || it.Equip {
		t.Fatalf("normal item = %+v", it)
	}

	equip := itemList2025(0x0B39, InventoryTypeInventory,
		equipItem2025(3, 1201, 0x2, 0x2, 7, [4]uint32{4001234, 0, 0, 0}),
		equipItem2025(4, 2301, 0x10, 0, 0, [4]uint32{}))
	if len(equip.Data) != 5+2*83 {
		t.Fatalf("fixture size %d", len(equip.Data))
	}
	items, ok, err = ParseInventoryItemList(equip)
	if err != nil || !ok || len(items) != 2 {
		t.Fatalf("equip list ok=%t err=%v items=%d", ok, err, len(items))
	}
	if it := items[0]; it.ItemID != 1201 || it.Location != 0x2 || !it.Equipped || it.WearLocation != 0x2 ||
		it.Refine != 7 || it.Grade != 2 || it.Options[0].Index != 8 || it.Options[0].Value != 15 || it.Cards[0] != 4001234 || !it.Identified || !it.Damaged || !it.Equip || it.Amount != 1 {
		t.Fatalf("equip item = %+v", it)
	}
	if it := items[1]; it.Index != 4 || it.ItemID != 2301 || it.Equipped || it.Location != 0x10 {
		t.Fatalf("second equip item = %+v", it)
	}

	// The inventory type byte routes 0x0B09/0x0B39 to cart and storage.
	cart := itemList2025(0x0B09, InventoryTypeCart, normalItem2025(1, 501, 9, [4]uint32{}))
	if _, ok, _ := ParseInventoryItemList(cart); ok {
		t.Fatal("cart list parsed as inventory")
	}
	if items, ok, err := ParseCartItemList(cart); err != nil || !ok || len(items) != 1 || items[0].ItemID != 501 {
		t.Fatalf("cart list ok=%t err=%v items=%+v", ok, err, items)
	}
	storage := itemList2025(0x0B39, InventoryTypeStorage, equipItem2025(9, 1202, 0x2, 0, 4, [4]uint32{}))
	if _, ok, _ := ParseInventoryItemList(storage); ok {
		t.Fatal("storage list parsed as inventory")
	}
	if items, ok, err := ParseStorageItemList(storage); err != nil || !ok || len(items) != 1 || items[0].Refine != 4 {
		t.Fatalf("storage list ok=%t err=%v items=%+v", ok, err, items)
	}

	bad := itemList2025(0x0B09, InventoryTypeInventory, make([]byte, 33))
	if _, _, err := ParseInventoryItemList(bad); err == nil {
		t.Fatal("truncated entry accepted")
	}
}

func TestParseInventoryBoundary2025(t *testing.T) {
	start := []byte{0, 0, 0, 0, InventoryTypeStorage}
	start = append(start, "Kafra\x00"...)
	put16(start, 0, 0x0B08)
	put16(start, 2, uint16(len(start)))
	b, ok, err := ParseInventoryBoundary(Packet{ID: 0x0B08, Data: start})
	if err != nil || !ok || !b.Start || b.Type != InventoryTypeStorage || b.Name != "Kafra" {
		t.Fatalf("start = %+v ok=%t err=%v", b, ok, err)
	}
	end := []byte{0x0B, 0x0B, InventoryTypeInventory, 0}
	b, ok, err = ParseInventoryBoundary(Packet{ID: 0x0B0B, Data: end})
	if err != nil || !ok || b.Start || b.Type != InventoryTypeInventory {
		t.Fatalf("end = %+v ok=%t err=%v", b, ok, err)
	}
}

func TestParseItemPickupAck2025(t *testing.T) {
	data := make([]byte, 85)
	put16(data, 0, 0x0B41)
	put16(data, 2, 7)
	put16(data, 4, 3)
	put32(data, 6, 1000123)
	data[10] = 1
	put32(data, 12, 4001234)
	put32(data, 28, 0x2)
	data[32] = 4
	data[33] = 0
	data[83] = 9
	data[84] = 3
	put16(data, 40, 12)
	put16(data, 42, 50)
	ack, ok, err := ParseItemPickupAck(Packet{ID: 0x0B41, Data: data})
	if err != nil || !ok {
		t.Fatalf("ok=%t err=%v", ok, err)
	}
	if ack.Index != 7 || ack.Amount != 3 || ack.ItemID != 1000123 || !ack.Identified || ack.Cards[0] != 4001234 ||
		ack.Location != 0x2 || ack.Type != 4 || ack.Result != 0 || ack.Refine != 9 || ack.Grade != 3 || ack.Options[0].Index != 12 || ack.Options[0].Value != 50 {
		t.Fatalf("ack = %+v", ack)
	}
}

func TestParseFloorItems2025(t *testing.T) {
	entry := make([]byte, 19)
	put16(entry, 0, 0x009D)
	put32(entry, 2, 5555)
	put32(entry, 6, 1000123)
	entry[10] = 1
	put16(entry, 11, 81)
	put16(entry, 13, 105)
	put16(entry, 15, 2)
	entry[17], entry[18] = 3, 4
	it, ok, err := ParseFloorItemEntry(Packet{ID: 0x009D, Data: entry})
	if err != nil || !ok || it.ID != 5555 || it.ItemID != 1000123 || !it.Identified || it.X != 81 || it.Y != 105 ||
		it.Amount != 2 || it.SubX != 3 || it.SubY != 4 || it.Falling {
		t.Fatalf("entry = %+v ok=%t err=%v", it, ok, err)
	}
	fall := make([]byte, 24)
	put16(fall, 0, 0x0ADD)
	put32(fall, 2, 5556)
	put32(fall, 6, 512)
	put16(fall, 10, 0)
	fall[12] = 1
	put16(fall, 13, 82)
	put16(fall, 15, 106)
	fall[17], fall[18] = 5, 6
	put16(fall, 19, 1)
	it, ok, err = ParseFloorItemEntry(Packet{ID: 0x0ADD, Data: fall})
	if err != nil || !ok || it.ID != 5556 || it.ItemID != 512 || it.X != 82 || it.Y != 106 || it.Amount != 1 ||
		it.SubX != 5 || it.SubY != 6 || !it.Falling {
		t.Fatalf("fall = %+v ok=%t err=%v", it, ok, err)
	}
}

func TestParseEquipAcks2025(t *testing.T) {
	wear := make([]byte, 11)
	put16(wear, 0, 0x0999)
	put16(wear, 2, 3)
	put32(wear, 4, 0x2)
	wear[10] = 0
	ack, ok, err := ParseInventoryEquipAck(Packet{ID: 0x0999, Data: wear})
	if err != nil || !ok || ack.Index != 3 || ack.Location != 0x2 || !ack.Success || ack.Unequip {
		t.Fatalf("wear = %+v ok=%t err=%v", ack, ok, err)
	}
	wear[10] = 1
	if ack, _, _ := ParseInventoryEquipAck(Packet{ID: 0x0999, Data: wear}); ack.Success {
		t.Fatal("wear failure reported as success")
	}
	off := make([]byte, 9)
	put16(off, 0, 0x099A)
	put16(off, 2, 3)
	put32(off, 4, 0x2)
	ack, ok, err = ParseInventoryEquipAck(Packet{ID: 0x099A, Data: off})
	if err != nil || !ok || ack.Index != 3 || !ack.Success || !ack.Unequip {
		t.Fatalf("takeoff = %+v ok=%t err=%v", ack, ok, err)
	}
}

func skillData2025(id uint16, inf uint32, level, sp, rng uint16, up uint8) []byte {
	e := make([]byte, skillData2025Size)
	put16(e, 0, id)
	put32(e, 2, inf)
	put16(e, 6, level)
	put16(e, 8, sp)
	put16(e, 10, rng)
	e[12] = up
	put16(e, 13, level)
	return e
}

func TestParseSkills2025(t *testing.T) {
	data := []byte{0, 0, 0, 0}
	data = append(data, skillData2025(1, 0, 9, 0, 0, 0)...)
	data = append(data, skillData2025(142, 4, 1, 3, 1, 1)...)
	put16(data, 0, 0x0B32)
	put16(data, 2, uint16(len(data)))
	list, ok, err := ParseSkillInfoList(Packet{ID: 0x0B32, Data: data})
	if err != nil || !ok || len(list.Skills) != 2 {
		t.Fatalf("list ok=%t err=%v %+v", ok, err, list)
	}
	if s := list.Skills[1]; s.ID != 142 || s.Type != 4 || s.Level != 1 || s.SPCost != 3 || s.Range != 1 || !s.Upgradable {
		t.Fatalf("skill = %+v", s)
	}
	for _, id := range []uint16{0x0B31, 0x0B33} {
		pkt := append([]byte{0, 0}, skillData2025(28, 16, 5, 13, 9, 1)...)
		put16(pkt, 0, id)
		up, ok, err := ParseSkillInfoUpdate(Packet{ID: id, Data: pkt})
		if err != nil || !ok || up.Skill.ID != 28 || up.Skill.Level != 5 || up.Skill.SPCost != 13 || up.Skill.Range != 9 {
			t.Fatalf("0x%04X = %+v ok=%t err=%v", id, up, ok, err)
		}
	}
	cast := make([]byte, 29)
	put16(cast, 0, 0x0B1A)
	put32(cast, 2, 150051)
	put32(cast, 6, 150052)
	put16(cast, 10, 81)
	put16(cast, 12, 105)
	put16(cast, 14, 28)
	put32(cast, 16, 6)
	put32(cast, 20, 1500)
	c, ok, err := ParseSkillCastNotify(Packet{ID: 0x0B1A, Data: cast})
	if err != nil || !ok || c.SourceID != 150051 || c.TargetID != 150052 || c.X != 81 || c.Y != 105 ||
		c.SkillID != 28 || c.Property != 6 || c.DelayTime != 1500 {
		t.Fatalf("cast = %+v ok=%t err=%v", c, ok, err)
	}
}

func TestParseHotkeyListV3(t *testing.T) {
	data := make([]byte, 271)
	put16(data, 0, 0x0B20)
	data[5] = 1 // skill
	put32(data, 6, 28)
	put16(data, 10, 5)
	data[5+7*26] = 0
	put32(data, 5+7*26+1, 1000123)
	list, ok, err := ParseHotkeyList(Packet{ID: 0x0B20, Data: data})
	if err != nil || !ok || len(list.Slots) != hotkeySlotsV3 {
		t.Fatalf("ok=%t err=%v slots=%d", ok, err, len(list.Slots))
	}
	if s := list.Slots[0]; s.Type != 1 || s.ID != 28 || s.Level != 5 {
		t.Fatalf("slot0 = %+v", s)
	}
	if s := list.Slots[26]; s.ID != 1000123 {
		t.Fatalf("slot26 = %+v", s)
	}
	put16(data, 3, 1) // extra tab is stored, not applied to the skill bar
	extra, ok, err := ParseHotkeyList(Packet{ID: 0x0B20, Data: data})
	if err != nil || !ok || extra.Tab != 1 {
		t.Fatalf("tab1 ok=%t err=%v tab=%d", ok, err, extra.Tab)
	}
}

func TestFixedLengths2025ItemPackets(t *testing.T) {
	table := PacketLengths20250716()
	for id, want := range map[uint16]int{
		0x0B41: 85, 0x0ADD: 24, 0x009D: 19, 0x0999: 11, 0x099A: 9, 0x0B31: 17,
		0x0B33: 17, 0x0B1A: 29, 0x0B20: 271, 0x0B0B: 4, 0x01C8: 15,
	} {
		if table[id] != want {
			t.Errorf("0x%04X length %d, want %d", id, table[id], want)
		}
	}
}

// loadHerculesCZHandlers reads the generated 20250716 client->server map
// (scripts/packets/gen-handlers-20250716.sh).
func loadHerculesCZHandlers(t *testing.T) map[uint16]string {
	t.Helper()
	f, err := os.Open("testdata/hercules_cz_handlers_20250716.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	out := map[uint16]string{}
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		fields := strings.Fields(sc.Text())
		if len(fields) != 2 {
			continue
		}
		id, err := strconv.ParseUint(fields[0], 0, 16)
		if err != nil {
			t.Fatal(err)
		}
		out[uint16(id)] = strings.TrimPrefix(fields[1], "clif->")
	}
	return out
}

// Every client->server packet goro builds for 20250716 must land on the
// intended Hercules handler with the length Hercules expects. Old 2008
// shuffled opcodes are live handlers for other things in this version
// (0x00F3 is global chat, 0x00A7 is walk), so a stale opcode is not just
// ignored.
func TestClientPackets20250716HitHerculesHandlers(t *testing.T) {
	handlers := loadHerculesCZHandlers(t)
	lengths := PacketLengths20250716()
	d := ClientDate20250716
	walk, _ := BuildWalkToXYPacketForClientDate(85, 105, d)
	name, _ := BuildNameRequestPacketForClientDate(150052, d)
	cases := []struct {
		name    string
		packet  []byte
		handler string
	}{
		{"walk", walk, "pWalkToXY"},
		{"tick", BuildTickSendPacketForClientDate(1, d), "pTickSend"},
		{"direction", BuildChangeDirectionPacketForClientDate(0, 2, d), "pChangeDir"},
		{"name", name, "pGetCharNameRequest"},
		{"action", BuildActionRequestPacketForClientDate(150052, ActionAttackOnce, d), "pActionRequest"},
		{"enter", BuildMapServerEnterPacketForClientDate(MapServerEnter{AccountID: 1, CharID: 2, AuthCode: 3}, d), "pWantToConnection"},
		{"chat", BuildGlobalChatPacketForClientDate("kgap09", "hi", d), "pGlobalMessage"},
		{"pickup", BuildItemPickupPacketForClientDate(5555, d), "pTakeItem"},
		{"use item", BuildUseInventoryItemPacketForClientDate(2, 2000031, d), "pUseItem"},
		{"drop", BuildDropInventoryItemPacketForClientDate(2, 1, d), "pDropItem"},
		{"wear", BuildWearEquipPacketForClientDate(3, 0x2, d), "pEquipItem"},
		{"takeoff", BuildTakeoffEquipPacket(3), "pUnequipItem"},
		{"to storage", BuildMoveToStoragePacketForClientDate(2, 1, d), "pMoveToKafra"},
		{"from storage", BuildMoveFromStoragePacketForClientDate(2, 1, d), "pMoveFromKafra"},
		{"close storage", BuildCloseStoragePacketForClientDate(d), "pCloseKafra"},
		{"skill to id", BuildUseSkillToIDPacketForClientDate(28, 1, 150052, d), "pUseSkillToId"},
		{"skill to ground", BuildUseSkillToGroundPacketForClientDate(12, 1, 85, 105, d), "pUseSkillToPos"},
		{"skill to ground text", BuildUseSkillToGroundWithTextPacketForClientDate(12, 1, 85, 105, "x", d), "pUseSkillToPosMoreInfo"},
		{"hotkey", BuildHotkeyPacket(0, HotkeySlot{Type: 1, ID: 28, Level: 1}), "pHotkey1"},
	}
	for _, tc := range cases {
		if len(tc.packet) < 2 {
			t.Errorf("%s: empty packet", tc.name)
			continue
		}
		id := ID(tc.packet)
		if got := handlers[id]; got != tc.handler {
			t.Errorf("%s: opcode 0x%04X goes to %q, want %q", tc.name, id, got, tc.handler)
		}
		if want := lengths[id]; want > 0 && want != len(tc.packet) {
			t.Errorf("%s: opcode 0x%04X len %d, Hercules expects %d", tc.name, id, len(tc.packet), want)
		}
	}
}
