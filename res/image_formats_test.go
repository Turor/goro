package res

import (
	"bytes"
	"encoding/binary"
	"hash/crc32"
	"image"
	"image/color"
	"image/png"
	"strings"
	"sync"
	"testing"
)

func tgaHeader(imageType, cmapType byte, cmapLen int, cmapBits, bpp byte, w, h int, descriptor byte) []byte {
	hdr := make([]byte, 18)
	hdr[1] = cmapType
	hdr[2] = imageType
	binary.LittleEndian.PutUint16(hdr[5:], uint16(cmapLen))
	hdr[7] = cmapBits
	binary.LittleEndian.PutUint16(hdr[12:], uint16(w))
	binary.LittleEndian.PutUint16(hdr[14:], uint16(h))
	hdr[16] = bpp
	hdr[17] = descriptor
	return hdr
}

func TestDecodeTGAGreyscale8(t *testing.T) {
	data := append(tgaHeader(3, 0, 0, 0, 8, 2, 1, 0x20), 10, 200)
	img, err := decodeTGA(data)
	if err != nil {
		t.Fatal(err)
	}
	if got := img.(*image.NRGBA).NRGBAAt(1, 0); got != (color.NRGBA{200, 200, 200, 255}) {
		t.Fatalf("pixel = %v", got)
	}
}

func TestDecodeTGAColorMappedAndRLE(t *testing.T) {
	data := tgaHeader(9, 1, 2, 24, 8, 3, 1, 0x20)
	data = append(data, 0, 0, 255 /* red */, 255, 0, 0 /* blue */)
	data = append(data, 0x81, 1, 0x00, 0) // 2 x index 1, then 1 x index 0
	img, err := decodeTGA(data)
	if err != nil {
		t.Fatal(err)
	}
	n := img.(*image.NRGBA)
	if n.NRGBAAt(0, 0) != (color.NRGBA{0, 0, 255, 255}) || n.NRGBAAt(1, 0) != (color.NRGBA{0, 0, 255, 255}) || n.NRGBAAt(2, 0) != (color.NRGBA{255, 0, 0, 255}) {
		t.Fatalf("pixels = %v %v %v", n.NRGBAAt(0, 0), n.NRGBAAt(1, 0), n.NRGBAAt(2, 0))
	}
}

func TestDecodeTGATrueColorBottomOrigin(t *testing.T) {
	data := append(tgaHeader(2, 0, 0, 0, 32, 1, 2, 0), 1, 2, 3, 4, 5, 6, 7, 8)
	img, err := decodeTGA(data)
	if err != nil {
		t.Fatal(err)
	}
	// Bottom-origin: the first pixel is the bottom row.
	if got := img.(*image.NRGBA).NRGBAAt(0, 1); got != (color.NRGBA{3, 2, 1, 4}) {
		t.Fatalf("pixel = %v", got)
	}
}

func bmpFile(bpp int, w, h int, extra []byte, palette []byte, pixels []byte, headerSize int, compression uint32, colorsUsed uint32) []byte {
	var b bytes.Buffer
	offset := 14 + headerSize + len(palette)
	b.WriteString("BM")
	binary.Write(&b, binary.LittleEndian, uint32(offset+len(pixels)))
	binary.Write(&b, binary.LittleEndian, uint32(0))
	binary.Write(&b, binary.LittleEndian, uint32(offset))
	info := make([]byte, headerSize)
	binary.LittleEndian.PutUint32(info[0:], uint32(headerSize))
	binary.LittleEndian.PutUint32(info[4:], uint32(w))
	binary.LittleEndian.PutUint32(info[8:], uint32(h))
	binary.LittleEndian.PutUint16(info[12:], 1)
	binary.LittleEndian.PutUint16(info[14:], uint16(bpp))
	binary.LittleEndian.PutUint32(info[16:], compression)
	binary.LittleEndian.PutUint32(info[32:], colorsUsed)
	copy(info[40:], extra)
	b.Write(info)
	b.Write(palette)
	b.Write(pixels)
	return b.Bytes()
}

func TestDecodeBMP16BitRGB555(t *testing.T) {
	// One row, two pixels: pure red, pure green; row padded to 4 bytes.
	data := bmpFile(16, 2, 1, nil, nil, []byte{0x00, 0x7c, 0xe0, 0x03}, 40, 0, 0)
	img, err := DecodeImageData(data)
	if err != nil {
		t.Fatal(err)
	}
	n := img.(*image.NRGBA)
	if n.NRGBAAt(0, 0) != (color.NRGBA{255, 0, 0, 255}) || n.NRGBAAt(1, 0) != (color.NRGBA{0, 255, 0, 255}) {
		t.Fatalf("pixels = %v %v", n.NRGBAAt(0, 0), n.NRGBAAt(1, 0))
	}
}

