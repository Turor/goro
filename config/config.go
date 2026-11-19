package config

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/kivutar/goro/glog"
)

type Config struct {
	// ConfigPath is the absolute file path selected by LoadConfig for saving settings.
	ConfigPath string
	Headless   bool
	DataDir    string
	// AIStateDir holds writable AI overrides when a platform uses read-only assets.
	// Empty keeps the desktop behavior of writing inside DataDir.
	AIStateDir    string
	Window        WindowConfig
	Packet        PacketConfig
	Login         LoginConfig
	Audio         AudioConfig
	Render        RenderConfig
	Network       NetworkConfig
	Fog           FogConfig
	Gameplay      GameplayConfig
	Script        ScriptConfig
	Log           glog.LogConfig
	ChatShortcuts ChatShortcuts
	UI            UISettings
	UILayout      UILayoutConfig
	Data          DataConfig
}

// DataConfig selects streaming of client data from an asset server.
type DataConfig struct {
	// Remote is "auto" (stream when the data directory has no archives),
	// "on" (also stream files missing from local archives) or "off".
	Remote string
	// RemoteURL is the asset server root (remoteclient2). Empty disables
	// streaming. It is never committed: it comes from data-server.ini, goro.ini,
	// --data-url or a release build's -X config.DefaultDataRemoteURL.
	RemoteURL string
	// CacheDir keeps streamed files. Empty means the user cache directory.
	CacheDir string
	// ManifestMode is "report" (log mismatches, do not block) or "enforce"
	// (fail closed after a manifest has been trusted). Empty means report.
	ManifestMode string
}

// Remote data modes.
const (
	DataRemoteAuto = "auto"
	DataRemoteOn   = "on"
	DataRemoteOff  = "off"

	// DataManifestReport logs signed-manifest mismatches and does not block.
	// DataManifestEnforce fails closed once a manifest has been trusted.
	DataManifestReport  = "report"
	DataManifestEnforce = "enforce"
)

// DefaultDataRemoteURL is the asset server a release build streams from when
// no config file names one. It is empty in source and set only at link time
// (go build -ldflags "-X github.com/kivutar/goro/config.DefaultDataRemoteURL=...")
// from operator config outside the repository.
var DefaultDataRemoteURL = ""

// DataServerFile is the optional, uncommitted file that names the asset
// server ([data] remote_url). It is read next to the executable, in the working
// directory and in --data-dir, before goro.ini; only its [data] section counts.
// See data-server.example.ini.
const DataServerFile = "data-server.ini"

// Streaming reports whether an asset server is configured and enabled.
func (d DataConfig) Streaming() bool {
	return d.Remote != DataRemoteOff && strings.TrimSpace(d.RemoteURL) != ""
}

// RemoteCacheDir returns where streamed files are kept.
func (d DataConfig) RemoteCacheDir() (string, error) {
	if d.CacheDir != "" {
		return filepath.Abs(d.CacheDir)
	}
	base, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	host := "default"
	if u, err := url.Parse(d.RemoteURL); err == nil && u.Host != "" {
		host = strings.NewReplacer(":", "_", "/", "_").Replace(u.Host)
	}
	return filepath.Join(base, "goro", "remote", host), nil
}

// UILayoutConfig selects the UI layout and theme override file
// (ui-layout.ini). Players need none of it: without the file the built-in
// layout is used and nothing is watched.
type UILayoutConfig struct {
	// Path is the override file. Empty means ui-layout.ini next to goro.ini.
	Path string
	// Watch polls the file and applies edits while the game runs (--ui-dev).
	Watch bool
	// DumpPath, when set, writes the full default layout there and exits.
	DumpPath string
	// Theme names the palette theme under ui-layout.ini: a built-in theme
	// (default, dark, complementary, analogous, triadic) or <name>.ini in the
	// themes folder. Empty means the default theme. The settings window saves
	// it as [ui_layout] theme.
	Theme string
}

// ThemesDirName is the folder next to goro.ini that holds user theme files.
const ThemesDirName = "themes"

// maxThemeNameLen bounds a saved theme name.
const maxThemeNameLen = 40

// ThemesDir returns the folder searched for user theme files.
func (cfg Config) ThemesDir() string {
	dir := "."
	if cfg.ConfigPath != "" {
		dir = filepath.Dir(cfg.ConfigPath)
	}
	return filepath.Join(dir, ThemesDirName)
}

// validateThemeName accepts the names a theme file may have: lower-case
// letters, digits, '_' and '-'. Empty means the default theme.
func validateThemeName(name string) error {
	if len(name) > maxThemeNameLen {
		return fmt.Errorf("ui_layout theme is longer than %d characters", maxThemeNameLen)
	}
	for _, r := range name {
		if !(r >= 'a' && r <= 'z' || r >= '0' && r <= '9' || r == '_' || r == '-') {
			return fmt.Errorf("ui_layout theme %q may use only a-z, 0-9, '_' and '-'", name)
		}
	}
	return nil
}

// SaveUITheme remembers the selected theme in goro.ini ([ui_layout] theme).
func (cfg Config) SaveUITheme(name string) (string, error) {
	name = strings.ToLower(strings.TrimSpace(name))
	if err := validateThemeName(name); err != nil {
		return "", err
	}
	return cfg.saveConfigValues(map[string]map[string]string{
		"uilayout": {"theme": name},
	})
}

// DefaultUILayoutFile is the override file name looked up next to goro.ini.
const DefaultUILayoutFile = "ui-layout.ini"

