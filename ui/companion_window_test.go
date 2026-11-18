package ui

import (
	"testing"

	"github.com/kivutar/goro/network"
	"github.com/kivutar/goro/session"
)

func TestHomunculusLineageMatchesStableClient(t *testing.T) {
	tests := []struct {
		classID, prevClass, want int
		name                     string
	}{
		{6001, 0, 0, "Lif"},
		{6005, 0, 0, "Lif"},
		{6009, 0, 0, "Lif"},
		{6013, 0, 0, "Lif"},
		{6002, 0, 1, "Amistr"},
		{6003, 0, 2, "Filir"},
		{6004, 0, 3, "Vanilmirth"},
		{6052, 0, 4, "Eleanor"},
		{6048, 6001, 0, "Eira from Lif"},
		{6049, 6002, 1, "Bayeri from Amistr"},
		{6051, 6003, 2, "Dieter from Filir"},
		{6050, 6004, 3, "Sera from Vanilmirth"},
		{6048, 0, 0, "Eira without prev"},
		{1, 0, -1, "unknown"},
	}
	for _, tt := range tests {
		if got := HomunculusLineage(tt.classID, tt.prevClass); got != tt.want {
			t.Fatalf("%s: lineage(%d,%d)=%d, want %d", tt.name, tt.classID, tt.prevClass, got, tt.want)
		}
		if tt.want >= 0 && HomunculusLineageName(tt.want) == "Homunculus" {
			t.Fatalf("%s: missing lineage name", tt.name)
		}
	}
}

func TestStableHomunculusIntimacyTextMatchesRobrowser(t *testing.T) {
	tests := []struct {
		value int
		want  string
	}{
		{0, "Awkward"},
		{9999, "Awkward"},
		{10000, "Shy"},
		{24999, "Shy"},
		{25000, "Neutral"},
		{74999, "Neutral"},
		{75000, "Cordial"},
		{90999, "Cordial"},
		{91000, "Loyal"},
	}
	for _, tt := range tests {
		if got := StableHomunculusIntimacyText(tt.value); got != tt.want {
			t.Fatalf("intimacy %d = %q, want %q", tt.value, got, tt.want)
		}
	}
}

func TestStablePetStatusAndResultMessages(t *testing.T) {
	if got := stablePetStatus(network.StablePetEntry{Flags: network.StableFlagInvEgg}); got != "Legacy egg on this character" {
		t.Fatalf("egg status = %q", got)
	}
	if got := stablePetStatus(network.StablePetEntry{Flags: network.StableFlagOut, OwnerName: "kgap01"}); got != "Out · kgap01" {
		t.Fatalf("out status = %q", got)
	}
	if got := stablePetStatus(network.StablePetEntry{Flags: network.StableFlagInStable}); got != "In account stable" {
		t.Fatalf("stable status = %q", got)
	}
	if got := StableResultMessage(network.StableOK); got != "" {
		t.Fatalf("ok message = %q", got)
	}
	if got := StableResultMessage(network.StableErrHavePet); got != "Return your current pet to its egg first, or Store it." {
		t.Fatalf("have-pet message = %q", got)
	}
	if got := StableResultMessage(99); got != "Stable action failed." {
		t.Fatalf("unknown message = %q", got)
	}
}

func TestCompanionWindowRemembersTabAndAppliesStableLists(t *testing.T) {
	ctx := Context{Session: &session.Session{Homunculus: session.Companion{Active: true, Name: "Lif", Level: 10}}, UIManager: NewManager(), ScreenW: 1024, ScreenH: 768}
	w := &CompanionWindow{}
	w.SetTab(CompanionTabStable)
	w.OpenWindow(ctx)
	if !w.IsOpen() || w.Tab() != CompanionTabStable {
		t.Fatal("did not open on the remembered tab")
	}
	if action := w.PopAction(); action.Kind != CompanionActionRequestLists {
		t.Fatalf("open action = %+v", action)
	}
	w.SetHomunculusList(network.StableHomunculusList{
		ActiveHomunID: 1,
		List:          []network.StableHomunculusEntry{{HomunID: 1, ClassID: 6001, Name: "Lif", Level: 10, Intimacy: 91000}},
	})
	w.SetPetList(network.StablePetList{
		ActivePetID: 9,
		List: []network.StablePetEntry{{
			PetID: 9, ClassID: 1002, Name: "Poring", Level: 5, Flags: network.StableFlagOut,
			Buff: "Active: +1 LUK", Intimate: 900,
		}},
	})
	w.SetOwnedCounts(network.StableOwnedList{List: []network.StableOwnedEntry{{ItemID: 909, Owned: 4}}})
	w.Update(ctx, network.PetProperty{}, false)
	if w.homunList.ActiveHomunID != 1 || len(w.petList.List) != 1 || w.owned[909] != 4 {
		t.Fatalf("lists not stored: homun=%+v pets=%+v owned=%v", w.homunList, w.petList, w.owned)
	}
	w.SetPetList(network.StablePetList{})
	if w.selectedPetID != 0 {
		t.Fatal("cleared pet list should drop the selection")
	}
}

func TestClampCompanionTab(t *testing.T) {
	if ClampCompanionTab(-1) != CompanionTabPet || ClampCompanionTab(4) != CompanionTabPet {
		t.Fatal("out of range tab")
	}
	if ClampCompanionTab(int(CompanionTabStable)) != CompanionTabStable {
		t.Fatal("stable tab")
	}
}

func TestCompanionClassNameFallsBackToID(t *testing.T) {
	if got := CompanionClassName(6001); got != "LIF" {
		t.Fatalf("lif name = %q", got)
	}
	if got := CompanionClassName(999999); got != "#999999" {
		t.Fatalf("unknown name = %q", got)
	}
}

func TestCompanionPetExpText(t *testing.T) {
	if got := companionPetExpText(network.StablePetEntry{Exp: 10, NextExp: 0}); got != "Max" {
		t.Fatalf("max = %q", got)
	}
	if got := companionPetExpText(network.StablePetEntry{Exp: 50, NextExp: 100}); got != "50 / 100 (50%)" {
		t.Fatalf("pct = %q", got)
	}
}
