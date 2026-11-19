package sim

import (
	"path/filepath"
	"testing"

	"github.com/kivutar/goro/network"
)

func TestFileRecorderWritesCapture(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cap.json")
	rec := NewFileRecorder(path, network.ClientDate20250716)
	rec.OnConnect()
	rec.OnPackets([]network.Packet{{ID: 0x0AC4, Data: []byte{0xc4, 0x0a, 0x04, 0x00}}})
	rec.OnConnect()
	rec.OnPackets([]network.Packet{{ID: 0x006B, Data: []byte{0x6b, 0x00, 0x04, 0x00}}})
	if err := rec.Close(); err != nil {
		t.Fatal(err)
	}
	cap, err := LoadCapture(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(cap.Streams) != 2 || cap.Streams[0].Name != "conn-1" || cap.Streams[1].Name != "conn-2" {
		t.Fatalf("streams = %+v", cap.Streams)
	}
	if len(cap.Streams[0].Records) != 1 || cap.Streams[0].Records[0].Dir != DirServerToClient {
		t.Fatalf("records = %+v", cap.Streams[0].Records)
	}
}

func TestParseCaptureRejectsBadMagic(t *testing.T) {
	if _, err := ParseCapture([]byte(`{"magic":"NOPE","client_date":20250716,"streams":[]}`)); err == nil {
		t.Fatal("bad magic accepted")
	}
}
