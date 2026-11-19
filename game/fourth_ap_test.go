package game

import (
	"testing"
	"time"

	"github.com/kivutar/goro/client"
	"github.com/kivutar/goro/network"
	"github.com/kivutar/goro/session"
)

func TestFourthClassAPAndCombatStats(t *testing.T) {
	ctx := client.Context{Session: &session.Session{}}
	mode := NewWorldMode()
	mode.handleNetworkPacket(ctx, testParameterChangePacket(network.StatusAP, 40), time.Now())
	mode.handleNetworkPacket(ctx, testParameterChangePacket(network.StatusMaxAP, 200), time.Now())
	mode.handleNetworkPacket(ctx, testParameterChangePacket(network.StatusPAtk, 18), time.Now())
	mode.handleNetworkPacket(ctx, testParameterChangePacket(network.StatusSMatk, 9), time.Now())
	mode.handleNetworkPacket(ctx, testParameterChangePacket(network.StatusRes, 7), time.Now())
	mode.handleNetworkPacket(ctx, testParameterChangePacket(network.StatusMRes, 6), time.Now())
	mode.handleNetworkPacket(ctx, testParameterChangePacket(network.StatusHPlus, 5), time.Now())
	mode.handleNetworkPacket(ctx, testParameterChangePacket(network.StatusCRate, 4), time.Now())
	if ctx.Session.Vitals.AP != 40 || ctx.Session.Vitals.MaxAP != 200 {
		t.Fatalf("ap = %d/%d, want 40/200", ctx.Session.Vitals.AP, ctx.Session.Vitals.MaxAP)
	}
	stats := ctx.Session.Stats
	if stats.PAtk != 18 || stats.SMatk != 9 || stats.Res != 7 || stats.MRes != 6 || stats.HPlus != 5 || stats.CRate != 4 {
		t.Fatalf("combat stats = %+v", stats)
	}
}
