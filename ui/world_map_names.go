package ui

import (
	"image"
	"image/color"
	"image/draw"
	"sync"

	"github.com/gogpu/ui/widget"
	"github.com/kivutar/goro/res"
	"github.com/kivutar/goro/ui/rotheme"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

// atlasImage is the picture the board records. Names are painted into that
// picture once. Drawing them as scene text turns every label into vector
// outlines, and replaying those outlines is what makes the map slow.
func (w *worldMapWidget) atlasImage() image.Image {
	if w == nil || w.image == nil || !w.showNames {
		if w == nil {
			return nil
		}
		return w.image
	}
	if w.named != nil && w.namedFor == w.image {
		return w.named
	}
	w.named = bakeWorldMapNames(w.image, w.entries, w.sourceSize)
	w.namedFor = w.image
	return w.named
}

func bakeWorldMapNames(src image.Image, entries []res.WorldMapEntry, source image.Point) *image.RGBA {
	bounds := src.Bounds()
	dst := image.NewRGBA(image.Rect(0, 0, bounds.Dx(), bounds.Dy()))
	draw.Draw(dst, dst.Bounds(), src, bounds.Min, draw.Src)
	if source.X <= 0 || source.Y <= 0 {
		return dst
	}
	sx := float64(dst.Bounds().Dx()) / float64(source.X)
	sy := float64(dst.Bounds().Dy()) / float64(source.Y)
	for _, entry := range entries {
		label := worldMapEntryLabel(entry)
		if label == "" {
			continue
		}
		cell := image.Rect(
			int(float64(entry.Bounds.Min.X)*sx),
			int(float64(entry.Bounds.Min.Y)*sy),
			int(float64(entry.Bounds.Max.X)*sx),
			int(float64(entry.Bounds.Max.Y)*sy),
		)
		paintWorldMapName(dst, cell, label)
	}
	return dst
}

func paintWorldMapName(dst *image.RGBA, cell image.Rectangle, label string) {
	if cell.Dx() < 8 || cell.Dy() < 6 {
		return
	}
	size := worldMapNameSize
	face := worldMapFace(size)
	if face == nil {
		return
	}
	width := font.MeasureString(face, label).Ceil()
	if width > cell.Dx() && width > 0 {
		size *= float32(cell.Dx()) / float32(width)
		if size < 5 {
			return
		}
		face = worldMapFace(size)
		if face == nil {
			return
		}
		width = font.MeasureString(face, label).Ceil()
	}
	labelH := int(worldMapNameH)
	if labelH > cell.Dy() {
		labelH = cell.Dy()
	}
	top := cell.Min.Y + (cell.Dy()-labelH)/2
	bar := image.Rect(cell.Min.X, top, cell.Max.X, top+labelH).Intersect(dst.Bounds())
	fillWidgetRect(dst, bar, worldMapLabelFill)
	ascent := face.Metrics().Ascent.Ceil()
	x := cell.Min.X + (cell.Dx()-width)/2
	y := top + (labelH+ascent)/2
	drawer := font.Drawer{
		Dst:  dst,
		Src:  image.NewUniform(widgetColorNRGBA(rotheme.Default.Colors.Text)),
		Face: face,
		Dot:  fixed.P(x, y),
	}
	drawer.DrawString(label)
}

func fillWidgetRect(dst *image.RGBA, rect image.Rectangle, c widget.Color) {
	if c.A <= 0 {
		return
	}
	cr, cg, cb, ca := c.R, c.G, c.B, c.A
	for y := rect.Min.Y; y < rect.Max.Y; y++ {
		for x := rect.Min.X; x < rect.Max.X; x++ {
			i := dst.PixOffset(x, y)
			inv := 1 - ca
			dst.Pix[i] = uint8(float32(dst.Pix[i])*inv + cr*255*ca)
			dst.Pix[i+1] = uint8(float32(dst.Pix[i+1])*inv + cg*255*ca)
			dst.Pix[i+2] = uint8(float32(dst.Pix[i+2])*inv + cb*255*ca) // theme:ok blue channel index
		}
	}
}

func widgetColorNRGBA(c widget.Color) color.NRGBA {
	// theme:ok converts an existing theme color to 8-bit channels
	return color.NRGBA{
		R: uint8(c.R * 255),
		G: uint8(c.G * 255),
		B: uint8(c.B * 255),
		A: uint8(c.A * 255),
	}
}

var (
	worldMapFontOnce sync.Once
	worldMapFont     *opentype.Font
	worldMapFaces    = map[int]font.Face{}
	worldMapFaceMu   sync.Mutex
)

func worldMapFace(size float32) font.Face {
	worldMapFontOnce.Do(func() {
		parsed, err := opentype.Parse(goregular.TTF)
		if err != nil {
			return
		}
		worldMapFont = parsed
	})
	if worldMapFont == nil || size < 5 {
		return nil
	}
	key := int(size + 0.5) // theme:ok round the font size to a cached face
	worldMapFaceMu.Lock()
	defer worldMapFaceMu.Unlock()
	if face, ok := worldMapFaces[key]; ok {
		return face
	}
	face, err := opentype.NewFace(worldMapFont, &opentype.FaceOptions{Size: float64(key), DPI: 72})
	if err != nil {
		return nil
	}
	worldMapFaces[key] = face
	return face
}