// UILayoutPath returns the override file to read.
func (cfg Config) UILayoutPath() string {
	if cfg.UILayout.Path != "" {
		return cfg.UILayout.Path
	}
	dir := "."
	if cfg.ConfigPath != "" {
		dir = filepath.Dir(cfg.ConfigPath)
	}
	return filepath.Join(dir, DefaultUILayoutFile)
}

type WindowConfig struct {
	Title      string
	Width      int
	Height     int
	Fullscreen bool
}

type PacketConfig struct {
	ClientDate int
	// ClientDateExplicit is true when client_date came from a config file or
	// --packet-client-date. When it is false, a data/clientinfo.xml
	// <packet_version> is preferred over the fork default.
	ClientDateExplicit bool
	Profile            int
}

// DefaultPacketClientDate is this fork's packet profile when neither the
// config nor data/clientinfo.xml selects a date. 20080910 remains available
// via --packet-client-date or [packet] client_date.
const DefaultPacketClientDate = 20250716

type LoginConfig struct {
	Username       string
	Password       string
	AutoLogin      bool
	ServerSlot     int
	CharServerSlot int
	CharSlot       int
	KeepID         bool
	SavedUsername  string
}

type AudioConfig struct {
	Disabled  bool
	BGM       bool
	BGMVolume float64
	SFXVolume float64
}

// DefaultMaxFPS caps rendering when no max_fps is configured. It covers
// 60-144 Hz displays; with vsync on, a cap selects mailbox presentation.
const DefaultMaxFPS = 144

// MaxAllowedFPS is the highest accepted frame cap. 0 remains unlimited.
const MaxAllowedFPS = 1000

type RenderConfig struct {
	GraphicsAPI        string
	Anisotropy         int
	TextureUpscaling   bool
	SmoothSprites      bool
	MSAA               bool
	Bloom              bool
	SSAO               bool
	FXAA               bool
	EnhancedLighting   bool
	VSync              bool
	MaxFPS             int
	FPS                bool
	NoUI               bool
	AsyncUI            bool
	UIProfile          bool
	BenchSeconds       int
	BenchWarmupSeconds int
	CPUProfile         string
	Stats              bool
	WorldDebugStats    bool
}

type NetworkConfig struct {
	Trace bool
	// Capture is a JSON path for framed s2c packets (Phase 5 golden replay).
	// Empty disables recording. Set with --capture or GORO_CAPTURE.
	Capture string
}

type FogConfig struct {
	Enabled bool
}

type GameplayConfig struct {
	NoShift     bool
	NoCtrl      bool
	LessEffects bool
	NoEffects   bool
	SimpleAura  bool
	SnapTargets bool
	SnapItems   bool
	ForceUserAI bool
}

type ScriptConfig struct {
	Path string
}

func LoadConfig(args []string) (Config, error) {
	// Parse once to find explicitly supplied paths, using the same flag rules as
	// the final pass. Defer validation until file settings have also been loaded.
	cli := defaultConfig()
	if err := parseCLI(&cli, args); err != nil {
		return Config{}, err
	}
	cfg := defaultConfig()

	// Defaults < data-server.ini (executable dir, ./, --data-dir) < ./goro.ini
	// < --data-dir/goro.ini < --config < command-line flags.
	for _, path := range dataServerPaths(cli.DataDir) {
		if err := applyDataServerFile(&cfg, path); err != nil {
			return Config{}, err
		}
	}
	cfg.ConfigPath = "goro.ini"
	if err := applyINIFile(&cfg, cfg.ConfigPath, false); err != nil {
		return Config{}, err
	}
	if cli.DataDir != "" {
		cfg.ConfigPath = filepath.Join(cli.DataDir, "goro.ini")
		if err := applyINIFile(&cfg, cfg.ConfigPath, false); err != nil {
			return Config{}, err
		}
	}
	if cli.ConfigPath != "" {
		cfg.ConfigPath = cli.ConfigPath
		if err := applyINIFile(&cfg, cfg.ConfigPath, true); err != nil {
			return Config{}, err
		}
	}
	if err := applyCLI(&cfg, args); err != nil {
		return Config{}, err
	}
	path, err := filepath.Abs(cfg.ConfigPath)
	if err != nil {
		return Config{}, err
	}
	cfg.ConfigPath = path
	cfg.DataDir = resolveDataDir(cfg.DataDir)
	return cfg, nil
}

type UserSettings struct {
	Fullscreen       bool
	VSync            bool
	MaxFPS           int
	FPS              bool
	Anisotropy       int
	TextureUpscaling bool
	SmoothSprites    bool
	MSAA             bool
	Bloom            bool
	SSAO             bool
	FXAA             bool
	EnhancedLighting bool
	BGMVolume        float64
	SFXVolume        float64
	NoShift          bool
	NoCtrl           bool
	LessEffects      bool
	SnapTargets      bool
	SnapItems        bool
}

func UserDataDir() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "goro"), nil
}

func NextScreenshotPath(now time.Time) (string, error) {
	dir, err := UserDataDir()
	if err != nil {
		return "", err
	}
	dir = filepath.Join(dir, "screenshots")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	name := fmt.Sprintf("goro-%s.png", now.Format("20060102-150405"))
	path := filepath.Join(dir, name)
	for i := 2; ; i++ {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			return path, nil
		} else if err != nil {
			return "", err
		}
		name = fmt.Sprintf("goro-%s-%02d.png", now.Format("20060102-150405"), i)
		path = filepath.Join(dir, name)
	}
}

