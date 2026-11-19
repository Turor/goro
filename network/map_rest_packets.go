package network

import (
	"encoding/binary"
	"fmt"
)

const (
	PacketZCNotifyTime      uint16 = 0x007f
	PacketZCSkillInfoDelete uint16 = 0x0441
	PacketZCQuestNotify     uint16 = 0x0446
	PacketZCMapPropertyR2   uint16 = 0x099b
	PacketZCSwapEquipInfo   uint16 = 0x0a9b
)

type MvpMark struct {
	GID     uint32
	ClassID uint16
	X       uint16
	Y       uint16
}

type MvpMinimap struct {
	Flags uint8
	Marks []MvpMark
}

type QuestNotifyEffect struct {
	NpcID  uint32
	X      int16
	Y      int16
	Effect uint16
	Color  uint16
}

type SkillDelete struct {
	SkillID uint16
}

type MapPropertyR2 struct {
	Type uint16
	Flag uint32
}

type SwapEquipSlot struct {
	Index    uint16
	Position uint32
}

type TuroranFeatures struct {
	Flags uint32
}

func ParseMvpMinimap(p Packet) (MvpMinimap, bool, error) {
	if p.ID != PacketZCMvpMinimap {
		return MvpMinimap{}, false, nil
	}
	if len(p.Data) < 7 {
		return MvpMinimap{}, false, fmt.Errorf("ZC_MVP_MINIMAP: invalid length %d", len(p.Data))
	}
	out := MvpMinimap{Flags: p.Data[4]}
	count := int(binary.LittleEndian.Uint16(p.Data[5:7]))
	body := p.Data[7:]
	const entry = 10
	for i := 0; i < count && (i+1)*entry <= len(body); i++ {
		e := body[i*entry : (i+1)*entry]
		out.Marks = append(out.Marks, MvpMark{
			GID:     binary.LittleEndian.Uint32(e[0:4]),
			ClassID: binary.LittleEndian.Uint16(e[4:6]),
			X:       binary.LittleEndian.Uint16(e[6:8]),
			Y:       binary.LittleEndian.Uint16(e[8:10]),
		})
	}
	return out, true, nil
}

func ParseQuestNotifyEffect(p Packet) (QuestNotifyEffect, bool, error) {
	if p.ID != PacketZCQuestNotify {
		return QuestNotifyEffect{}, false, nil
	}
	if len(p.Data) < 14 {
		return QuestNotifyEffect{}, false, fmt.Errorf("ZC_QUEST_NOTIFY_EFFECT: invalid length %d", len(p.Data))
	}
	return QuestNotifyEffect{
		NpcID:  binary.LittleEndian.Uint32(p.Data[2:6]),
		X:      int16(binary.LittleEndian.Uint16(p.Data[6:8])),
		Y:      int16(binary.LittleEndian.Uint16(p.Data[8:10])),
		Effect: binary.LittleEndian.Uint16(p.Data[10:12]),
		Color:  binary.LittleEndian.Uint16(p.Data[12:14]),
	}, true, nil
}

func ParseSkillInfoDelete(p Packet) (SkillDelete, bool, error) {
	if p.ID != PacketZCSkillInfoDelete {
		return SkillDelete{}, false, nil
	}
	if len(p.Data) < 4 {
		return SkillDelete{}, false, fmt.Errorf("ZC_SKILLINFO_DELETE: invalid length %d", len(p.Data))
	}
	return SkillDelete{SkillID: binary.LittleEndian.Uint16(p.Data[2:4])}, true, nil
}

func ParseMapPropertyR2(p Packet) (MapPropertyR2, bool, error) {
	if p.ID != PacketZCMapPropertyR2 {
		return MapPropertyR2{}, false, nil
	}
	if len(p.Data) < 8 {
		return MapPropertyR2{}, false, fmt.Errorf("ZC_MAPPROPERTY_R2: invalid length %d", len(p.Data))
	}
	return MapPropertyR2{
		Type: binary.LittleEndian.Uint16(p.Data[2:4]),
		Flag: binary.LittleEndian.Uint32(p.Data[4:8]),
	}, true, nil
}

func ParseSwapEquipInfo(p Packet) ([]SwapEquipSlot, bool, error) {
	if p.ID != PacketZCSwapEquipInfo {
		return nil, false, nil
	}
	if len(p.Data) < 4 {
		return nil, false, fmt.Errorf("ZC_SEND_SWAP_EQUIPITEM_INFO: invalid length %d", len(p.Data))
	}
	body := p.Data[4:]
	const entry = 6
	if len(body)%entry != 0 {
		return nil, true, nil
	}
	out := make([]SwapEquipSlot, 0, len(body)/entry)
	for i := 0; i+entry <= len(body); i += entry {
		out = append(out, SwapEquipSlot{
			Index:    binary.LittleEndian.Uint16(body[i : i+2]),
			Position: binary.LittleEndian.Uint32(body[i+2 : i+6]),
		})
	}
	return out, true, nil
}

func ParseTuroranFeatures(p Packet) (TuroranFeatures, bool, error) {
	if p.ID != 0x0EF1 {
		return TuroranFeatures{}, false, nil
	}
	if len(p.Data) < 6 {
		return TuroranFeatures{}, false, fmt.Errorf("ZC_TURORAN_FEATURES: invalid length %d", len(p.Data))
	}
	return TuroranFeatures{Flags: binary.LittleEndian.Uint32(p.Data[2:6])}, true, nil
}

func ParseNotifyTime(p Packet) (uint32, bool, error) {
	if p.ID != PacketZCNotifyTime {
		return 0, false, nil
	}
	if len(p.Data) < 6 {
		return 0, false, fmt.Errorf("ZC_NOTIFY_TIME: invalid length %d", len(p.Data))
	}
	return binary.LittleEndian.Uint32(p.Data[2:6]), true, nil
}
