package ui

import (
	"image"
	"image/color"
	"testing"

	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/uitest"
	"github.com/gogpu/ui/widget"
	"github.com/kivutar/goro/res"
)

func TestWorldMapNamesAreBakedIntoThePicture(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 64, 48))
	for i := range img.Pix {
		img.Pix[i] = 255
	}
	board := &worldMapWidget{
		image: img, sourceSize: img.Bounds().Size(), showNames: true,
		entries: []res.WorldMapEntry{{MapName: "prontera", Name: "Prontera", Bounds: image.Rect(0, 0, 40, 30)}},
	}
	board.SetBounds(geometry.NewRect(0, 0, 64, 48))
	canvas := &uitest.MockCanvas{}
	board.Draw(widget.NewContext(), canvas)
	if len(canvas.Texts) != 0 || len(canvas.StyledTexts) != 0 {
		t.Fatal("map names were drawn as scene text")
	}
	if len(canvas.Images) != 1 {
		t.Fatalf("images = %d, want the baked picture", len(canvas.Images))
	}
	baked, ok := canvas.Images[0].Image.(*image.RGBA)
	if !ok || baked == img {
		t.Fatal("names were not painted into their own picture")
	}
	if !namePixelsChanged(img, baked) {
		t.Fatal("baked picture has no name pixels")
	}
	again := &uitest.MockCanvas{}
	board.Draw(widget.NewContext(), again)
	if again.Images[0].Image != baked {
		t.Fatal("names were painted again")
	}
}

func namePixelsChanged(src, dst *image.RGBA) bool {
	n := min(len(src.Pix), len(dst.Pix))
	for i := 0; i < n; i++ {
		if src.Pix[i] != dst.Pix[i] {
			return true
		}
	}
	return false
}

func TestWidgetColorNRGBAKeepsOpaqueText(t *testing.T) {
	got := widgetColorNRGBA(widget.Color{R: 0, G: 0, B: 0, A: 1})
	if got != (color.NRGBA{A: 255}) {
		t.Fatalf("color = %+v", got)
	}
}
