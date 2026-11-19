package network

import (
	"encoding/binary"
	"fmt"
)

const (
	PacketZCOpenRefiningUI     uint16 = 0x0aa0
	PacketCZRefiningSelectItem uint16 = 0x0aa1
	PacketZCRefiningMaterials  uint16 = 0x0aa2
	PacketCZReqRefining        uint16 = 0x0aa3
	PacketCZCloseRefiningUI    uint16 = 0x0aa4
	PacketZCBroadcastRefine    uint16 = 0x0ada

	refineMaterialEntrySize = 9 // itemId u32 + chance i8 + zeny i32
	refineMaterialHeaderLen = 7 // type2 + len2 + itemIndex2 + blessing1
)

type RefineMaterial struct {
	ItemID uint32
	Chance int8
	Zeny   int32
}

type RefineMaterialList struct {
	Index     uint16
	Blessing  bool
	Materials []RefineMaterial
}

type RefineBroadcast struct {
	Name        string
	ItemID      uint32
	RefineLevel int8
	Status      int8
}

func ParseOpenRefiningUI(p Packet) (bool, error) {
	if p.ID != PacketZCOpenRefiningUI {
		return false, nil
	}
	if len(p.Data) != 2 {
		return false, fmt.Errorf("ZC_OPEN_REFINING_UI: invalid length %d", len(p.Data))
	}
	return true, nil
}

func ParseRefineMaterialList(p Packet) (RefineMaterialList, bool, error) {
	if p.ID != PacketZCRefiningMaterials {
		return RefineMaterialList{}, false, nil
	}
	if len(p.Data) < refineMaterialHeaderLen {
		return RefineMaterialList{}, false, fmt.Errorf("ZC_REFINING_MATERIAL_LIST: invalid length %d", len(p.Data))
	}
	wireLen := int(binary.LittleEndian.Uint16(p.Data[2:4]))
	if wireLen != len(p.Data) {
		return RefineMaterialList{}, false, fmt.Errorf("ZC_REFINING_MATERIAL_LIST: length field %d != %d", wireLen, len(p.Data))
	}
	body := len(p.Data) - refineMaterialHeaderLen
	if body%refineMaterialEntrySize != 0 {
		return RefineMaterialList{}, false, fmt.Errorf("ZC_REFINING_MATERIAL_LIST: body %d not a multiple of %d", body, refineMaterialEntrySize)
	}
	count := body / refineMaterialEntrySize
	list := RefineMaterialList{
		Index:     binary.LittleEndian.Uint16(p.Data[4:6]),
		Blessing:  p.Data[6] != 0,
		Materials: make([]RefineMaterial, count),
	}
	off := refineMaterialHeaderLen
	for i := 0; i < count; i++ {
		list.Materials[i] = RefineMaterial{
			ItemID: binary.LittleEndian.Uint32(p.Data[off : off+4]),
			Chance: int8(p.Data[off+4]),
			Zeny:   int32(binary.LittleEndian.Uint32(p.Data[off+5 : off+9])),
		}
		off += refineMaterialEntrySize
	}
	return list, true, nil
}

func ParseRefineBroadcast(p Packet) (RefineBroadcast, bool, error) {
	if p.ID != PacketZCBroadcastRefine {
		return RefineBroadcast{}, false, nil
	}
	const nameLen = 24
	if len(p.Data) != 2+nameLen+4+1+1 {
		return RefineBroadcast{}, false, fmt.Errorf("ZC_BROADCAST_ITEMREFINING_RESULT: invalid length %d", len(p.Data))
	}
	return RefineBroadcast{
		Name:        decodeROFixedString(p.Data[2 : 2+nameLen]),
		ItemID:      binary.LittleEndian.Uint32(p.Data[2+nameLen : 2+nameLen+4]),
		RefineLevel: int8(p.Data[2+nameLen+4]),
		Status:      int8(p.Data[2+nameLen+5]),
	}, true, nil
}

func BuildRefiningSelectItemPacket(index uint16) []byte {
	data := make([]byte, 4)
	binary.LittleEndian.PutUint16(data[0:2], PacketCZRefiningSelectItem)
	binary.LittleEndian.PutUint16(data[2:4], index)
	return data
}

func BuildReqRefiningPacket(index uint16, itemID uint32, blessing bool) []byte {
	data := make([]byte, 9)
	binary.LittleEndian.PutUint16(data[0:2], PacketCZReqRefining)
	binary.LittleEndian.PutUint16(data[2:4], index)
	binary.LittleEndian.PutUint32(data[4:8], itemID)
	if blessing {
		data[8] = 1
	}
	return data
}

func BuildCloseRefiningUIPacket() []byte {
	data := make([]byte, 2)
	binary.LittleEndian.PutUint16(data[0:2], PacketCZCloseRefiningUI)
	return data
}

func (c *Client) SendRefiningSelectItem(index uint16) error {
	return c.Send(BuildRefiningSelectItemPacket(index))
}

func (c *Client) SendReqRefining(index uint16, itemID uint32, blessing bool) error {
	return c.Send(BuildReqRefiningPacket(index, itemID, blessing))
}

func (c *Client) SendCloseRefiningUI() error {
	return c.Send(BuildCloseRefiningUIPacket())
}
