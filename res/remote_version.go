package res

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/kivutar/goro/glog"
	"github.com/kivutar/goro/updater"
)

const (
	pollMinDefault       = 60 * time.Second
	pollMaxDefault       = 120 * time.Second
	pollOfflineMax       = 10 * time.Minute
	deltaDebounceDefault = 2 * time.Second
	dataManifestMax      = 64 << 20
	changesMax           = 8 << 20
	versionMax           = 64 << 10
)

// persistedDataVersion is cachedVersion on disk, keyed by the cache directory
// (already per host).
type persistedDataVersion struct {
	Version uint64 `json:"version"`
}

type changesResponse struct {
	Version uint64   `json:"version"`
	Paths   []string `json:"paths"`
	Full    bool     `json:"full"`
}

func pollInterval(jitter func() float64, min, max time.Duration) time.Duration {
	if min <= 0 {
		min = pollMinDefault
	}
	if max < min {
		max = min
	}
	span := max - min
	j := 0.0
	if jitter != nil {
		j = jitter()
	}
	if j < 0 {
		j = 0
	}
	if j > 1 {
		j = 1
	}
	return min + time.Duration(j*float64(span))
}

func offlinePollInterval(cur time.Duration) time.Duration {
	next := cur * 2
	if next < remoteOfflineBackoff {
		next = remoteOfflineBackoff
	}
	if next > pollOfflineMax {
		next = pollOfflineMax
	}
	return next
}

func defaultJitter() float64 { return rand.Float64() }

func (r *remoteSource) versionPath() string {
	return filepath.Join(r.cacheDir, "data-version.json")
}

func (r *remoteSource) trustedKeysPath() string {
	return filepath.Join(r.cacheDir, "trusted-keys.json")
}

func (r *remoteSource) loadCachedVersion() {
	raw, err := os.ReadFile(r.versionPath())
	if err != nil {
		return
	}
	var p persistedDataVersion
	if json.Unmarshal(raw, &p) != nil || p.Version == 0 {
		return
	}
	r.mu.Lock()
	r.cachedVersion = p.Version
	r.haveCached = true
	r.mu.Unlock()
	r.seenVersion.Store(p.Version)
}

func (r *remoteSource) persistCachedVersion() {
	r.mu.Lock()
	v := r.cachedVersion
	r.mu.Unlock()
	raw, err := json.Marshal(persistedDataVersion{Version: v})
	if err != nil {
		return
	}
	if err := writeFileAtomic(r.versionPath(), raw); err != nil {
		r.logOnce("data-version", "remote data: cannot persist data version: %v", err)
	}
}

func (r *remoteSource) noteVersionHeader(h http.Header) {
	s := strings.TrimSpace(h.Get("X-Data-Version"))
	if s == "" {
		return
	}
	v, err := strconv.ParseUint(s, 10, 64)
	if err != nil || v == 0 {
		return
	}
	r.noteVersion(v)
}

func (r *remoteSource) noteVersion(v uint64) {
	r.headerSeen.Store(true)
	for {
		old := r.seenVersion.Load()
		if v <= old {
			break
		}
		if r.seenVersion.CompareAndSwap(old, v) {
			break
		}
	}
	r.mu.Lock()
	if !r.haveCached {
		r.cachedVersion = v
		r.haveCached = true
		r.mu.Unlock()
		r.persistCachedVersion()
		go r.loadDataManifest()
		return
	}
	cached := r.cachedVersion
	r.mu.Unlock()
	if v > cached {
		r.wakeDelta()
	}
}

func (r *remoteSource) wakeDelta() {
	select {
	case r.deltaWake <- struct{}{}:
	default:
	}
}

func (r *remoteSource) startVersionWorkers() {
	go r.deltaLoop()
	go r.pollLoop()
}

func (r *remoteSource) stop() {
	r.stopOnce.Do(func() { close(r.stopCh) })
}

func (r *remoteSource) deltaLoop() {
	for {
		select {
		case <-r.stopCh:
			return
		case <-r.deltaWake:
		}
		timer := time.NewTimer(r.deltaDebounce)
		select {
		case <-r.stopCh:
			timer.Stop()
			return
		case <-timer.C:
		}
		r.runDelta()
	}
}

func (r *remoteSource) runDelta() {
	for {
		select {
		case <-r.stopCh:
			return
		default:
		}
		r.mu.Lock()
		cached := r.cachedVersion
		r.mu.Unlock()
		seen := r.seenVersion.Load()
		if seen <= cached {
			return
		}
		select {
		case r.background <- struct{}{}:
		case <-r.stopCh:
			return
		}
		body, err := r.getMeta("changes?since="+strconv.FormatUint(cached, 10), changesMax)
		<-r.background
		if err != nil {
			return
		}
		var resp changesResponse
		if json.Unmarshal(body, &resp) != nil || resp.Version == 0 {
			r.logOnce("changes", "remote data: bad changes response")
			return
		}
		if resp.Full {
			r.markAllStale()
			r.refreshIndex()
			r.queuePaths(dataVersionFullReloadPaths)
		} else {
			for _, p := range resp.Paths {
				r.evict(p)
			}
			r.queuePaths(resp.Paths)
		}
		r.mu.Lock()
		r.cachedVersion = resp.Version
		r.haveCached = true
		r.mu.Unlock()
		r.persistCachedVersion()
		r.loadDataManifest()
	}
}

