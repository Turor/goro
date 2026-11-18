package network

import (
	"encoding/binary"
	"fmt"
	"math"
	"net"
	"strings"
)

type AccountAcceptLogin struct {
	AuthCode   uint32
	AccountID  uint32
	UserLevel  uint32
	LastLogin  string
	Sex        uint8
	CharServer []CharServer
}

type AccountRefuseLogin struct {
	ErrorCode   uint8
	UnblockTime string
}

type Character struct {
	ID        uint32
	Exp       int64
	Money     int64
	Name      string
	Slot      uint8
	Level     int16
	JobLevel  int16
	Job       int16
	HP        int
	MaxHP     int
	SP        int
	MaxSP     int
	Str       uint8
	Agi       uint8
	Vit       uint8
	Int       uint8
	Dex       uint8
	Luk       uint8
	Hair      int16
	HairColor uint8
	HeadPal   int16
	BodyPal   int16
	Weapon    int16
	Shield    int16
	HeadTop   int16
	HeadMid   int16
	HeadLow   int16
	Option    uint32
}

type CharList struct {
	Layout     string
	Characters []Character
}

type ZoneServerNotify struct {
	CharID  uint32
	MapName string
	Address string
	Port    uint16
}

type MapAcceptEnter struct {
	ServerTick uint32
	X          int
	Y          int
	Dir        int
}

type CharServer struct {
	Address   string
	Port      uint16
	Name      string
	UserCount uint16
	State     uint16
	Property  uint16
}

// PacketACAcceptLogin2 is AC_ACCEPT_LOGIN for PACKETVER >= 20170315: a
// 64-byte header (auth token added) and 160-byte server rows.
const PacketACAcceptLogin2 uint16 = 0x0AC4

// AC_REFUSE_LOGIN_R2 / R3 carry a 32-bit error code (PACKETVER >= 20101123,
// R3 from 20180627).
const (
	PacketACRefuseLoginR2 uint16 = 0x083E
	PacketACRefuseLoginR3 uint16 = 0x0B02
)

// IsAccountAcceptLogin reports whether a packet is any AC_ACCEPT_LOGIN form.
func IsAccountAcceptLogin(id uint16) bool {
	return id == 0x0069 || id == PacketACAcceptLogin2
}

// IsAccountRefuseLogin reports whether a packet is any AC_REFUSE_LOGIN form.
func IsAccountRefuseLogin(id uint16) bool {
	return id == 0x006A || id == PacketACRefuseLoginR2 || id == PacketACRefuseLoginR3
}

func ParseAccountAcceptLogin(packet Packet) (AccountAcceptLogin, error) {
	header, serverSize := 47, 32
	switch packet.ID {
	case 0x0069:
	case PacketACAcceptLogin2:
		header, serverSize = 64, 160
	default:
		return AccountAcceptLogin{}, fmt.Errorf("unexpected packet 0x%04X", packet.ID)
	}
	if len(packet.Data) < header {
		return AccountAcceptLogin{}, fmt.Errorf("AC_ACCEPT_LOGIN 0x%04X too short: %d", packet.ID, len(packet.Data))
	}

	packetLen := int(binary.LittleEndian.Uint16(packet.Data[2:4]))
	if packetLen > len(packet.Data) {
		return AccountAcceptLogin{}, fmt.Errorf("AC_ACCEPT_LOGIN length %d exceeds data %d", packetLen, len(packet.Data))
	}

	sex := packet.Data[46]
	if sex >= 10 {
		sex -= 10
	}

	out := AccountAcceptLogin{
		AuthCode:  binary.LittleEndian.Uint32(packet.Data[4:8]),
		AccountID: binary.LittleEndian.Uint32(packet.Data[8:12]),
		UserLevel: binary.LittleEndian.Uint32(packet.Data[12:16]),
		LastLogin: fixedString(packet.Data[20:46]),
		Sex:       sex,
	}

	serverBytes := packetLen - header
	if serverBytes < 0 {
		serverBytes = 0
	}
	serverCount := serverBytes / serverSize
	for i := 0; i < serverCount; i++ {
		base := header + i*serverSize
		raw := packet.Data[base : base+32]
		ip := net.IPv4(raw[0], raw[1], raw[2], raw[3])
		out.CharServer = append(out.CharServer, CharServer{
			Address:   ip.String(),
			Port:      binary.LittleEndian.Uint16(raw[4:6]),
			Name:      fixedString(raw[6:26]),
			UserCount: binary.LittleEndian.Uint16(raw[26:28]),
			State:     binary.LittleEndian.Uint16(raw[28:30]),
			Property:  binary.LittleEndian.Uint16(raw[30:32]),
		})
	}

	return out, nil
}

