package ui

import (
	"strings"
	"testing"

	"github.com/kivutar/goro/session"
)

func TestGearWindowIncludesEquipmentAndStats(t *testing.T) {
	window := &EquipmentWindow{}
	ctx := Context{
		Session: &session.Session{Stats: session.Stats{Str: 12, Points: 3, Attack: 40, AttackBonus: 2}},
		ScreenW: 1280,
		ScreenH: 720,
	}
	window.Toggle(ctx)
	width, height := gearWindowSize()
	if window.width != width || window.height != height {
		t.Fatalf("gear size = %dx%d, want %dx%d", window.width, window.height, width, height)
	}
	if width <= equipmentWindowWidth {
		t.Fatal("gear window is not wider than the paperdoll")
	}
	labels := characterLabels(window.widgetTree(ctx, nil, nil))
	for _, want := range []string{"STR", "40 + 2", "Composite", "Points 3"} {
		if !strings.Contains(labels, want) {
			t.Fatalf("gear labels missing %q: %q", want, labels)
		}
	}
}

func TestViewedEquipmentStaysPaperdollSized(t *testing.T) {
	if equipmentWindowWidth >= gearWindowSizeWidth() {
		t.Fatal("viewed equipment width must stay the paperdoll width")
	}
}

func gearWindowSizeWidth() int {
	width, _ := gearWindowSize()
	return width
}
