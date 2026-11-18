package res

import (
	"errors"
	"image"
	"image/color"
	"strings"
	"sync"
	"time"

	"github.com/kivutar/goro/glog"
)

const (
	missingCheckerboardSize = 64
	missingCheckerboardCell = 8
)

var (
	missingCheckerboardOnce sync.Once
	missingCheckerboard     *image.NRGBA
	missingTextureLog       sync.Map
)

// MissingCheckerboard is the magenta-and-black placeholder used when a
// texture file is permanently missing. Magenta is an RO chroma key, so this
// image must not go through applyROTransparency or effect black-keying.
func MissingCheckerboard() *image.NRGBA {
	missingCheckerboardOnce.Do(func() {
		img := image.NewNRGBA(image.Rect(0, 0, missingCheckerboardSize, missingCheckerboardSize))
		magenta := color.NRGBA{R: 255, B: 255, A: 255}
		black := color.NRGBA{A: 255}
		for y := 0; y < missingCheckerboardSize; y++ {
			for x := 0; x < missingCheckerboardSize; x++ {
				if ((x/missingCheckerboardCell)+(y/missingCheckerboardCell))%2 == 0 {
					img.SetNRGBA(x, y, magenta)
				} else {
					img.SetNRGBA(x, y, black)
				}
			}
		}
		missingCheckerboard = img
	})
	return missingCheckerboard
}

// IsMissingCheckerboard reports whether img is the shared placeholder.
func IsMissingCheckerboard(img image.Image) bool {
	return img != nil && img == MissingCheckerboard()
}

// RecordMissingTexture logs path once process-wide. It returns true on the
// first recording of that path.
func RecordMissingTexture(path string) bool {
	path = strings.TrimSpace(path)
	if path == "" {
		return false
	}
	if _, loaded := missingTextureLog.LoadOrStore(path, struct{}{}); loaded {
		return false
	}
	glog.Warnf("missing texture path=%s", path)
	return true
}

// LoadTexture is LoadImage with a magenta-and-black checkerboard fallback for
// permanent misses. ErrRemotePending is returned so the caller can retry;
// that state is not logged as a missing texture.
func LoadTexture(manager *Manager, candidates []string) (image.Image, string, error) {
	img, source, err := LoadImage(manager, candidates)
	return textureFromLoad(img, source, err, candidates)
}

// LoadTextureExact is LoadImageExact with the missing-texture fallback.
func LoadTextureExact(manager *Manager, candidates []string) (image.Image, string, error) {
	img, source, err := LoadImageExact(manager, candidates)
	return textureFromLoad(img, source, err, candidates)
}

// LoadTextureWithin is LoadImageWithin with the missing-texture fallback.
func LoadTextureWithin(manager *Manager, candidates []string, budget time.Duration) (image.Image, string, error) {
	img, source, err := LoadImageWithin(manager, candidates, budget)
	return textureFromLoad(img, source, err, candidates)
}

func textureFromLoad(img image.Image, source string, err error, candidates []string) (image.Image, string, error) {
	if err == nil {
		return img, source, nil
	}
	if errors.Is(err, ErrRemotePending) {
		return nil, source, err
	}
	path := strings.TrimSpace(source)
	if path == "" && len(candidates) > 0 {
		path = candidates[0]
	}
	RecordMissingTexture(path)
	return MissingCheckerboard(), path, nil
}
