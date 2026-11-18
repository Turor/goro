package network

import (
	"encoding/binary"
	"fmt"

	"github.com/kivutar/goro/glog"
)

// Turoran account pet stable + per-character homunculus stable
// (hercules-prere-plus src/plugins/turoran_stable.c).
//
// Client-to-server: 0x0EF0 open (len 2), 0x0EF1 homunculus swap (len 6),
// 0x0EF2 pet action (len 8). Incoming 0x0EF1 is ZC_TURORAN_FEATURES
// (item marks), not a stable packet; both directions are 6 bytes.
const (
	PacketCZTuroranStableOpen   uint16 = 0x0EF0
	PacketCZTuroranStableHomun  uint16 = 0x0EF1
	PacketCZTuroranStablePet    uint16 = 0x0EF2
	PacketZCTuroranStableOwned  uint16 = 0x0EF6
	PacketZCTuroranStableHomun  uint16 = 0x0EF8
	PacketZCTuroranStablePet    uint16 = 0x0EF9
	PacketZCTuroranStableResult uint16 = 0x0EFA
)

const (
	StableHomunculusHeaderLen = 9
	StableHomunculusEntryLen  = 42
	StablePetHeaderLen        = 9
	StablePetEntryLen         = 191
	StableOwnedHeaderLen      = 6
	StableOwnedEntryLen       = 8
	StableNameLen             = 24
	StablePetBuffLen          = 96
	StableResultLen           = 8
)

const (
	StablePetStore      uint16 = 0
	StablePetSummon     uint16 = 1
	StablePetDepositEgg uint16 = 2
	StablePetToEgg      uint16 = 3
	StablePetFeed       uint16 = 4
	StablePetEquip      uint16 = 5
	StablePetUnequip    uint16 = 6
	StablePetEvolve     uint16 = 7
)

const (
	StableOK           int16 = 0
	StableErr          int16 = 1
	StableErrNoHomun   int16 = 2
	StableErrType      int16 = 3
	StableErrInUse     int16 = 4
	StableErrInventory int16 = 5
	StableErrNotYours  int16 = 6
	StableErrNoPet     int16 = 7
	StableErrHavePet   int16 = 8
	StableErrDead      int16 = 9
	StableErrNotActive int16 = 10
	StableErrNoFood    int16 = 11
	StableErrNoAcce    int16 = 12
	StableErrEvolve    int16 = 13
)

const (
	StableFlagInStable uint8 = 0x01
	StableFlagOut      uint8 = 0x02
	StableFlagInvEgg   uint8 = 0x04
)

type StableHomunculusEntry struct {
	HomunID    uint32
	ClassID    uint16
	PrevClass  uint16
	Level      uint16
	Intimacy   uint32
	Hunger     uint16
	Vaporize   uint8
	RenameFlag uint8
	Name       string
}

type StableHomunculusList struct {
	ActiveHomunID uint32
	List          []StableHomunculusEntry
}

type StablePetEntry struct {
	PetID      uint32
	ClassID    uint32
	Level      uint16
	Intimate   uint16
	Hungry     uint16
	EggID      uint32
	Equip      uint32
	CharID     uint32
	Incubate   uint8
	RenameFlag uint8
	Flags      uint8
	Name       string
	OwnerName  string
	Atk        uint16
	Atk2       uint16
	ASPD       uint16
	Exp        uint32
	NextExp    uint32
	FoodID     uint32
	Buff       string
}

type StablePetList struct {
	ActivePetID uint32
	List        []StablePetEntry
}

type StableOwnedEntry struct {
	ItemID uint32
	Owned  uint32
}

type StableOwnedList struct {
	List []StableOwnedEntry
}

type StableResult struct {
	Result int16
	ID     uint32
}

