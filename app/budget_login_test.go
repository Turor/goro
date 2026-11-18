//go:build budget

package app

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/kivutar/goro/config"
	"github.com/kivutar/goro/network"
	"github.com/kivutar/goro/render"
	"github.com/kivutar/goro/sim"
)

// TestHeadlessStartupToLoginBudget times New + first CA_LOGIN against an
// in-process fake server. Opt in with `go test -tags nofakecgo,budget`.
func TestHeadlessStartupToLoginBudget(t *testing.T) {
	t.Setenv("DISPLAY", "")
	t.Setenv("WAYLAND_DISPLAY", "")
	budget := sim.HeadlessLoginBudget
	if s := os.Getenv(sim.EnvHeadlessLoginBudget); s != "" {
		if d, err := time.ParseDuration(s + "s"); err == nil && d > 0 {
			budget = d
		}
	}

	listener, err := net.ListenTCP("tcp4", &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1)})
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	if err := listener.SetDeadline(time.Now().Add(budget + 2*time.Second)); err != nil {
		t.Fatal(err)
	}
	port := uint16(listener.Addr().(*net.TCPAddr).Port)
	root := t.TempDir()
	connections := fmt.Sprintf(`<clientinfo><connection><address>127.0.0.1</address><port>%d</port><version>55</version></connection></clientinfo>`, port)
	if err := os.WriteFile(filepath.Join(root, "clientinfo.xml"), []byte(connections), 0o600); err != nil {
		t.Fatal(err)
	}
	cfg := config.Config{
		Headless: true, DataDir: root,
		Audio:  config.AudioConfig{Disabled: true},
		Window: config.WindowConfig{Width: 800, Height: 600},
		Packet: config.PacketConfig{ClientDate: 20080910, ClientDateExplicit: true},
		Login:  config.LoginConfig{AutoLogin: true, Username: "tester", Password: "test-password", CharSlot: 0},
	}

	start := time.Now()
	g, err := New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() { done <- render.RunHeadless(ctx, g, cfg.Window) }()

	conn, err := listener.Accept()
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	want := network.BuildAccountLoginPacket(network.AccountLogin{Username: "tester", Password: "test-password", Version: 55})
	got := make([]byte, len(want))
	if err := conn.SetReadDeadline(time.Now().Add(budget)); err != nil {
		t.Fatal(err)
	}
	if _, err := io.ReadFull(conn, got); err != nil {
		t.Fatal(err)
	}
	elapsed := time.Since(start)
	if !bytes.Equal(got, want) {
		t.Fatalf("login packet = %x, want %x", got, want)
	}
	cancel()
	g.RequestQuit()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
	}
	if elapsed > budget {
		t.Fatalf("headless startup-to-login %s exceeds budget %s", elapsed, budget)
	}
	t.Logf("headless startup-to-login %s (budget %s)", elapsed, budget)
}
