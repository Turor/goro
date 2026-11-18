package actors

import (
	"image/color"
	"strings"
	"testing"

	"github.com/kivutar/goro/ui/rotheme"
	worldstate "github.com/kivutar/goro/world"
)

func TestActorNameColorsFollowPalette(t *testing.T) {
	t.Cleanup(rotheme.ResetLayout)
	layout, _, err := rotheme.ParseLayout(strings.NewReader("[palette]\nhighlight = #ff00ff\n[speech_bubble]\nmax_width = 300\n"), "test")
	if err != nil {
		t.Fatal(err)
	}
	layout.Apply()
	if got := ActorNameLabelColor(worldstate.Actor{IsAdmin: true}, false); got != (color.RGBA{R: 255, B: 255, A: 255}) {
		t.Fatalf("admin name did not follow @highlight: %v", got)
	}
	if speechBubbleMaxWidth != 300 {
		t.Fatalf("speech bubble width = %v", speechBubbleMaxWidth)
	}
	rotheme.ResetLayout()
	if got := ActorNameLabelColor(worldstate.Actor{IsAdmin: true}, false); got != (color.RGBA{R: 255, G: 255, A: 255}) {
		t.Fatalf("reset admin name = %v", got)
	}
}