func ParseAccountRefuseLogin(packet Packet) (AccountRefuseLogin, error) {
	if packet.ID == PacketACRefuseLoginR2 || packet.ID == PacketACRefuseLoginR3 {
		if len(packet.Data) < 26 {
			return AccountRefuseLogin{}, fmt.Errorf("AC_REFUSE_LOGIN 0x%04X too short: %d", packet.ID, len(packet.Data))
		}
		code := binary.LittleEndian.Uint32(packet.Data[2:6])
		if code > 0xFF {
			code = 0xFF
		}
		return AccountRefuseLogin{
			ErrorCode:   uint8(code),
			UnblockTime: fixedString(packet.Data[6:26]),
		}, nil
	}
	if packet.ID != 0x006A {
		return AccountRefuseLogin{}, fmt.Errorf("unexpected packet 0x%04X", packet.ID)
	}
	if len(packet.Data) < 23 {
		return AccountRefuseLogin{}, fmt.Errorf("AC_REFUSE_LOGIN too short: %d", len(packet.Data))
	}

	return AccountRefuseLogin{
		ErrorCode:   packet.Data[2],
		UnblockTime: fixedString(packet.Data[3:23]),
	}, nil
}

// PacketHCAckCharInfoPerPage is HC_ACK_CHARINFO_PER_PAGE for PACKETVER
// >= 20201007 main: a 4-byte header and 175-byte CHARACTER_INFO rows. Hercules
// sends an empty one as the paging terminator.
const PacketHCAckCharInfoPerPage uint16 = 0x0B72

// PacketHCAcceptMakeChar2025 is HC_ACCEPT_MAKECHAR with a 175-byte
// CHARACTER_INFO (PACKETVER >= 20201007 main).
const PacketHCAcceptMakeChar2025 uint16 = 0x0B6F

// PacketHCNotifyZoneServer2 is HC_NOTIFY_ZONESVR with a 128-byte domain
// (PACKETVER >= 20170315).
const PacketHCNotifyZoneServer2 uint16 = 0x0AC5

// PacketZCAcceptEnter2 is ZC_ACCEPT_ENTER with font and sex (PACKETVER
// >= 20141022).
const PacketZCAcceptEnter2 uint16 = 0x0A18

// characterInfo2025Size is CHARACTER_INFO with 64-bit EXP, 64-bit HP/SP,
// body, map name, robe, slot change, rename and sex (PACKETVER >= 20201007
// main; Hercules char_mmo_char_tobuf, MAX_CHAR_BUF).
const characterInfo2025Size = 175

// IsCharList reports whether a packet carries a character list.
func IsCharList(id uint16) bool {
	return id == 0x006B || id == PacketHCAckCharInfoPerPage
}

// IsZoneServerNotify reports whether a packet is any HC_NOTIFY_ZONESVR form.
func IsZoneServerNotify(id uint16) bool {
	return id == 0x0071 || id == PacketHCNotifyZoneServer2
}

// IsMapAcceptEnter reports whether a packet is any ZC_ACCEPT_ENTER form.
func IsMapAcceptEnter(id uint16) bool {
	return id == 0x0073 || id == 0x02EB || id == PacketZCAcceptEnter2
}

// IsMakeCharacterAccept reports whether a packet is any HC_ACCEPT_MAKECHAR form.
func IsMakeCharacterAccept(id uint16) bool {
	return id == 0x006D || id == PacketHCAcceptMakeChar2025
}

