package network

import (
	"encoding/binary"
	"fmt"
)

// PACKETVER 20250716 (Hercules main client) item, skill and hotkey packets.
// Layouts follow Hercules src/map/packets_struct.h for that version: item IDs
// and card slots are 32-bit, item lists carry an inventory type byte, and the
// equipment entries carry option and forge data.
const (
	PacketZCInventoryStart         uint16 = 0x0B08
	PacketZCInventoryItemList2025  uint16 = 0x0B09
	PacketZCInventoryEquipList2025 uint16 = 0x0B39
	PacketZCInventoryEnd           uint16 = 0x0B0B
	PacketZCItemPickupAck2025      uint16 = 0x0B41
	PacketZCItemFallEntry2025      uint16 = 0x0ADD
	PacketZCReqWearEquipAck2       uint16 = 0x0999
	PacketZCReqTakeoffEquipAck2    uint16 = 0x099A
	PacketZCSkillInfoList2025      uint16 = 0x0B32
	PacketZCAddSkill2025           uint16 = 0x0B31
	PacketZCSkillInfoUpdate2       uint16 = 0x0B33
	PacketZCUseSkillAck2025        uint16 = 0x0B1A
	PacketZCShortcutKeyListV3      uint16 = 0x0B20
	PacketCZReqWearEquip2          uint16 = 0x0998
	PacketCZMoveToStorage2025      uint16 = 0x0364
	PacketCZMoveFromStorage2025    uint16 = 0x0365
	PacketCZUseSkillToGround2025   uint16 = 0x0366
)

// Inventory types carried by 0x0B08/0x0B09/0x0B39/0x0B0B.
const (
	InventoryTypeInventory    uint8 = 0
	InventoryTypeCart         uint8 = 1
	InventoryTypeStorage      uint8 = 2
	InventoryTypeGuildStorage uint8 = 3
)

const (
	normalItemInfo2025Size = 34
	equipItemInfo2025Size  = 83
	skillData2025Size      = 15
	hotkeySlotSize         = 7
	hotkeySlotsV3          = 38
)

// InventoryBoundary is ZC_INVENTORY_START (0x0B08) or ZC_INVENTORY_END
// (0x0B0B). They bracket the item lists of one inventory type.
type InventoryBoundary struct {
	Type  uint8
	Name  string
	Start bool
}

func ParseInventoryBoundary(packet Packet) (InventoryBoundary, bool, error) {
	switch packet.ID {
	case PacketZCInventoryStart:
		if len(packet.Data) < 5 {
			return InventoryBoundary{}, false, fmt.Errorf("ZC_INVENTORY_START too short: %d", len(packet.Data))
		}
		return InventoryBoundary{Type: packet.Data[4], Name: decodeROFixedString(packet.Data[5:]), Start: true}, true, nil
	case PacketZCInventoryEnd:
		if len(packet.Data) < 4 {
			return InventoryBoundary{}, false, fmt.Errorf("ZC_INVENTORY_END too short: %d", len(packet.Data))
		}
		return InventoryBoundary{Type: packet.Data[2]}, true, nil
	default:
		return InventoryBoundary{}, false, nil
	}
}

// parseItemList2025 decodes 0x0B09/0x0B39 when its inventory type matches.
func parseItemList2025(packet Packet, want uint8) ([]InventoryItem, bool, error) {
	if packet.ID != PacketZCInventoryItemList2025 && packet.ID != PacketZCInventoryEquipList2025 {
		return nil, false, nil
	}
	if len(packet.Data) < 5 {
		return nil, false, fmt.Errorf("item list 0x%04X too short: %d", packet.ID, len(packet.Data))
	}
	if packet.Data[4] != want {
		return nil, false, nil
	}
	if packet.ID == PacketZCInventoryItemList2025 {
		return parseNormalItems2025(packet)
	}
	return parseEquipItems2025(packet)
}

