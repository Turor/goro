package network

import (
	"encoding/binary"
	"net"
	"testing"
	"time"
)

// Fixtures in this file follow the Hercules PACKETVER 20250716 structs
// (login/packets_ac_struct.h, char/char.c char_mmo_char_tobuf and
// map/packets_struct.h) that TuroranRO TEST runs.

func put16(b []byte, off int, v uint16) { binary.LittleEndian.PutUint16(b[off:], v) }
func put32(b []byte, off int, v uint32) { binary.LittleEndian.PutUint32(b[off:], v) }
func put64(b []byte, off int, v uint64) { binary.LittleEndian.PutUint64(b[off:], v) }

func TestPacketLengths20250716(t *testing.T) {
	table := PacketLengthsForClientDate(ClientDate20250716)
	for id, want := range map[uint16]int{
		0x0AC4: -1, 0x0AC5: 156, 0x082D: -1, 0x006B: -1, 0x0B72: -1, 0x0B6F: 177,
		0x02EB: 13, 0x0A18: 14, 0x09FF: -1, 0x09FE: -1, 0x09FD: -1, 0x0ACB: 12,
		0x0ACC: 18, 0x0B09: -1, 0x0B39: -1, 0x0B1A: 29, 0x0B8D: -1, 0x0B8E: 18,
		// Turoran custom packets that every client receives.
		0x0EF1: 6, 0x0EF5: -1,
		0x0EE8: -1, 0x0EE9: 7, 0x0EF6: -1, 0x0EF8: -1, 0x0EF9: -1, 0x0EFA: 8,
		0x0EFB: -1, 0x0EFC: 8, 0x0EFD: -1, 0x0EFE: -1, 0x0EFF: 8,
	} {
		if got, ok := table[id]; !ok || got != want {
			t.Errorf("length 0x%04X = %d (known %t), want %d", id, got, ok, want)
		}
	}
	if old := PacketLengthsForClientDate(20080910); old[0x0EF1] != 0 || old[0x0069] != -1 {
		t.Fatalf("2008 profile changed: 0x0EF1=%d 0x0069=%d", old[0x0EF1], old[0x0069])
	}
	if !IsClientDate20250716(20250716) || IsClientDate20250716(20080910) {
		t.Fatal("IsClientDate20250716 boundaries wrong")
	}
}

func TestFramer20250716TuroranCustomPackets(t *testing.T) {
	features := make([]byte, 6)
	put16(features, 0, 0x0EF1)
	put32(features, 2, 1)
	mvp := make([]byte, 7)
	put16(mvp, 0, 0x0EF5)
	put16(mvp, 2, 7)
	mapMove := make([]byte, 22)
	put16(mapMove, 0, 0x0091)
	copy(mapMove[2:], "int_land.gat")
	stream := append(append(append([]byte{}, features...), mvp...), mapMove...)

	framer := NewFramer(PacketLengths20250716())
	var got []Packet
	for _, b := range stream { // byte by byte: lengths must hold across reads
		packets, err := framer.Push([]byte{b})
		if err != nil {
			t.Fatalf("frame error: %v", err)
		}
		got = append(got, packets...)
	}
	if len(got) != 3 || got[0].ID != 0x0EF1 || got[1].ID != 0x0EF5 || got[2].ID != 0x0091 {
		t.Fatalf("packets = %v", got)
	}
}

func TestParseAccountAcceptLogin2(t *testing.T) {
	data := make([]byte, 64+2*160)
	put16(data, 0, PacketACAcceptLogin2)
	put16(data, 2, uint16(len(data)))
	put32(data, 4, 0x11223344)
	put32(data, 8, 2000031)
	put32(data, 12, 0)
	data[46] = 1
	for i, name := range []string{"TuroranRO-TEST", "Second"} {
		row := data[64+i*160:]
		copy(row[0:4], []byte{192, 168, 11, 110})
		put16(row, 4, 6131)
		copy(row[6:26], name)
		put16(row, 26, 3)
	}
	login, err := ParseAccountAcceptLogin(Packet{ID: PacketACAcceptLogin2, Data: data})
	if err != nil {
		t.Fatal(err)
	}
	if login.AccountID != 2000031 || login.AuthCode != 0x11223344 || login.Sex != 1 {
		t.Fatalf("login = %+v", login)
	}
	if len(login.CharServer) != 2 || login.CharServer[0].Address != "192.168.11.110" || login.CharServer[0].Port != 6131 ||
		login.CharServer[0].Name != "TuroranRO-TEST" || login.CharServer[1].Name != "Second" {
		t.Fatalf("servers = %+v", login.CharServer)
	}
	if !IsAccountAcceptLogin(0x0AC4) || !IsAccountAcceptLogin(0x0069) {
		t.Fatal("IsAccountAcceptLogin")
	}
}

