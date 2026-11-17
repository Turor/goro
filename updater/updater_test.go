package updater

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const testSHA = "0123456789abcdef0123456789abcdef01234567"

func testKey(t *testing.T) ed25519.PrivateKey {
	t.Helper()
	seed := bytes.Repeat([]byte{7}, ed25519.SeedSize)
	return ed25519.NewKeyFromSeed(seed)
}

func signed(t *testing.T, key ed25519.PrivateKey, m Manifest) []byte {
	t.Helper()
	p, err := CanonicalPayload(m)
	if err != nil {
		t.Fatal(err)
	}
	env, err := SignEnvelope(p, key)
	if err != nil {
		t.Fatal(err)
	}
	return []byte(env)
}

func TestEmbeddedKeyIsTheTuroranReleaseKey(t *testing.T) {
	if got := hex.EncodeToString(EmbeddedPublicKey); got != "fb3dd3d6b9e1955ee03a4786596f7d5562b7d96d7b6c177e6d38cf7abca7de18" {
		t.Fatalf("embedded key changed: %s", got)
	}
}

func TestEnvelopeRoundTripAndTamper(t *testing.T) {
	key := testKey(t)
	pub := key.Public().(ed25519.PublicKey)
	m := Manifest{Version: "main-0123456789ab", Build: 42, Files: []File{
		{Path: "goro-windows-amd64-" + testSHA + ".exe", SHA256: strings.Repeat("a", 64), Size: 3, URL: "goro-windows-amd64-" + testSHA + ".exe"},
		{Path: "goro-linux-amd64-" + testSHA, SHA256: strings.Repeat("b", 64), Size: 5, URL: "goro-linux-amd64-" + testSHA},
	}}
	env := signed(t, key, m)
	got, err := OpenEnvelope(env, pub)
	if err != nil {
		t.Fatal(err)
	}
	if got.Build != 42 || got.Files[0].Path != "goro-linux-amd64-"+testSHA {
		t.Fatalf("unexpected manifest %+v", got)
	}
	tampered := bytes.Replace(env, []byte("\n42\n"), []byte("\n43\n"), 1)
	if _, err := OpenEnvelope(tampered, pub); err == nil {
		t.Fatal("tampered build accepted")
	}
	other := ed25519.NewKeyFromSeed(bytes.Repeat([]byte{9}, 32)).Public().(ed25519.PublicKey)
	if _, err := OpenEnvelope(env, other); err == nil {
		t.Fatal("wrong key accepted")
	}
	if _, err := OpenEnvelope(env, make([]byte, 32)); err == nil {
		t.Fatal("zero key accepted")
	}
}

func TestKorangarManifestIsRejected(t *testing.T) {
	key := testKey(t)
	payload := "KRUP1\nmain-x\n1\n1\nkorangar-linux-" + testSHA + "\n" + strings.Repeat("a", 64) + "\n1\nx\n"
	sig := ed25519.Sign(key, []byte(payload))
	env := "KRUP1SIG\n" + hex.EncodeToString(sig) + "\n" + payload
	if _, err := OpenEnvelope([]byte(env), key.Public().(ed25519.PublicKey)); err == nil {
		t.Fatal("KRUP1 envelope accepted by goro")
	}
}

func TestNonCanonicalPayloadRejected(t *testing.T) {
	cases := []string{
		"GORO1\nv\n1\n1\nb\n" + strings.Repeat("a", 64) + "\n1\nb\nextra\n",
		"GORO1\nv\n01\n1\nb\n" + strings.Repeat("a", 64) + "\n1\nb\n",
		"GORO1\nv\n1\n2\nz\n" + strings.Repeat("a", 64) + "\n1\nz\na\n" + strings.Repeat("a", 64) + "\n1\na\n",
		"GORO1\nv\n1\n1\n../x\n" + strings.Repeat("a", 64) + "\n1\nx\n",
		"GORO1\nv\n1\n1\nupdates/x\n" + strings.Repeat("a", 64) + "\n1\nx\n",
		"GORO1\nv\n0\n1\nx\n" + strings.Repeat("a", 64) + "\n1\nx\n",
	}
	for _, c := range cases {
		if _, err := ParsePayload(c); err == nil {
			t.Errorf("accepted %q", c)
		}
	}
}