func (cfg Config) SaveVisualSettings(fogEnabled, simpleAura, noEffects bool) (string, error) {
	return cfg.saveConfigValues(map[string]map[string]string{
		"fog": {
			"enabled": formatINIValueBool(fogEnabled),
		},
		"gameplay": {
			"simple_aura": formatINIValueBool(simpleAura),
			"no_effects":  formatINIValueBool(noEffects),
		},
	})
}

func (cfg Config) SaveUserSettings(settings UserSettings) (string, error) {
	if err := validateAnisotropy(settings.Anisotropy); err != nil {
		return "", err
	}
	if err := validateMaxFPS(settings.MaxFPS); err != nil {
		return "", err
	}
	if settings.BGMVolume < 0 || settings.BGMVolume > 1 {
		return "", fmt.Errorf("bgm volume must be between 0 and 1")
	}
	if settings.SFXVolume < 0 || settings.SFXVolume > 1 {
		return "", fmt.Errorf("sfx volume must be between 0 and 1")
	}
	values := map[string]map[string]string{
		"window": {
			"fullscreen": formatINIValueBool(settings.Fullscreen),
		},
		"render": {
			"anisotropy":        strconv.Itoa(settings.Anisotropy),
			"texture_upscaling": formatINIValueBool(settings.TextureUpscaling),
			"smooth_sprites":    formatINIValueBool(settings.SmoothSprites),
			"msaa":              formatINIValueBool(settings.MSAA),
			"bloom":             formatINIValueBool(settings.Bloom),
			"ssao":              formatINIValueBool(settings.SSAO),
			"fxaa":              formatINIValueBool(settings.FXAA),
			"enhanced_lighting": formatINIValueBool(settings.EnhancedLighting),
			"vsync":             formatINIValueBool(settings.VSync),
			"max_fps":           strconv.Itoa(settings.MaxFPS),
			"fps":               formatINIValueBool(settings.FPS),
		},
		"audio": {
			"bgm_volume": formatINIValueFloat(settings.BGMVolume),
			"sfx_volume": formatINIValueFloat(settings.SFXVolume),
		},
		"gameplay": {
			"no_shift":     formatINIValueBool(settings.NoShift),
			"no_ctrl":      formatINIValueBool(settings.NoCtrl),
			"less_effects": formatINIValueBool(settings.LessEffects),
			"snap":         formatINIValueBool(settings.SnapTargets),
			"itemsnap":     formatINIValueBool(settings.SnapItems),
		},
	}
	return cfg.saveConfigValues(values)
}

// SaveScriptPath saves an explicit script choice independently of other settings.
func (cfg Config) SaveScriptPath(path string) (string, error) {
	if strings.ContainsAny(path, "\r\n\x00") {
		return "", fmt.Errorf("script path must be a single line without NUL characters")
	}
	return cfg.saveConfigValues(map[string]map[string]string{
		"script": {"path": `"` + path + `"`},
	})
}

// SaveLoginID remembers only the ID, independently of explicit login credentials.
func (cfg Config) SaveLoginID(username string, keep bool) (string, error) {
	if !keep {
		username = ""
	}
	if strings.ContainsAny(username, "\r\n\x00") {
		return "", fmt.Errorf("login ID must be a single line without NUL characters")
	}
	return cfg.saveConfigValues(map[string]map[string]string{
		"login": {
			"keep_id":        formatINIValueBool(keep),
			"saved_username": `"` + username + `"`,
		},
	})
}

func (cfg Config) saveConfigValues(values map[string]map[string]string) (string, error) {
	path := cfg.ConfigPath
	if path == "" {
		return "", fmt.Errorf("no config file selected")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return "", err
	}
	existing, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return "", err
	}
	data := upsertINIValues(string(existing), values)
	if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
		return "", err
	}
	return path, nil
}

func defaultConfig() Config {
	return Config{
		Data:          DataConfig{Remote: DataRemoteAuto, RemoteURL: DefaultDataRemoteURL, ManifestMode: DataManifestReport},
		ChatShortcuts: defaultChatShortcuts(),
		Window: WindowConfig{
			Title:      "goro",
			Width:      1280,
			Height:     720,
			Fullscreen: false,
		},
		Packet: PacketConfig{
			ClientDate: DefaultPacketClientDate,
			Profile:    23,
		},
		Login: LoginConfig{
			CharSlot: -1,
		},
		Audio: AudioConfig{
			BGM:       true,
			BGMVolume: 0.55,
			SFXVolume: 0.55,
		},
		Render: RenderConfig{
			GraphicsAPI:        "vulkan",
			Anisotropy:         8,
			SmoothSprites:      true,
			AsyncUI:            true,
			VSync:              true,
			MaxFPS:             DefaultMaxFPS,
			BenchWarmupSeconds: 0,
		},
		Fog: FogConfig{
			Enabled: true,
		},
		Gameplay: GameplayConfig{
			NoCtrl: true,
		},
		Log: glog.LogConfig{
			Level: "info",
		},
	}
}

// dataServerPaths lists the data-server.ini candidates, lowest priority first,
// without duplicates.
func dataServerPaths(dataDir string) []string {
	var dirs []string
	if exe, err := os.Executable(); err == nil {
		if resolved, err := filepath.EvalSymlinks(exe); err == nil {
			exe = resolved
		}
		dirs = append(dirs, filepath.Dir(exe))
	}
	dirs = append(dirs, ".")
	if dataDir != "" {
		dirs = append(dirs, dataDir)
	}
	seen := map[string]bool{}
	var paths []string
	for _, dir := range dirs {
		abs, err := filepath.Abs(dir)
		if err != nil || seen[abs] {
			continue
		}
		seen[abs] = true
		paths = append(paths, filepath.Join(abs, DataServerFile))
	}
	return paths
}

