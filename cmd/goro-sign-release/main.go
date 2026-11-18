// Command goro-sign-release signs and verifies TuroranRO goro release
// manifests (GORO1) and signed data manifests (GDAT1). It never prints
// private key material.
//
//	goro-sign-release sign --private-key SEED --version V --build N --out PATH \
//	    --file REL=SOURCE [--file REL=SOURCE]... [--url REL=URL]... \
//	    [--next-key HEX]...
//	goro-sign-release sign-data --private-key SEED --version N --out PATH \
//	    --file REL=SOURCE [--file REL=SOURCE]... [--root DIR] [--next-key HEX]...
//	goro-sign-release verify --manifest PATH [--public-key HEX]
//	goro-sign-release verify-data --manifest PATH [--public-key HEX]
//	goro-sign-release public-key --private-key SEED
//
// SEED is a path to a 32-byte Ed25519 seed (raw, or 64 hex characters). The
// environment variable GORO_UPDATE_SIGNING_KEY may name that path instead.
package main

import (
	"crypto/ed25519"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"io"
	iofs "io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/kivutar/goro/updater"
)

type multi []string

func (m *multi) String() string     { return strings.Join(*m, ",") }
func (m *multi) Set(v string) error { *m = append(*m, v); return nil }

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "goro-sign-release:", err)
		os.Exit(1)
	}
}

func run(args []string, out io.Writer) error {
	if len(args) == 0 {
		return errors.New("usage: goro-sign-release sign|sign-data|verify|verify-data|public-key [flags]")
	}
	switch args[0] {
	case "sign":
		return sign(args[1:], out)
	case "sign-data":
		return signData(args[1:], out)
	case "verify":
		return verify(args[1:], out)
	case "verify-data":
		return verifyData(args[1:], out)
	case "public-key":
		fs := flag.NewFlagSet("public-key", flag.ContinueOnError)
		keyPath := fs.String("private-key", os.Getenv("GORO_UPDATE_SIGNING_KEY"), "seed path")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		key, err := loadSeed(*keyPath)
		if err != nil {
			return err
		}
		pub := key.Public().(ed25519.PublicKey)
		fmt.Fprintln(out, hex.EncodeToString(pub))
		if !pub.Equal(updater.EmbeddedPublicKey) {
			fmt.Fprintln(out, "warning: does not match the key embedded in the client")
		}
		return nil
	}
	return fmt.Errorf("unknown command %q", args[0])
}

func sign(args []string, out io.Writer) error {
	fs := flag.NewFlagSet("sign", flag.ContinueOnError)
	keyPath := fs.String("private-key", os.Getenv("GORO_UPDATE_SIGNING_KEY"), "seed path")
	version := fs.String("version", "", "release version, e.g. main-0123456789ab")
	build := fs.Uint64("build", 0, "monotonic build number (unix time)")
	outPath := fs.String("out", "", "manifest output path")
	var files, urls, nextKeys multi
	fs.Var(&files, "file", "REL=SOURCE, repeatable")
	fs.Var(&urls, "url", "REL=URL override, repeatable (default: REL, relative to the manifest)")
	fs.Var(&nextKeys, "next-key", "trusted public key hex to carry for rotation, repeatable")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *outPath == "" || len(files) == 0 {
		return errors.New("--out and at least one --file are required")
	}
	key, err := loadSeed(*keyPath)
	if err != nil {
		return err
	}
	overrides := map[string]string{}
	for _, u := range urls {
		rel, target, ok := strings.Cut(u, "=")
		if !ok {
			return fmt.Errorf("bad --url %q", u)
		}
		overrides[rel] = target
	}
	next, err := parseNextKeyFlags(nextKeys)
	if err != nil {
		return err
	}
	m := updater.Manifest{Version: *version, Build: *build, NextKeys: next}
	for _, f := range files {
		rel, src, ok := strings.Cut(f, "=")
		if !ok {
			return fmt.Errorf("bad --file %q", f)
		}
		data, err := os.ReadFile(src)
		if err != nil {
			return err
		}
		u := rel
		if o, ok := overrides[rel]; ok {
			u = o
		}
		m.Files = append(m.Files, updater.File{Path: rel, SHA256: updater.SHA256Hex(data), Size: int64(len(data)), URL: u})
	}
	payload, err := updater.CanonicalPayload(m)
	if err != nil {
		return err
	}
	env, err := updater.SignEnvelope(payload, key)
	if err != nil {
		return err
	}
	// Round-trip with the public half before writing anything.
	if _, err := updater.OpenEnvelope([]byte(env), key.Public().(ed25519.PublicKey)); err != nil {
		return fmt.Errorf("self-verify failed: %w", err)
	}
	if err := os.WriteFile(*outPath, []byte(env), 0o644); err != nil {
		return err
	}
	fmt.Fprintf(out, "signed %s version=%s build=%d files=%d\n", *outPath, m.Version, m.Build, len(m.Files))
	return nil
}

