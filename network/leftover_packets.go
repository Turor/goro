package network

import (
	"encoding/binary"
	"fmt"
)

const (
	PacketZCReformOpen  uint16 = 0x0b8f
	PacketCZReformClose uint16 = 0x0b90
	PacketCZReformItem  uint16 = 0x0b91
	PacketZCReformAck   uint16 = 0x0b92

	PacketZCLaphineOpen  uint16 = 0x0a4e
	PacketCZLaphineMake  uint16 = 0x0a4f
	PacketZCLaphineAck   uint16 = 0x0a50
	PacketCZLaphineClose uint16 = 0x0a70

	PacketZCLaphineUpgOpen  uint16 = 0x0ab4
	PacketCZLaphineUpgClose uint16 = 0x0ab5
	PacketCZLaphineUpgMake  uint16 = 0x0ab6
	PacketZCLaphineUpgAck   uint16 = 0x0ab7

	PacketCZStylingClose uint16 = 0x0a48
	UITypeStyling        uint8  = 1

	PacketZCEdenList   uint16 = 0x0efb
	PacketZCEdenResult uint16 = 0x0efc
	PacketZCEdenShop   uint16 = 0x0efd
	PacketCZEdenReq    uint16 = 0x0ef3
	PacketZCMvpMinimap uint16 = 0x0ef5

	EdenActList     uint16 = 0
	EdenActAccept   uint16 = 1
	EdenActComplete uint16 = 2
	EdenActAbandon  uint16 = 3
	EdenActBuy      uint16 = 4

	edenQuestSize = 203
	edenShopSize  = 20
)

type ReformOpen struct {
	ItemID int32
}

type ReformAck struct {
	Index  uint16
	Result int8
}

type LaphineOpen struct {
	ItemID  uint32
	Upgrade bool
}

type EdenQuest struct {
	QuestID uint32
	State   uint8
	Name    string
	Map     string
	Merit   int16
}

type EdenList struct {
	Level  int16
	Merit  int32
	Quests []EdenQuest
}

type EdenShopItem struct {
	Slot       uint16
	ItemID     uint32
	Amount     uint16
	PriceMerit int32
	PriceZeny  int32
}

type EdenShop struct {
	Merit int32
	Items []EdenShopItem
}

type EdenResult struct {
	Result  int16
	QuestID uint32
}

func ParseReformOpen(p Packet) (ReformOpen, bool, error) {
	if p.ID != PacketZCReformOpen {
		return ReformOpen{}, false, nil
	}
	if len(p.Data) < 6 {
		return ReformOpen{}, false, fmt.Errorf("ZC_OPEN_REFORM_UI: invalid length %d", len(p.Data))
	}
	return ReformOpen{ItemID: int32(binary.LittleEndian.Uint32(p.Data[2:6]))}, true, nil
}

func ParseReformAck(p Packet) (ReformAck, bool, error) {
	if p.ID != PacketZCReformAck {
		return ReformAck{}, false, nil
	}
	if len(p.Data) < 5 {
		return ReformAck{}, false, fmt.Errorf("ZC_ITEM_REFORM_ACK: invalid length %d", len(p.Data))
	}
	return ReformAck{
		Index:  binary.LittleEndian.Uint16(p.Data[2:4]),
		Result: int8(p.Data[4]),
	}, true, nil
}

func ParseLaphineOpen(p Packet) (LaphineOpen, bool, error) {
	upgrade := p.ID == PacketZCLaphineUpgOpen
	if p.ID != PacketZCLaphineOpen && !upgrade {
		return LaphineOpen{}, false, nil
	}
	if len(p.Data) < 6 {
		return LaphineOpen{}, false, fmt.Errorf("laphine open: invalid length %d", len(p.Data))
	}
	return LaphineOpen{ItemID: binary.LittleEndian.Uint32(p.Data[2:6]), Upgrade: upgrade}, true, nil
}

func ParseLaphineAck(p Packet) (int16, bool, error) {
	if p.ID != PacketZCLaphineAck && p.ID != PacketZCLaphineUpgAck {
		return 0, false, nil
	}
	if len(p.Data) < 4 {
		return 0, false, fmt.Errorf("laphine ack: invalid length %d", len(p.Data))
	}
	return int16(binary.LittleEndian.Uint16(p.Data[2:4])), true, nil
}

func ParseEdenList(p Packet) (EdenList, bool, error) {
	if p.ID != PacketZCEdenList {
		return EdenList{}, false, nil
	}
	if len(p.Data) < 12 {
		return EdenList{}, false, fmt.Errorf("ZC_TURORAN_EDEN_LIST: invalid length %d", len(p.Data))
	}
	list := EdenList{
		Level: int16(binary.LittleEndian.Uint16(p.Data[4:6])),
		Merit: int32(binary.LittleEndian.Uint32(p.Data[8:12])),
	}
	count := int(p.Data[6])
	body := p.Data[12:]
	for i := 0; i < count && (i+1)*edenQuestSize <= len(body); i++ {
		e := body[i*edenQuestSize : (i+1)*edenQuestSize]
		list.Quests = append(list.Quests, EdenQuest{
			QuestID: binary.LittleEndian.Uint32(e[0:4]),
			State:   e[4],
			Merit:   int16(binary.LittleEndian.Uint16(e[39:41])),
			Map:     decodeROFixedString(e[41:57]),
			Name:    decodeROFixedString(e[59:107]),
		})
	}
	return list, true, nil
}

