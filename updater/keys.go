package updater

import (
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"sync"
)

// KeyRing is the client's list of trusted Ed25519 public keys: the compiled-in
// set plus extras accepted from a verified next-key list. It never holds a
// private key.
type KeyRing struct {
	mu       sync.Mutex
	embedded []ed25519.PublicKey
	extra    []ed25519.PublicKey
	path     string
}

type storedPublicKeys struct {
	PublicKeys []string `json:"public_keys"`
}

// NewKeyRing loads extras from path (if it exists) and starts with embedded.
// path may be empty, in which case extras are not persisted.
func NewKeyRing(path string, embedded []ed25519.PublicKey) *KeyRing {
	k := &KeyRing{path: path}
	if len(embedded) == 0 {
		embedded = TrustedPublicKeys
	}
	k.embedded = uniqueKeys(embedded)
	if path != "" {
		k.extra = loadExtraKeys(path)
	}
	return k
}

// Keys returns the current trusted list (embedded first, then extras).
func (k *KeyRing) Keys() []ed25519.PublicKey {
	if k == nil {
		return append([]ed25519.PublicKey(nil), TrustedPublicKeys...)
	}
	k.mu.Lock()
	defer k.mu.Unlock()
	return uniqueKeys(append(append([]ed25519.PublicKey(nil), k.embedded...), k.extra...))
}

// AcceptNextKeys unions verified public keys into the ring and persists extras
// that are not already compiled in. Empty input is a no-op.
func (k *KeyRing) AcceptNextKeys(next []ed25519.PublicKey) error {
	if k == nil || len(next) == 0 {
		return nil
	}
	cleaned := uniqueKeys(next)
	if len(cleaned) == 0 {
		return errors.New("no usable next-keys")
	}
	k.mu.Lock()
	defer k.mu.Unlock()
	k.extra = uniqueKeys(append(k.extra, cleaned...))
	if k.path == "" {
		return nil
	}
	return persistExtraKeys(k.path, k.extra, k.embedded)
}

func uniqueKeys(in []ed25519.PublicKey) []ed25519.PublicKey {
	seen := map[string]bool{}
	var out []ed25519.PublicKey
	var hexes []string
	index := map[string]ed25519.PublicKey{}
	for _, k := range in {
		if len(k) != ed25519.PublicKeySize || isZero(k) {
			continue
		}
		h := hex.EncodeToString(k)
		if seen[h] {
			continue
		}
		seen[h] = true
		hexes = append(hexes, h)
		index[h] = append(ed25519.PublicKey(nil), k...)
	}
	sort.Strings(hexes)
	for _, h := range hexes {
		out = append(out, index[h])
	}
	return out
}

func loadExtraKeys(path string) []ed25519.PublicKey {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var stored storedPublicKeys
	if json.Unmarshal(raw, &stored) != nil {
		return nil
	}
	var keys []ed25519.PublicKey
	for _, h := range stored.PublicKeys {
		k, err := parsePublicKeyHex(h)
		if err != nil {
			continue
		}
		keys = append(keys, k)
	}
	return uniqueKeys(keys)
}

func persistExtraKeys(path string, extra, embedded []ed25519.PublicKey) error {
	emb := map[string]bool{}
	for _, k := range embedded {
		if len(k) == ed25519.PublicKeySize {
			emb[hex.EncodeToString(k)] = true
		}
	}
	var hexes []string
	for _, k := range uniqueKeys(extra) {
		h := hex.EncodeToString(k)
		if emb[h] {
			continue
		}
		hexes = append(hexes, h)
	}
	sort.Strings(hexes)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	raw, err := json.MarshalIndent(storedPublicKeys{PublicKeys: hexes}, "", "  ")
	if err != nil {
		return err
	}
	raw = append(raw, '\n')
	return writeAtomic(path, raw)
}
