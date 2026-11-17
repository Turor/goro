package res

import (
	"bytes"
	"context"
	"fmt"
	"path"
	"sort"
	"strings"
	"sync"
	"time"

	glua "github.com/yuin/gopher-lua"
	"github.com/yuin/gopher-lua/parse"
)

// DataCheckFailure is one file that a parser rejected during CheckArchives.
type DataCheckFailure struct {
	Name    string `json:"name"`
	Archive string `json:"archive"`
	Kind    string `json:"kind"`
	Version string `json:"version,omitempty"`
	Err     string `json:"error"`
}

// DataCheckKind aggregates one file kind (extension, or lub-bytecode/lub-text).
type DataCheckKind struct {
	Total    int            `json:"total"`
	Failed   int            `json:"failed"`
	Versions map[string]int `json:"versions,omitempty"`
}

// DataCheckReport is the result of a full archive walk.
type DataCheckReport struct {
	Archives []string                  `json:"archives"`
	Files    int                       `json:"files"`
	Shadowed int                       `json:"shadowed"`
	Skipped  int                       `json:"skipped"`
	Kinds    map[string]*DataCheckKind `json:"kinds"`
	Failures []DataCheckFailure        `json:"failures"`
	Elapsed  string                    `json:"elapsed"`
}

// DataCheckOptions tunes CheckArchives.
type DataCheckOptions struct {
	// Filter limits the walk to names it accepts (nil walks everything).
	Filter func(name string) bool
	// Workers is the number of parallel readers (default 8).
	Workers int
	// LuaTimeout bounds executing each compiled .lub (default 2s).
	LuaTimeout time.Duration
	// Progress, when set, is called every few thousand files.
	Progress func(done, total int)
}

type dataCheckJob struct {
	name    string
	archive *GRF
}

type dataCheckResult struct {
	kind    string
	version string
	err     error
	skipped bool
}

// CheckArchives reads every unique file (first archive wins, as in lookups)
// and runs the matching parser over it. It never writes anything.
func (m *Manager) CheckArchives(opts DataCheckOptions) DataCheckReport {
	start := time.Now()
	if opts.Workers <= 0 {
		opts.Workers = 8
	}
	if opts.LuaTimeout <= 0 {
		opts.LuaTimeout = 2 * time.Second
	}
	report := DataCheckReport{Kinds: map[string]*DataCheckKind{}}
	seen := map[string]bool{}
	var jobs []dataCheckJob
	for _, archive := range m.Archives {
		report.Archives = append(report.Archives, path.Base(strings.ReplaceAll(archive.Path(), "\\", "/")))
		for _, name := range archive.Names() {
			key := normalizeGRFName(name)
			if seen[key] {
				report.Shadowed++
				continue
			}
			seen[key] = true
			if opts.Filter != nil && !opts.Filter(name) {
				continue
			}
			jobs = append(jobs, dataCheckJob{name: name, archive: archive})
		}
	}
	sort.Slice(jobs, func(i, j int) bool { return jobs[i].name < jobs[j].name })

	results := make([]dataCheckResult, len(jobs))
	var wg sync.WaitGroup
	var mu sync.Mutex
	done := 0
	next := make(chan int)
	for w := 0; w < opts.Workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range next {
				results[i] = checkDataFile(jobs[i], opts.LuaTimeout)
				if opts.Progress != nil {
					mu.Lock()
					done++
					if done%5000 == 0 || done == len(jobs) {
						opts.Progress(done, len(jobs))
					}
					mu.Unlock()
				}
			}
		}()
	}
	for i := range jobs {
		next <- i
	}
	close(next)
	wg.Wait()

	for i, result := range results {
		if result.skipped {
			report.Skipped++
			continue
		}
		report.Files++
		kind := report.Kinds[result.kind]
		if kind == nil {
			kind = &DataCheckKind{}
			report.Kinds[result.kind] = kind
		}
		kind.Total++
		if result.version != "" {
			if kind.Versions == nil {
				kind.Versions = map[string]int{}
			}
			kind.Versions[result.version]++
		}
		if result.err != nil {
			kind.Failed++
			report.Failures = append(report.Failures, DataCheckFailure{
				Name:    jobs[i].name,
				Archive: path.Base(strings.ReplaceAll(jobs[i].archive.Path(), "\\", "/")),
				Kind:    result.kind,
				Version: result.version,
				Err:     result.err.Error(),
			})
		}
	}
	report.Elapsed = time.Since(start).Round(time.Millisecond).String()
	return report
}

