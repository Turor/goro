package main

import (
	"bytes"
	"crypto/ed25519"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kivutar/goro/updater"
)

func TestSignVerifyRoundTrip(t *testing.T) {
	dir := t.TempDir()
	seed := filepath.Join(dir, "seed")
	if err := os.WriteFile(seed, bytes.Repeat([]byte{3}, 32), 0o600); err != nil {
		t.Fatal(err)
	}
	bin := filepath.Join(dir, "goro-linux-amd64-0123456789ab")
	if err := os.WriteFile(bin, []byte("binary"), 0o644); err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(dir, "manifest.goroup")
	var buf bytes.Buffer
	if err := run([]string{"sign", "--private-key", seed, "--version", "main-0123456789ab", "--build", "7", "--out", out, "--file", "goro-linux-amd64-0123456789ab=" + bin}, &buf); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(buf.String(), hex.EncodeToString(bytes.Repeat([]byte{3}, 32))) {
		t.Fatal("seed printed")
	}
	pub := ed25519.NewKeyFromSeed(bytes.Repeat([]byte{3}, 32)).Public().(ed25519.PublicKey)
	buf.Reset()
	if err := run([]string{"verify", "--manifest", out, "--public-key", hex.EncodeToString(pub)}, &buf); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "ok version=main-0123456789ab build=7 files=1") {
		t.Fatalf("verify output %q", buf.String())
	}
	// The embedded release key must not accept a test-key manifest.
	if err := run([]string{"verify", "--manifest", out}, &buf); err == nil {
		t.Fatal("embedded key accepted a test signature")
	}
}

func TestSignDataAndNextKeys(t *testing.T) {
	dir := t.TempDir()
	seed := filepath.Join(dir, "seed")
	if err := os.WriteFile(seed, bytes.Repeat([]byte{5}, 32), 0o600); err != nil {
		t.Fatal(err)
	}
	src := filepath.Join(dir, "skillid.lub")
	if err := os.WriteFile(src, []byte("lua"), 0o644); err != nil {
		t.Fatal(err)
	}
	pub := ed25519.NewKeyFromSeed(bytes.Repeat([]byte{5}, 32)).Public().(ed25519.PublicKey)
	next := hex.EncodeToString(bytes.Repeat([]byte{9}, 32))
	out := filepath.Join(dir, "data.gdat")
	var buf bytes.Buffer
	if err := run([]string{"sign-data", "--private-key", seed, "--version", "42", "--out", out, "--file", "data/lua files/skillid.lub=" + src, "--next-key", hex.EncodeToString(pub), "--next-key", next}, &buf); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(buf.String(), hex.EncodeToString(bytes.Repeat([]byte{5}, 32))) {
		t.Fatal("seed printed")
	}
	buf.Reset()
	if err := run([]string{"verify-data", "--manifest", out, "--public-key", hex.EncodeToString(pub)}, &buf); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "ok version=42 files=1 next-keys=2") {
		t.Fatalf("verify-data output %q", buf.String())
	}
}

func TestSignDataRootDetached(t *testing.T) {
	dir := t.TempDir()
	seed := filepath.Join(dir, "seed")
	if err := os.WriteFile(seed, bytes.Repeat([]byte{7}, 32), 0o600); err != nil {
		t.Fatal(err)
	}
	pub := ed25519.NewKeyFromSeed(bytes.Repeat([]byte{7}, 32)).Public().(ed25519.PublicKey)
	body := []byte("bmp")
	root, _, err := updater.BuildDataTree(12, 9, 1_700_000_000, []updater.DataFile{{
		Path: "data/texture/abc.bmp", SHA256: updater.SHA256Hex(body), Size: int64(len(body)),
	}}, nil)
	if err != nil {
		t.Fatal(err)
	}
	rootPath := filepath.Join(dir, "root")
	if err := os.WriteFile(rootPath, []byte(root), 0o644); err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := run([]string{"sign-data-root", "--private-key", seed, "--root", rootPath}, &buf); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(buf.String(), hex.EncodeToString(bytes.Repeat([]byte{7}, 32))) {
		t.Fatal("seed printed")
	}
	if _, err := os.Stat(filepath.Join(dir, "root.sig")); err != nil {
		t.Fatal(err)
	}
	buf.Reset()
	if err := run([]string{"verify-data-root", "--root", rootPath, "--public-key", hex.EncodeToString(pub)}, &buf); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "ok version=12 previous=9 created=1700000000 next-keys=0") {
		t.Fatalf("verify-data-root output %q", buf.String())
	}
	if strings.Contains(buf.String(), "texture/abc.bmp") {
		t.Fatal("verify-data-root printed shard entries")
	}
	raw, err := os.ReadFile(rootPath)
	if err != nil {
		t.Fatal(err)
	}
	raw[len(raw)-2] ^= 1
	if err := os.WriteFile(rootPath, raw, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := run([]string{"verify-data-root", "--root", rootPath, "--public-key", hex.EncodeToString(pub)}, &buf); err == nil {
		t.Fatal("tampered root verified")
	}
}

func TestRefusesZeroAndShortSeeds(t *testing.T) {
	dir := t.TempDir()
	for name, data := range map[string][]byte{"zero": make([]byte, 32), "short": []byte("abc")} {
		p := filepath.Join(dir, name)
		_ = os.WriteFile(p, data, 0o600)
		if _, err := loadSeed(p); err == nil {
			t.Errorf("%s seed accepted", name)
		}
	}
}