func parseNormalItems2025(packet Packet) ([]InventoryItem, bool, error) {
	const entrySize = normalItemInfo2025Size
	body := packet.Data[5:]
	if len(body)%entrySize != 0 {
		return nil, false, fmt.Errorf("ZC_NORMAL_ITEMLIST 0x%04X invalid length: %d", packet.ID, len(packet.Data))
	}
	items := make([]InventoryItem, 0, len(body)/entrySize)
	for offset := 0; offset+entrySize <= len(body); offset += entrySize {
		e := body[offset : offset+entrySize]
		items = append(items, InventoryItem{
			Index:      binary.LittleEndian.Uint16(e[0:2]),
			ItemID:     binary.LittleEndian.Uint32(e[2:6]),
			Type:       e[6],
			Amount:     binary.LittleEndian.Uint16(e[7:9]),
			Location:   binary.LittleEndian.Uint32(e[9:13]),
			Cards:      readItemCards32(e, 13),
			Identified: e[33]&1 != 0,
		})
	}
	return items, true, nil
}

func parseEquipItems2025(packet Packet) ([]InventoryItem, bool, error) {
	const entrySize = equipItemInfo2025Size
	body := packet.Data[5:]
	if len(body)%entrySize != 0 {
		return nil, false, fmt.Errorf("ZC_EQUIPMENT_ITEMLIST 0x%04X invalid length: %d", packet.ID, len(packet.Data))
	}
	items := make([]InventoryItem, 0, len(body)/entrySize)
	for offset := 0; offset+entrySize <= len(body); offset += entrySize {
		e := body[offset : offset+entrySize]
		wearState := binary.LittleEndian.Uint32(e[11:15])
		flag := e[82]
		items = append(items, InventoryItem{
			Index:        binary.LittleEndian.Uint16(e[0:2]),
			ItemID:       binary.LittleEndian.Uint32(e[2:6]),
			Type:         e[6],
			Location:     binary.LittleEndian.Uint32(e[7:11]),
			WearLocation: wearState,
			Amount:       1,
			Equip:        true,
			Equipped:     wearState != 0,
			Cards:        readItemCards32(e, 15),
			Refine:       e[80],
			Grade:        e[81],
			Identified:   flag&1 != 0,
			Damaged:      flag&2 != 0,
			Options:      readFiveItemOptions(e, 40),
			ForgeOptions: readThreeItemOptions(e, 65),
		})
	}
	return items, true, nil
}

func parseItemPickupAck2025(packet Packet) (ItemPickupAck, bool, error) {
	if len(packet.Data) < 85 {
		return ItemPickupAck{}, false, fmt.Errorf("ZC_ITEM_PICKUP_ACK 0x%04X too short: %d", packet.ID, len(packet.Data))
	}
	return ItemPickupAck{
		Index:        binary.LittleEndian.Uint16(packet.Data[2:4]),
		Amount:       binary.LittleEndian.Uint16(packet.Data[4:6]),
		ItemID:       binary.LittleEndian.Uint32(packet.Data[6:10]),
		Identified:   packet.Data[10] != 0,
		Damaged:      packet.Data[11] != 0,
		Cards:        readItemCards32(packet.Data, 12),
		Location:     binary.LittleEndian.Uint32(packet.Data[28:32]),
		Type:         packet.Data[32],
		Result:       packet.Data[33],
		Refine:       packet.Data[83],
		Grade:        packet.Data[84],
		Options:      readFiveItemOptions(packet.Data, 40),
		ForgeOptions: readThreeItemOptions(packet.Data, 65),
	}, true, nil
}

// readItemOptions reads count packed options (int16 index, int16 value, uint8 param).
// Empty slots, index 0, are omitted.
func readFiveItemOptions(data []byte, offset int) (out [5]ItemOption) {
	readItemOptionsInto(data, offset, out[:])
	return out
}

func readThreeItemOptions(data []byte, offset int) (out [3]ItemOption) {
	readItemOptionsInto(data, offset, out[:])
	return out
}

