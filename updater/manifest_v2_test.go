package updater

import (
	"bytes"
	"crypto/ed25519"
	"os"
	"strings"
	"testing"
)

func TestManifestV2NormalizeFixture(t *testing.T) {
	raw, err := os.ReadFile("../testdata/manifest-v2-normalize.txt")
	if err != nil {
		t.Fatal(err)
	}
	if bytes.HasPrefix(raw, []byte{0xEF, 0xBB, 0xBF}) {
		t.Fatal("fixture has a BOM")
	}
	if bytes.Contains(raw, []byte{'\r'}) {
		t.Fatal("fixture must be LF only")
	}
	refs := map[string]string{
		"texture/abc.bmp":                  "312e1081f58d46229528909b2e9892fda39a1052c93a037164f0fbb021b58819",
		"system/iteminfo.lua":              "428c74e1539a11afbd067b4493256d9b6220476d230bf6f54187037dfcb43a9e",
		"bgm/theme.mp3":                    "24e72b04ea46e23aeb0bfc7884de088b273f3486c6ae37727c6039238710577e",
		"lua files/skillinfoz/skillid.lub": "ae1d00f9929efdea4c66b6c64c71ca414c3233546f8d2775cd5bba68dce717f5",
		"data/foo":                         "0a26bd10d4064b5c0ad5d5d133ead6dea8434e88ddf9360dae0501781a52d941",
		"ai/test.mtd":                      "56bb9c5237fa84479ab645b7163c8792d1f9ccb111251ebf4bc3ea39efc51d93",
		"texture/유.bmp":                    "60e90ae32634d0d17bc0e7690656d36614cc07f146ae5704bea710474828567a",
		"texture/유저인터페이스/login_interface/titlebar_left.bmp": "62a29913aeecb33cc194e310bb37535f489295251fdf6883be5d1dda4885d0fb",
	}
	for path, full := range refs {
		if got := SHA256Hex([]byte(path)); got != full {
			t.Fatalf("sha256(%q) = %s", path, got)
		}
		if ShardID(path) != full[:2] {
			t.Fatalf("shard(%q) = %s", path, ShardID(path))
		}
	}
	n := 0
	for i, line := range strings.Split(string(raw), "\n") {
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.Split(line, "\t")
		if len(parts) != 3 {
			t.Fatalf("line %d: want 3 columns, got %q", i+1, line)
		}
		in, want, shard := parts[0], parts[1], parts[2]
		var got string
		var ok bool
		if strings.Contains(in, "%") {
			got, ok = DecodeThenNormalize(in)
		} else {
			got, ok = NormalizeManifestPath(in)
			got2, ok2 := DecodeThenNormalize(in)
			if got != got2 || ok != ok2 {
				t.Fatalf("line %d: DecodeThenNormalize disagreed: %q/%v vs %q/%v", i+1, got, ok, got2, ok2)
			}
		}
		n++
		if want == "REJECT" {
			if ok {
				t.Fatalf("line %d: %q normalized to %q", i+1, in, got)
			}
			if shard != "-" {
				t.Fatalf("line %d: reject shard %q", i+1, shard)
			}
			continue
		}
		if !ok || got != want {
			t.Fatalf("line %d: %q got %q ok=%v want %q", i+1, in, got, ok, want)
		}
		if ShardID(got) != shard {
			t.Fatalf("line %d: shard(%q)=%s want %s", i+1, got, ShardID(got), shard)
		}
	}
	if n < 20 {
		t.Fatalf("fixture checked %d vectors, want the shared set", n)
	}
}