func TestAllowedURL(t *testing.T) {
	ok := []string{
		"https://ro.turoran.com/client/goro/main/latest",
		"https://garage.turoran.com/korangar-releases/goro/main/manifest.goroup?X-Amz-Signature=abc",
		"https://garage.turoran.com/korangar-releases/goro/main/goro-linux-amd64-" + testSHA,
		"http://127.0.0.1:8080/x",
		"file:///tmp/m",
	}
	bad := []string{
		"https://ro.turoran.com/client/korangar/main/latest",
		"https://garage.turoran.com/korangar-releases/main/korangar-linux-" + testSHA,
		"https://garage.turoran.com/korangar-releases/goro/main/../main/x",
		"https://evil.example/client/goro/main/latest",
		"http://ro.turoran.com/client/goro/main/latest",
		"https://ro.turoran.com:8443/client/goro/main/latest",
		"https://user@ro.turoran.com/client/goro/main/latest",
		"http://192.168.11.54/x",
		"file://host/tmp/m",
	}
	for _, u := range ok {
		if !allowedURL(u) {
			t.Errorf("rejected %s", u)
		}
	}
	for _, u := range bad {
		if allowedURL(u) {
			t.Errorf("accepted %s", u)
		}
	}
}

func TestPlatformFile(t *testing.T) {
	m := Manifest{Files: []File{
		{Path: "goro-linux-amd64-" + testSHA},
		{Path: "goro-windows-amd64-" + testSHA + ".exe"},
		{Path: "goro-windows-amd64-" + testSHA},
		{Path: "README.txt"},
	}}
	if f, ok := platformFile(m, "linux", "amd64"); !ok || f.Path != "goro-linux-amd64-"+testSHA {
		t.Fatalf("linux pick %v %v", f, ok)
	}
	if f, ok := platformFile(m, "windows", "amd64"); !ok || !strings.HasSuffix(f.Path, ".exe") {
		t.Fatalf("windows pick %v %v", f, ok)
	}
	if _, ok := platformFile(m, "linux", "arm64"); ok {
		t.Fatal("arm64 should have no file")
	}
}

func TestSelfCheckOutput(t *testing.T) {
	if err := checkSelfCheckOutput(SelfCheckToken+" build=42 version=v\n", 42); err != nil {
		t.Fatal(err)
	}
	if err := checkSelfCheckOutput(SelfCheckToken+" build=41 version=v\n", 42); err == nil {
		t.Fatal("wrong build accepted")
	}
	if err := checkSelfCheckOutput("hello\n", 42); err == nil {
		t.Fatal("missing token accepted")
	}
}

func TestHandleFlags(t *testing.T) {
	var out bytes.Buffer
	if !HandleFlags([]string{SelfCheckFlag}, &out) || !strings.HasPrefix(out.String(), SelfCheckToken+" build=") {
		t.Fatalf("self-check output %q", out.String())
	}
	if HandleFlags([]string{"--data-dir", "x"}, &out) {
		t.Fatal("plain args handled")
	}
}

func TestEnabled(t *testing.T) {
	t.Setenv(EnvSkip, "")
	t.Setenv(EnvForce, "")
	if Enabled(nil) {
		t.Fatal("dev build (no Channel) must not update")
	}
	t.Setenv(EnvForce, "1")
	if !Enabled(nil) {
		t.Fatal("forced check disabled")
	}
	if Enabled([]string{"--headless"}) {
		t.Fatal("headless must not update")
	}
	t.Setenv(EnvSkip, "1")
	if Enabled(nil) {
		t.Fatal("skip env ignored")
	}
}

// fixture is an install directory with a fake running binary and a release
// directory holding a signed manifest and the new binary.
type fixture struct {
	root, exe, rel string
	newBin         []byte
	key            ed25519.PrivateKey
	selfChecks     []string
}

func newFixture(t *testing.T, build uint64, goos string) *fixture {
	t.Helper()
	f := &fixture{root: t.TempDir(), rel: t.TempDir(), key: testKey(t), newBin: []byte("new goro binary")}
	name := "goro"
	if goos == "windows" {
		name = "goro.exe"
	}
	f.exe = filepath.Join(f.root, name)
	if err := os.WriteFile(f.exe, []byte("old goro binary"), 0o755); err != nil {
		t.Fatal(err)
	}
	f.writeRelease(t, build, f.newBin)
	return f
}

func (f *fixture) artifactName(goos string) string {
	if goos == "windows" {
		return "goro-windows-amd64-" + testSHA + ".exe"
	}
	return "goro-linux-amd64-" + testSHA
}

