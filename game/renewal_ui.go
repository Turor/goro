package game

import (
	"fmt"

	"github.com/kivutar/goro/client"
	"github.com/kivutar/goro/glog"
	"github.com/kivutar/goro/network"
	gameui "github.com/kivutar/goro/ui"
)

func (m *WorldMode) handleRenewalUIPacket(ctx client.Context, pkt network.Packet) bool {
	if ui, ok, err := network.ParseUIOpen(pkt); err != nil {
		glog.Errorf("parse UI open 0x%04X: %v", pkt.ID, err)
		return true
	} else if ok {
		m.applyUIOpen(ctx, ui)
		return true
	}
	if open, err := network.ParseOpenRefiningUI(pkt); err != nil {
		glog.Errorf("parse refine UI 0x%04X: %v", pkt.ID, err)
		return true
	} else if open {
		m.ui.refineUI.Open(ctx)
		return true
	}
	if list, ok, err := network.ParseRefineMaterialList(pkt); err != nil {
		glog.Errorf("parse refine materials 0x%04X: %v", pkt.ID, err)
		return true
	} else if ok {
		m.ui.refineUI.SetMaterials(ctx, list)
		return true
	}
	if result, ok, err := network.ParseRefineBroadcast(pkt); err != nil {
		glog.Errorf("parse refine broadcast 0x%04X: %v", pkt.ID, err)
		return true
	} else if ok {
		status := "failed"
		if result.Status == 0 {
			status = "succeeded"
		}
		m.ui.console.AddSystemMessage("%s refined item %d to +%d (%s).", result.Name, result.ItemID, result.RefineLevel, status)
		return true
	}
	if resp, ok, err := network.ParseEnchantResponse(pkt); err != nil {
		glog.Errorf("parse enchant response 0x%04X: %v", pkt.ID, err)
		return true
	} else if ok {
		msg := fmt.Sprintf("Enchant result %d (item %d)", resp.MsgID, resp.ItemID)
		m.ui.enchantUI.SetMessage(ctx, msg)
		m.ui.console.AddSystemMessage("%s", msg)
		return true
	}
	if m.handleBankPacket(ctx, pkt) {
		return true
	}
	if m.handleRodexPacket(ctx, pkt) {
		return true
	}
	if m.handleAttendanceAchievePackets(ctx, pkt) {
		return true
	}
	if m.handleGradePacket(ctx, pkt) {
		return true
	}
	if open, ok, err := network.ParseSearchStoreOpen(pkt); err != nil {
		glog.Errorf("parse search store open 0x%04X: %v", pkt.ID, err)
		return true
	} else if ok {
		_ = open
		m.ui.searchStore.Open(ctx)
		return true
	}
	if ack, ok, err := network.ParseSearchStoreAck(pkt); err != nil {
		glog.Errorf("parse search store 0x%04X: %v", pkt.ID, err)
		return true
	} else if ok {
		m.ui.searchStore.SetAck(ctx, ack)
		return true
	}
	if cfg, ok, err := network.ParseConfigFlag(pkt); err != nil {
		glog.Errorf("parse config 0x%04X: %v", pkt.ID, err)
		return true
	} else if ok {
		if ctx.Session != nil {
			switch cfg.Type {
			case network.ConfigPetFeed:
				ctx.Session.PetAutoFeed = cfg.Value != 0
			case network.ConfigHomunFeed:
				ctx.Session.HomunAutoFeed = cfg.Value != 0
			}
		}
		return true
	}
	if add, ok, err := network.ParseRodexAddAck(pkt); err != nil {
		glog.Errorf("parse rodex add 0x%04X: %v", pkt.ID, err)
		return true
	} else if ok {
		if add.Result == 0 {
			m.ui.rodex.SetMessage(ctx, "Item attached.")
		} else {
			m.ui.rodex.SetMessage(ctx, "Could not attach item.")
		}
		return true
	}
	if open, ok, err := network.ParseRouletteOpen(pkt); err != nil {
		glog.Errorf("parse roulette open 0x%04X: %v", pkt.ID, err)
		return true
	} else if ok {
		m.ui.roulette.SetOpen(ctx, open)
		return true
	}
	if spin, ok, err := network.ParseRouletteSpin(pkt); err != nil {
		glog.Errorf("parse roulette spin 0x%04X: %v", pkt.ID, err)
		return true
	} else if ok {
		m.ui.roulette.SetSpin(ctx, spin)
		return true
	}
	if items, ok, err := network.ParseMarketOpen(pkt); err != nil {
		glog.Errorf("parse market 0x%04X: %v", pkt.ID, err)
		return true
	} else if ok {
		m.ui.market.OpenList(ctx, items)
		return true
	}
	if items, ok, err := network.ParseBarterList(pkt); err != nil {
		glog.Errorf("parse barter 0x%04X: %v", pkt.ID, err)
		return true
	} else if ok {
		m.ui.barter.OpenList(ctx, items)
		return true
	}
	if info, ok, err := network.ParseClanInfo(pkt); err != nil {
		glog.Errorf("parse clan 0x%04X: %v", pkt.ID, err)
		return true
	} else if ok {
		m.ui.clan.SetInfo(ctx, info)
		return true
	}
	if info, ok, err := network.ParseClanOnline(pkt); err != nil {
		glog.Errorf("parse clan online 0x%04X: %v", pkt.ID, err)
		return true
	} else if ok {
		m.ui.clan.SetInfo(ctx, info)
		return true
	}
	if msg, ok, err := network.ParseClanChat(pkt); err != nil {
		glog.Errorf("parse clan chat 0x%04X: %v", pkt.ID, err)
		return true
	} else if ok {
		m.ui.console.AddMessage("%s", msg)
		return true
	}
	if open, ok, err := network.ParseReformOpen(pkt); err != nil {
		glog.Errorf("parse reform open 0x%04X: %v", pkt.ID, err)
		return true
	} else if ok {
		m.ui.reform.Open(ctx, open.ItemID)
		return true
	}
	if ack, ok, err := network.ParseReformAck(pkt); err != nil {
		glog.Errorf("parse reform ack 0x%04X: %v", pkt.ID, err)
		return true
	} else if ok {
		m.ui.reform.SetAck(ctx, ack)
		return true
	}
	if open, ok, err := network.ParseLaphineOpen(pkt); err != nil {
		glog.Errorf("parse laphine open 0x%04X: %v", pkt.ID, err)
		return true
	} else if ok {
		m.ui.laphine.Open(ctx, open)
		return true
	}
	if result, ok, err := network.ParseLaphineAck(pkt); err != nil {
		glog.Errorf("parse laphine ack 0x%04X: %v", pkt.ID, err)
		return true
	} else if ok {
		m.ui.laphine.SetResult(ctx, result)
		return true
	}
	if list, ok, err := network.ParseEdenList(pkt); err != nil {
		glog.Errorf("parse eden list 0x%04X: %v", pkt.ID, err)
		return true
	} else if ok {
		m.ui.eden.SetList(ctx, list)
		return true
	}
	if shop, ok, err := network.ParseEdenShop(pkt); err != nil {
		glog.Errorf("parse eden shop 0x%04X: %v", pkt.ID, err)
		return true
	} else if ok {
		m.ui.eden.SetShop(ctx, shop)
		return true
	}
	if res, ok, err := network.ParseEdenResult(pkt); err != nil {
		glog.Errorf("parse eden result 0x%04X: %v", pkt.ID, err)
		return true
	} else if ok {
		m.ui.eden.SetResult(ctx, res)
		return true
	}
	return false
}