func verify(args []string, out io.Writer) error {
	fs := flag.NewFlagSet("verify", flag.ContinueOnError)
	path := fs.String("manifest", "", "manifest path")
	pubHex := fs.String("public-key", "", "public key hex (default: embedded)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	pub := updater.EmbeddedPublicKey
	if *pubHex != "" {
		b, err := hex.DecodeString(*pubHex)
		if err != nil || len(b) != ed25519.PublicKeySize {
			return errors.New("bad --public-key")
		}
		pub = b
	}
	data, err := os.ReadFile(*path)
	if err != nil {
		return err
	}
	m, err := updater.OpenEnvelope(data, pub)
	if err != nil {
		return err
	}
	fmt.Fprintf(out, "ok version=%s build=%s files=%d next-keys=%d\n", m.Version, strconv.FormatUint(m.Build, 10), len(m.Files), len(m.NextKeys))
	for _, f := range m.Files {
		fmt.Fprintf(out, "  %s %s %d\n", f.Path, f.SHA256, f.Size)
	}
	return nil
}

func signData(args []string, out io.Writer) error {
	fs := flag.NewFlagSet("sign-data", flag.ContinueOnError)
	keyPath := fs.String("private-key", os.Getenv("GORO_UPDATE_SIGNING_KEY"), "seed path")
	version := fs.Uint64("version", 0, "monotonic data version")
	outPath := fs.String("out", "", "manifest output path")
	root := fs.String("root", "", "directory of files to list (optional)")
	var files, nextKeys multi
	fs.Var(&files, "file", "REL=SOURCE, repeatable")
	fs.Var(&nextKeys, "next-key", "trusted public key hex to carry for rotation, repeatable")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *outPath == "" || (*root == "" && len(files) == 0) {
		return errors.New("--out and at least one --file or --root are required")
	}
	key, err := loadSeed(*keyPath)
	if err != nil {
		return err
	}
	next, err := parseNextKeyFlags(nextKeys)
	if err != nil {
		return err
	}
	m := updater.DataManifest{Version: *version, NextKeys: next}
	for _, f := range files {
		rel, src, ok := strings.Cut(f, "=")
		if !ok {
			return fmt.Errorf("bad --file %q", f)
		}
		data, err := os.ReadFile(src)
		if err != nil {
			return err
		}
		m.Files = append(m.Files, updater.DataFile{Path: rel, SHA256: updater.SHA256Hex(data), Size: int64(len(data))})
	}
	if *root != "" {
		err := filepath.WalkDir(*root, func(path string, d iofs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				if strings.HasPrefix(d.Name(), ".") {
					return iofs.SkipDir
				}
				return nil
			}
			if strings.HasPrefix(d.Name(), ".") {
				return nil
			}
			info, err := d.Info()
			if err != nil || !info.Mode().IsRegular() {
				return err
			}
			rel, err := filepath.Rel(*root, path)
			if err != nil {
				return err
			}
			rel = filepath.ToSlash(rel)
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			m.Files = append(m.Files, updater.DataFile{Path: rel, SHA256: updater.SHA256Hex(data), Size: int64(len(data))})
			return nil
		})
		if err != nil {
			return err
		}
	}
	payload, err := updater.CanonicalDataPayload(m)
	if err != nil {
		return err
	}
	env, err := updater.SignDataEnvelope(payload, key)
	if err != nil {
		return err
	}
	if _, err := updater.OpenDataEnvelope([]byte(env), key.Public().(ed25519.PublicKey)); err != nil {
		return fmt.Errorf("self-verify failed: %w", err)
	}
	if err := os.WriteFile(*outPath, []byte(env), 0o644); err != nil {
		return err
	}
	fmt.Fprintf(out, "signed %s version=%d files=%d next-keys=%d\n", *outPath, m.Version, len(m.Files), len(m.NextKeys))
	return nil
}

func verifyData(args []string, out io.Writer) error {
	fs := flag.NewFlagSet("verify-data", flag.ContinueOnError)
	path := fs.String("manifest", "", "manifest path")
	pubHex := fs.String("public-key", "", "public key hex (default: embedded)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	pub := updater.EmbeddedPublicKey
	if *pubHex != "" {
		b, err := hex.DecodeString(*pubHex)
		if err != nil || len(b) != ed25519.PublicKeySize {
			return errors.New("bad --public-key")
		}
		pub = b
	}
	data, err := os.ReadFile(*path)
	if err != nil {
		return err
	}
	m, err := updater.OpenDataEnvelope(data, pub)
	if err != nil {
		return err
	}
	fmt.Fprintf(out, "ok version=%d files=%d next-keys=%d\n", m.Version, len(m.Files), len(m.NextKeys))
	for _, f := range m.Files {
		fmt.Fprintf(out, "  %s %s %d\n", f.Path, f.SHA256, f.Size)
	}
	return nil
}

func parseNextKeyFlags(values []string) ([]ed25519.PublicKey, error) {
	var keys []ed25519.PublicKey
	for _, h := range values {
		h = strings.ToLower(strings.TrimSpace(h))
		b, err := hex.DecodeString(h)
		if err != nil || len(b) != ed25519.PublicKeySize {
			return nil, fmt.Errorf("bad --next-key")
		}
		keys = append(keys, ed25519.PublicKey(b))
	}
	return keys, nil
}

func loadSeed(path string) (ed25519.PrivateKey, error) {
	if path == "" {
		return nil, errors.New("pass --private-key PATH or set GORO_UPDATE_SIGNING_KEY")
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, errors.New("cannot read the signing seed")
	}
	seed := raw
	if t := strings.TrimSpace(string(raw)); len(t) == 64 {
		if b, err := hex.DecodeString(t); err == nil {
			seed = b
		}
	}
	if len(seed) != ed25519.SeedSize {
		return nil, errors.New("signing seed must be 32 bytes or 64 hex characters")
	}
	allZero := true
	for _, b := range seed {
		if b != 0 {
			allZero = false
		}
	}
	if allZero {
		return nil, errors.New("refusing an all-zero seed")
	}
	return ed25519.NewKeyFromSeed(seed), nil
}