func ParseCharList(packet Packet) (CharList, error) {
	if !IsCharList(packet.ID) {
		return CharList{}, fmt.Errorf("unexpected packet 0x%04X", packet.ID)
	}
	if len(packet.Data) < 4 {
		return CharList{}, fmt.Errorf("HC_ACCEPT_ENTER too short: %d", len(packet.Data))
	}
	packetLen := int(binary.LittleEndian.Uint16(packet.Data[2:4]))
	if packetLen > len(packet.Data) {
		return CharList{}, fmt.Errorf("HC_ACCEPT_ENTER length %d exceeds data %d", packetLen, len(packet.Data))
	}

	var (
		layout                 string
		headerSize, recordSize int
		err                    error
	)
	if packet.ID == PacketHCAckCharInfoPerPage {
		layout, headerSize, recordSize = "per_page_4_175", 4, characterInfo2025Size
		if (packetLen-headerSize)%recordSize != 0 {
			return CharList{}, fmt.Errorf("unsupported HC_ACK_CHARINFO_PER_PAGE length=%d", packetLen)
		}
	} else {
		layout, headerSize, recordSize, err = resolveCharListLayout(packetLen)
		if err != nil {
			return CharList{}, err
		}
	}

	out := CharList{Layout: layout}
	for offset := headerSize; offset+recordSize <= packetLen; offset += recordSize {
		var character Character
		switch recordSize {
		case 106:
			character = parseLegacyCharacter106(packet.Data[offset : offset+recordSize])
		case characterInfo2025Size:
			character = parseCharacter175(packet.Data[offset : offset+recordSize])
		default:
			character = parseCharacter108(packet.Data[offset : offset+recordSize])
		}
		out.Characters = append(out.Characters, character)
	}
	return out, nil
}

func ParseMakeCharacterAccept(packet Packet) (Character, error) {
	if packet.ID == PacketHCAcceptMakeChar2025 {
		if len(packet.Data) < 2+characterInfo2025Size {
			return Character{}, fmt.Errorf("HC_ACCEPT_MAKECHAR 0x0B6F too short: %d", len(packet.Data))
		}
		return parseCharacter175(packet.Data[2 : 2+characterInfo2025Size]), nil
	}
	if packet.ID != 0x006D {
		return Character{}, fmt.Errorf("unexpected packet 0x%04X", packet.ID)
	}
	if len(packet.Data) < 110 {
		return Character{}, fmt.Errorf("HC_ACCEPT_MAKECHAR too short: %d", len(packet.Data))
	}
	return parseCharacter108(packet.Data[2:110]), nil
}

func ParseMakeCharacterRefuse(packet Packet) (uint8, error) {
	if packet.ID != 0x006E {
		return 0, fmt.Errorf("unexpected packet 0x%04X", packet.ID)
	}
	if len(packet.Data) < 3 {
		return 0, fmt.Errorf("HC_REFUSE_MAKECHAR too short: %d", len(packet.Data))
	}
	return packet.Data[2], nil
}

func ParseDeleteCharacterAccept(packet Packet) error {
	if packet.ID != 0x006F {
		return fmt.Errorf("unexpected packet 0x%04X", packet.ID)
	}
	if len(packet.Data) < 2 {
		return fmt.Errorf("HC_ACCEPT_DELETECHAR too short: %d", len(packet.Data))
	}
	return nil
}

func ParseDeleteCharacterRefuse(packet Packet) (uint8, error) {
	if packet.ID != 0x0070 {
		return 0, fmt.Errorf("unexpected packet 0x%04X", packet.ID)
	}
	if len(packet.Data) < 3 {
		return 0, fmt.Errorf("HC_REFUSE_DELETECHAR too short: %d", len(packet.Data))
	}
	return packet.Data[2], nil
}

