package network

import (
	"encoding/binary"
	"fmt"
)

const (
	PacketCZGradeSelect    uint16 = 0x0b59
	PacketZCGradeMaterials uint16 = 0x0b5a
	PacketCZGradeRequest   uint16 = 0x0b5b
	PacketCZGradeClose     uint16 = 0x0b5c
	PacketZCGradeAck       uint16 = 0x0b5d
	PacketZCGradeBroadcast uint16 = 0x0b5e

	gradeBlessingSize  = 16
	gradeMaterialSize  = 17
	gradeListHeaderLen = 34 // type2+len2+index2+chance4+blessing16+protect4+protectAmt4
)

type GradeBlessing struct {
	ID     int32
	Amount int32
	Max    int32
	Bonus  int32
}

type GradeMaterial struct {
	ItemID    int32
	Amount    int32
	Price     int32
	Downgrade bool
	Breakable bool
}

type GradeMaterialList struct {
	Index     uint16
	Chance    int32
	Blessing  GradeBlessing
	Materials []GradeMaterial
}

type GradeAck struct {
	Index  uint16
	Grade  int16
	Result int32
}

type GradeBroadcast struct {
	Name   string
	ItemID uint32
	Grade  int16
	Status int8
}

func ParseGradeMaterialList(p Packet) (GradeMaterialList, bool, error) {
	if p.ID != PacketZCGradeMaterials {
		return GradeMaterialList{}, false, nil
	}
	if len(p.Data) < gradeListHeaderLen {
		return GradeMaterialList{}, false, fmt.Errorf("ZC_GRADE_ENCHANT_MATERIAL_LIST: invalid length %d", len(p.Data))
	}
	wireLen := int(binary.LittleEndian.Uint16(p.Data[2:4]))
	if wireLen != len(p.Data) {
		return GradeMaterialList{}, false, fmt.Errorf("ZC_GRADE_ENCHANT_MATERIAL_LIST: length field %d != %d", wireLen, len(p.Data))
	}
	body := len(p.Data) - gradeListHeaderLen
	if body%gradeMaterialSize != 0 {
		return GradeMaterialList{}, false, fmt.Errorf("ZC_GRADE_ENCHANT_MATERIAL_LIST: body %d", body)
	}
	list := GradeMaterialList{
		Index:  binary.LittleEndian.Uint16(p.Data[4:6]),
		Chance: int32(binary.LittleEndian.Uint32(p.Data[6:10])),
		Blessing: GradeBlessing{
			ID:     int32(binary.LittleEndian.Uint32(p.Data[10:14])),
			Amount: int32(binary.LittleEndian.Uint32(p.Data[14:18])),
			Max:    int32(binary.LittleEndian.Uint32(p.Data[18:22])),
			Bonus:  int32(binary.LittleEndian.Uint32(p.Data[22:26])),
		},
	}
	count := body / gradeMaterialSize
	list.Materials = make([]GradeMaterial, count)
	off := gradeListHeaderLen
	for i := 0; i < count; i++ {
		list.Materials[i] = GradeMaterial{
			ItemID:    int32(binary.LittleEndian.Uint32(p.Data[off : off+4])),
			Amount:    int32(binary.LittleEndian.Uint32(p.Data[off+4 : off+8])),
			Price:     int32(binary.LittleEndian.Uint32(p.Data[off+8 : off+12])),
			Downgrade: p.Data[off+12] != 0,
			Breakable: p.Data[off+16] != 0,
		}
		off += gradeMaterialSize
	}
	return list, true, nil
}

func ParseGradeAck(p Packet) (GradeAck, bool, error) {
	if p.ID != PacketZCGradeAck {
		return GradeAck{}, false, nil
	}
	if len(p.Data) != 10 {
		return GradeAck{}, false, fmt.Errorf("ZC_GRADE_ENCHANT_ACK: invalid length %d", len(p.Data))
	}
	return GradeAck{
		Index:  binary.LittleEndian.Uint16(p.Data[2:4]),
		Grade:  int16(binary.LittleEndian.Uint16(p.Data[4:6])),
		Result: int32(binary.LittleEndian.Uint32(p.Data[6:10])),
	}, true, nil
}

func ParseGradeBroadcast(p Packet) (GradeBroadcast, bool, error) {
	if p.ID != PacketZCGradeBroadcast {
		return GradeBroadcast{}, false, nil
	}
	const nameLen = 24
	if len(p.Data) != 2+nameLen+4+2+1 {
		return GradeBroadcast{}, false, fmt.Errorf("ZC_GRADE_ENCHANT_BROADCAST: invalid length %d", len(p.Data))
	}
	return GradeBroadcast{
		Name:   decodeROFixedString(p.Data[2 : 2+nameLen]),
		ItemID: binary.LittleEndian.Uint32(p.Data[2+nameLen : 2+nameLen+4]),
		Grade:  int16(binary.LittleEndian.Uint16(p.Data[2+nameLen+4 : 2+nameLen+6])),
		Status: int8(p.Data[2+nameLen+6]),
	}, true, nil
}

func BuildGradeSelectPacket(index uint16) []byte {
	data := make([]byte, 4)
	binary.LittleEndian.PutUint16(data[0:2], PacketCZGradeSelect)
	binary.LittleEndian.PutUint16(data[2:4], index)
	return data
}

func BuildGradeRequestPacket(index uint16, materialIndex int32, bless bool, blessAmt int32) []byte {
	data := make([]byte, 14)
	binary.LittleEndian.PutUint16(data[0:2], PacketCZGradeRequest)
	binary.LittleEndian.PutUint16(data[2:4], index)
	binary.LittleEndian.PutUint32(data[4:8], uint32(materialIndex))
	if bless {
		data[8] = 1
	}
	binary.LittleEndian.PutUint32(data[9:13], uint32(blessAmt))
	return data
}

func BuildGradeClosePacket() []byte {
	data := make([]byte, 2)
	binary.LittleEndian.PutUint16(data[0:2], PacketCZGradeClose)
	return data
}

func (c *Client) SendGradeSelect(index uint16) error {
	return c.Send(BuildGradeSelectPacket(index))
}

func (c *Client) SendGradeRequest(index uint16, materialIndex int32, bless bool, blessAmt int32) error {
	return c.Send(BuildGradeRequestPacket(index, materialIndex, bless, blessAmt))
}

func (c *Client) SendGradeClose() error {
	return c.Send(BuildGradeClosePacket())
}