func (m *WorldMode) applyUIOpen(ctx client.Context, ui network.UIOpen) {
	switch ui.Type {
	case network.UITypeBank:
		m.openBank(ctx)
	case network.UITypeEnchant:
		m.ui.enchantUI.Open(ctx, ui.Data)
	case network.UITypeAttendance:
		m.ui.attendance.Open(ctx, ui.Data)
	case network.UITypeGrade:
		m.ui.gradeUI.Open(ctx)
	case network.UITypeStyling:
		m.ui.styling.Open(ctx)
	default:
		glog.Debugf("unhandled UI type %d data=%d", ui.Type, ui.Data)
	}
}

func (m *WorldMode) handleBankPacket(ctx client.Context, pkt network.Packet) bool {
	if ack, ok, err := network.ParseBankOpenAck(pkt); err != nil {
		glog.Errorf("parse bank open 0x%04X: %v", pkt.ID, err)
		return true
	} else if ok {
		if pkt.ID == network.PacketZCBankClose {
			m.ui.bank.Close()
			return true
		}
		if ack.Reason != network.BankReasonOK {
			m.ui.console.AddErrorMessage("Bank is unavailable.")
			return true
		}
		if !m.ui.bank.IsOpen() {
			m.ui.bank.Open(ctx, 0, bankZeny(ctx))
		}
		if ctx.Network != nil {
			_ = ctx.Network.SendBankCheck(ctx.Session.AccountID)
		}
		return true
	}
	if bal, ok, err := network.ParseBankCheck(pkt); err != nil {
		glog.Errorf("parse bank check 0x%04X: %v", pkt.ID, err)
		return true
	} else if ok {
		if !m.ui.bank.IsOpen() {
			m.ui.bank.Open(ctx, bal.Vault, bankZeny(ctx))
		} else {
			m.ui.bank.SetBalances(ctx, bal.Vault, bankZeny(ctx), "")
		}
		return true
	}
	if bal, ok, err := network.ParseBankAck(pkt); err != nil {
		glog.Errorf("parse bank ack 0x%04X: %v", pkt.ID, err)
		return true
	} else if ok {
		if ctx.Session != nil {
			ctx.Session.Inventory.Zeny = int64(bal.Zeny)
		}
		msg := ""
		if bal.Reason != network.BankReasonOK {
			msg = "Bank transaction failed."
		}
		m.ui.bank.SetBalances(ctx, bal.Vault, bal.Zeny, msg)
		return true
	}
	return false
}