func ParseZoneServerNotify(packet Packet) (ZoneServerNotify, error) {
	if !IsZoneServerNotify(packet.ID) {
		return ZoneServerNotify{}, fmt.Errorf("unexpected packet 0x%04X", packet.ID)
	}
	if len(packet.Data) < 28 {
		return ZoneServerNotify{}, fmt.Errorf("HC_NOTIFY_ZONESVR too short: %d", len(packet.Data))
	}

	mapName := normalizeMapName(fixedString(packet.Data[6:22]))

	return ZoneServerNotify{
		CharID:  binary.LittleEndian.Uint32(packet.Data[2:6]),
		MapName: mapName,
		Address: net.IPv4(packet.Data[22], packet.Data[23], packet.Data[24], packet.Data[25]).String(),
		Port:    binary.LittleEndian.Uint16(packet.Data[26:28]),
	}, nil
}

func ParseMapAcceptEnter(packet Packet) (MapAcceptEnter, error) {
	if !IsMapAcceptEnter(packet.ID) {
		return MapAcceptEnter{}, fmt.Errorf("unexpected packet 0x%04X", packet.ID)
	}
	if len(packet.Data) < 9 {
		return MapAcceptEnter{}, fmt.Errorf("ZC_ACCEPT_ENTER too short: %d", len(packet.Data))
	}

	return MapAcceptEnter{
		ServerTick: binary.LittleEndian.Uint32(packet.Data[2:6]),
		X:          (int(packet.Data[6]) << 2) | int(packet.Data[7]>>6),
		Y:          (int(packet.Data[7]&0x3f) << 4) | int(packet.Data[8]>>4),
		Dir:        int(packet.Data[8] & 0x0f),
	}, nil
}

func resolveCharListLayout(packetLen int) (layout string, headerSize int, recordSize int, err error) {
	type candidate struct {
		name   string
		header int
		record int
	}
	for _, c := range []candidate{
		{name: "expanded_27_175", header: 27, record: characterInfo2025Size},
		{name: "expanded_27_108", header: 27, record: 108},
		{name: "legacy_24_108", header: 24, record: 108},
		{name: "legacy_4_106", header: 4, record: 106},
	} {
		if packetLen >= c.header && (packetLen-c.header)%c.record == 0 {
			return c.name, c.header, c.record, nil
		}
	}
	return "", 0, 0, fmt.Errorf("unsupported HC_ACCEPT_ENTER layout length=%d", packetLen)
}

func parseCharacter108(data []byte) Character {
	return Character{
		ID:        binary.LittleEndian.Uint32(data[0:4]),
		Exp:       int64(int32(binary.LittleEndian.Uint32(data[4:8]))),
		Money:     int64(int32(binary.LittleEndian.Uint32(data[8:12]))),
		JobLevel:  int16(binary.LittleEndian.Uint32(data[16:20])),
		HP:        int(int16(binary.LittleEndian.Uint16(data[42:44]))),
		MaxHP:     int(int16(binary.LittleEndian.Uint16(data[44:46]))),
		SP:        int(int16(binary.LittleEndian.Uint16(data[46:48]))),
		MaxSP:     int(int16(binary.LittleEndian.Uint16(data[48:50]))),
		Option:    binary.LittleEndian.Uint32(data[28:32]),
		Job:       int16(binary.LittleEndian.Uint16(data[52:54])),
		Hair:      int16(binary.LittleEndian.Uint16(data[54:56])),
		Weapon:    int16(binary.LittleEndian.Uint16(data[56:58])),
		Level:     int16(binary.LittleEndian.Uint16(data[58:60])),
		HeadLow:   int16(binary.LittleEndian.Uint16(data[62:64])),
		Shield:    int16(binary.LittleEndian.Uint16(data[64:66])),
		HeadTop:   int16(binary.LittleEndian.Uint16(data[66:68])),
		HeadMid:   int16(binary.LittleEndian.Uint16(data[68:70])),
		Name:      fixedString(data[74:98]),
		Str:       data[98],
		Agi:       data[99],
		Vit:       data[100],
		Int:       data[101],
		Dex:       data[102],
		Luk:       data[103],
		Slot:      data[104],
		HairColor: data[105],
		HeadPal:   int16(binary.LittleEndian.Uint16(data[70:72])),
		BodyPal:   int16(binary.LittleEndian.Uint16(data[72:74])),
	}
}

