package sim

import (
	"encoding/binary"

	"github.com/kivutar/goro/network"
)

const (
	fixtureAccountID uint32 = 2000001
	fixtureCharID    uint32 = 150001
	fixtureAuthCode  uint32 = 0x44556677
	fixtureCharName         = "kgap01"
	fixtureMapName          = "int_land"
	// RFC 5737 documentation address. Captures must never use live TEST IPs.
	fixtureCharIP   = "192.0.2.10"
	fixtureCharPort = uint16(6121)
	fixtureMapPort  = uint16(5121)
	fixtureFromX    = 81
	fixtureFromY    = 105
	fixtureToX      = 82
	fixtureToY      = 105
	fixtureMobID    = uint32(3000001)
	fixtureItemID   = uint32(501)
)

// LoginWalkCapture is the reduced 20250716 golden: login, char select, map
// enter, then one walk plus fight, cast, storage, warp, item-mark and Warpra
// packets. It is synthesized from Hercules PACKETVER 20250716 layouts, not
// a live TEST dump.
func LoginWalkCapture() Capture {
	loginPkt := accountAcceptLogin2()
	charList := charList2025()
	zone := zoneNotify2()
	enter := mapAcceptEnter2()
	features := packet(0x0EF1, 6)
	put32(features, 2, 1)
	walk := selfMoveAck(fixtureFromX, fixtureFromY, fixtureToX, fixtureToY)
	mob := idleMob2025()
	act := actorAction(fixtureAccountID, fixtureMobID, 12)
	cast := skillCast2025(fixtureAccountID, fixtureMobID, 19, 400)
	storeStart := inventoryStart(network.InventoryTypeStorage, "Kafra")
	storeItems := storageNormalList()
	storeEnd := inventoryEnd(network.InventoryTypeStorage)
	warp := mapChangeSameServer("geffen", 119, 113)
	marks := varPacket(0x0EE8, nil)
	warpra := varPacket(0x0EFE, nil)

	return Capture{
		Magic:      Magic,
		ClientDate: network.ClientDate20250716,
		Notes:      "reduced 20250716 fixture: login + walk (and framed fight/cast/storage/warp/marks/Warpra). Not a live TEST capture.",
		Streams: []Stream{
			{
				Name: "login",
				Records: []Record{
					{Dir: DirServerToClient, Hex: encodeHex(loginPkt), Note: "AC_ACCEPT_LOGIN 0x0AC4"},
				},
			},
			{
				Name:      "char",
				PrefixHex: encodeHex(put32(make([]byte, 4), 0, fixtureAccountID)),
				Records: []Record{
					{Dir: DirServerToClient, Hex: encodeHex(charList), Note: "HC_ACCEPT_ENTER 0x006B"},
					{Dir: DirServerToClient, Hex: encodeHex(zone), Note: "HC_NOTIFY_ZONESVR 0x0AC5"},
				},
			},
			{
				Name: "map",
				Records: []Record{
					{Dir: DirServerToClient, Hex: encodeHex(enter), Note: "ZC_ACCEPT_ENTER 0x0A18"},
					{Dir: DirServerToClient, Hex: encodeHex(features), Note: "ZC_TURORAN_FEATURES 0x0EF1"},
					{Dir: DirServerToClient, Hex: encodeHex(mob), Note: "ZC_NOTIFY_STANDENTRY 0x09FF"},
					{Dir: DirServerToClient, Hex: encodeHex(act), Note: "ZC_NOTIFY_ACT 0x008A"},
					{Dir: DirServerToClient, Hex: encodeHex(cast), Note: "ZC_USESKILL_ACK 0x0B1A"},
					{Dir: DirServerToClient, Hex: encodeHex(storeStart), Note: "ZC_INVENTORY_START storage"},
					{Dir: DirServerToClient, Hex: encodeHex(storeItems), Note: "ZC_INVENTORY_ITEMLIST storage"},
					{Dir: DirServerToClient, Hex: encodeHex(storeEnd), Note: "ZC_INVENTORY_END storage"},
					{Dir: DirServerToClient, Hex: encodeHex(walk), Note: "ZC_NOTIFY_PLAYERMOVE 0x0087"},
					{Dir: DirServerToClient, Hex: encodeHex(warp), Note: "ZC_NPCACK_MAPMOVE 0x0091"},
					{Dir: DirServerToClient, Hex: encodeHex(marks), Note: "ZC_ITEM_MARKS_LIST 0x0EE8 (framed, not decoded)"},
					{Dir: DirServerToClient, Hex: encodeHex(warpra), Note: "ZC_WARPRALIST 0x0EFE (framed, not decoded)"},
				},
			},
		},
	}
}

func packet(id uint16, size int) []byte {
	b := make([]byte, size)
	put16(b, 0, id)
	return b
}

func varPacket(id uint16, body []byte) []byte {
	b := make([]byte, 4+len(body))
	put16(b, 0, id)
	put16(b, 2, uint16(len(b)))
	copy(b[4:], body)
	return b
}

func put16(b []byte, off int, v uint16) []byte {
	binary.LittleEndian.PutUint16(b[off:], v)
	return b
}

func put32(b []byte, off int, v uint32) []byte {
	binary.LittleEndian.PutUint32(b[off:], v)
	return b
}

func put64(b []byte, off int, v uint64) []byte {
	binary.LittleEndian.PutUint64(b[off:], v)
	return b
}

