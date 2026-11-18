package updater

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"strings"
	"testing"
)

func generateSeed(t *testing.T) (ed25519.PublicKey, ed25519.PrivateKey) {
	t.Helper()
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	return pub, priv
}

func signedData(t *testing.T, key ed25519.PrivateKey, m DataManifest) []byte {
	t.Helper()
	p, err := CanonicalDataPayload(m)
	if err != nil {
		t.Fatal(err)
	}
	env, err := SignDataEnvelope(p, key)
	if err != nil {
		t.Fatal(err)
	}
	return []byte(env)
}

func sampleDataManifest(version uint64, body []byte) DataManifest {
	return DataManifest{
		Version: version,
		Files: []DataFile{
			{Path: "data/lua files/skillinfoz/skillid.lub", SHA256: SHA256Hex(body), Size: int64(len(body))},
			{Path: "System/itemInfo.lua", SHA256: SHA256Hex([]byte("items")), Size: 5},
		},
	}
}

func TestDataManifestRoundTripAndBadSignature(t *testing.T) {
	pub, priv := generateSeed(t)
	body := []byte("skill-id-table")
	env := signedData(t, priv, sampleDataManifest(100, body))
	got, err := OpenDataEnvelope(env, pub)
	if err != nil {
		t.Fatal(err)
	}
	if got.Version != 100 || len(got.Files) != 2 {
		t.Fatalf("unexpected %+v", got)
	}
	if err := got.VerifyFile("data/lua files/skillinfoz/skillid.lub", body); err != nil {
		t.Fatal(err)
	}

	tampered := bytes.Replace(env, []byte("\n100\n"), []byte("\n101\n"), 1)
	if _, err := OpenDataEnvelope(tampered, pub); err == nil {
		t.Fatal("tampered version accepted")
	}
	other, _ := generateSeed(t)
	if _, err := OpenDataEnvelope(env, other); err == nil {
		t.Fatal("wrong key accepted")
	}
	if _, err := OpenDataEnvelope(env, make([]byte, 32)); err == nil {
		t.Fatal("zero key accepted")
	}
}

func TestDataManifestRollbackVersionRejected(t *testing.T) {
	pub, priv := generateSeed(t)
	env := signedData(t, priv, sampleDataManifest(50, []byte("v50")))
	got, err := OpenDataEnvelope(env, pub)
	if err != nil {
		t.Fatal(err)
	}
	if err := got.RejectRollback(100); err == nil {
		t.Fatal("older version accepted")
	}
	if err := got.RejectRollback(50); err != nil {
		t.Fatalf("equal version rejected: %v", err)
	}
	if err := got.RejectRollback(0); err != nil {
		t.Fatalf("fresh cache rejected: %v", err)
	}
}

func TestDataManifestDigestMismatchRejected(t *testing.T) {
	pub, priv := generateSeed(t)
	body := []byte("skill-id-table")
	env := signedData(t, priv, sampleDataManifest(7, body))
	got, err := OpenDataEnvelope(env, pub)
	if err != nil {
		t.Fatal(err)
	}
	if err := got.VerifyFile("data/lua files/skillinfoz/skillid.lub", []byte("tampered")); err == nil {
		t.Fatal("digest mismatch accepted")
	}
	if err := got.VerifyFile("data/lua files/skillinfoz/skillid.lub", append(body, 'x')); err == nil {
		t.Fatal("size mismatch accepted")
	}
	if err := got.VerifyFile("data/missing.txt", body); err == nil {
		t.Fatal("unknown path accepted")
	}
}

func TestDataManifestTrustedKeyList(t *testing.T) {
	pub1, priv1 := generateSeed(t)
	pub2, priv2 := generateSeed(t)
	body := []byte("table")
	env1 := signedData(t, priv1, sampleDataManifest(1, body))
	env2 := signedData(t, priv2, sampleDataManifest(2, body))
	keys := []ed25519.PublicKey{pub1, pub2}
	if _, err := OpenDataEnvelopeWithKeys(env1, keys); err != nil {
		t.Fatalf("first trusted key: %v", err)
	}
	if _, err := OpenDataEnvelopeWithKeys(env2, keys); err != nil {
		t.Fatalf("second trusted key: %v", err)
	}
	other, _ := generateSeed(t)
	if _, err := OpenDataEnvelopeWithKeys(env1, []ed25519.PublicKey{other}); err == nil {
		t.Fatal("untrusted key accepted")
	}
}

func TestDataManifestNextKeyAcceptance(t *testing.T) {
	oldPub, oldPriv := generateSeed(t)
	newPub, newPriv := generateSeed(t)
	body := []byte("table")
	m := sampleDataManifest(3, body)
	m.NextKeys = []ed25519.PublicKey{oldPub, newPub}
	env := signedData(t, oldPriv, m)
	got, err := OpenDataEnvelope(env, oldPub)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.NextKeys) != 2 {
		t.Fatalf("next-keys %d", len(got.NextKeys))
	}

	dir := t.TempDir()
	ring := NewKeyRing(dir+"/trusted-keys.json", []ed25519.PublicKey{oldPub})
	if err := ring.AcceptNextKeys(got.NextKeys); err != nil {
		t.Fatal(err)
	}
	later := signedData(t, newPriv, sampleDataManifest(4, body))
	if _, err := OpenDataEnvelopeWithKeys(later, ring.Keys()); err != nil {
		t.Fatalf("new key after next-keys not trusted: %v", err)
	}
	reloaded := NewKeyRing(dir+"/trusted-keys.json", []ed25519.PublicKey{oldPub})
	if _, err := OpenDataEnvelopeWithKeys(later, reloaded.Keys()); err != nil {
		t.Fatalf("persisted next-key not loaded: %v", err)
	}
}

func TestGORO1RejectedAsDataManifest(t *testing.T) {
	key := testKey(t)
	env := signed(t, key, Manifest{Version: "main-x", Build: 1, Files: []File{
		{Path: "goro-linux-amd64-" + testSHA, SHA256: strings.Repeat("a", 64), Size: 1, URL: "x"},
	}})
	if _, err := OpenDataEnvelope(env, key.Public().(ed25519.PublicKey)); err == nil {
		t.Fatal("GORO1 envelope accepted as GDAT1")
	}
}

func TestDataPathAllowsSpacesRejectsTraversal(t *testing.T) {
	if err := validateDataPath("data/lua files/x.lub"); err != nil {
		t.Fatal(err)
	}
	for _, p := range []string{"/abs", "a\\b", "../x", "a/../b", "a//b", "C:/x", "a/\x00b"} {
		if err := validateDataPath(p); err == nil {
			t.Errorf("accepted %q", p)
		}
	}
}

func TestCanonicalDataPayloadOmitsEmptyNextKeys(t *testing.T) {
	m := sampleDataManifest(1, []byte("x"))
	p, err := CanonicalDataPayload(m)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(p, "next-keys") {
		t.Fatal("empty next-keys must be omitted")
	}
}

func TestTrustedPublicKeysStartsWithEmbedded(t *testing.T) {
	if len(TrustedPublicKeys) == 0 || !bytes.Equal(TrustedPublicKeys[0], EmbeddedPublicKey) {
		t.Fatal("compiled-in list must start with the embedded release key")
	}
	if got := hex.EncodeToString(EmbeddedPublicKey); got != "fb3dd3d6b9e1955ee03a4786596f7d5562b7d96d7b6c177e6d38cf7abca7de18" {
		t.Fatalf("embedded key changed: %s", got)
	}
}
