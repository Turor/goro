package sim

import (
	"testing"

	"github.com/kivutar/goro/network"
)

func TestFrameStreamSkipsClientToServerAndPrefixHex(t *testing.T) {
	pkt := []byte{0x6b, 0x00, 0x04, 0x00}
	stream := Stream{
		Name:      "char",
		PrefixHex: "81841e00",
		Records: []Record{
			{Dir: DirClientToServer, Hex: "0000"},
			{Dir: DirServerToClient, Hex: encodeHex(pkt)},
		},
	}
	packets, err := FrameStream(stream, network.ClientDate20250716)
	if err != nil {
		t.Fatal(err)
	}
	if len(packets) != 1 || packets[0].ID != 0x006B {
		t.Fatalf("packets = %+v", packets)
	}
}