func TestParseAccountRefuseLoginR3(t *testing.T) {
	data := make([]byte, 26)
	put16(data, 0, PacketACRefuseLoginR3)
	put32(data, 2, 1)
	refusal, err := ParseAccountRefuseLogin(Packet{ID: PacketACRefuseLoginR3, Data: data})
	if err != nil || refusal.ErrorCode != 1 {
		t.Fatalf("refusal = %+v, %v", refusal, err)
	}
	if !IsAccountRefuseLogin(0x0B02) || !IsAccountRefuseLogin(0x083E) {
		t.Fatal("IsAccountRefuseLogin")
	}
}

func characterInfo2025Fixture(slot uint8, name string) []byte {
	b := make([]byte, characterInfo2025Size)
	put32(b, 0, 150051)
	put64(b, 4, 5_000_000_000)
	put32(b, 12, 3000)
	put32(b, 24, 50)
	put32(b, 36, 0x20)
	put64(b, 50, 125_000)
	put64(b, 58, 250_000)
	put64(b, 66, 900)
	put64(b, 74, 1800)
	put16(b, 84, 4252)
	put16(b, 86, 17)
	put16(b, 90, 1201)
	put16(b, 92, 250)
	put16(b, 96, 11)
	put16(b, 98, 2101)
	put16(b, 100, 5)
	put16(b, 102, 6)
	put16(b, 104, 3)
	put16(b, 106, 2)
	copy(b[108:132], name)
	copy(b[132:138], []byte{99, 98, 97, 96, 95, 94})
	put16(b, 138, uint16(slot))
	copy(b[142:158], "int_land.gat")
	b[174] = 1
	return b
}

func checkCharacter2025(t *testing.T, c Character, slot uint8, name string) {
	t.Helper()
	if c.ID != 150051 || c.Exp != 5_000_000_000 || c.Money != 3000 || c.JobLevel != 50 || c.Option != 0x20 {
		t.Fatalf("character header = %+v", c)
	}
	if c.HP != 125_000 || c.MaxHP != 250_000 || c.SP != 900 || c.MaxSP != 1800 {
		t.Fatalf("character vitals = %d/%d %d/%d", c.HP, c.MaxHP, c.SP, c.MaxSP)
	}
	if c.Job != 4252 || c.Hair != 17 || c.Weapon != 1201 || c.Level != 250 || c.HeadLow != 11 || c.Shield != 2101 ||
		c.HeadTop != 5 || c.HeadMid != 6 || c.HairColor != 3 || c.HeadPal != 3 || c.BodyPal != 2 {
		t.Fatalf("character look = %+v", c)
	}
	if c.Name != name || c.Slot != slot || c.Str != 99 || c.Luk != 94 {
		t.Fatalf("character identity = %+v", c)
	}
}

func TestParseCharList2025(t *testing.T) {
	// HC_ACCEPT_ENTER (0x006B): 27-byte header, 175-byte rows.
	data := make([]byte, 27)
	put16(data, 0, 0x006B)
	data = append(data, characterInfo2025Fixture(0, "kgap09")...)
	data = append(data, characterInfo2025Fixture(2, "Second")...)
	put16(data, 2, uint16(len(data)))
	list, err := ParseCharList(Packet{ID: 0x006B, Data: data})
	if err != nil {
		t.Fatal(err)
	}
	if list.Layout != "expanded_27_175" || len(list.Characters) != 2 {
		t.Fatalf("list = %s %d", list.Layout, len(list.Characters))
	}
	checkCharacter2025(t, list.Characters[0], 0, "kgap09")
	checkCharacter2025(t, list.Characters[1], 2, "Second")

	// HC_ACK_CHARINFO_PER_PAGE (0x0B72): 4-byte header.
	page := make([]byte, 4)
	put16(page, 0, PacketHCAckCharInfoPerPage)
	page = append(page, characterInfo2025Fixture(1, "Paged")...)
	put16(page, 2, uint16(len(page)))
	list, err = ParseCharList(Packet{ID: PacketHCAckCharInfoPerPage, Data: page})
	if err != nil || len(list.Characters) != 1 {
		t.Fatalf("page = %+v, %v", list, err)
	}
	checkCharacter2025(t, list.Characters[0], 1, "Paged")

	empty := []byte{0x72, 0x0B, 4, 0}
	if list, err := ParseCharList(Packet{ID: PacketHCAckCharInfoPerPage, Data: empty}); err != nil || len(list.Characters) != 0 {
		t.Fatalf("empty page = %+v, %v", list, err)
	}
}

