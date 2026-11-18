package maps

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"math/rand/v2"
	"strings"
	"time"

	"github.com/kivutar/goro/client"
	"github.com/kivutar/goro/render"
	"github.com/kivutar/goro/res"
	"github.com/kivutar/goro/ui/rotheme"
)

const (
	loadingMap = iota
	loadingModels
	loadingTextures
	loadingSounds
	loadingPrepared
)

// Loading prepares one map on the game thread, yielding between assets so the
// loading screen can redraw. GPU uploads and scene warmup follow preparation.
type Loading struct {
	RSWSource string
	Presented bool
	stage     int
	tasks     []func()
	completed int
	uploads   int
	images    render.ImageGroup
	// prefetch downloads the next stage's streamed files in parallel; the
	// loading screen keeps drawing until it is done.
	prefetch   *res.RemoteBatch
	background *render.Image
	unblockUI  func()
}

func NewLoading(ctx client.Context) *Loading {
	l := &Loading{RSWSource: ctx.World.MapName}
	l.blockUI(ctx)
	if ctx.Resources != nil {
		for _, index := range append(rand.Perm(10), -1) {
			name := fmt.Sprintf("loading%02d.jpg", index+1)
			candidates := []string{
				"data/texture/유저인터페이스/" + name,
				"texture/유저인터페이스/" + name,
				"data/texture/interface/" + name,
				name,
			}
			for _, candidate := range candidates {
				data, err := ctx.Resources.ReadFileExact(candidate)
				if err != nil {
					continue
				}
				// Loading artwork is opaque: sprite magenta-keying does not apply.
				if img, _, err := image.Decode(bytes.NewReader(data)); err == nil {
					l.background = l.images.Add(render.NewImageFromImage(img))
					break
				}
			}
			if l.background != nil {
				break
			}
		}
	}
	l.prefetch = ctx.Resources.PrefetchRemote(mapFileCandidates(ctx.World.MapName))
	l.tasks = []func(){
		func() { ctx.World.GND, _, _ = LoadGND(ctx.Resources, ctx.World.MapName) },
		func() {
			rsw, source, err := LoadRSW(ctx.Resources, ctx.World.MapName)
			if err == nil {
				ctx.World.RSW = rsw
				l.RSWSource = source
			}
		},
	}
	return l
}

func (l *Loading) Prepared() bool { return l.stage == loadingPrepared }

// Advance reports completion once. Each batch waits for a submitted frame;
// skipped GPU frames cannot turn loading back into a blocking loop.
func (l *Loading) Advance(ctx client.Context, scene *System) bool {
	l.blockUI(ctx)
	if !l.Presented || l.Prepared() {
		return false
	}
	if !l.prefetch.Done() {
		return false
	}
	l.Presented = false
	deadline := time.Now().Add(8 * time.Millisecond)
	for l.completed < len(l.tasks) {
		l.tasks[l.completed]()
		l.tasks[l.completed] = nil
		l.completed++
		if time.Now().After(deadline) {
			break
		}
	}
	if l.completed < len(l.tasks) {
		return false
	}
	l.stage++
	l.completed = 0
	switch l.stage {
	case loadingModels:
		l.tasks = ModelLoadTasks(ctx)
		l.prefetch = ctx.Resources.PrefetchRemote(modelFileCandidates(ctx))
	case loadingTextures:
		l.tasks = scene.TextureLoadTasks(ctx)
		l.prefetch = ctx.Resources.PrefetchRemote(textureFileCandidates(ctx))
	case loadingSounds:
		l.tasks = SoundLoadTasks(ctx)
	case loadingPrepared:
		l.tasks = nil
		l.uploads = len(scene.Images.Uploads)
		return true
	}
	return false
}

func (l *Loading) Percent(pendingUploads, coveredFrames int) int {
	if l.Prepared() {
		progress := 98
		if l.uploads > 0 {
			progress -= 8 * pendingUploads / l.uploads
		}
		return min(100, progress+2*coveredFrames/FadePrewarmFrames) // theme:ok progress percentage, not a size
	}
	// Stage weights reserve the final ten percent for uploads and warmup.
	limits := [...]int{0, 10, 30, 85, 90}
	progress := limits[l.stage]
	if len(l.tasks) > 0 {
		progress += (limits[l.stage+1] - progress) * l.completed / len(l.tasks)
	}
	return progress
}

