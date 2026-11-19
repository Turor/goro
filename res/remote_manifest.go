package res

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/kivutar/goro/glog"
	"github.com/kivutar/goro/updater"
)

const (
	manifestModeReport  = "report"
	manifestModeEnforce = "enforce"
	dataRootMax         = 1 << 20
	dataSigMax          = 256
	dataShardMax        = 32 << 20
)

// signedRoot is a signature-checked GDAT2 root kept in memory.
type signedRoot struct {
	version uint64
	raw     []byte
	sig     []byte
	parsed  updater.DataRoot
}

type cachedShard struct {
	byPath map[string]updater.DataFile
}

type shardCall struct {
	done  chan struct{}
	shard *cachedShard
	err   error
}

func normalizeManifestMode(mode string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(mode)) {
	case "", manifestModeReport:
		return manifestModeReport, nil
	case manifestModeEnforce:
		return manifestModeEnforce, nil
	default:
		return "", fmt.Errorf("data manifest mode must be report or enforce, not %q", mode)
	}
}

func (r *remoteSource) enforcing() bool {
	return r.manifestMode == manifestModeEnforce
}

func (r *remoteSource) trusted() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.manifestTrusted
}

func (r *remoteSource) manifestDir() string {
	return filepath.Join(r.cacheDir, "manifest-v2")
}

// loadPersistedRoot re-verifies the accepted root on every start. A version
// cursor alone is not trust: the signature is checked again before any file
// is treated as covered.
func (r *remoteSource) loadPersistedRoot() {
	dir := r.manifestDir()
	root, errR := os.ReadFile(filepath.Join(dir, "root"))
	sig, errS := os.ReadFile(filepath.Join(dir, "root.sig"))
	verRaw, errV := os.ReadFile(filepath.Join(dir, "version"))
	var fileVer uint64
	if errV == nil {
		fileVer, _ = strconv.ParseUint(strings.TrimSpace(string(verRaw)), 10, 64)
	}
	if errR == nil && errS == nil {
		parsed, err := updater.OpenDataRoot(root, sig, r.keyRing.Keys())
		if err == nil {
			if err := r.keyRing.AcceptNextKeys(parsed.NextKeys); err != nil {
				glog.Warnf("remote data: next-keys not persisted: %v", err)
			}
			r.installRoot(root, sig, parsed)
			if fileVer != parsed.Version {
				if err := r.persistAccepted(root, sig, parsed.Version); err != nil {
					r.logOnce("manifest-persist", "remote data: cannot persist data manifest: %v", err)
				}
			}
			return
		}
		r.logOnce("manifest-persist", "remote data: persisted data manifest failed verification: %v", err)
	}
	// A version cursor without a verifiable root still means this host was
	// trusted before. Do not install the bad root, and do not start over at 0.
	if fileVer > 0 {
		r.mu.Lock()
		r.manifestTrusted = true
		if fileVer > r.acceptedManifest {
			r.acceptedManifest = fileVer
		}
		r.mu.Unlock()
		r.noteManifestAttack()
		if errR != nil || errS != nil {
			r.logOnce("manifest-persist", "remote data: trusted data manifest is missing from the cache")
		}
	}
}

func (r *remoteSource) installRoot(raw, sig []byte, parsed updater.DataRoot) {
	prev := r.acceptedManifestVersion()
	r.manifest.Store(&signedRoot{
		version: parsed.Version,
		raw:     append([]byte(nil), raw...),
		sig:     append([]byte(nil), sig...),
		parsed:  parsed,
	})
	r.mu.Lock()
	advanced := parsed.Version > r.acceptedManifest
	r.acceptedManifest = parsed.Version
	r.manifestTrusted = true
	r.mu.Unlock()
	if advanced {
		r.shardMu.Lock()
		r.shards = map[string]*cachedShard{}
		r.shardMu.Unlock()
		// The first accepted root, and every newer one, must re-check bodies
		// already in the cache. Reloading the same version does not.
		if parsed.Version > prev {
			r.markAllStale()
		}
		glog.Infof("remote data: accepted signed manifest version %d", parsed.Version)
	}
}

