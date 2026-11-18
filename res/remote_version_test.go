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

func signedDataManifest(t *testing.T, priv ed25519.PrivateKey, version uint64, files map[string][]byte) []byte {
	t.Helper()
	list := make([]updater.DataFile, 0, len(files))
	for p, b := range files {
		list = append(list, updater.DataFile{Path: p, SHA256: updater.SHA256Hex(b), Size: int64(len(b))})
	}
	payload, err := updater.CanonicalDataPayload(updater.DataManifest{Version: version, Files: list})
	if err != nil {
		t.Fatal(err)
	}
	env, err := updater.SignDataEnvelope(payload, priv)
	if err != nil {
		t.Fatal(err)
	}
	return []byte(env)
}

func testKey(t *testing.T) (ed25519.PublicKey, ed25519.PrivateKey) {
	t.Helper()
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	return pub, priv
}

func TestRemoteAppliesChangesSinceAndEvictsPaths(t *testing.T) {
	pub, priv := testKey(t)
	srv := newFakeAssetServer(t)
	srv.dataVersion = 10
	srv.index = []string{"data\\a.txt", "data\\b.txt"}
	srv.put("data/a.txt", []byte("a1"), "ea")
	srv.put("data/b.txt", []byte("b1"), "eb")
	srv.manifest = signedDataManifest(t, priv, 10, map[string][]byte{
		"data/a.txt": []byte("a1"),
		"data/b.txt": []byte("b1"),
	})
	srv.changes = func(since uint64) []byte {
		if since != 10 {
			t.Errorf("since=%d", since)
		}
		b, _ := json.Marshal(changesResponse{Version: 11, Paths: []string{"data/a.txt"}})
		return b
	}
	m := versionedManager(t, srv, RemoteOptions{TrustedKeys: []ed25519.PublicKey{pub}})
	m.WaitRemoteIndex(3 * time.Second)
	if _, err := m.ReadFile("data\\a.txt"); err != nil {
		t.Fatal(err)
	}
	hitsBefore := srv.hitCount("data/a.txt")

	srv.mu.Lock()
	srv.dataVersion = 11
	srv.files["data/a.txt"] = []byte("a2")
	srv.etags["data/a.txt"] = "ea2"
	srv.manifest = signedDataManifest(t, priv, 11, map[string][]byte{
		"data/a.txt": []byte("a2"),
		"data/b.txt": []byte("b1"),
	})
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
	pub, priv := testKey(t)
	srv := newFakeAssetServer(t)
	srv.dataVersion = 5
	srv.index = []string{"data\\a.txt", "data\\b.txt"}
	srv.put("data/a.txt", []byte("a1"), "e1")
	srv.put("data/b.txt", []byte("b1"), "eb")
	srv.manifest = signedDataManifest(t, priv, 5, map[string][]byte{
		"data/a.txt": []byte("a1"),
		"data/b.txt": []byte("b1"),
	})
	srv.changes = func(since uint64) []byte {
		b, _ := json.Marshal(changesResponse{Version: 50, Full: true})
		return b
	}
	m := versionedManager(t, srv, RemoteOptions{TrustedKeys: []ed25519.PublicKey{pub}})
	m.WaitRemoteIndex(3 * time.Second)
	if _, err := m.ReadFile("data\\a.txt"); err != nil {
		t.Fatal(err)
	}
	srv.mu.Lock()
	srv.dataVersion = 50
	srv.files["data/a.txt"] = []byte("a-new")
	srv.etags["data/a.txt"] = "e2"
	srv.manifest = signedDataManifest(t, priv, 50, map[string][]byte{
		"data/a.txt": []byte("a-new"),
		"data/b.txt": []byte("b1"),
	})
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

func TestRemoteManifestChecksBodyCachedBeforeIt(t *testing.T) {
	pub, priv := testKey(t)
	bad := []byte("tampered")
	good := []byte("good-bytes")
	srv := newFakeAssetServer(t)
	srv.dataVersion = 9
	srv.index = []string{"data\\prontera.gat"}
	srv.put("data/prontera.gat", bad, "etag-bad")
	mgr := versionedManager(t, srv, RemoteOptions{TrustedKeys: []ed25519.PublicKey{pub}})
	if !mgr.WaitRemoteIndex(3 * time.Second) {
		t.Fatal("index")
	}
	got, err := mgr.ReadFile("data\\prontera.gat")
	if err != nil || string(got) != "tampered" {
		t.Fatalf("prime cache: %q %v", got, err)
	}
	hits := srv.hitCount("data/prontera.gat")
	srv.mu.Lock()
	srv.manifest = signedDataManifest(t, priv, 9, map[string][]byte{
		"data/prontera.gat": good,
	})
	srv.mu.Unlock()
	if !mgr.remote.loadDataManifest() || !mgr.WaitRemoteManifest(time.Second) {
		t.Fatal("manifest not accepted")
	}
	got, err = mgr.ReadFile("data\\prontera.gat")
	if err == nil && string(got) == "tampered" {
		t.Fatal("pre-manifest cache was served without a digest match")
	}
	if err == nil {
		t.Fatalf("mismatched cache returned %q", got)
	}
	// Same ETag: the recheck is a 304, so the rejection is of the cached body.
	if srv.hitCount("data/prontera.gat") <= hits {
		t.Fatal("cached body was not rechecked after the manifest")
	}
	dataPath, _ := mgr.remote.cachePaths("data/prontera.gat")
	if _, statErr := os.Stat(dataPath); !os.IsNotExist(statErr) {
		t.Fatal("mismatched cached body was kept")
	}
}

func TestRemoteManifestKeepsCachedBodyThatMatches(t *testing.T) {
	pub, priv := testKey(t)
	good := []byte("good-bytes")
	srv := newFakeAssetServer(t)
	srv.dataVersion = 9
	srv.index = []string{"data\\prontera.gat"}
	srv.put("data/prontera.gat", good, "etag-good")
	mgr := versionedManager(t, srv, RemoteOptions{TrustedKeys: []ed25519.PublicKey{pub}})
	mgr.WaitRemoteIndex(3 * time.Second)
	if _, err := mgr.ReadFile("data\\prontera.gat"); err != nil {
		t.Fatal(err)
	}
	srv.mu.Lock()
	srv.manifest = signedDataManifest(t, priv, 9, map[string][]byte{
		"data/prontera.gat": good,
	})
	srv.mu.Unlock()
	if !mgr.remote.loadDataManifest() {
		t.Fatal("manifest")
	}
	got, err := mgr.ReadFile("data\\prontera.gat")
	if err != nil || string(got) != "good-bytes" {
		t.Fatalf("matching cache: %q %v", got, err)
	}
}

func TestRemoteUnsignedChangesDoNotEvictOrAdvance(t *testing.T) {
	srv := newFakeAssetServer(t)
	srv.dataVersion = 10
	srv.index = []string{"data\\a.gat", "data\\b.gat"}
	srv.put("data/a.gat", []byte("a1"), "ea")
	srv.put("data/b.gat", []byte("b1"), "eb")
	srv.changes = func(since uint64) []byte {
		b, _ := json.Marshal(changesResponse{Version: 11, Paths: []string{"data/a.gat"}})
		return b
	}
	m := versionedManager(t, srv, RemoteOptions{})
	m.WaitRemoteIndex(3 * time.Second)
	if _, err := m.ReadFile("data\\a.gat"); err != nil {
		t.Fatal(err)
	}
	srv.mu.Lock()
	srv.dataVersion = 11
	srv.files["data/a.gat"] = []byte("a2")
	srv.etags["data/a.gat"] = "ea2"
	srv.mu.Unlock()
	if _, err := m.ReadFile("data\\b.gat"); err != nil {
		t.Fatal(err)
	}
	time.Sleep(200 * time.Millisecond)
	if m.RemoteCachedVersion() != 10 {
		t.Fatalf("unsigned changes advanced cached version to %d", m.RemoteCachedVersion())
	}
	data, err := m.ReadFile("data\\a.gat")
	if err != nil || string(data) != "a1" {
		t.Fatalf("unsigned changes evicted the cache: %q %v", data, err)
	}
	srv.mu.Lock()
	hits := srv.changesHits
	srv.mu.Unlock()
	if hits > 4 {
		t.Fatalf("changes requests spun: %d", hits)
	}
}

func TestRemoteChangesRejectsRollbackAndDoesNotSpin(t *testing.T) {
	srv := newFakeAssetServer(t)
	srv.dataVersion = 10
	srv.index = []string{"data\\a.gat", "data\\b.gat"}
	srv.put("data/a.gat", []byte("a1"), "ea")
	srv.put("data/b.gat", []byte("b1"), "eb")
	srv.changes = func(since uint64) []byte {
		b, _ := json.Marshal(changesResponse{Version: 4, Paths: []string{"data/a.gat"}})
		return b
	}
	m := versionedManager(t, srv, RemoteOptions{})
	m.WaitRemoteIndex(3 * time.Second)
	if _, err := m.ReadFile("data\\a.gat"); err != nil {
		t.Fatal(err)
	}
	srv.mu.Lock()
	srv.dataVersion = 40
	srv.files["data/a.gat"] = []byte("a2")
	srv.etags["data/a.gat"] = "ea2"
	srv.mu.Unlock()
	if _, err := m.ReadFile("data\\b.gat"); err != nil {
		t.Fatal(err)
	}
	time.Sleep(250 * time.Millisecond)
	if m.RemoteCachedVersion() != 10 {
		t.Fatalf("rollback moved cached version to %d", m.RemoteCachedVersion())
	}
	data, err := m.ReadFile("data\\a.gat")
	if err != nil || string(data) != "a1" {
		t.Fatalf("rollback evicted the cache: %q %v", data, err)
	}
	srv.mu.Lock()
	hits := srv.changesHits
	srv.mu.Unlock()
	if hits == 0 || hits > 4 {
		t.Fatalf("changes hits = %d, want a bounded retry", hits)
	}
}

func TestRemoteChangesStallDoesNotTightLoop(t *testing.T) {
	pub, priv := testKey(t)
	bodyA := []byte("a1")
	bodyB := []byte("b1")
	srv := newFakeAssetServer(t)
	srv.dataVersion = 10
	srv.index = []string{"data\\a.gat", "data\\b.gat"}
	srv.put("data/a.gat", bodyA, "ea")
	srv.put("data/b.gat", bodyB, "eb")
	srv.manifest = signedDataManifest(t, priv, 15, map[string][]byte{
		"data/a.gat": bodyA,
		"data/b.gat": bodyB,
	})
	srv.changes = func(since uint64) []byte {
		b, _ := json.Marshal(changesResponse{Version: 15, Paths: []string{"data/a.gat"}})
		return b
	}
	m := versionedManager(t, srv, RemoteOptions{TrustedKeys: []ed25519.PublicKey{pub}})
	m.WaitRemoteIndex(3 * time.Second)
	if !m.WaitRemoteManifest(3 * time.Second) {
		t.Fatal("manifest")
	}
	if _, err := m.ReadFile("data\\a.gat"); err != nil {
		t.Fatal(err)
	}
	srv.mu.Lock()
	srv.dataVersion = 40
	srv.mu.Unlock()
	if _, err := m.ReadFile("data\\b.gat"); err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) && m.RemoteCachedVersion() != 15 {
		time.Sleep(10 * time.Millisecond)
	}
	if m.RemoteCachedVersion() != 15 {
		t.Fatalf("cached %d, want 15 (not seen 40)", m.RemoteCachedVersion())
	}
	time.Sleep(250 * time.Millisecond)
	srv.mu.Lock()
	hits := srv.changesHits
	srv.mu.Unlock()
	if hits > 6 {
		t.Fatalf("stalled changes version tight-looped: %d GETs", hits)
	}
}
