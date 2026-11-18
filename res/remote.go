package res

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/kivutar/goro/glog"
)

// DefaultRemoteDataURL is the TuroranRO asset server (remoteclient2), the same
// source roBrowser and Korangar read. It already merges the client data in
// player priority order: System/ and the text overlays first, then slot-0
// translation.grf, then the other DATA.INI archives.
const DefaultRemoteDataURL = "https://ro.turoran.com/client/"

// Size caps for one streamed file and for the archive index (vars for tests).
var (
	remoteMaxFile  int64 = 128 << 20
	remoteMaxIndex int64 = 64 << 20
)

const (
	remoteOfflineBackoff = 30 * time.Second
	remoteRequestTimeout = 2 * time.Minute
	// remoteMaxConns caps connections to the asset server; remoteBackground
	// caps concurrent background work (refreshes and prefetches).
	remoteMaxConns   = 6
	remoteBackground = 4
)

// ErrRemotePending reports a streamed file that is still downloading after
// the caller's wait budget. It is a not-found error, so lookups fall through;
// ask again on a later frame.
var ErrRemotePending = fmt.Errorf("%w: still downloading", errResourceNotFound)

// RemoteOptions configures streaming of client data from an asset server.
type RemoteOptions struct {
	// BaseURL is the asset server root; files are requested at BaseURL+path.
	BaseURL string
	// CacheDir keeps downloaded files between runs.
	CacheDir string
	// Aliases maps a canonical client name (lowercase, slash separated) to the
	// server file that replaces it, for tables the server ships under a
	// versioned name (see TuroranRemoteAliases).
	Aliases map[string]string
	// Client overrides the HTTP client (tests).
	Client *http.Client
}

// TuroranRemoteAliases are the canonical names of the tables roBrowser's
// DBManager loads under Turoran file names. They match the picks in
// scripts/data/assemble-turoran-data.sh.
var TuroranRemoteAliases = map[string]string{
	"data/luafiles514/lua files/skillinfoz/skillid.lub":                "data/luafiles514/lua files/skillinfoz/skillid.turoran4.lub",
	"data/luafiles514/lua files/skillinfoz/skillinfolist.lub":          "data/luafiles514/lua files/skillinfoz/skillinfolist.turoran13.lub",
	"data/luafiles514/lua files/skillinfoz/skilltreeview.lub":          "data/luafiles514/lua files/skillinfoz/skilltreeview.fourth.lub",
	"data/luafiles514/lua files/skillinfoz/skilldescript.lub":          "System/skilldescript.turoran.re2.lub",
	"data/luafiles514/lua files/datainfo/petinfo.lub":                  "data/luafiles514/lua files/datainfo/petinfo.turoran.lub",
	"data/luafiles514/lua files/datainfo/addrandomoptionnametable.lub": "data/luafiles514/lua files/datainfo/addrandomoptionnametable.turoran.lub",
	"data/cardprefixnametable.txt":                                     "data/cardprefixnametable.turoran.txt",
	"data/num2cardillustnametable.txt":                                 "data/num2cardillustnametable.turoran.txt",
	"system/iteminfo.lua":                                              "System/itemInfo.turoran.renewal11.lua",
}

// remoteSource fetches files on demand and caches them on disk. Every failure
// is reported as errResourceNotFound so lookups fall through to the next
// candidate instead of aborting.
type remoteSource struct {
	base     *url.URL
	host     string // cache key prefix, so one cache_dir can serve several hosts
	cacheDir string
	aliases  map[string]string
	client   *http.Client

	// index holds the lowercase names the server's archives contain
	// (GET list-files). Until it arrives every eligible name is requested.
	index      atomic.Pointer[remoteIndex]
	indexReady chan struct{}

	mu           sync.Mutex
	missing      map[string]struct{}
	validated    map[string]struct{}
	inflight     map[string]*remoteCall
	logged       map[string]struct{}
	offlineUntil time.Time
	requests     atomic.Int64
	background   chan struct{}
	pending      map[string]*remotePending
	// refreshed is called after a background refresh (tests).
	refreshed func(string)
}

type remoteIndex struct {
	names      map[string]struct{}
	generation string
}

type remotePending struct {
	done   chan struct{}
	data   []byte
	source string
	err    error
}

// connError marks failures to reach the server (dial, TLS, no response), the
// only errors that put the source in its offline back-off.
type connError struct{ err error }

func (e connError) Error() string { return e.err.Error() }
func (e connError) Unwrap() error { return e.err }