func (r *remoteSource) persistAccepted(raw, sig []byte, version uint64) error {
	dir := r.manifestDir()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	if err := writeFileAtomic(filepath.Join(dir, "root"), raw); err != nil {
		return err
	}
	sigFile := sig
	if len(sigFile) == 0 || sigFile[len(sigFile)-1] != '\n' {
		sigFile = append(append([]byte(nil), sig...), '\n')
	}
	if err := writeFileAtomic(filepath.Join(dir, "root.sig"), sigFile); err != nil {
		return err
	}
	return writeFileAtomic(filepath.Join(dir, "version"), []byte(strconv.FormatUint(version, 10)+"\n"))
}

func (r *remoteSource) manifestLoop() {
	r.loadDataManifest()
	for {
		select {
		case <-r.stopCh:
			return
		case <-r.manifestWake:
		}
		r.loadDataManifest()
	}
}

func (r *remoteSource) wakeManifest() {
	select {
	case r.manifestWake <- struct{}{}:
	default:
	}
}

// considerManifest refetches the signed root when the live version moves
// past the one already attempted. It does not apply unsigned changes.
func (r *remoteSource) considerManifest(v uint64) {
	if v == 0 {
		return
	}
	r.mu.Lock()
	if v <= r.manifestTried || v <= r.acceptedManifest {
		r.mu.Unlock()
		return
	}
	r.manifestTried = v
	r.mu.Unlock()
	r.wakeManifest()
}

// loadDataManifest fetches the current signed root. A failed fetch leaves a
// previously accepted root in place and returns false so an unsigned changes
// cursor is not treated as covered.
func (r *remoteSource) loadDataManifest() bool {
	r.manifestMu.Lock()
	defer r.manifestMu.Unlock()
	ok := r.fetchAndAcceptRoot()
	r.mu.Lock()
	if seen := r.seenVersion.Load(); seen > r.manifestTried {
		r.manifestTried = seen
	}
	r.mu.Unlock()
	return ok
}

func (r *remoteSource) fetchAndAcceptRoot() bool {
	rootBody, err := r.getMeta("data-manifest/root", dataRootMax)
	if err != nil {
		r.manifestUnavailable(err)
		return false
	}
	sigBody, err := r.getMeta("data-manifest/root.sig", dataSigMax)
	if err != nil {
		r.manifestUnavailable(err)
		return false
	}
	parsed, err := updater.OpenDataRoot(rootBody, sigBody, r.keyRing.Keys())
	if err != nil {
		r.logOnce("data-manifest", "remote data: signed data manifest rejected: %v", err)
		r.noteManifestAttack()
		return false
	}
	accepted := r.acceptedManifestVersion()
	if err := parsed.RejectRollback(accepted); err != nil {
		r.logOnce("data-manifest", "remote data: %v", err)
		r.noteManifestAttack()
		return false
	}
	if err := r.keyRing.AcceptNextKeys(parsed.NextKeys); err != nil {
		glog.Warnf("remote data: next-keys not persisted: %v", err)
	}
	if err := r.persistAccepted(rootBody, sigBody, parsed.Version); err != nil {
		r.logOnce("manifest-persist", "remote data: cannot persist data manifest: %v", err)
	}
	r.installRoot(rootBody, sigBody, parsed)
	return true
}

func (r *remoteSource) manifestUnavailable(err error) {
	if r.trusted() {
		r.logOnce("manifest-missing", "remote data: previously trusted data manifest unavailable (%v); not skipping checks", err)
		r.noteManifestAttack()
		return
	}
	r.logOnce("manifest-missing", "remote data: signed data manifest unavailable: %v", err)
}

func (r *remoteSource) noteManifestAttack() {
	r.mu.Lock()
	r.manifestAttack = true
	r.mu.Unlock()
}

// verifyDigest checks a streamed body against the accepted manifest. In
// report mode a mismatch is logged and the body is still returned. In
// enforce mode the body must not be cached or used.
func (r *remoteSource) verifyDigest(p string, body []byte) error {
	if err := r.checkFile(p, body, false); err != nil {
		if !r.enforcing() {
			r.logOnce("digest:"+p, "remote data: %v", err)
			return nil
		}
		return err
	}
	return nil
}

// verifyLocal checks a loose or archive body when its normalized path is
// listed. Unlisted local files stay usable. Report mode logs and does not
// block.
func (r *remoteSource) verifyLocal(name string, body []byte) error {
	if err := r.checkFile(name, body, true); err != nil {
		if !r.enforcing() {
			r.logOnce("digest-local:"+name, "remote data: %v", err)
			return nil
		}
		return err
	}
	return nil
}

