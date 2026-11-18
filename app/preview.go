package app

import (
	"os"
	"strconv"
	"strings"

	"github.com/kivutar/goro/game"
	"github.com/kivutar/goro/glog"
	"github.com/kivutar/goro/session"
	gameui "github.com/kivutar/goro/ui"
	"github.com/kivutar/goro/world"
)

// previewEnv names a developer-only switch that skips login and enters a map
// offline so renewal maps and sprites can be checked without a server:
//
//	GORO_PREVIEW_MAP=prontera,156,180,4060
//	GORO_PREVIEW_ACTORS=1002,1115,4218,10001
//	GORO_PREVIEW_WINDOWS=shop
//	GORO_PREVIEW_SCREEN=charselect   (character select with sample characters)
//
// Nothing is sent: the network client is never connected.
const (
	previewEnv       = "GORO_PREVIEW_MAP"
	previewActorsEnv = "GORO_PREVIEW_ACTORS"
	// previewWindowsEnv opens sample windows (comma list) in the preview.
	previewWindowsEnv = "GORO_PREVIEW_WINDOWS"
	previewScreenEnv  = "GORO_PREVIEW_SCREEN"
)

type previewSpec struct {
	Map    string
	X, Y   int
	Job    int16
	Actors []int16
}

func parsePreview(mapValue, actorsValue string) (previewSpec, bool) {
	fields := strings.Split(strings.TrimSpace(mapValue), ",")
	spec := previewSpec{Map: strings.TrimSpace(fields[0]), X: 150, Y: 150}
	if spec.Map == "" {
		return previewSpec{}, false
	}
	ints := make([]int, 0, 3)
	for _, field := range fields[1:] {
		value, err := strconv.Atoi(strings.TrimSpace(field))
		if err != nil {
			return previewSpec{}, false
		}
		ints = append(ints, value)
	}
	if len(ints) >= 2 {
		spec.X, spec.Y = ints[0], ints[1]
	}
	if len(ints) >= 3 {
		spec.Job = int16(ints[2])
	}
	for _, field := range strings.Split(actorsValue, ",") {
		field = strings.TrimSpace(field)
		if field == "" {
			continue
		}
		value, err := strconv.Atoi(field)
		if err != nil {
			return previewSpec{}, false
		}
		spec.Actors = append(spec.Actors, int16(value))
	}
	return spec, true
}

// previewMode returns an offline world mode when the preview switch is set.
func (g *Game) previewMode() game.Mode {
	if strings.EqualFold(strings.TrimSpace(os.Getenv(previewScreenEnv)), "charselect") && !g.cfg.Headless {
		return g.previewCharacterSelect()
	}
	spec, ok := parsePreview(os.Getenv(previewEnv), os.Getenv(previewActorsEnv))
	if !ok || g.cfg.Headless {
		return nil
	}
	glog.Infof("offline map preview map=%s x=%d y=%d job=%d actors=%d", spec.Map, spec.X, spec.Y, spec.Job, len(spec.Actors))
	character := session.Character{ID: 1, Name: "Preview", Job: spec.Job, Level: 99, JobLevel: 50, HP: 100, MaxHP: 100, SP: 100, MaxSP: 100, Hair: 1}
	g.session.Characters = []session.Character{character}
	g.session.Selected = character
	g.session.CharID = character.ID
	g.session.Sex = 1
	g.world.MapName = spec.Map
	g.world.Player.Name = character.Name
	g.world.Player.Job = character.Job
	g.world.Player.Head = character.Hair
	g.world.Player.Sex = 1
	g.world.SetPlayerPosition(spec.X, spec.Y, 0)
	for i, job := range spec.Actors {
		column, row := i%10, i/10
		x, y := spec.X-9+2*column, spec.Y+3+2*row
		g.world.UpsertActor(world.Actor{ID: uint32(100000 + i), Name: strconv.Itoa(int(job)), Job: job, Head: 1, Sex: byte(i % 2), X: x, Y: y, FromX: x, FromY: y, ToX: x, ToY: y})
	}
	mode := game.NewWorldMode()
	if windows := strings.TrimSpace(os.Getenv(previewWindowsEnv)); windows != "" {
		mode.SetPreviewWindows(strings.Split(windows, ","))
	}
	return mode
}

// previewCharacterSelect shows the character select screen offline with
// sample characters (no character server is configured, so none is dialed).
func (g *Game) previewCharacterSelect() game.Mode {
	glog.Infof("offline character select preview")
	// theme:ok sample characters for the offline preview
	g.session.Characters = []session.Character{
		{ID: 1, Slot: 0, Name: "Preview", Job: 4060, Level: 99, JobLevel: 50, HP: 100, MaxHP: 100, SP: 100, MaxSP: 100, Hair: 1, Str: 90, Agi: 30, Vit: 50, Int: 1, Dex: 60, Luk: 9},
		{ID: 2, Slot: 1, Name: "Acolyte", Job: 4, Level: 42, JobLevel: 30, HP: 900, MaxHP: 900, SP: 300, MaxSP: 300, Hair: 5, Str: 1, Agi: 9, Vit: 40, Int: 60, Dex: 30, Luk: 1},
	}
	g.session.Sex = 1
	return game.NewCharacterSelectMode(g.modeContext(), gameui.ChatConsole{})
}
