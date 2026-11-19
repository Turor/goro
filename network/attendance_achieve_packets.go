package network

import (
	"encoding/binary"
	"fmt"
)

const (
	PacketCZAttendanceReward  uint16 = 0x0aef
	PacketZCUIAction          uint16 = 0x0af0
	PacketZCAchievementList   uint16 = 0x0a23
	PacketZCAchievementUpdate uint16 = 0x0a24
	PacketCZAchievementReward uint16 = 0x0a25
	PacketZCAchievementReward uint16 = 0x0a26

	UITypeAttendance   uint8 = 7
	UITypeGrade        uint8 = 8
	CZUITypeAttendance uint8 = 5

	achievementObjectiveCount = 10
	achievementInfoSize       = 50 // id4 + done1 + obj40 + at4 + reward1
	achievementListHeader     = 22
)

type UIAction struct {
	Type int32
	Data int32
}

type Achievement struct {
	ID          uint32
	Completed   bool
	Objectives  [achievementObjectiveCount]uint32
	CompletedAt uint32
	Rewarded    bool
}

type AchievementList struct {
	Total      uint32
	Points     uint32
	Rank       uint16
	RankPoints uint32
	NextPoints uint32
	Items      []Achievement
}

func ParseUIAction(p Packet) (UIAction, bool, error) {
	if p.ID != PacketZCUIAction {
		return UIAction{}, false, nil
	}
	if len(p.Data) != 10 {
		return UIAction{}, false, fmt.Errorf("ZC_UI_ACTION: invalid length %d", len(p.Data))
	}
	return UIAction{
		Type: int32(binary.LittleEndian.Uint32(p.Data[2:6])),
		Data: int32(binary.LittleEndian.Uint32(p.Data[6:10])),
	}, true, nil
}

func parseAchievement(data []byte) Achievement {
	var a Achievement
	a.ID = binary.LittleEndian.Uint32(data[0:4])
	a.Completed = data[4] != 0
	for i := 0; i < achievementObjectiveCount; i++ {
		a.Objectives[i] = binary.LittleEndian.Uint32(data[5+i*4 : 9+i*4])
	}
	a.CompletedAt = binary.LittleEndian.Uint32(data[45:49])
	a.Rewarded = data[49] != 0
	return a
}

func ParseAchievementList(p Packet) (AchievementList, bool, error) {
	if p.ID != PacketZCAchievementList {
		return AchievementList{}, false, nil
	}
	if len(p.Data) < achievementListHeader {
		return AchievementList{}, false, fmt.Errorf("ZC_ALL_ACH_LIST: invalid length %d", len(p.Data))
	}
	wireLen := int(binary.LittleEndian.Uint16(p.Data[2:4]))
	if wireLen != len(p.Data) {
		return AchievementList{}, false, fmt.Errorf("ZC_ALL_ACH_LIST: length field %d != %d", wireLen, len(p.Data))
	}
	list := AchievementList{
		Total:      binary.LittleEndian.Uint32(p.Data[4:8]),
		Points:     binary.LittleEndian.Uint32(p.Data[8:12]),
		Rank:       binary.LittleEndian.Uint16(p.Data[12:14]),
		RankPoints: binary.LittleEndian.Uint32(p.Data[14:18]),
		NextPoints: binary.LittleEndian.Uint32(p.Data[18:22]),
	}
	body := len(p.Data) - achievementListHeader
	if body%achievementInfoSize != 0 {
		return AchievementList{}, false, fmt.Errorf("ZC_ALL_ACH_LIST: body %d", body)
	}
	count := body / achievementInfoSize
	list.Items = make([]Achievement, count)
	off := achievementListHeader
	for i := 0; i < count; i++ {
		list.Items[i] = parseAchievement(p.Data[off : off+achievementInfoSize])
		off += achievementInfoSize
	}
	return list, true, nil
}

func ParseAchievementUpdate(p Packet) (AchievementList, bool, error) {
	if p.ID != PacketZCAchievementUpdate {
		return AchievementList{}, false, nil
	}
	const hdr = 16
	if len(p.Data) != hdr+achievementInfoSize {
		return AchievementList{}, false, fmt.Errorf("ZC_ACH_UPDATE: invalid length %d", len(p.Data))
	}
	list := AchievementList{
		Points:     binary.LittleEndian.Uint32(p.Data[2:6]),
		Rank:       binary.LittleEndian.Uint16(p.Data[6:8]),
		RankPoints: binary.LittleEndian.Uint32(p.Data[8:12]),
		NextPoints: binary.LittleEndian.Uint32(p.Data[12:16]),
		Items:      []Achievement{parseAchievement(p.Data[hdr:])},
	}
	return list, true, nil
}

func ParseAchievementRewardAck(p Packet) (uint32, bool, bool, error) {
	if p.ID != PacketZCAchievementReward {
		return 0, false, false, nil
	}
	if len(p.Data) != 7 {
		return 0, false, false, fmt.Errorf("ZC_REQ_ACH_REWARD_ACK: invalid length %d", len(p.Data))
	}
	failed := p.Data[2] != 0
	id := binary.LittleEndian.Uint32(p.Data[3:7])
	return id, failed, true, nil
}

func BuildOpenUIPacket(uiType uint8) []byte {
	data := make([]byte, 3)
	binary.LittleEndian.PutUint16(data[0:2], PacketCZOpenUI)
	data[2] = uiType
	return data
}

func BuildAttendanceRewardPacket() []byte {
	data := make([]byte, 2)
	binary.LittleEndian.PutUint16(data[0:2], PacketCZAttendanceReward)
	return data
}

func BuildAchievementRewardPacket(id uint32) []byte {
	data := make([]byte, 6)
	binary.LittleEndian.PutUint16(data[0:2], PacketCZAchievementReward)
	binary.LittleEndian.PutUint32(data[2:6], id)
	return data
}

func (c *Client) SendOpenUI(uiType uint8) error {
	return c.Send(BuildOpenUIPacket(uiType))
}

func (c *Client) SendAttendanceReward() error {
	return c.Send(BuildAttendanceRewardPacket())
}

func (c *Client) SendAchievementReward(id uint32) error {
	return c.Send(BuildAchievementRewardPacket(id))
}
