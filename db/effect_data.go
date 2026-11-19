package db

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// EffectTables is one overlay of EffectSpecs, SkillEffects and SkillActions.
type EffectTables struct {
	Effects      map[int]EffectSpec
	SkillEffects map[uint16]SkillEffectSpec
	SkillActions map[uint16]SkillActionKind
}

type effectTablesFile struct {
	Effects      map[string]json.RawMessage `json:"effects"`
	SkillEffects map[string]json.RawMessage `json:"skill_effects"`
	SkillActions map[string]string          `json:"skill_actions"`
}

type effectSpecJSON struct {
	DurationNS         *int64            `json:"duration_ns"`
	DurationMS         *float64          `json:"duration_ms"`
	Duration           *int64            `json:"Duration"`
	CameraShakeNS      *int64            `json:"camera_shake_ns"`
	CameraShakeMS      *float64          `json:"camera_shake_ms"`
	CameraShake        *int64            `json:"CameraShake"`
	CameraShakeDelayNS *int64            `json:"camera_shake_delay_ns"`
	CameraShakeDelayMS *float64          `json:"camera_shake_delay_ms"`
	CameraShakeDelay   *int64            `json:"CameraShakeDelay"`
	DetachLocalActor   bool              `json:"detach_local_actor"`
	DetachLocalActorGo bool              `json:"DetachLocalActor"`
	SFX                []string          `json:"sfx"`
	SFXGo              []string          `json:"SFX"`
	SFXDelaysNS        []int64           `json:"sfx_delays_ns"`
	SFXDelaysMS        []float64         `json:"sfx_delays_ms"`
	SFXDelaysGo        []int64           `json:"SFXDelays"`
	SFXRandMin         int               `json:"sfx_rand_min"`
	SFXRandMinGo       int               `json:"SFXRandMin"`
	SFXRandMax         int               `json:"sfx_rand_max"`
	SFXRandMaxGo       int               `json:"SFXRandMax"`
	Layer              int               `json:"layer"`
	LayerGo            int               `json:"Layer"`
	ZOrder             int               `json:"z_order"`
	ZOrderGo           int               `json:"ZOrder"`
	Depth              float64           `json:"depth"`
	DepthGo            float64           `json:"Depth"`
	Components         []json.RawMessage `json:"components"`
	ComponentsGo       []json.RawMessage `json:"Components"`
}

