package ui

import (
	"strings"
	"testing"

	"github.com/gogpu/ui/widget"
	"github.com/kivutar/goro/session"
	"github.com/kivutar/goro/ui/rotheme"
)

func TestCharacterWindowShowsFourthClassAP(t *testing.T) {
	ctx := Context{Session: &session.Session{Vitals: session.Vitals{HP: 10, MaxHP: 20, SP: 3, MaxSP: 8, AP: 40, MaxAP: 200}}}
	window := &CharacterWindow{}
	window.Update(ctx)
	if window.height != characterWindowHeight+int(rotheme.ProgressBarHeight+characterVitalsGap) {
		t.Fatalf("height = %d, want base plus an AP row", window.height)
	}
	var labels []string
	var walk func(widget.Widget)
	walk = func(current widget.Widget) {
		if text, ok := current.(interface{ Content() string }); ok {
			labels = append(labels, text.Content())
		}
		for _, child := range current.Children() {
			walk(child)
		}
	}
	walk(window.widgetTree(ctx))
	if !strings.Contains(strings.Join(labels, "\n"), "AP") {
		t.Fatalf("character window labels = %q, want AP", labels)
	}
}