func (m *WorldMode) handleRodexPacket(ctx client.Context, pkt network.Packet) bool {
	if icon, ok, err := network.ParseRodexIcon(pkt); err != nil {
		glog.Errorf("parse rodex icon 0x%04X: %v", pkt.ID, err)
		return true
	} else if ok {
		m.ui.rodex.SetUnread(icon.Show)
		return true
	}
	if list, ok, err := network.ParseRodexMailList(pkt); err != nil {
		glog.Errorf("parse rodex list 0x%04X: %v", pkt.ID, err)
		return true
	} else if ok {
		if !m.ui.rodex.IsOpen() {
			m.ui.rodex.OpenMailbox(ctx)
		}
		m.ui.rodex.SetList(ctx, list)
		return true
	}
	if msg, ok, err := network.ParseRodexRead(pkt); err != nil {
		glog.Errorf("parse rodex read 0x%04X: %v", pkt.ID, err)
		return true
	} else if ok {
		m.ui.rodex.SetRead(ctx, msg)
		return true
	}
	if result, ok, err := network.ParseRodexSendResult(pkt); err != nil {
		glog.Errorf("parse rodex send 0x%04X: %v", pkt.ID, err)
		return true
	} else if ok {
		if result == 0 {
			m.ui.rodex.SetMessage(ctx, "Mail sent.")
			if ctx.Network != nil {
				_ = ctx.Network.SendRodexRefresh()
			}
		} else {
			m.ui.rodex.SetMessage(ctx, "Mail could not be sent.")
		}
		return true
	}
	if ack, ok, err := network.ParseRodexMailIDAck(pkt); err != nil {
		glog.Errorf("parse rodex ack 0x%04X: %v", pkt.ID, err)
		return true
	} else if ok {
		_ = ack
		if ctx.Network != nil {
			_ = ctx.Network.SendRodexRefresh()
		}
		return true
	}
	return false
}

func (m *WorldMode) openBank(ctx client.Context) {
	if ctx.Network == nil || ctx.Session == nil {
		return
	}
	if err := ctx.Network.SendBankOpen(ctx.Session.AccountID); err != nil {
		m.ui.console.AddErrorMessage("Bank request failed.")
	}
}