type remoteCall struct {
	done chan struct{}
	data []byte
	err  error
}

type remoteMeta struct {
	Path       string `json:"path"`
	ETag       string `json:"etag,omitempty"`
	Generation string `json:"generation,omitempty"`
}

func newRemoteSource(opts RemoteOptions) (*remoteSource, error) {
	raw := strings.TrimSpace(opts.BaseURL)
	if raw == "" {
		return nil, errors.New("empty remote data URL")
	}
	if !strings.HasSuffix(raw, "/") {
		raw += "/"
	}
	base, err := url.Parse(raw)
	if err != nil || (base.Scheme != "https" && base.Scheme != "http") || base.Host == "" {
		return nil, fmt.Errorf("invalid remote data URL %q", opts.BaseURL)
	}
	if opts.CacheDir == "" {
		return nil, errors.New("empty remote data cache directory")
	}
	if err := os.MkdirAll(filepath.Join(opts.CacheDir, "files"), 0o755); err != nil {
		return nil, fmt.Errorf("remote data cache: %w", err)
	}
	client := opts.Client
	if client == nil {
		transport := http.DefaultTransport.(*http.Transport).Clone()
		transport.DialContext = (&net.Dialer{Timeout: 5 * time.Second, KeepAlive: 30 * time.Second}).DialContext
		transport.ResponseHeaderTimeout = 20 * time.Second
		transport.MaxIdleConnsPerHost = remoteMaxConns
		transport.MaxConnsPerHost = remoteMaxConns
		client = &http.Client{Transport: transport}
	}
	aliases := make(map[string]string, len(opts.Aliases))
	for from, to := range opts.Aliases {
		key := strings.ToLower(remotePath(from))
		target := remotePath(to)
		if key != "" && target != "" {
			aliases[key] = target
		}
	}
	return &remoteSource{
		base:       base,
		host:       strings.ToLower(base.Host),
		cacheDir:   opts.CacheDir,
		background: make(chan struct{}, remoteBackground),
		pending:    make(map[string]*remotePending),
		aliases:    aliases,
		client:     client,
		indexReady: make(chan struct{}),
		missing:    make(map[string]struct{}),
		validated:  make(map[string]struct{}),
		inflight:   make(map[string]*remoteCall),
		logged:     make(map[string]struct{}),
	}, nil
}

// EnableRemote streams files that are not in the data directory or its local
// archives from an asset server. It returns at once; the server's name index
// loads in the background, so startup never waits for the network.
func (m *Manager) EnableRemote(opts RemoteOptions) error {
	source, err := newRemoteSource(opts)
	if err != nil {
		return err
	}
	m.remote = source
	// newManager may already have resolved resnametable.txt (reading
	// clientinfo falls back to aliases) while nothing was streamed; load the
	// table again so the server's copy counts. EnableRemote must run before
	// the manager is shared between goroutines.
	m.resourceAliasesOnce = sync.Once{}
	m.resourceAliases = nil
	go source.loadIndex()
	glog.Infof("remote data: streaming from %s (cache %s)", source.base.Redacted(), opts.CacheDir)
	return nil
}

// RemoteEnabled reports whether files can be streamed from an asset server.
func (m *Manager) RemoteEnabled() bool { return m != nil && m.remote != nil }

// WaitRemoteIndex blocks until the server's name index has loaded or failed,
// or the timeout passes. Used by tools and tests; the game never waits.
func (m *Manager) WaitRemoteIndex(timeout time.Duration) bool {
	if m == nil || m.remote == nil {
		return false
	}
	select {
	case <-m.remote.indexReady:
		return m.remote.index.Load() != nil
	case <-time.After(timeout):
		return false
	}
}

// RemoteRequests returns how many HTTP requests were sent for files.
func (m *Manager) RemoteRequests() int64 {
	if m == nil || m.remote == nil {
		return 0
	}
	return m.remote.requests.Load()
}

// remotePath turns a client resource name (backslashes, EUC-KR or UTF-8) into
// the server path: UTF-8, slash separated, "data/" lowercase and "System/"
// spelled as the server's directory. Only data/ and System/ are served.
func remotePath(name string) string {
	clean := resourceAliasPath(name)
	if clean == "" {
		return ""
	}
	first, rest, ok := strings.Cut(clean, "/")
	if !ok || rest == "" {
		return ""
	}
	// The server list stays local: never stream a login server table.
	switch strings.ToLower(path.Base(clean)) {
	case "clientinfo.xml", "sclientinfo.xml":
		return ""
	}
	switch strings.ToLower(first) {
	case "data":
		return "data/" + rest
	case "system":
		return "System/" + rest
	}
	return ""
}