type effectComponentJSON struct {
	Kind             string          `json:"kind"`
	KindGo           json.RawMessage `json:"Kind"`
	FuncName         string          `json:"func"`
	FuncNameGo       string          `json:"FuncName"`
	DurationNS       *int64          `json:"duration_ns"`
	DurationMS       *float64        `json:"duration_ms"`
	Duration         *int64          `json:"Duration"`
	DelayNS          *int64          `json:"delay_ns"`
	DelayMS          *float64        `json:"delay_ms"`
	Delay            *int64          `json:"Delay"`
	RepeatDelayNS    *int64          `json:"repeat_delay_ns"`
	RepeatDelay      *int64          `json:"RepeatDelay"`
	DuplicateDelayNS *int64          `json:"duplicate_delay_ns"`
	DuplicateDelay   *int64          `json:"DuplicateDelay"`
	Repeat           bool            `json:"repeat"`
	RepeatGo         bool            `json:"Repeat"`
	STRFile          string          `json:"file"`
	STRFileGo        string          `json:"STRFile"`
	STRMinFile       string          `json:"min_file"`
	STRMinFileGo     string          `json:"STRMinFile"`
	STRRandMin       int             `json:"str_rand_min"`
	STRRandMinGo     int             `json:"STRRandMin"`
	STRRandMax       int             `json:"str_rand_max"`
	STRRandMaxGo     int             `json:"STRRandMax"`
	AttachedEntity   bool            `json:"attached_entity"`
	AttachedEntityGo bool            `json:"AttachedEntity"`
	RenderBefore     bool            `json:"render_before"`
	RenderBeforeGo   bool            `json:"RenderBefore"`
	Layer            int             `json:"layer"`
	LayerGo          int             `json:"Layer"`
	ZOrder           int             `json:"z_order"`
	ZOrderGo         int             `json:"ZOrder"`
	Depth            float64         `json:"depth"`
	DepthGo          float64         `json:"Depth"`
	TexturePath      string          `json:"texture_path"`
	TexturePathGo    string          `json:"TexturePath"`
	TextureName      string          `json:"texture_name"`
	TextureNameGo    string          `json:"TextureName"`
	TextureFile      string          `json:"texture_file"`
	TextureFileGo    string          `json:"TextureFile"`
	TextureFiles     []string        `json:"texture_files"`
	TextureFilesGo   []string        `json:"TextureFiles"`
	SpriteFile       string          `json:"sprite"`
	SpriteFileGo     string          `json:"SpriteFile"`
	SpriteHead       bool            `json:"sprite_head"`
	SpriteHeadGo     bool            `json:"SpriteHead"`
	SpriteRepeat     bool            `json:"sprite_repeat"`
	SpriteRepeatGo   bool            `json:"SpriteRepeat"`
	SpriteYOffset    float64         `json:"sprite_y_offset"`
	SpriteYOffsetGo  float64         `json:"SpriteYOffset"`
	SpriteXOffset    float64         `json:"sprite_x_offset"`
	SpriteXOffsetGo  float64         `json:"SpriteXOffset"`
	PosX             float64         `json:"pos_x"`
	PosXGo           float64         `json:"PosX"`
	PosY             float64         `json:"pos_y"`
	PosYGo           float64         `json:"PosY"`
	PosZ             float64         `json:"pos_z"`
	PosZGo           float64         `json:"PosZ"`
	Color            []int           `json:"color"`
	ColorGo          *struct {
		R, G, B, A uint8
	} `json:"Color"`
	Fade               bool                `json:"fade"`
	FadeGo             bool                `json:"Fade"`
	FadeIn             bool                `json:"fade_in"`
	FadeInGo           bool                `json:"FadeIn"`
	FadeOut            bool                `json:"fade_out"`
	FadeOutGo          bool                `json:"FadeOut"`
	BlendAdditive      bool                `json:"blend_additive"`
	BlendAdditiveGo    bool                `json:"BlendAdditive"`
	Overlay            bool                `json:"overlay"`
	OverlayGo          bool                `json:"Overlay"`
	WorldSizedSprite   bool                `json:"world_sized_sprite"`
	WorldSizedSpriteGo bool                `json:"WorldSizedSprite"`
	AlphaMax           float64             `json:"alpha_max"`
	AlphaMaxGo         float64             `json:"AlphaMax"`
	HitRing            *EffectHitRing      `json:"hit_ring"`
	HitRingGo          *EffectHitRing      `json:"HitRing"`
	HitParticles       *EffectHitParticles `json:"hit_particles"`
	HitParticlesGo     *EffectHitParticles `json:"HitParticles"`
}

