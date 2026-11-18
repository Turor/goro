package sim

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/kivutar/goro/network"
)

// Magic is the capture file identifier.
const Magic = "GORO-REPLAY-1"

const (
	DirServerToClient = "s2c"
	DirClientToServer = "c2s"
)

// Capture is an in-repo golden of framed (or frameable) 20250716 traffic.
// Addresses in fixtures are documentation hosts (192.0.2.0/24), never live
// TEST or production IPs.
type Capture struct {
	Magic      string   `json:"magic"`
	ClientDate int      `json:"client_date"`
	Notes      string   `json:"notes,omitempty"`
	Streams    []Stream `json:"streams"`
}

// Stream is one TCP connection (login, char, or map).
type Stream struct {
	Name       string   `json:"name"`
	PrefixHex  string   `json:"prefix_hex,omitempty"`
	SkipPrefix int      `json:"skip_prefix,omitempty"`
	Records    []Record `json:"records"`
}

// Record is one packet on the wire. Hex is the full packet, including the
// 2-byte opcode. Only s2c records are replayed into game state.
type Record struct {
	Dir  string `json:"dir"`
	Hex  string `json:"hex"`
	Note string `json:"note,omitempty"`
}

// LoadCapture reads a JSON capture from path.
func LoadCapture(path string) (Capture, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Capture{}, err
	}
	return ParseCapture(data)
}

// ParseCapture decodes a JSON capture.
func ParseCapture(data []byte) (Capture, error) {
	var cap Capture
	if err := json.Unmarshal(data, &cap); err != nil {
		return Capture{}, fmt.Errorf("parse capture: %w", err)
	}
	if cap.Magic != Magic {
		return Capture{}, fmt.Errorf("capture magic %q, want %q", cap.Magic, Magic)
	}
	if cap.ClientDate <= 0 {
		cap.ClientDate = network.ClientDate20250716
	}
	return cap, nil
}

// MarshalCapture writes a capture as indented JSON.
func MarshalCapture(cap Capture) ([]byte, error) {
	if cap.Magic == "" {
		cap.Magic = Magic
	}
	if cap.ClientDate <= 0 {
		cap.ClientDate = network.ClientDate20250716
	}
	data, err := json.MarshalIndent(cap, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(data, '\n'), nil
}

func decodeHex(s string) ([]byte, error) {
	s = strings.Map(func(r rune) rune {
		if r == ' ' || r == '\n' || r == '\t' || r == '\r' {
			return -1
		}
		return r
	}, s)
	if s == "" {
		return nil, nil
	}
	return hex.DecodeString(s)
}

func encodeHex(data []byte) string {
	return hex.EncodeToString(data)
}

// FileRecorder appends framed s2c packets from live connections. Stream
// names are conn-1, conn-2, … (login, char, map in a typical session).
type FileRecorder struct {
	mu   sync.Mutex
	path string
	cap  Capture
}

// NewFileRecorder starts an empty capture that will be written on Close.
func NewFileRecorder(path string, clientDate int) *FileRecorder {
	if clientDate <= 0 {
		clientDate = network.ClientDate20250716
	}
	return &FileRecorder{
		path: path,
		cap: Capture{
			Magic:      Magic,
			ClientDate: clientDate,
			Notes:      "live capture; replace addresses with 192.0.2.0/24 before committing",
		},
	}
}

// OnConnect starts a new stream for a TCP reconnect.
func (r *FileRecorder) OnConnect() {
	if r == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.cap.Streams = append(r.cap.Streams, Stream{
		Name: fmt.Sprintf("conn-%d", len(r.cap.Streams)+1),
	})
}

// OnPackets records framed server-to-client packets on the current stream.
func (r *FileRecorder) OnPackets(packets []network.Packet) {
	if r == nil || len(packets) == 0 {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.cap.Streams) == 0 {
		r.cap.Streams = append(r.cap.Streams, Stream{Name: "conn-1"})
	}
	i := len(r.cap.Streams) - 1
	for _, pkt := range packets {
		r.cap.Streams[i].Records = append(r.cap.Streams[i].Records, Record{
			Dir: DirServerToClient,
			Hex: encodeHex(pkt.Data),
		})
	}
}

// Close writes the capture JSON. It is safe to call more than once.
func (r *FileRecorder) Close() error {
	if r == nil {
		return nil
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.path == "" {
		return nil
	}
	data, err := MarshalCapture(r.cap)
	if err != nil {
		return err
	}
	return os.WriteFile(r.path, data, 0o600)
}

// EqualReports whether two captures have the same magic, date, stream names
// and packet bytes. Notes may differ.
func CapturesEqual(a, b Capture) bool {
	if a.Magic != b.Magic || a.ClientDate != b.ClientDate || len(a.Streams) != len(b.Streams) {
		return false
	}
	for i := range a.Streams {
		as, bs := a.Streams[i], b.Streams[i]
		if as.Name != bs.Name || as.PrefixHex != bs.PrefixHex || as.SkipPrefix != bs.SkipPrefix || len(as.Records) != len(bs.Records) {
			return false
		}
		for j := range as.Records {
			if as.Records[j].Dir != bs.Records[j].Dir {
				return false
			}
			ah, _ := decodeHex(as.Records[j].Hex)
			bh, _ := decodeHex(bs.Records[j].Hex)
			if !bytes.Equal(ah, bh) {
				return false
			}
		}
	}
	return true
}
