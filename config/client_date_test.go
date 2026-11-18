package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultPacketClientDateIs20250716(t *testing.T) {
	isolateUserConfig(t)
	cfg, err := LoadConfig(nil)
	if err != nil {
		t.Fatal(err)
	}
	if DefaultPacketClientDate != 20250716 {
		t.Fatalf("DefaultPacketClientDate = %d, want 20250716", DefaultPacketClientDate)
	}
	if cfg.Packet.ClientDate != 20250716 {
		t.Fatalf("default client date = %d, want 20250716", cfg.Packet.ClientDate)
	}
	if cfg.Packet.ClientDateExplicit {
		t.Fatal("default client date must not be marked explicit")
	}
}

func TestConfigFileSelectsPacketClientDate(t *testing.T) {
	isolateUserConfig(t)
	dir := t.TempDir()
	path := filepath.Join(dir, "goro.ini")
	if err := os.WriteFile(path, []byte("[packet]\nclient_date = 20080910\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	cfg, err := LoadConfig([]string{"--config", path})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Packet.ClientDate != 20080910 {
		t.Fatalf("config client date = %d, want 20080910", cfg.Packet.ClientDate)
	}
	if !cfg.Packet.ClientDateExplicit {
		t.Fatal("config client date should be explicit")
	}
}

func TestCLISelectsPacketClientDate(t *testing.T) {
	isolateUserConfig(t)
	cfg, err := LoadConfig([]string{"--packet-client-date", "20080910"})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Packet.ClientDate != 20080910 {
		t.Fatalf("cli client date = %d, want 20080910", cfg.Packet.ClientDate)
	}
	if !cfg.Packet.ClientDateExplicit {
		t.Fatal("cli client date should be explicit")
	}
}

func TestCLIOverridesConfigPacketClientDate(t *testing.T) {
	isolateUserConfig(t)
	dir := t.TempDir()
	path := filepath.Join(dir, "goro.ini")
	if err := os.WriteFile(path, []byte("[packet]\nclient_date = 20080910\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	cfg, err := LoadConfig([]string{"--config", path, "--packet-client-date", "20211103"})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Packet.ClientDate != 20211103 || !cfg.Packet.ClientDateExplicit {
		t.Fatalf("cli override = %d explicit=%v", cfg.Packet.ClientDate, cfg.Packet.ClientDateExplicit)
	}
}

func TestResolvePacketClientDate(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name      string
		cfg       Config
		packetVer int
		want      int
	}{
		{"default with no clientinfo", Config{Packet: PacketConfig{ClientDate: DefaultPacketClientDate}}, 0, 20250716},
		{"clientinfo preferred when unset", Config{Packet: PacketConfig{ClientDate: DefaultPacketClientDate}}, 20250716, 20250716},
		{"clientinfo can select 20080910 when unset", Config{Packet: PacketConfig{ClientDate: DefaultPacketClientDate}}, 20080910, 20080910},
		{"explicit 20080910 beats clientinfo", Config{Packet: PacketConfig{ClientDate: 20080910, ClientDateExplicit: true}}, 20250716, 20080910},
		{"explicit 20250716 beats other clientinfo", Config{Packet: PacketConfig{ClientDate: 20250716, ClientDateExplicit: true}}, 20080910, 20250716},
		{"zero cfg falls back to default", Config{}, 0, 20250716},
		{"zero cfg uses clientinfo", Config{}, 20190530, 20190530},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.cfg.ResolvePacketClientDate(tc.packetVer)
			if got != tc.want {
				t.Fatalf("got %d, want %d", got, tc.want)
			}
		})
	}
}
