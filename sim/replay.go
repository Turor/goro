package sim

import (
	"fmt"
	"os"
	"time"

	"github.com/kivutar/goro/client"
	"github.com/kivutar/goro/config"
	"github.com/kivutar/goro/game"
	"github.com/kivutar/goro/game/login"
	"github.com/kivutar/goro/network"
	"github.com/kivutar/goro/res"
	"github.com/kivutar/goro/session"
	worldstate "github.com/kivutar/goro/world"
)

// Driver replays a capture through the framer, packet decoders, and game
// state. It does not open a socket.
type Driver struct {
	ClientDate int
	Session    *session.Session
	World      *worldstate.World
	Mode       *game.WorldMode
	Ctx        client.Context
	Framed     []network.Packet
}

// NewDriver builds an offline replay driver. dataDir may be empty; a temp
// directory is used so res.Manager can start without archives.
func NewDriver(clientDate int, dataDir string) (*Driver, error) {
	if clientDate <= 0 {
		clientDate = network.ClientDate20250716
	}
	if dataDir == "" {
		dir, err := os.MkdirTemp("", "goro-replay-")
		if err != nil {
			return nil, err
		}
		dataDir = dir
	}
	resources, err := res.NewManager(dataDir)
	if err != nil {
		return nil, err
	}
	sess := session.New()
	world := worldstate.New()
	mode := game.NewWorldMode()
	d := &Driver{
		ClientDate: clientDate,
		Session:    sess,
		World:      world,
		Mode:       mode,
	}
	d.Ctx = client.Context{
		Config:    config.Config{Headless: true, Packet: config.PacketConfig{ClientDate: clientDate, ClientDateExplicit: true}},
		Resources: resources,
		Session:   sess,
		World:     world,
		Network:   network.NewClient(clientDate, false),
	}
	return d, nil
}

// Replay frames every s2c stream and applies packets to session/world.
func (d *Driver) Replay(cap Capture) error {
	if d == nil {
		return fmt.Errorf("nil replay driver")
	}
	if cap.ClientDate == 0 {
		cap.ClientDate = d.ClientDate
	}
	for _, stream := range cap.Streams {
		packets, err := FrameStream(stream, cap.ClientDate)
		d.Framed = append(d.Framed, packets...)
		if err != nil {
			return err
		}
		for _, pkt := range packets {
			if err := d.ApplyPacket(pkt); err != nil {
				return fmt.Errorf("stream %s packet 0x%04X: %w", stream.Name, pkt.ID, err)
			}
		}
	}
	return nil
}

// ApplyPacket updates session and world from one already-framed s2c packet.
func (d *Driver) ApplyPacket(pkt network.Packet) error {
	if d == nil {
		return fmt.Errorf("nil replay driver")
	}
	handled, err := login.ApplySessionPacket(d.Ctx, pkt)
	if err != nil {
		return err
	}
	if handled {
		return nil
	}
	if _, stop := d.Mode.ApplyReplayPacket(d.Ctx, pkt, time.Now()); stop {
		// Map changes stop a live frame; replay keeps going so later
		// packets in the same capture still update state.
		return nil
	}
	return nil
}