func ParseStableHomunculusList(packet Packet) (StableHomunculusList, bool, error) {
	if packet.ID != PacketZCTuroranStableHomun {
		return StableHomunculusList{}, false, nil
	}
	if len(packet.Data) < StableHomunculusHeaderLen {
		return StableHomunculusList{}, false, fmt.Errorf("ZC_TURORAN_STABLE_HOMUN too short: %d", len(packet.Data))
	}
	packetLen := packetBodyLen(packet.Data)
	body := packet.Data[StableHomunculusHeaderLen:packetLen]
	count := int(packet.Data[8])
	if count < 0 || count*StableHomunculusEntryLen > len(body) {
		return StableHomunculusList{}, false, fmt.Errorf("ZC_TURORAN_STABLE_HOMUN truncated: count=%d body=%d", count, len(body))
	}
	list := StableHomunculusList{
		ActiveHomunID: binary.LittleEndian.Uint32(packet.Data[4:8]),
		List:          make([]StableHomunculusEntry, 0, count),
	}
	for i := 0; i < count; i++ {
		entry := body[i*StableHomunculusEntryLen : (i+1)*StableHomunculusEntryLen]
		list.List = append(list.List, StableHomunculusEntry{
			HomunID:    binary.LittleEndian.Uint32(entry[0:4]),
			ClassID:    binary.LittleEndian.Uint16(entry[4:6]),
			PrevClass:  binary.LittleEndian.Uint16(entry[6:8]),
			Level:      binary.LittleEndian.Uint16(entry[8:10]),
			Intimacy:   binary.LittleEndian.Uint32(entry[10:14]),
			Hunger:     binary.LittleEndian.Uint16(entry[14:16]),
			Vaporize:   entry[16],
			RenameFlag: entry[17],
			Name:       decodeROFixedString(entry[18:42]),
		})
	}
	return list, true, nil
}

func ParseStablePetList(packet Packet) (StablePetList, bool, error) {
	if packet.ID != PacketZCTuroranStablePet {
		return StablePetList{}, false, nil
	}
	if len(packet.Data) < StablePetHeaderLen {
		return StablePetList{}, false, fmt.Errorf("ZC_TURORAN_STABLE_PET too short: %d", len(packet.Data))
	}
	packetLen := packetBodyLen(packet.Data)
	body := packet.Data[StablePetHeaderLen:packetLen]
	count := int(packet.Data[8])
	if count < 0 || count*StablePetEntryLen > len(body) {
		return StablePetList{}, false, fmt.Errorf("ZC_TURORAN_STABLE_PET truncated: count=%d body=%d", count, len(body))
	}
	list := StablePetList{
		ActivePetID: binary.LittleEndian.Uint32(packet.Data[4:8]),
		List:        make([]StablePetEntry, 0, count),
	}
	for i := 0; i < count; i++ {
		entry := body[i*StablePetEntryLen : (i+1)*StablePetEntryLen]
		list.List = append(list.List, StablePetEntry{
			PetID:      binary.LittleEndian.Uint32(entry[0:4]),
			ClassID:    binary.LittleEndian.Uint32(entry[4:8]),
			Level:      binary.LittleEndian.Uint16(entry[8:10]),
			Intimate:   binary.LittleEndian.Uint16(entry[10:12]),
			Hungry:     binary.LittleEndian.Uint16(entry[12:14]),
			EggID:      binary.LittleEndian.Uint32(entry[14:18]),
			Equip:      binary.LittleEndian.Uint32(entry[18:22]),
			CharID:     binary.LittleEndian.Uint32(entry[22:26]),
			Incubate:   entry[26],
			RenameFlag: entry[27],
			Flags:      entry[28],
			Name:       decodeROFixedString(entry[29:53]),
			OwnerName:  decodeROFixedString(entry[53:77]),
			Atk:        binary.LittleEndian.Uint16(entry[77:79]),
			Atk2:       binary.LittleEndian.Uint16(entry[79:81]),
			ASPD:       binary.LittleEndian.Uint16(entry[81:83]),
			Exp:        binary.LittleEndian.Uint32(entry[83:87]),
			NextExp:    binary.LittleEndian.Uint32(entry[87:91]),
			FoodID:     binary.LittleEndian.Uint32(entry[91:95]),
			Buff:       decodeROFixedString(entry[95:191]),
		})
	}
	return list, true, nil
}

