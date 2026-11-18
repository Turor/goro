package companion

import (
	"fmt"
	"testing"

	"github.com/kivutar/goro/client"
	"github.com/kivutar/goro/network"
	gameui "github.com/kivutar/goro/ui"
)

func TestApplyStableResultRefreshesOnOKAndReportsErrors(t *testing.T) {
	var windows Windows
	var messages []string
	s := &System{}
	s.Bind(Bindings{Windows: &windows, Error: func(format string, args ...any) {
		messages = append(messages, fmt.Sprintf(format, args...))
	}})
	s.ApplyStableHomunculusList(network.StableHomunculusList{
		ActiveHomunID: 7,
		List:          []network.StableHomunculusEntry{{HomunID: 7, ClassID: 6001, Name: "Lif"}},
	})
	s.ApplyStablePetList(network.StablePetList{
		ActivePetID: 9,
		List:        []network.StablePetEntry{{PetID: 9, ClassID: 1002, Name: "Poring"}},
	})
	s.ApplyStableOwnedList(network.StableOwnedList{List: []network.StableOwnedEntry{{ItemID: 909, Owned: 2}}})
	if windows.Companion.HomunculusList().ActiveHomunID != 7 || windows.Companion.PetList().ActivePetID != 9 {
		t.Fatalf("lists not stored: homun=%+v pet=%+v", windows.Companion.HomunculusList(), windows.Companion.PetList())
	}

	s.ApplyStableResult(client.Context{}, network.StableResult{Result: network.StableErrHavePet, ID: 9})
	if len(messages) != 1 || messages[0] != gameui.StableResultMessage(network.StableErrHavePet) {
		t.Fatalf("error binding = %v", messages)
	}

	messages = nil
	s.ApplyStableResult(client.Context{}, network.StableResult{Result: network.StableOK})
	if len(messages) != 0 {
		t.Fatalf("ok result reported an error: %v", messages)
	}
}

func TestHandleCompanionWindowActionEvolveWithoutRecipe(t *testing.T) {
	var windows Windows
	var messages []string
	s := &System{}
	s.Bind(Bindings{Windows: &windows, Error: func(format string, args ...any) {
		messages = append(messages, fmt.Sprintf(format, args...))
	}})
	s.HandleCompanionWindowAction(client.Context{}, gameui.CompanionWindowAction{
		Kind:  gameui.CompanionActionPetEvolve,
		PetID: 12,
	})
	if len(messages) != 1 || messages[0] != "Evolution recipes are not loaded." {
		t.Fatalf("evolve leftover = %v", messages)
	}
}

func TestHandleCompanionWindowActionNilNetworkReportsError(t *testing.T) {
	var windows Windows
	var messages []string
	s := &System{}
	s.Bind(Bindings{Windows: &windows, Error: func(format string, args ...any) {
		messages = append(messages, fmt.Sprintf(format, args...))
	}})
	s.HandleCompanionWindowAction(client.Context{}, gameui.CompanionWindowAction{Kind: gameui.CompanionActionPetStore})
	if len(messages) != 1 {
		t.Fatalf("store without network = %v", messages)
	}
	s.HandleCompanionWindowAction(client.Context{}, gameui.CompanionWindowAction{Kind: gameui.CompanionActionHomunculusStore})
	if len(messages) != 2 {
		t.Fatalf("homun store without network = %v", messages)
	}
}
