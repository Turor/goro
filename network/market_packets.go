package network

import (
	"encoding/binary"
	"fmt"
)

const (
	PacketCZOpenRoulette  uint16 = 0x0a19
	PacketZCOpenRoulette  uint16 = 0x0a1a
	PacketCZRouletteInfo  uint16 = 0x0a1b
	PacketZCRouletteInfo  uint16 = 0x0a1c
	PacketCZCloseRoulette uint16 = 0x0a1d
	PacketZCCloseRoulette uint16 = 0x0a1e
	PacketCZSpinRoulette  uint16 = 0x0a1f
	PacketZCSpinRoulette  uint16 = 0x0a20
	PacketCZClaimRoulette uint16 = 0x0a21
	PacketZCClaimRoulette uint16 = 0x0a22

	PacketZCMarketOpen     uint16 = 0x09d5
	PacketZCMarketOpen2    uint16 = 0x0b7a
	PacketCZMarketPurchase uint16 = 0x09d6
	PacketZCMarketResult   uint16 = 0x09d7
	PacketCZMarketClose    uint16 = 0x09d8

	PacketCZBarterPurchase uint16 = 0x0b0f

	PacketZCClanInfo   uint16 = 0x098a
	PacketZCClanOnline uint16 = 0x0988
	PacketZCClanLeave  uint16 = 0x0989
	PacketCZClanChat   uint16 = 0x098d
	PacketZCClanChat   uint16 = 0x098e

	PacketCZHotkeyChangeV2 uint16 = 0x0b21
)

type RouletteOpen struct {
	Result uint8
	Gold   int32
	Silver int32
	Bronze int32
}

type RouletteSpin struct {
	Result uint8
	Step   uint16
	Index  uint16
	ItemID uint16
	Gold   int32
	Silver int32
	Bronze int32
}

type MarketItem struct {
	ItemID uint32
	Type   uint8
	Price  uint32
	Qty    uint32
}

type BarterItem struct {
	ItemID   uint32
	Amount   uint32
	Currency uint32
	Cost     uint32
	Index    uint32
}

type ClanInfo struct {
	ID     uint32
	Name   string
	Master string
	Map    string
	Online uint16
	Total  uint16
}

func ParseRouletteOpen(p Packet) (RouletteOpen, bool, error) {
	if p.ID != PacketZCOpenRoulette {
		return RouletteOpen{}, false, nil
	}
	if len(p.Data) < 23 {
		return RouletteOpen{}, false, fmt.Errorf("ZC_ACK_OPEN_ROULETTE: invalid length %d", len(p.Data))
	}
	return RouletteOpen{
		Result: p.Data[2],
		Gold:   int32(binary.LittleEndian.Uint32(p.Data[11:15])),
		Silver: int32(binary.LittleEndian.Uint32(p.Data[15:19])),
		Bronze: int32(binary.LittleEndian.Uint32(p.Data[19:23])),
	}, true, nil
}

func ParseRouletteSpin(p Packet) (RouletteSpin, bool, error) {
	if p.ID != PacketZCSpinRoulette {
		return RouletteSpin{}, false, nil
	}
	if len(p.Data) < 21 {
		return RouletteSpin{}, false, fmt.Errorf("ZC_ACK_GENERATE_ROULETTE: invalid length %d", len(p.Data))
	}
	return RouletteSpin{
		Result: p.Data[2],
		Step:   binary.LittleEndian.Uint16(p.Data[3:5]),
		Index:  binary.LittleEndian.Uint16(p.Data[5:7]),
		ItemID: binary.LittleEndian.Uint16(p.Data[7:9]),
		Gold:   int32(binary.LittleEndian.Uint32(p.Data[9:13])),
		Silver: int32(binary.LittleEndian.Uint32(p.Data[13:17])),
		Bronze: int32(binary.LittleEndian.Uint32(p.Data[17:21])),
	}, true, nil
}