func ParseEdenShop(p Packet) (EdenShop, bool, error) {
	if p.ID != PacketZCEdenShop {
		return EdenShop{}, false, nil
	}
	if len(p.Data) < 8 {
		return EdenShop{}, false, fmt.Errorf("ZC_TURORAN_EDEN_SHOP: invalid length %d", len(p.Data))
	}
	shop := EdenShop{Merit: int32(binary.LittleEndian.Uint32(p.Data[4:8]))}
	if len(p.Data) < 10 {
		return shop, true, nil
	}
	count := int(binary.LittleEndian.Uint16(p.Data[8:10]))
	body := p.Data[10:]
	for i := 0; i < count && (i+1)*edenShopSize <= len(body); i++ {
		e := body[i*edenShopSize : (i+1)*edenShopSize]
		shop.Items = append(shop.Items, EdenShopItem{
			Slot:       binary.LittleEndian.Uint16(e[0:2]),
			ItemID:     binary.LittleEndian.Uint32(e[4:8]),
			Amount:     binary.LittleEndian.Uint16(e[8:10]),
			PriceMerit: int32(binary.LittleEndian.Uint32(e[10:14])),
			PriceZeny:  int32(binary.LittleEndian.Uint32(e[14:18])),
		})
	}
	return shop, true, nil
}

func ParseEdenResult(p Packet) (EdenResult, bool, error) {
	if p.ID != PacketZCEdenResult {
		return EdenResult{}, false, nil
	}
	if len(p.Data) < 8 {
		return EdenResult{}, false, fmt.Errorf("ZC_TURORAN_EDEN_RESULT: invalid length %d", len(p.Data))
	}
	return EdenResult{
		Result:  int16(binary.LittleEndian.Uint16(p.Data[2:4])),
		QuestID: binary.LittleEndian.Uint32(p.Data[4:8]),
	}, true, nil
}

func BuildReformItemPacket(itemID int32, index uint16) []byte {
	data := make([]byte, 8)
	binary.LittleEndian.PutUint16(data[0:2], PacketCZReformItem)
	binary.LittleEndian.PutUint32(data[2:6], uint32(itemID))
	binary.LittleEndian.PutUint16(data[6:8], index)
	return data
}

func BuildLaphineMakePacket(itemID uint32, index, count uint16) []byte {
	data := make([]byte, 12)
	binary.LittleEndian.PutUint16(data[0:2], PacketCZLaphineMake)
	binary.LittleEndian.PutUint16(data[2:4], 12)
	binary.LittleEndian.PutUint32(data[4:8], itemID)
	binary.LittleEndian.PutUint16(data[8:10], index)
	binary.LittleEndian.PutUint16(data[10:12], count)
	return data
}

func BuildLaphineUpgPacket(itemID uint32, index uint16) []byte {
	data := make([]byte, 8)
	binary.LittleEndian.PutUint16(data[0:2], PacketCZLaphineUpgMake)
	binary.LittleEndian.PutUint32(data[2:6], itemID)
	binary.LittleEndian.PutUint16(data[6:8], index)
	return data
}

func BuildEdenReqPacket(action uint16, questID uint32) []byte {
	data := make([]byte, 8)
	binary.LittleEndian.PutUint16(data[0:2], PacketCZEdenReq)
	binary.LittleEndian.PutUint16(data[2:4], action)
	binary.LittleEndian.PutUint32(data[4:8], questID)
	return data
}

func (c *Client) SendReformClose() error { return c.Send(BuildEmptyPacket(PacketCZReformClose)) }
func (c *Client) SendReformItem(itemID int32, index uint16) error {
	return c.Send(BuildReformItemPacket(itemID, index))
}
func (c *Client) SendLaphineClose() error { return c.Send(BuildEmptyPacket(PacketCZLaphineClose)) }
func (c *Client) SendLaphineMake(itemID uint32, index, count uint16) error {
	return c.Send(BuildLaphineMakePacket(itemID, index, count))
}
func (c *Client) SendLaphineUpgClose() error {
	return c.Send(BuildEmptyPacket(PacketCZLaphineUpgClose))
}
func (c *Client) SendLaphineUpg(itemID uint32, index uint16) error {
	return c.Send(BuildLaphineUpgPacket(itemID, index))
}
func (c *Client) SendStylingClose() error { return c.Send(BuildEmptyPacket(PacketCZStylingClose)) }
func (c *Client) SendEdenReq(action uint16, questID uint32) error {
	return c.Send(BuildEdenReqPacket(action, questID))
}
