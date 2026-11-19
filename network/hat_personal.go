package network

import (
	"encoding/binary"
	"fmt"
)

const (
	PacketZCHatEffect           uint16 = 0x0a3b
	PacketZCPersonalInformation uint16 = 0x097b
)

type HatEffect struct {
	GID     uint32
	Enabled bool
	IDs     []uint16
}

type PersonalInfo struct {
	Exp   int32
	Death int32
	Drop  int32
}

func ParseHatEffect(p Packet) (HatEffect, bool, error) {
	if p.ID != PacketZCHatEffect {
		return HatEffect{}, false, nil
	}
	if len(p.Data) < 9 {
		return HatEffect{}, false, fmt.Errorf("ZC_HAT_EFFECT: invalid length %d", len(p.Data))
	}
	out := HatEffect{
		GID:     binary.LittleEndian.Uint32(p.Data[4:8]),
		Enabled: p.Data[8] != 0,
	}
	body := p.Data[9:]
	for i := 0; i+2 <= len(body); i += 2 {
		out.IDs = append(out.IDs, binary.LittleEndian.Uint16(body[i:i+2]))
	}
	return out, true, nil
}

func ParsePersonalInformation(p Packet) (PersonalInfo, bool, error) {
	if p.ID != PacketZCPersonalInformation {
		return PersonalInfo{}, false, nil
	}
	if len(p.Data) < 16 {
		return PersonalInfo{}, false, fmt.Errorf("ZC_PERSONAL_INFORMATION2: invalid length %d", len(p.Data))
	}
	return PersonalInfo{
		Exp:   int32(binary.LittleEndian.Uint32(p.Data[4:8])),
		Death: int32(binary.LittleEndian.Uint32(p.Data[8:12])),
		Drop:  int32(binary.LittleEndian.Uint32(p.Data[12:16])),
	}, true, nil
}
