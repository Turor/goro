package network

import (
	"encoding/binary"
	"fmt"
)

const (
	PacketZCConfig uint16 = 0x02d9

	ConfigPetFeed   uint32 = 2
	ConfigHomunFeed uint32 = 3

	PacketCZEquipSwitchAdd    uint16 = 0x0a97
	PacketCZEquipSwitchRemove uint16 = 0x0a99
	PacketCZEquipSwitchAll    uint16 = 0x0a9c
	PacketCZEquipSwitchSingle uint16 = 0x0ace

	PacketCZSearchStore      uint16 = 0x0835
	PacketZCSearchStoreAck   uint16 = 0x0b64
	PacketZCSearchStoreOpen  uint16 = 0x083a
	PacketCZSearchStoreNext  uint16 = 0x0838
	PacketCZSearchStoreClick uint16 = 0x083c
	PacketCZSearchStoreClose uint16 = 0x083b

	PacketCZBarterClose uint16 = 0x0b12
	PacketZCBarterList  uint16 = 0x0b78

	PacketCZRodexAddItem    uint16 = 0x0a04
	PacketZCRodexAddItemAck uint16 = 0x0b3f
	PacketCZRodexRemoveItem uint16 = 0x0a06
	PacketCZCheckName       uint16 = 0x0a13
	PacketZCCheckName       uint16 = 0x0a51
)

type ConfigFlag struct {
	Type  uint32
	Value uint32
}

type SearchStoreOpen struct {
	Effect uint16
	Uses   uint8
}

type SearchStoreHit struct {
	StoreID  uint32
	AID      uint32
	ShopName string
	ItemID   uint32
	Price    uint32
	Amount   uint16
}

type SearchStoreAck struct {
	FirstPage bool
	NextPage  bool
	Uses      uint8
	Hits      []SearchStoreHit
}

type RodexAddAck struct {
	Result uint8
	Index  uint16
	Count  uint16
	ItemID uint32
}

func ParseConfigFlag(p Packet) (ConfigFlag, bool, error) {
	if p.ID != PacketZCConfig {
		return ConfigFlag{}, false, nil
	}
	if len(p.Data) != 10 {
		return ConfigFlag{}, false, fmt.Errorf("ZC_CONFIG: invalid length %d", len(p.Data))
	}
	return ConfigFlag{
		Type:  binary.LittleEndian.Uint32(p.Data[2:6]),
		Value: binary.LittleEndian.Uint32(p.Data[6:10]),
	}, true, nil
}

func ParseSearchStoreOpen(p Packet) (SearchStoreOpen, bool, error) {
	if p.ID != PacketZCSearchStoreOpen {
		return SearchStoreOpen{}, false, nil
	}
	if len(p.Data) < 5 {
		return SearchStoreOpen{}, false, fmt.Errorf("ZC_OPEN_SEARCH_STORE: invalid length %d", len(p.Data))
	}
	return SearchStoreOpen{
		Effect: binary.LittleEndian.Uint16(p.Data[2:4]),
		Uses:   p.Data[4],
	}, true, nil
}

func ParseSearchStoreAck(p Packet) (SearchStoreAck, bool, error) {
	if p.ID != PacketZCSearchStoreAck {
		return SearchStoreAck{}, false, nil
	}
	if len(p.Data) < 7 {
		return SearchStoreAck{}, false, fmt.Errorf("ZC_SEARCH_STORE_INFO_ACK: invalid length %d", len(p.Data))
	}
	ack := SearchStoreAck{FirstPage: p.Data[4] != 0, NextPage: p.Data[5] != 0, Uses: p.Data[6]}
	// 20250716 PACKET_ZC_SEARCH_STORE_INFO_ACK_sub:
	// storeId4 + AID4 + shopName80 + itemId4 + type1 + price4 + amount2
	// + EQUIPSLOTINFO 16 + options 5*5 + forge 3*5 + refine1 + grade1 = 157
	const entry = 157
	body := p.Data[7:]
	for i := 0; i+entry <= len(body); i += entry {
		e := body[i : i+entry]
		ack.Hits = append(ack.Hits, SearchStoreHit{
			StoreID:  binary.LittleEndian.Uint32(e[0:4]),
			AID:      binary.LittleEndian.Uint32(e[4:8]),
			ShopName: decodeROFixedString(e[8:88]),
			ItemID:   binary.LittleEndian.Uint32(e[88:92]),
			Price:    binary.LittleEndian.Uint32(e[93:97]),
			Amount:   binary.LittleEndian.Uint16(e[97:99]),
		})
	}
	return ack, true, nil
}

