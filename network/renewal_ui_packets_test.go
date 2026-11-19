package network

import (
	"encoding/binary"
	"testing"
)

func TestPincodeEncryptRoundTrip(t *testing.T) {
	const seed uint32 = 0x1234
	plain := "5821"
	enc := EncryptPincode(seed, plain)
	if len(enc) != 4 {
		t.Fatalf("encrypted len=%d", len(enc))
	}
	if got := decryptPincode(seed, enc); got != plain {
		t.Fatalf("decrypt(%q) = %q, want %q (enc=%q)", enc, got, plain, enc)
	}
	pkt := BuildPincodeCheckPacket(2000001, enc)
	if len(pkt) != 10 || ID(pkt) != PacketCZPincodeCheck {
		t.Fatalf("check packet %x", pkt)
	}
	if binary.LittleEndian.Uint32(pkt[2:6]) != 2000001 {
		t.Fatalf("aid")
	}
	change := BuildPincodeChangePacket(7, "1111", "2222")
	if len(change) != 14 || ID(change) != PacketCZPincodeChange {
		t.Fatalf("change len=%d", len(change))
	}
}

func TestParsePincodeLoginState(t *testing.T) {
	data := make([]byte, 12)
	binary.LittleEndian.PutUint16(data[0:2], PacketHCPincodeLoginState)
	binary.LittleEndian.PutUint32(data[2:6], 99)
	binary.LittleEndian.PutUint32(data[6:10], 42)
	binary.LittleEndian.PutUint16(data[10:12], PincodeLoginAsk)
	st, ok, err := ParsePincodeLoginState(Packet{ID: PacketHCPincodeLoginState, Data: data})
	if err != nil || !ok || st.Seed != 99 || st.AccountID != 42 || st.State != PincodeLoginAsk {
		t.Fatalf("parse %+v ok=%v err=%v", st, ok, err)
	}
	data2 := make([]byte, 13)
	binary.LittleEndian.PutUint16(data2[0:2], PacketHCPincodeLoginState2)
	binary.LittleEndian.PutUint32(data2[2:6], 1)
	binary.LittleEndian.PutUint32(data2[6:10], 2)
	binary.LittleEndian.PutUint16(data2[10:12], PincodeLoginWrong)
	data2[12] = byte(PincodeLoginFlagWrong)
	st, ok, err = ParsePincodeLoginState(Packet{ID: PacketHCPincodeLoginState2, Data: data2})
	if err != nil || !ok || !st.HasFlag || st.Flag != PincodeLoginFlagWrong {
		t.Fatalf("login2 %+v ok=%v err=%v", st, ok, err)
	}
}

func TestBankPackets(t *testing.T) {
	open := BuildBankOpenPacket(9)
	if len(open) != 6 || ID(open) != PacketCZBankOpen {
		t.Fatalf("open %x", open)
	}
	dep := BuildBankDepositPacket(9, 500)
	if len(dep) != 10 || int32(binary.LittleEndian.Uint32(dep[6:10])) != 500 {
		t.Fatalf("deposit")
	}
	ack := make([]byte, 16)
	binary.LittleEndian.PutUint16(ack[0:2], PacketZCBankDeposit)
	binary.LittleEndian.PutUint16(ack[2:4], 0)
	binary.LittleEndian.PutUint64(ack[4:12], 1500)
	binary.LittleEndian.PutUint32(ack[12:16], 800)
	bal, ok, err := ParseBankAck(Packet{ID: PacketZCBankDeposit, Data: ack})
	if err != nil || !ok || bal.Vault != 1500 || bal.Zeny != 800 {
		t.Fatalf("ack %+v err=%v", bal, err)
	}
	ui := make([]byte, 11)
	binary.LittleEndian.PutUint16(ui[0:2], PacketZCUIOpen2)
	ui[2] = UITypeEnchant
	binary.LittleEndian.PutUint64(ui[3:11], 77)
	opened, ok, err := ParseUIOpen(Packet{ID: PacketZCUIOpen2, Data: ui})
	if err != nil || !ok || opened.Type != UITypeEnchant || opened.Data != 77 {
		t.Fatalf("ui %+v err=%v", opened, err)
	}
}

