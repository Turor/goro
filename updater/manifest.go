package updater

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

const (
	payloadMagic  = "GORO1"
	envelopeMagic = "GORO1SIG"
	maxFiles      = 64
	maxTokenLen   = 512
)

// EmbeddedPublicKey is the Turoran release public key (the same controller key
// that signs Korangar releases). The matching seed never leaves the signing
// controller and is never stored in this repository.
var EmbeddedPublicKey = mustHexKey("fb3dd3d6b9e1955ee03a4786596f7d5562b7d96d7b6c177e6d38cf7abca7de18")

func mustHexKey(s string) ed25519.PublicKey {
	b, err := hex.DecodeString(s)
	if err != nil || len(b) != ed25519.PublicKeySize {
		panic("updater: bad embedded public key")
	}
	return ed25519.PublicKey(b)
}

// File is one artifact listed in a manifest.
type File struct {
	Path   string
	SHA256 string
	Size   int64
	URL    string
}

// Manifest is a verified release description.
type Manifest struct {
	Version string
	Build   uint64
	Files   []File
}

// CanonicalPayload returns the exact bytes that are signed.
func CanonicalPayload(m Manifest) (string, error) {
	if err := validateVersion(m.Version); err != nil {
		return "", err
	}
	if m.Build == 0 {
		return "", errors.New("build must be a positive monotonic counter or timestamp")
	}
	if len(m.Files) == 0 {
		return "", errors.New("manifest has no files")
	}
	if len(m.Files) > maxFiles {
		return "", errors.New("too many files")
	}
	files := append([]File(nil), m.Files...)
	sort.Slice(files, func(i, j int) bool { return files[i].Path < files[j].Path })
	var b strings.Builder
	b.WriteString(payloadMagic + "\n")
	b.WriteString(m.Version + "\n")
	b.WriteString(strconv.FormatUint(m.Build, 10) + "\n")
	b.WriteString(strconv.Itoa(len(files)) + "\n")
	seen := map[string]bool{}
	for _, f := range files {
		if seen[f.Path] {
			return "", fmt.Errorf("duplicate path %s", f.Path)
		}
		seen[f.Path] = true
		if err := validateRelativePath(f.Path); err != nil {
			return "", err
		}
		if err := validateToken("url", f.URL); err != nil {
			return "", err
		}
		if !isLowerHex(f.SHA256, 64) {
			return "", fmt.Errorf("sha256 must be 64 lowercase hex chars for %s", f.Path)
		}
		if f.Size < 0 {
			return "", fmt.Errorf("negative size for %s", f.Path)
		}
		b.WriteString(f.Path + "\n")
		b.WriteString(f.SHA256 + "\n")
		b.WriteString(strconv.FormatInt(f.Size, 10) + "\n")
		b.WriteString(f.URL + "\n")
	}
	return b.String(), nil
}

// SignEnvelope signs a canonical payload with an Ed25519 private key.
func SignEnvelope(payload string, key ed25519.PrivateKey) (string, error) {
	if !strings.HasPrefix(payload, payloadMagic+"\n") {
		return "", errors.New("payload is not a GORO1 document")
	}
	if len(key) != ed25519.PrivateKeySize {
		return "", errors.New("bad private key length")
	}
	sig := ed25519.Sign(key, []byte(payload))
	return envelopeMagic + "\n" + hex.EncodeToString(sig) + "\n" + payload, nil
}

// OpenEnvelope verifies the signature and parses the payload.
func OpenEnvelope(data []byte, pub ed25519.PublicKey) (Manifest, error) {
	if len(pub) != ed25519.PublicKeySize || isZero(pub) {
		return Manifest{}, errors.New("update public key is unset")
	}
	text := string(data)
	rest, ok := strings.CutPrefix(text, envelopeMagic+"\n")
	if !ok {
		return Manifest{}, errors.New("missing GORO1SIG header")
	}
	sigHex, payload, ok := strings.Cut(rest, "\n")
	if !ok {
		return Manifest{}, errors.New("truncated signature envelope")
	}
	if !isLowerHex(sigHex, 128) {
		return Manifest{}, errors.New("signature must be 128 lowercase hex chars")
	}
	if !strings.HasPrefix(payload, payloadMagic+"\n") {
		return Manifest{}, errors.New("signed payload must be canonical GORO1")
	}
	sig, _ := hex.DecodeString(sigHex)
	if !ed25519.Verify(pub, []byte(payload), sig) {
		return Manifest{}, errors.New("manifest signature rejected")
	}
	return ParsePayload(payload)
}

