package effects

import (
	"image"
	"sync"
	"time"

	"github.com/kivutar/goro/client"
	"github.com/kivutar/goro/game/scene"
	"github.com/kivutar/goro/render"
	"github.com/kivutar/goro/res"
)

var (
	missingTextureImageOnce sync.Once
	missingTextureImage     *render.Image
)

func missingTextureRenderImage() *render.Image {
	missingTextureImageOnce.Do(func() {
		missingTextureImage = render.NewImageFromImage(res.MissingCheckerboard())
	})
	return missingTextureImage
}

// IsMissingTexture reports whether img is the shared missing-texture placeholder.
func IsMissingTexture(img *render.Image) bool {
	return img != nil && img == missingTextureRenderImage()
}

// TextureLoaded reports a real texture, not a missing-file placeholder.
func TextureLoaded(img *render.Image) bool {
	return img != nil && !IsMissingTexture(img)
}

// Host connects visual effects to the game's sound, camera and actor presentation.
type Host interface {
	EffectSound(time.Time, Instance, ...string)
	EffectSoundAtPosition(time.Time, float64, float64, float64, ...string)
	EffectSoundGlobal(time.Time, ...string)
	EffectShake(time.Time, time.Duration)
	EffectStarted(client.Context, Instance, time.Time)
	EffectCameraYaw(client.Context) float64
	EffectSpriteTexture(*render.Image) *render.Image
}

// System owns active effects and the resources used to render them.
// Its zero value is ready to use; a host adds sound and camera integration.
type System struct {
	host           Host
	nextRandomSeed uint32
	Active         []Instance
	specs          map[int]Spec
	StrEffects     map[string]*res.STR
	StrEffectMiss  map[string]struct{}
	EffectViews    map[string]*scene.SpriteView
	EffectViewMiss map[string]struct{}
	ActorBlurs     map[uint32][]ActorBlur
	StarFlights    map[uint32]*StarFlight
	whitePixel     *render.Image
	textures       map[string]*render.Image
	textureMiss    map[string]struct{}
}

// Bind shares the world's texture cache, including its map-entry invalidation.
func (m *System) Bind(host Host, textures map[string]*render.Image, missing map[string]struct{}) {
	m.host = host
	m.textures = textures
	m.textureMiss = missing
}

func (m *System) cachedTexture(key string) (*render.Image, bool) {
	if m.textures == nil {
		return nil, false
	}
	texture, ok := m.textures[key]
	return texture, ok
}

func (m *System) storeLoadedTexture(key string, img image.Image, convert func(image.Image) *render.Image) *render.Image {
	if img == nil {
		return nil
	}
	if m.textures == nil {
		m.textures = make(map[string]*render.Image)
	}
	if m.textureMiss == nil {
		m.textureMiss = make(map[string]struct{})
	}
	var texture *render.Image
	if res.IsMissingCheckerboard(img) {
		m.textureMiss[key] = struct{}{}
		texture = missingTextureRenderImage()
	} else {
		texture = convert(img)
	}
	m.textures[key] = texture
	return texture
}

func (m *System) scheduleSound(at time.Time, paths ...string) {
	if m.host != nil {
		m.host.EffectSoundGlobal(at, paths...)
	}
}

func (m *System) scheduleSoundAtWorldEffect(at time.Time, effect Instance, paths ...string) {
	if m.host != nil {
		m.host.EffectSound(at, effect, paths...)
	}
}

func (m *System) scheduleSoundVolumeAtPosition(at time.Time, volume, x, y float64, paths ...string) {
	if m.host != nil {
		m.host.EffectSoundAtPosition(at, volume, x, y, paths...)
	}
}

func (m *System) startCameraShake(at time.Time, duration time.Duration) {
	if m.host != nil {
		m.host.EffectShake(at, duration)
	}
}

func (m *System) applyWorldEffectSideEffects(ctx client.Context, effect Instance, at time.Time) {
	if m.host != nil {
		m.host.EffectStarted(ctx, effect, at)
	}
}

func (m *System) cameraYaw(ctx client.Context) float64 {
	if m.host != nil {
		return m.host.EffectCameraYaw(ctx)
	}
	return scene.CameraYawForMap(ctx)
}

func (m *System) spriteTexture(image *render.Image) *render.Image {
	if m.host != nil {
		return m.host.EffectSpriteTexture(image)
	}
	return image
}
