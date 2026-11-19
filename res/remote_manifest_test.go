package res

import (
	"crypto/ed25519"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/kivutar/goro/updater"
)

func enforceOpts(cache string, pub ed25519.PublicKey) RemoteOptions {
	return RemoteOptions{
		CacheDir:      cache,
		TrustedKeys:   []ed25519.PublicKey{pub},
		ManifestMode:  "enforce",
		PollMin:       time.Hour,
		PollMax:       time.Hour,
		DeltaDebounce: time.Millisecond,
	}
}

func TestRemoteRestartVerifiesPersistedRoot(t *testing.T) {
	pub, priv := testKey(t)
	srv := newFakeAssetServer(t)
	good := []byte("good-bytes")
	srv.dataVersion = 9
	srv.index = []string{"data\\a.txt"}
	srv.put("data/a.txt", good, "e")
	publishSignedDataRoot(t, srv, priv, 9, map[string][]byte{"data/a.txt": good})
	cache := t.TempDir()
	opts := enforceOpts(cache, pub)
	m := versionedManager(t, srv, opts)
	if !m.WaitRemoteManifest(3 * time.Second) {
		t.Fatal("manifest")
	}
	if _, err := os.Stat(filepath.Join(cache, "manifest-v2", "root")); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(cache, "data-version.json")); err != nil {
		t.Fatal("version cursor was not persisted")
	}
	m.StopRemote()

	srv.put("data/a.txt", []byte("tampered"), "e2")
	srv.mu.Lock()
	srv.v2Block = true
	srv.mu.Unlock()
	m2 := versionedManager(t, srv, opts)
	got, err := m2.ReadFile("data\\a.txt")
	if err == nil {
		t.Fatalf("restart served %q with the live manifest hidden", got)
	}
	if m2.remote.acceptedManifestVersion() != 9 {
		t.Fatalf("accepted %d", m2.remote.acceptedManifestVersion())
	}
}

func TestRemoteCorruptPersistedRootIsNotTrusted(t *testing.T) {
	pub, priv := testKey(t)
	srv := newFakeAssetServer(t)
	good := []byte("good-bytes")
	bad := []byte("tampered")
	srv.dataVersion = 9
	srv.index = []string{"data\\a.txt"}
	srv.put("data/a.txt", good, "e")
	publishSignedDataRoot(t, srv, priv, 9, map[string][]byte{"data/a.txt": good})
	cache := t.TempDir()
	opts := enforceOpts(cache, pub)
	m := versionedManager(t, srv, opts)
	if !m.WaitRemoteManifest(3 * time.Second) {
		t.Fatal("manifest")
	}
	m.StopRemote()

	forged, _, err := updater.BuildDataTree(9, 0, 1, []updater.DataFile{{
		Path: "data/a.txt", SHA256: updater.SHA256Hex(bad), Size: int64(len(bad)),
	}}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(cache, "manifest-v2", "root"), []byte(forged), 0o644); err != nil {
		t.Fatal(err)
	}
	srv.put("data/a.txt", bad, "e2")
	srv.mu.Lock()
	srv.v2Block = true
	srv.mu.Unlock()
	m2 := versionedManager(t, srv, opts)
	if m2.remote.manifest.Load() != nil {
		t.Fatal("corrupt persisted root was accepted without a valid signature")
	}
	if got, err := m2.ReadFile("data\\a.txt"); err == nil {
		t.Fatalf("tampered body served: %q", got)
	}
}

func TestRemoteNormalizedSpellingMatchesManifest(t *testing.T) {
	pub, priv := testKey(t)
	body := []byte("bmp-bytes")
	srv := newFakeAssetServer(t)
	srv.dataVersion = 4
	srv.index = []string{`data\Texture\ABC.bmp`}
	srv.put("data/Texture/ABC.bmp", body, "e")
	publishSignedDataRoot(t, srv, priv, 4, map[string][]byte{`data\Texture\ABC.bmp`: body})
	m := versionedManager(t, srv, enforceOpts(t.TempDir(), pub))
	if !m.WaitRemoteManifest(3 * time.Second) {
		t.Fatal("manifest")
	}
	got, err := m.ReadFile(`data\Texture\ABC.bmp`)
	if err != nil || string(got) != string(body) {
		t.Fatalf("normalized path rejected: %q %v", got, err)
	}
}

func TestRemoteReportModeServesMismatch(t *testing.T) {
	pub, priv := testKey(t)
	good := []byte("good-bytes")
	srv := newFakeAssetServer(t)
	srv.dataVersion = 9
	srv.index = []string{"data\\a.txt"}
	srv.put("data/a.txt", []byte("tampered"), "e")
	publishSignedDataRoot(t, srv, priv, 9, map[string][]byte{"data/a.txt": good})
	m := versionedManager(t, srv, RemoteOptions{TrustedKeys: []ed25519.PublicKey{pub}, ManifestMode: "report"})
	if !m.WaitRemoteManifest(3 * time.Second) {
		t.Fatal("manifest")
	}
	got, err := m.ReadFile("data\\a.txt")
	if err != nil || string(got) != "tampered" {
		t.Fatalf("report mode blocked gameplay: %q %v", got, err)
	}
}

