package ui

import (
	"strings"
	"testing"

	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
	"github.com/kivutar/goro/session"
)

func TestGearTabsSplitAllocationAndKeepComposite(t *testing.T) {
	window := &EquipmentWindow{}
	ctx := Context{
		Session: &session.Session{Stats: session.Stats{
			Str: 12, Points: 3, Attack: 40, AttackBonus: 2,
			Pow: 4, TraitPoints: 2, PAtk: 9,
		}},
		ScreenW: 1280,
		ScreenH: 720,
	}
	window.Toggle(ctx)
	labels := characterLabels(window.widgetTree(ctx, nil, nil))
	for _, want := range []string{"Points 3", "STR", "Composite", "ATK 40 + 2", "P.ATK"} {
		if !strings.Contains(labels, want) {
			t.Fatalf("basic tab missing %q: %q", want, labels)
		}
	}
	if strings.Contains(labels, "POW") {
		t.Fatal("basic tab shows 4th-class allocation")
	}
	window.showStatTab(gearStatTabTrait)
	labels = characterLabels(window.content)
	for _, want := range []string{"Trait Points 2", "POW", "Composite", "P.ATK"} {
		if !strings.Contains(labels, want) {
			t.Fatalf("4th tab missing %q: %q", want, labels)
		}
	}
	if strings.Contains(labels, "STR") {
		t.Fatal("4th tab still shows basic allocation")
	}
	_, height := gearWindowSize()
	if height < ROWindowTitleHeight+gearStatsContentHeight()+ROWindowFooterHeight {
		t.Fatalf("gear height %d does not fit the displayed attributes", height)
	}
}

func TestResizedWindowGrowsTheRightPane(t *testing.T) {
	window := &EquipmentWindow{}
	ctx := Context{Session: &session.Session{}, ScreenW: 1280, ScreenH: 900}
	window.Toggle(ctx)
	width, height := gearWindowSize()
	window.resizeTo(ctx, width+80, height+50)
	frame, ok := window.content.(*windowFrame)
	if !ok {
		t.Fatal("gear content is not a window frame")
	}
	frame.Layout(widget.NewContext(), geometry.Tight(geometry.Sz(float32(window.width), float32(window.height))))
	grown := false
	for _, box := range windowBoxes(frame) {
		if box.Style().ExplicitWidth >= float32(gearStatsColumnWidth()+80) {
			grown = true
		}
	}
	if !grown {
		t.Fatal("resized window did not widen the status column")
	}
}
