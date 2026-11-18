package companion

import (
	"github.com/kivutar/goro/client"
	"github.com/kivutar/goro/glog"
	"github.com/kivutar/goro/network"
	gameui "github.com/kivutar/goro/ui"
)

func (m *System) OpenCompanionWindow(ctx client.Context) {
	m.bindings.Windows.Companion.OpenWindow(ctx)
	m.requestStableLists(ctx)
}

func (m *System) ToggleCompanionWindow(ctx client.Context) {
	if m.bindings.Windows.Companion.IsOpen() {
		m.bindings.Windows.Companion.Close()
		m.bindings.Windows.Companion.Publish(ctx)
		return
	}
	m.OpenCompanionWindow(ctx)
}

func (m *System) OpenCompanionTab(ctx client.Context, tab gameui.CompanionTab) {
	m.bindings.Windows.Companion.OpenTab(ctx, tab)
	m.requestStableLists(ctx)
}

func (m *System) ApplyStableHomunculusList(list network.StableHomunculusList) {
	m.bindings.Windows.Companion.SetHomunculusList(list)
}

func (m *System) ApplyStablePetList(list network.StablePetList) {
	m.bindings.Windows.Companion.SetPetList(list)
}

func (m *System) ApplyStableOwnedList(list network.StableOwnedList) {
	m.bindings.Windows.Companion.SetOwnedCounts(list)
}

func (m *System) ApplyStableResult(ctx client.Context, result network.StableResult) {
	if result.Result == network.StableOK {
		m.requestStableLists(ctx)
		return
	}
	msg := gameui.StableResultMessage(result.Result)
	if msg == "" {
		msg = "Stable action failed."
	}
	if m.bindings.Error != nil {
		m.bindings.Error("%s", msg)
	}
	glog.Debugf("stable result code=%d id=%d", result.Result, result.ID)
}

func (m *System) HandleCompanionWindowAction(ctx client.Context, action gameui.CompanionWindowAction) {
	switch action.Kind {
	case gameui.CompanionActionRequestLists:
		m.requestStableLists(ctx)
	case gameui.CompanionActionHomunculusStore:
		m.sendStableHomunculus(ctx, 0)
	case gameui.CompanionActionHomunculusSwap:
		m.sendStableHomunculus(ctx, action.HomunculusID)
	case gameui.CompanionActionPetStore:
		m.sendStablePet(ctx, network.StablePetStore, 0)
	case gameui.CompanionActionPetSummon:
		m.sendStablePet(ctx, network.StablePetSummon, action.PetID)
	case gameui.CompanionActionPetDepositEgg:
		m.sendStablePet(ctx, network.StablePetDepositEgg, action.PetID)
	case gameui.CompanionActionPetToEgg:
		m.sendStablePet(ctx, network.StablePetToEgg, action.PetID)
	case gameui.CompanionActionPetFeed:
		m.sendStablePet(ctx, network.StablePetFeed, action.PetID)
	case gameui.CompanionActionPetUnequip:
		m.sendStablePet(ctx, network.StablePetUnequip, action.PetID)
	case gameui.CompanionActionPetEvolve:
		if action.EvolveEggID == 0 {
			if m.bindings.Error != nil {
				m.bindings.Error("Evolution recipes are not loaded.")
			}
			return
		}
		m.sendStablePet(ctx, network.StablePetEvolve, action.EvolveEggID)
	case gameui.CompanionActionPetPerformance:
		m.sendClassicPet(ctx, network.PetCommandPerformance)
	case gameui.CompanionActionPetClassicFeed:
		m.sendClassicPet(ctx, network.PetCommandFeed)
	case gameui.CompanionActionHomunculusSkill:
		m.HandleHomunculusInfoAction(ctx, gameui.HomunculusInfoAction{Kind: gameui.HomunculusInfoActionSkill})
	case gameui.CompanionActionHomunculusFeed:
		m.HandleHomunculusInfoAction(ctx, gameui.HomunculusInfoAction{Kind: gameui.HomunculusInfoActionFeed})
	case gameui.CompanionActionHomunculusDelete:
		m.HandleHomunculusInfoAction(ctx, gameui.HomunculusInfoAction{Kind: gameui.HomunculusInfoActionDelete})
	case gameui.CompanionActionMercenarySkill:
		m.HandleMercenaryInfoAction(ctx, gameui.MercenaryInfoAction{Kind: gameui.MercenaryInfoActionSkill})
	case gameui.CompanionActionMercenaryDelete:
		m.HandleMercenaryInfoAction(ctx, gameui.MercenaryInfoAction{Kind: gameui.MercenaryInfoActionDelete})
	}
}

func (m *System) requestStableLists(ctx client.Context) {
	if ctx.Network == nil {
		return
	}
	if err := ctx.Network.SendStableOpen(); err != nil {
		if m.bindings.Error != nil {
			m.bindings.Error("Stable request failed.")
		}
		glog.Warnf("stable open failed: %v", err)
	}
}

func (m *System) sendStableHomunculus(ctx client.Context, homunID uint32) {
	if ctx.Network == nil {
		if m.bindings.Error != nil {
			m.bindings.Error("Stable homunculus request failed: not connected.")
		}
		return
	}
	if err := ctx.Network.SendStableHomunculus(homunID); err != nil {
		if m.bindings.Error != nil {
			m.bindings.Error("Stable homunculus request failed.")
		}
		glog.Warnf("stable homunculus failed id=%d: %v", homunID, err)
	}
}

func (m *System) sendStablePet(ctx client.Context, action uint16, id uint32) {
	if ctx.Network == nil {
		if m.bindings.Error != nil {
			m.bindings.Error("Stable pet request failed: not connected.")
		}
		return
	}
	if err := ctx.Network.SendStablePet(action, id); err != nil {
		if m.bindings.Error != nil {
			m.bindings.Error("Stable pet request failed.")
		}
		glog.Warnf("stable pet failed action=%d id=%d: %v", action, id, err)
	}
}

func (m *System) sendClassicPet(ctx client.Context, command uint8) {
	if ctx.Network == nil {
		if m.bindings.Error != nil {
			m.bindings.Error("Pet command failed: not connected.")
		}
		return
	}
	if err := ctx.Network.SendPetCommand(command); err != nil {
		if m.bindings.Error != nil {
			m.bindings.Error("Pet command failed.")
		}
		glog.Warnf("pet command failed command=%d: %v", command, err)
	}
}