// ParsePayload parses a canonical payload without checking a signature.
func ParsePayload(payload string) (Manifest, error) {
	lines := strings.Split(payload, "\n")
	next := func() (string, error) {
		if len(lines) == 0 {
			return "", errors.New("truncated manifest")
		}
		l := lines[0]
		lines = lines[1:]
		return l, nil
	}
	if l, _ := next(); l != payloadMagic {
		return Manifest{}, errors.New("payload magic mismatch")
	}
	var m Manifest
	var err error
	if m.Version, err = next(); err != nil {
		return Manifest{}, err
	}
	bs, err := next()
	if err != nil {
		return Manifest{}, err
	}
	if m.Build, err = strconv.ParseUint(bs, 10, 64); err != nil {
		return Manifest{}, errors.New("bad build")
	}
	cs, err := next()
	if err != nil {
		return Manifest{}, err
	}
	count, err := strconv.Atoi(cs)
	if err != nil || count <= 0 || count > maxFiles {
		return Manifest{}, errors.New("file count out of range")
	}
	for i := 0; i < count; i++ {
		var f File
		var size string
		if f.Path, err = next(); err != nil {
			return Manifest{}, err
		}
		if f.SHA256, err = next(); err != nil {
			return Manifest{}, err
		}
		if size, err = next(); err != nil {
			return Manifest{}, err
		}
		if f.Size, err = strconv.ParseInt(size, 10, 64); err != nil {
			return Manifest{}, errors.New("bad size")
		}
		if f.URL, err = next(); err != nil {
			return Manifest{}, err
		}
		m.Files = append(m.Files, f)
	}
	if len(lines) != 1 || lines[0] != "" {
		return Manifest{}, errors.New("trailing manifest data")
	}
	canonical, err := CanonicalPayload(m)
	if err != nil {
		return Manifest{}, err
	}
	if canonical != payload {
		return Manifest{}, errors.New("payload is not canonical")
	}
	return m, nil
}

// SHA256Hex returns the lowercase hex digest of data.
func SHA256Hex(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func validateVersion(v string) error {
	// Must look like a name: no separators, and not only dots ("." / "..").
	if len(v) < 1 || len(v) > 64 || strings.Trim(v, ".") == "" || strings.HasPrefix(v, ".") {
		return fmt.Errorf("unsafe version %q", v)
	}
	for _, c := range v {
		if !(c >= '0' && c <= '9' || c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c == '.' || c == '_' || c == '+' || c == '-') {
			return fmt.Errorf("unsafe version %q", v)
		}
	}
	return nil
}

func validateToken(label, v string) error {
	if v == "" || len(v) > maxTokenLen {
		return fmt.Errorf("unsafe %s", label)
	}
	for _, c := range v {
		if c < 0x21 || c == 0x7f {
			return fmt.Errorf("unsafe %s", label)
		}
	}
	return nil
}

// validateRelativePath accepts slash-separated relative paths with no empty,
// "." or ".." components, no backslashes or drive letters, and nothing under
// the updater's own updates/ directory.
func validateRelativePath(p string) error {
	if err := validateToken("path", p); err != nil {
		return err
	}
	if strings.HasPrefix(p, "/") || strings.Contains(p, "\\") || strings.Contains(p, ":") {
		return fmt.Errorf("unsafe path %q", p)
	}
	for _, part := range strings.Split(p, "/") {
		if part == "" || part == "." || part == ".." {
			return fmt.Errorf("unsafe path %q", p)
		}
	}
	if p == "updates" || strings.HasPrefix(p, "updates/") {
		return fmt.Errorf("path %q is reserved", p)
	}
	return nil
}

func isLowerHex(s string, n int) bool {
	if len(s) != n {
		return false
	}
	for _, c := range s {
		if !(c >= '0' && c <= '9' || c >= 'a' && c <= 'f') {
			return false
		}
	}
	return true
}

func isZero(b []byte) bool {
	for _, x := range b {
		if x != 0 {
			return false
		}
	}
	return true
}
