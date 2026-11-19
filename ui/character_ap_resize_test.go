package ui

import (
	"strings"
	"testing"

	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
	"github.com/kivutar/goro/session"
	"github.com/kivutar/goro/ui/rotheme"
)

func characterLabels(root widget.Widget) string {
	var labels []string
	var walk func(widget.Widget)
	walk = func(current widget.Widget) {
		if current == nil {
			return
		}
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

func mountedFrameHeight(root widget.Widget) float32 {
	root.Layout(widget.NewContext(), geometry.Loose(geometry.Sz(400, 800)))
	bounds, ok := root.(interface{ Bounds() geometry.Rect })
	if !ok {
		return -1
	}
	return bounds.Bounds().Height()
}

func TestCharacterWindowAPResizesAfterOpen(t *testing.T) {
	ctx := Context{Session: &session.Session{Vitals: session.Vitals{HP: 10, MaxHP: 20, SP: 3, MaxSP: 8}}, ScreenW: 800, ScreenH: 600}
	window := &CharacterWindow{}
	window.Update(ctx)
	if window.baseHeight() != characterWindowHeight {
		t.Fatalf("height without AP = %d, want %d", window.baseHeight(), characterWindowHeight)
	}
	if strings.Contains(characterLabels(window.content), "AP ") {
		t.Fatal("AP row shown when max AP is zero")
	}
	ctx.Session.Vitals.AP, ctx.Session.Vitals.MaxAP = 40, 200
	window.Update(ctx)
	want := characterWindowHeight + int(rotheme.ProgressBarHeight+characterVitalsGap)
	if window.baseHeight() != want {
		t.Fatalf("height after AP = %d, want %d", window.baseHeight(), want)
	}
	if mountedFrameHeight(window.content) != float32(want) {
		t.Fatalf("mounted frame height = %.1f, want %d", mountedFrameHeight(window.content), want)
	}
	if !strings.Contains(characterLabels(window.content), "AP") {
		t.Fatal("mounted window did not show AP")
	}
	ctx.Session.Vitals.AP, ctx.Session.Vitals.MaxAP = 0, 0
	window.Update(ctx)
	if window.baseHeight() != characterWindowHeight || mountedFrameHeight(window.content) != float32(characterWindowHeight) {
		t.Fatalf("mounted height after clear = base %d frame %.1f", window.baseHeight(), mountedFrameHeight(window.content))
	}
}

func TestCharacterWindowCompactAPUsesItsOwnLine(t *testing.T) {
	ctx := Context{Session: &session.Session{Vitals: session.Vitals{HP: 10, MaxHP: 20, SP: 3, MaxSP: 8}}, ScreenW: 800, ScreenH: 600}
	window := &CharacterWindow{}
	window.SetCompact(true)
	window.Update(ctx)
	ctx.Session.Vitals.AP, ctx.Session.Vitals.MaxAP = 4, 12
	window.Update(ctx)
	want := characterWindowCompactHeight + int(rotheme.ProgressBarHeight+characterVitalsGap)
	if window.baseHeight() != want || mountedFrameHeight(window.content) != float32(want) {
		t.Fatalf("compact mounted height = base %d frame %.1f, want %d", window.baseHeight(), mountedFrameHeight(window.content), want)
	}
	if !strings.Contains(characterLabels(window.content), "AP 4 / 12") {
		t.Fatalf("compact labels = %q", characterLabels(window.content))
	}
}
