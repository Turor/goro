package sim

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestLiveRunnerSkipsWithoutOptIn(t *testing.T) {
	t.Setenv(EnvLiveTest, "")
	t.Setenv(EnvLiveAccountFile, "")
	if err := CheckLiveEnv(); !errors.Is(err, ErrLiveDisabled) {
		t.Fatalf("CheckLiveEnv = %v, want ErrLiveDisabled", err)
	}
}

func TestLiveRunnerSkipsWithoutAccountFile(t *testing.T) {
	t.Setenv(EnvLiveTest, "1")
	t.Setenv(EnvLiveAccountFile, "")
	if err := CheckLiveEnv(); !errors.Is(err, ErrLiveNoAccount) {
		t.Fatalf("CheckLiveEnv = %v, want ErrLiveNoAccount", err)
	}
}

func TestLiveRunnerRejectsNonKgapUsername(t *testing.T) {
	for _, name := range []string{"not-a-kgap", "korangar_m1_1108", "kgap", "kgap01x"} {
		path := filepath.Join(t.TempDir(), "acct.json")
		body := []byte(`{"username":"` + name + `","password":"dummy"}`)
		if err := os.WriteFile(path, body, 0o600); err != nil {
			t.Fatal(err)
		}
		t.Setenv(EnvLiveTest, "1")
		t.Setenv(EnvLiveAccountFile, path)
		if err := CheckLiveEnv(); !errors.Is(err, ErrLiveNotKgap) {
			t.Fatalf("username %q: CheckLiveEnv = %v, want ErrLiveNotKgap", name, err)
		}
	}
}

func TestLiveRunnerAcceptsKgapUsername(t *testing.T) {
	path := filepath.Join(t.TempDir(), "acct.json")
	if err := os.WriteFile(path, []byte(`{"username":"kgap01","password":"dummy"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv(EnvLiveTest, "1")
	t.Setenv(EnvLiveAccountFile, path)
	if err := CheckLiveEnv(); err != nil {
		t.Fatal(err)
	}
	if !IsKgapUsername("kgap01") || IsKgapUsername("player") {
		t.Fatal("IsKgapUsername boundaries")
	}
}