func (m *Manager) verifyLocalData(name string, data []byte) error {
	if m == nil || m.remote == nil {
		return nil
	}
	return m.remote.verifyLocal(name, data)
}

func (r *remoteSource) checkFile(name string, body []byte, local bool) error {
	key, ok := updater.DecodeThenNormalize(name)
	if !ok {
		if local {
			return nil
		}
		return fmt.Errorf("unsafe data path %q", name)
	}
	root := r.manifest.Load()
	if root == nil {
		if !local && r.enforcing() && r.trusted() {
			return errors.New("signed data manifest is not available")
		}
		return nil
	}
	entry, listed, err := r.lookup(root, key)
	if err != nil {
		return err
	}
	if !listed {
		if local {
			return nil
		}
		return fmt.Errorf("%s is not in the signed data manifest", key)
	}
	if int64(len(body)) != entry.Size || updater.SHA256Hex(body) != entry.SHA256 {
		return fmt.Errorf("%s digest mismatch", key)
	}
	return nil
}

func (r *remoteSource) lookup(root *signedRoot, key string) (updater.DataFile, bool, error) {
	id := updater.ShardID(key)
	ref, ok := root.parsed.Shard(id)
	if !ok {
		return updater.DataFile{}, false, fmt.Errorf("shard %s is not in the signed root", id)
	}
	cacheKey := strconv.FormatUint(root.version, 10) + "/" + id
	r.shardMu.Lock()
	if s, ok := r.shards[cacheKey]; ok {
		r.shardMu.Unlock()
		f, listed := s.byPath[key]
		return f, listed, nil
	}
	if call, ok := r.shardFlight[cacheKey]; ok {
		r.shardMu.Unlock()
		<-call.done
		if call.err != nil {
			return updater.DataFile{}, false, call.err
		}
		f, listed := call.shard.byPath[key]
		return f, listed, nil
	}
	call := &shardCall{done: make(chan struct{})}
	r.shardFlight[cacheKey] = call
	r.shardMu.Unlock()

	call.shard, call.err = r.loadShard(root, ref)
	r.shardMu.Lock()
	delete(r.shardFlight, cacheKey)
	if call.err == nil {
		if r.shards == nil {
			r.shards = map[string]*cachedShard{}
		}
		r.shards[cacheKey] = call.shard
	}
	r.shardMu.Unlock()
	close(call.done)
	if call.err != nil {
		return updater.DataFile{}, false, call.err
	}
	f, listed := call.shard.byPath[key]
	return f, listed, nil
}

func (r *remoteSource) loadShard(root *signedRoot, ref updater.ShardRef) (*cachedShard, error) {
	if ref.Size <= 0 || ref.Size > dataShardMax {
		return nil, fmt.Errorf("shard %s size %d is not usable", ref.ID, ref.Size)
	}
	path := filepath.Join(r.manifestDir(), "shards", strconv.FormatUint(root.version, 10), ref.ID)
	if body, err := os.ReadFile(path); err == nil && int64(len(body)) == ref.Size && updater.SHA256Hex(body) == ref.SHA256 {
		if shard, err := r.parseShard(root.version, ref.ID, body); err == nil {
			return shard, nil
		}
	}
	rel := "data-manifest/v/" + strconv.FormatUint(root.version, 10) + "/shards/" + ref.ID
	body, err := r.getMeta(rel, ref.Size)
	if err != nil {
		return nil, fmt.Errorf("shard %s: %w", ref.ID, err)
	}
	if int64(len(body)) != ref.Size || updater.SHA256Hex(body) != ref.SHA256 {
		return nil, fmt.Errorf("shard %s digest mismatch", ref.ID)
	}
	shard, err := r.parseShard(root.version, ref.ID, body)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err == nil {
		_ = writeFileAtomic(path, body)
	}
	return shard, nil
}

func (r *remoteSource) parseShard(version uint64, id string, body []byte) (*cachedShard, error) {
	parsed, err := updater.ParseDataShard(string(body))
	if err != nil {
		return nil, err
	}
	if parsed.Version != version || parsed.ID != id {
		return nil, fmt.Errorf("shard %s does not match the accepted root", id)
	}
	byPath := make(map[string]updater.DataFile, len(parsed.Files))
	for _, f := range parsed.Files {
		byPath[f.Path] = f
	}
	return &cachedShard{byPath: byPath}, nil
}
