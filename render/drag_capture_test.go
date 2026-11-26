package render

import (
	"image/color"
	"testing"

	"github.com/gogpu/ui/geometry"
)

func TestCoverDragCaptureReplacesOnlyTheWindow(t *testing.T) {
	dst := NewImage(8, 6)
	dst.Fill(color.RGBA{R: 200, A: 255})
	src := NewImage(4, 3)
	src.Fill(color.RGBA{B: 180, A: 255})
	coverDragCapture(dst, geometry.NewRect(0, 0, 8, 6), geometry.NewRect(0, 0, 4, 3), src)

	if got := dst.pix.RGBAAt(1, 1); got.B != 180 || got.R != 0 {
		t.Fatalf("window pixel = %+v, want the window's own pixels", got)
	}
	if got := dst.pix.RGBAAt(6, 4); got.R != 200 {
		t.Fatalf("pixel outside the window = %+v, want the region crop", got)
	}
}