func (m *WorldMode) updateBank(ctx client.Context) bool {
	if !m.ui.bank.IsOpen() {
		return false
	}
	consumed := m.ui.bank.Update(ctx)
	action := m.ui.bank.PopAction()
	if ctx.Network == nil || ctx.Session == nil {
		return consumed
	}
	switch action.Kind {
	case gameui.BankActionDeposit:
		_ = ctx.Network.SendBankDeposit(ctx.Session.AccountID, action.Amount)
	case gameui.BankActionWithdraw:
		_ = ctx.Network.SendBankWithdraw(ctx.Session.AccountID, action.Amount)
	case gameui.BankActionClose:
		_ = ctx.Network.SendBankClose(ctx.Session.AccountID)
	}
	return consumed
}

func (m *WorldMode) useRodex(ctx client.Context) bool {
	return ctx.Config.Packet.ClientDate >= network.ClientDate20250716
}

func bankZeny(ctx client.Context) int32 {
	if ctx.Session == nil {
		return 0
	}
	if ctx.Session.Inventory.Zeny > 1<<31-1 {
		return 1<<31 - 1
	}
	return int32(ctx.Session.Inventory.Zeny)
}

func (m *WorldMode) handleAttendanceAchievePackets(ctx client.Context, pkt network.Packet) bool {
	if act, ok, err := network.ParseUIAction(pkt); err != nil {
		glog.Errorf("parse UI action 0x%04X: %v", pkt.ID, err)
		return true
	} else if ok {
		m.ui.attendance.SetProgress(ctx, int(act.Data))
		return true
	}
	if list, ok, err := network.ParseAchievementList(pkt); err != nil {
		glog.Errorf("parse achievement list 0x%04X: %v", pkt.ID, err)
		return true
	} else if ok {
		m.ui.achievements.SetList(ctx, list)
		return true
	}
	if upd, ok, err := network.ParseAchievementUpdate(pkt); err != nil {
		glog.Errorf("parse achievement update 0x%04X: %v", pkt.ID, err)
		return true
	} else if ok {
		cur := m.ui.achievements.List
		cur.Points = upd.Points
		cur.Rank = upd.Rank
		cur.RankPoints = upd.RankPoints
		cur.NextPoints = upd.NextPoints
		if len(upd.Items) == 1 {
			found := false
			for i, a := range cur.Items {
				if a.ID == upd.Items[0].ID {
					cur.Items[i] = upd.Items[0]
					found = true
					break
				}
			}
			if !found {
				cur.Items = append(cur.Items, upd.Items[0])
			}
		}
		m.ui.achievements.SetList(ctx, cur)
		return true
	}
	if id, failed, ok, err := network.ParseAchievementRewardAck(pkt); err != nil {
		glog.Errorf("parse achievement reward 0x%04X: %v", pkt.ID, err)
		return true
	} else if ok {
		if failed {
			m.ui.achievements.SetMessage(ctx, "Reward failed.")
		} else {
			m.ui.achievements.SetMessage(ctx, "Reward claimed.")
		}
		_ = id
		return true
	}
	return false
}

func (m *WorldMode) handleGradePacket(ctx client.Context, pkt network.Packet) bool {
	if list, ok, err := network.ParseGradeMaterialList(pkt); err != nil {
		glog.Errorf("parse grade materials 0x%04X: %v", pkt.ID, err)
		return true
	} else if ok {
		if !m.ui.gradeUI.IsOpen() {
			m.ui.gradeUI.Open(ctx)
		}
		m.ui.gradeUI.SetMaterials(ctx, list)
		return true
	}
	if ack, ok, err := network.ParseGradeAck(pkt); err != nil {
		glog.Errorf("parse grade ack 0x%04X: %v", pkt.ID, err)
		return true
	} else if ok {
		m.ui.gradeUI.SetAck(ctx, ack)
		return true
	}
	if bc, ok, err := network.ParseGradeBroadcast(pkt); err != nil {
		glog.Errorf("parse grade broadcast 0x%04X: %v", pkt.ID, err)
		return true
	} else if ok {
		status := "failed"
		if bc.Status != 0 {
			status = "succeeded"
		}
		m.ui.console.AddSystemMessage("%s graded item %d to %d (%s).", bc.Name, bc.ItemID, bc.Grade, status)
		return true
	}
	return false
}
