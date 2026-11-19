package ui

import (
	"testing"

	"github.com/kivutar/goro/session"
)

func TestIdentifiedItemNameShowsGradeAndOptionCount(t *testing.T) {
	item := session.InventoryItem{
		ItemID:     1201,
		Identified: true,
		Refine:     7,
		Grade:      4,
		Options: [5]session.ItemOption{
			{Index: 1, Value: 120},
			{Index: 0x7FFE, Value: 1},
			{Index: 2, Value: -3},
		},
	}
	got := ItemDisplayName(nil, item)
	want := "+7 [A] item 1201 [2 Option]"
	if got != want {
		t.Fatalf("name = %q, want %q", got, want)
	}
	lines := itemInfoDescriptionLines(Context{}, item)
	if len(lines) < 2 {
		t.Fatalf("option lines = %d", len(lines))
	}
	if itemInfoLinePlainText(lines[0]) != "UNKNOWN RANDOM OPTION 120" {
		t.Fatalf("first option = %q", itemInfoLinePlainText(lines[0]))
	}
	if itemInfoLinePlainText(lines[1]) != "UNKNOWN RANDOM OPTION -3" {
		t.Fatalf("second option = %q", itemInfoLinePlainText(lines[1]))
	}
}

func TestUnidentifiedItemHidesGradeAndOptions(t *testing.T) {
	item := session.InventoryItem{ItemID: 1201, Grade: 2, Options: [5]session.ItemOption{{Index: 1, Value: 5}}}
	if got := ItemDisplayName(nil, item); got != "item 1201" {
		t.Fatalf("name = %q", got)
	}
}
