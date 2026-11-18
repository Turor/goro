package maps

import (
	"image/color"
	"testing"

	"github.com/kivutar/goro/render"
	"github.com/kivutar/goro/ui/rotheme"
)

func TestLoadingBarFollowsLayoutTokens(t *testing.T) {
	t.Cleanup(rotheme.ResetLayout)
	if got := fadedLoadingColor(rotheme.LoadingBarFill, 128); got != (color.RGBA{R: 66, G: 99, B: 165, A: 128}) {
		t.Fatalf("default faded fill %v, want the old literal with the fade alpha", got)
	}
	if err := rotheme.SetToken("loading", "bar_height", "30"); err != nil {
		t.Fatal(err)
	}
	if err := rotheme.SetToken("loading", "bar_fill", "#ff000080"); err != nil {
		t.Fatal(err)
	}
	if rotheme.LoadingBarHeight != 30 {
		t.Fatalf("bar height %v, want 30", rotheme.LoadingBarHeight)
	}
	if got := fadedLoadingColor(rotheme.LoadingBarFill, 255); got != (color.RGBA{R: 255, A: 128}) {
		t.Fatalf("themed fill %v", got)
	}
	(&Loading{}).Draw(render.NewFrame(800, 600), 50, 255)
	rotheme.ResetLayout()
	if rotheme.LoadingBarHeight != 15 || rotheme.LoadingBarFill != (color.RGBA{R: 66, G: 99, B: 165, A: 255}) {
		t.Fatalf("defaults not restored: height %v fill %v", rotheme.LoadingBarHeight, rotheme.LoadingBarFill)
	}
}
