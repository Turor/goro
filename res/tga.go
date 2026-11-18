package res

import (
	"encoding/binary"
	"fmt"
	"image"
	"image/color"
)

// decodeTGA decodes the TGA variants found in RO data: true-colour (types 2,
// 10), grey-scale (3, 11) and colour-mapped (1, 9) images, raw or RLE, at 8,
// 15/16, 24 or 32 bits per pixel.
func decodeTGA(data []byte) (image.Image, error) {
	if len(data) < 18 {
		return nil, fmt.Errorf("invalid tga: short header")
	}
	idLength := int(data[0])
	colorMapType := data[1]
	imageType := data[2]
	mapFirst := int(binary.LittleEndian.Uint16(data[3:5]))
	mapLength := int(binary.LittleEndian.Uint16(data[5:7]))
	mapBits := int(data[7])
	width := int(binary.LittleEndian.Uint16(data[12:14]))
	height := int(binary.LittleEndian.Uint16(data[14:16]))
	bpp := int(data[16])
	descriptor := data[17]
	if width <= 0 || height <= 0 {
		return nil, fmt.Errorf("invalid tga size %dx%d", width, height)
	}
	offset := 18 + idLength
	if offset > len(data) {
		return nil, fmt.Errorf("invalid tga id length")
	}

	var palette []color.NRGBA
	if colorMapType == 1 {
		entryBytes := (mapBits + 7) / 8
		if entryBytes < 2 || entryBytes > 4 {
			return nil, fmt.Errorf("unsupported tga color map depth %d", mapBits)
		}
		if offset+mapLength*entryBytes > len(data) {
			return nil, fmt.Errorf("truncated tga color map")
		}
		palette = make([]color.NRGBA, mapFirst+mapLength)
		for i := 0; i < mapLength; i++ {
			palette[mapFirst+i] = tgaColor(data[offset:offset+entryBytes], false)
			offset += entryBytes
		}
	} else if colorMapType != 0 {
		return nil, fmt.Errorf("unsupported tga color map type %d", colorMapType)
	}

	var pixel func(px []byte) color.NRGBA
	switch imageType &^ 8 {
	case 1:
		if palette == nil || (bpp != 8 && bpp != 16) {
			return nil, fmt.Errorf("unsupported tga color-mapped image (map=%d bpp=%d)", colorMapType, bpp)
		}
		pixel = func(px []byte) color.NRGBA {
			index := int(px[0])
			if len(px) > 1 {
				index |= int(px[1]) << 8
			}
			if index < len(palette) {
				return palette[index]
			}
			return color.NRGBA{A: 255}
		}
	case 2:
		if bpp != 15 && bpp != 16 && bpp != 24 && bpp != 32 {
			return nil, fmt.Errorf("unsupported tga bpp %d", bpp)
		}
		alpha := bpp == 32 || (bpp == 16 && descriptor&0x0f != 0)
		pixel = func(px []byte) color.NRGBA { return tgaColor(px, alpha) }
	case 3:
		if bpp != 8 && bpp != 16 {
			return nil, fmt.Errorf("unsupported tga grey-scale bpp %d", bpp)
		}
		pixel = func(px []byte) color.NRGBA {
			a := uint8(255)
			if len(px) > 1 {
				a = px[1]
			}
			return color.NRGBA{R: px[0], G: px[0], B: px[0], A: a}
		}
	default:
		return nil, fmt.Errorf("unsupported tga type %d", imageType)
	}

	bytesPerPixel := (bpp + 7) / 8
	img := image.NewNRGBA(image.Rect(0, 0, width, height))
	topOrigin := descriptor&0x20 != 0
	total := width * height
	if imageType&8 == 0 {
		if offset+total*bytesPerPixel > len(data) {
			return nil, fmt.Errorf("truncated tga pixels")
		}
		for i := 0; i < total; i++ {
			setTGAPixel(img, i, topOrigin, pixel(data[offset:offset+bytesPerPixel]))
			offset += bytesPerPixel
		}
		return img, nil
	}
	for i := 0; i < total; {
		if offset >= len(data) {
			return nil, fmt.Errorf("truncated tga rle")
		}
		header := data[offset]
		offset++
		count := int(header&0x7f) + 1
		if header&0x80 != 0 {
			if offset+bytesPerPixel > len(data) {
				return nil, fmt.Errorf("truncated tga rle pixel")
			}
			c := pixel(data[offset : offset+bytesPerPixel])
			offset += bytesPerPixel
			for j := 0; j < count && i < total; j++ {
				setTGAPixel(img, i, topOrigin, c)
				i++
			}
			continue
		}
		for j := 0; j < count && i < total; j++ {
			if offset+bytesPerPixel > len(data) {
				return nil, fmt.Errorf("truncated tga raw packet")
			}
			setTGAPixel(img, i, topOrigin, pixel(data[offset:offset+bytesPerPixel]))
			offset += bytesPerPixel
			i++
		}
	}
	return img, nil
}

// tgaColor reads a little-endian BGR(A) or 15/16-bit ARGB1555 value.
func tgaColor(px []byte, alpha bool) color.NRGBA {
	switch len(px) {
	case 2:
		v := uint16(px[0]) | uint16(px[1])<<8
		c := color.NRGBA{R: expand5(v >> 10), G: expand5(v >> 5), B: expand5(v), A: 255}
		if alpha && v&0x8000 == 0 {
			c.A = 0
		}
		return c
	case 3:
		return color.NRGBA{R: px[2], G: px[1], B: px[0], A: 255}
	default:
		a := uint8(255)
		if alpha {
			a = px[3]
		}
		return color.NRGBA{R: px[2], G: px[1], B: px[0], A: a}
	}
}

func expand5(v uint16) uint8 {
	v &= 0x1f
	return uint8(v<<3 | v>>2)
}

func setTGAPixel(img *image.NRGBA, index int, topOrigin bool, c color.NRGBA) {
	width := img.Bounds().Dx()
	x := index % width
	y := index / width
	if !topOrigin {
		y = img.Bounds().Dy() - 1 - y
	}
	img.SetNRGBA(x, y, c)
}
