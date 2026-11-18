package login

import (
	"github.com/kivutar/goro/client"
	"github.com/kivutar/goro/game/actors"
	"github.com/kivutar/goro/network"
	"github.com/kivutar/goro/session"
)

// ApplySessionPacket updates session and world state from a login, character,
// zone, or map-accept packet without touching UI. It reports whether the
// packet belongs to that login flow.
func ApplySessionPacket(ctx client.Context, pkt network.Packet) (bool, error) {
	if ctx.Session == nil {
		return false, nil
	}
	if network.IsAccountAcceptLogin(pkt.ID) {
		login, err := network.ParseAccountAcceptLogin(pkt)
		if err != nil {
			return true, err
		}
		ctx.Session.AccountID = login.AccountID
		ctx.Session.AuthCode = login.AuthCode
		ctx.Session.UserLevel = login.UserLevel
		ctx.Session.Sex = login.Sex
		ctx.Session.CharServers = convertCharServers(login.CharServer)
		ctx.Session.CharServerIndex = 0
		return true, nil
	}
	if network.IsCharList(pkt.ID) {
		list, err := network.ParseCharList(pkt)
		if err != nil {
			return true, err
		}
		if pkt.ID == network.PacketHCAckCharInfoPerPage && len(list.Characters) == 0 {
			return true, nil
		}
		ctx.Session.Characters = convertCharacters(list.Characters)
		if len(ctx.Session.Characters) > 0 && ctx.Session.Selected.ID == 0 {
			ctx.Session.SelectCharacter(ctx.Session.Characters[0])
		}
		return true, nil
	}
	if network.IsZoneServerNotify(pkt.ID) {
		zone, err := network.ParseZoneServerNotify(pkt)
		if err != nil {
			return true, err
		}
		for _, character := range ctx.Session.Characters {
			if character.ID == zone.CharID {
				ctx.Session.SelectCharacter(character)
				break
			}
		}
		ctx.Session.CharID = zone.CharID
		ctx.Session.Zone = session.ZoneServer{
			Address: zone.Address,
			Port:    zone.Port,
			MapName: zone.MapName,
		}
		if ctx.World != nil {
			ctx.World.MapName = zone.MapName
		}
		return true, nil
	}
	if network.IsMapAcceptEnter(pkt.ID) {
		enter, err := network.ParseMapAcceptEnter(pkt)
		if err != nil {
			return true, err
		}
		actors.ApplyMapAcceptEnter(ctx, enter)
		return true, nil
	}
	return false, nil
}