func isRemoteTextFile(name string) bool {
	switch strings.ToLower(path.Ext(name)) {
	case ".lub", ".lua", ".txt", ".xml", ".json", ".ini":
		return true
	}
	return false
}

// resolve returns the server paths to try for a client name, best first, or
// none if the server cannot have it. A Turoran alias comes before the
// canonical name, which stays as the fallback for servers without it.
// Archive files (data/) are checked against the index; text tables may also
// come from the server's loose overlays, which list-files does not report.
func (r *remoteSource) resolve(name string) []string {
	p := remotePath(name)
	if p == "" {
		return nil
	}
	var out []string
	if target, ok := r.aliases[strings.ToLower(p)]; ok {
		out = append(out, target)
	}
	if r.eligible(p) {
		out = append(out, p)
	}
	return out
}

func (r *remoteSource) eligible(p string) bool {
	if strings.HasPrefix(p, "System/") || isRemoteTextFile(p) {
		return true
	}
	index := r.index.Load()
	if index == nil {
		return true
	}
	_, ok := index.names[strings.ToLower(p)]
	return ok
}

// has reports whether the server is known to have name, without a request.
func (r *remoteSource) has(name string) bool {
	index := r.index.Load()
	if index == nil {
		return false
	}
	for _, p := range r.resolve(name) {
		if _, ok := index.names[strings.ToLower(p)]; ok {
			return true
		}
	}
	return false
}

func (r *remoteSource) read(name string) ([]byte, error) {
	for _, p := range r.resolve(name) {
		if data, err := r.readPath(p); err == nil {
			return data, nil
		}
	}
	return nil, fmt.Errorf("%w: %s", errResourceNotFound, name)
}

func (r *remoteSource) readPath(p string) ([]byte, error) {
	r.mu.Lock()
	if _, ok := r.missing[p]; ok {
		r.mu.Unlock()
		return nil, errResourceNotFound
	}
	if call, ok := r.inflight[p]; ok {
		r.mu.Unlock()
		<-call.done
		return call.data, call.err
	}
	call := &remoteCall{done: make(chan struct{})}
	r.inflight[p] = call
	r.mu.Unlock()

	call.data, call.err = r.fetch(p)
	if call.err != nil {
		call.err = fmt.Errorf("%w: %s (%v)", errResourceNotFound, p, call.err)
	}
	r.mu.Lock()
	delete(r.inflight, p)
	r.mu.Unlock()
	close(call.done)
	return call.data, call.err
}

func (r *remoteSource) cachePaths(p string) (string, string) {
	sum := sha256.Sum256([]byte(r.host + "\x00" + p))
	name := hex.EncodeToString(sum[:])
	dir := filepath.Join(r.cacheDir, "files", name[:2])
	return filepath.Join(dir, name), filepath.Join(dir, name+".json")
}

