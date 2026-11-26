package network

import (
	"encoding/binary"

	"github.com/kivutar/goro/glog"
)

const (
	PacketCZStorageMove   uint16 = 0x0EE2
	PacketZCStorageResult uint16 = 0x0EE3
	PacketZCStorageList   uint16 = 0x0EE4

	StorageWherePersonal byte = 0
	StorageWhereGuild    byte = 1

	StorageActionList     byte = 0
	StorageActionDeposit  byte = 1
	StorageActionWithdraw byte = 2

	StorageResultOK        byte = 0
	StorageResultNotLoaded byte = 1
	StorageResultNoItem    byte = 2
	StorageResultCannot    byte = 3
	StorageResultBusy      byte = 4
	StorageResultDenied    byte = 5

	storageListHeader  = 12
	storageEntrySize   = 28
	storageListReplace = 0x01
)

type StorageMoveResult struct {
	Result       byte
	Where        byte
	Action       byte
	Type         byte
	StorageIndex uint16
	Amount       uint16
	StorageLeft  uint16
	ItemID       uint32
}

type StorageList struct {
	Where   byte
	Replace bool
	Chunk   uint16
	Chunks  uint16
	Items   []InventoryItem
}

func BuildStorageMovePacket(where, action byte, index uint16, amount uint32) []byte {
	if amount == 0 {
		amount = 1
	}
	packet := make([]byte, 10)
	binary.LittleEndian.PutUint16(packet[0:2], PacketCZStorageMove)
	packet[2] = where
	packet[3] = action
	binary.LittleEndian.PutUint16(packet[4:6], index)
	binary.LittleEndian.PutUint32(packet[6:10], amount)
	return packet
}

func (c *Client) SendStorageMove(where, action byte, index uint16, amount uint32) error {
	packet := BuildStorageMovePacket(where, action, index, amount)
	err := c.Send(packet)
	if err != nil {
		glog.Warnf("send storage move failed opcode=0x%04X where=%d action=%d index=%d amount=%d: %v", PacketCZStorageMove, where, action, index, amount, err)
	}
	return err
}

func ParseStorageMoveResult(packet Packet) (StorageMoveResult, bool, error) {
	if packet.ID != PacketZCStorageResult {
		return StorageMoveResult{}, false, nil
	}
	if len(packet.Data) < 18 {
		return StorageMoveResult{}, false, errShort(packet)
	}
	return StorageMoveResult{
		Result:       packet.Data[2],
		Where:        packet.Data[3],
		Action:       packet.Data[4],
		Type:         packet.Data[5],
		StorageIndex: binary.LittleEndian.Uint16(packet.Data[6:8]),
		Amount:       binary.LittleEndian.Uint16(packet.Data[8:10]),
		StorageLeft:  binary.LittleEndian.Uint16(packet.Data[10:12]),
		ItemID:       binary.LittleEndian.Uint32(packet.Data[12:16]),
	}, true, nil
}

func ParseStorageList(packet Packet) (StorageList, bool, error) {
	if packet.ID != PacketZCStorageList {
		return StorageList{}, false, nil
	}
	if len(packet.Data) < storageListHeader {
		return StorageList{}, false, errShort(packet)
	}
	count := int(binary.LittleEndian.Uint16(packet.Data[10:12]))
	if storageListHeader+count*storageEntrySize > len(packet.Data) {
		return StorageList{}, false, errShort(packet)
	}
	items := make([]InventoryItem, 0, count)
	for i := 0; i < count; i++ {
		off := storageListHeader + i*storageEntrySize
		entry := packet.Data[off : off+storageEntrySize]
		var cards [4]uint32
		for card := 0; card < 4; card++ {
			cards[card] = binary.LittleEndian.Uint32(entry[12+card*4 : 16+card*4])
		}
		items = append(items, InventoryItem{
			Index:      binary.LittleEndian.Uint16(entry[0:2]),
			ItemID:     binary.LittleEndian.Uint32(entry[2:6]),
			Amount:     binary.LittleEndian.Uint16(entry[6:8]),
			Type:       entry[8],
			Identified: entry[9] != 0,
			Refine:     entry[10],
			Grade:      entry[11],
			Cards:      cards,
		})
	}
	return StorageList{
		Where:   packet.Data[4],
		Replace: packet.Data[5]&storageListReplace != 0,
		Chunk:   binary.LittleEndian.Uint16(packet.Data[6:8]),
		Chunks:  binary.LittleEndian.Uint16(packet.Data[8:10]),
		Items:   items,
	}, true, nil
}

func errShort(packet Packet) error {
	return &packetShortError{id: packet.ID, n: len(packet.Data)}
}

type packetShortError struct {
	id uint16
	n  int
}

func (e *packetShortError) Error() string {
	return "storage packet too short"
}