func readItemOptionsInto(data []byte, offset int, out []ItemOption) {
	for i := range out {
		at := offset + i*5
		if at+5 > len(data) {
			return
		}
		out[i] = ItemOption{
			Index: int16(binary.LittleEndian.Uint16(data[at : at+2])),
			Value: int16(binary.LittleEndian.Uint16(data[at+2 : at+4])),
			Param: data[at+4],
		}
	}
}

func parseItemEntry2025(packet Packet) (FloorItemEntry, bool, error) {
	if len(packet.Data) < 19 {
		return FloorItemEntry{}, false, fmt.Errorf("ZC_ITEM_ENTRY too short: %d", len(packet.Data))
	}
	return FloorItemEntry{
		ID:         binary.LittleEndian.Uint32(packet.Data[2:6]),
		ItemID:     binary.LittleEndian.Uint32(packet.Data[6:10]),
		Identified: packet.Data[10] != 0,
		X:          int(binary.LittleEndian.Uint16(packet.Data[11:13])),
		Y:          int(binary.LittleEndian.Uint16(packet.Data[13:15])),
		Amount:     binary.LittleEndian.Uint16(packet.Data[15:17]),
		SubX:       packet.Data[17],
		SubY:       packet.Data[18],
	}, true, nil
}

func parseItemFallEntry2025(packet Packet) (FloorItemEntry, bool, error) {
	if len(packet.Data) < 24 {
		return FloorItemEntry{}, false, fmt.Errorf("ZC_ITEM_FALL_ENTRY 0x0ADD too short: %d", len(packet.Data))
	}
	return FloorItemEntry{
		ID:         binary.LittleEndian.Uint32(packet.Data[2:6]),
		ItemID:     binary.LittleEndian.Uint32(packet.Data[6:10]),
		Identified: packet.Data[12] != 0,
		X:          int(binary.LittleEndian.Uint16(packet.Data[13:15])),
		Y:          int(binary.LittleEndian.Uint16(packet.Data[15:17])),
		SubX:       packet.Data[17],
		SubY:       packet.Data[18],
		Amount:     binary.LittleEndian.Uint16(packet.Data[19:21]),
		Falling:    true,
	}, true, nil
}

func parseEquipAck2025(packet Packet) (InventoryEquipAck, bool, error) {
	switch packet.ID {
	case PacketZCReqWearEquipAck2:
		if len(packet.Data) < 11 {
			return InventoryEquipAck{}, false, fmt.Errorf("ZC_REQ_WEAR_EQUIP_ACK2 too short: %d", len(packet.Data))
		}
		return InventoryEquipAck{
			Index:    binary.LittleEndian.Uint16(packet.Data[2:4]),
			Location: binary.LittleEndian.Uint32(packet.Data[4:8]),
			Success:  packet.Data[10] == 0,
		}, true, nil
	case PacketZCReqTakeoffEquipAck2:
		if len(packet.Data) < 9 {
			return InventoryEquipAck{}, false, fmt.Errorf("ZC_REQ_TAKEOFF_EQUIP_ACK2 too short: %d", len(packet.Data))
		}
		return InventoryEquipAck{
			Index:    binary.LittleEndian.Uint16(packet.Data[2:4]),
			Location: binary.LittleEndian.Uint32(packet.Data[4:8]),
			Success:  packet.Data[8] == 0,
			Unequip:  true,
		}, true, nil
	default:
		return InventoryEquipAck{}, false, nil
	}
}

// parseSkillData2025 decodes the 15-byte SKILLDATA. The skill name is no
// longer sent; it comes from the client skill tables.
func parseSkillData2025(data []byte) SkillInfo {
	return SkillInfo{
		ID:         binary.LittleEndian.Uint16(data[0:2]),
		Type:       binary.LittleEndian.Uint32(data[2:6]),
		Level:      int(binary.LittleEndian.Uint16(data[6:8])),
		SPCost:     int(binary.LittleEndian.Uint16(data[8:10])),
		Range:      int(binary.LittleEndian.Uint16(data[10:12])),
		Upgradable: data[12] != 0,
	}
}

