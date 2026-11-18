package res

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/json"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/kivutar/goro/updater"
)

func TestPollJitterBounds(t *testing.T) {
	seen := map[time.Duration]struct{}{}
	for i := 0; i < 400; i++ {
		d := pollInterval(defaultJitter, pollMinDefault, pollMaxDefault)
		if d < pollMinDefault || d > pollMaxDefault {
			t.Fatalf("poll interval %s outside 60–120s", d)
		}
		seen[d] = struct{}{}
	}
	if len(seen) < 10 {
		t.Fatalf("jitter collapsed to %d values", len(seen))
	}
	if d := offlinePollInterval(60 * time.Second); d != 120*time.Second {
		t.Fatalf("offline double %s", d)
	}
	if d := offlinePollInterval(6 * time.Minute); d != pollOfflineMax {
		t.Fatalf("offline cap %s", d)
	}
}

func versionedManager(t *testing.T, srv *fakeAssetServer, opts RemoteOptions) *Manager {
	t.Helper()
	if opts.PollMin == 0 {
		opts.PollMin = time.Hour
		opts.PollMax = time.Hour
	}
	if opts.DeltaDebounce == 0 {
		opts.DeltaDebounce = 5 * time.Millisecond
	}
	opts.BaseURL = srv.url()
	if opts.CacheDir == "" {
		opts.CacheDir = t.TempDir()
	}
	m, err := NewManager(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if err := m.EnableRemote(opts); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(m.StopRemote)
	return m
}

func TestRemoteAdoptsFirstDataVersionWithoutDelta(t *testing.T) {
	srv := newFakeAssetServer(t)
	srv.dataVersion = 1000
	srv.index = []string{"data\\prontera.gat"}
	srv.put("data/prontera.gat", []byte("map"), "etag-1")
	var changes atomic.Int64
	srv.changes = func(since uint64) []byte {
		changes.Add(1)
		b, _ := json.Marshal(changesResponse{Version: 1000, Paths: []string{}})
		return b
	}
	m := versionedManager(t, srv, RemoteOptions{})
	if !m.WaitRemoteIndex(3 * time.Second) {
		t.Fatal("index")
	}
	if m.RemoteSeenVersion() != 1000 || m.RemoteCachedVersion() != 1000 {
		t.Fatalf("seen=%d cached=%d", m.RemoteSeenVersion(), m.RemoteCachedVersion())
	}
	time.Sleep(30 * time.Millisecond)
	if changes.Load() != 0 {
		t.Fatal("fresh cache must not call changes")
	}
	raw, err := os.ReadFile(filepath.Join(m.remote.cacheDir, "data-version.json"))
	if err != nil {
		t.Fatal(err)
	}
	var p persistedDataVersion
	if json.Unmarshal(raw, &p) != nil || p.Version != 1000 {
		t.Fatalf("persisted %+v", p)
	}
}

func TestRemoteAppliesChangesSinceAndEvictsPaths(t *testing.T) {
	srv := newFakeAssetServer(t)
	srv.dataVersion = 10
	srv.index = []string{"data\\a.txt", "data\\b.txt"}
	srv.put("data/a.txt", []byte("a1"), "ea")
	srv.put("data/b.txt", []byte("b1"), "eb")
	srv.changes = func(since uint64) []byte {
		if since != 10 {
			t.Errorf("since=%d", since)
		}
		b, _ := json.Marshal(changesResponse{Version: 11, Paths: []string{"data/a.txt"}})
		return b
	}
	m := versionedManager(t, srv, RemoteOptions{})
	m.WaitRemoteIndex(3 * time.Second)
	if _, err := m.ReadFile("data\\a.txt"); err != nil {
		t.Fatal(err)
	}
	hitsBefore := srv.hitCount("data/a.txt")

	srv.mu.Lock()
	srv.dataVersion = 11
	srv.files["data/a.txt"] = []byte("a2")
	srv.etags["data/a.txt"] = "ea2"
	srv.mu.Unlock()
	if _, err := m.ReadFile("data\\b.txt"); err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) && m.RemoteCachedVersion() != 11 {
		time.Sleep(10 * time.Millisecond)
	}
	if m.RemoteCachedVersion() != 11 {
		t.Fatalf("cached %d", m.RemoteCachedVersion())
	}
	data, err := m.ReadFile("data\\a.txt")
	if err != nil || string(data) != "a2" {
		t.Fatalf("after delta: %q %v", data, err)
	}
	if srv.hitCount("data/a.txt") <= hitsBefore {
		t.Fatal("stale path was not revalidated")
	}
}