func TestParseMakeCharacterAccept2025(t *testing.T) {
	data := append([]byte{0x6F, 0x0B}, characterInfo2025Fixture(4, "Fresh")...)
	if len(data) != 177 {
		t.Fatalf("fixture length %d", len(data))
	}
	c, err := ParseMakeCharacterAccept(Packet{ID: PacketHCAcceptMakeChar2025, Data: data})
	if err != nil {
		t.Fatal(err)
	}
	checkCharacter2025(t, c, 4, "Fresh")
}

func TestParseZoneServerNotify2(t *testing.T) {
	data := make([]byte, 156)
	put16(data, 0, PacketHCNotifyZoneServer2)
	put32(data, 2, 150051)
	copy(data[6:22], "int_land.gat")
	copy(data[22:26], []byte{192, 168, 11, 110})
	put16(data, 26, 5131)
	zone, err := ParseZoneServerNotify(Packet{ID: PacketHCNotifyZoneServer2, Data: data})
	if err != nil {
		t.Fatal(err)
	}
	if zone.CharID != 150051 || zone.MapName != "int_land" || zone.Address != "192.168.11.110" || zone.Port != 5131 {
		t.Fatalf("zone = %+v", zone)
	}
}

func TestParseMapAcceptEnter2(t *testing.T) {
	data := make([]byte, 14)
	put16(data, 0, PacketZCAcceptEnter2)
	put32(data, 2, 1234)
	pos := packPosForTest(81, 105, 4)
	copy(data[6:9], pos[:])
	enter, err := ParseMapAcceptEnter(Packet{ID: PacketZCAcceptEnter2, Data: data})
	if err != nil || enter.X != 81 || enter.Y != 105 || enter.Dir != 4 || enter.ServerTick != 1234 {
		t.Fatalf("enter = %+v, %v", enter, err)
	}
}

func packPosForTest(x, y, dir int) [3]byte {
	return [3]byte{byte(x >> 2), byte((x&3)<<6 | (y>>4)&0x3f), byte((y&0xf)<<4 | dir&0xf)}
}

func packMoveForTest(x0, y0, x1, y1 int) [6]byte {
	return [6]byte{
		byte(x0 >> 2),
		byte((x0&3)<<6 | (y0>>4)&0x3f),
		byte((y0&0xf)<<4 | (x1>>6)&0xf),
		byte((x1&0x3f)<<2 | (y1>>8)&3),
		byte(y1),
		0x88,
	}
}

// unitFixture2025 builds 0x09FF/0x09FE/0x09FD from Hercules packet_idle_unit,
// packet_spawn_unit and packet_unit_walking at PACKETVER 20250716.
func unitFixture2025(id uint16) []byte {
	size := map[uint16]int{0x09FF: 108, 0x09FE: 107, 0x09FD: 114}[id]
	b := make([]byte, size)
	put16(b, 0, id)
	put16(b, 2, uint16(size))
	b[4] = 0 // player
	put32(b, 5, 2000031)
	put32(b, 9, 150051)
	put16(b, 13, 150)
	put16(b, 15, 1)
	put16(b, 17, 2)
	put32(b, 19, 0x10)
	put16(b, 23, 4252)
	put16(b, 25, 17)
	put32(b, 27, 1201)
	put32(b, 31, 2101)
	put16(b, 35, 11)
	base := 37
	if id == 0x09FD {
		put32(b, 37, 777)
		base = 41
	}
	put16(b, base, 5)
	put16(b, base+2, 6)
	put16(b, base+4, 3)
	put16(b, base+6, 2)
	put16(b, base+8, 1)
	put16(b, base+10, 20) // robe
	put32(b, base+12, 42)
	put16(b, base+16, 7)
	put32(b, base+20, 0x400)
	b[base+25] = 1
	pos := base + 26
	switch id {
	case 0x09FD:
		move := packMoveForTest(81, 105, 85, 107)
		copy(b[pos:], move[:])
		put16(b, pos+8, 250)
		copy(b[90:], "Walker")
	case 0x09FF:
		p := packPosForTest(81, 105, 6)
		copy(b[pos:], p[:])
		b[pos+5] = 2 // sitting
		put16(b, pos+6, 250)
		copy(b[84:], "Idler")
	default:
		p := packPosForTest(81, 105, 6)
		copy(b[pos:], p[:])
		put16(b, pos+5, 250)
		copy(b[83:], "Spawn")
	}
	return b
}

