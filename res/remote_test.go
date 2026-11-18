package res

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"golang.org/x/text/encoding/korean"
)

// fakeAssetServer mimics remoteclient2: GET /list-files returns the archive
// names (backslashes), and every other path serves a file or 404.
type fakeAssetServer struct {
	t      *testing.T
	mu     sync.Mutex
	index  []string
	files  map[string][]byte // decoded URL path without the leading slash
	etags  map[string]string
	hits   map[string]int
	raw    []string // raw request paths, to check percent-encoding
	gate   chan struct{}
	server *httptest.Server
}

func newFakeAssetServer(t *testing.T) *fakeAssetServer {
	t.Helper()
	f := &fakeAssetServer{t: t, files: map[string][]byte{}, etags: map[string]string{}, hits: map[string]int{}}
	f.server = httptest.NewServer(http.HandlerFunc(f.serve))
	t.Cleanup(f.server.Close)
	return f
}

func (f *fakeAssetServer) serve(w http.ResponseWriter, r *http.Request) {
	p := strings.TrimPrefix(r.URL.Path, "/client/")
	if p == "list-files" {
		f.mu.Lock()
		gate := f.gate
		f.mu.Unlock()
		if gate != nil {
			<-gate
		}
		f.mu.Lock()
		body, _ := json.Marshal(f.index)
		f.mu.Unlock()
		_, _ = w.Write(body)
		return
	}
	f.mu.Lock()
	f.hits[p]++
	f.raw = append(f.raw, r.URL.EscapedPath())
	data, ok := f.files[p]
	etag := f.etags[p]
	f.mu.Unlock()
	if !ok {
		http.NotFound(w, r)
		return
	}
	if etag != "" {
		w.Header().Set("ETag", etag)
		if r.Header.Get("If-None-Match") == etag {
			w.WriteHeader(http.StatusNotModified)
			return
		}
	}
	_, _ = w.Write(data)
}

func (f *fakeAssetServer) put(p string, data []byte, etag string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.files[p] = data
	f.etags[p] = etag
}

func (f *fakeAssetServer) hitCount(p string) int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.hits[p]
}

func (f *fakeAssetServer) totalHits() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	n := 0
	for _, c := range f.hits {
		n += c
	}
	return n
}

func (f *fakeAssetServer) url() string { return f.server.URL + "/client/" }

func streamingManager(t *testing.T, root, cache, url string, aliases map[string]string) *Manager {
	t.Helper()
	m, err := NewManager(root)
	if err != nil {
		t.Fatal(err)
	}
	if err := m.EnableRemote(RemoteOptions{BaseURL: url, CacheDir: cache, Aliases: aliases}); err != nil {
		t.Fatal(err)
	}
	return m
}

func TestRemoteStreamsMapMissingLocallyAndReusesCache(t *testing.T) {
	srv := newFakeAssetServer(t)
	srv.index = []string{"data\\prontera.rsw", "data\\prontera.gnd", "data\\prontera.gat"}
	for _, ext := range []string{"rsw", "gnd", "gat"} {
		srv.put("data/prontera."+ext, []byte("prontera-"+ext), "")
	}
	root, cache := t.TempDir(), t.TempDir()
	m := streamingManager(t, root, cache, srv.url(), nil)
	if !m.WaitRemoteIndex(5 * time.Second) {
		t.Fatal("index did not load")
	}
	for _, ext := range []string{"rsw", "gnd", "gat"} {
		// The map loaders try data\ first, like the GRF names.
		data, source, err := m.ReadFileCandidates([]string{"data\\prontera." + ext, "data/prontera." + ext, "prontera." + ext})
		if err != nil || string(data) != "prontera-"+ext {
			t.Fatalf("%s: got %q from %q, %v", ext, data, source, err)
		}
	}
	if !m.HasFileExact("data\\prontera.gat") || m.HasFileExact("data\\nowhere.gat") {
		t.Fatal("HasFileExact must follow the server index")
	}

	// A new run with the server gone still loads the map from the cache.
	srv.server.Close()
	offline := streamingManager(t, root, cache, srv.url(), nil)
	data, err := offline.ReadFile("data\\prontera.gnd")
	if err != nil || string(data) != "prontera-gnd" {
		t.Fatalf("cached map: got %q, %v", data, err)
	}
}

func TestRemoteLocalFilesStayExplicitOverrides(t *testing.T) {
	srv := newFakeAssetServer(t)
	srv.index = []string{"data\\clientinfo.xml", "data\\local.txt"}
	srv.put("data/local.txt", []byte("server"), "")
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "data"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "data", "local.txt"), []byte("loose"), 0o644); err != nil {
		t.Fatal(err)
	}
	m := streamingManager(t, root, t.TempDir(), srv.url(), nil)
	data, err := m.ReadFile("data\\local.txt")
	if err != nil || string(data) != "loose" {
		t.Fatalf("got %q, %v; the loose file must win", data, err)
	}
	if n := srv.totalHits(); n != 0 {
		t.Fatalf("%d requests for a local file", n)
	}
}