func parseSkillInfoList2025(packet Packet) (SkillInfoList, bool, error) {
	if len(packet.Data) < 4 {
		return SkillInfoList{}, false, fmt.Errorf("ZC_SKILLINFO_LIST 0x0B32 too short: %d", len(packet.Data))
	}
	body := packet.Data[4:]
	if len(body)%skillData2025Size != 0 {
		return SkillInfoList{}, false, fmt.Errorf("ZC_SKILLINFO_LIST 0x0B32 bad body len: %d", len(body))
	}
	skills := make([]SkillInfo, 0, len(body)/skillData2025Size)
	for offset := 0; offset < len(body); offset += skillData2025Size {
		skills = append(skills, parseSkillData2025(body[offset:offset+skillData2025Size]))
	}
	return SkillInfoList{Skills: skills}, true, nil
}

func parseSkillInfoUpdate2025(packet Packet) (SkillInfoUpdate, bool, error) {
	if len(packet.Data) < 2+skillData2025Size {
		return SkillInfoUpdate{}, false, fmt.Errorf("skill info 0x%04X too short: %d", packet.ID, len(packet.Data))
	}
	return SkillInfoUpdate{Skill: parseSkillData2025(packet.Data[2 : 2+skillData2025Size])}, true, nil
}

func parseSkillCastNotify2025(packet Packet) (SkillCastNotify, bool, error) {
	if len(packet.Data) < 29 {
		return SkillCastNotify{}, false, fmt.Errorf("ZC_USESKILL_ACK 0x0B1A too short: %d", len(packet.Data))
	}
	return SkillCastNotify{
		SourceID:  binary.LittleEndian.Uint32(packet.Data[2:6]),
		TargetID:  binary.LittleEndian.Uint32(packet.Data[6:10]),
		X:         binary.LittleEndian.Uint16(packet.Data[10:12]),
		Y:         binary.LittleEndian.Uint16(packet.Data[12:14]),
		SkillID:   binary.LittleEndian.Uint16(packet.Data[14:16]),
		Property:  binary.LittleEndian.Uint32(packet.Data[16:20]),
		DelayTime: binary.LittleEndian.Uint32(packet.Data[20:24]),
	}, true, nil
}

// parseHotkeyListV3 decodes ZC_SHORTCUT_KEY_LIST_V3 (0x0B20). Tab 0 is the skill
// bar (38 slots). Extra tabs parse into HotkeyList.Tab for session.Hotkeys.Extra.
func parseHotkeyListV3(packet Packet) (HotkeyList, bool, error) {
	const headerSize = 5
	if len(packet.Data) < headerSize+hotkeySlotsV3*hotkeySlotSize {
		return HotkeyList{}, false, fmt.Errorf("ZC_SHORTCUT_KEY_LIST_V3 too short: %d", len(packet.Data))
	}
	tab := binary.LittleEndian.Uint16(packet.Data[3:5])
	slots := make([]HotkeySlot, hotkeySlotsV3)
	for i := range slots {
		offset := headerSize + i*hotkeySlotSize
		slots[i] = HotkeySlot{
			Type:  packet.Data[offset],
			ID:    binary.LittleEndian.Uint32(packet.Data[offset+1 : offset+5]),
			Level: binary.LittleEndian.Uint16(packet.Data[offset+5 : offset+7]),
		}
	}
	return HotkeyList{Tab: tab, Slots: slots}, true, nil
}

// BuildWearEquipPacketForClientDate sends CZ_REQ_WEAR_EQUIP. 20250716 uses
// 0x0998 with a 32-bit location; 0x00A9 is the 2008 layout.
func BuildWearEquipPacketForClientDate(index uint16, location uint32, clientDate int) []byte {
	if !IsClientDate20250716(clientDate) {
		return BuildWearEquipPacket(index, uint16(location))
	}
	packet := make([]byte, 8)
	binary.LittleEndian.PutUint16(packet[0:2], PacketCZReqWearEquip2)
	binary.LittleEndian.PutUint16(packet[2:4], index)
	binary.LittleEndian.PutUint32(packet[4:8], location)
	return packet
}