// applyDataServerFile applies the [data] section of a data-server.ini; a
// missing file is fine, other sections are rejected so the file cannot carry
// unrelated settings.
func applyDataServerFile(cfg *Config, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("open %s: %w", path, err)
	}
	section := ""
	for i, raw := range strings.Split(string(data), "\n") {
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			section = normalizeKey(strings.TrimSpace(line[1 : len(line)-1]))
			continue
		}
		if section != "data" {
			return fmt.Errorf("%s line %d: only the [data] section is allowed", path, i+1)
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			return fmt.Errorf("%s line %d: expected key=value", path, i+1)
		}
		if err := applyConfigValue(cfg, section, normalizeKey(key), cleanINIValue(value)); err != nil {
			return fmt.Errorf("%s line %d: %w", path, i+1, err)
		}
	}
	return validateConfig(cfg)
}

func applyINIFile(cfg *Config, path string, explicit bool) error {
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) && !explicit {
			return nil
		}
		return fmt.Errorf("open config %s: %w", path, err)
	}
	defer file.Close()

	layoutBefore := cfg.UILayout.Path
	if err := applyINI(cfg, file); err != nil {
		return fmt.Errorf("read config %s: %w", path, err)
	}
	// A relative [ui_layout] file is relative to the INI that names it, not
	// to the working directory. --ui-layout stays relative to the CWD.
	if layout := cfg.UILayout.Path; layout != layoutBefore && layout != "" && !filepath.IsAbs(layout) {
		cfg.UILayout.Path = filepath.Join(filepath.Dir(path), layout)
	}
	return nil
}

