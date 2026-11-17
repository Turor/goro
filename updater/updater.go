package updater

import (
	"bytes"
	"context"
	"crypto/ed25519"
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
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// Link-time release identity. A release build sets all three, for example:
//
//	-ldflags "-X github.com/kivutar/goro/updater.Channel=main
//	          -X github.com/kivutar/goro/updater.Version=main-0123456789ab
//	          -X github.com/kivutar/goro/updater.Build=1763419200"
//
// An empty Channel (every plain `go build`) disables the launch check, so a
// developer binary is never replaced.
var (
	Channel = ""
	Version = "dev"
	Build   = "0"
)

const (
	// DefaultCatalogURL is the public main-channel catalog. remoteclient2
	// answers it with short-lived presigned Garage URLs.
	DefaultCatalogURL = "https://ro.turoran.com/client/goro/main/latest"

	// EnvSkip set to 1 disables the launch check.
	EnvSkip = "GORO_SKIP_UPDATE"
	// EnvURL overrides the catalog or manifest URL (still allowlisted).
	EnvURL = "GORO_UPDATE_MANIFEST_URL"
	// EnvForce runs the check even in a development build (Channel unset).
	EnvForce = "GORO_UPDATE_FORCE"

	// SelfCheckFlag makes the binary print SelfCheckToken and exit.
	SelfCheckFlag  = "--update-self-check"
	SelfCheckToken = "goro-update-self-check-ok"
	// BuildInfoFlag prints the link-time release identity and exits.
	BuildInfoFlag = "--build-info"

	catalogLimit  = 64 * 1024
	manifestLimit = 256 * 1024
	fileLimit     = 512 * 1024 * 1024
	lockStale     = 15 * time.Minute
)

// Outcome of a launch check.
type Outcome int

const (
	Skipped Outcome = iota
	UpToDate
	Applied
)

func (o Outcome) String() string {
	switch o {
	case UpToDate:
		return "up to date"
	case Applied:
		return "applied"
	default:
		return "skipped"
	}
}

// Result describes what a launch check did.
type Result struct {
	Outcome Outcome
	Reason  string
	Version string
	Build   uint64
}

// Options controls one launch check. Zero values pick production defaults.
type Options struct {
	URL         string
	InstallRoot string
	Executable  string // absolute path of the running binary
	PublicKey   ed25519.PublicKey
	Installed   uint64 // build compiled into the running binary
	GOOS        string
	GOARCH      string
	HTTPClient  *http.Client
	// SelfCheck runs the staged binary; nil runs it with SelfCheckFlag.
	SelfCheck func(exe string, want uint64) error
	Logf      func(format string, args ...any)
}

// HandleFlags answers SelfCheckFlag and BuildInfoFlag. It returns true when
// the process should exit immediately afterwards.
func HandleFlags(args []string, out io.Writer) bool {
	for _, a := range args {
		switch a {
		case SelfCheckFlag:
			fmt.Fprintf(out, "%s build=%s version=%s\n", SelfCheckToken, Build, Version)
			return true
		case BuildInfoFlag:
			ch := Channel
			if ch == "" {
				ch = "dev"
			}
			fmt.Fprintf(out, "goro version=%s build=%s channel=%s %s/%s\n", Version, Build, ch, runtime.GOOS, runtime.GOARCH)
			return true
		}
	}
	return false
}

// Enabled reports whether a launch check should run for these args.
func Enabled(args []string) bool {
	if strings.TrimSpace(os.Getenv(EnvSkip)) == "1" {
		return false
	}
	for _, a := range args {
		if a == "--headless" || a == "-headless" || strings.HasPrefix(a, "--headless=") || strings.HasPrefix(a, "-headless=") {
			// Simulation and CI runs never touch the network or the binary.
			return false
		}
	}
	return Channel != "" || os.Getenv(EnvForce) == "1"
}

// CheckOnLaunch runs the production launch check for the current process.
// When an update is applied it starts the new binary with the same args and
// returns Applied; the caller should then exit.
func CheckOnLaunch(args []string, logf func(string, ...any)) (Result, error) {
	exe, err := os.Executable()
	if err != nil {
		return Result{Reason: "no executable path"}, err
	}
	if resolved, err := filepath.EvalSymlinks(exe); err == nil {
		exe = resolved
	}
	exe, _ = filepath.Abs(exe)
	target := os.Getenv(EnvURL)
	if strings.TrimSpace(target) == "" {
		target = DefaultCatalogURL
	}
	installed, _ := strconv.ParseUint(Build, 10, 64)
	res, err := Run(context.Background(), Options{
		URL:         target,
		InstallRoot: filepath.Dir(exe),
		Executable:  exe,
		Installed:   installed,
		Logf:        logf,
	})
	if err != nil || res.Outcome != Applied {
		return res, err
	}
	if err := relaunch(exe, args); err != nil {
		return res, fmt.Errorf("update %s applied but relaunch failed: %w", res.Version, err)
	}
	return res, nil
}

func relaunch(exe string, args []string) error {
	cmd := exec.Command(exe, args...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	cmd.Env = append(os.Environ(), EnvSkip+"=1")
	return cmd.Start()
}

// Run performs one check: fetch, verify, download, swap. Network trouble and a
// missing channel are reported as Skipped with a nil error. Any error leaves
// the installed binary unchanged.
func Run(ctx context.Context, o Options) (Result, error) {
	o.defaults()
	if strings.TrimSpace(o.URL) == "" {
		return Result{Reason: "no update url"}, nil
	}
	if isZero(o.PublicKey) {
		return Result{}, errors.New("update public key is unset")
	}
	if rel, err := filepath.Rel(o.InstallRoot, o.Executable); err != nil || strings.HasPrefix(rel, "..") {
		return Result{}, errors.New("executable is outside the install root")
	}
	cleanupOld(o.Executable, o.logf)

	fetched, err := o.fetch(ctx, o.URL, catalogLimit+manifestLimit)
	if err != nil {
		if isUnreachable(err) {
			return Result{Reason: "channel unreachable: " + err.Error()}, nil
		}
		return Result{}, err
	}
	manifestURL := o.URL
	var cat *catalog
	envelope := fetched
	if looksLikeJSON(fetched) {
		cat, err = parseCatalog(fetched)
		if err != nil {
			return Result{}, err
		}
		if !allowedURL(cat.ManifestURL) {
			return Result{}, errors.New("catalog manifest_url is not an allowlisted release url")
		}
		for _, f := range cat.Files {
			if !allowedURL(f.URL) {
				return Result{}, fmt.Errorf("catalog file url for %s is not allowlisted", f.Path)
			}
		}
		manifestURL = cat.ManifestURL
		envelope, err = o.fetch(ctx, manifestURL, manifestLimit)
		if err != nil {
			if isUnreachable(err) {
				return Result{Reason: "manifest unreachable: " + err.Error()}, nil
			}
			return Result{}, err
		}
	}
	m, err := OpenEnvelope(envelope, o.PublicKey)
	if err != nil {
		return Result{}, err
	}
	res := Result{Version: m.Version, Build: m.Build}
	installed := o.Installed
	if applied := appliedBuild(o.InstallRoot); applied > installed {
		installed = applied
	}
	if m.Build == installed {
		res.Outcome = UpToDate
		return res, nil
	}
	if m.Build < installed {
		res.Reason = fmt.Sprintf("rollback protection: manifest build %d is older than installed %d", m.Build, installed)
		return res, nil
	}
	file, ok := platformFile(m, o.GOOS, o.GOARCH)
	if !ok {
		res.Reason = fmt.Sprintf("release %s has no %s/%s build", m.Version, o.GOOS, o.GOARCH)
		return res, nil
	}
	for _, f := range m.Files {
		if f.Path != file.Path {
			o.logf("update: ignoring %s (only the %s/%s binary is installed)", f.Path, o.GOOS, o.GOARCH)
		}
	}
	src, err := artifactURL(manifestURL, file, cat)
	if err != nil {
		return res, err
	}
	if err := o.apply(ctx, m, file, src); err != nil {
		return res, err
	}
	res.Outcome = Applied
	return res, nil
}

func (o *Options) defaults() {
	if o.PublicKey == nil {
		o.PublicKey = EmbeddedPublicKey
	}
	if o.GOOS == "" {
		o.GOOS = runtime.GOOS
	}
	if o.GOARCH == "" {
		o.GOARCH = runtime.GOARCH
	}
	if o.HTTPClient == nil {
		o.HTTPClient = &http.Client{
			Timeout: 10 * time.Minute,
			// Every hop must itself be allowlisted, so redirects are refused.
			CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
		}
	}
	if o.SelfCheck == nil {
		o.SelfCheck = runSelfCheck
	}
}

func (o *Options) logf(format string, args ...any) {
	if o.Logf != nil {
		o.Logf(format, args...)
	}
}

func (o *Options) apply(ctx context.Context, m Manifest, f File, src string) error {
	updates := filepath.Join(o.InstallRoot, "updates")
	if err := os.MkdirAll(updates, 0o755); err != nil {
		return fmt.Errorf("install directory is not writable: %w", err)
	}
	unlock, err := lock(updates)
	if err != nil {
		return err
	}
	defer unlock()

	staging := filepath.Join(updates, "staging", m.Version)
	_ = os.RemoveAll(staging)
	if err := os.MkdirAll(staging, 0o755); err != nil {
		return err
	}
	defer os.RemoveAll(staging)
	staged := filepath.Join(staging, filepath.Base(o.Executable))
	o.logf("update: downloading %s (%d bytes) for %s", m.Version, f.Size, f.Path)
	if err := o.download(ctx, src, staged, f); err != nil {
		return err
	}
	if err := os.Chmod(staged, 0o755); err != nil {
		return err
	}
	// Check the staged copy before touching the live binary.
	if err := o.SelfCheck(staged, m.Build); err != nil {
		return fmt.Errorf("staged binary failed self-check: %w", err)
	}

	exe := o.Executable
	old := exe + ".old"
	_ = os.Remove(old)
	// Renaming the running image is the supported way to replace it on
	// Windows; on Unix the running process keeps its inode.
	if err := os.Rename(exe, old); err != nil {
		return fmt.Errorf("move running binary aside: %w", err)
	}
	restore := func() {
		_ = os.Remove(exe)
		_ = os.Rename(old, exe)
	}
	if err := os.Rename(staged, exe); err != nil {
		restore()
		return fmt.Errorf("install new binary: %w", err)
	}
	if err := o.SelfCheck(exe, m.Build); err != nil {
		restore()
		return fmt.Errorf("installed binary failed self-check: %w", err)
	}
	if err := writeAtomic(filepath.Join(updates, "applied.build"), []byte(strconv.FormatUint(m.Build, 10)+"\n")); err != nil {
		restore()
		return err
	}
	_ = writeAtomic(filepath.Join(updates, "applied.version"), []byte(m.Version+"\n"))
	o.logf("update: installed %s (build %d)", m.Version, m.Build)
	return nil
}

func (o *Options) download(ctx context.Context, src, dst string, f File) error {
	if f.Size > fileLimit {
		return errors.New("artifact exceeds size limit")
	}
	body, err := o.open(ctx, src, f.Size)
	if err != nil {
		return err
	}
	defer body.Close()
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o755)
	if err != nil {
		return err
	}
	h := sha256.New()
	n, err := io.Copy(io.MultiWriter(out, h), io.LimitReader(body, f.Size+1))
	if err == nil {
		err = out.Sync()
	}
	if cerr := out.Close(); err == nil {
		err = cerr
	}
	if err != nil {
		return err
	}
	if n != f.Size {
		return fmt.Errorf("size mismatch for %s: got %d want %d", f.Path, n, f.Size)
	}
	if got := hex.EncodeToString(h.Sum(nil)); got != f.SHA256 {
		return fmt.Errorf("sha256 mismatch for %s", f.Path)
	}
	return nil
}

func (o *Options) fetch(ctx context.Context, rawURL string, limit int64) ([]byte, error) {
	body, err := o.open(ctx, rawURL, limit)
	if err != nil {
		return nil, err
	}
	defer body.Close()
	data, err := io.ReadAll(io.LimitReader(body, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > limit {
		return nil, errors.New("response exceeds limit")
	}
	return data, nil
}

// open returns a reader for an allowlisted URL.
func (o *Options) open(ctx context.Context, rawURL string, limit int64) (io.ReadCloser, error) {
	if !allowedURL(rawURL) {
		return nil, fmt.Errorf("url is not allowlisted: %s", redact(rawURL))
	}
	u, _ := url.Parse(rawURL)
	if u.Scheme == "file" {
		return os.Open(filepath.FromSlash(fileURLPath(u)))
	}
	ctx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		cancel()
		return nil, err
	}
	req.Header.Set("User-Agent", "goro-updater/"+Version)
	resp, err := o.HTTPClient.Do(req)
	if err != nil {
		cancel()
		return nil, unreachable{err}
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		cancel()
		err := fmt.Errorf("http status %d for %s", resp.StatusCode, redact(rawURL))
		if resp.StatusCode == http.StatusNotFound || resp.StatusCode >= 500 {
			return nil, unreachable{err}
		}
		return nil, err
	}
	if resp.ContentLength > limit {
		resp.Body.Close()
		cancel()
		return nil, errors.New("content-length exceeds limit")
	}
	return &cancelBody{resp.Body, cancel}, nil
}

type cancelBody struct {
	io.ReadCloser
	cancel context.CancelFunc
}

func (c *cancelBody) Close() error {
	err := c.ReadCloser.Close()
	c.cancel()
	return err
}

type unreachable struct{ err error }

func (u unreachable) Error() string { return u.err.Error() }
func (u unreachable) Unwrap() error { return u.err }

func isUnreachable(err error) bool {
	var u unreachable
	if errors.As(err, &u) {
		return true
	}
	var ne net.Error
	if errors.As(err, &ne) {
		return true
	}
	return errors.Is(err, os.ErrNotExist)
}

// redact drops presigned query strings from log lines.
func redact(raw string) string {
	if i := strings.IndexByte(raw, '?'); i >= 0 {
		return raw[:i] + "?…"
	}
	return raw
}

type catalog struct {
	ManifestURL string `json:"manifest_url"`
	ExpiresIn   int    `json:"expires_in"`
	Files       []struct {
		Path string `json:"path"`
		URL  string `json:"url"`
	} `json:"files"`
}

func looksLikeJSON(b []byte) bool {
	t := bytes.TrimLeft(b, " \t\r\n")
	return len(t) > 0 && t[0] == '{'
}

func parseCatalog(b []byte) (*catalog, error) {
	var c catalog
	if err := json.Unmarshal(b, &c); err != nil {
		return nil, fmt.Errorf("bad release catalog: %w", err)
	}
	if c.ManifestURL == "" {
		return nil, errors.New("release catalog has no manifest_url")
	}
	return &c, nil
}

// releaseName matches goro-<os>-<arch>-<git sha>[.exe].
var releaseName = regexp.MustCompile(`^goro-(linux|windows|darwin)-(amd64|arm64)-[0-9a-f]{7,64}(\.exe)?$`)

// platformFile picks the binary for goos/goarch.
func platformFile(m Manifest, goos, goarch string) (File, bool) {
	for _, f := range m.Files {
		if !releaseName.MatchString(f.Path) {
			continue
		}
		parts := strings.SplitN(f.Path, "-", 4)
		exe := strings.HasSuffix(f.Path, ".exe")
		if parts[1] == goos && parts[2] == goarch && exe == (goos == "windows") {
			return f, true
		}
	}
	return File{}, false
}

func artifactURL(manifestURL string, f File, cat *catalog) (string, error) {
	if cat != nil {
		for _, e := range cat.Files {
			if e.Path == f.Path {
				return e.URL, nil
			}
		}
	}
	base, err := url.Parse(manifestURL)
	if err != nil {
		return "", err
	}
	ref, err := url.Parse(f.URL)
	if err != nil {
		return "", err
	}
	resolved := base.ResolveReference(ref).String()
	if !allowedURL(resolved) {
		return "", fmt.Errorf("artifact url for %s is not allowlisted", f.Path)
	}
	return resolved, nil
}

// allowedURL accepts file:// (tests and offline installs), loopback http,
// the public catalog, and presigned objects under the goro release prefix.
func allowedURL(raw string) bool {
	u, err := url.Parse(raw)
	if err != nil || u.User != nil {
		return false
	}
	switch u.Scheme {
	case "file":
		return u.Host == "" || u.Host == "localhost"
	case "http":
		h := u.Hostname()
		return h == "127.0.0.1" || h == "localhost" || h == "::1"
	case "https":
		if u.Port() != "" && u.Port() != "443" {
			return false
		}
		switch u.Hostname() {
		case "ro.turoran.com":
			return u.Path == "/client/goro/main/latest"
		case "garage.turoran.com":
			name, ok := strings.CutPrefix(u.Path, "/korangar-releases/goro/main/")
			return ok && (name == "manifest.goroup" || releaseName.MatchString(name))
		}
	}
	return false
}

func fileURLPath(u *url.URL) string {
	p := u.Path
	// file:///C:/x -> C:/x on Windows.
	if runtime.GOOS == "windows" && len(p) > 2 && p[0] == '/' && p[2] == ':' {
		p = p[1:]
	}
	return p
}

func appliedBuild(root string) uint64 {
	b, err := os.ReadFile(filepath.Join(root, "updates", "applied.build"))
	if err != nil {
		return 0
	}
	n, _ := strconv.ParseUint(strings.TrimSpace(string(b)), 10, 64)
	return n
}

func runSelfCheck(exe string, want uint64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, exe, SelfCheckFlag)
	cmd.Env = append(os.Environ(), EnvSkip+"=1")
	out, err := cmd.Output()
	if err != nil {
		return err
	}
	return checkSelfCheckOutput(string(out), want)
}

func checkSelfCheckOutput(out string, want uint64) error {
	for _, line := range strings.Split(out, "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 || fields[0] != SelfCheckToken {
			continue
		}
		got, ok := strings.CutPrefix(fields[1], "build=")
		if !ok {
			return errors.New("self-check printed no build")
		}
		if got != strconv.FormatUint(want, 10) {
			return fmt.Errorf("self-check build %s does not match manifest build %d", got, want)
		}
		return nil
	}
	return errors.New("self-check token missing")
}

func cleanupOld(exe string, logf func(string, ...any)) {
	old := exe + ".old"
	if _, err := os.Stat(old); err == nil {
		// On Windows this can fail while the previous process is still
		// exiting; the next launch tries again.
		if err := os.Remove(old); err != nil && logf != nil {
			logf("update: could not remove %s yet: %v", filepath.Base(old), err)
		}
	}
}

func lock(dir string) (func(), error) {
	path := filepath.Join(dir, ".lock")
	for attempt := 0; attempt < 2; attempt++ {
		f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
		if err == nil {
			fmt.Fprintf(f, "%d\n", os.Getpid())
			f.Close()
			return func() { _ = os.Remove(path) }, nil
		}
		if st, serr := os.Stat(path); serr == nil && time.Since(st.ModTime()) > lockStale {
			_ = os.Remove(path)
			continue
		}
		return nil, errors.New("another goro is applying an update")
	}
	return nil, errors.New("could not take update lock")
}

func writeAtomic(path string, data []byte) error {
	tmp := path + ".partial"
	f, err := os.Create(tmp)
	if err != nil {
		return err
	}
	if _, err := f.Write(data); err != nil {
		f.Close()
		return err
	}
	if err := f.Sync(); err != nil {
		f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}
