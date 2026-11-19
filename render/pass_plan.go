package render

// Pass is one recorded render pass. The GPU backend walks this list.
type Pass struct {
	Name                  string
	Fullscreen            bool
	SubmitsOpaqueGeometry bool
	// DepthInput is "opaque-depth" when the pass reads the opaque world depth.
	DepthInput string
	// After names the pass this one follows, such as "world-resolve".
	After string
}

// PassOptions selects optional passes.
type PassOptions struct {
	SSAO     bool
	FXAA     bool
	Enhanced bool
}

// PlanFramePasses records the passes for one frame.
// Ambient occlusion reads the opaque pass depth and does not submit that
// geometry again. FXAA is one fullscreen pass after the world color resolve.
func PlanFramePasses(opt PassOptions) []Pass {
	passes := make([]Pass, 0, 6)
	if opt.Enhanced {
		passes = append(passes, Pass{Name: "directional-shadow"})
	}
	passes = append(passes, Pass{Name: "opaque-world", SubmitsOpaqueGeometry: true})
	if opt.SSAO {
		passes = append(passes, Pass{Name: "ssao-estimate", DepthInput: "opaque-depth"})
		passes = append(passes, Pass{Name: "ssao-composite", Fullscreen: true})
	}
	passes = append(passes, Pass{Name: "world-resolve"})
	if opt.FXAA {
		passes = append(passes, Pass{Name: "fxaa", Fullscreen: true, After: "world-resolve"})
	}
	passes = append(passes, Pass{Name: "screen"})
	return passes
}

// OpaqueGeometrySubmissions counts passes that draw opaque world geometry.
func OpaqueGeometrySubmissions(passes []Pass) int {
	count := 0
	for _, pass := range passes {
		if pass.SubmitsOpaqueGeometry {
			count++
		}
	}
	return count
}

// PassByName returns the first pass with name.
func PassByName(passes []Pass, name string) (Pass, bool) {
	for _, pass := range passes {
		if pass.Name == name {
			return pass, true
		}
	}
	return Pass{}, false
}

// FXAAFollowsResolve reports whether exactly one FXAA pass sits immediately
// after the world color resolve.
func FXAAFollowsResolve(passes []Pass) bool {
	resolve := -1
	fxaa := 0
	fxaaAt := -1
	for i, pass := range passes {
		if pass.Name == "world-resolve" && resolve < 0 {
			resolve = i
		}
		if pass.Name == "fxaa" {
			fxaa++
			fxaaAt = i
		}
	}
	return fxaa == 1 && resolve >= 0 && fxaaAt == resolve+1
}