type skillEffectSpecJSON struct {
	SFX                      []string `json:"sfx"`
	SFXGo                    []string `json:"SFX"`
	EffectTiming             string   `json:"effect_timing"`
	EffectTimingGo           int      `json:"EffectTiming"`
	EffectIDs                []int    `json:"effect_ids"`
	EffectIDsGo              []int    `json:"EffectIDs"`
	EffectIDsOnCaster        []int    `json:"effect_ids_on_caster"`
	EffectIDsOnCasterGo      []int    `json:"EffectIDsOnCaster"`
	GroundEffectIDs          []int    `json:"ground_effect_ids"`
	GroundEffectIDsGo        []int    `json:"GroundEffectIDs"`
	HitEffectIDs             []int    `json:"hit_effect_ids"`
	HitEffectIDsGo           []int    `json:"HitEffectIDs"`
	HitEffectIDsOnCaster     []int    `json:"hit_effect_ids_on_caster"`
	HitEffectIDsOnCasterGo   []int    `json:"HitEffectIDsOnCaster"`
	BeforeHitEffectIDs       []int    `json:"before_hit_effect_ids"`
	BeforeHitEffectIDsGo     []int    `json:"BeforeHitEffectIDs"`
	BeforeHitEffectIDsSelf   []int    `json:"before_hit_effect_ids_self"`
	BeforeHitEffectIDsSelfGo []int    `json:"BeforeHitEffectIDsSelf"`
	BeginCastEffectIDs       []int    `json:"begin_cast_effect_ids"`
	BeginCastEffectIDsGo     []int    `json:"BeginCastEffectIDs"`
	SuccessEffectIDs         []int    `json:"success_effect_ids"`
	SuccessEffectIDsGo       []int    `json:"SuccessEffectIDs"`
	SuccessEffectIDsSelf     []int    `json:"success_effect_ids_self"`
	SuccessEffectIDsSelfGo   []int    `json:"SuccessEffectIDsSelf"`
	HideCastBar              bool     `json:"hide_cast_bar"`
	HideCastBarGo            bool     `json:"HideCastBar"`
	HideCastAura             bool     `json:"hide_cast_aura"`
	HideCastAuraGo           bool     `json:"HideCastAura"`
}

// ParseEffectTablesJSON decodes an effects overlay. Unknown kinds and
// non-numeric IDs are errors so a bad file can be rejected without installing.
func ParseEffectTablesJSON(data []byte) (EffectTables, error) {
	var file effectTablesFile
	if err := json.Unmarshal(data, &file); err != nil {
		return EffectTables{}, fmt.Errorf("effects.json: %w", err)
	}
	out := EffectTables{
		Effects:      map[int]EffectSpec{},
		SkillEffects: map[uint16]SkillEffectSpec{},
		SkillActions: map[uint16]SkillActionKind{},
	}
	for key, raw := range file.Effects {
		id, err := strconv.Atoi(key)
		if err != nil {
			return EffectTables{}, fmt.Errorf("effects.json: effect id %q: %w", key, err)
		}
		spec, err := parseEffectSpecJSON(raw)
		if err != nil {
			return EffectTables{}, fmt.Errorf("effects.json: effect %d: %w", id, err)
		}
		out.Effects[id] = spec
	}
	for key, raw := range file.SkillEffects {
		id, err := parseSkillID(key)
		if err != nil {
			return EffectTables{}, fmt.Errorf("effects.json: skill effect id %q: %w", key, err)
		}
		spec, err := parseSkillEffectSpecJSON(raw)
		if err != nil {
			return EffectTables{}, fmt.Errorf("effects.json: skill %d: %w", id, err)
		}
		out.SkillEffects[id] = spec
	}
	for key, name := range file.SkillActions {
		id, err := parseSkillID(key)
		if err != nil {
			return EffectTables{}, fmt.Errorf("effects.json: skill action id %q: %w", key, err)
		}
		kind, err := parseSkillActionKind(name)
		if err != nil {
			return EffectTables{}, fmt.Errorf("effects.json: skill %d action: %w", id, err)
		}
		out.SkillActions[id] = kind
	}
	return out, nil
}

// MarshalEffectTablesJSON writes tables in the overlay schema (pretty-printed).
func MarshalEffectTablesJSON(tables EffectTables) ([]byte, error) {
	file := effectTablesFile{
		Effects:      map[string]json.RawMessage{},
		SkillEffects: map[string]json.RawMessage{},
		SkillActions: map[string]string{},
	}
	for id, spec := range tables.Effects {
		raw, err := marshalEffectSpecJSON(spec)
		if err != nil {
			return nil, err
		}
		file.Effects[strconv.Itoa(id)] = raw
	}
	for id, spec := range tables.SkillEffects {
		raw, err := marshalSkillEffectSpecJSON(spec)
		if err != nil {
			return nil, err
		}
		file.SkillEffects[strconv.Itoa(int(id))] = raw
	}
	for id, kind := range tables.SkillActions {
		file.SkillActions[strconv.Itoa(int(id))] = skillActionKindName(kind)
	}
	return json.MarshalIndent(file, "", "  ")
}