func (l *Loading) Draw(screen *render.Frame, percent int, alpha uint8) {
	width, height := float64(screen.Bounds().Dx()), float64(screen.Bounds().Dy())
	render.DrawRect(screen, 0, 0, width, height, fadedLoadingColor(rotheme.LoadingBackground, alpha))
	if l.background != nil {
		var opts render.DrawImageOptions
		opts.GeoM.Scale(width/float64(l.background.Bounds().Dx()), height/float64(l.background.Bounds().Dy()))
		opts.ColorScale.ScaleAlpha(float32(alpha) / 255) // theme:ok alpha byte to fraction
		opts.Filter = render.FilterLinear
		screen.DrawImage(l.background, &opts)
	}
	w, h := render.SnapScreenPoint(screen, min(rotheme.LoadingBarMaxWidth, width-rotheme.LoadingBarSideMargin), rotheme.LoadingBarHeight)
	x, y := render.SnapScreenPoint(screen, (width-w)/2, height*rotheme.LoadingBarYRatio)
	// Snap each inset once so opposite edges have the same physical thickness.
	borderX, borderY := render.SnapScreenPoint(screen, rotheme.LoadingBarBorder, rotheme.LoadingBarBorder)
	padX, padY := render.SnapScreenPoint(screen, rotheme.LoadingBarPadding, rotheme.LoadingBarPadding)
	render.DrawRect(screen, x, y, w, h, fadedLoadingColor(rotheme.LoadingBarBorderColor, alpha))
	render.DrawRect(screen, x+borderX, y+borderY, w-2*borderX, h-2*borderY, fadedLoadingColor(rotheme.LoadingBarTrack, alpha))           // theme:ok inset on both edges
	render.DrawRect(screen, x+padX, y+padY, float64(percent)*(w-2*padX)/100, h-2*padY, fadedLoadingColor(rotheme.LoadingBarFill, alpha)) // theme:ok inset on both edges, percent to fraction
	text := fmt.Sprintf("%d%%", percent)
	// Overlay labels include padding around the glyphs.
	render.DrawCenteredUITextAtSize(screen, text, width/2, y-rotheme.LoadingLabelGap, fadedLoadingColor(rotheme.LoadingLabelColor, alpha), rotheme.LoadingLabelSize, false)
}

// fadedLoadingColor scales a theme color alpha by the screen fade alpha. With
// the default opaque colors this is exactly the fade alpha.
func fadedLoadingColor(c color.RGBA, alpha uint8) color.RGBA {
	c.A = uint8(uint32(c.A) * uint32(alpha) / 255) // theme:ok alpha channel scale
	return c
}

func (l *Loading) Release() {
	if l.unblockUI != nil {
		l.unblockUI()
		l.unblockUI = nil
	}
	l.images.Release()
	l.background = nil
	l.tasks = nil
}

func (l *Loading) blockUI(ctx client.Context) {
	// Startup can create the map before the renderer attaches the UI app.
	if app, ok := ctx.UIApp.(interface{ BlockInput() func() }); ok && l.unblockUI == nil {
		l.unblockUI = app.BlockInput()
	}
}

func mapFileCandidates(mapName string) [][]string {
	base := strings.TrimSuffix(strings.TrimSuffix(mapName, ".gat"), ".rsw")
	var lists [][]string
	for _, ext := range []string{"gnd", "rsw", "gat"} {
		lists = append(lists, []string{"data\\" + base + "." + ext})
	}
	return lists
}

func modelFileCandidates(ctx client.Context) [][]string {
	if ctx.World == nil || ctx.World.RSW == nil {
		return nil
	}
	var lists [][]string
	for _, placement := range ctx.World.RSW.Models {
		if placement.Filename != "" {
			lists = append(lists, res.RSMModelCandidates(placement.Filename))
		}
	}
	return lists
}

func textureFileCandidates(ctx client.Context) [][]string {
	if ctx.Config.Headless || ctx.World == nil {
		return nil
	}
	var lists [][]string
	if ctx.World.GND != nil {
		for _, name := range ctx.World.GND.Textures {
			lists = append(lists, res.GroundTextureCandidates(name))
		}
	}
	for _, model := range ctx.World.RSM {
		if model == nil {
			continue
		}
		for _, name := range model.Textures {
			lists = append(lists, res.GroundTextureCandidates(name))
		}
	}
	return lists
}
