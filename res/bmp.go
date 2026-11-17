package res

import (
	"encoding/binary"
	"fmt"
	"image"
	"image/color"
	"math/bits"
)

// decodeBMPFallback decodes the uncompressed BMPs that golang.org/x/image/bmp
// rejects but the RO client draws: 16-bit (RGB555 or bitfields), 32-bit
// BI_BITFIELDS with V4/V5 headers, and palette images whose pixels index past
// a short colour table (those pixels come out black).
func decodeBMPFallback(data []byte) (image.Image, error) {
	if len(data) < 26 || data[0] != 'B' || data[1] != 'M' {
		return nil, fmt.Errorf("not a bmp")
	}
	pixelOffset := int(binary.LittleEndian.Uint32(data[10:14]))
	headerSize := int(binary.LittleEndian.Uint32(data[14:18]))
	if headerSize < 40 || 14+headerSize > len(data) {
		return nil, fmt.Errorf("unsupported bmp header size %d", headerSize)
	}
	width := int(int32(binary.LittleEndian.Uint32(data[18:22])))
	rawHeight := int(int32(binary.LittleEndian.Uint32(data[22:26])))
	bpp := int(binary.LittleEndian.Uint16(data[28:30]))
	compression := binary.LittleEndian.Uint32(data[30:34])
	colorsUsed := int(binary.LittleEndian.Uint32(data[46:50]))
	topDown := rawHeight < 0
	height := rawHeight
	if topDown {
		height = -height
	}
	if width <= 0 || height <= 0 || width > 16384 || height > 16384 {
		return nil, fmt.Errorf("invalid bmp size %dx%d", width, rawHeight)
	}
	if compression != 0 && compression != 3 && compression != 6 {
		return nil, fmt.Errorf("unsupported bmp compression %d", compression)
	}

	var masks [4]uint32 // r, g, b, a
	switch bpp {
	case 16:
		masks = [4]uint32{0x7c00, 0x03e0, 0x001f, 0}
	case 32:
		masks = [4]uint32{0x00ff0000, 0x0000ff00, 0x000000ff, 0}
	}
	if compression == 3 || compression == 6 {
		maskOffset := 14 + 40
		count := 3
		if compression == 6 || headerSize >= 56 {
			count = 4
		}
		if maskOffset+count*4 > len(data) {
			return nil, fmt.Errorf("truncated bmp bitfields")
		}
		for i := 0; i < count; i++ {
			masks[i] = binary.LittleEndian.Uint32(data[maskOffset+i*4:])
		}
	}

	var palette []color.NRGBA
	if bpp <= 8 {
		entries := colorsUsed
		if entries == 0 || entries > 1<<bpp {
			entries = 1 << bpp
		}
		start := 14 + headerSize
		if start+entries*4 > len(data) {
			return nil, fmt.Errorf("truncated bmp palette")
		}
		palette = make([]color.NRGBA, entries)
		for i := range palette {
			p := data[start+i*4:]
			palette[i] = color.NRGBA{R: p[2], G: p[1], B: p[0], A: 255}
		}
	} else if bpp != 16 && bpp != 24 && bpp != 32 {
		return nil, fmt.Errorf("unsupported bmp bpp %d", bpp)
	}

	stride := (width*bpp + 31) / 32 * 4
	if pixelOffset < 0 || pixelOffset+stride*height > len(data) {
		return nil, fmt.Errorf("truncated bmp pixels")
	}
	img := image.NewNRGBA(image.Rect(0, 0, width, height))
	for row := 0; row < height; row++ {
		y := height - 1 - row
		if topDown {
			y = row
		}
		line := data[pixelOffset+row*stride:]
		for x := 0; x < width; x++ {
			var c color.NRGBA
			switch bpp {
			case 1, 2, 4, 8:
				bit := x * bpp
				index := int(line[bit/8]>>(8-bpp-bit%8)) & (1<<bpp - 1)
				if index < len(palette) {
					c = palette[index]
				} else {
					c = color.NRGBA{A: 255}
				}
			case 16:
				c = bmpMasked(uint32(binary.LittleEndian.Uint16(line[x*2:])), masks)
			case 24:
				p := line[x*3:]
				c = color.NRGBA{R: p[2], G: p[1], B: p[0], A: 255}
			case 32:
				c = bmpMasked(binary.LittleEndian.Uint32(line[x*4:]), masks)
			}
			img.SetNRGBA(x, y, c)
		}
	}
	return img, nil
}

func bmpMasked(v uint32, masks [4]uint32) color.NRGBA {
	c := color.NRGBA{R: bmpChannel(v, masks[0]), G: bmpChannel(v, masks[1]), B: bmpChannel(v, masks[2]), A: 255}
	if masks[3] != 0 {
		c.A = bmpChannel(v, masks[3])
	}
	return c
}

// bmpChannel extracts the masked bits and scales them to 0..255.
func bmpChannel(v, mask uint32) uint8 {
	if mask == 0 {
		return 0
	}
	shift := bits.TrailingZeros32(mask)
	width := bits.OnesCount32(mask)
	value := (v & mask) >> shift
	if width >= 8 {
		return uint8(value >> (width - 8))
	}
	max := uint32(1)<<width - 1
	return uint8(value * 255 / max)
}
