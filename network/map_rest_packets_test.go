package network

import (
	"encoding/binary"
	"testing"
)

func TestParseMvpMinimapAndQuestNotify(t *testing.T) {
	data := make([]byte, 7+10)
	binary.LittleEndian.PutUint16(data[0:2], PacketZCMvpMinimap)
	binary.LittleEndian.PutUint16(data[2:4], uint16(len(data)))
	data[4] = 1
	binary.LittleEndian.PutUint16(data[5:7], 1)
	binary.LittleEndian.PutUint32(data[7:11], 150001)
	binary.LittleEndian.PutUint16(data[11:13], 1039)
	binary.LittleEndian.PutUint16(data[13:15], 80)
	binary.LittleEndian.PutUint16(data[15:17], 90)
	got, ok, err := ParseMvpMinimap(Packet{ID: PacketZCMvpMinimap, Data: data})
	if err != nil || !ok || got.Flags != 1 || len(got.Marks) != 1 || got.Marks[0].X != 80 {
		t.Fatalf("%+v ok=%t err=%v", got, ok, err)
	}
	q := make([]byte, 14)
	binary.LittleEndian.PutUint16(q[0:2], PacketZCQuestNotify)
	binary.LittleEndian.PutUint32(q[2:6], 77)
	binary.LittleEndian.PutUint16(q[6:8], 10)
	binary.LittleEndian.PutUint16(q[8:10], 20)
	qn, ok, err := ParseQuestNotifyEffect(Packet{ID: PacketZCQuestNotify, Data: q})
	if err != nil || !ok || qn.NpcID != 77 || qn.X != 10 {
		t.Fatalf("%+v", qn)
	}
	del := []byte{0x41, 0x04, 28, 0}
	d, ok, err := ParseSkillInfoDelete(Packet{ID: PacketZCSkillInfoDelete, Data: del})
	if err != nil || !ok || d.SkillID != 28 {
		t.Fatalf("%+v", d)
	}
}