func ParseMarketOpen(p Packet) ([]MarketItem, bool, error) {
	if p.ID != PacketZCMarketOpen && p.ID != PacketZCMarketOpen2 {
		return nil, false, nil
	}
	if len(p.Data) < 4 {
		return nil, false, fmt.Errorf("ZC_NPC_MARKET_OPEN: invalid length %d", len(p.Data))
	}
	entry := 19
	if p.ID == PacketZCMarketOpen {
		entry = 15
	}
	body := p.Data[4:]
	if entry == 0 || len(body)%entry != 0 {
		return nil, true, nil
	}
	items := make([]MarketItem, 0, len(body)/entry)
	for i := 0; i+entry <= len(body); i += entry {
		e := body[i : i+entry]
		items = append(items, MarketItem{
			ItemID: binary.LittleEndian.Uint32(e[0:4]),
			Type:   e[4],
			Price:  binary.LittleEndian.Uint32(e[5:9]),
			Qty:    binary.LittleEndian.Uint32(e[9:13]),
		})
	}
	return items, true, nil
}

func ParseBarterList(p Packet) ([]BarterItem, bool, error) {
	if p.ID != PacketZCBarterList {
		return nil, false, nil
	}
	if len(p.Data) < 4 {
		return nil, false, fmt.Errorf("ZC_NPC_BARTER_MARKET_ITEMINFO: invalid length %d", len(p.Data))
	}
	const entry = 31
	body := p.Data[4:]
	if len(body)%entry != 0 {
		return nil, true, nil
	}
	items := make([]BarterItem, 0, len(body)/entry)
	for i := 0; i+entry <= len(body); i += entry {
		e := body[i : i+entry]
		items = append(items, BarterItem{
			ItemID:   binary.LittleEndian.Uint32(e[0:4]),
			Amount:   binary.LittleEndian.Uint32(e[5:9]),
			Currency: binary.LittleEndian.Uint32(e[9:13]),
			Cost:     binary.LittleEndian.Uint32(e[13:17]),
			Index:    binary.LittleEndian.Uint32(e[21:25]),
		})
	}
	return items, true, nil
}

func ParseClanInfo(p Packet) (ClanInfo, bool, error) {
	if p.ID != PacketZCClanInfo {
		return ClanInfo{}, false, nil
	}
	if len(p.Data) < 4+4+24+24+16+2 {
		return ClanInfo{}, false, fmt.Errorf("ZC_CLANINFO: invalid length %d", len(p.Data))
	}
	return ClanInfo{
		ID:     binary.LittleEndian.Uint32(p.Data[4:8]),
		Name:   decodeROFixedString(p.Data[8:32]),
		Master: decodeROFixedString(p.Data[32:56]),
		Map:    decodeROFixedString(p.Data[56:72]),
	}, true, nil
}

func ParseClanOnline(p Packet) (ClanInfo, bool, error) {
	if p.ID != PacketZCClanOnline {
		return ClanInfo{}, false, nil
	}
	if len(p.Data) < 6 {
		return ClanInfo{}, false, fmt.Errorf("ZC_NOTIFY_CLAN_CONNECTINFO: invalid length %d", len(p.Data))
	}
	return ClanInfo{
		Online: binary.LittleEndian.Uint16(p.Data[2:4]),
		Total:  binary.LittleEndian.Uint16(p.Data[4:6]),
	}, true, nil
}

func ParseClanChat(p Packet) (string, bool, error) {
	if p.ID != PacketZCClanChat {
		return "", false, nil
	}
	if len(p.Data) < 4+24 {
		return "", false, fmt.Errorf("ZC_NOTIFY_CLAN_CHAT: invalid length %d", len(p.Data))
	}
	name := decodeROFixedString(p.Data[4:28])
	msg := decodeROFixedString(p.Data[28:])
	if name == "" {
		return msg, true, nil
	}
	return name + " : " + msg, true, nil
}