// OverlayEffectTables clones the compiled tables and replaces each ID present
// in overlay. Missing overlay sections leave the compiled entries in place.
func OverlayEffectTables(overlay EffectTables) (map[int]EffectSpec, map[uint16]SkillEffectSpec, map[uint16]SkillActionKind) {
	specs, skills, actions := CompiledEffectTables()
	for id, spec := range overlay.Effects {
		specs[id] = cloneEffectSpec(spec)
	}
	for id, spec := range overlay.SkillEffects {
		skills[id] = cloneSkillEffectSpec(spec)
	}
	for id, kind := range overlay.SkillActions {
		actions[id] = kind
	}
	return specs, skills, actions
}

func parseSkillID(key string) (uint16, error) {
	id, err := strconv.Atoi(key)
	if err != nil {
		return 0, err
	}
	if id < 0 || id > 0xffff {
		return 0, fmt.Errorf("out of range")
	}
	return uint16(id), nil
}

func parseEffectSpecJSON(raw json.RawMessage) (EffectSpec, error) {
	var spec EffectSpec
	if err := json.Unmarshal(raw, &spec); err != nil {
		spec = EffectSpec{}
	}
	var aux effectSpecJSON
	if err := json.Unmarshal(raw, &aux); err != nil {
		if len(spec.Components) == 0 && spec.Duration == 0 && spec.Layer == 0 && len(spec.SFX) == 0 {
			return EffectSpec{}, err
		}
		return spec, nil
	}
	if spec.Duration == 0 {
		spec.Duration = pickDuration(aux.DurationNS, aux.DurationMS, aux.Duration)
	}
	if spec.CameraShake == 0 {
		spec.CameraShake = pickDuration(aux.CameraShakeNS, aux.CameraShakeMS, aux.CameraShake)
	}
	if spec.CameraShakeDelay == 0 {
		spec.CameraShakeDelay = pickDuration(aux.CameraShakeDelayNS, aux.CameraShakeDelayMS, aux.CameraShakeDelay)
	}
	if !spec.DetachLocalActor {
		spec.DetachLocalActor = aux.DetachLocalActor || aux.DetachLocalActorGo
	}
	if len(spec.SFX) == 0 {
		spec.SFX = firstStrings(aux.SFX, aux.SFXGo)
	}
	if spec.SFXRandMin == 0 {
		spec.SFXRandMin = firstInt(aux.SFXRandMin, aux.SFXRandMinGo)
	}
	if spec.SFXRandMax == 0 {
		spec.SFXRandMax = firstInt(aux.SFXRandMax, aux.SFXRandMaxGo)
	}
	if spec.Layer == 0 {
		spec.Layer = firstInt(aux.Layer, aux.LayerGo)
	}
	if spec.ZOrder == 0 {
		spec.ZOrder = firstInt(aux.ZOrder, aux.ZOrderGo)
	}
	if spec.Depth == 0 {
		spec.Depth = firstFloat(aux.Depth, aux.DepthGo)
	}
	if len(spec.SFXDelays) == 0 {
		if len(aux.SFXDelaysNS) > 0 {
			spec.SFXDelays = durationsFromNS(aux.SFXDelaysNS)
		} else if len(aux.SFXDelaysMS) > 0 {
			spec.SFXDelays = durationsFromMS(aux.SFXDelaysMS)
		} else if len(aux.SFXDelaysGo) > 0 {
			spec.SFXDelays = durationsFromNS(aux.SFXDelaysGo)
		}
	}
	components := aux.Components
	if len(components) == 0 {
		components = aux.ComponentsGo
	}
	if len(components) > 0 {
		parsed := make([]EffectComponent, 0, len(components))
		for i, craw := range components {
			component, err := parseEffectComponentJSON(craw)
			if err != nil {
				return EffectSpec{}, fmt.Errorf("component %d: %w", i, err)
			}
			parsed = append(parsed, component)
		}
		spec.Components = parsed
	}
	return spec, nil
}