func (r *remoteSource) queuePaths(paths []string) {
	if r.updated == nil {
		return
	}
	for _, p := range paths {
		if p != "" {
			r.updated(p)
		}
	}
}

func (r *remoteSource) markAllStale() {
	r.cacheGen.Add(1)
	r.mu.Lock()
	r.validated = map[string]struct{}{}
	r.missing = map[string]struct{}{}
	r.stale = map[string]struct{}{}
	r.mu.Unlock()
}

func (r *remoteSource) pollLoop() {
	interval := pollInterval(r.jitter, r.pollMin, r.pollMax)
	for {
		select {
		case <-r.stopCh:
			return
		case <-time.After(interval):
		}
		if r.headerSeen.Swap(false) {
			interval = pollInterval(r.jitter, r.pollMin, r.pollMax)
			continue
		}
		r.mu.Lock()
		offline := time.Now().Before(r.offlineUntil)
		r.mu.Unlock()
		if offline {
			interval = offlinePollInterval(interval)
			continue
		}
		select {
		case r.background <- struct{}{}:
		case <-r.stopCh:
			return
		}
		err := r.pollVersion()
		<-r.background
		if err != nil {
			interval = offlinePollInterval(interval)
			continue
		}
		interval = pollInterval(r.jitter, r.pollMin, r.pollMax)
	}
}

func (r *remoteSource) pollVersion() error {
	body, err := r.getMeta("version", versionMax)
	if err != nil {
		return err
	}
	if len(body) == 0 {
		return nil
	}
	var v struct {
		Version uint64 `json:"version"`
	}
	if json.Unmarshal(body, &v) == nil && v.Version > 0 {
		r.noteVersion(v.Version)
	}
	return nil
}

func (r *remoteSource) getMeta(rel string, limit int64) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), remoteRequestTimeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, r.base.String()+rel, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "goro-remote-data/1")
	resp, err := r.client.Do(req)
	if err != nil {
		r.mu.Lock()
		r.offlineUntil = time.Now().Add(remoteOfflineBackoff)
		r.mu.Unlock()
		return nil, connError{err}
	}
	defer resp.Body.Close()
	r.noteVersionHeader(resp.Header)
	if resp.StatusCode != http.StatusOK {
		_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 64<<10))
		if resp.StatusCode == http.StatusNotFound {
			return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
		}
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	if resp.ContentLength > limit {
		return nil, fmt.Errorf("%s is larger than %d bytes", rel, limit)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(body)) > limit {
		return nil, fmt.Errorf("%s is larger than %d bytes", rel, limit)
	}
	return body, nil
}

func (r *remoteSource) loadDataManifest() {
	body, err := r.getMeta("data-manifest", dataManifestMax)
	if err != nil {
		return
	}
	keys := r.keyRing.Keys()
	m, err := updater.OpenDataEnvelopeWithKeys(body, keys)
	if err != nil {
		r.logOnce("data-manifest", "remote data: signed data manifest rejected: %v", err)
		return
	}
	r.mu.Lock()
	accepted := r.acceptedManifest
	r.mu.Unlock()
	if err := m.RejectRollback(accepted); err != nil {
		r.logOnce("data-manifest", "remote data: %v", err)
		return
	}
	if err := r.keyRing.AcceptNextKeys(m.NextKeys); err != nil {
		glog.Warnf("remote data: next-keys not persisted: %v", err)
	}
	stored := m
	r.dataManifest.Store(&stored)
	r.mu.Lock()
	r.acceptedManifest = m.Version
	r.mu.Unlock()
}

func (r *remoteSource) verifyDigest(p string, body []byte) error {
	m := r.dataManifest.Load()
	if m == nil {
		return nil
	}
	return m.VerifyFile(p, body)
}

// StopRemote stops the version poll and delta worker. Tests call it via
// t.Cleanup; the game process exits instead.
func (m *Manager) StopRemote() {
	if m == nil || m.remote == nil {
		return
	}
	m.remote.stop()
}

// RemoteSeenVersion is the newest X-Data-Version observed.
func (m *Manager) RemoteSeenVersion() uint64 {
	if m == nil || m.remote == nil {
		return 0
	}
	return m.remote.seenVersion.Load()
}

// RemoteCachedVersion is the version the cache is consistent with.
func (m *Manager) RemoteCachedVersion() uint64 {
	if m == nil || m.remote == nil {
		return 0
	}
	m.remote.mu.Lock()
	defer m.remote.mu.Unlock()
	return m.remote.cachedVersion
}

// WaitRemoteManifest blocks until a verified data manifest is loaded or the
// timeout passes. Tests use it; the game never waits.
func (m *Manager) WaitRemoteManifest(timeout time.Duration) bool {
	if m == nil || m.remote == nil {
		return false
	}
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if m.remote.dataManifest.Load() != nil {
			return true
		}
		time.Sleep(5 * time.Millisecond)
	}
	return false
}

// dataVersionFullReloadPaths are the tables a full recheck rebuilds live.
var dataVersionFullReloadPaths = []string{
	"data/luafiles514/lua files/skillinfoz/skillid.lub",
	"data/luafiles514/lua files/skillinfoz/skillinfolist.lub",
	"data/luafiles514/lua files/skillinfoz/skilltreeview.lub",
	"data/luafiles514/lua files/skillinfoz/skilldescript.lub",
}
