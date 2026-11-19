package network

import (
	"encoding/binary"
	"fmt"
)

const (
	PacketZCRodexIcon       uint16 = 0x09e7
	PacketCZRodexOpen       uint16 = 0x0ac0
	PacketCZRodexRefresh    uint16 = 0x0ac1
	PacketZCRodexMailList   uint16 = 0x0ac2
	PacketCZRodexClose      uint16 = 0x09e9
	PacketCZRodexRead       uint16 = 0x09ea
	PacketZCRodexRead       uint16 = 0x0b63
	PacketCZRodexNext       uint16 = 0x09ee
	PacketCZRodexDelete     uint16 = 0x09f5
	PacketZCRodexDelete     uint16 = 0x09f6
	PacketCZRodexGetZeny    uint16 = 0x09f1
	PacketZCRodexGetZeny    uint16 = 0x09f2
	PacketCZRodexGetItems   uint16 = 0x09f3
	PacketZCRodexGetItems   uint16 = 0x09f4
	PacketCZRodexSend       uint16 = 0x0a6e
	PacketZCRodexSendResult uint16 = 0x09ed
	PacketCZRodexCheckName  uint16 = 0x0a13
	PacketZCRodexCheckName  uint16 = 0x0a51

	rodexNameLen        = 24
	rodexOpenPacketLen  = 26 // type + 3x int64
	rodexReadHeaderLen  = 24 // type2+len2+opentype1+mailid8+textlen2+zeny8+itemcnt1
	rodexMailListHeader = 5  // type2+len2+isEnd1
)

type RodexIcon struct {
	Show bool
}

type RodexMailHeader struct {
	OpenType uint8
	MailID   int64
	Read     bool
	Type     uint8
	Sender   string
	Expire   int32
	Title    string
}

type RodexMailList struct {
	End  bool
	Mail []RodexMailHeader
}

type RodexRead struct {
	OpenType uint8
	MailID   int64
	Zeny     int64
	Body     string
	Items    int
}

type RodexMailIDAck struct {
	OpenType uint8
	MailID   int64
	Result   uint8
}

func ParseRodexIcon(p Packet) (RodexIcon, bool, error) {
	if p.ID != PacketZCRodexIcon {
		return RodexIcon{}, false, nil
	}
	if len(p.Data) != 3 {
		return RodexIcon{}, false, fmt.Errorf("ZC_RODEX_ICON: invalid length %d", len(p.Data))
	}
	return RodexIcon{Show: p.Data[2] != 0}, true, nil
}

func ParseRodexMailList(p Packet) (RodexMailList, bool, error) {
	if p.ID != PacketZCRodexMailList {
		return RodexMailList{}, false, nil
	}
	if len(p.Data) < rodexMailListHeader {
		return RodexMailList{}, false, fmt.Errorf("ZC_RODEX_MAIL_LIST: invalid length %d", len(p.Data))
	}
	wireLen := int(binary.LittleEndian.Uint16(p.Data[2:4]))
	if wireLen != len(p.Data) {
		return RodexMailList{}, false, fmt.Errorf("ZC_RODEX_MAIL_LIST: length field %d != %d", wireLen, len(p.Data))
	}
	list := RodexMailList{End: p.Data[4] != 0}
	off := rodexMailListHeader
	for off < len(p.Data) {
		// openType1 + mailID8 + isread1 + type1 + sender24 + expire4 + titleLen2
		const fixed = 1 + 8 + 1 + 1 + 24 + 4 + 2
		if off+fixed > len(p.Data) {
			return RodexMailList{}, false, fmt.Errorf("ZC_RODEX_MAIL_LIST: truncated header at %d", off)
		}
		titleLen := int(binary.LittleEndian.Uint16(p.Data[off+39 : off+41]))
		if titleLen < 0 || off+fixed+titleLen > len(p.Data) {
			return RodexMailList{}, false, fmt.Errorf("ZC_RODEX_MAIL_LIST: truncated title at %d", off)
		}
		entry := RodexMailHeader{
			OpenType: p.Data[off],
			MailID:   int64(binary.LittleEndian.Uint64(p.Data[off+1 : off+9])),
			Read:     p.Data[off+9] != 0,
			Type:     p.Data[off+10],
			Sender:   decodeROFixedString(p.Data[off+11 : off+35]),
			Expire:   int32(binary.LittleEndian.Uint32(p.Data[off+35 : off+39])),
			Title:    decodeROFixedString(p.Data[off+41 : off+41+titleLen]),
		}
		list.Mail = append(list.Mail, entry)
		off += fixed + titleLen
	}
	return list, true, nil
}

func ParseRodexRead(p Packet) (RodexRead, bool, error) {
	if p.ID != PacketZCRodexRead {
		return RodexRead{}, false, nil
	}
	if len(p.Data) < rodexReadHeaderLen {
		return RodexRead{}, false, fmt.Errorf("ZC_ACK_READ_RODEX: invalid length %d", len(p.Data))
	}
	textLen := int(binary.LittleEndian.Uint16(p.Data[13:15]))
	if textLen < 0 || rodexReadHeaderLen+textLen > len(p.Data) {
		return RodexRead{}, false, fmt.Errorf("ZC_ACK_READ_RODEX: truncated body")
	}
	return RodexRead{
		OpenType: p.Data[4],
		MailID:   int64(binary.LittleEndian.Uint64(p.Data[5:13])),
		Zeny:     int64(binary.LittleEndian.Uint64(p.Data[15:23])),
		Items:    int(p.Data[23]),
		Body:     string(p.Data[rodexReadHeaderLen : rodexReadHeaderLen+textLen]),
	}, true, nil
}