func (f *fixture) writeRelease(t *testing.T, build uint64, bin []byte) {
	t.Helper()
	files := []File{}
	for _, goos := range []string{"linux", "windows"} {
		n := f.artifactName(goos)
		if err := os.WriteFile(filepath.Join(f.rel, n), bin, 0o644); err != nil {
			t.Fatal(err)
		}
		files = append(files, File{Path: n, SHA256: SHA256Hex(bin), Size: int64(len(bin)), URL: n})
	}
	env := signed(t, f.key, Manifest{Version: "main-0123456789ab", Build: build, Files: files})
	if err := os.WriteFile(filepath.Join(f.rel, "manifest.goroup"), env, 0o644); err != nil {
		t.Fatal(err)
	}
}

func (f *fixture) opts(url, goos string) Options {
	return Options{
		URL: url, InstallRoot: f.root, Executable: f.exe,
		PublicKey: f.key.Public().(ed25519.PublicKey), Installed: 10,
		GOOS: goos, GOARCH: "amd64",
		SelfCheck: func(exe string, want uint64) error {
			f.selfChecks = append(f.selfChecks, exe)
			return nil
		},
	}
}

func (f *fixture) manifestURL() string {
	return "file://" + filepath.ToSlash(filepath.Join(f.rel, "manifest.goroup"))
}

func TestApplyFromFileURL(t *testing.T) {
	for _, goos := range []string{"linux", "windows"} {
		t.Run(goos, func(t *testing.T) {
			f := newFixture(t, 20, goos)
			res, err := Run(context.Background(), f.opts(f.manifestURL(), goos))
			if err != nil || res.Outcome != Applied {
				t.Fatalf("res=%+v err=%v", res, err)
			}
			got, _ := os.ReadFile(f.exe)
			if !bytes.Equal(got, f.newBin) {
				t.Fatalf("exe not replaced: %q", got)
			}
			if old, _ := os.ReadFile(f.exe + ".old"); string(old) != "old goro binary" {
				t.Fatalf("old binary not kept aside: %q", old)
			}
			if appliedBuild(f.root) != 20 {
				t.Fatal("applied.build not recorded")
			}
			if len(f.selfChecks) != 2 {
				t.Fatalf("want staged and installed self-checks, got %v", f.selfChecks)
			}
			if _, err := os.Stat(filepath.Join(f.root, "updates", "staging", "main-0123456789ab")); !os.IsNotExist(err) {
				t.Fatal("staging not cleaned")
			}
			// Second launch: same build is up to date, and .old is removed.
			res, err = Run(context.Background(), f.opts(f.manifestURL(), goos))
			if err != nil || res.Outcome != UpToDate {
				t.Fatalf("second run res=%+v err=%v", res, err)
			}
			if _, err := os.Stat(f.exe + ".old"); !os.IsNotExist(err) {
				t.Fatal(".old not cleaned on next launch")
			}
		})
	}
}

func TestRollbackAndSameBuild(t *testing.T) {
	f := newFixture(t, 5, "linux")
	res, err := Run(context.Background(), f.opts(f.manifestURL(), "linux"))
	if err != nil || res.Outcome != Skipped || !strings.Contains(res.Reason, "rollback") {
		t.Fatalf("res=%+v err=%v", res, err)
	}
	f.writeRelease(t, 10, f.newBin)
	if res, _ := Run(context.Background(), f.opts(f.manifestURL(), "linux")); res.Outcome != UpToDate {
		t.Fatalf("same build: %+v", res)
	}
	assertUnchanged(t, f)
}

