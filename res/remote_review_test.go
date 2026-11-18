package res

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestRemoteLoadsResnametableAfterEmptyDirStartup(t *testing.T) {
	srv := newFakeAssetServer(t)
	srv.index = []string{"data\\resnametable.txt", "data\\real.gat"}
	srv.put("data/resnametable.txt", []byte("alias.gat#real.gat#\n"), "")
	srv.put("data/real.gat", []byte("real"), "")
	// No local clientinfo: newManager falls back to the alias table before
	// streaming is enabled.
	m := streamingManager(t, t.TempDir(), t.TempDir(), srv.url(), nil)
	m.WaitRemoteIndex(5 * time.Second)
	data, err := m.ReadFile("data\\alias.gat")
	if err != nil || string(data) != "real" {
		t.Fatalf("alias.gat: got %q, %v; resnametable.txt must load from the server", data, err)
	}
	if srv.hitCount("data/resnametable.txt") != 1 {
		t.Fatal("resnametable.txt was not requested")
	}
}

func TestRemoteNeverStreamsClientInfo(t *testing.T) {
	srv := newFakeAssetServer(t)
	srv.index = []string{"data\\clientinfo.xml", "data\\sclientinfo.xml"}
	for _, p := range []string{"data/clientinfo.xml", "data/sclientinfo.xml", "System/clientinfo.xml", "System/sclientinfo.xml"} {
		srv.put(p, []byte("<clientinfo/>"), "")
	}
	m := streamingManager(t, t.TempDir(), t.TempDir(), srv.url(), nil)
	m.WaitRemoteIndex(5 * time.Second)
	for _, name := range clientInfoCandidates {
		if _, err := m.ReadFile(name); err == nil {
			t.Errorf("%s was streamed", name)
		}
		if m.HasFileExact(name) {
			t.Errorf("%s reported as available", name)
		}
	}
	for _, p := range []string{"data/clientinfo.xml", "data/sclientinfo.xml", "System/clientinfo.xml", "System/sclientinfo.xml"} {
		if n := srv.hitCount(p); n != 0 {
			t.Fatalf("%s requested %d times", p, n)
		}
	}
}

func TestRemotePerFileErrorDoesNotTakeServerOffline(t *testing.T) {
	srv := newFakeAssetServer(t)
	srv.put("data/broken.txt", nil, "")
	srv.broken["data/broken.txt"] = true
	srv.put("data/fine.txt", []byte("fine"), "")
	m := streamingManager(t, t.TempDir(), t.TempDir(), srv.url(), nil)
	if _, err := m.ReadFile("data\\broken.txt"); !errors.Is(err, errResourceNotFound) {
		t.Fatalf("broken file: %v", err)
	}
	data, err := m.ReadFile("data\\fine.txt")
	if err != nil || string(data) != "fine" {
		t.Fatalf("a short body must not trigger the offline back-off: %q %v", data, err)
	}
}

func TestRemoteSizeCaps(t *testing.T) {
	oldFile, oldIndex := remoteMaxFile, remoteMaxIndex
	t.Cleanup(func() { remoteMaxFile, remoteMaxIndex = oldFile, oldIndex })
	remoteMaxFile = 8

	srv := newFakeAssetServer(t)
	srv.put("data/small.txt", []byte("12345678"), "")
	srv.put("data/large.txt", []byte("123456789"), "")
	m := streamingManager(t, t.TempDir(), t.TempDir(), srv.url(), nil)
	if data, err := m.ReadFile("data\\small.txt"); err != nil || string(data) != "12345678" {
		t.Fatalf("file at the cap: %q %v", data, err)
	}
	if _, err := m.ReadFile("data\\large.txt"); !errors.Is(err, errResourceNotFound) {
		t.Fatalf("file over the cap: %v", err)
	}
	srv.put("data/after.txt", []byte("ok"), "")
	if data, err := m.ReadFile("data\\after.txt"); err != nil || string(data) != "ok" {
		t.Fatalf("an oversized file must not take the server offline: %q %v", data, err)
	}

	remoteMaxIndex = 16
	big := newFakeAssetServer(t)
	big.index = []string{"data\\a-rather-long-name.spr", "data\\another-long-name.spr"}
	m = streamingManager(t, t.TempDir(), t.TempDir(), big.url(), nil)
	if m.WaitRemoteIndex(5 * time.Second) {
		t.Fatal("an index over the cap must be rejected")
	}
}

