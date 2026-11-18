package config

import "testing"

func TestCaptureCLIAndEnv(t *testing.T) {
	cfg := defaultConfig()
	if err := applyCLI(&cfg, []string{"--capture", "sim/testdata/out.json"}); err != nil {
		t.Fatal(err)
	}
	if cfg.Network.Capture != "sim/testdata/out.json" {
		t.Fatalf("capture = %q", cfg.Network.Capture)
	}

	cfg = defaultConfig()
	t.Setenv("GORO_CAPTURE", "from-env.json")
	if err := applyCLI(&cfg, nil); err != nil {
		t.Fatal(err)
	}
	if cfg.Network.Capture != "from-env.json" {
		t.Fatalf("env capture = %q", cfg.Network.Capture)
	}

	cfg = defaultConfig()
	t.Setenv("GORO_CAPTURE", "from-env.json")
	if err := applyCLI(&cfg, []string{"--capture", "from-flag.json"}); err != nil {
		t.Fatal(err)
	}
	if cfg.Network.Capture != "from-flag.json" {
		t.Fatalf("flag should win: %q", cfg.Network.Capture)
	}
}