func ParseRodexAddAck(p Packet) (RodexAddAck, bool, error) {
	if p.ID != PacketZCRodexAddItemAck {
		return RodexAddAck{}, false, nil
	}
	if len(p.Data) < 11 {
		return RodexAddAck{}, false, fmt.Errorf("ZC_ACK_ADD_ITEM_RODEX: invalid length %d", len(p.Data))
	}
	return RodexAddAck{
		Result: p.Data[2],
		Index:  binary.LittleEndian.Uint16(p.Data[3:5]),
		Count:  binary.LittleEndian.Uint16(p.Data[5:7]),
		ItemID: binary.LittleEndian.Uint32(p.Data[7:11]),
	}, true, nil
}

func BuildConfigPacket(typ, value uint32) []byte {
	data := make([]byte, 10)
	binary.LittleEndian.PutUint16(data[0:2], PacketCZConfig)
	binary.LittleEndian.PutUint32(data[2:6], typ)
	binary.LittleEndian.PutUint32(data[6:10], value)
	return data
}

func BuildEquipSwitchAddPacket(index uint16, position uint32) []byte {
	data := make([]byte, 8)
	binary.LittleEndian.PutUint16(data[0:2], PacketCZEquipSwitchAdd)
	binary.LittleEndian.PutUint16(data[2:4], index)
	binary.LittleEndian.PutUint32(data[4:8], position)
	return data
}

func BuildEquipSwitchAllPacket() []byte {
	data := make([]byte, 2)
	binary.LittleEndian.PutUint16(data[0:2], PacketCZEquipSwitchAll)
	return data
}

func BuildSearchStorePacket(itemID uint32) []byte {
	data := make([]byte, 15+4)
	binary.LittleEndian.PutUint16(data[0:2], PacketCZSearchStore)
	binary.LittleEndian.PutUint16(data[2:4], uint16(len(data)))
	data[4] = 0 // vending
	data[13] = 1
	binary.LittleEndian.PutUint32(data[15:19], itemID)
	return data
}

func BuildRodexAddItemPacket(index, count uint16) []byte {
	data := make([]byte, 6)
	binary.LittleEndian.PutUint16(data[0:2], PacketCZRodexAddItem)
	binary.LittleEndian.PutUint16(data[2:4], index)
	binary.LittleEndian.PutUint16(data[4:6], count)
	return data
}

func BuildCheckNamePacket(name string) []byte {
	data := make([]byte, 26)
	binary.LittleEndian.PutUint16(data[0:2], PacketCZCheckName)
	copy(data[2:26], name)
	return data
}

func (c *Client) SendConfig(typ, value uint32) error { return c.Send(BuildConfigPacket(typ, value)) }
func (c *Client) SendEquipSwitchAdd(index uint16, position uint32) error {
	return c.Send(BuildEquipSwitchAddPacket(index, position))
}
func (c *Client) SendEquipSwitchAll() error           { return c.Send(BuildEquipSwitchAllPacket()) }
func (c *Client) SendSearchStore(itemID uint32) error { return c.Send(BuildSearchStorePacket(itemID)) }
func (c *Client) SendRodexAddItem(index, count uint16) error {
	return c.Send(BuildRodexAddItemPacket(index, count))
}
func (c *Client) SendCheckName(name string) error { return c.Send(BuildCheckNamePacket(name)) }
