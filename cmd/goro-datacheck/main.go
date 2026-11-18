// Command goro-datacheck walks a Ragnarok data directory (DATA.INI order,
// first archive wins) and reports every file goro's parsers reject. It only
// reads the archives.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/kivutar/goro/res"
)

func main() {
	dataDir := flag.String("data-dir", "", "Ragnarok data directory with DATA.INI")
	jsonOut := flag.String("json", "", "write the full report as JSON to this path")
	only := flag.String("only", "", "comma-separated extensions to check (e.g. rsw,gnd,lub)")
	workers := flag.Int("workers", 8, "parallel readers")
	maxExamples := flag.Int("examples", 5, "failure examples printed per kind and error")
	dataURL := flag.String("data-url", "", "stream files missing from --data-dir from this asset server (e.g. https://ro.turoran.com/client/)")
	dataCache := flag.String("data-cache", "", "cache directory for streamed files (default: a temporary directory)")
	streamMap := flag.String("stream-map", "", "with --data-url: load this map (rsw, gnd, gat and its ground textures) and exit")
	flag.Parse()
	if *streamMap != "" {
		os.Exit(runStreamMap(*dataDir, *dataURL, *dataCache, *streamMap))
	}
	if *dataDir == "" {
		fmt.Fprintln(os.Stderr, "--data-dir is required")
		os.Exit(2)
	}
	manager, err := res.NewManager(*dataDir)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	opts := res.DataCheckOptions{Workers: *workers, Progress: func(done, total int) {
		fmt.Fprintf(os.Stderr, "\r%d/%d", done, total)
	}}
	if *only != "" {
		exts := map[string]bool{}
		for _, ext := range strings.Split(*only, ",") {
			exts["."+strings.TrimPrefix(strings.ToLower(strings.TrimSpace(ext)), ".")] = true
		}
		opts.Filter = func(name string) bool {
			lower := strings.ToLower(name)
			if i := strings.LastIndexByte(lower, '.'); i >= 0 {
				return exts[lower[i:]]
			}
			return false
		}
	}
	report := manager.CheckArchives(opts)
	fmt.Fprintln(os.Stderr)
	fmt.Printf("archives: %s\n", strings.Join(report.Archives, ", "))
	fmt.Printf("checked %d files (%d skipped kinds, %d shadowed by an earlier archive) in %s\n",
		report.Files, report.Skipped, report.Shadowed, report.Elapsed)
	kinds := make([]string, 0, len(report.Kinds))
	for kind := range report.Kinds {
		kinds = append(kinds, kind)
	}
	sort.Strings(kinds)
	fmt.Printf("%-14s %8s %7s  versions\n", "kind", "total", "failed")
	for _, kind := range kinds {
		k := report.Kinds[kind]
		fmt.Printf("%-14s %8d %7d  %s\n", kind, k.Total, k.Failed, formatVersions(k.Versions))
	}
	type groupKey struct{ kind, err string }
	groups := map[groupKey][]res.DataCheckFailure{}
	for _, f := range report.Failures {
		key := groupKey{f.Kind, normalizeErr(f.Err)}
		groups[key] = append(groups[key], f)
	}
	keys := make([]groupKey, 0, len(groups))
	for key := range groups {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool { return len(groups[keys[i]]) > len(groups[keys[j]]) })
	for _, key := range keys {
		failures := groups[key]
		fmt.Printf("\n[%s] %d x %s\n", key.kind, len(failures), key.err)
		for i, f := range failures {
			if i >= *maxExamples {
				break
			}
			fmt.Printf("  %s (%s v%s): %s\n", f.Name, f.Archive, f.Version, f.Err)
		}
	}
	if *jsonOut != "" {
		data, err := json.MarshalIndent(report, "", "  ")
		if err == nil {
			err = os.WriteFile(*jsonOut, data, 0o644)
		}
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
}

func formatVersions(versions map[string]int) string {
	keys := make([]string, 0, len(versions))
	for v := range versions {
		keys = append(keys, v)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, v := range keys {
		parts = append(parts, fmt.Sprintf("%s:%d", v, versions[v]))
	}
	return strings.Join(parts, " ")
}

// normalizeErr strips digits so similar errors group together.
func normalizeErr(err string) string {
	var b strings.Builder
	digit := false
	for _, r := range err {
		if r >= '0' && r <= '9' {
			if !digit {
				b.WriteByte('N')
			}
			digit = true
			continue
		}
		digit = false
		b.WriteRune(r)
	}
	s := b.String()
	if len(s) > 120 {
		s = s[:120]
	}
	return s
}
