package ui

import (
	"image"
	"image/draw"

	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
	xdraw "golang.org/x/image/draw"
)

// drawFittedImage draws img inside dst. A resized window scales the cell, so
// the bitmap has to follow that rect instead of staying at its file size.
func drawFittedImage(canvas widget.Canvas, img image.Image, dst geometry.Rect) {
	if canvas == nil || img == nil || dst.Width() < 1 || dst.Height() < 1 {
		return
	}
	bounds := img.Bounds()
	width, height := int(dst.Width()), int(dst.Height())
	if width == bounds.Dx() && height == bounds.Dy() {
		canvas.DrawImage(img, dst.Min)
		return
	}
	fitted := image.NewRGBA(image.Rect(0, 0, width, height))
	xdraw.NearestNeighbor.Scale(fitted, fitted.Bounds(), img, bounds, draw.Over, nil)
	canvas.DrawImage(fitted, dst.Min)
}