func packPos(x, y, dir int) [3]byte {
	return [3]byte{byte(x >> 2), byte((x&3)<<6 | (y>>4)&0x3f), byte((y&0xf)<<4 | dir&0xf)}
}

func packMove(x0, y0, x1, y1 int) [6]byte {
	return [6]byte{
		byte(x0 >> 2),
		byte((x0&3)<<6 | (y0>>4)&0x3f),
		byte((y0&0xf)<<4 | (x1>>6)&0xf),
		byte((x1&0x3f)<<2 | (y1>>8)&3),
		byte(y1),
		0x88,
	}
}

func accountAcceptLogin2() []byte {
	data := make([]byte, 64+160)
	put16(data, 0, network.PacketACAcceptLogin2)
	put16(data, 2, uint16(len(data)))
	put32(data, 4, fixtureAuthCode)
	put32(data, 8, fixtureAccountID)
	data[46] = 1
	row := data[64:]
	copy(row[0:4], []byte{192, 0, 2, 10})
	put16(row, 4, fixtureCharPort)
	copy(row[6:26], "TuroranRO-TEST")
	put16(row, 26, 1)
	return data
}

func characterInfo2025(slot uint8, name string) []byte {
	b := make([]byte, 175)
	put32(b, 0, fixtureCharID)
	put64(b, 4, 1000)
	put32(b, 12, 50)
	put32(b, 24, 1)
	put64(b, 50, 100)
	put64(b, 58, 100)
	put64(b, 66, 50)
	put64(b, 74, 50)
	put16(b, 84, 0) // novice
	put16(b, 86, 1)
	put16(b, 90, 1)
	copy(b[108:132], name)
	copy(b[132:138], []byte{1, 1, 1, 1, 1, 1})
	put16(b, 138, uint16(slot))
	copy(b[142:158], fixtureMapName+".gat")
	b[174] = 1
	return b
}

func charList2025() []byte {
	data := make([]byte, 27)
	put16(data, 0, 0x006B)
	data = append(data, characterInfo2025(0, fixtureCharName)...)
	put16(data, 2, uint16(len(data)))
	return data
}

func zoneNotify2() []byte {
	data := make([]byte, 156)
	put16(data, 0, network.PacketHCNotifyZoneServer2)
	put32(data, 2, fixtureCharID)
	copy(data[6:22], fixtureMapName+".gat")
	copy(data[22:26], []byte{192, 0, 2, 10})
	put16(data, 26, fixtureMapPort)
	return data
}

func mapAcceptEnter2() []byte {
	data := make([]byte, 14)
	put16(data, 0, network.PacketZCAcceptEnter2)
	put32(data, 2, 1000)
	pos := packPos(fixtureFromX, fixtureFromY, 4)
	copy(data[6:9], pos[:])
	return data
}

func selfMoveAck(fromX, fromY, toX, toY int) []byte {
	data := make([]byte, 12)
	put16(data, 0, 0x0087)
	put32(data, 2, 1100)
	move := packMove(fromX, fromY, toX, toY)
	copy(data[6:12], move[:])
	return data
}

func idleMob2025() []byte {
	b := make([]byte, 108)
	put16(b, 0, 0x09FF)
	put16(b, 2, uint16(len(b)))
	b[4] = 5 // mob
	put32(b, 5, fixtureMobID)
	put16(b, 13, 200)
	put16(b, 23, 1002) // Poring
	base := 37
	pos := base + 26
	p := packPos(83, 105, 4)
	copy(b[pos:], p[:])
	put16(b, pos+6, 1)
	copy(b[84:], "Poring")
	return b
}

func actorAction(src, dst uint32, damage int16) []byte {
	b := make([]byte, 29)
	put16(b, 0, 0x008A)
	put32(b, 2, src)
	put32(b, 6, dst)
	put32(b, 10, 1200)
	put32(b, 14, 500)
	put32(b, 18, 400)
	put16(b, 22, uint16(damage))
	put16(b, 24, 1)
	b[26] = 0 // attack once
	return b
}

func skillCast2025(src, dst uint32, skillID uint16, delay uint32) []byte {
	b := make([]byte, 29)
	put16(b, 0, network.PacketZCUseSkillAck2025)
	put32(b, 2, src)
	put32(b, 6, dst)
	put16(b, 14, skillID)
	put32(b, 20, delay)
	return b
}

func inventoryStart(invType uint8, name string) []byte {
	b := []byte{0, 0, 0, 0, invType}
	b = append(b, name...)
	b = append(b, 0)
	put16(b, 0, network.PacketZCInventoryStart)
	put16(b, 2, uint16(len(b)))
	return b
}

func inventoryEnd(invType uint8) []byte {
	b := []byte{0, 0, invType, 0}
	put16(b, 0, network.PacketZCInventoryEnd)
	return b
}

func storageNormalList() []byte {
	entry := make([]byte, 34)
	put16(entry, 0, 1)
	put32(entry, 2, fixtureItemID)
	entry[6] = 0 // healing
	put16(entry, 7, 10)
	entry[33] = 1
	b := []byte{0, 0, 0, 0, network.InventoryTypeStorage}
	b = append(b, entry...)
	put16(b, 0, network.PacketZCInventoryItemList2025)
	put16(b, 2, uint16(len(b)))
	return b
}

func mapChangeSameServer(name string, x, y int) []byte {
	b := make([]byte, 22)
	put16(b, 0, 0x0091)
	copy(b[2:18], name+".gat")
	put16(b, 18, uint16(x))
	put16(b, 20, uint16(y))
	return b
}
