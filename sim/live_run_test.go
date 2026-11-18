//go:build live

package sim

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

// TestLiveLoginScenario runs scenarios/login.lua against TEST. Opt in with
// GORO_LIVE_TEST=1 and GORO_LIVE_ACCOUNT_FILE. Never part of the default gate.
func TestLiveLoginScenario(t *testing.T) {
	if err := CheckLiveEnv(); err != nil {
		t.Skip(err.Error())
	}
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller")
	}
	root := filepath.Join(filepath.Dir(file), "..")
	runner := filepath.Join(root, "scripts", "sim", "live-runner.sh")
	cmd := exec.Command(runner, filepath.Join(root, "scenarios", "login.lua"))
	cmd.Dir = root
	cmd.Env = os.Environ()
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("live runner: %v", err)
	}
}
