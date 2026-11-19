package network

import (
	"encoding/binary"
	"fmt"
)

const (
	PacketZCBankCheck    uint16 = 0x09a6
	PacketCZBankDeposit  uint16 = 0x09a7
	PacketZCBankDeposit  uint16 = 0x09a8
	PacketCZBankWithdraw uint16 = 0x09a9
	PacketZCBankWithdraw uint16 = 0x09aa
	PacketCZBankCheck    uint16 = 0x09ab
	PacketCZBankOpen     uint16 = 0x09b6
	PacketZCBankOpen     uint16 = 0x09b7
	PacketCZBankClose    uint16 = 0x09b8
	PacketZCBankClose    uint16 = 0x09b9

	PacketZCUIOpen  uint16 = 0x0ae2
	PacketZCUIOpen2 uint16 = 0x0b9a
	PacketCZOpenUI  uint16 = 0x0a68

	UITypeBank    uint8 = 0
	UITypeEnchant uint8 = 10

	BankReasonOK uint16 = 0
)

type BankBalance struct {
	Vault  int64
	Zeny   int32
	Reason uint16
}

type BankOpenAck struct {
	Reason uint16
}

type UIOpen struct {
	Type uint8
	Data int64
}

func ParseBankCheck(p Packet) (BankBalance, bool, error) {
	if p.ID != PacketZCBankCheck {
		return BankBalance{}, false, nil
	}
	if len(p.Data) != 12 {
		return BankBalance{}, false, fmt.Errorf("ZC_BANKING_CHECK: invalid length %d", len(p.Data))
	}
	return BankBalance{
		Vault:  int64(binary.LittleEndian.Uint64(p.Data[2:10])),
		Reason: binary.LittleEndian.Uint16(p.Data[10:12]),
	}, true, nil
}

func ParseBankAck(p Packet) (BankBalance, bool, error) {
	if p.ID != PacketZCBankDeposit && p.ID != PacketZCBankWithdraw {
		return BankBalance{}, false, nil
	}
	if len(p.Data) != 16 {
		return BankBalance{}, false, fmt.Errorf("ZC_BANKING_ACK 0x%04X: invalid length %d", p.ID, len(p.Data))
	}
	return BankBalance{
		Reason: binary.LittleEndian.Uint16(p.Data[2:4]),
		Vault:  int64(binary.LittleEndian.Uint64(p.Data[4:12])),
		Zeny:   int32(binary.LittleEndian.Uint32(p.Data[12:16])),
	}, true, nil
}

func ParseBankOpenAck(p Packet) (BankOpenAck, bool, error) {
	if p.ID != PacketZCBankOpen && p.ID != PacketZCBankClose {
		return BankOpenAck{}, false, nil
	}
	if len(p.Data) != 4 {
		return BankOpenAck{}, false, fmt.Errorf("ZC_BANKING_OPEN/CLOSE 0x%04X: invalid length %d", p.ID, len(p.Data))
	}
	return BankOpenAck{Reason: binary.LittleEndian.Uint16(p.Data[2:4])}, true, nil
}

func ParseUIOpen(p Packet) (UIOpen, bool, error) {
	switch p.ID {
	case PacketZCUIOpen:
		if len(p.Data) < 3 {
			return UIOpen{}, false, fmt.Errorf("ZC_UI_OPEN: invalid length %d", len(p.Data))
		}
		ui := UIOpen{Type: p.Data[2]}
		if len(p.Data) >= 7 {
			ui.Data = int64(int32(binary.LittleEndian.Uint32(p.Data[3:7])))
		}
		return ui, true, nil
	case PacketZCUIOpen2:
		if len(p.Data) != 11 {
			return UIOpen{}, false, fmt.Errorf("ZC_UI_OPEN2: invalid length %d", len(p.Data))
		}
		return UIOpen{
			Type: p.Data[2],
			Data: int64(binary.LittleEndian.Uint64(p.Data[3:11])),
		}, true, nil
	default:
		return UIOpen{}, false, nil
	}
}

func BuildBankDepositPacket(accountID uint32, money int32) []byte {
	data := make([]byte, 10)
	binary.LittleEndian.PutUint16(data[0:2], PacketCZBankDeposit)
	binary.LittleEndian.PutUint32(data[2:6], accountID)
	binary.LittleEndian.PutUint32(data[6:10], uint32(money))
	return data
}

func BuildBankWithdrawPacket(accountID uint32, money int32) []byte {
	data := make([]byte, 10)
	binary.LittleEndian.PutUint16(data[0:2], PacketCZBankWithdraw)
	binary.LittleEndian.PutUint32(data[2:6], accountID)
	binary.LittleEndian.PutUint32(data[6:10], uint32(money))
	return data
}

func BuildBankCheckPacket(accountID uint32) []byte {
	data := make([]byte, 6)
	binary.LittleEndian.PutUint16(data[0:2], PacketCZBankCheck)
	binary.LittleEndian.PutUint32(data[2:6], accountID)
	return data
}

func BuildBankOpenPacket(accountID uint32) []byte {
	data := make([]byte, 6)
	binary.LittleEndian.PutUint16(data[0:2], PacketCZBankOpen)
	binary.LittleEndian.PutUint32(data[2:6], accountID)
	return data
}

func BuildBankClosePacket(accountID uint32) []byte {
	data := make([]byte, 6)
	binary.LittleEndian.PutUint16(data[0:2], PacketCZBankClose)
	binary.LittleEndian.PutUint32(data[2:6], accountID)
	return data
}

func (c *Client) SendBankDeposit(accountID uint32, money int32) error {
	return c.Send(BuildBankDepositPacket(accountID, money))
}

func (c *Client) SendBankWithdraw(accountID uint32, money int32) error {
	return c.Send(BuildBankWithdrawPacket(accountID, money))
}

func (c *Client) SendBankCheck(accountID uint32) error {
	return c.Send(BuildBankCheckPacket(accountID))
}

func (c *Client) SendBankOpen(accountID uint32) error {
	return c.Send(BuildBankOpenPacket(accountID))
}

func (c *Client) SendBankClose(accountID uint32) error {
	return c.Send(BuildBankClosePacket(accountID))
}