func parseEffectComponentJSON(raw json.RawMessage) (EffectComponent, error) {
	var c EffectComponent
	_ = json.Unmarshal(raw, &c)
	var aux effectComponentJSON
	if err := json.Unmarshal(raw, &aux); err != nil {
		if c.Kind == 0 {
			return EffectComponent{}, err
		}
		return c, nil
	}
	kind := c.Kind
	if kind == 0 {
		parsed, err := parseComponentKind(aux.Kind, aux.KindGo)
		if err != nil {
			return EffectComponent{}, err
		}
		kind = parsed
	}
	if c.FuncName == "" {
		c.FuncName = firstString(aux.FuncName, aux.FuncNameGo)
	}
	if c.Duration == 0 {
		c.Duration = pickDuration(aux.DurationNS, aux.DurationMS, aux.Duration)
	}
	if c.Delay == 0 {
		c.Delay = pickDuration(aux.DelayNS, aux.DelayMS, aux.Delay)
	}
	if c.RepeatDelay == 0 {
		c.RepeatDelay = pickDuration(aux.RepeatDelayNS, nil, aux.RepeatDelay)
	}
	if c.DuplicateDelay == 0 {
		c.DuplicateDelay = pickDuration(aux.DuplicateDelayNS, nil, aux.DuplicateDelay)
	}
	c.Repeat = c.Repeat || aux.Repeat || aux.RepeatGo
	if c.STRFile == "" {
		c.STRFile = firstString(aux.STRFile, aux.STRFileGo)
	}
	if c.STRMinFile == "" {
		c.STRMinFile = firstString(aux.STRMinFile, aux.STRMinFileGo)
	}
	if c.STRRandMin == 0 {
		c.STRRandMin = firstInt(aux.STRRandMin, aux.STRRandMinGo)
	}
	if c.STRRandMax == 0 {
		c.STRRandMax = firstInt(aux.STRRandMax, aux.STRRandMaxGo)
	}
	c.AttachedEntity = c.AttachedEntity || aux.AttachedEntity || aux.AttachedEntityGo
	c.RenderBefore = c.RenderBefore || aux.RenderBefore || aux.RenderBeforeGo
	if c.Layer == 0 {
		c.Layer = firstInt(aux.Layer, aux.LayerGo)
	}
	if c.ZOrder == 0 {
		c.ZOrder = firstInt(aux.ZOrder, aux.ZOrderGo)
	}
	if c.Depth == 0 {
		c.Depth = firstFloat(aux.Depth, aux.DepthGo)
	}
	if c.TexturePath == "" {
		c.TexturePath = firstString(aux.TexturePath, aux.TexturePathGo)
	}
	if c.TextureName == "" {
		c.TextureName = firstString(aux.TextureName, aux.TextureNameGo)
	}
	if c.TextureFile == "" {
		c.TextureFile = firstString(aux.TextureFile, aux.TextureFileGo)
	}
	if len(c.TextureFiles) == 0 {
		c.TextureFiles = firstStrings(aux.TextureFiles, aux.TextureFilesGo)
	}
	if c.SpriteFile == "" {
		c.SpriteFile = firstString(aux.SpriteFile, aux.SpriteFileGo)
	}
	c.SpriteHead = c.SpriteHead || aux.SpriteHead || aux.SpriteHeadGo
	c.SpriteRepeat = c.SpriteRepeat || aux.SpriteRepeat || aux.SpriteRepeatGo
	if c.SpriteYOffset == 0 {
		c.SpriteYOffset = firstFloat(aux.SpriteYOffset, aux.SpriteYOffsetGo)
	}
	if c.SpriteXOffset == 0 {
		c.SpriteXOffset = firstFloat(aux.SpriteXOffset, aux.SpriteXOffsetGo)
	}
	if c.PosX == 0 {
		c.PosX = firstFloat(aux.PosX, aux.PosXGo)
	}
	if c.PosY == 0 {
		c.PosY = firstFloat(aux.PosY, aux.PosYGo)
	}
	if c.PosZ == 0 {
		c.PosZ = firstFloat(aux.PosZ, aux.PosZGo)
	}
	c.Fade = c.Fade || aux.Fade || aux.FadeGo
	c.FadeIn = c.FadeIn || aux.FadeIn || aux.FadeInGo
	c.FadeOut = c.FadeOut || aux.FadeOut || aux.FadeOutGo
	c.BlendAdditive = c.BlendAdditive || aux.BlendAdditive || aux.BlendAdditiveGo
	c.Overlay = c.Overlay || aux.Overlay || aux.OverlayGo
	c.WorldSizedSprite = c.WorldSizedSprite || aux.WorldSizedSprite || aux.WorldSizedSpriteGo
	if c.AlphaMax == 0 {
		c.AlphaMax = firstFloat(aux.AlphaMax, aux.AlphaMaxGo)
	}
	if c.HitRing == nil {
		c.HitRing = firstHitRing(aux.HitRing, aux.HitRingGo)
	}
	if c.HitParticles == nil {
		c.HitParticles = firstHitParticles(aux.HitParticles, aux.HitParticlesGo)
	}
	if c.Color.A == 0 && c.Color.R == 0 && c.Color.G == 0 && c.Color.B == 0 {
		if len(aux.Color) >= 4 {
			c.Color.R, c.Color.G, c.Color.B, c.Color.A = uint8(aux.Color[0]), uint8(aux.Color[1]), uint8(aux.Color[2]), uint8(aux.Color[3])
		} else if aux.ColorGo != nil {
			c.Color.R, c.Color.G, c.Color.B, c.Color.A = aux.ColorGo.R, aux.ColorGo.G, aux.ColorGo.B, aux.ColorGo.A
		}
	}
	c.Kind = kind
	if c.Kind == 0 {
		return EffectComponent{}, fmt.Errorf("missing kind")
	}
	return c, nil
}

