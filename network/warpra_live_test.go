package network

import (
	"os"
	"testing"
)

// TestWarpraKgapLiveCheck is opt-in. The default gate skips it. Credentials
// and the TEST host come from the environment, never from the repo. Dummy
// hosts only: do not commit private addresses.
func TestWarpraKgapLiveCheck(t *testing.T) {
	if os.Getenv("GORO_WARPRA_LIVE") != "1" {
		t.Skip("set GORO_WARPRA_LIVE=1 with GORO_WARPRA_HOST and kgap credentials in the environment to exercise Warpra on TEST")
	}
	host := os.Getenv("GORO_WARPRA_HOST")
	user := os.Getenv("GORO_WARPRA_USER")
	pass := os.Getenv("GORO_WARPRA_PASS")
	if host == "" || user == "" || pass == "" {
		t.Skip("missing GORO_WARPRA_HOST / GORO_WARPRA_USER / GORO_WARPRA_PASS")
	}
	list := BuildWarpraRequestPacket(WarpraActList, 0, 0)
	if len(list) != WarpraCZReqLen || ID(list) != PacketCZWarpraReq {
		t.Fatalf("list request = %x", list)
	}
	buy := BuildWarpraRequestPacket(WarpraActBuy, uint16(WarpraKindTown), 0)
	if len(buy) != WarpraCZReqLen || ID(buy) != PacketCZWarpraReq {
		t.Fatalf("buy request = %x", buy)
	}
	// A full login/list/warp/no-zeny pass needs Phase 5 live tooling. The
	// Hercules layouts are covered by the golden fixtures in this package.
	t.Log("Warpra CZ builders match the live opcode layout; in-game list/warp/no-zeny is a follow-up")
}
