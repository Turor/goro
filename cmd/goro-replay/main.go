// Command goro-replay frames raw 20250716 dumps and packs them into a golden
// capture JSON. It never talks to a server.
package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/kivutar/goro/network"
	"github.com/kivutar/goro/sim"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	switch os.Args[1] {
	case "frame":
		os.Exit(cmdFrame(os.Args[2:]))
	case "pack":
		os.Exit(cmdPack(os.Args[2:]))
	default:
		usage()
		os.Exit(2)
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, `goro-replay frames raw PACKETVER 20250716 dumps into golden captures.

  goro-replay frame [-date 20250716] [-skip 4] FILE
  goro-replay pack -out FILE [-date 20250716] [-note TEXT] name=FILE...

Raw files are concatenated s2c bytes from one TCP connection. Char-server
dumps that start with Hercules' 4-byte account ID use -skip 4. Replace live
addresses with 192.0.2.0/24 before committing a capture.
`)
}

func cmdFrame(args []string) int {
	fs := flag.NewFlagSet("frame", flag.ExitOnError)
	date := fs.Int("date", network.ClientDate20250716, "packet client date")
	skip := fs.Int("skip", 0, "unframed prefix bytes to drop (char-server account ID is 4)")
	fs.Parse(args)
	if fs.NArg() != 1 {
		fmt.Fprintln(os.Stderr, "frame: need exactly one FILE")
		return 2
	}
	data, err := os.ReadFile(fs.Arg(0))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	stream := sim.Stream{Name: fs.Arg(0), SkipPrefix: *skip, Records: []sim.Record{
		{Dir: sim.DirServerToClient, Hex: hex.EncodeToString(data)},
	}}
	packets, err := sim.FrameStream(stream, *date)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	for _, pkt := range packets {
		fmt.Printf("0x%04X len=%d\n", pkt.ID, len(pkt.Data))
	}
	return 0
}

func cmdPack(args []string) int {
	fs := flag.NewFlagSet("pack", flag.ExitOnError)
	out := fs.String("out", "", "output capture JSON")
	date := fs.Int("date", network.ClientDate20250716, "packet client date")
	note := fs.String("note", "", "capture notes")
	fs.Parse(args)
	if *out == "" || fs.NArg() == 0 {
		fmt.Fprintln(os.Stderr, "pack: need -out FILE and name=FILE arguments")
		return 2
	}
	cap := sim.Capture{Magic: sim.Magic, ClientDate: *date, Notes: *note}
	for _, arg := range fs.Args() {
		name, path, ok := strings.Cut(arg, "=")
		if !ok || name == "" || path == "" {
			fmt.Fprintf(os.Stderr, "pack: expected name=FILE, got %q\n", arg)
			return 2
		}
		data, err := os.ReadFile(path)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
		stream := sim.Stream{Name: name}
		if name == "char" && len(data) >= 4 {
			stream.PrefixHex = hex.EncodeToString(data[:4])
			data = data[4:]
		}
		stream.Records = []sim.Record{{Dir: sim.DirServerToClient, Hex: hex.EncodeToString(data)}}
		if _, err := sim.FrameStream(stream, *date); err != nil {
			fmt.Fprintf(os.Stderr, "pack %s: %v\n", name, err)
			return 1
		}
		cap.Streams = append(cap.Streams, stream)
	}
	body, err := sim.MarshalCapture(cap)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	if err := os.WriteFile(*out, body, 0o600); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	return 0
}
