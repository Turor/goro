package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/kivutar/goro/res"
)

// runStreamMap loads one map the way the game does, streaming whatever the
// data directory lacks, and reports where each file came from.
func runStreamMap(dataDir, dataURL, cacheDir, mapName string) int {
	if dataURL == "" {
		fmt.Fprintln(os.Stderr, "--stream-map needs --data-url")
		return 2
	}
	if dataDir == "" {
		dir, err := os.MkdirTemp("", "goro-stream-data-")
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
		defer os.RemoveAll(dir)
		dataDir = dir
	}
	if cacheDir == "" {
		dir, err := os.MkdirTemp("", "goro-stream-cache-")
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
		defer os.RemoveAll(dir)
		cacheDir = dir
	}
	manager, err := res.NewManager(dataDir)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	start := time.Now()
	if err := manager.EnableRemote(res.RemoteOptions{BaseURL: dataURL, CacheDir: cacheDir, Aliases: res.TuroranRemoteAliases}); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	fmt.Printf("local archives: %d, remote enabled in %s\n", len(manager.Archives), time.Since(start).Round(time.Millisecond))
	indexed := manager.WaitRemoteIndex(time.Minute)
	fmt.Printf("archive index loaded: %v after %s\n", indexed, time.Since(start).Round(time.Millisecond))

	base := strings.TrimSuffix(strings.TrimSuffix(mapName, ".gat"), ".rsw")
	read := func(ext string) ([]byte, bool) {
		data, source, err := manager.ReadFileCandidates([]string{"data\\" + base + "." + ext, "data/" + base + "." + ext, base + "." + ext})
		if err != nil {
			fmt.Printf("%s.%s: MISSING: %v\n", base, ext, err)
			return nil, false
		}
		fmt.Printf("%s.%s: %d bytes via %s\n", base, ext, len(data), source)
		return data, true
	}
	ok := true
	if data, found := read("rsw"); found {
		rsw, err := res.ParseRSW(data)
		if err != nil {
			fmt.Printf("rsw parse: %v\n", err)
			ok = false
		} else {
			fmt.Printf("rsw: %d models\n", len(rsw.Models))
		}
	} else {
		ok = false
	}
	if data, found := read("gat"); found {
		if _, err := res.ParseGAT(data); err != nil {
			fmt.Printf("gat parse: %v\n", err)
			ok = false
		}
	} else {
		ok = false
	}
	if data, found := read("gnd"); found {
		gnd, err := res.ParseGND(data)
		if err != nil {
			fmt.Printf("gnd parse: %v\n", err)
			ok = false
		} else {
			loaded := 0
			for _, texture := range gnd.Textures {
				if _, _, err := res.LoadImage(manager, res.GroundTextureCandidates(texture)); err == nil {
					loaded++
				} else {
					fmt.Printf("ground texture missing: %s\n", texture)
				}
			}
			fmt.Printf("gnd: %d/%d ground textures loaded\n", loaded, len(gnd.Textures))
			if loaded != len(gnd.Textures) {
				ok = false
			}
		}
	} else {
		ok = false
	}
	fmt.Printf("remote requests: %d, total %s\n", manager.RemoteRequests(), time.Since(start).Round(time.Millisecond))
	if !ok {
		return 1
	}
	return 0
}
