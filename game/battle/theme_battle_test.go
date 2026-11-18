package battle

import (
	"image/color"
	"strings"
	"testing"

	"github.com/kivutar/goro/ui/rotheme"
)

func applyLayout(t *testing.T, src string) {
	t.Helper()
	layout, _, err := rotheme.ParseLayout(strings.NewReader(src), "test")
	if err != nil {
		t.Fatal(err)
	}
	layout.Apply()
}

// World overlays read their tokens on every draw, so a reload (or a palette
// change) reaches them without rebuilding anything.
func TestWorldOverlayTokensFollowLayoutAndPalette(t *testing.T) {
	t.Cleanup(rotheme.ResetLayout)
	if got := damageFloaterColor(FloaterIncoming, color.RGBA{}); got != (color.RGBA{R: 255, G: 64, B: 64, A: 255}) {
		t.Fatalf("built-in incoming color = %v", got)
	}
	applyLayout(t, "[palette]\nbad = #2040ff\nworld_text = #808080\n[actor_bars]\nwidth = 80\nframe = 2\n[damage_numbers]\ncritical = #123456\n")
	if got := damageFloaterColor(FloaterCritical, color.RGBA{}); got != (color.RGBA{R: 0x12, G: 0x34, B: 0x56, A: 255}) {
		t.Fatalf("explicit critical color = %v", got)
	}
	if got := damageFloaterColor(FloaterIncoming, color.RGBA{}); got == (color.RGBA{R: 255, G: 64, B: 64, A: 255}) || got.B <= got.R {
		t.Fatalf("incoming did not follow @bad: %v", got)
	}
	if got := damageFloaterColor(FloaterMiss, color.RGBA{}); got != (color.RGBA{R: 128, G: 128, B: 128, A: 255}) {
		t.Fatalf("miss did not follow @world_text: %v", got)
	}
	if got := actorOverlayBarFillWidth(1); got != 76 {
		t.Fatalf("full bar fill = %v, want width 80 minus a 2px frame on each side", got)
	}
	rotheme.ResetLayout()
	if got := actorOverlayBarFillWidth(1); got != 58 {
		t.Fatalf("reset fill width = %v", got)
	}
}
