package network

import (
	"encoding/binary"
	"fmt"
)

const (
	PacketCZRequestRandomEnchant  uint16 = 0x0b9b
	PacketCZRequestPerfectEnchant uint16 = 0x0b9c
	PacketCZRequestUpgradeEnchant uint16 = 0x0b9d
	PacketCZRequestResetEnchant   uint16 = 0x0b9e
	PacketZCResponseEnchant       uint16 = 0x0b9f
	PacketCZCloseUIEnchant        uint16 = 0x0ba0
)

type EnchantResponse struct {
	MsgID  int32
	ItemID int32
}

func ParseEnchantResponse(p Packet) (EnchantResponse, bool, error) {
	if p.ID != PacketZCResponseEnchant {
		return EnchantResponse{}, false, nil
	}
	if len(p.Data) != 10 {
		return EnchantResponse{}, false, fmt.Errorf("ZC_RESPONSE_ENCHANT: invalid length %d", len(p.Data))
	}
	return EnchantResponse{
		MsgID:  int32(binary.LittleEndian.Uint32(p.Data[2:6])),
		ItemID: int32(binary.LittleEndian.Uint32(p.Data[6:10])),
	}, true, nil
}

func BuildRandomEnchantPacket(group int64, index uint16) []byte {
	data := make([]byte, 12)
	binary.LittleEndian.PutUint16(data[0:2], PacketCZRequestRandomEnchant)
	binary.LittleEndian.PutUint64(data[2:10], uint64(group))
	binary.LittleEndian.PutUint16(data[10:12], ToClientInventoryIndex(index))
	return data
}

func BuildPerfectEnchantPacket(group int64, index uint16, itemID int32) []byte {
	data := make([]byte, 16)
	binary.LittleEndian.PutUint16(data[0:2], PacketCZRequestPerfectEnchant)
	binary.LittleEndian.PutUint64(data[2:10], uint64(group))
	binary.LittleEndian.PutUint16(data[10:12], ToClientInventoryIndex(index))
	binary.LittleEndian.PutUint32(data[12:16], uint32(itemID))
	return data
}

func BuildUpgradeEnchantPacket(group int64, index uint16, slot uint16) []byte {
	data := make([]byte, 14)
	binary.LittleEndian.PutUint16(data[0:2], PacketCZRequestUpgradeEnchant)
	binary.LittleEndian.PutUint64(data[2:10], uint64(group))
	binary.LittleEndian.PutUint16(data[10:12], ToClientInventoryIndex(index))
	binary.LittleEndian.PutUint16(data[12:14], slot)
	return data
}

func BuildResetEnchantPacket(group int64, index uint16) []byte {
	data := make([]byte, 12)
	binary.LittleEndian.PutUint16(data[0:2], PacketCZRequestResetEnchant)
	binary.LittleEndian.PutUint64(data[2:10], uint64(group))
	binary.LittleEndian.PutUint16(data[10:12], ToClientInventoryIndex(index))
	return data
}

func BuildCloseEnchantUIPacket() []byte {
	data := make([]byte, 2)
	binary.LittleEndian.PutUint16(data[0:2], PacketCZCloseUIEnchant)
	return data
}

func (c *Client) SendRandomEnchant(group int64, index uint16) error {
	return c.Send(BuildRandomEnchantPacket(group, index))
}

func (c *Client) SendPerfectEnchant(group int64, index uint16, itemID int32) error {
	return c.Send(BuildPerfectEnchantPacket(group, index, itemID))
}

func (c *Client) SendUpgradeEnchant(group int64, index uint16, slot uint16) error {
	return c.Send(BuildUpgradeEnchantPacket(group, index, slot))
}

func (c *Client) SendResetEnchant(group int64, index uint16) error {
	return c.Send(BuildResetEnchantPacket(group, index))
}

func (c *Client) SendCloseEnchantUI() error {
	return c.Send(BuildCloseEnchantUIPacket())
}
