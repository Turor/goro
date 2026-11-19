package game

import (
	"encoding/binary"
	"testing"
	"time"

	"github.com/kivutar/goro/network"
	gameui "github.com/kivutar/goro/ui"
)

func TestStablePacketsApplyToCompanionWindow(t *testing.T) {
	ctx := chatShortcutTestContext(t)
	m := NewWorldMode()

	homun := make([]byte, network.StableHomunculusHeaderLen+network.StableHomunculusEntryLen)
	binary.LittleEndian.PutUint16(homun[0:2], network.PacketZCTuroranStableHomun)
	binary.LittleEndian.PutUint16(homun[2:4], uint16(len(homun)))
	binary.LittleEndian.PutUint32(homun[4:8], 11)
	homun[8] = 1
	binary.LittleEndian.PutUint32(homun[9:13], 11)
	binary.LittleEndian.PutUint16(homun[13:15], 6001)
	copy(homun[27:51], []byte("Lif"))

	if next, stop := m.handleNetworkPacket(ctx, network.Packet{ID: network.PacketZCTuroranStableHomun, Data: homun}, time.Now()); next != nil || stop {
		t.Fatal("stable homunculus list stopped the frame")
	}
	if got := m.ui.companions.Companion.HomunculusList(); got.ActiveHomunID != 11 || len(got.List) != 1 || got.List[0].Name != "Lif" {
		t.Fatalf("homunculus list = %+v", got)
	}

	pet := make([]byte, network.StablePetHeaderLen+network.StablePetEntryLen)
	binary.LittleEndian.PutUint16(pet[0:2], network.PacketZCTuroranStablePet)
	binary.LittleEndian.PutUint16(pet[2:4], uint16(len(pet)))
	binary.LittleEndian.PutUint32(pet[4:8], 22)
	pet[8] = 1
	binary.LittleEndian.PutUint32(pet[9:13], 22)
	binary.LittleEndian.PutUint32(pet[13:17], 1002)
	copy(pet[38:62], []byte("Poring"))

	if next, stop := m.handleNetworkPacket(ctx, network.Packet{ID: network.PacketZCTuroranStablePet, Data: pet}, time.Now()); next != nil || stop {
		t.Fatal("stable pet list stopped the frame")
	}
	if got := m.ui.companions.Companion.PetList(); got.ActivePetID != 22 || len(got.List) != 1 || got.List[0].Name != "Poring" {
		t.Fatalf("pet list = %+v", got)
	}

	owned := make([]byte, network.StableOwnedHeaderLen+network.StableOwnedEntryLen)
	binary.LittleEndian.PutUint16(owned[0:2], network.PacketZCTuroranStableOwned)
	binary.LittleEndian.PutUint16(owned[2:4], uint16(len(owned)))
	binary.LittleEndian.PutUint16(owned[4:6], 1)
	binary.LittleEndian.PutUint32(owned[6:10], 909)
	binary.LittleEndian.PutUint32(owned[10:14], 4)
	if next, stop := m.handleNetworkPacket(ctx, network.Packet{ID: network.PacketZCTuroranStableOwned, Data: owned}, time.Now()); next != nil || stop {
		t.Fatal("stable owned list stopped the frame")
	}

	result := make([]byte, 8)
	binary.LittleEndian.PutUint16(result[0:2], network.PacketZCTuroranStableResult)
	binary.LittleEndian.PutUint16(result[2:4], uint16(network.StableErrHavePet))
	binary.LittleEndian.PutUint32(result[4:8], 22)
	if next, stop := m.handleNetworkPacket(ctx, network.Packet{ID: network.PacketZCTuroranStableResult, Data: result}, time.Now()); next != nil || stop {
		t.Fatal("stable result stopped the frame")
	}
	messages := m.ui.console.Messages()
	if len(messages) != 1 || messages[0].Text != gameui.StableResultMessage(network.StableErrHavePet) {
		t.Fatalf("result message = %+v", messages)
	}

	features := make([]byte, 6)
	binary.LittleEndian.PutUint16(features[0:2], 0x0EF1)
	if next, stop := m.handleNetworkPacket(ctx, network.Packet{ID: 0x0EF1, Data: features}, time.Now()); next != nil || stop {
		t.Fatal("incoming 0x0EF1 features packet stopped the frame")
	}
	if got := m.ui.companions.Companion.HomunculusList(); got.ActiveHomunID != 11 {
		t.Fatal("incoming 0x0EF1 was parsed as a stable homunculus list")
	}
}