func TestRemoteIndexETagSkipsRedownload(t *testing.T) {
	srv := newFakeAssetServer(t)
	srv.index = []string{"data\\a.spr"}
	srv.indexETag = `"idx-1"`
	cache := t.TempDir()
	first := streamingManager(t, t.TempDir(), cache, srv.url(), nil)
	if !first.WaitRemoteIndex(5 * time.Second) {
		t.Fatal("index did not load")
	}
	indexPath, _ := first.remote.indexPaths()
	before, err := os.Stat(indexPath)
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(20 * time.Millisecond)
	second := streamingManager(t, t.TempDir(), cache, srv.url(), nil)
	if !second.WaitRemoteIndex(5*time.Second) || !second.HasFileExact("data\\a.spr") {
		t.Fatal("cached index not used")
	}
	after, err := os.Stat(indexPath)
	if err != nil {
		t.Fatal(err)
	}
	srv.mu.Lock()
	full := srv.indexFull
	srv.mu.Unlock()
	if full != 1 || !after.ModTime().Equal(before.ModTime()) {
		t.Fatalf("index downloaded %d times, rewritten=%v; an unchanged index must be revalidated, not refetched", full, !after.ModTime().Equal(before.ModTime()))
	}
}

func TestRemoteRevalidatesArchiveFilesWhenIndexChanges(t *testing.T) {
	srv := newFakeAssetServer(t)
	srv.index = []string{"data\\a.spr"}
	srv.put("data/a.spr", []byte("v1"), `"1"`)
	cache := t.TempDir()
	first := streamingManager(t, t.TempDir(), cache, srv.url(), nil)
	first.WaitRemoteIndex(5 * time.Second)
	if data, _ := first.ReadFile("data\\a.spr"); string(data) != "v1" {
		t.Fatalf("first: %q", data)
	}

	// Same index: the cached archive file is used without a request.
	same := streamingManager(t, t.TempDir(), cache, srv.url(), nil)
	same.WaitRemoteIndex(5 * time.Second)
	if data, _ := same.ReadFile("data\\a.spr"); string(data) != "v1" || srv.hitCount("data/a.spr") != 1 {
		t.Fatalf("unchanged index: %q after %d requests", data, srv.hitCount("data/a.spr"))
	}

	// The server's archives changed: revalidate and take the new file.
	srv.mu.Lock()
	srv.index = []string{"data\\a.spr", "data\\b.spr"}
	srv.mu.Unlock()
	srv.put("data/a.spr", []byte("v2"), `"2"`)
	changed := streamingManager(t, t.TempDir(), cache, srv.url(), nil)
	changed.WaitRemoteIndex(5 * time.Second)
	if data, _ := changed.ReadFile("data\\a.spr"); string(data) != "v2" {
		t.Fatalf("changed index: %q", data)
	}
}

func TestRemoteCacheIsKeyedByHost(t *testing.T) {
	one, two := newFakeAssetServer(t), newFakeAssetServer(t)
	one.put("data/x.spr", []byte("one"), "")
	two.put("data/x.spr", []byte("two"), "")
	cache := t.TempDir()
	a := streamingManager(t, t.TempDir(), cache, one.url(), nil)
	b := streamingManager(t, t.TempDir(), cache, two.url(), nil)
	if data, _ := a.ReadFile("data\\x.spr"); string(data) != "one" {
		t.Fatalf("host one: %q", data)
	}
	if data, _ := b.ReadFile("data\\x.spr"); string(data) != "two" {
		t.Fatalf("host two read %q from a shared cache_dir", data)
	}
}