func TestRemoteEncodesKoreanPathsAsUTF8(t *testing.T) {
	srv := newFakeAssetServer(t)
	const utf8Name = "data/texture/유저인터페이스/item/sm_bash.bmp"
	srv.index = []string{strings.ReplaceAll(utf8Name, "/", "\\")}
	srv.put(utf8Name, []byte("icon"), "")
	m := streamingManager(t, t.TempDir(), t.TempDir(), srv.url(), nil)
	m.WaitRemoteIndex(5 * time.Second)

	eucKR, err := korean.EUCKR.NewEncoder().String(strings.ReplaceAll(utf8Name, "/", "\\"))
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{eucKR, strings.ReplaceAll(utf8Name, "/", "\\")} {
		data, err := m.ReadFile(name)
		if err != nil || string(data) != "icon" {
			t.Fatalf("%q: got %q, %v", name, data, err)
		}
	}
	srv.mu.Lock()
	defer srv.mu.Unlock()
	if len(srv.raw) != 1 {
		t.Fatalf("requests %v, want one (shared cache for both spellings)", srv.raw)
	}
	want := "/client/data/texture/%EC%9C%A0%EC%A0%80%EC%9D%B8%ED%84%B0%ED%8E%98%EC%9D%B4%EC%8A%A4/item/sm_bash.bmp"
	if srv.raw[0] != want {
		t.Fatalf("request path %s, want %s", srv.raw[0], want)
	}
}

func TestRemoteMissingFilesFailGracefullyAndAreAskedOnce(t *testing.T) {
	srv := newFakeAssetServer(t)
	srv.index = []string{"data\\known.spr"}
	m := streamingManager(t, t.TempDir(), t.TempDir(), srv.url(), nil)
	m.WaitRemoteIndex(5 * time.Second)

	for range 3 {
		if _, err := m.ReadFile("data\\missing.txt"); !errors.Is(err, errResourceNotFound) {
			t.Fatalf("missing text: %v, want errResourceNotFound", err)
		}
	}
	if n := srv.hitCount("data/missing.txt"); n != 1 {
		t.Fatalf("missing text requested %d times, want 1", n)
	}
	// Archive files the index does not list are never requested.
	if _, err := m.ReadFile("data\\texture\\nothere.bmp"); !errors.Is(err, errResourceNotFound) {
		t.Fatalf("unlisted binary: %v", err)
	}
	// Names outside data/ and System/ are not served.
	if _, err := m.ReadFile("texture\\item\\x.bmp"); !errors.Is(err, errResourceNotFound) {
		t.Fatalf("bare name: %v", err)
	}
	if n := srv.totalHits(); n != 1 {
		t.Fatalf("%d requests, want only the one text probe", n)
	}
	// A missing candidate falls through to the next one.
	srv.put("data/known.spr", []byte("spr"), "")
	data, source, err := m.ReadFileCandidates([]string{"data\\missing.txt", "data\\known.spr"})
	if err != nil || string(data) != "spr" || source != "data\\known.spr" {
		t.Fatalf("fallthrough: %q %q %v", data, source, err)
	}
}

func TestRemoteUnreachableServerDoesNotAbortLookups(t *testing.T) {
	srv := newFakeAssetServer(t)
	srv.server.Close()
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "data"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "data", "b.txt"), []byte("local"), 0o644); err != nil {
		t.Fatal(err)
	}
	m := streamingManager(t, root, t.TempDir(), srv.url(), nil)
	data, _, err := m.ReadFileCandidates([]string{"data\\a.txt", "data\\b.txt"})
	if err != nil || string(data) != "local" {
		t.Fatalf("got %q, %v", data, err)
	}
	start := time.Now()
	for range 20 {
		_, _ = m.ReadFile("data\\c.txt")
	}
	if time.Since(start) > 2*time.Second {
		t.Fatal("an offline server must not be retried on every lookup")
	}
}