func parseSkillEffectSpecJSON(raw json.RawMessage) (SkillEffectSpec, error) {
	var spec SkillEffectSpec
	_ = json.Unmarshal(raw, &spec)
	var aux skillEffectSpecJSON
	if err := json.Unmarshal(raw, &aux); err != nil {
		return spec, nil
	}
	if len(spec.SFX) == 0 {
		spec.SFX = firstStrings(aux.SFX, aux.SFXGo)
	}
	if len(spec.EffectIDs) == 0 {
		spec.EffectIDs = firstInts(aux.EffectIDs, aux.EffectIDsGo)
	}
	if len(spec.EffectIDsOnCaster) == 0 {
		spec.EffectIDsOnCaster = firstInts(aux.EffectIDsOnCaster, aux.EffectIDsOnCasterGo)
	}
	if len(spec.GroundEffectIDs) == 0 {
		spec.GroundEffectIDs = firstInts(aux.GroundEffectIDs, aux.GroundEffectIDsGo)
	}
	if len(spec.HitEffectIDs) == 0 {
		spec.HitEffectIDs = firstInts(aux.HitEffectIDs, aux.HitEffectIDsGo)
	}
	if len(spec.HitEffectIDsOnCaster) == 0 {
		spec.HitEffectIDsOnCaster = firstInts(aux.HitEffectIDsOnCaster, aux.HitEffectIDsOnCasterGo)
	}
	if len(spec.BeforeHitEffectIDs) == 0 {
		spec.BeforeHitEffectIDs = firstInts(aux.BeforeHitEffectIDs, aux.BeforeHitEffectIDsGo)
	}
	if len(spec.BeforeHitEffectIDsSelf) == 0 {
		spec.BeforeHitEffectIDsSelf = firstInts(aux.BeforeHitEffectIDsSelf, aux.BeforeHitEffectIDsSelfGo)
	}
	if len(spec.BeginCastEffectIDs) == 0 {
		spec.BeginCastEffectIDs = firstInts(aux.BeginCastEffectIDs, aux.BeginCastEffectIDsGo)
	}
	if len(spec.SuccessEffectIDs) == 0 {
		spec.SuccessEffectIDs = firstInts(aux.SuccessEffectIDs, aux.SuccessEffectIDsGo)
	}
	if len(spec.SuccessEffectIDsSelf) == 0 {
		spec.SuccessEffectIDsSelf = firstInts(aux.SuccessEffectIDsSelf, aux.SuccessEffectIDsSelfGo)
	}
	if !spec.HideCastBar {
		spec.HideCastBar = aux.HideCastBar || aux.HideCastBarGo
	}
	if !spec.HideCastAura {
		spec.HideCastAura = aux.HideCastAura || aux.HideCastAuraGo
	}
	if spec.EffectTiming == SkillEffectAtHit {
		switch strings.ToLower(strings.TrimSpace(aux.EffectTiming)) {
		case "action_start":
			spec.EffectTiming = SkillEffectAtActionStart
		case "", "hit":
		default:
			if aux.EffectTiming != "" {
				return SkillEffectSpec{}, fmt.Errorf("unknown effect_timing %q", aux.EffectTiming)
			}
		}
		if aux.EffectTimingGo == int(SkillEffectAtActionStart) {
			spec.EffectTiming = SkillEffectAtActionStart
		}
	}
	return spec, nil
}

