package game

import (
	"testing"
	"time"

	"github.com/kivutar/goro/client"
	"github.com/kivutar/goro/network"
	"github.com/kivutar/goro/session"
)

func TestFourthClassCombatStatsKeepSignedRange(t *testing.T) {
	ctx := client.Context{Session: &session.Session{}}
	mode := NewWorldMode()
	mode.handleNetworkPacket(ctx, testParameterChangePacket(network.StatusPAtk, 300), time.Now())
	mode.handleNetworkPacket(ctx, testParameterChangePacket(network.StatusSMatk, 0xFFFFFFFB), time.Now())
	if ctx.Session.Stats.PAtk != 300 || ctx.Session.Stats.SMatk != -5 {
		t.Fatalf("combat = patk %d smatk %d, want 300 and -5", ctx.Session.Stats.PAtk, ctx.Session.Stats.SMatk)
	}
}
