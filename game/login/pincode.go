package login

import (
	"github.com/kivutar/goro/client"
	"github.com/kivutar/goro/glog"
	"github.com/kivutar/goro/network"
	gameui "github.com/kivutar/goro/ui"
)

func (m *Screen) handlePincodePacket(ctx client.Context, pkt network.Packet) bool {
	if st, ok, err := network.ParsePincodeLoginState(pkt); err != nil {
		glog.Errorf("parse PIN login 0x%04X: %v", pkt.ID, err)
		return true
	} else if ok {
		m.applyPincodeLogin(ctx, st)
		return true
	}
	if st, ok, err := network.ParsePincodeMakeState(pkt); err != nil {
		glog.Errorf("parse PIN make 0x%04X: %v", pkt.ID, err)
		return true
	} else if ok {
		m.pincodeSeed = st.Seed
		switch st.State {
		case network.PincodeMakeSuccess:
			m.pincode.ShowMessage(ctx, "PIN set. Enter it to continue.")
			m.pincode.OpenPrompt(ctx, gameui.PincodeModeAsk, "Enter your PIN.")
		case network.PincodeMakeRestrictPW:
			m.pincode.ShowMessage(ctx, "That PIN is not allowed.")
		default:
			m.pincode.ShowMessage(ctx, "Could not set PIN.")
		}
		return true
	}
	if st, ok, err := network.ParsePincodeEditState(pkt); err != nil {
		glog.Errorf("parse PIN edit 0x%04X: %v", pkt.ID, err)
		return true
	} else if ok {
		m.pincodeSeed = st.Seed
		if st.State == network.PincodeEditSuccess {
			m.pincode.ShowMessage(ctx, "PIN changed. Enter the new PIN.")
		} else {
			m.pincode.ShowMessage(ctx, "Could not change PIN.")
		}
		return true
	}
	return false
}

func (m *Screen) applyPincodeLogin(ctx client.Context, st network.PincodeLoginState) {
	m.pincodeSeed = st.Seed
	m.pincodeAccount = st.AccountID
	switch st.State {
	case network.PincodeLoginOK:
		m.pincodeRequired = false
		m.pincode.CloseUnlocked(ctx)
		if m.autoSelectCharacter(ctx) {
			return
		}
	case network.PincodeLoginAsk:
		m.pincodeRequired = true
		m.pincode.OpenPrompt(ctx, gameui.PincodeModeAsk, "Enter your PIN.")
	case network.PincodeLoginNotSet:
		m.pincodeRequired = true
		m.pincode.OpenPrompt(ctx, gameui.PincodeModeSet, "Create a 4-digit PIN.")
	case network.PincodeLoginExpired:
		m.pincodeRequired = true
		m.pincode.OpenPrompt(ctx, gameui.PincodeModeChange, "Your PIN expired. Choose a new one.")
	case network.PincodeLoginWrong:
		m.pincodeRequired = true
		m.pincode.OpenPrompt(ctx, gameui.PincodeModeAsk, "Incorrect PIN. Try again.")
	case network.PincodeLoginRestrictPW:
		m.pincodeRequired = true
		m.pincode.OpenPrompt(ctx, gameui.PincodeModeAsk, "That PIN is not allowed.")
	default:
		if st.HasFlag && st.Flag == network.PincodeLoginFlagLocked {
			m.pincodeRequired = true
			m.pincode.OpenPrompt(ctx, gameui.PincodeModeLocked, "PIN locked.")
		}
	}
}

func (m *Screen) updatePincode(ctx client.Context) bool {
	if !m.pincode.IsOpen() {
		return false
	}
	consumed := m.pincode.Update(ctx)
	action := m.pincode.PopAction()
	if ctx.Network == nil {
		return consumed
	}
	switch action.Kind {
	case gameui.PincodeActionCheck:
		_ = ctx.Network.SendPincodeCheck(m.pincodeAccount, m.pincodeSeed, action.PIN)
	case gameui.PincodeActionSet:
		_ = ctx.Network.SendPincodeFirst(m.pincodeAccount, m.pincodeSeed, action.PIN)
	case gameui.PincodeActionChange:
		_ = ctx.Network.SendPincodeChange(m.pincodeAccount, m.pincodeSeed, action.Old, action.PIN)
	}
	return consumed
}

func (m *Screen) pinBlocksCharacterSelect() bool {
	return m.pincodeRequired || m.pincode.IsOpen()
}