func TestDecodeBMP8BitShortPaletteOutOfRangeIsBlack(t *testing.T) {
	palette := []byte{0, 0, 255, 0} // one entry: red
	data := bmpFile(8, 2, 1, nil, palette, []byte{0, 7, 0, 0}, 40, 0, 1)
	img, err := DecodeImageData(data)
	if err != nil {
		t.Fatal(err)
	}
	n := img.(*image.NRGBA)
	if n.NRGBAAt(0, 0) != (color.NRGBA{255, 0, 0, 255}) || n.NRGBAAt(1, 0) != (color.NRGBA{0, 0, 0, 255}) {
		t.Fatalf("pixels = %v %v", n.NRGBAAt(0, 0), n.NRGBAAt(1, 0))
	}
}

func TestDecodeBMP32BitV5Bitfields(t *testing.T) {
	masks := make([]byte, 16)
	binary.LittleEndian.PutUint32(masks[0:], 0x00ff0000)
	binary.LittleEndian.PutUint32(masks[4:], 0x0000ff00)
	binary.LittleEndian.PutUint32(masks[8:], 0x000000ff)
	binary.LittleEndian.PutUint32(masks[12:], 0xff000000)
	data := bmpFile(32, 1, 1, masks, nil, []byte{0x30, 0x20, 0x10, 0x80}, 124, 3, 0)
	img, err := DecodeImageData(data)
	if err != nil {
		t.Fatal(err)
	}
	if got := img.(*image.NRGBA).NRGBAAt(0, 0); got != (color.NRGBA{0x10, 0x20, 0x30, 0x80}) {
		t.Fatalf("pixel = %v", got)
	}
}

func TestDecodePNGWithBadChunkChecksum(t *testing.T) {
	var buf bytes.Buffer
	src := image.NewNRGBA(image.Rect(0, 0, 1, 1))
	src.SetNRGBA(0, 0, color.NRGBA{1, 2, 3, 255})
	if err := png.Encode(&buf, src); err != nil {
		t.Fatal(err)
	}
	data := buf.Bytes()
	// Corrupt the IHDR CRC.
	ihdrEnd := 8 + 8 + 13
	binary.BigEndian.PutUint32(data[ihdrEnd:], crc32.ChecksumIEEE([]byte("wrong")))
	img, err := DecodeImageData(data)
	if err != nil {
		t.Fatal(err)
	}
	if got := img.(*image.NRGBA).NRGBAAt(0, 0); got != (color.NRGBA{1, 2, 3, 255}) {
		t.Fatalf("pixel = %v", got)
	}
}

// TestDecodeAllRealImagesWhenConfigured decodes every BMP/TGA/PNG/JPG that
// lookups can reach under GORO_DATA_DIR (first archive wins).
func TestDecodeAllRealImagesWhenConfigured(t *testing.T) {
	manager := realDataManager(t)
	type job struct {
		archive *GRF
		name    string
	}
	var jobs []job
	seen := map[string]bool{}
	for _, archive := range manager.Archives {
		for _, name := range archive.Names() {
			key := strings.ToLower(name)
			if seen[key] {
				continue
			}
			seen[key] = true
			for _, ext := range []string{".bmp", ".tga", ".png", ".jpg", ".jpeg"} {
				if strings.HasSuffix(key, ext) {
					jobs = append(jobs, job{archive, name})
				}
			}
		}
	}
	if len(jobs) == 0 {
		t.Skip("no images under GORO_DATA_DIR")
	}
	var mu sync.Mutex
	var failures []string
	var wg sync.WaitGroup
	next := make(chan job)
	for w := 0; w < 8; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range next {
				data, err := j.archive.ReadFile(j.name)
				if err == nil {
					_, err = DecodeImageData(data)
				}
				if err != nil {
					mu.Lock()
					failures = append(failures, j.name+": "+err.Error())
					mu.Unlock()
				}
			}
		}()
	}
	for _, j := range jobs {
		next <- j
	}
	close(next)
	wg.Wait()
	for i, f := range failures {
		if i >= 20 {
			break
		}
		t.Error(f)
	}
	if len(failures) > 0 {
		t.Fatalf("%d of %d images failed to decode", len(failures), len(jobs))
	}
	t.Logf("decoded %d images", len(jobs))
}