func BuildEmptyPacket(id uint16) []byte {
	data := make([]byte, 2)
	binary.LittleEndian.PutUint16(data, id)
	return data
}

func BuildMarketPurchasePacket(itemID uint32, qty int32) []byte {
	data := make([]byte, 12)
	binary.LittleEndian.PutUint16(data[0:2], PacketCZMarketPurchase)
	binary.LittleEndian.PutUint16(data[2:4], 12)
	binary.LittleEndian.PutUint32(data[4:8], itemID)
	binary.LittleEndian.PutUint32(data[8:12], uint32(qty))
	return data
}

func BuildBarterPurchasePacket(itemID, amount, shopIndex uint32) []byte {
	data := make([]byte, 18)
	binary.LittleEndian.PutUint16(data[0:2], PacketCZBarterPurchase)
	binary.LittleEndian.PutUint16(data[2:4], 18)
	binary.LittleEndian.PutUint32(data[4:8], itemID)
	binary.LittleEndian.PutUint32(data[8:12], amount)
	binary.LittleEndian.PutUint16(data[12:14], 0)
	binary.LittleEndian.PutUint32(data[14:18], shopIndex)
	return data
}

func BuildClanChatPacket(msg string) []byte {
	data := make([]byte, 4+len(msg)+1)
	binary.LittleEndian.PutUint16(data[0:2], PacketCZClanChat)
	binary.LittleEndian.PutUint16(data[2:4], uint16(len(data)))
	copy(data[4:], msg)
	return data
}

func BuildHotkeyPacketV2(tab, index uint16, slot HotkeySlot) []byte {
	data := make([]byte, 13)
	binary.LittleEndian.PutUint16(data[0:2], PacketCZHotkeyChangeV2)
	binary.LittleEndian.PutUint16(data[2:4], tab)
	binary.LittleEndian.PutUint16(data[4:6], index)
	data[6] = slot.Type
	binary.LittleEndian.PutUint32(data[7:11], slot.ID)
	binary.LittleEndian.PutUint16(data[11:13], slot.Level)
	return data
}

func BuildClaimRoulettePacket() []byte {
	data := make([]byte, 3)
	binary.LittleEndian.PutUint16(data[0:2], PacketCZClaimRoulette)
	return data
}

func (c *Client) SendOpenRoulette() error  { return c.Send(BuildEmptyPacket(PacketCZOpenRoulette)) }
func (c *Client) SendRouletteInfo() error  { return c.Send(BuildEmptyPacket(PacketCZRouletteInfo)) }
func (c *Client) SendCloseRoulette() error { return c.Send(BuildEmptyPacket(PacketCZCloseRoulette)) }
func (c *Client) SendSpinRoulette() error  { return c.Send(BuildEmptyPacket(PacketCZSpinRoulette)) }
func (c *Client) SendClaimRoulette() error { return c.Send(BuildClaimRoulettePacket()) }
func (c *Client) SendMarketClose() error   { return c.Send(BuildEmptyPacket(PacketCZMarketClose)) }
func (c *Client) SendBarterClose() error   { return c.Send(BuildEmptyPacket(PacketCZBarterClose)) }
func (c *Client) SendMarketPurchase(itemID uint32, qty int32) error {
	return c.Send(BuildMarketPurchasePacket(itemID, qty))
}
func (c *Client) SendBarterPurchase(itemID, amount, shopIndex uint32) error {
	return c.Send(BuildBarterPurchasePacket(itemID, amount, shopIndex))
}
func (c *Client) SendClanChat(msg string) error { return c.Send(BuildClanChatPacket(msg)) }
func (c *Client) SendHotkeyTab(tab, index uint16, slot HotkeySlot) error {
	if IsClientDate20250716(c.clientDate) {
		return c.Send(BuildHotkeyPacketV2(tab, index, slot))
	}
	return c.SendHotkey(index, slot)
}
