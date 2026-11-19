package network

import (
	"encoding/binary"
	"testing"
)

func TestAttendanceAndAchievementPackets(t *testing.T) {
	open := BuildOpenUIPacket(UITypeAttendance)
	if len(open) != 3 || ID(open) != PacketCZOpenUI || open[2] != UITypeAttendance {
		t.Fatalf("open ui %v", open)
	}
	claim := BuildAttendanceRewardPacket()
	if len(claim) != 2 || ID(claim) != PacketCZAttendanceReward {
		t.Fatalf("claim")
	}
	act := make([]byte, 10)
	binary.LittleEndian.PutUint16(act[0:2], PacketZCUIAction)
	binary.LittleEndian.PutUint32(act[6:10], 3)
	got, ok, err := ParseUIAction(Packet{ID: PacketZCUIAction, Data: act})
	if err != nil || !ok || got.Data != 3 {
		t.Fatalf("action %+v", got)
	}
	info := make([]byte, achievementInfoSize)
	binary.LittleEndian.PutUint32(info[0:4], 17)
	info[4] = 1
	hdr := make([]byte, achievementListHeader)
	binary.LittleEndian.PutUint16(hdr[0:2], PacketZCAchievementList)
	binary.LittleEndian.PutUint16(hdr[2:4], uint16(achievementListHeader+achievementInfoSize))
	binary.LittleEndian.PutUint32(hdr[8:12], 50)
	listPkt := append(hdr, info...)
	list, ok, err := ParseAchievementList(Packet{ID: PacketZCAchievementList, Data: listPkt})
	if err != nil || !ok || list.Points != 50 || len(list.Items) != 1 || list.Items[0].ID != 17 || !list.Items[0].Completed {
		t.Fatalf("list %+v err=%v", list, err)
	}
	rew := BuildAchievementRewardPacket(17)
	if len(rew) != 6 || binary.LittleEndian.Uint32(rew[2:6]) != 17 {
		t.Fatalf("reward")
	}
}

func TestGradeEnchantPackets(t *testing.T) {
	sel := BuildGradeSelectPacket(9)
	if binary.LittleEndian.Uint16(sel[2:4]) != 9 {
		t.Fatalf("select")
	}
	req := BuildGradeRequestPacket(9, 1, true, 2)
	if len(req) != 14 || req[8] != 1 || binary.LittleEndian.Uint32(req[9:13]) != 2 {
		t.Fatalf("req %v", req)
	}
	body := make([]byte, gradeListHeaderLen+gradeMaterialSize)
	binary.LittleEndian.PutUint16(body[0:2], PacketZCGradeMaterials)
	binary.LittleEndian.PutUint16(body[2:4], uint16(len(body)))
	binary.LittleEndian.PutUint16(body[4:6], 9)
	binary.LittleEndian.PutUint32(body[6:10], 40)
	binary.LittleEndian.PutUint32(body[gradeListHeaderLen:gradeListHeaderLen+4], 6241)
	list, ok, err := ParseGradeMaterialList(Packet{ID: PacketZCGradeMaterials, Data: body})
	if err != nil || !ok || list.Index != 9 || list.Chance != 40 || len(list.Materials) != 1 || list.Materials[0].ItemID != 6241 {
		t.Fatalf("list %+v err=%v", list, err)
	}
}