func TestRemoteBackgroundWorkIsBounded(t *testing.T) {
	srv := newFakeAssetServer(t)
	cache := t.TempDir()
	var names []string
	for i := range 3 * remoteBackground {
		p := "data/t" + string(rune('a'+i)) + ".txt"
		srv.put(p, []byte("v1"), `"1"`)
		names = append(names, strings.ReplaceAll(p, "/", "\\"))
	}
	warm := streamingManager(t, t.TempDir(), cache, srv.url(), nil)
	for _, name := range names {
		if _, err := warm.ReadFile(name); err != nil {
			t.Fatal(err)
		}
	}
	srv.mu.Lock()
	for _, name := range names {
		p := strings.ReplaceAll(name, "\\", "/")
		srv.files[p] = []byte("v2")
		srv.etags[p] = `"2"`
		srv.delay[p] = 30 * time.Millisecond
	}
	srv.maxActive = 0
	srv.mu.Unlock()

	m := streamingManager(t, t.TempDir(), cache, srv.url(), nil)
	var wg sync.WaitGroup
	wg.Add(len(names))
	m.remote.refreshed = func(string) { wg.Done() }
	for _, name := range names {
		if _, err := m.ReadFile(name); err != nil {
			t.Fatal(err)
		}
	}
	wg.Wait()
	srv.mu.Lock()
	defer srv.mu.Unlock()
	if srv.maxActive > remoteBackground {
		t.Fatalf("%d concurrent refreshes, cap %d", srv.maxActive, remoteBackground)
	}
}

func TestReadFileCandidatesWithinDoesNotStallFrames(t *testing.T) {
	srv := newFakeAssetServer(t)
	srv.index = []string{"data\\slow.bmp"}
	srv.put("data/slow.bmp", []byte("slow"), "")
	srv.delay["data/slow.bmp"] = 200 * time.Millisecond
	m := streamingManager(t, t.TempDir(), t.TempDir(), srv.url(), nil)
	m.WaitRemoteIndex(5 * time.Second)

	names := []string{"data\\slow.bmp"}
	start := time.Now()
	if _, _, err := m.ReadFileCandidatesWithin(names, 10*time.Millisecond); !errors.Is(err, ErrRemotePending) || !errors.Is(err, errResourceNotFound) {
		t.Fatalf("first frame: %v, want ErrRemotePending", err)
	}
	if _, _, err := m.ReadFileCandidatesWithin(names, 10*time.Millisecond); !errors.Is(err, ErrRemotePending) {
		t.Fatalf("second frame: %v", err)
	}
	if elapsed := time.Since(start); elapsed > 100*time.Millisecond {
		t.Fatalf("frames waited %s", elapsed)
	}
	deadline := time.Now().Add(5 * time.Second)
	for {
		data, _, err := m.ReadFileCandidatesWithin(names, 10*time.Millisecond)
		if err == nil {
			if string(data) != "slow" {
				t.Fatalf("got %q", data)
			}
			break
		}
		if !errors.Is(err, ErrRemotePending) || time.Now().After(deadline) {
			t.Fatalf("download never completed: %v", err)
		}
		time.Sleep(20 * time.Millisecond)
	}
	if n := srv.hitCount("data/slow.bmp"); n != 1 {
		t.Fatalf("%d requests, want 1", n)
	}
}

func TestPrefetchRemoteDownloadsInParallel(t *testing.T) {
	srv := newFakeAssetServer(t)
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "data"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "data", "local.spr"), []byte("local"), 0o644); err != nil {
		t.Fatal(err)
	}
	var lists [][]string
	for i := range remoteBackground {
		p := "data/m" + string(rune('a'+i)) + ".spr"
		srv.index = append(srv.index, strings.ReplaceAll(p, "/", "\\"))
		srv.put(p, []byte(p), "")
		srv.delay[p] = 150 * time.Millisecond
		lists = append(lists, []string{strings.ReplaceAll(p, "/", "\\")})
	}
	srv.index = append(srv.index, "data\\local.spr")
	lists = append(lists, []string{"data\\local.spr"})
	m := streamingManager(t, root, t.TempDir(), srv.url(), nil)
	m.WaitRemoteIndex(5 * time.Second)

	start := time.Now()
	batch := m.PrefetchRemote(lists)
	for !batch.Done() {
		if time.Since(start) > 5*time.Second {
			t.Fatal("prefetch never finished")
		}
		time.Sleep(5 * time.Millisecond)
	}
	if elapsed := time.Since(start); elapsed > 450*time.Millisecond {
		t.Fatalf("prefetch took %s; it must run in parallel", elapsed)
	}
	before := srv.totalHits()
	for _, list := range lists {
		if _, _, err := m.ReadFileCandidates(list); err != nil {
			t.Fatal(err)
		}
	}
	if srv.totalHits() != before || srv.hitCount("data/local.spr") != 0 {
		t.Fatal("prefetched files were requested again, or a local file was fetched")
	}
	var nilBatch *RemoteBatch
	if !nilBatch.Done() {
		t.Fatal("nil batch must be done")
	}
}
