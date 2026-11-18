package main

import (
	"bytes"
	"crypto/ed25519"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"
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
