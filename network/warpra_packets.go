package network

import (
	"encoding/binary"
	"fmt"
	"strings"
)

// Turoran Warpra (turoran_warpra.c) packet IDs and Hercules layouts.
const (
	PacketCZWarpraReq    uint16 = 0x0EF4
	PacketZCWarpraList   uint16 = 0x0EFE
	PacketZCWarpraResult uint16 = 0x0EFF

	WarpraActList uint16 = 0
	WarpraActWarp uint16 = 1
	WarpraActBuy  uint16 = 2

	WarpraOK      int16 = 0
	WarpraErr     int16 = 1
	WarpraLocked  int16 = 2
	WarpraNoZeny  int16 = 3
	WarpraInvalid int16 = 4
	WarpraBusy    int16 = 5
	WarpraLevel   int16 = 6

	WarpraKindTown    uint8 = 0
	WarpraKindField   uint8 = 1
	WarpraKindDungeon uint8 = 2
	WarpraKindSpecial uint8 = 3

	WarpraDefaultPrice  uint32 = 50_000_000
	WarpraCZReqLen             = 8
	WarpraZCResultLen          = 8
	WarpraListFlagMore  uint8  = 0x01
	WarpraListHeaderLen        = 17 // id(2)+len(2)+zeny(4)+price(4)+flags(1)+total(2)+count(2)
	WarpraMapLen               = 16
	WarpraNameLen              = 40
	WarpraGroupNameLen         = 24
	// kind(1)+id(2)+unlocked(1)+x(2)+y(2)+group(1)+map(16)+name(40)+group_name(24)
	WarpraEntrySize = 1 + 2 + 1 + 2 + 2 + 1 + WarpraMapLen + WarpraNameLen + WarpraGroupNameLen
)

type WarpraLoc struct {
	Kind      uint8
	ID        uint16
	Unlocked  bool
	X, Y      uint16
	Group     uint8
	Map       string
	Name      string
	GroupName string
}

type WarpraList struct {
	Zeny  uint32
	Price uint32
	Flags uint8
	Total uint16
	Locs  []WarpraLoc
}

func (l WarpraList) More() bool {
	return l.Flags&WarpraListFlagMore != 0
}

type WarpraResult struct {
	Result int16
	Kind   uint16
	ID     uint16
}

func ParseWarpraList(p Packet) (WarpraList, bool, error) {
	if p.ID != PacketZCWarpraList {
		return WarpraList{}, false, nil
	}
	if len(p.Data) < WarpraListHeaderLen {
		return WarpraList{}, false, fmt.Errorf("warpra list: invalid length %d", len(p.Data))
	}
	wireLen := int(binary.LittleEndian.Uint16(p.Data[2:4]))
	if wireLen != len(p.Data) {
		return WarpraList{}, false, fmt.Errorf("warpra list: length field %d != payload %d", wireLen, len(p.Data))
	}
	body := len(p.Data) - WarpraListHeaderLen
	if body%WarpraEntrySize != 0 {
		return WarpraList{}, false, fmt.Errorf("warpra list: body %d is not a multiple of entry size %d", body, WarpraEntrySize)
	}
	count := int(binary.LittleEndian.Uint16(p.Data[15:17]))
	if count != body/WarpraEntrySize {
		return WarpraList{}, false, fmt.Errorf("warpra list: count %d != %d entries", count, body/WarpraEntrySize)
	}
	list := WarpraList{
		Zeny:  binary.LittleEndian.Uint32(p.Data[4:8]),
		Price: binary.LittleEndian.Uint32(p.Data[8:12]),
		Flags: p.Data[12],
		Total: binary.LittleEndian.Uint16(p.Data[13:15]),
		Locs:  make([]WarpraLoc, count),
	}
	if list.Price == 0 {
		list.Price = WarpraDefaultPrice
	}
	off := WarpraListHeaderLen
	for i := range list.Locs {
		list.Locs[i] = parseWarpraEntry(p.Data[off : off+WarpraEntrySize])
		off += WarpraEntrySize
	}
	return list, true, nil
}

func parseWarpraEntry(data []byte) WarpraLoc {
	return WarpraLoc{
		Kind:      data[0],
		ID:        binary.LittleEndian.Uint16(data[1:3]),
		Unlocked:  data[3] != 0,
		X:         binary.LittleEndian.Uint16(data[4:6]),
		Y:         binary.LittleEndian.Uint16(data[6:8]),
		Group:     data[8],
		Map:       fixedString(data[9 : 9+WarpraMapLen]),
		Name:      fixedString(data[9+WarpraMapLen : 9+WarpraMapLen+WarpraNameLen]),
		GroupName: fixedString(data[9+WarpraMapLen+WarpraNameLen:]),
	}
}

