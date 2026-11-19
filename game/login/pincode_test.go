package login

import (
	"encoding/binary"
	"testing"

	"github.com/kivutar/goro/client"
	"github.com/kivutar/goro/network"
	"github.com/kivutar/goro/session"
)

func TestPincodeAskBlocksCharacterSelect(t *testing.T) {
	mode := &Screen{}
	ctx := client.Context{Session: &session.Session{AccountID: 42}}
	data := make([]byte, 12)
	binary.LittleEndian.PutUint16(data[0:2], network.PacketHCPincodeLoginState)
	binary.LittleEndian.PutUint32(data[2:6], 99)
	binary.LittleEndian.PutUint32(data[6:10], 42)
	binary.LittleEndian.PutUint16(data[10:12], network.PincodeLoginAsk)
	if !mode.handlePincodePacket(ctx, network.Packet{ID: network.PacketHCPincodeLoginState, Data: data}) {
		t.Fatal("PIN packet not handled")
	}
	if !mode.pinBlocksCharacterSelect() {
		t.Fatal("PIN did not block character select")
	}
	mode.submitSelectedCharacter(ctx)
	if mode.charSelectPending {
		t.Fatal("select went out while PIN was required")
	}
	ok := make([]byte, 12)
	binary.LittleEndian.PutUint16(ok[0:2], network.PacketHCPincodeLoginState)
	binary.LittleEndian.PutUint32(ok[2:6], 99)
	binary.LittleEndian.PutUint32(ok[6:10], 42)
	binary.LittleEndian.PutUint16(ok[10:12], network.PincodeLoginOK)
	if !mode.handlePincodePacket(ctx, network.Packet{ID: network.PacketHCPincodeLoginState, Data: ok}) {
		t.Fatal("PIN OK not handled")
	}
	if mode.pinBlocksCharacterSelect() {
		t.Fatal("PIN still blocking after OK")
	}
}