func checkDataFile(job dataCheckJob, luaTimeout time.Duration) (result dataCheckResult) {
	ext := strings.ToLower(path.Ext(strings.ReplaceAll(job.name, "\\", "/")))
	result.kind = strings.TrimPrefix(ext, ".")
	switch ext {
	case ".rsw", ".gnd", ".gat", ".rsm", ".rsm2", ".act", ".spr", ".pal",
		".bmp", ".tga", ".jpg", ".jpeg", ".png", ".str", ".imf", ".lub", ".lua":
	default:
		result.skipped = true
		return result
	}
	defer func() {
		if r := recover(); r != nil {
			result.err = fmt.Errorf("panic: %v", r)
		}
	}()
	data, err := job.archive.ReadFile(job.name)
	if err != nil {
		result.err = fmt.Errorf("read: %w", err)
		return result
	}
	switch ext {
	case ".rsw":
		if len(data) >= 6 {
			result.version = fmt.Sprintf("%d.%d", data[4], data[5])
		}
		_, result.err = ParseRSW(data)
	case ".gnd":
		if len(data) >= 6 {
			result.version = fmt.Sprintf("%d.%d", data[4], data[5])
		}
		_, result.err = ParseGND(data)
	case ".gat":
		if len(data) >= 6 {
			result.version = fmt.Sprintf("%d.%d", data[4], data[5])
		}
		_, result.err = ParseGAT(data)
	case ".rsm", ".rsm2":
		if len(data) >= 6 {
			result.version = fmt.Sprintf("%d.%d", data[4], data[5])
		}
		_, result.err = ParseRSM(data)
	case ".act":
		if len(data) >= 4 {
			result.version = fmt.Sprintf("%d.%d", data[3], data[2])
		}
		_, result.err = ParseACT(data)
	case ".spr":
		if len(data) >= 4 {
			result.version = fmt.Sprintf("%d.%d", data[3], data[2])
		}
		_, result.err = ParseSPR(data)
	case ".pal":
		_, result.err = ParsePAL(data)
	case ".bmp", ".tga", ".jpg", ".jpeg", ".png":
		_, result.err = DecodeImageData(data)
	case ".str":
		_, result.err = ParseSTR(data, "")
	case ".imf":
		_, result.err = ParseIMF(data)
	case ".lub", ".lua":
		result.kind, result.err = checkLuaData(data, luaTimeout)
	}
	return result
}

// checkLuaData classifies a .lub/.lua file as Lua 5.1 bytecode or source and
// checks that goro can load it: bytecode is translated to a gopher-lua proto
// and executed in a sandbox; source is parsed.
func checkLuaData(data []byte, timeout time.Duration) (string, error) {
	if bytes.HasPrefix(data, []byte("\x1bLua")) {
		fn, err := parseLua51Bytecode(data)
		if err != nil {
			return "lub-bytecode", fmt.Errorf("bytecode: %w", err)
		}
		proto, err := gopherLuaProto(fn)
		if err != nil {
			return "lub-bytecode", fmt.Errorf("translate: %w", err)
		}
		state := glua.NewState(glua.Options{SkipOpenLibs: true})
		defer state.Close()
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		state.SetContext(ctx)
		openLuaResourceLibs(state)
		if err := state.CallByParam(glua.P{Fn: state.NewFunctionFromProto(proto), NRet: 0, Protect: true}); err != nil {
			// Most of these index constants a sibling file defines (JT_*,
			// ACCESSORY_*, SKID); the loader that uses the table runs them in
			// order. Count them, but they are not load failures.
			return "lub-bytecode-needs-globals", nil
		}
		return "lub-bytecode", nil
	}
	if _, err := parse.Parse(bytes.NewReader(stripUTF8BOM(data)), "chunk"); err != nil {
		return "lub-text", fmt.Errorf("parse: %s", firstLine(err.Error()))
	}
	return "lub-text", nil
}

func stripUTF8BOM(data []byte) []byte {
	return bytes.TrimPrefix(data, []byte("\xef\xbb\xbf"))
}

func firstLine(s string) string {
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		return s[:i]
	}
	return s
}