func ParseWarpraResult(p Packet) (WarpraResult, bool, error) {
	if p.ID != PacketZCWarpraResult {
		return WarpraResult{}, false, nil
	}
	if len(p.Data) != WarpraZCResultLen {
		return WarpraResult{}, false, fmt.Errorf("warpra result: invalid length %d", len(p.Data))
	}
	return WarpraResult{
		Result: int16(binary.LittleEndian.Uint16(p.Data[2:4])),
		Kind:   binary.LittleEndian.Uint16(p.Data[4:6]),
		ID:     binary.LittleEndian.Uint16(p.Data[6:8]),
	}, true, nil
}

func BuildWarpraRequestPacket(action, kind, id uint16) []byte {
	data := make([]byte, WarpraCZReqLen)
	binary.LittleEndian.PutUint16(data[0:2], PacketCZWarpraReq)
	binary.LittleEndian.PutUint16(data[2:4], action)
	binary.LittleEndian.PutUint16(data[4:6], kind)
	binary.LittleEndian.PutUint16(data[6:8], id)
	return data
}

func (c *Client) SendWarpraRequest(action, kind, id uint16) error {
	return c.Send(BuildWarpraRequestPacket(action, kind, id))
}

func (c *Client) SendWarpraList() error {
	return c.SendWarpraRequest(WarpraActList, 0, 0)
}

func (c *Client) SendWarpraWarp(kind uint8, id uint16) error {
	return c.SendWarpraRequest(WarpraActWarp, uint16(kind), id)
}

func (c *Client) SendWarpraBuy(kind uint8, id uint16) error {
	return c.SendWarpraRequest(WarpraActBuy, uint16(kind), id)
}

// EncodeWarpraList writes one ZC_LIST chunk in the Hercules layout. Tests use
// it as a golden fixture; the live server is the source of catalog contents.
func EncodeWarpraList(list WarpraList) []byte {
	count := len(list.Locs)
	length := WarpraListHeaderLen + count*WarpraEntrySize
	data := make([]byte, length)
	binary.LittleEndian.PutUint16(data[0:2], PacketZCWarpraList)
	binary.LittleEndian.PutUint16(data[2:4], uint16(length))
	binary.LittleEndian.PutUint32(data[4:8], list.Zeny)
	price := list.Price
	if price == 0 {
		price = WarpraDefaultPrice
	}
	binary.LittleEndian.PutUint32(data[8:12], price)
	data[12] = list.Flags
	total := list.Total
	if total == 0 {
		total = uint16(count)
	}
	binary.LittleEndian.PutUint16(data[13:15], total)
	binary.LittleEndian.PutUint16(data[15:17], uint16(count))
	off := WarpraListHeaderLen
	for _, loc := range list.Locs {
		encodeWarpraEntry(data[off:off+WarpraEntrySize], loc)
		off += WarpraEntrySize
	}
	return data
}

func EncodeWarpraResult(result WarpraResult) []byte {
	data := make([]byte, WarpraZCResultLen)
	binary.LittleEndian.PutUint16(data[0:2], PacketZCWarpraResult)
	binary.LittleEndian.PutUint16(data[2:4], uint16(result.Result))
	binary.LittleEndian.PutUint16(data[4:6], result.Kind)
	binary.LittleEndian.PutUint16(data[6:8], result.ID)
	return data
}

func encodeWarpraEntry(dst []byte, loc WarpraLoc) {
	dst[0] = loc.Kind
	binary.LittleEndian.PutUint16(dst[1:3], loc.ID)
	if loc.Unlocked {
		dst[3] = 1
	}
	binary.LittleEndian.PutUint16(dst[4:6], loc.X)
	binary.LittleEndian.PutUint16(dst[6:8], loc.Y)
	dst[8] = loc.Group
	copy(dst[9:9+WarpraMapLen], loc.Map)
	copy(dst[9+WarpraMapLen:9+WarpraMapLen+WarpraNameLen], loc.Name)
	copy(dst[9+WarpraMapLen+WarpraNameLen:], loc.GroupName)
}

func NormalizeWarpraMapID(mapID string) string {
	id := strings.ToLower(strings.TrimSpace(mapID))
	id = strings.TrimSuffix(id, ".gat")
	id = strings.TrimSuffix(id, ".rsw")
	return id
}