func TestDataRootRoundTripDetachedSignature(t *testing.T) {
	pub, priv := generateSeed(t)
	body := []byte("bmp-bytes")
	root, shards, err := BuildDataTree(9, 0, 1_700_000_000, []DataFile{
		{Path: `data\Texture\ABC.bmp`, SHA256: SHA256Hex(body), Size: int64(len(body))},
		{Path: "System/itemInfo.lua", SHA256: SHA256Hex([]byte("items")), Size: 5},
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := shards["31"]; !ok {
		t.Fatal("missing shard 31")
	}
	sig, err := SignDataRoot(root, priv)
	if err != nil {
		t.Fatal(err)
	}
	for _, rawSig := range [][]byte{[]byte(sig), []byte(sig + "\n")} {
		got, err := OpenDataRoot([]byte(root), rawSig, []ed25519.PublicKey{pub})
		if err != nil {
			t.Fatal(err)
		}
		if got.Version != 9 || got.CreatedUnix != 1_700_000_000 {
			t.Fatalf("root %+v", got)
		}
	}
	shard, err := ParseDataShard(shards["31"])
	if err != nil {
		t.Fatal(err)
	}
	f, ok := shard.File("texture/abc.bmp")
	if !ok || f.SHA256 != SHA256Hex(body) {
		t.Fatalf("shard entry %+v ok=%v", f, ok)
	}
	// data/ prefix is stripped once, so the client spelling is not the key.
	if _, ok := shard.File(`data\Texture\ABC.bmp`); ok {
		t.Fatal("shard kept the client spelling")
	}
	tampered := []byte(root)
	tampered[len(dataRootMagic)+1] ^= 1
	if _, err := OpenDataRoot(tampered, []byte(sig+"\n"), []ed25519.PublicKey{pub}); err == nil {
		t.Fatal("tampered root accepted")
	}
	other, _ := generateSeed(t)
	if _, err := OpenDataRoot([]byte(root), []byte(sig), []ed25519.PublicKey{other}); err == nil {
		t.Fatal("wrong key accepted")
	}
	if _, err := ParseDataRoot(root + "\n"); err == nil {
		t.Fatal("extra newline accepted")
	}
}

func TestDataRootNextKeysAndRollback(t *testing.T) {
	oldPub, oldPriv := generateSeed(t)
	_, newPriv := generateSeed(t)
	newPub := newPriv.Public().(ed25519.PublicKey)
	body := []byte("table")
	root, _, err := BuildDataTree(3, 2, 10, []DataFile{
		{Path: "ai/test.mtd", SHA256: SHA256Hex(body), Size: int64(len(body))},
	}, []ed25519.PublicKey{oldPub, newPub})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Count(root, "next-keys") != 1 {
		t.Fatalf("next-keys section missing:\n%s", root)
	}
	sig, err := SignDataRoot(root, oldPriv)
	if err != nil {
		t.Fatal(err)
	}
	got, err := OpenDataRoot([]byte(root), []byte(sig+"\n"), []ed25519.PublicKey{oldPub})
	if err != nil {
		t.Fatal(err)
	}
	if len(got.NextKeys) != 2 {
		t.Fatalf("next-keys %d", len(got.NextKeys))
	}
	if err := got.RejectRollback(3); err != nil {
		t.Fatal(err)
	}
	if err := got.RejectRollback(4); err == nil {
		t.Fatal("rollback accepted")
	}
	// A skipped version is not a rollback.
	if err := got.RejectRollback(1); err != nil {
		t.Fatal(err)
	}
	empty, _, err := BuildDataTree(4, 3, 11, []DataFile{
		{Path: "ai/test.mtd", SHA256: SHA256Hex(body), Size: int64(len(body))},
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(empty, "next-keys") {
		t.Fatal("empty next-keys must be omitted")
	}
	laterSig, err := SignDataRoot(empty, newPriv)
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	ring := NewKeyRing(dir+"/trusted-keys.json", []ed25519.PublicKey{oldPub})
	if err := ring.AcceptNextKeys(got.NextKeys); err != nil {
		t.Fatal(err)
	}
	if _, err := OpenDataRoot([]byte(empty), []byte(laterSig), ring.Keys()); err != nil {
		t.Fatalf("rotated key: %v", err)
	}
}

func TestDataRootRejectsLooseTreeAndBadShard(t *testing.T) {
	if _, err := SignDataRoot("GDAT1\n1\n0\n", ed25519.PrivateKey(bytes.Repeat([]byte{1}, 64))); err == nil {
		t.Fatal("GDAT1 accepted as a root")
	}
	_, _, err := BuildDataTree(1, 0, 1, []DataFile{
		{Path: "data/../secret", SHA256: strings.Repeat("ab", 32), Size: 1},
	}, nil)
	if err == nil {
		t.Fatal("traversal path accepted")
	}
	_, _, err = BuildDataTree(1, 0, 1, []DataFile{
		{Path: "data/a.txt", SHA256: strings.Repeat("ab", 32), Size: 1},
		{Path: "Data/A.TXT", SHA256: strings.Repeat("cd", 32), Size: 1},
	}, nil)
	if err == nil {
		t.Fatal("duplicate normalized path accepted")
	}
	root, shards, err := BuildDataTree(8, 0, 1, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	empty, err := ParseDataShard(shards["00"])
	if err != nil || len(empty.Files) != 0 || empty.Version != 8 {
		t.Fatalf("empty shard: %+v %v", empty, err)
	}
	if _, err := ParseDataRoot(strings.Replace(root, "\n8\n", "\n08\n", 1)); err == nil {
		t.Fatal("leading zero accepted")
	}
}