func TestRefineUIPackets(t *testing.T) {
	sel := BuildRefiningSelectItemPacket(3)
	if binary.LittleEndian.Uint16(sel[2:4]) != 3 {
		t.Fatalf("select index %v", sel)
	}
	req := BuildReqRefiningPacket(3, 985, true)
	if len(req) != 9 || req[8] != 1 || binary.LittleEndian.Uint32(req[4:8]) != 985 {
		t.Fatalf("req %v", req)
	}
	body := make([]byte, refineMaterialHeaderLen+refineMaterialEntrySize)
	binary.LittleEndian.PutUint16(body[0:2], PacketZCRefiningMaterials)
	binary.LittleEndian.PutUint16(body[2:4], uint16(len(body)))
	binary.LittleEndian.PutUint16(body[4:6], 3)
	body[6] = 1
	binary.LittleEndian.PutUint32(body[7:11], 985)
	body[11] = 80
	binary.LittleEndian.PutUint32(body[12:16], 1000)
	list, ok, err := ParseRefineMaterialList(Packet{ID: PacketZCRefiningMaterials, Data: body})
	if err != nil || !ok || list.Index != 3 || !list.Blessing || len(list.Materials) != 1 || list.Materials[0].Chance != 80 {
		t.Fatalf("list %+v err=%v", list, err)
	}
}

func TestEnchantUIPackets(t *testing.T) {
	pkt := BuildRandomEnchantPacket(12, 4)
	if len(pkt) != 12 || binary.LittleEndian.Uint16(pkt[10:12]) != 4 {
		t.Fatalf("random %v", pkt)
	}
	perf := BuildPerfectEnchantPacket(12, 4, 6223)
	if len(perf) != 16 || int32(binary.LittleEndian.Uint32(perf[12:16])) != 6223 {
		t.Fatalf("perfect")
	}
	resp := make([]byte, 10)
	binary.LittleEndian.PutUint16(resp[0:2], PacketZCResponseEnchant)
	binary.LittleEndian.PutUint32(resp[2:6], 9)
	binary.LittleEndian.PutUint32(resp[6:10], 1101)
	got, ok, err := ParseEnchantResponse(Packet{ID: PacketZCResponseEnchant, Data: resp})
	if err != nil || !ok || got.MsgID != 9 || got.ItemID != 1101 {
		t.Fatalf("%+v", got)
	}
}

func TestRodexPackets(t *testing.T) {
	icon := []byte{0xe7, 0x09, 1}
	got, ok, err := ParseRodexIcon(Packet{ID: PacketZCRodexIcon, Data: icon})
	if err != nil || !ok || !got.Show {
		t.Fatalf("icon")
	}
	open := BuildRodexOpenPacket()
	if len(open) != 26 || ID(open) != PacketCZRodexOpen {
		t.Fatalf("open len=%d", len(open))
	}
	title := "Hello"
	entryFixed := 1 + 8 + 1 + 1 + 24 + 4 + 2
	list := make([]byte, rodexMailListHeader+entryFixed+len(title))
	binary.LittleEndian.PutUint16(list[0:2], PacketZCRodexMailList)
	binary.LittleEndian.PutUint16(list[2:4], uint16(len(list)))
	list[4] = 1
	off := rodexMailListHeader
	list[off] = 1
	binary.LittleEndian.PutUint64(list[off+1:off+9], 99)
	copy(list[off+11:off+35], []byte("Alice"))
	binary.LittleEndian.PutUint16(list[off+39:off+41], uint16(len(title)))
	copy(list[off+41:], title)
	parsed, ok, err := ParseRodexMailList(Packet{ID: PacketZCRodexMailList, Data: list})
	if err != nil || !ok || !parsed.End || len(parsed.Mail) != 1 || parsed.Mail[0].MailID != 99 || parsed.Mail[0].Title != title {
		t.Fatalf("list %+v err=%v", parsed, err)
	}
	body := "hi"
	read := make([]byte, rodexReadHeaderLen+len(body))
	binary.LittleEndian.PutUint16(read[0:2], PacketZCRodexRead)
	binary.LittleEndian.PutUint16(read[2:4], uint16(len(read)))
	read[4] = 1
	binary.LittleEndian.PutUint64(read[5:13], 99)
	binary.LittleEndian.PutUint16(read[13:15], uint16(len(body)))
	binary.LittleEndian.PutUint64(read[15:23], 250)
	copy(read[24:], body)
	msg, ok, err := ParseRodexRead(Packet{ID: PacketZCRodexRead, Data: read})
	if err != nil || !ok || msg.MailID != 99 || msg.Zeny != 250 || msg.Body != body {
		t.Fatalf("read %+v err=%v", msg, err)
	}
	send := BuildRodexSendPacket("Bob", "Alice", "T", "B", 10, 0)
	if ID(send) != PacketCZRodexSend || binary.LittleEndian.Uint16(send[2:4]) != uint16(len(send)) {
		t.Fatalf("send")
	}
}
