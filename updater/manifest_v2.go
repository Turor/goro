package updater

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"
)

const (
	dataRootMagic  = "GDAT2ROOT"
	dataShardMagic = "GDAT2SHARD"
	shardSlots     = 256
)

// ShardRef is one shard's digest in a GDAT2 root. ID is two lowercase hex
// digits. SHA256 and Size cover the shard's canonical UTF-8 bytes.
type ShardRef struct {
	ID     string
	SHA256 string
	Size   int64
}

// DataRoot is a verified GDAT2 root. The signature is detached (root.sig);
// it is not part of this document.
type DataRoot struct {
	Version         uint64
	PreviousVersion uint64
	CreatedUnix     int64
	Shards          [shardSlots]ShardRef
	NextKeys        []ed25519.PublicKey
}

// DataShard is one GDAT2 shard: the files whose normalized path hashes into
// that shard id.
type DataShard struct {
	Version uint64
	ID      string
	Files   []DataFile
	byPath  map[string]DataFile
}

// ShardID is the first byte of SHA-256(UTF-8 of the normalized path), as two
// lowercase hex digits. normalized must already be a manifest key.
func ShardID(normalized string) string {
	sum := sha256.Sum256([]byte(normalized))
	return hex.EncodeToString(sum[:1])
}

// NormalizeManifestPath applies the shared manifest key rule. The bool is
// false when the path must not appear in a manifest.
func NormalizeManifestPath(input string) (string, bool) {
	if !utf8.ValidString(input) {
		return "", false
	}
	for _, r := range input {
		if r < 0x20 || r == 0x7f {
			return "", false
		}
	}
	p := strings.ToLower(input)
	p = strings.ReplaceAll(p, `\`, "/")
	if strings.HasPrefix(p, "/") {
		p = p[1:]
	}
	for strings.Contains(p, "//") {
		p = strings.ReplaceAll(p, "//", "/")
	}
	for strings.HasPrefix(p, "./") {
		p = p[2:]
	}
	if strings.HasSuffix(p, "/") {
		p = p[:len(p)-1]
	}
	if p == "data" {
		p = ""
	} else if strings.HasPrefix(p, "data/") {
		p = p[len("data/"):]
	}
	if p == "" || len(p) > maxDataPathLen || strings.Contains(p, `\`) || strings.Contains(p, ":") {
		return "", false
	}
	for _, part := range strings.Split(p, "/") {
		if part == "" || part == "." || part == ".." {
			return "", false
		}
	}
	if err := validateDataPath(p); err != nil {
		return "", false
	}
	return p, true
}

// DecodeThenNormalize percent-decodes a URL path when it contains '%' and
// then runs NormalizeManifestPath. Inputs with no '%' are not decoded.
func DecodeThenNormalize(input string) (string, bool) {
	if strings.Contains(input, "%") {
		decoded, ok := decodePathSegments(input)
		if !ok {
			return "", false
		}
		input = decoded
	}
	return NormalizeManifestPath(input)
}

func decodePathSegments(s string) (string, bool) {
	var b strings.Builder
	b.Grow(len(s))
	start := 0
	for i := 0; i <= len(s); i++ {
		if i == len(s) || s[i] == '/' || s[i] == '\\' {
			dec, err := url.PathUnescape(s[start:i])
			if err != nil {
				return "", false
			}
			b.WriteString(dec)
			if i < len(s) {
				b.WriteByte(s[i])
			}
			start = i + 1
		}
	}
	return b.String(), true
}

// BuildDataTree normalizes paths, groups them into 256 shards, and returns
// the canonical root plus each shard's canonical bytes (keyed by shard id).
func BuildDataTree(version, previous uint64, created int64, files []DataFile, next []ed25519.PublicKey) (string, map[string]string, error) {
	if version == 0 {
		return "", nil, errors.New("data version must be a positive monotonic counter")
	}
	grouped := map[string][]DataFile{}
	seen := map[string]bool{}
	for _, f := range files {
		key, ok := NormalizeManifestPath(f.Path)
		if !ok {
			return "", nil, fmt.Errorf("unsafe path %q", f.Path)
		}
		if seen[key] {
			return "", nil, fmt.Errorf("duplicate path %s", key)
		}
		seen[key] = true
		if !isLowerHex(f.SHA256, 64) || f.Size < 0 {
			return "", nil, fmt.Errorf("bad digest for %s", key)
		}
		id := ShardID(key)
		grouped[id] = append(grouped[id], DataFile{Path: key, SHA256: f.SHA256, Size: f.Size})
	}
	shards := make(map[string]string, shardSlots)
	root := DataRoot{Version: version, PreviousVersion: previous, CreatedUnix: created, NextKeys: next}
	for i := 0; i < shardSlots; i++ {
		id := fmt.Sprintf("%02x", i)
		body, err := CanonicalDataShard(DataShard{Version: version, ID: id, Files: grouped[id]})
		if err != nil {
			return "", nil, err
		}
		shards[id] = body
		sum := sha256.Sum256([]byte(body))
		root.Shards[i] = ShardRef{ID: id, SHA256: hex.EncodeToString(sum[:]), Size: int64(len(body))}
	}
	canonical, err := CanonicalDataRoot(root)
	if err != nil {
		return "", nil, err
	}
	return canonical, shards, nil
}

// CanonicalDataRoot returns the exact bytes that are signed for a GDAT2 root.
func CanonicalDataRoot(r DataRoot) (string, error) {
	if r.Version == 0 {
		return "", errors.New("data version must be a positive monotonic counter")
	}
	if r.CreatedUnix < 0 {
		return "", errors.New("bad created time")
	}
	var b strings.Builder
	b.WriteString(dataRootMagic + "\n")
	b.WriteString(strconv.FormatUint(r.Version, 10) + "\n")
	b.WriteString(strconv.FormatUint(r.PreviousVersion, 10) + "\n")
	b.WriteString(strconv.FormatInt(r.CreatedUnix, 10) + "\n")
	b.WriteString(strconv.Itoa(shardSlots) + "\n")
	for i := 0; i < shardSlots; i++ {
		id := fmt.Sprintf("%02x", i)
		sh := r.Shards[i]
		if sh.ID != id {
			return "", fmt.Errorf("shard %s missing or out of order", id)
		}
		if !isLowerHex(sh.SHA256, 64) {
			return "", fmt.Errorf("bad shard digest %s", id)
		}
		if sh.Size < 0 {
			return "", fmt.Errorf("bad shard size %s", id)
		}
		b.WriteString(id + " " + sh.SHA256 + " " + strconv.FormatInt(sh.Size, 10) + "\n")
	}
	if err := writeNextKeys(&b, r.NextKeys); err != nil {
		return "", err
	}
	return b.String(), nil
}

// CanonicalDataShard returns the exact bytes hashed by the root for one shard.
func CanonicalDataShard(s DataShard) (string, error) {
	if s.Version == 0 {
		return "", errors.New("data version must be a positive monotonic counter")
	}
	if !isLowerHex(s.ID, 2) {
		return "", errors.New("bad shard id")
	}
	files := append([]DataFile(nil), s.Files...)
	sort.Slice(files, func(i, j int) bool { return files[i].Path < files[j].Path })
	if len(files) > maxDataFiles {
		return "", errors.New("too many files")
	}
	var b strings.Builder
	b.WriteString(dataShardMagic + "\n")
	b.WriteString(strconv.FormatUint(s.Version, 10) + "\n")
	b.WriteString(s.ID + "\n")
	b.WriteString(strconv.Itoa(len(files)) + "\n")
	seen := map[string]bool{}
	for _, f := range files {
		if seen[f.Path] {
			return "", fmt.Errorf("duplicate path %s", f.Path)
		}
		seen[f.Path] = true
		if err := validManifestKey(f.Path); err != nil {
			return "", err
		}
		if ShardID(f.Path) != s.ID {
			return "", fmt.Errorf("path %s belongs in shard %s", f.Path, ShardID(f.Path))
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
	return b.String(), nil
}

// validManifestKey checks a path that is already the output of one
// NormalizeManifestPath call. It does not strip a data/ prefix again, because
// data/foo is a legal key for a source path of data/data/foo.
func validManifestKey(p string) error {
	if err := validateDataPath(p); err != nil {
		return err
	}
	if p != strings.ToLower(p) || strings.HasPrefix(p, "/") || strings.HasSuffix(p, "/") || strings.Contains(p, "//") || strings.HasPrefix(p, "./") {
		return fmt.Errorf("path %q is not a manifest key", p)
	}
	return nil
}

// SignDataRoot signs a canonical GDAT2 root. The returned signature is 128
// lowercase hex characters and does not include a trailing newline.
func SignDataRoot(canonical string, key ed25519.PrivateKey) (string, error) {
	if _, err := ParseDataRoot(canonical); err != nil {
		return "", err
	}
	if len(key) != ed25519.PrivateKeySize {
		return "", errors.New("bad private key length")
	}
	return hex.EncodeToString(ed25519.Sign(key, []byte(canonical))), nil
}

// ParseDetachedSignature accepts 128 lowercase hex characters, with or
// without one trailing newline.
func ParseDetachedSignature(raw []byte) ([]byte, error) {
	s := string(raw)
	if strings.HasSuffix(s, "\n") {
		s = strings.TrimSuffix(s, "\n")
		s = strings.TrimSuffix(s, "\r")
	}
	if strings.ContainsAny(s, "\r\n \t") || !isLowerHex(s, 128) {
		return nil, errors.New("signature must be 128 lowercase hex chars")
	}
	sig, err := hex.DecodeString(s)
	if err != nil || len(sig) != ed25519.SignatureSize {
		return nil, errors.New("signature must be 128 lowercase hex chars")
	}
	return sig, nil
}

// OpenDataRoot verifies a detached Ed25519 signature over the exact root
// bytes, then parses the canonical document. The first trusted key that
// verifies is accepted.
func OpenDataRoot(root, sig []byte, keys []ed25519.PublicKey) (DataRoot, error) {
	rawSig, err := ParseDetachedSignature(sig)
	if err != nil {
		return DataRoot{}, err
	}
	if len(keys) == 0 {
		return DataRoot{}, errors.New("no trusted public keys")
	}
	ok := false
	for _, pub := range keys {
		if len(pub) != ed25519.PublicKeySize || isZero(pub) {
			continue
		}
		if ed25519.Verify(pub, root, rawSig) {
			ok = true
			break
		}
	}
	if !ok {
		return DataRoot{}, errors.New("manifest signature rejected")
	}
	return ParseDataRoot(string(root))
}

// ParseDataRoot parses a canonical GDAT2 root without checking a signature.
func ParseDataRoot(payload string) (DataRoot, error) {
	lines, err := canonicalLines(payload)
	if err != nil {
		return DataRoot{}, err
	}
	next := lineReader(&lines)
	if l, _ := next(); l != dataRootMagic {
		return DataRoot{}, errors.New("payload magic mismatch")
	}
	var r DataRoot
	vs, err := next()
	if err != nil {
		return DataRoot{}, err
	}
	if r.Version, err = parseCanonicalUint(vs); err != nil || r.Version == 0 {
		return DataRoot{}, errors.New("bad data version")
	}
	ps, err := next()
	if err != nil {
		return DataRoot{}, err
	}
	if r.PreviousVersion, err = parseCanonicalUint(ps); err != nil {
		return DataRoot{}, errors.New("bad previous version")
	}
	cs, err := next()
	if err != nil {
		return DataRoot{}, err
	}
	created, err := parseCanonicalUint(cs)
	if err != nil || created > uint64(^uint64(0)>>1) {
		return DataRoot{}, errors.New("bad created time")
	}
	r.CreatedUnix = int64(created)
	ns, err := next()
	if err != nil || ns != strconv.Itoa(shardSlots) {
		return DataRoot{}, errors.New("root must list 256 shards")
	}
	for i := 0; i < shardSlots; i++ {
		line, err := next()
		if err != nil {
			return DataRoot{}, err
		}
		id := fmt.Sprintf("%02x", i)
		fields := strings.Split(line, " ")
		if len(fields) != 3 || fields[0] != id || !isLowerHex(fields[1], 64) {
			return DataRoot{}, fmt.Errorf("bad shard line %s", id)
		}
		size, err := parseCanonicalUint(fields[2])
		if err != nil || size > uint64(^uint64(0)>>1) {
			return DataRoot{}, fmt.Errorf("bad shard size %s", id)
		}
		r.Shards[i] = ShardRef{ID: id, SHA256: fields[1], Size: int64(size)}
	}
	r.NextKeys, err = parseNextKeys(&lines, next)
	if err != nil {
		return DataRoot{}, err
	}
	if len(lines) != 1 || lines[0] != "" {
		return DataRoot{}, errors.New("trailing data manifest")
	}
	canonical, err := CanonicalDataRoot(r)
	if err != nil {
		return DataRoot{}, err
	}
	if canonical != payload {
		return DataRoot{}, errors.New("payload is not canonical")
	}
	return r, nil
}

// ParseDataShard parses a canonical GDAT2 shard without checking the root digest.
func ParseDataShard(payload string) (DataShard, error) {
	lines, err := canonicalLines(payload)
	if err != nil {
		return DataShard{}, err
	}
	next := lineReader(&lines)
	if l, _ := next(); l != dataShardMagic {
		return DataShard{}, errors.New("payload magic mismatch")
	}
	var s DataShard
	vs, err := next()
	if err != nil {
		return DataShard{}, err
	}
	if s.Version, err = parseCanonicalUint(vs); err != nil || s.Version == 0 {
		return DataShard{}, errors.New("bad data version")
	}
	if s.ID, err = next(); err != nil || !isLowerHex(s.ID, 2) {
		return DataShard{}, errors.New("bad shard id")
	}
	cs, err := next()
	if err != nil {
		return DataShard{}, err
	}
	count, err := parseCanonicalUint(cs)
	if err != nil || count > maxDataFiles {
		return DataShard{}, errors.New("file count out of range")
	}
	for i := uint64(0); i < count; i++ {
		var f DataFile
		var size string
		if f.Path, err = next(); err != nil {
			return DataShard{}, err
		}
		if f.SHA256, err = next(); err != nil {
			return DataShard{}, err
		}
		if size, err = next(); err != nil {
			return DataShard{}, err
		}
		n, err := parseCanonicalUint(size)
		if err != nil || n > uint64(^uint64(0)>>1) {
			return DataShard{}, errors.New("bad size")
		}
		f.Size = int64(n)
		s.Files = append(s.Files, f)
	}
	if len(lines) != 1 || lines[0] != "" {
		return DataShard{}, errors.New("trailing data manifest")
	}
	canonical, err := CanonicalDataShard(s)
	if err != nil {
		return DataShard{}, err
	}
	if canonical != payload {
		return DataShard{}, errors.New("payload is not canonical")
	}
	s.byPath = make(map[string]DataFile, len(s.Files))
	for _, f := range s.Files {
		s.byPath[f.Path] = f
	}
	return s, nil
}

// File returns the shard entry for a normalized path.
func (s DataShard) File(path string) (DataFile, bool) {
	if s.byPath == nil {
		s.byPath = make(map[string]DataFile, len(s.Files))
		for _, f := range s.Files {
			s.byPath[f.Path] = f
		}
	}
	f, ok := s.byPath[path]
	return f, ok
}

// Shard returns the root's digest line for a two-hex shard id.
func (r DataRoot) Shard(id string) (ShardRef, bool) {
	n, err := strconv.ParseUint(id, 16, 8)
	if err != nil || fmt.Sprintf("%02x", n) != id {
		return ShardRef{}, false
	}
	ref := r.Shards[n]
	if ref.ID != id {
		return ShardRef{}, false
	}
	return ref, true
}

// RejectRollback reports an error when this root is older than the version
// the client already accepted. Equal versions and skipped versions are allowed.
func (r DataRoot) RejectRollback(accepted uint64) error {
	if accepted > 0 && r.Version < accepted {
		return fmt.Errorf("data version %d is older than accepted %d", r.Version, accepted)
	}
	return nil
}

func canonicalLines(payload string) ([]string, error) {
	if payload == "" || !strings.HasSuffix(payload, "\n") || strings.Contains(payload, "\r") {
		return nil, errors.New("truncated data manifest")
	}
	return strings.Split(payload, "\n"), nil
}

func lineReader(lines *[]string) func() (string, error) {
	return func() (string, error) {
		if len(*lines) == 0 {
			return "", errors.New("truncated data manifest")
		}
		l := (*lines)[0]
		*lines = (*lines)[1:]
		return l, nil
	}
}

func parseCanonicalUint(s string) (uint64, error) {
	if s == "" || (len(s) > 1 && s[0] == '0') {
		return 0, errors.New("bad integer")
	}
	n, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return 0, errors.New("bad integer")
	}
	return n, nil
}