func (r *remoteSource) fetch(p string) ([]byte, error) {
	dataPath, metaPath := r.cachePaths(p)
	cached, cacheErr := os.ReadFile(dataPath)
	var meta remoteMeta
	if cacheErr == nil {
		if raw, err := os.ReadFile(metaPath); err == nil {
			_ = json.Unmarshal(raw, &meta)
		}
		if meta.Path != p {
			// Hash collision or a torn write: refetch.
			cacheErr = errors.New("cache metadata mismatch")
		}
	}
	generation := ""
	if index := r.index.Load(); index != nil {
		generation = index.generation
	}
	if cacheErr == nil {
		switch r.revalidation(p, meta, generation) {
		case revalidateNever:
			return cached, nil
		case revalidateBackground:
			// Serve the cached table now so startup never waits on the
			// network; a changed overlay is picked up on the next run.
			r.markValidated(p)
			go r.refresh(p, dataPath, metaPath, meta, generation)
			return cached, nil
		}
	}

	r.mu.Lock()
	offline := time.Now().Before(r.offlineUntil)
	r.mu.Unlock()
	if offline {
		if cacheErr == nil {
			return cached, nil
		}
		return nil, errors.New("asset server offline")
	}

	etag := ""
	if cacheErr == nil {
		etag = meta.ETag
	}
	body, newETag, status, err := r.get(p, etag)
	switch {
	case err != nil:
		var conn connError
		if errors.As(err, &conn) {
			r.mu.Lock()
			r.offlineUntil = time.Now().Add(remoteOfflineBackoff)
			r.mu.Unlock()
			r.logOnce("offline", "remote data: %s unreachable, using cached files for %s: %v", r.base.Redacted(), remoteOfflineBackoff, err)
		} else {
			// One bad file (too large, cut short) must not stall the rest.
			r.logOnce("file:"+p, "remote data: %s: %v", p, err)
		}
		if cacheErr == nil {
			return cached, nil
		}
		return nil, err
	case status == http.StatusNotModified && cacheErr == nil:
		r.markValidated(p)
		if meta.Generation != generation && generation != "" {
			meta.Generation = generation
			r.writeMeta(metaPath, meta)
		}
		return cached, nil
	case status == http.StatusOK:
		r.markValidated(p)
		if err := r.store(dataPath, metaPath, body, remoteMeta{Path: p, ETag: newETag, Generation: generation}); err != nil {
			r.logOnce("cache", "remote data: cannot write cache %s: %v", r.cacheDir, err)
		}
		return body, nil
	case status == http.StatusNotFound:
		r.mu.Lock()
		r.missing[p] = struct{}{}
		r.mu.Unlock()
		_ = os.Remove(dataPath)
		_ = os.Remove(metaPath)
		glog.Debugf("remote data: missing %s", p)
		return nil, errors.New("not on asset server")
	default:
		r.logOnce("status:"+p, "remote data: %s returned HTTP %d", p, status)
		if cacheErr == nil {
			return cached, nil
		}
		return nil, fmt.Errorf("HTTP %d", status)
	}
}

type revalidateMode int

const (
	revalidateNever revalidateMode = iota
	revalidateBackground
	revalidateNow
)

// revalidation: text tables (the overlays the server sends no-cache) are
// refreshed in the background once per run; archive files are checked again
// only after the server's archive index changed since they were downloaded.
func (r *remoteSource) revalidation(p string, meta remoteMeta, generation string) revalidateMode {
	r.mu.Lock()
	_, done := r.validated[p]
	r.mu.Unlock()
	if done {
		return revalidateNever
	}
	if isRemoteTextFile(p) || strings.HasPrefix(p, "System/") {
		return revalidateBackground
	}
	if generation != "" && meta.Generation != "" && meta.Generation != generation {
		return revalidateNow
	}
	return revalidateNever
}

// refresh revalidates a cached file and stores a newer copy if there is one.
func (r *remoteSource) refresh(p, dataPath, metaPath string, meta remoteMeta, generation string) {
	r.background <- struct{}{}
	defer func() { <-r.background }()
	body, etag, status, err := r.get(p, meta.ETag)
	switch {
	case err != nil:
		r.logOnce("refresh", "remote data: cannot refresh cached files: %v", err)
	case status == http.StatusOK:
		if err := r.store(dataPath, metaPath, body, remoteMeta{Path: p, ETag: etag, Generation: generation}); err != nil {
			r.logOnce("cache", "remote data: cannot write cache %s: %v", r.cacheDir, err)
		}
	case status == http.StatusNotFound:
		_ = os.Remove(dataPath)
		_ = os.Remove(metaPath)
	}
	if r.refreshed != nil {
		r.refreshed(p)
	}
}

func (r *remoteSource) markValidated(p string) {
	r.mu.Lock()
	r.validated[p] = struct{}{}
	r.mu.Unlock()
}

// fileURL percent-encodes every segment, so Korean names go out as UTF-8.
func (r *remoteSource) fileURL(p string) string {
	parts := strings.Split(p, "/")
	for i, part := range parts {
		parts[i] = url.PathEscape(part)
	}
	return r.base.String() + strings.Join(parts, "/")
}

func (r *remoteSource) get(p, etag string) ([]byte, string, int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), remoteRequestTimeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, r.fileURL(p), nil)
	if err != nil {
		return nil, "", 0, err
	}
	req.Header.Set("User-Agent", "goro-remote-data/1")
	if etag != "" {
		req.Header.Set("If-None-Match", etag)
	}
	r.requests.Add(1)
	resp, err := r.client.Do(req)
	if err != nil {
		return nil, "", 0, connError{err}
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 64<<10))
		return nil, "", resp.StatusCode, nil
	}
	if resp.ContentLength > remoteMaxFile {
		return nil, "", 0, fmt.Errorf("%s is larger than %d bytes", p, remoteMaxFile)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, remoteMaxFile+1))
	if err != nil {
		return nil, "", 0, err
	}
	if int64(len(body)) > remoteMaxFile {
		return nil, "", 0, fmt.Errorf("%s is larger than %d bytes", p, remoteMaxFile)
	}
	if resp.ContentLength >= 0 && int64(len(body)) != resp.ContentLength {
		return nil, "", 0, fmt.Errorf("%s: short body", p)
	}
	return body, resp.Header.Get("ETag"), http.StatusOK, nil
}

