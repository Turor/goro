package res

import "testing"

func TestParseClientInfo(t *testing.T) {
	info, err := ParseClientInfo([]byte(`
<clientinfo>
  <servicetype>korea</servicetype>
  <connection>
    <display>TestRO</display>
    <address>127.0.0.1</address>
    <port>6900</port>
    <version>55</version>
    <langtype>0</langtype>
    <aid><admin>2000002</admin></aid>
    <yellow>
      <admin>2000003</admin>
      <admin>2000002</admin>
      <admin>not-a-number</admin>
    </yellow>
  </connection>
</clientinfo>`))
	if err != nil {
		t.Fatal(err)
	}
	if info.ServiceType != "korea" {
		t.Fatalf("service type = %q", info.ServiceType)
	}
	if len(info.Connections) != 1 {
		t.Fatalf("connections = %d", len(info.Connections))
	}
	conn := info.Connections[0]
	if conn.Display != "TestRO" || conn.Address != "127.0.0.1" || conn.Port != 6900 || conn.Version != 55 {
		t.Fatalf("unexpected connection: %+v", conn)
	}
	wantAdmins := []uint32{2000003, 2000002}
	if len(conn.AdminList) != len(wantAdmins) {
		t.Fatalf("admin list = %v, want %v", conn.AdminList, wantAdmins)
	}
	for i := range wantAdmins {
		if conn.AdminList[i] != wantAdmins[i] {
			t.Fatalf("admin list = %v, want %v", conn.AdminList, wantAdmins)
		}
	}
}

func TestParseClientInfoPacketVersion(t *testing.T) {
	info, err := ParseClientInfo([]byte(`
<clientinfo>
  <connection>
    <display>TEST</display>
    <address>192.0.2.10</address>
    <port>6910</port>
    <version>55</version>
    <packet_version>20250716</packet_version>
  </connection>
  <connection>
    <display>Classic</display>
    <address>192.0.2.11</address>
    <port>6900</port>
    <version>55</version>
    <packet_version>20080910</packet_version>
  </connection>
  <connection>
    <display>NoDate</display>
    <address>192.0.2.12</address>
    <port>6900</port>
    <version>55</version>
    <packet_version>not-a-date</packet_version>
  </connection>
</clientinfo>`))
	if err != nil {
		t.Fatal(err)
	}
	if len(info.Connections) != 3 {
		t.Fatalf("connections = %d", len(info.Connections))
	}
	if info.Connections[0].Version != 55 || info.Connections[0].PacketVersion != 20250716 {
		t.Fatalf("slot 0 = %+v", info.Connections[0])
	}
	if info.Connections[1].PacketVersion != 20080910 {
		t.Fatalf("slot 1 packet version = %d", info.Connections[1].PacketVersion)
	}
	if info.Connections[2].PacketVersion != 0 {
		t.Fatalf("invalid packet_version should be ignored, got %d", info.Connections[2].PacketVersion)
	}
	if info.PacketVersionForSlot(0) != 20250716 || info.PacketVersionForSlot(1) != 20080910 {
		t.Fatalf("PacketVersionForSlot = %d, %d", info.PacketVersionForSlot(0), info.PacketVersionForSlot(1))
	}
	if info.PacketVersionForSlot(-1) != 0 || info.PacketVersionForSlot(3) != 0 {
		t.Fatal("out-of-range slots should return 0")
	}
}

func TestParseClientInfoPacketVersionOptional(t *testing.T) {
	info, err := ParseClientInfo([]byte(`
<clientinfo>
  <connection>
    <display>Local</display>
    <address>127.0.0.1</address>
  </connection>
</clientinfo>`))
	if err != nil {
		t.Fatal(err)
	}
	if info.Connections[0].PacketVersion != 0 || info.PacketVersionForSlot(0) != 0 {
		t.Fatalf("missing packet_version = %d", info.Connections[0].PacketVersion)
	}
}