func ParseStableOwnedList(packet Packet) (StableOwnedList, bool, error) {
	if packet.ID != PacketZCTuroranStableOwned {
		return StableOwnedList{}, false, nil
	}
	if len(packet.Data) < StableOwnedHeaderLen {
		return StableOwnedList{}, false, fmt.Errorf("ZC_TURORAN_STABLE_OWNED too short: %d", len(packet.Data))
	}
	packetLen := packetBodyLen(packet.Data)
	body := packet.Data[StableOwnedHeaderLen:packetLen]
	count := int(binary.LittleEndian.Uint16(packet.Data[4:6]))
	if count < 0 || count*StableOwnedEntryLen > len(body) {
		return StableOwnedList{}, false, fmt.Errorf("ZC_TURORAN_STABLE_OWNED truncated: count=%d body=%d", count, len(body))
	}
	list := StableOwnedList{List: make([]StableOwnedEntry, 0, count)}
	for i := 0; i < count; i++ {
		entry := body[i*StableOwnedEntryLen : (i+1)*StableOwnedEntryLen]
		list.List = append(list.List, StableOwnedEntry{
			ItemID: binary.LittleEndian.Uint32(entry[0:4]),
			Owned:  binary.LittleEndian.Uint32(entry[4:8]),
		})
	}
	return list, true, nil
}

func ParseStableResult(packet Packet) (StableResult, bool, error) {
	if packet.ID != PacketZCTuroranStableResult {
		return StableResult{}, false, nil
	}
	if len(packet.Data) < StableResultLen {
		return StableResult{}, false, fmt.Errorf("ZC_TURORAN_STABLE_RESULT too short: %d", len(packet.Data))
	}
	return StableResult{
		Result: int16(binary.LittleEndian.Uint16(packet.Data[2:4])),
		ID:     binary.LittleEndian.Uint32(packet.Data[4:8]),
	}, true, nil
}

func packetBodyLen(data []byte) int {
	if len(data) < 4 {
		return len(data)
	}
	packetLen := int(binary.LittleEndian.Uint16(data[2:4]))
	if packetLen <= 0 || packetLen > len(data) {
		return len(data)
	}
	return packetLen
}

func BuildStableOpenPacket() []byte {
	var w Writer
	w.Uint16(PacketCZTuroranStableOpen)
	return w.Bytes()
}

func BuildStableHomunculusPacket(homunID uint32) []byte {
	var w Writer
	w.Uint16(PacketCZTuroranStableHomun)
	w.Uint32(homunID)
	return w.Bytes()
}

func BuildStablePetPacket(action uint16, id uint32) []byte {
	var w Writer
	w.Uint16(PacketCZTuroranStablePet)
	w.Uint16(action)
	w.Uint32(id)
	return w.Bytes()
}

func (c *Client) SendStableOpen() error {
	packet := BuildStableOpenPacket()
	err := c.Send(packet)
	if err == nil {
		glog.Debugf("sent CZ_TURORAN_STABLE_OPEN opcode=0x%04X client_date=%d", ID(packet), c.clientDate)
	} else {
		glog.Warnf("send CZ_TURORAN_STABLE_OPEN failed opcode=0x%04X len=%d client_date=%d: %v", ID(packet), len(packet), c.clientDate, err)
	}
	return err
}

func (c *Client) SendStableHomunculus(homunID uint32) error {
	packet := BuildStableHomunculusPacket(homunID)
	err := c.Send(packet)
	if err == nil {
		glog.Debugf("sent CZ_TURORAN_STABLE_HOMUN opcode=0x%04X homun=%d client_date=%d", ID(packet), homunID, c.clientDate)
	} else {
		glog.Warnf("send CZ_TURORAN_STABLE_HOMUN failed opcode=0x%04X len=%d homun=%d client_date=%d: %v", ID(packet), len(packet), homunID, c.clientDate, err)
	}
	return err
}

func (c *Client) SendStablePet(action uint16, id uint32) error {
	packet := BuildStablePetPacket(action, id)
	err := c.Send(packet)
	if err == nil {
		glog.Debugf("sent CZ_TURORAN_STABLE_PET opcode=0x%04X action=%d id=%d client_date=%d", ID(packet), action, id, c.clientDate)
	} else {
		glog.Warnf("send CZ_TURORAN_STABLE_PET failed opcode=0x%04X len=%d action=%d id=%d client_date=%d: %v", ID(packet), len(packet), action, id, c.clientDate, err)
	}
	return err
}