func parseCLI(cfg *Config, args []string) error {
	fs := flag.NewFlagSet("goro", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	windowed := false
	fs.StringVar(&cfg.ConfigPath, "config", cfg.ConfigPath, "path to goro ini configuration (also used to save settings)")
	fs.BoolVar(&cfg.Headless, "headless", false, "run without a window or audio (implies autologin)")
	fs.StringVar(&cfg.DataDir, "data-dir", cfg.DataDir, "Ragnarok data directory (loads goro.ini from this directory)")
	fs.StringVar(&cfg.Data.Remote, "data-remote", cfg.Data.Remote, "stream client data from the asset server: auto (when there are no local archives), on, off")
	fs.StringVar(&cfg.Data.RemoteURL, "data-url", cfg.Data.RemoteURL, "asset server URL for streamed client data")
	fs.StringVar(&cfg.Data.CacheDir, "data-cache", cfg.Data.CacheDir, "directory for streamed client data (default: user cache directory)")
	fs.StringVar(&cfg.Data.ManifestMode, "data-manifest-mode", cfg.Data.ManifestMode, "signed data manifest: report (log mismatches) or enforce (fail closed)")
	fs.StringVar(&cfg.Window.Title, "title", cfg.Window.Title, "window title")
	fs.IntVar(&cfg.Window.Width, "width", cfg.Window.Width, "window width")
	fs.IntVar(&cfg.Window.Height, "height", cfg.Window.Height, "window height")
	fs.BoolVar(&cfg.Window.Fullscreen, "fullscreen", cfg.Window.Fullscreen, "start in fullscreen mode")
	fs.BoolVar(&windowed, "windowed", false, "force windowed mode")
	fs.IntVar(&cfg.Packet.ClientDate, "packet-client-date", cfg.Packet.ClientDate, "packet client date")
	fs.IntVar(&cfg.Packet.Profile, "packet-profile", cfg.Packet.Profile, "packet profile")
	fs.StringVar(&cfg.Login.Username, "username", cfg.Login.Username, "login username")
	fs.StringVar(&cfg.Login.Password, "password", cfg.Login.Password, "login password")
	fs.BoolVar(&cfg.Login.AutoLogin, "autologin", cfg.Login.AutoLogin, "attempt login automatically")
	fs.IntVar(&cfg.Login.ServerSlot, "server-slot", cfg.Login.ServerSlot, "login server entry in clientinfo.xml for autologin, starting at 0")
	fs.IntVar(&cfg.Login.CharServerSlot, "char-server-slot", cfg.Login.CharServerSlot, "character server entry returned after login for autologin, starting at 0")
	fs.IntVar(&cfg.Login.CharSlot, "char-slot", cfg.Login.CharSlot, "character slot to select after autologin, 0 to 8")
	fs.BoolVar(&cfg.Audio.BGM, "bgm", cfg.Audio.BGM, "enable BGM")
	fs.BoolVar(&cfg.Audio.Disabled, "no-audio", cfg.Audio.Disabled, "disable all audio output")
	fs.Float64Var(&cfg.Audio.BGMVolume, "bgm-volume", cfg.Audio.BGMVolume, "BGM volume from 0 to 1")
	fs.Float64Var(&cfg.Audio.SFXVolume, "sfx-volume", cfg.Audio.SFXVolume, "SFX volume from 0 to 1")
	fs.StringVar(&cfg.Render.GraphicsAPI, "graphics-api", cfg.Render.GraphicsAPI, "graphics API: auto, vulkan, dx12, metal, gles, software")
	fs.IntVar(&cfg.Render.Anisotropy, "anisotropy", cfg.Render.Anisotropy, "anisotropic texture filtering: 0 (off), 2, 4, 8, 16")
	fs.BoolVar(&cfg.Render.TextureUpscaling, "texture-upscaling", cfg.Render.TextureUpscaling, "upscale terrain, model, and world sprite textures 2x with xBR")
	fs.BoolVar(&cfg.Render.SSAO, "ssao", cfg.Render.SSAO, "enable subtle ambient occlusion for terrain and models")
	fs.BoolVar(&cfg.Render.FXAA, "fxaa", cfg.Render.FXAA, "enable FXAA on the resolved world color")
	fs.BoolVar(&cfg.Render.EnhancedLighting, "enhanced-lighting", cfg.Render.EnhancedLighting, "shade the world from the map sun, tiled point lights, and a directional shadow")
	fs.BoolVar(&cfg.Render.Bloom, "bloom", cfg.Render.Bloom, "enable subtle bloom for bright world pixels")
	fs.BoolVar(&cfg.Render.MSAA, "msaa", cfg.Render.MSAA, "enable 4x multisample anti-aliasing for the world")
	fs.BoolVar(&cfg.Render.SmoothSprites, "smooth-sprites", cfg.Render.SmoothSprites, "smooth world sprites (false uses nearest-neighbor filtering)")
	fs.BoolVar(&cfg.Render.VSync, "vsync", cfg.Render.VSync, "enable vsync")
	fs.IntVar(&cfg.Render.MaxFPS, "max-fps", cfg.Render.MaxFPS, "cap the frame rate: 0 for unlimited, or 1-1000 (with vsync, a cap selects low-latency mailbox presentation)")
	fs.BoolVar(&cfg.Render.FPS, "fps", cfg.Render.FPS, "show measured FPS counter")
	fs.BoolVar(&cfg.Render.NoUI, "no-ui", cfg.Render.NoUI, "disable UI rendering for benchmarking")
	fs.BoolVar(&cfg.Render.AsyncUI, "async-ui", cfg.Render.AsyncUI, "rasterize UI off the draw thread")
	fs.BoolVar(&cfg.Render.UIProfile, "profile-ui", cfg.Render.UIProfile, "log aggregate UI frame and redraw timings")
	fs.IntVar(&cfg.Render.BenchSeconds, "bench-seconds", cfg.Render.BenchSeconds, "quit after benchmarking for this many seconds")
	fs.IntVar(&cfg.Render.BenchWarmupSeconds, "bench-warmup-seconds", cfg.Render.BenchWarmupSeconds, "benchmark warmup seconds")
	fs.StringVar(&cfg.Render.CPUProfile, "cpu-profile", cfg.Render.CPUProfile, "write CPU profile to this path during benchmark")
	fs.BoolVar(&cfg.Render.Stats, "render-stats", cfg.Render.Stats, "show render stats")
	fs.BoolVar(&cfg.Render.WorldDebugStats, "world-debug-stats", cfg.Render.WorldDebugStats, "show world renderer debug stats")
	fs.BoolVar(&cfg.Network.Trace, "net-trace", cfg.Network.Trace, "log network reads and writes")
	fs.StringVar(&cfg.Network.Capture, "capture", cfg.Network.Capture, "write framed server packets to this JSON capture (golden replay)")
	fs.BoolVar(&cfg.Fog.Enabled, "fog", cfg.Fog.Enabled, "enable map fog")
	fs.BoolVar(&cfg.Gameplay.NoShift, "no-shift", cfg.Gameplay.NoShift, "allow support skills to target enemies without holding Shift")
	fs.BoolVar(&cfg.Gameplay.NoCtrl, "no-ctrl", cfg.Gameplay.NoCtrl, "keep attacking with one click without holding Ctrl")
	fs.BoolVar(&cfg.Gameplay.LessEffects, "mineffect", cfg.Gameplay.LessEffects, "use simpler visual effects")
	fs.BoolVar(&cfg.Gameplay.NoEffects, "no-effects", cfg.Gameplay.NoEffects, "hide optional visual effects")
	fs.BoolVar(&cfg.Gameplay.SnapTargets, "snap", cfg.Gameplay.SnapTargets, "magnetize attack and enemy skill cursors to targets")
	fs.BoolVar(&cfg.Gameplay.SnapItems, "itemsnap", cfg.Gameplay.SnapItems, "magnetize pickup cursor to floor items")
	fs.BoolVar(&cfg.Gameplay.SnapItems, "item-snap", cfg.Gameplay.SnapItems, "magnetize pickup cursor to floor items")
	fs.BoolVar(&cfg.Gameplay.ForceUserAI, "force-user-ai", cfg.Gameplay.ForceUserAI, "start companion AI in USER_AI/USER_AI_M custom mode")
	fs.StringVar(&cfg.Script.Path, "script", cfg.Script.Path, "Lua script to run while in game")
	fs.StringVar(&cfg.Log.Level, "log-level", cfg.Log.Level, "minimum log level: debug, info, warn, error, fatal")
	fs.StringVar(&cfg.Log.File, "log-file", cfg.Log.File, "append logs to this file")
	fs.StringVar(&cfg.UILayout.Path, "ui-layout", cfg.UILayout.Path, "UI layout and theme override file (default ui-layout.ini next to goro.ini)")
	fs.BoolVar(&cfg.UILayout.Watch, "ui-dev", cfg.UILayout.Watch, "apply UI layout file edits while running; F10 reloads")
	fs.StringVar(&cfg.UILayout.DumpPath, "dump-ui-layout", cfg.UILayout.DumpPath, "write every UI layout key with its default to this path (- for stdout) and exit")

	if err := fs.Parse(args); err != nil {
		return err
	}
	var pathErr error
	fs.Visit(func(f *flag.Flag) {
		if (f.Name == "config" || f.Name == "data-dir") && f.Value.String() == "" {
			pathErr = fmt.Errorf("--%s requires a non-empty path", f.Name)
		}
		if f.Name == "packet-client-date" {
			cfg.Packet.ClientDateExplicit = true
		}
	})
	if pathErr != nil {
		return pathErr
	}
	if windowed {
		cfg.Window.Fullscreen = false
	}
	if cfg.Network.Capture == "" {
		cfg.Network.Capture = strings.TrimSpace(os.Getenv("GORO_CAPTURE"))
	}
	return nil
}

func applyCLI(cfg *Config, args []string) error {
	if err := parseCLI(cfg, args); err != nil {
		return err
	}
	if cfg.Headless {
		cfg.Login.AutoLogin = true
		cfg.Audio.Disabled = true
		username := cfg.Login.Username
		if username == "" && cfg.Login.KeepID {
			username = cfg.Login.SavedUsername
		}
		if strings.TrimSpace(username) == "" || cfg.Login.Password == "" {
			return fmt.Errorf("headless mode requires a login ID and password")
		}
		if cfg.Login.CharSlot < 0 {
			return fmt.Errorf("headless mode requires --char-slot (0 to 8)")
		}
	}
	return validateConfig(cfg)
}

func applyINI(cfg *Config, r io.Reader) error {
	scanner := bufio.NewScanner(r)
	section := ""
	lineNo := 0
	for scanner.Scan() {
		lineNo++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			section = normalizeKey(strings.TrimSpace(line[1 : len(line)-1]))
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			return fmt.Errorf("line %d: expected key=value", lineNo)
		}
		if err := applyConfigValue(cfg, section, normalizeKey(key), cleanINIValue(value)); err != nil {
			return fmt.Errorf("line %d: %w", lineNo, err)
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	return validateConfig(cfg)
}

func applyConfigValue(cfg *Config, section, key, value string) error {
	if section == "ui" {
		return cfg.UI.apply(key, value)
	}
	if section == "chatshortcuts" {
		digit, err := strconv.Atoi(key)
		if err != nil || digit < 0 || digit > 9 {
			return fmt.Errorf("invalid chat shortcut key %q", key)
		}
		cfg.ChatShortcuts[(digit+9)%10] = value
		return nil
	}
	switch section + "." + key {
	case ".datadir", "data.dir", "data.datadir", "config.datadir", "core.datadir":
		cfg.DataDir = value
	case "data.remote":
		cfg.Data.Remote = strings.ToLower(strings.TrimSpace(value))
	case "data.remoteurl", "data.url":
		cfg.Data.RemoteURL = value
	case "data.cachedir", "data.cache":
		cfg.Data.CacheDir = value
	case "data.manifestmode":
		cfg.Data.ManifestMode = strings.ToLower(strings.TrimSpace(value))
	case "window.title":
		cfg.Window.Title = value
	case "window.width":
		return setInt(value, &cfg.Window.Width)
	case "window.height":
		return setInt(value, &cfg.Window.Height)
	case "window.fullscreen":
		return setBool(value, &cfg.Window.Fullscreen)
	case "packet.clientdate":
		if err := setInt(value, &cfg.Packet.ClientDate); err != nil {
			return err
		}
		cfg.Packet.ClientDateExplicit = true
		return nil
	case "packet.profile":
		return setInt(value, &cfg.Packet.Profile)
	case "login.username":
		cfg.Login.Username = value
	case "login.password":
		cfg.Login.Password = value
	case "login.autologin":
		return setBool(value, &cfg.Login.AutoLogin)
	case "login.serverslot":
		return setInt(value, &cfg.Login.ServerSlot)
	case "login.charserverslot":
		return setInt(value, &cfg.Login.CharServerSlot)
	case "login.charslot":
		return setInt(value, &cfg.Login.CharSlot)
	case "login.keepid":
		return setBool(value, &cfg.Login.KeepID)
	case "login.savedusername":
		cfg.Login.SavedUsername = value
	case "audio.bgm":
		return setBool(value, &cfg.Audio.BGM)
	case "audio.noaudio":
		return setBool(value, &cfg.Audio.Disabled)
	case "audio.bgmvolume":
		return setFloat(value, &cfg.Audio.BGMVolume)
	case "audio.sfxvolume":
		return setFloat(value, &cfg.Audio.SFXVolume)
	case "render.graphicsapi":
		cfg.Render.GraphicsAPI = value
	case "render.anisotropy":
		return setInt(value, &cfg.Render.Anisotropy)
	case "render.textureupscaling":
		return setBool(value, &cfg.Render.TextureUpscaling)
	case "render.ssao":
		return setBool(value, &cfg.Render.SSAO)
	case "render.fxaa":
		return setBool(value, &cfg.Render.FXAA)
	case "render.enhancedlighting":
		return setBool(value, &cfg.Render.EnhancedLighting)
	case "render.bloom":
		return setBool(value, &cfg.Render.Bloom)
	case "render.msaa":
		return setBool(value, &cfg.Render.MSAA)
	case "render.smoothsprites":
		return setBool(value, &cfg.Render.SmoothSprites)
	case "render.vsync":
		return setBool(value, &cfg.Render.VSync)
	case "render.maxfps":
		return setInt(value, &cfg.Render.MaxFPS)
	case "render.fps":
		return setBool(value, &cfg.Render.FPS)
	case "render.noui":
		return setBool(value, &cfg.Render.NoUI)
	case "render.asyncui":
		return setBool(value, &cfg.Render.AsyncUI)
	case "render.profileui":
		return setBool(value, &cfg.Render.UIProfile)
	case "render.benchseconds":
		return setInt(value, &cfg.Render.BenchSeconds)
	case "render.benchwarmupseconds":
		return setInt(value, &cfg.Render.BenchWarmupSeconds)
	case "render.cpuprofile":
		cfg.Render.CPUProfile = value
	case "render.stats":
		return setBool(value, &cfg.Render.Stats)
	case "render.worlddebugstats":
		return setBool(value, &cfg.Render.WorldDebugStats)
	case "network.trace":
		return setBool(value, &cfg.Network.Trace)
	case "network.capture":
		cfg.Network.Capture = value
	case "fog.enabled":
		return setBool(value, &cfg.Fog.Enabled)
	case "gameplay.noshift":
		return setBool(value, &cfg.Gameplay.NoShift)
	case "gameplay.noctrl":
		return setBool(value, &cfg.Gameplay.NoCtrl)
	case "gameplay.lesseffects", "gameplay.lesseffect", "gameplay.mineffect", "gameplay.less_effects", "gameplay.less_effect":
		return setBool(value, &cfg.Gameplay.LessEffects)
	case "gameplay.simpleaura":
		return setBool(value, &cfg.Gameplay.SimpleAura)
	case "gameplay.noeffects":
		return setBool(value, &cfg.Gameplay.NoEffects)
	case "gameplay.snap", "gameplay.snaptargets", "gameplay.targetsnap":
		return setBool(value, &cfg.Gameplay.SnapTargets)
	case "gameplay.itemsnap", "gameplay.snapitems", "gameplay.itemsnapping":
		return setBool(value, &cfg.Gameplay.SnapItems)
	case "gameplay.forceuserai":
		return setBool(value, &cfg.Gameplay.ForceUserAI)
	case ".script", "script.path":
		cfg.Script.Path = value
	case "log.level":
		cfg.Log.Level = strings.ToLower(value)
	case "log.file":
		cfg.Log.File = value
	case "uilayout.file":
		cfg.UILayout.Path = value
	case "uilayout.watch":
		return setBool(value, &cfg.UILayout.Watch)
	case "uilayout.theme":
		name := strings.ToLower(strings.TrimSpace(value))
		if err := validateThemeName(name); err != nil {
			return err
		}
		cfg.UILayout.Theme = name
	default:
		return fmt.Errorf("unknown key %q in section %q", key, section)
	}
	return nil
}

// ResolvePacketClientDate returns the packet profile to use at startup.
// An explicit config or --packet-client-date value wins. Otherwise a positive
// clientinfo <packet_version> is used when present. The fork default is
// 20250716; 20080910 is still available when selected explicitly.
func (cfg Config) ResolvePacketClientDate(clientInfoPacketVersion int) int {
	if cfg.Packet.ClientDateExplicit {
		return cfg.Packet.ClientDate
	}
	if clientInfoPacketVersion > 0 {
		return clientInfoPacketVersion
	}
	if cfg.Packet.ClientDate > 0 {
		return cfg.Packet.ClientDate
	}
	return DefaultPacketClientDate
}

func validateConfig(cfg *Config) error {
	if err := validateAnisotropy(cfg.Render.Anisotropy); err != nil {
		return err
	}
	if err := validateMaxFPS(cfg.Render.MaxFPS); err != nil {
		return err
	}
	if cfg.Window.Width <= 0 {
		return fmt.Errorf("window width must be positive")
	}
	if cfg.Window.Height <= 0 {
		return fmt.Errorf("window height must be positive")
	}
	if cfg.Packet.ClientDate <= 0 {
		return fmt.Errorf("packet client date must be positive")
	}
	if cfg.Login.CharSlot < -1 || cfg.Login.CharSlot > 8 {
		return fmt.Errorf("character slot must be between 0 and 8")
	}
	if cfg.Login.ServerSlot < 0 {
		return fmt.Errorf("--server-slot must be non-negative")
	}
	if cfg.Login.CharServerSlot < 0 {
		return fmt.Errorf("--char-server-slot must be non-negative")
	}
	if cfg.Audio.BGMVolume < 0 || cfg.Audio.BGMVolume > 1 {
		return fmt.Errorf("bgm volume must be between 0 and 1")
	}
	if cfg.Audio.SFXVolume < 0 || cfg.Audio.SFXVolume > 1 {
		return fmt.Errorf("sfx volume must be between 0 and 1")
	}
	if cfg.Render.BenchSeconds < 0 || cfg.Render.BenchWarmupSeconds < 0 {
		return fmt.Errorf("benchmark durations must be non-negative")
	}
	switch cfg.Data.Remote {
	case DataRemoteAuto, DataRemoteOn, DataRemoteOff:
	default:
		return fmt.Errorf("data remote must be auto, on or off, not %q", cfg.Data.Remote)
	}
	if cfg.Data.Remote == DataRemoteOn && strings.TrimSpace(cfg.Data.RemoteURL) == "" {
		return fmt.Errorf("data remote = on needs remote_url (see %s)", DataServerFile)
	}
	if cfg.Data.Streaming() {
		u, err := url.Parse(cfg.Data.RemoteURL)
		if err != nil || (u.Scheme != "https" && u.Scheme != "http") || u.Host == "" {
			return fmt.Errorf("invalid data URL %q", cfg.Data.RemoteURL)
		}
	}
	switch strings.ToLower(strings.TrimSpace(cfg.Data.ManifestMode)) {
	case "", DataManifestReport:
		cfg.Data.ManifestMode = DataManifestReport
	case DataManifestEnforce:
		cfg.Data.ManifestMode = DataManifestEnforce
	default:
		return fmt.Errorf("data manifest mode must be report or enforce, not %q", cfg.Data.ManifestMode)
	}
	switch cfg.Log.Level {
	case "debug", "info", "warn", "warning", "error", "fatal":
	default:
		return fmt.Errorf("invalid log level %q", cfg.Log.Level)
	}
	return nil
}

func validateAnisotropy(level int) error {
	switch level {
	case 0, 2, 4, 8, 16:
		return nil
	default:
		return fmt.Errorf("anisotropy must be 0 (off), 2, 4, 8, or 16")
	}
}

func validateMaxFPS(fps int) error {
	if fps < 0 || fps > MaxAllowedFPS {
		return fmt.Errorf("max fps must be 0 (unlimited) or between 1 and %d", MaxAllowedFPS)
	}
	return nil
}

func upsertINIValues(src string, values map[string]map[string]string) string {
	sectionOrder := []string{"window", "render", "audio", "fog", "gameplay", "login", "chatshortcuts", "script", "ui", "uilayout"}
	seenSections := make(map[string]bool)
	written := make(map[string]map[string]bool)
	for section := range values {
		written[section] = make(map[string]bool)
	}
	src = strings.ReplaceAll(src, "\r\n", "\n")
	src = strings.ReplaceAll(src, "\r", "\n")
	lines := strings.Split(src, "\n")
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	out := make([]string, 0, len(lines)+12)
	currentSection := ""
	flushMissing := func(section string) {
		sectionValues, ok := values[section]
		if !ok {
			return
		}
		for _, key := range sortedINIKeys(sectionValues) {
			if written[section][normalizeKey(key)] {
				continue
			}
			out = append(out, fmt.Sprintf("%s = %s", key, sectionValues[key]))
			written[section][normalizeKey(key)] = true
		}
	}
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]") {
			flushMissing(currentSection)
			currentSection = normalizeKey(strings.TrimSpace(trimmed[1 : len(trimmed)-1]))
			seenSections[currentSection] = true
			out = append(out, line)
			continue
		}
		if sectionValues, ok := values[currentSection]; ok && trimmed != "" && !strings.HasPrefix(trimmed, "#") && !strings.HasPrefix(trimmed, ";") {
			if key, _, hasKey := strings.Cut(trimmed, "="); hasKey {
				normalizedKey := normalizeKey(key)
				replaced := false
				for saveKey, saveValue := range sectionValues {
					if normalizeKey(saveKey) != normalizedKey {
						continue
					}
					out = append(out, fmt.Sprintf("%s = %s", strings.TrimSpace(key), saveValue))
					written[currentSection][normalizedKey] = true
					replaced = true
					break
				}
				if replaced {
					continue
				}
			}
		}
		out = append(out, line)
	}
	flushMissing(currentSection)
	for _, section := range sectionOrder {
		if _, ok := values[section]; !ok || seenSections[section] {
			continue
		}
		if len(out) > 0 && strings.TrimSpace(out[len(out)-1]) != "" {
			out = append(out, "")
		}
		out = append(out, "["+iniSectionHeader(section)+"]")
		flushMissing(section)
	}
	return strings.Join(out, "\n") + "\n"
}