func TestRemoteTuroranAliasesPreferVersionedTables(t *testing.T) {
	srv := newFakeAssetServer(t)
	srv.put("System/itemInfo.turoran.renewal11.lua", []byte("turoran"), "")
	srv.put("System/itemInfo.lua", []byte("stock"), "")
	srv.put("data/luafiles514/lua files/skillinfoz/skillid.lub", []byte("stock skillid"), "")
	m := streamingManager(t, t.TempDir(), t.TempDir(), srv.url(), TuroranRemoteAliases)

	_, data, ok := m.ReadFirst(itemInfoLuaCandidates)
	if !ok || string(data) != "turoran" {
		t.Fatalf("itemInfo: got %q, want the Turoran renewal table", data)
	}
	// Without the versioned file the canonical one is the fallback.
	_, data, ok = m.ReadFirst(skillIDLuaCandidates)
	if !ok || string(data) != "stock skillid" {
		t.Fatalf("skillid fallback: got %q", data)
	}
	if n := srv.hitCount("data/luafiles514/lua files/skillinfoz/skillid.turoran4.lub"); n != 1 {
		t.Fatalf("alias asked %d times, want 1", n)
	}
	if err := checkAssembleScriptPicks(TuroranRemoteAliases); err != nil {
		t.Fatal(err)
	}
}

// The streaming aliases and scripts/data/assemble-turoran-data.sh must pick
// the same Turoran tables.
func checkAssembleScriptPicks(aliases map[string]string) error {
	script, err := os.ReadFile(filepath.Join("..", "scripts", "data", "assemble-turoran-data.sh"))
	if err != nil {
		return err
	}
	picks := 0
	for _, line := range strings.Split(string(script), "\n") {
		if !strings.HasPrefix(line, "pick ") {
			continue
		}
		fields := strings.Split(line, "\" \"")
		if len(fields) != 2 {
			continue
		}
		src := strings.Trim(fields[0], "pick \"")
		dst := strings.Trim(fields[1], "\"")
		src = strings.NewReplacer("$overlay/", "", "$system/", "System/").Replace(src)
		dst = strings.NewReplacer("$lua/", "data/luafiles514/lua files/", "$out/", "").Replace(dst)
		if got := aliases[strings.ToLower(dst)]; got != src {
			return errors.New("alias for " + dst + " is " + got + ", script picks " + src)
		}
		picks++
	}
	if picks != len(aliases) {
		return errors.New("script and aliases list different tables")
	}
	return nil
}

func TestRemoteStartupDoesNotWaitForIndex(t *testing.T) {
	srv := newFakeAssetServer(t)
	gate := make(chan struct{})
	srv.gate = gate
	defer close(gate)
	srv.put("data/early.txt", []byte("early"), "")
	start := time.Now()
	m := streamingManager(t, t.TempDir(), t.TempDir(), srv.url(), nil)
	data, err := m.ReadFile("data\\early.txt")
	if err != nil || string(data) != "early" {
		t.Fatalf("got %q, %v", data, err)
	}
	if time.Since(start) > 2*time.Second {
		t.Fatal("startup waited for the archive index")
	}
	if m.WaitRemoteIndex(10 * time.Millisecond) {
		t.Fatal("index should still be loading")
	}
}

func TestRemoteTextTablesRefreshInBackground(t *testing.T) {
	srv := newFakeAssetServer(t)
	srv.put("data/table.txt", []byte("v1"), `"1"`)
	cache := t.TempDir()
	first := streamingManager(t, t.TempDir(), cache, srv.url(), nil)
	if data, err := first.ReadFile("data\\table.txt"); err != nil || string(data) != "v1" {
		t.Fatalf("first: %q %v", data, err)
	}

	srv.put("data/table.txt", []byte("v2"), `"2"`)
	second := streamingManager(t, t.TempDir(), cache, srv.url(), nil)
	refreshed := make(chan string, 1)
	second.remote.refreshed = func(p string) { refreshed <- p }
	// The cached copy is served at once, the new one fetched behind it.
	if data, err := second.ReadFile("data\\table.txt"); err != nil || string(data) != "v1" {
		t.Fatalf("second: %q %v", data, err)
	}
	select {
	case <-refreshed:
	case <-time.After(5 * time.Second):
		t.Fatal("no background refresh")
	}
	third := streamingManager(t, t.TempDir(), cache, srv.url(), nil)
	third.remote.refreshed = func(string) {}
	if data, err := third.ReadFile("data\\table.txt"); err != nil || string(data) != "v2" {
		t.Fatalf("third: %q %v", data, err)
	}
}

func TestRemotePathOnlyServesDataAndSystem(t *testing.T) {
	cases := map[string]string{
		"data\\Sprite\\a.spr":   "data/Sprite/a.spr",
		"DATA/x.txt":            "data/x.txt",
		"system\\itemInfo.lua":  "System/itemInfo.lua",
		"data\\..\\etc\\passwd": "",
		"..\\data\\x":           "",
		"texture\\x.bmp":        "",
		"data":                  "",
	}
	for in, want := range cases {
		if got := remotePath(in); got != want {
			t.Errorf("remotePath(%q) = %q, want %q", in, got, want)
		}
	}
}