func TestParseActorEntry2025(t *testing.T) {
	for _, id := range []uint16{0x09FF, 0x09FE, 0x09FD} {
		entry, ok, err := ParseActorEntry(Packet{ID: id, Data: unitFixture2025(id)})
		if err != nil || !ok {
			t.Fatalf("0x%04X: ok=%t err=%v", id, ok, err)
		}
		if entry.ID != 2000031 || entry.Speed != 150 || entry.BodyState != 1 || entry.HealthState != 2 || entry.EffectState != 0x10 {
			t.Fatalf("0x%04X state = %+v", id, entry)
		}
		if entry.Job != 4252 || entry.Head != 17 || entry.Weapon != 1201 || entry.Shield != 2101 || entry.HeadLow != 11 ||
			entry.HeadTop != 5 || entry.HeadMid != 6 || entry.HeadPal != 3 || entry.BodyPal != 2 || entry.HeadDir != 1 {
			t.Fatalf("0x%04X look = %+v", id, entry)
		}
		if entry.GuildID != 42 || entry.EmblemVersion != 7 || entry.Opt3State != 0x400 || entry.Sex != 1 || entry.Level != 250 {
			t.Fatalf("0x%04X identity = %+v", id, entry)
		}
		if id == 0x09FD {
			if !entry.Moving || entry.FromX != 81 || entry.FromY != 105 || entry.ToX != 85 || entry.ToY != 107 || entry.MoveStartTick != 777 {
				t.Fatalf("walking = %+v", entry)
			}
		} else if entry.X != 81 || entry.Y != 105 || entry.Dir != 6 || entry.Moving {
			t.Fatalf("0x%04X position = %+v", id, entry)
		}
	}
	if _, _, err := ParseActorEntry(Packet{ID: 0x09FF, Data: make([]byte, 40)}); err == nil {
		t.Fatal("short 0x09FF accepted")
	}
}

func TestParse64BitParameterAndExp(t *testing.T) {
	data := make([]byte, 12)
	put16(data, 0, PacketZCLongLongParChange)
	put16(data, 2, 1) // base exp
	put64(data, 4, 6_000_000_000)
	change, ok, err := ParseParameterChange(Packet{ID: PacketZCLongLongParChange, Data: data})
	if err != nil || !ok || change.VarID != 1 || change.Value != 6_000_000_000 {
		t.Fatalf("change = %+v ok=%t err=%v", change, ok, err)
	}
	exp := make([]byte, 18)
	put16(exp, 0, PacketZCNotifyExp2)
	put32(exp, 2, 2000031)
	put64(exp, 6, 1500)
	put16(exp, 14, 1)
	put16(exp, 16, 0)
	notify, ok, err := ParseExpNotify(Packet{ID: PacketZCNotifyExp2, Data: exp})
	if err != nil || !ok || notify.AID != 2000031 || notify.Amount != 1500 || notify.VarID != 1 {
		t.Fatalf("exp = %+v ok=%t err=%v", notify, ok, err)
	}
}

func TestAccountLoginFullLengthPasswordClearsClientType(t *testing.T) {
	packet := BuildAccountLoginPacket(AccountLogin{Version: 55, Username: "kgap09", Password: "abcdefghijklmnopqrstuvwx", ClientType: 1})
	if len(packet) != 55 || packet[54] != 0 {
		t.Fatalf("len=%d client type=%d, want 55 and 0", len(packet), packet[54])
	}
	packet = BuildAccountLoginPacket(AccountLogin{Version: 55, Username: "kgap09", Password: "short", ClientType: 1})
	if packet[54] != 1 {
		t.Fatalf("short password client type = %d, want 1", packet[54])
	}
}

// TestClientSkipsCharServerAccountIDPrefix covers the bare 4-byte account
// ID Hercules writes before HC_ACCEPT_ENTER. Its bytes here (0x0B 0x72 ...)
// spell a real 20250716 opcode, so framer resync alone would mis-frame them.
func TestClientSkipsCharServerAccountIDPrefix(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	served := make(chan error, 1)
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			served <- err
			return
		}
		defer conn.Close()
		enter := make([]byte, 17)
		if _, err := conn.Read(enter); err != nil {
			served <- err
			return
		}
		slots := make([]byte, 29)
		put16(slots, 0, 0x082D)
		put16(slots, 2, 29)
		reply := append([]byte{0x72, 0x0B, 0x09, 0x00}, slots...)
		_, err = conn.Write(reply)
		served <- err
		time.Sleep(200 * time.Millisecond)
	}()

	client := NewClient(ClientDate20250716, false)
	addr := listener.Addr().(*net.TCPAddr)
	if err := client.Connect(t.Context(), "127.0.0.1", addr.Port); err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	if err := client.SendCharServerEnter(0x00090B72, 1, 0, 1); err != nil {
		t.Fatal(err)
	}
	if err := <-served; err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(2 * time.Second)
	var packets []Packet
	for time.Now().Before(deadline) && len(packets) == 0 {
		packets = client.DrainPackets()
		time.Sleep(10 * time.Millisecond)
	}
	if errs := client.DrainErrors(); len(errs) > 0 {
		t.Fatalf("frame errors: %v", errs)
	}
	if len(packets) != 1 || packets[0].ID != 0x082D {
		t.Fatalf("packets = %v, want one 0x082D", packets)
	}
}