func ParseRodexSendResult(p Packet) (uint8, bool, error) {
	if p.ID != PacketZCRodexSendResult {
		return 0, false, nil
	}
	if len(p.Data) != 3 {
		return 0, false, fmt.Errorf("ZC_WRITE_MAIL_RESULT: invalid length %d", len(p.Data))
	}
	return p.Data[2], true, nil
}

func ParseRodexMailIDAck(p Packet) (RodexMailIDAck, bool, error) {
	switch p.ID {
	case PacketZCRodexDelete:
		if len(p.Data) != 11 {
			return RodexMailIDAck{}, false, fmt.Errorf("ZC_ACK_DELETE_MAIL: invalid length %d", len(p.Data))
		}
		return RodexMailIDAck{
			OpenType: p.Data[2],
			MailID:   int64(binary.LittleEndian.Uint64(p.Data[3:11])),
		}, true, nil
	case PacketZCRodexGetZeny, PacketZCRodexGetItems:
		if len(p.Data) != 12 {
			return RodexMailIDAck{}, false, fmt.Errorf("ZC_ACK_FROM_MAIL 0x%04X: invalid length %d", p.ID, len(p.Data))
		}
		return RodexMailIDAck{
			MailID:   int64(binary.LittleEndian.Uint64(p.Data[2:10])),
			OpenType: p.Data[10],
			Result:   p.Data[11],
		}, true, nil
	default:
		return RodexMailIDAck{}, false, nil
	}
}

func BuildRodexOpenPacket() []byte {
	data := make([]byte, rodexOpenPacketLen)
	binary.LittleEndian.PutUint16(data[0:2], PacketCZRodexOpen)
	return data
}

func BuildRodexRefreshPacket() []byte {
	data := make([]byte, 2+8+16)
	binary.LittleEndian.PutUint16(data[0:2], PacketCZRodexRefresh)
	return data
}

func BuildRodexClosePacket() []byte {
	data := make([]byte, 2)
	binary.LittleEndian.PutUint16(data[0:2], PacketCZRodexClose)
	return data
}

func BuildRodexReadPacket(openType uint8, mailID int64) []byte {
	data := make([]byte, 11)
	binary.LittleEndian.PutUint16(data[0:2], PacketCZRodexRead)
	data[2] = openType
	binary.LittleEndian.PutUint64(data[3:11], uint64(mailID))
	return data
}

func BuildRodexDeletePacket(openType uint8, mailID int64) []byte {
	data := make([]byte, 11)
	binary.LittleEndian.PutUint16(data[0:2], PacketCZRodexDelete)
	data[2] = openType
	binary.LittleEndian.PutUint64(data[3:11], uint64(mailID))
	return data
}

func BuildRodexGetZenyPacket(openType uint8, mailID int64) []byte {
	data := make([]byte, 11)
	binary.LittleEndian.PutUint16(data[0:2], PacketCZRodexGetZeny)
	binary.LittleEndian.PutUint64(data[2:10], uint64(mailID))
	data[10] = openType
	return data
}

func BuildRodexGetItemsPacket(openType uint8, mailID int64) []byte {
	data := make([]byte, 11)
	binary.LittleEndian.PutUint16(data[0:2], PacketCZRodexGetItems)
	binary.LittleEndian.PutUint64(data[2:10], uint64(mailID))
	data[10] = openType
	return data
}

func BuildRodexSendPacket(receiver, sender, title, body string, zeny int64, receiverCharID int32) []byte {
	titleBytes := []byte(title)
	bodyBytes := []byte(body)
	header := 2 + 2 + 24 + 24 + 8 + 2 + 2 + 4
	data := make([]byte, header+len(titleBytes)+len(bodyBytes))
	binary.LittleEndian.PutUint16(data[0:2], PacketCZRodexSend)
	binary.LittleEndian.PutUint16(data[2:4], uint16(len(data)))
	copy(data[4:28], receiver)
	copy(data[28:52], sender)
	binary.LittleEndian.PutUint64(data[52:60], uint64(zeny))
	binary.LittleEndian.PutUint16(data[60:62], uint16(len(titleBytes)))
	binary.LittleEndian.PutUint16(data[62:64], uint16(len(bodyBytes)))
	binary.LittleEndian.PutUint32(data[64:68], uint32(receiverCharID))
	copy(data[68:], titleBytes)
	copy(data[68+len(titleBytes):], bodyBytes)
	return data
}

func (c *Client) SendRodexOpen() error    { return c.Send(BuildRodexOpenPacket()) }
func (c *Client) SendRodexRefresh() error { return c.Send(BuildRodexRefreshPacket()) }
func (c *Client) SendRodexClose() error   { return c.Send(BuildRodexClosePacket()) }
func (c *Client) SendRodexRead(openType uint8, mailID int64) error {
	return c.Send(BuildRodexReadPacket(openType, mailID))
}
func (c *Client) SendRodexDelete(openType uint8, mailID int64) error {
	return c.Send(BuildRodexDeletePacket(openType, mailID))
}
func (c *Client) SendRodexGetZeny(openType uint8, mailID int64) error {
	return c.Send(BuildRodexGetZenyPacket(openType, mailID))
}
func (c *Client) SendRodexGetItems(openType uint8, mailID int64) error {
	return c.Send(BuildRodexGetItemsPacket(openType, mailID))
}
func (c *Client) SendRodexSend(receiver, sender, title, body string, zeny int64, receiverCharID int32) error {
	return c.Send(BuildRodexSendPacket(receiver, sender, title, body, zeny, receiverCharID))
}