// parseCharacter175 decodes the PACKETVER 20250716 CHARACTER_INFO.
func parseCharacter175(data []byte) Character {
	u16 := func(off int) int16 { return int16(binary.LittleEndian.Uint16(data[off : off+2])) }
	clampInt := func(v int64) int {
		if v > math.MaxInt32 {
			return math.MaxInt32
		}
		if v < math.MinInt32 {
			return math.MinInt32
		}
		return int(v)
	}
	i64 := func(off int) int64 { return int64(binary.LittleEndian.Uint64(data[off : off+8])) }
	return Character{
		ID:        binary.LittleEndian.Uint32(data[0:4]),
		Exp:       i64(4),
		Money:     int64(int32(binary.LittleEndian.Uint32(data[12:16]))),
		JobLevel:  int16(binary.LittleEndian.Uint32(data[24:28])),
		Option:    binary.LittleEndian.Uint32(data[36:40]),
		HP:        clampInt(i64(50)),
		MaxHP:     clampInt(i64(58)),
		SP:        clampInt(i64(66)),
		MaxSP:     clampInt(i64(74)),
		Job:       u16(84),
		Hair:      u16(86),
		Weapon:    u16(90),
		Level:     u16(92),
		HeadLow:   u16(96),
		Shield:    u16(98),
		HeadTop:   u16(100),
		HeadMid:   u16(102),
		HairColor: uint8(binary.LittleEndian.Uint16(data[104:106])),
		HeadPal:   u16(104),
		BodyPal:   u16(106),
		Name:      fixedString(data[108:132]),
		Str:       data[132],
		Agi:       data[133],
		Vit:       data[134],
		Int:       data[135],
		Dex:       data[136],
		Luk:       data[137],
		Slot:      uint8(binary.LittleEndian.Uint16(data[138:140])),
	}
}

func parseLegacyCharacter106(data []byte) Character {
	character := Character{
		ID:       binary.LittleEndian.Uint32(data[0:4]),
		Exp:      int64(int32(binary.LittleEndian.Uint32(data[4:8]))),
		Money:    int64(int32(binary.LittleEndian.Uint32(data[8:12]))),
		JobLevel: int16(binary.LittleEndian.Uint32(data[16:20])),
		HP:       int(int16(binary.LittleEndian.Uint16(data[42:44]))),
		MaxHP:    int(int16(binary.LittleEndian.Uint16(data[44:46]))),
		SP:       int(int16(binary.LittleEndian.Uint16(data[46:48]))),
		MaxSP:    int(int16(binary.LittleEndian.Uint16(data[48:50]))),
		Option:   binary.LittleEndian.Uint32(data[28:32]),
		Job:      int16(binary.LittleEndian.Uint16(data[52:54])),
		Hair:     int16(binary.LittleEndian.Uint16(data[54:56])),
		Weapon:   int16(binary.LittleEndian.Uint16(data[56:58])),
		Level:    int16(binary.LittleEndian.Uint16(data[58:60])),
		HeadLow:  int16(binary.LittleEndian.Uint16(data[62:64])),
		Shield:   int16(binary.LittleEndian.Uint16(data[64:66])),
		HeadTop:  int16(binary.LittleEndian.Uint16(data[66:68])),
		HeadMid:  int16(binary.LittleEndian.Uint16(data[68:70])),
		Name:     fixedString(data[74:98]),
		Str:      data[98],
		Agi:      data[99],
		Vit:      data[100],
		Int:      data[101],
		Dex:      data[102],
		Luk:      data[103],
	}
	if len(data) >= 106 {
		character.Slot = uint8(binary.LittleEndian.Uint16(data[104:106]))
	}
	character.HeadPal = int16(binary.LittleEndian.Uint16(data[70:72]))
	character.BodyPal = int16(binary.LittleEndian.Uint16(data[72:74]))
	if character.HairColor == 0 && character.HeadPal > 0 && character.HeadPal <= 255 {
		character.HairColor = uint8(character.HeadPal)
	}
	return character
}

func fixedString(data []byte) string {
	if index := strings.IndexByte(string(data), 0); index >= 0 {
		data = data[:index]
	}
	return strings.TrimSpace(string(data))
}
