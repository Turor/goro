package sim

import (
	"fmt"

	"github.com/kivutar/goro/network"
)

// FrameStream runs the 20250716 (or capture-date) length table over a
// stream's concatenated s2c bytes. Char-server captures may set SkipPrefix
// or PrefixHex for the bare 4-byte account ID Hercules writes before the
// first packet.
func FrameStream(stream Stream, clientDate int) ([]network.Packet, error) {
	if clientDate <= 0 {
		clientDate = network.ClientDate20250716
	}
	if stream.PrefixHex != "" {
		if _, err := decodeHex(stream.PrefixHex); err != nil {
			return nil, fmt.Errorf("stream %s prefix: %w", stream.Name, err)
		}
	}
	var raw []byte
	for i, rec := range stream.Records {
		if rec.Dir != "" && rec.Dir != DirServerToClient {
			continue
		}
		pkt, err := decodeHex(rec.Hex)
		if err != nil {
			return nil, fmt.Errorf("stream %s record %d: %w", stream.Name, i, err)
		}
		raw = append(raw, pkt...)
	}
	if skip := stream.SkipPrefix; skip > 0 {
		if skip > len(raw) {
			return nil, fmt.Errorf("stream %s skip_prefix %d longer than payload %d", stream.Name, skip, len(raw))
		}
		raw = raw[skip:]
	}
	framer := network.NewFramer(network.PacketLengthsForClientDate(clientDate))
	packets, err := framer.Push(raw)
	if err != nil {
		return packets, fmt.Errorf("stream %s: %w", stream.Name, err)
	}
	return packets, nil
}

// FrameCapture frames every s2c stream. Partial results are returned with
// the first framing error.
func FrameCapture(cap Capture) ([]network.Packet, error) {
	var all []network.Packet
	for _, stream := range cap.Streams {
		packets, err := FrameStream(stream, cap.ClientDate)
		all = append(all, packets...)
		if err != nil {
			return all, err
		}
	}
	return all, nil
}

// PacketIDs returns the opcodes in order.
func PacketIDs(packets []network.Packet) []uint16 {
	ids := make([]uint16, len(packets))
	for i, pkt := range packets {
		ids[i] = pkt.ID
	}
	return ids
}
