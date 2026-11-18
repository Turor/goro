package game

import (
	"time"

	"github.com/kivutar/goro/client"
	"github.com/kivutar/goro/network"
)

// ApplyReplayPacket applies one already-framed server packet the same way a
// live WorldMode session would. Golden replay tests use this so fixtures
// exercise the framer, the decoders, and game state together without a
// network.
func (m *WorldMode) ApplyReplayPacket(ctx client.Context, pkt network.Packet, now time.Time) (Mode, bool) {
	if m == nil {
		return nil, false
	}
	if now.IsZero() {
		now = time.Now()
	}
	return m.handleNetworkPacket(ctx, pkt, now)
}
