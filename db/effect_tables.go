package db

import (
	"sync/atomic"
	"time"
)

// Effect layer values for data-driven draw order. Zero means "use the
// component RenderBefore flag" so compiled tables keep today's behaviour.
const (
	EffectLayerAuto    = 0
	EffectLayerGround  = -1
	EffectLayerActor   = 1
	EffectLayerOverlay = 2
)

// Ez2EffectIDBase is the first numeric ID assigned to named ez2streffect.bson
// entries that are not already in the compiled tables.
const Ez2EffectIDBase = 20000

var (
	compiledEffectSpecs    map[int]EffectSpec
	compiledSkillEffects   map[uint16]SkillEffectSpec
	compiledSkillActions   map[uint16]SkillActionKind
	effectTablesGeneration atomic.Uint64
	effectNameIDs          map[string]int
	effectTablesFromData   atomic.Bool
)

func init() {
	compiledEffectSpecs = cloneEffectSpecs(EffectSpecs)
	compiledSkillEffects = cloneSkillEffects(SkillEffects)
	compiledSkillActions = cloneSkillActions(SkillActions)
}

// EffectTablesGeneration changes whenever the live effect tables are replaced.
func EffectTablesGeneration() uint64 { return effectTablesGeneration.Load() }

// EffectTablesUseDataFile reports whether the live tables came from a data
// file or BSON merge rather than the compiled fallback alone.
func EffectTablesUseDataFile() bool { return effectTablesFromData.Load() }

// CompiledEffectTables returns clones of the built-in tables, for export and
// as the fallback when no data file is present.
func CompiledEffectTables() (map[int]EffectSpec, map[uint16]SkillEffectSpec, map[uint16]SkillActionKind) {
	return cloneEffectSpecs(compiledEffectSpecs), cloneSkillEffects(compiledSkillEffects), cloneSkillActions(compiledSkillActions)
}

// EffectNameID returns the numeric ID assigned to a named BSON effect.
func EffectNameID(name string) (int, bool) {
	id, ok := effectNameIDs[name]
	return id, ok
}

// InstallEffectTables replaces the live EffectSpecs, SkillEffects and
// SkillActions maps. Call it from the game thread. Pass fromData=true when
// the tables include a loaded data file or BSON merge.
func InstallEffectTables(specs map[int]EffectSpec, skills map[uint16]SkillEffectSpec, actions map[uint16]SkillActionKind, names map[string]int, fromData bool) {
	if specs == nil {
		specs = map[int]EffectSpec{}
	}
	if skills == nil {
		skills = map[uint16]SkillEffectSpec{}
	}
	if actions == nil {
		actions = map[uint16]SkillActionKind{}
	}
	EffectSpecs = specs
	SkillEffects = skills
	SkillActions = actions
	effectNameIDs = names
	effectTablesFromData.Store(fromData)
	effectTablesGeneration.Add(1)
}

// RestoreCompiledEffectTables installs a clone of the built-in tables.
func RestoreCompiledEffectTables() {
	InstallEffectTables(cloneEffectSpecs(compiledEffectSpecs), cloneSkillEffects(compiledSkillEffects), cloneSkillActions(compiledSkillActions), nil, false)
}

// EffectDrawsBeforeEntities reports whether a component belongs in the ground
// (before entities) pass. A non-zero Layer on the component or spec wins over
// RenderBefore so Crazy Uproar-style actor effects can be data-driven.
func EffectDrawsBeforeEntities(spec EffectSpec, component EffectComponent) bool {
	layer := component.Layer
	if layer == 0 {
		layer = spec.Layer
	}
	switch {
	case layer < 0:
		return true
	case layer > 0:
		return false
	default:
		return component.RenderBefore
	}
}

// EffectDrawZOrder is the sort key within a draw pass.
func EffectDrawZOrder(spec EffectSpec, component EffectComponent) int {
	if component.ZOrder != 0 {
		return component.ZOrder
	}
	return spec.ZOrder
}

// EffectDrawDepth is extra camera-plane depth bias. Zero means the renderer
// keeps its built-in default (for example StrEffectDepthBias on STR quads).
func EffectDrawDepth(spec EffectSpec, component EffectComponent) float64 {
	if component.Depth != 0 {
		return component.Depth
	}
	return spec.Depth
}

func cloneEffectSpecs(in map[int]EffectSpec) map[int]EffectSpec {
	out := make(map[int]EffectSpec, len(in))
	for id, spec := range in {
		out[id] = cloneEffectSpec(spec)
	}
	return out
}

func cloneEffectSpec(spec EffectSpec) EffectSpec {
	spec.SFX = append([]string(nil), spec.SFX...)
	spec.SFXDelays = append([]time.Duration(nil), spec.SFXDelays...)
	if spec.Components != nil {
		components := make([]EffectComponent, len(spec.Components))
		for i, component := range spec.Components {
			components[i] = cloneEffectComponent(component)
		}
		spec.Components = components
	}
	return spec
}

func cloneEffectComponent(component EffectComponent) EffectComponent {
	component.TextureFiles = append([]string(nil), component.TextureFiles...)
	if component.HitRing != nil {
		ring := *component.HitRing
		component.HitRing = &ring
	}
	if component.HitParticles != nil {
		particles := *component.HitParticles
		component.HitParticles = &particles
	}
	return component
}

func cloneSkillEffects(in map[uint16]SkillEffectSpec) map[uint16]SkillEffectSpec {
	out := make(map[uint16]SkillEffectSpec, len(in))
	for id, spec := range in {
		out[id] = cloneSkillEffectSpec(spec)
	}
	return out
}

func cloneSkillEffectSpec(spec SkillEffectSpec) SkillEffectSpec {
	spec.SFX = append([]string(nil), spec.SFX...)
	spec.EffectIDs = append([]int(nil), spec.EffectIDs...)
	spec.EffectIDsOnCaster = append([]int(nil), spec.EffectIDsOnCaster...)
	spec.GroundEffectIDs = append([]int(nil), spec.GroundEffectIDs...)
	spec.HitEffectIDs = append([]int(nil), spec.HitEffectIDs...)
	spec.HitEffectIDsOnCaster = append([]int(nil), spec.HitEffectIDsOnCaster...)
	spec.BeforeHitEffectIDs = append([]int(nil), spec.BeforeHitEffectIDs...)
	spec.BeforeHitEffectIDsSelf = append([]int(nil), spec.BeforeHitEffectIDsSelf...)
	spec.BeginCastEffectIDs = append([]int(nil), spec.BeginCastEffectIDs...)
	spec.SuccessEffectIDs = append([]int(nil), spec.SuccessEffectIDs...)
	spec.SuccessEffectIDsSelf = append([]int(nil), spec.SuccessEffectIDsSelf...)
	return spec
}

func cloneSkillActions(in map[uint16]SkillActionKind) map[uint16]SkillActionKind {
	out := make(map[uint16]SkillActionKind, len(in))
	for id, kind := range in {
		out[id] = kind
	}
	return out
}