func (r *remoteSource) store(dataPath, metaPath string, body []byte, meta remoteMeta) error {
	if err := os.MkdirAll(filepath.Dir(dataPath), 0o755); err != nil {
		return err
	}
	if err := writeFileAtomic(dataPath, body); err != nil {
		return err
	}
	r.writeMeta(metaPath, meta)
	return nil
}

func (r *remoteSource) writeMeta(metaPath string, meta remoteMeta) {
	raw, err := json.Marshal(meta)
	if err == nil {
		err = writeFileAtomic(metaPath, raw)
	}
	if err != nil {
		r.logOnce("cache", "remote data: cannot write cache %s: %v", r.cacheDir, err)
	}
}

func writeFileAtomic(name string, data []byte) error {
	tmp, err := os.CreateTemp(filepath.Dir(name), ".tmp-*")
	if err != nil {
		return err
	}
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmp.Name())
		return err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmp.Name())
		return err
	}
	return os.Rename(tmp.Name(), name)
}

func (r *remoteSource) logOnce(key, format string, args ...any) {
	r.mu.Lock()
	_, seen := r.logged[key]
	r.logged[key] = struct{}{}
	r.mu.Unlock()
	if !seen {
		glog.Warnf(format, args...)
	}
}

// loadIndex loads the cached archive index first, then asks the server for
// a newer one with its ETag. An unchanged list is neither parsed nor written
// again. Either way indexReady closes, so waiters never hang.
func (r *remoteSource) loadIndex() {
	defer close(r.indexReady)
	indexPath, etagPath := r.indexPaths()
	cachedETag := ""
	if cached, err := os.ReadFile(indexPath); err == nil {
		if index, err := parseRemoteIndex(cached); err == nil {
			r.index.Store(index)
			if raw, err := os.ReadFile(etagPath); err == nil {
				cachedETag = strings.TrimSpace(string(raw))
			}
		}
	}
	body, etag, status, err := r.getIndex(cachedETag)
	if err == nil && status == http.StatusNotModified && r.index.Load() != nil {
		return
	}
	if err != nil || status != http.StatusOK {
		r.logOnce("index", "remote data: archive index unavailable (status %d): %v", status, err)
		return
	}
	if current := r.index.Load(); current != nil && current.generation == remoteIndexGeneration(body) {
		r.writeIndexETag(etagPath, etag)
		return
	}
	index, err := parseRemoteIndex(body)
	if err != nil {
		r.logOnce("index", "remote data: bad archive index: %v", err)
		return
	}
	r.index.Store(index)
	if err := writeFileAtomic(indexPath, body); err == nil {
		r.writeIndexETag(etagPath, etag)
	}
	glog.Infof("remote data: %d archive files indexed", len(index.names))
}

func (r *remoteSource) indexPaths() (string, string) {
	sum := sha256.Sum256([]byte(r.host))
	name := "list-files-" + hex.EncodeToString(sum[:8])
	return filepath.Join(r.cacheDir, name+".json"), filepath.Join(r.cacheDir, name+".etag")
}

func (r *remoteSource) writeIndexETag(etagPath, etag string) {
	if etag == "" {
		_ = os.Remove(etagPath)
		return
	}
	_ = writeFileAtomic(etagPath, []byte(etag))
}

func (r *remoteSource) getIndex(etag string) ([]byte, string, int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), remoteRequestTimeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, r.base.String()+"list-files", nil)
	if err != nil {
		return nil, "", 0, err
	}
	req.Header.Set("User-Agent", "goro-remote-data/1")
	if etag != "" {
		req.Header.Set("If-None-Match", etag)
	}
	resp, err := r.client.Do(req)
	if err != nil {
		return nil, "", 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, "", resp.StatusCode, nil
	}
	if resp.ContentLength > remoteMaxIndex {
		return nil, "", 0, errors.New("archive index too large")
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, remoteMaxIndex+1))
	if err != nil {
		return nil, "", 0, err
	}
	if int64(len(body)) > remoteMaxIndex {
		return nil, "", 0, errors.New("archive index too large")
	}
	return body, resp.Header.Get("ETag"), http.StatusOK, nil
}

