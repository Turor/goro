package ui

import (
	"strings"
	"testing"

	"github.com/gogpu/ui/widget"
	"github.com/kivutar/goro/session"
	"github.com/kivutar/goro/ui/rotheme"
)

func characterLabels(root widget.Widget) string {
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
	walk(root)
	return strings.Join(labels, "\n")
}

func TestCharacterWindowAPResizesAfterOpen(t *testing.T) {
	ctx := Context{Session: &session.Session{Vitals: session.Vitals{HP: 10, MaxHP: 20, SP: 3, MaxSP: 8}}, ScreenW: 800, ScreenH: 600}
	window := &CharacterWindow{}
	window.Update(ctx)
	if window.height != characterWindowHeight {
		t.Fatalf("height without AP = %d, want %d", window.height, characterWindowHeight)
	}
	if strings.Contains(characterLabels(window.widgetTree(ctx)), "AP ") {
		t.Fatal("AP row shown when max AP is zero")
	}
	ctx.Session.Vitals.AP, ctx.Session.Vitals.MaxAP = 40, 200
	window.Update(ctx)
	want := characterWindowHeight + int(rotheme.ProgressBarHeight+characterVitalsGap)
	if window.height != want {
		t.Fatalf("height after AP = %d, want %d", window.height, want)
	}
	if !strings.Contains(characterLabels(window.widgetTree(ctx)), "AP") {
		t.Fatal("open window did not show AP")
	}
	ctx.Session.Vitals.AP, ctx.Session.Vitals.MaxAP = 0, 0
	window.Update(ctx)
	if window.height != characterWindowHeight {
		t.Fatalf("height after AP cleared = %d, want %d", window.height, characterWindowHeight)
	}
}

func TestCharacterWindowCompactAPUsesItsOwnLine(t *testing.T) {
	ctx := Context{Session: &session.Session{Vitals: session.Vitals{HP: 10, MaxHP: 20, SP: 3, MaxSP: 8, AP: 4, MaxAP: 12}}, ScreenW: 800, ScreenH: 600}
	window := &CharacterWindow{}
	window.SetCompact(true)
	window.Update(ctx)
	want := characterWindowCompactHeight + int(rotheme.ProgressBarHeight+characterVitalsGap)
	if window.height != want {
		t.Fatalf("compact height = %d, want %d", window.height, want)
	}
	labels := characterLabels(window.widgetTree(ctx))
	if !strings.Contains(labels, "AP 4 / 12") {
		t.Fatalf("compact labels = %q", labels)
	}
}
