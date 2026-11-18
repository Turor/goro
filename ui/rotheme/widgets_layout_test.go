package rotheme

import (
	"testing"

	"github.com/gogpu/ui/core/button"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/uitest"
	"github.com/gogpu/ui/widget"
)

// Buttons, bars and glyphs are painted widgets, not windows. Rebuilding means
// constructing them again (sizes are copied then) and drawing again (colours
// and arrow paths are read at paint time). OnLayoutApplied refreshes the
// directional-arrow paths during SetToken.
func TestWidgetsFollowLayoutTokens(t *testing.T) {
	t.Cleanup(ResetLayout)

	defButtonH := buttonBoxHeight()
	defBarH := progressBarLaidOutHeight()
	defShadow := buttonShadowColor()
	defTrack := sliderTrackPaint()
	defArrow := directionalArrows[IconButtonLeft].fillPath

	mustSetToken(t, "button", "padding_y", "12")
	mustSetToken(t, "button", "shadow_far", "#ff00aa80")
	mustSetToken(t, "progress_bar", "height", "20")
	mustSetToken(t, "slider", "track", "#112233")
	mustSetToken(t, "icon_button", "arrow_border_inset", "3")

	if got := buttonBoxHeight(); got == defButtonH {
		t.Fatalf("button height stayed %.1f after padding_y changed", got)
	}
	if got := progressBarLaidOutHeight(); got == defBarH {
		t.Fatalf("progress bar height stayed %.1f after height changed", got)
	}
	if got := buttonShadowColor(); got == defShadow {
		t.Fatalf("button shadow stayed %v", got)
	}
	if got := sliderTrackPaint(); got == defTrack {
		t.Fatalf("slider track stayed %v", got)
	}
	if directionalArrows[IconButtonLeft].fillPath == defArrow {
		t.Fatalf("directional arrow was not rebuilt after the border inset changed (inset %.2f)", directionalArrowBorderInset)
	}

	ResetLayout()
	if got := buttonBoxHeight(); got != defButtonH {
		t.Fatalf("button height = %.1f after reset, want %.1f", got, defButtonH)
	}
	if got := progressBarLaidOutHeight(); got != defBarH {
		t.Fatalf("progress bar height = %.1f after reset, want %.1f", got, defBarH)
	}
	if got := buttonShadowColor(); got != defShadow {
		t.Fatalf("button shadow = %v after reset, want %v", got, defShadow)
	}
	if got := sliderTrackPaint(); got != defTrack {
		t.Fatalf("slider track = %v after reset, want %v", got, defTrack)
	}
	if directionalArrows[IconButtonLeft].fillPath != defArrow {
		t.Fatal("directional arrow did not return to the default path")
	}
}

func mustSetToken(t *testing.T, section, key, raw string) {
	t.Helper()
	if err := SetToken(section, key, raw); err != nil {
		t.Fatalf("set %s.%s = %s: %v", section, key, raw, err)
	}
}

func buttonBoxHeight() float32 {
	box := Button("Ok", nil)
	return box.Layout(widget.NewContext(), geometry.Loose(geometry.Sz(240, 80))).Height
}

func progressBarLaidOutHeight() float32 {
	bar := ProgressBar()
	return bar.Layout(widget.NewContext(), geometry.Loose(geometry.Sz(200, 80))).Height
}

func buttonShadowColor() widget.Color {
	canvas := &uitest.MockCanvas{}
	ButtonPainter{}.PaintButton(canvas, button.PaintState{Bounds: geometry.NewRect(0, 0, 80, 22)})
	if len(canvas.RoundRects) == 0 {
		return widget.Color{}
	}
	return canvas.RoundRects[0].Color
}

func sliderTrackPaint() widget.Color {
	canvas := &uitest.MockCanvas{}
	paintHorizontalSliderTrack(canvas, geometry.NewRect(0, 0, 100, SliderTrackHeight), 40, false)
	if len(canvas.Rects) == 0 {
		return widget.Color{}
	}
	return canvas.Rects[0].Color
}