func remoteIndexGeneration(body []byte) string {
	sum := sha256.Sum256(body)
	return hex.EncodeToString(sum[:])
}

func parseRemoteIndex(body []byte) (*remoteIndex, error) {
	var names []string
	if err := json.Unmarshal(body, &names); err != nil {
		return nil, err
	}
	if len(names) == 0 {
		return nil, errors.New("empty archive index")
	}
	index := &remoteIndex{names: make(map[string]struct{}, len(names))}
	for _, name := range names {
		if p := remotePath(name); p != "" {
			index.names[strings.ToLower(p)] = struct{}{}
		}
	}
	index.generation = remoteIndexGeneration(body)
	return index, nil
}

// ReadFileCandidatesWithin is ReadFileCandidates for per-frame callers: when
// a lookup needs the network it waits at most budget, then returns
// ErrRemotePending while the download finishes in the background. Asking
// again later returns the file (from the cache) without waiting twice. Do not
// record ErrRemotePending as a permanent miss.
func (m *Manager) ReadFileCandidatesWithin(names []string, budget time.Duration) ([]byte, string, error) {
	if m == nil || m.remote == nil || budget <= 0 {
		return m.ReadFileCandidates(names)
	}
	r := m.remote
	key := strings.Join(names, "\x00")
	r.mu.Lock()
	call, running := r.pending[key]
	if !running {
		call = &remotePending{done: make(chan struct{})}
		r.pending[key] = call
		go func() {
			call.data, call.source, call.err = m.ReadFileCandidates(names)
			r.mu.Lock()
			delete(r.pending, key)
			r.mu.Unlock()
			close(call.done)
		}()
	}
	r.mu.Unlock()
	if running {
		// A previous frame already waited; never stall twice for one file.
		select {
		case <-call.done:
			return call.data, call.source, call.err
		default:
			return nil, "", ErrRemotePending
		}
	}
	timer := time.NewTimer(budget)
	defer timer.Stop()
	select {
	case <-call.done:
		return call.data, call.source, call.err
	case <-timer.C:
		return nil, "", ErrRemotePending
	}
}

// RemoteBatch tracks background prefetches. A nil batch is done.
type RemoteBatch struct {
	remaining atomic.Int64
}

// Done reports whether every prefetch in the batch has finished.
func (b *RemoteBatch) Done() bool { return b == nil || b.remaining.Load() <= 0 }

// PrefetchRemote downloads, in parallel and in the background, the file each
// candidate list would stream, so a loading screen can keep drawing while
// the network works. Lists that a local file satisfies first are skipped.
func (m *Manager) PrefetchRemote(lists [][]string) *RemoteBatch {
	if m == nil || m.remote == nil {
		return nil
	}
	r := m.remote
	seen := make(map[string]struct{})
	var paths []string
	for _, candidates := range lists {
		for _, name := range candidates {
			if m.hasLocal(name) {
				break
			}
			if p := r.prefetchPath(name); p != "" {
				if _, ok := seen[p]; !ok {
					seen[p] = struct{}{}
					paths = append(paths, p)
				}
				break
			}
		}
	}
	if len(paths) == 0 {
		return nil
	}
	batch := &RemoteBatch{}
	batch.remaining.Store(int64(len(paths)))
	for _, p := range paths {
		go func() {
			defer batch.remaining.Add(-1)
			r.background <- struct{}{}
			defer func() { <-r.background }()
			_, _ = r.readPath(p)
		}()
	}
	return batch
}

// prefetchPath picks the server file a lookup of name would read: listed
// archive files once the index is in, otherwise the first resolvable path.
func (r *remoteSource) prefetchPath(name string) string {
	paths := r.resolve(name)
	if len(paths) == 0 {
		return ""
	}
	index := r.index.Load()
	if index == nil {
		return paths[0]
	}
	for _, p := range paths {
		if _, ok := index.names[strings.ToLower(p)]; ok {
			return p
		}
	}
	return ""
}

func (m *Manager) hasLocal(name string) bool {
	if _, ok := m.Find(name); ok {
		return true
	}
	for _, archive := range m.Archives {
		if archive != nil && archive.Has(name) {
			return true
		}
	}
	return false
}