func marshalEffectSpecJSON(spec EffectSpec) (json.RawMessage, error) {
	return json.Marshal(spec)
}

func marshalSkillEffectSpecJSON(spec SkillEffectSpec) (json.RawMessage, error) {
	return json.Marshal(spec)
}

func pickDuration(ns *int64, ms *float64, goNS *int64) time.Duration {
	if ns != nil {
		return time.Duration(*ns)
	}
	if goNS != nil {
		return time.Duration(*goNS)
	}
	if ms != nil {
		return time.Duration(*ms * float64(time.Millisecond))
	}
	return 0
}

func durationsFromNS(values []int64) []time.Duration {
	out := make([]time.Duration, len(values))
	for i, v := range values {
		out[i] = time.Duration(v)
	}
	return out
}

func durationsFromMS(values []float64) []time.Duration {
	out := make([]time.Duration, len(values))
	for i, v := range values {
		out[i] = time.Duration(v * float64(time.Millisecond))
	}
	return out
}

func firstString(a, b string) string {
	if a != "" {
		return a
	}
	return b
}

func firstStrings(a, b []string) []string {
	if len(a) > 0 {
		return a
	}
	return b
}

func firstInts(a, b []int) []int {
	if len(a) > 0 {
		return a
	}
	return b
}

func firstInt(a, b int) int {
	if a != 0 {
		return a
	}
	return b
}

func firstFloat(a, b float64) float64 {
	if a != 0 {
		return a
	}
	return b
}

func firstHitRing(a, b *EffectHitRing) *EffectHitRing {
	if a != nil {
		return a
	}
	return b
}

func firstHitParticles(a, b *EffectHitParticles) *EffectHitParticles {
	if a != nil {
		return a
	}
	return b
}

func (k EffectComponentKind) MarshalJSON() ([]byte, error) {
	if name := componentKindName(k); name != "" {
		return json.Marshal(name)
	}
	return json.Marshal(int(k))
}

func (k *EffectComponentKind) UnmarshalJSON(data []byte) error {
	kind, err := parseComponentKind("", data)
	if err != nil {
		var name string
		if json.Unmarshal(data, &name) == nil {
			kind, err = parseComponentKind(name, nil)
		}
	}
	if err != nil {
		return err
	}
	*k = kind
	return nil
}

