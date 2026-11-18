package updater

import (
	"crypto/ed25519"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"
)

const (
	dataPayloadMagic  = "GDAT1"
	dataEnvelopeMagic = "GDAT1SIG"
	maxDataFiles      = 1_000_000
	maxDataPathLen    = 512
)

// DataFile is one served asset listed in a signed data manifest.
type DataFile struct {
	Path   string
	SHA256 string
	Size   int64
}

// DataManifest is a verified listing of every file in one data version.
type DataManifest struct {
	Version  uint64
	Files    []DataFile
	NextKeys []ed25519.PublicKey
	byPath   map[string]DataFile
}

// CanonicalDataPayload returns the exact bytes that are signed for a data
// version. Files are sorted by path bytes; next-keys are optional.
func CanonicalDataPayload(m DataManifest) (string, error) {
	if m.Version == 0 {
		return "", errors.New("data version must be a positive monotonic counter")
	}
	if len(m.Files) == 0 {
		return "", errors.New("data manifest has no files")
	}
	if len(m.Files) > maxDataFiles {
		return "", errors.New("too many files")
	}
	files := append([]DataFile(nil), m.Files...)
	sort.Slice(files, func(i, j int) bool { return files[i].Path < files[j].Path })
	var b strings.Builder
	b.WriteString(dataPayloadMagic + "\n")
	b.WriteString(strconv.FormatUint(m.Version, 10) + "\n")
	b.WriteString(strconv.Itoa(len(files)) + "\n")
	seen := map[string]bool{}
	for _, f := range files {
		if seen[f.Path] {
			return "", fmt.Errorf("duplicate path %s", f.Path)
		}
		seen[f.Path] = true
		if err := validateDataPath(f.Path); err != nil {
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
	}
	if err := writeNextKeys(&b, m.NextKeys); err != nil {
		return "", err
	}
	return b.String(), nil
}

// SignDataEnvelope signs a canonical GDAT1 payload with an Ed25519 private key.
func SignDataEnvelope(payload string, key ed25519.PrivateKey) (string, error) {
	if !strings.HasPrefix(payload, dataPayloadMagic+"\n") {
		return "", errors.New("payload is not a GDAT1 document")
	}
	if len(key) != ed25519.PrivateKeySize {
		return "", errors.New("bad private key length")
	}
	sig := ed25519.Sign(key, []byte(payload))
	return dataEnvelopeMagic + "\n" + hex.EncodeToString(sig) + "\n" + payload, nil
}

// OpenDataEnvelope verifies a GDAT1 envelope with a single public key.
func OpenDataEnvelope(data []byte, pub ed25519.PublicKey) (DataManifest, error) {
	return OpenDataEnvelopeWithKeys(data, []ed25519.PublicKey{pub})
}

// OpenDataEnvelopeWithKeys verifies a GDAT1 envelope against any trusted key.
func OpenDataEnvelopeWithKeys(data []byte, keys []ed25519.PublicKey) (DataManifest, error) {
	payload, err := openSignedPayload(data, dataEnvelopeMagic, dataPayloadMagic, keys)
	if err != nil {
		return DataManifest{}, err
	}
	return ParseDataPayload(payload)
}

// ParseDataPayload parses a canonical GDAT1 payload without checking a signature.
func ParseDataPayload(payload string) (DataManifest, error) {
	lines := strings.Split(payload, "\n")
	next := func() (string, error) {
		if len(lines) == 0 {
			return "", errors.New("truncated data manifest")
		}
		l := lines[0]
		lines = lines[1:]
		return l, nil
	}
	if l, _ := next(); l != dataPayloadMagic {
		return DataManifest{}, errors.New("payload magic mismatch")
	}
	var m DataManifest
	vs, err := next()
	if err != nil {
		return DataManifest{}, err
	}
	if m.Version, err = strconv.ParseUint(vs, 10, 64); err != nil {
		return DataManifest{}, errors.New("bad data version")
	}
	cs, err := next()
	if err != nil {
		return DataManifest{}, err
	}
	count, err := strconv.Atoi(cs)
	if err != nil || count <= 0 || count > maxDataFiles {
		return DataManifest{}, errors.New("file count out of range")
	}
	for i := 0; i < count; i++ {
		var f DataFile
		var size string
		if f.Path, err = next(); err != nil {
			return DataManifest{}, err
		}
		if f.SHA256, err = next(); err != nil {
			return DataManifest{}, err
		}
		if size, err = next(); err != nil {
			return DataManifest{}, err
		}
		if f.Size, err = strconv.ParseInt(size, 10, 64); err != nil {
			return DataManifest{}, errors.New("bad size")
		}
		m.Files = append(m.Files, f)
	}
	m.NextKeys, err = parseNextKeys(&lines, next)
	if err != nil {
		return DataManifest{}, err
	}
	if len(lines) != 1 || lines[0] != "" {
		return DataManifest{}, errors.New("trailing data manifest")
	}
	canonical, err := CanonicalDataPayload(m)
	if err != nil {
		return DataManifest{}, err
	}
	if canonical != payload {
		return DataManifest{}, errors.New("payload is not canonical")
	}
	m.byPath = make(map[string]DataFile, len(m.Files))
	for _, f := range m.Files {
		m.byPath[f.Path] = f
	}
	return m, nil
}

// RejectRollback reports an error when this manifest's version is older than
// the version the client already accepted.
func (m DataManifest) RejectRollback(accepted uint64) error {
	if accepted > 0 && m.Version < accepted {
		return fmt.Errorf("data version %d is older than accepted %d", m.Version, accepted)
	}
	return nil
}

// VerifyFile checks a streamed body against the signed digest. A mismatch
// must be discarded and never cached or used. Unknown paths are rejected
// when a manifest is in force.
func (m DataManifest) VerifyFile(path string, data []byte) error {
	if m.byPath == nil {
		m.byPath = make(map[string]DataFile, len(m.Files))
		for _, f := range m.Files {
			m.byPath[f.Path] = f
		}
	}
	f, ok := m.byPath[path]
	if !ok {
		return fmt.Errorf("%s is not in the signed data manifest", path)
	}
	if int64(len(data)) != f.Size {
		return fmt.Errorf("%s size mismatch: got %d want %d", path, len(data), f.Size)
	}
	if got := SHA256Hex(data); got != f.SHA256 {
		return fmt.Errorf("%s digest mismatch", path)
	}
	return nil
}

// File returns the signed entry for path, if any.
func (m DataManifest) File(path string) (DataFile, bool) {
	if m.byPath == nil {
		return DataFile{}, false
	}
	f, ok := m.byPath[path]
	return f, ok
}

// validateDataPath accepts slash-separated relative paths with spaces (RO
// tables live under "lua files"), no empty/"."/".." components, and no
// backslashes or drive letters.
func validateDataPath(p string) error {
	if p == "" || len(p) > maxDataPathLen || !utf8.ValidString(p) {
		return fmt.Errorf("unsafe path %q", p)
	}
	if strings.HasPrefix(p, "/") || strings.Contains(p, "\\") || strings.Contains(p, ":") {
		return fmt.Errorf("unsafe path %q", p)
	}
	for _, c := range p {
		if c < 0x20 || c == 0x7f {
			return fmt.Errorf("unsafe path %q", p)
		}
	}
	for _, part := range strings.Split(p, "/") {
		if part == "" || part == "." || part == ".." {
			return fmt.Errorf("unsafe path %q", p)
		}
	}
	return nil
}
