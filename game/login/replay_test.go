package login

import (
	"testing"

	"github.com/kivutar/goro/client"
	"github.com/kivutar/goro/network"
	"github.com/kivutar/goro/session"
	worldstate "github.com/kivutar/goro/world"
)

func TestApplySessionPacketLoginFlow(t *testing.T) {
	sess := session.New()
	world := worldstate.New()
	ctx := client.Context{Session: sess, World: world}

	loginPkt := make([]byte, 64+160)
	put16 := func(b []byte, off int, v uint16) { b[off] = byte(v); b[off+1] = byte(v >> 8) }
	put32 := func(b []byte, off int, v uint32) {
		b[off] = byte(v)
		b[off+1] = byte(v >> 8)
		b[off+2] = byte(v >> 16)
		b[off+3] = byte(v >> 24)
	}
	put16(loginPkt, 0, network.PacketACAcceptLogin2)
	put16(loginPkt, 2, uint16(len(loginPkt)))
	put32(loginPkt, 4, 9)
	put32(loginPkt, 8, 2000001)
	copy(loginPkt[64:68], []byte{192, 0, 2, 10})
	put16(loginPkt[64:], 4, 6121)
	copy(loginPkt[64+6:64+26], "TuroranRO-TEST")

	ok, err := ApplySessionPacket(ctx, network.Packet{ID: network.PacketACAcceptLogin2, Data: loginPkt})
	if err != nil || !ok {
		t.Fatalf("login ok=%t err=%v", ok, err)
	}
	if sess.AccountID != 2000001 || sess.AuthCode != 9 {
		t.Fatalf("session login = aid=%d auth=%d", sess.AccountID, sess.AuthCode)
	}
	if len(sess.CharServers) != 1 || sess.CharServers[0].Address != "192.0.2.10" {
		t.Fatalf("servers = %+v", sess.CharServers)
	}

	if ok, err := ApplySessionPacket(ctx, network.Packet{ID: 0x0087, Data: make([]byte, 12)}); ok || err != nil {
		t.Fatalf("walk packet handled as login: ok=%t err=%v", ok, err)
	}
}