func TestRemoteLocalListedFileMustMatch(t *testing.T) {
	pub, priv := testKey(t)
	srv := newFakeAssetServer(t)
	srv.dataVersion = 3
	publishSignedDataRoot(t, srv, priv, 3, map[string][]byte{
		"data/listed.txt": []byte("signed-bytes"),
	})
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "data"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "data", "listed.txt"), []byte("local-bytes"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "data", "extra.txt"), []byte("extra"), 0o644); err != nil {
		t.Fatal(err)
	}
	m, err := NewManager(root)
	if err != nil {
		t.Fatal(err)
	}
	if err := m.EnableRemote(RemoteOptions{
		BaseURL: srv.url(), CacheDir: t.TempDir(), TrustedKeys: []ed25519.PublicKey{pub},
		ManifestMode: "enforce", PollMin: time.Hour, PollMax: time.Hour, DeltaDebounce: time.Millisecond,
	}); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(m.StopRemote)
	if !m.WaitRemoteManifest(3 * time.Second) {
		t.Fatal("manifest")
	}
	if _, err := m.ReadFile(`data\listed.txt`); err == nil {
		t.Fatal("listed local file with a different digest was used")
	}
	got, err := m.ReadFile("data/extra.txt")
	if err != nil || string(got) != "extra" {
		t.Fatalf("unlisted local file: %q %v", got, err)
	}
}

func TestRemoteFetchesOnlyTheShardItNeeds(t *testing.T) {
	pub, priv := testKey(t)
	bmp := []byte("bmp")
	lua := []byte("lua")
	srv := newFakeAssetServer(t)
	srv.dataVersion = 6
	srv.index = []string{`data\texture\abc.bmp`}
	srv.put("data/texture/abc.bmp", bmp, "e")
	srv.put("System/itemInfo.lua", lua, "l")
	publishSignedDataRoot(t, srv, priv, 6, map[string][]byte{
		"data/texture/abc.bmp": bmp,
		"System/itemInfo.lua":  lua,
	})
	m := versionedManager(t, srv, enforceOpts(t.TempDir(), pub))
	if !m.WaitRemoteManifest(3 * time.Second) {
		t.Fatal("manifest")
	}
	if _, err := m.ReadFile(`data\texture\abc.bmp`); err != nil {
		t.Fatal(err)
	}
	bmpKey, ok := updater.NormalizeManifestPath("data/texture/abc.bmp")
	luaKey, ok2 := updater.NormalizeManifestPath("System/itemInfo.lua")
	if !ok || !ok2 {
		t.Fatal("normalize")
	}
	if bmpKey == luaKey || updater.ShardID(bmpKey) == updater.ShardID(luaKey) {
		t.Fatal("fixture paths share a shard")
	}
	if srv.metaCount("data-manifest/v/6/shards/"+updater.ShardID(bmpKey)) == 0 {
		t.Fatal("needed shard was not fetched")
	}
	if n := srv.metaCount("data-manifest/v/6/shards/" + updater.ShardID(luaKey)); n != 0 {
		t.Fatalf("unused shard fetched %d times", n)
	}
}

func TestRemoteRollbackRejectedAfterRestart(t *testing.T) {
	pub, priv := testKey(t)
	good := []byte("version-ten")
	srv := newFakeAssetServer(t)
	srv.dataVersion = 10
	srv.index = []string{"data\\a.txt"}
	srv.put("data/a.txt", good, "e")
	publishSignedDataRoot(t, srv, priv, 10, map[string][]byte{"data/a.txt": good})
	cache := t.TempDir()
	opts := enforceOpts(cache, pub)
	m := versionedManager(t, srv, opts)
	if !m.WaitRemoteManifest(3 * time.Second) {
		t.Fatal("manifest")
	}
	m.StopRemote()
	publishSignedDataRoot(t, srv, priv, 4, map[string][]byte{"data/a.txt": []byte("old")})
	m2 := versionedManager(t, srv, opts)
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) && srv.metaCount("data-manifest/root") < 2 {
		time.Sleep(10 * time.Millisecond)
	}
	got, err := m2.ReadFile("data\\a.txt")
	if err != nil || string(got) != string(good) {
		t.Fatalf("rollback replaced the accepted manifest: %q %v", got, err)
	}
	if m2.remote.acceptedManifestVersion() != 10 {
		t.Fatalf("accepted %d", m2.remote.acceptedManifestVersion())
	}
}