func TestHashMismatchLeavesInstallUnchanged(t *testing.T) {
	f := newFixture(t, 20, "linux")
	if err := os.WriteFile(filepath.Join(f.rel, f.artifactName("linux")), []byte("bad goro binary"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Run(context.Background(), f.opts(f.manifestURL(), "linux")); err == nil || !strings.Contains(err.Error(), "sha256 mismatch") {
		t.Fatalf("err=%v", err)
	}
	assertUnchanged(t, f)
}

func TestBadSignatureLeavesInstallUnchanged(t *testing.T) {
	f := newFixture(t, 20, "linux")
	o := f.opts(f.manifestURL(), "linux")
	o.PublicKey = EmbeddedPublicKey // test key != release key
	if _, err := Run(context.Background(), o); err == nil {
		t.Fatal("manifest signed by another key accepted")
	}
	assertUnchanged(t, f)
}

func TestSelfCheckFailureRestoresBinary(t *testing.T) {
	f := newFixture(t, 20, "linux")
	o := f.opts(f.manifestURL(), "linux")
	calls := 0
	o.SelfCheck = func(string, uint64) error {
		calls++
		if calls == 2 {
			return errors.New("boom")
		}
		return nil
	}
	if _, err := Run(context.Background(), o); err == nil {
		t.Fatal("self-check failure ignored")
	}
	assertUnchanged(t, f)
}

func TestNoPlatformBuildIsSkipped(t *testing.T) {
	f := newFixture(t, 20, "linux")
	o := f.opts(f.manifestURL(), "linux")
	o.GOARCH = "arm64"
	res, err := Run(context.Background(), o)
	if err != nil || res.Outcome != Skipped || !strings.Contains(res.Reason, "no linux/arm64") {
		t.Fatalf("res=%+v err=%v", res, err)
	}
	assertUnchanged(t, f)
}

func TestUnreachableChannelIsSkipped(t *testing.T) {
	f := newFixture(t, 20, "linux")
	srv := httptest.NewServer(http.NotFoundHandler())
	defer srv.Close()
	res, err := Run(context.Background(), f.opts(srv.URL+"/latest", "linux"))
	if err != nil || res.Outcome != Skipped {
		t.Fatalf("res=%+v err=%v", res, err)
	}
	srv.Close()
	res, err = Run(context.Background(), f.opts(srv.URL+"/latest", "linux"))
	if err != nil || res.Outcome != Skipped {
		t.Fatalf("closed server res=%+v err=%v", res, err)
	}
	assertUnchanged(t, f)
}

// The catalog flow mirrors remoteclient2: a JSON document with a manifest URL
// and per-file URLs (presigned in production).
func TestApplyThroughCatalog(t *testing.T) {
	f := newFixture(t, 20, "linux")
	mux := http.NewServeMux()
	var srv *httptest.Server
	mux.HandleFunc("/client/goro/main/latest", func(w http.ResponseWriter, r *http.Request) {
		n := f.artifactName("linux")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"expires_in":   600,
			"manifest_url": srv.URL + "/obj/manifest.goroup?sig=1",
			"files":        []map[string]string{{"path": n, "url": srv.URL + "/obj/" + n + "?sig=2"}},
		})
	})
	mux.Handle("/obj/", http.StripPrefix("/obj/", http.FileServer(http.Dir(f.rel))))
	srv = httptest.NewServer(mux)
	defer srv.Close()
	res, err := Run(context.Background(), f.opts(srv.URL+"/client/goro/main/latest", "linux"))
	if err != nil || res.Outcome != Applied {
		t.Fatalf("res=%+v err=%v", res, err)
	}
}

func TestCatalogWithForeignHostIsRejected(t *testing.T) {
	f := newFixture(t, 20, "linux")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"manifest_url":"https://evil.example/manifest.goroup","files":[]}`))
	}))
	defer srv.Close()
	if _, err := Run(context.Background(), f.opts(srv.URL, "linux")); err == nil {
		t.Fatal("foreign manifest host accepted")
	}
	assertUnchanged(t, f)
}

func TestConcurrentApplyIsLocked(t *testing.T) {
	f := newFixture(t, 20, "linux")
	if err := os.MkdirAll(filepath.Join(f.root, "updates"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(f.root, "updates", ".lock"), []byte("1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Run(context.Background(), f.opts(f.manifestURL(), "linux")); err == nil || !strings.Contains(err.Error(), "another goro") {
		t.Fatalf("err=%v", err)
	}
	assertUnchanged(t, f)
}

func assertUnchanged(t *testing.T, f *fixture) {
	t.Helper()
	got, _ := os.ReadFile(f.exe)
	if string(got) != "old goro binary" {
		t.Fatalf("installed binary changed: %q", got)
	}
	if appliedBuild(f.root) != 0 {
		t.Fatal("applied.build written")
	}
}

func TestSizeMismatchLeavesInstallUnchanged(t *testing.T) {
	f := newFixture(t, 20, "linux")
	if err := os.WriteFile(filepath.Join(f.rel, f.artifactName("linux")), []byte("longer goro binary"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Run(context.Background(), f.opts(f.manifestURL(), "linux")); err == nil || !strings.Contains(err.Error(), "size mismatch") {
		t.Fatalf("err=%v", err)
	}
	assertUnchanged(t, f)
}
