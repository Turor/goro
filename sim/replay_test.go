package sim

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/kivutar/goro/network"
)

func testdataPath(t *testing.T, name string) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller")
	}
	return filepath.Join(filepath.Dir(file), "testdata", name)
}

func TestGoldenLoginWalkFixtureFileMatchesBuilder(t *testing.T) {
	built, err := MarshalCapture(LoginWalkCapture())
	if err != nil {
		t.Fatal(err)
	}
	path := testdataPath(t, "login-walk.json")
	stored, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(built, stored) {
		t.Fatalf("testdata/login-walk.json drifted from LoginWalkCapture(); regenerate the file")
	}
}

func TestGoldenLoginWalkReplay(t *testing.T) {
	cap, err := LoadCapture(testdataPath(t, "login-walk.json"))
	if err != nil {
		t.Fatal(err)
	}
	if cap.ClientDate != network.ClientDate20250716 {
		t.Fatalf("client_date = %d", cap.ClientDate)
	}

	framed, err := FrameCapture(cap)
	if err != nil {
		t.Fatal(err)
	}
	wantIDs := []uint16{
		network.PacketACAcceptLogin2,
		0x006B,
		network.PacketHCNotifyZoneServer2,
		network.PacketZCAcceptEnter2,
		0x0EF1,
		0x09FF,
		0x008A,
		network.PacketZCUseSkillAck2025,
		network.PacketZCInventoryStart,
		network.PacketZCInventoryItemList2025,
		network.PacketZCInventoryEnd,
		0x0087,
		0x0091,
		0x0EE8,
		0x0EFE,
	}
	got := PacketIDs(framed)
	if len(got) != len(wantIDs) {
		t.Fatalf("framed %d packets %v, want %d %v", len(got), fmtIDs(got), len(wantIDs), fmtIDs(wantIDs))
	}
	for i, id := range wantIDs {
		if got[i] != id {
			t.Fatalf("packet %d id 0x%04X, want 0x%04X", i, got[i], id)
		}
	}

	// Byte-at-a-time framing must agree (length-table drift shows up here).
	stream := cap.Streams[2]
	var raw []byte
	for _, rec := range stream.Records {
		b, err := decodeHex(rec.Hex)
		if err != nil {
			t.Fatal(err)
		}
		raw = append(raw, b...)
	}
	framer := network.NewFramer(network.PacketLengths20250716())
	var stepped []network.Packet
	for _, b := range raw {
		pkts, err := framer.Push([]byte{b})
		if err != nil {
			t.Fatalf("byte-wise frame: %v", err)
		}
		stepped = append(stepped, pkts...)
	}
	if len(stepped) != 12 {
		t.Fatalf("map stream byte-wise packets = %d, want 12", len(stepped))
	}

	driver, err := NewDriver(cap.ClientDate, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	sawWalk := false
	for _, pkt := range framed {
		if err := driver.ApplyPacket(pkt); err != nil {
			t.Fatal(err)
		}
		if pkt.ID == 0x0087 {
			sawWalk = true
			if driver.Session.PlayerX != fixtureToX || driver.Session.PlayerY != fixtureToY {
				t.Fatalf("walk ack position = %d,%d want %d,%d", driver.Session.PlayerX, driver.Session.PlayerY, fixtureToX, fixtureToY)
			}
		}
	}
	if !sawWalk {
		t.Fatal("walk ack 0x0087 was not applied")
	}

	sess := driver.Session
	if sess.AccountID != fixtureAccountID || sess.AuthCode != fixtureAuthCode {
		t.Fatalf("login state aid=%d auth=%d", sess.AccountID, sess.AuthCode)
	}
	if len(sess.CharServers) != 1 || sess.CharServers[0].Address != fixtureCharIP || sess.CharServers[0].Port != fixtureCharPort {
		t.Fatalf("char servers = %+v", sess.CharServers)
	}
	if sess.CharServers[0].Name != "TuroranRO-TEST" {
		t.Fatalf("char server name = %q", sess.CharServers[0].Name)
	}
	if len(sess.Characters) != 1 || sess.Characters[0].Name != fixtureCharName || sess.Characters[0].ID != fixtureCharID {
		t.Fatalf("characters = %+v", sess.Characters)
	}
	if sess.Zone.MapName != fixtureMapName || sess.CharID != fixtureCharID {
		t.Fatalf("zone = %+v char=%d", sess.Zone, sess.CharID)
	}
	if sess.Zone.Address != fixtureCharIP {
		t.Fatalf("zone address = %q (must stay a documentation host)", sess.Zone.Address)
	}
	if driver.World.MapName != fixtureMapName {
		t.Fatalf("world map = %q", driver.World.MapName)
	}
	if actor, ok := driver.World.Actors[fixtureMobID]; !ok || actor.Job != 1002 {
		t.Fatalf("mob actor = %+v ok=%t", actor, ok)
	}
	if !sess.Storage.Open || len(sess.Storage.Items) != 1 || sess.Storage.Items[0].ItemID != fixtureItemID || sess.Storage.Items[0].Amount != 10 {
		t.Fatalf("storage = %+v", sess.Storage)
	}
}

func TestFixtureWalkPacketDecodes(t *testing.T) {
	cap := LoginWalkCapture()
	var walk []byte
	for _, rec := range cap.Streams[2].Records {
		if rec.Note == "ZC_NOTIFY_PLAYERMOVE 0x0087" {
			var err error
			walk, err = decodeHex(rec.Hex)
			if err != nil {
				t.Fatal(err)
			}
		}
	}
	if len(walk) != 12 {
		t.Fatalf("walk len %d", len(walk))
	}
	ack, ok, err := network.ParseSelfMoveAck(network.Packet{ID: 0x0087, Data: walk})
	if err != nil || !ok {
		t.Fatalf("parse walk ok=%t err=%v", ok, err)
	}
	if ack.FromX != fixtureFromX || ack.FromY != fixtureFromY || ack.ToX != fixtureToX || ack.ToY != fixtureToY {
		t.Fatalf("walk ack = %+v", ack)
	}
}

func TestReplayDoesNotNeedNetwork(t *testing.T) {
	driver, err := NewDriver(network.ClientDate20250716, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if err := driver.Replay(LoginWalkCapture()); err != nil {
		t.Fatal(err)
	}
	if driver.Ctx.Network.Status() != "offline" {
		t.Fatalf("network status = %q, want offline", driver.Ctx.Network.Status())
	}
}

func fmtIDs(ids []uint16) string {
	out := "["
	for i, id := range ids {
		if i > 0 {
			out += " "
		}
		out += "0x" + itoaHex(id)
	}
	return out + "]"
}

func itoaHex(id uint16) string {
	const hexdigits = "0123456789abcdef"
	return string([]byte{
		hexdigits[id>>12], hexdigits[(id>>8)&0xf],
		hexdigits[(id>>4)&0xf], hexdigits[id&0xf],
	})
}