func TestRemoteTooOldChangesTriggersFullRecheck(t *testing.T) {
	srv := newFakeAssetServer(t)
	srv.dataVersion = 5
	srv.index = []string{"data\\a.txt", "data\\b.txt"}
	srv.put("data/a.txt", []byte("a1"), "e1")
	srv.put("data/b.txt", []byte("b1"), "eb")
	srv.changes = func(since uint64) []byte {
		b, _ := json.Marshal(changesResponse{Version: 50, Full: true})
		return b
	}
	m := versionedManager(t, srv, RemoteOptions{})
	m.WaitRemoteIndex(3 * time.Second)
	if _, err := m.ReadFile("data\\a.txt"); err != nil {
		t.Fatal(err)
	}
	srv.mu.Lock()
	srv.dataVersion = 50
	srv.files["data/a.txt"] = []byte("a-new")
	srv.etags["data/a.txt"] = "e2"
	srv.mu.Unlock()
	if _, err := m.ReadFile("data\\b.txt"); err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) && m.RemoteCachedVersion() != 50 {
		time.Sleep(10 * time.Millisecond)
	}
	if m.RemoteCachedVersion() != 50 {
		t.Fatalf("cached %d after full", m.RemoteCachedVersion())
	}
	if m.remote.cacheGen.Load() == 0 {
		t.Fatal("full recheck must bump the cache generation")
	}
	data, err := m.ReadFile("data\\a.txt")
	if err != nil || string(data) != "a-new" {
		t.Fatalf("full recheck did not revalidate: %q %v", data, err)
	}
}

func TestRemotePollReadsVersionWhenIdle(t *testing.T) {
	srv := newFakeAssetServer(t)
	srv.dataVersion = 3
	srv.index = []string{"data\\a.txt"}
	srv.put("data/a.txt", []byte("a"), "")
	m := versionedManager(t, srv, RemoteOptions{PollMin: 15 * time.Millisecond, PollMax: 30 * time.Millisecond})
	m.WaitRemoteIndex(3 * time.Second)
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		srv.mu.Lock()
		n := srv.versionHits
		srv.mu.Unlock()
		if n >= 1 {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("idle poll never hit GET version")
}

func TestRemoteDigestMismatchIsDiscarded(t *testing.T) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	good := []byte("good-bytes")
	mft := updater.DataManifest{Version: 9, Files: []updater.DataFile{
		{Path: "data/a.txt", SHA256: updater.SHA256Hex(good), Size: int64(len(good))},
	}}
	payload, err := updater.CanonicalDataPayload(mft)
	if err != nil {
		t.Fatal(err)
	}
	env, err := updater.SignDataEnvelope(payload, priv)
	if err != nil {
		t.Fatal(err)
	}
	srv := newFakeAssetServer(t)
	srv.dataVersion = 9
	srv.index = []string{"data\\a.txt", "data\\b.txt"}
	srv.put("data/a.txt", []byte("tampered"), "e")
	srv.put("data/b.txt", good, "g")
	srv.manifest = []byte(env)
	mgr := versionedManager(t, srv, RemoteOptions{TrustedKeys: []ed25519.PublicKey{pub}})
	mgr.WaitRemoteIndex(3 * time.Second)
	if !mgr.WaitRemoteManifest(3 * time.Second) {
		t.Fatal("manifest not loaded")
	}
	if _, err := mgr.ReadFile("data\\a.txt"); err == nil {
		t.Fatal("digest mismatch was used")
	}
	dataPath, _ := mgr.remote.cachePaths("data/a.txt")
	if _, err := os.Stat(dataPath); !os.IsNotExist(err) {
		t.Fatal("mismatched body was cached")
	}
}

func TestRemoteSignedManifestVerifiesMatchingFile(t *testing.T) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	good := []byte("good-bytes")
	mft := updater.DataManifest{Version: 9, Files: []updater.DataFile{
		{Path: "data/a.txt", SHA256: updater.SHA256Hex(good), Size: int64(len(good))},
	}}
	payload, err := updater.CanonicalDataPayload(mft)
	if err != nil {
		t.Fatal(err)
	}
	env, err := updater.SignDataEnvelope(payload, priv)
	if err != nil {
		t.Fatal(err)
	}
	srv := newFakeAssetServer(t)
	srv.dataVersion = 9
	srv.index = []string{"data\\a.txt"}
	srv.put("data/a.txt", good, "e")
	srv.manifest = []byte(env)
	mgr := versionedManager(t, srv, RemoteOptions{TrustedKeys: []ed25519.PublicKey{pub}})
	mgr.WaitRemoteIndex(3 * time.Second)
	if !mgr.WaitRemoteManifest(3 * time.Second) {
		t.Fatal("manifest not loaded")
	}
	data, err := mgr.ReadFile("data\\a.txt")
	if err != nil || string(data) != "good-bytes" {
		t.Fatalf("got %q %v", data, err)
	}
}