func (t SkillEffectTiming) MarshalJSON() ([]byte, error) {
	if t == SkillEffectAtActionStart {
		return json.Marshal("action_start")
	}
	return json.Marshal("hit")
}

func (t *SkillEffectTiming) UnmarshalJSON(data []byte) error {
	var name string
	if err := json.Unmarshal(data, &name); err == nil {
		switch strings.ToLower(name) {
		case "", "hit":
			*t = SkillEffectAtHit
		case "action_start":
			*t = SkillEffectAtActionStart
		default:
			return fmt.Errorf("unknown effect_timing %q", name)
		}
		return nil
	}
	var n int
	if err := json.Unmarshal(data, &n); err != nil {
		return err
	}
	*t = SkillEffectTiming(n)
	return nil
}

func parseComponentKind(name string, raw json.RawMessage) (EffectComponentKind, error) {
	if name == "" && len(raw) > 0 {
		var asString string
		if err := json.Unmarshal(raw, &asString); err == nil {
			name = asString
		} else {
			var asInt int
			if err := json.Unmarshal(raw, &asInt); err != nil {
				return 0, fmt.Errorf("kind: %w", err)
			}
			kind := EffectComponentKind(asInt)
			if componentKindName(kind) == "" && kind != 0 {
				return 0, fmt.Errorf("unknown kind %d", asInt)
			}
			return kind, nil
		}
	}
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "", "str":
		if name == "" {
			return 0, fmt.Errorf("missing kind")
		}
		return EffectComponentSTR, nil
	case "cylinder":
		return EffectComponentCylinder, nil
	case "2d":
		return EffectComponent2D, nil
	case "3d":
		return EffectComponent3D, nil
	case "spr":
		return EffectComponentSPR, nil
	case "func":
		return EffectComponentFUNC, nil
	case "quadhorn":
		return EffectComponentQuadHorn, nil
	default:
		return 0, fmt.Errorf("unknown kind %q", name)
	}
}

func componentKindName(kind EffectComponentKind) string {
	switch kind {
	case EffectComponentSTR:
		return "str"
	case EffectComponentCylinder:
		return "cylinder"
	case EffectComponent2D:
		return "2d"
	case EffectComponent3D:
		return "3d"
	case EffectComponentSPR:
		return "spr"
	case EffectComponentFUNC:
		return "func"
	case EffectComponentQuadHorn:
		return "quadhorn"
	default:
		return ""
	}
}

func parseSkillActionKind(name string) (SkillActionKind, error) {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "", "default":
		return SkillActionDefault, nil
	case "idle":
		return SkillActionIdle, nil
	case "skill":
		return SkillActionSkill, nil
	case "attack":
		return SkillActionAttack, nil
	case "attack1":
		return SkillActionAttack1, nil
	case "attack2":
		return SkillActionAttack2, nil
	case "attack3":
		return SkillActionAttack3, nil
	case "pickup":
		return SkillActionPickup, nil
	case "action":
		return SkillActionAction, nil
	case "readyfight":
		return SkillActionReadyfight, nil
	case "none":
		return SkillActionNone, nil
	case "attack_fixed_frame":
		return SkillActionAttackFixedFrame, nil
	case "dance":
		return SkillActionDance, nil
	default:
		return 0, fmt.Errorf("unknown skill action %q", name)
	}
}

func skillActionKindName(kind SkillActionKind) string {
	switch kind {
	case SkillActionIdle:
		return "idle"
	case SkillActionSkill:
		return "skill"
	case SkillActionAttack:
		return "attack"
	case SkillActionAttack1:
		return "attack1"
	case SkillActionAttack2:
		return "attack2"
	case SkillActionAttack3:
		return "attack3"
	case SkillActionPickup:
		return "pickup"
	case SkillActionAction:
		return "action"
	case SkillActionReadyfight:
		return "readyfight"
	case SkillActionNone:
		return "none"
	case SkillActionAttackFixedFrame:
		return "attack_fixed_frame"
	case SkillActionDance:
		return "dance"
	default:
		return "default"
	}
}