// iniSectionHeader is how a new section is written: the documented name for
// sections whose normalized key drops an underscore.
func iniSectionHeader(section string) string {
	if section == "uilayout" {
		return "ui_layout"
	}
	return section
}

func sortedINIKeys(values map[string]string) []string {
	preferred := []string{"fullscreen", "vsync", "max_fps", "fps", "bgm_volume", "sfx_volume", "no_shift", "no_ctrl", "keep_id", "saved_username", "1", "2", "3", "4", "5", "6", "7", "8", "9", "0"}
	keys := make([]string, 0, len(values))
	seen := make(map[string]bool, len(values))
	for _, key := range preferred {
		if _, ok := values[key]; ok {
			keys = append(keys, key)
			seen[key] = true
		}
	}
	for key := range values {
		if !seen[key] {
			keys = append(keys, key)
		}
	}
	return keys
}

func formatINIValueBool(value bool) string {
	return strconv.FormatBool(value)
}

func formatINIValueFloat(value float64) string {
	return strconv.FormatFloat(value, 'f', 2, 64)
}

func cleanINIValue(value string) string {
	value = strings.TrimSpace(value)
	if len(value) >= 2 {
		if (value[0] == '"' && value[len(value)-1] == '"') || (value[0] == '\'' && value[len(value)-1] == '\'') {
			return value[1 : len(value)-1]
		}
	}
	return value
}

func normalizeKey(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = strings.ReplaceAll(value, "-", "")
	value = strings.ReplaceAll(value, "_", "")
	return value
}

func setBool(raw string, dst *bool) error {
	value, err := strconv.ParseBool(raw)
	if err == nil {
		*dst = value
		return nil
	}
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "yes", "on", "enabled":
		*dst = true
		return nil
	case "no", "off", "disabled":
		*dst = false
		return nil
	default:
		return fmt.Errorf("invalid bool %q", raw)
	}
}

func setInt(raw string, dst *int) error {
	value, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil {
		return fmt.Errorf("invalid int %q", raw)
	}
	*dst = value
	return nil
}

func setFloat(raw string, dst *float64) error {
	value, err := strconv.ParseFloat(strings.TrimSpace(raw), 64)
	if err != nil {
		return fmt.Errorf("invalid float %q", raw)
	}
	*dst = value
	return nil
}

func resolveDataDir(value string) string {
	if value != "" {
		if abs, err := filepath.Abs(value); err == nil {
			return abs
		}
		return value
	}

	wd, err := os.Getwd()
	if err != nil {
		return "."
	}
	return wd
}
