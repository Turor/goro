package bot

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/kivutar/goro/client"
)

func TestScriptQuitCallsRequestQuit(t *testing.T) {
	path := filepath.Join(t.TempDir(), "quit.lua")
	if err := os.WriteFile(path, []byte(`
function tick()
	assert(goro.quit())
end
`), 0o600); err != nil {
		t.Fatal(err)
	}
	called := false
	ctx := client.Context{
		RequestQuit: func() { called = true },
	}
	script, err := New(ctx, path, Bindings{})
	if err != nil {
		t.Fatal(err)
	}
	defer script.Close()
	if err := script.Tick(); err != nil {
		t.Fatal(err)
	}
	if !called {
		t.Fatal("goro.quit did not call RequestQuit")
	}
}

func TestScriptQuitWithoutHandler(t *testing.T) {
	path := filepath.Join(t.TempDir(), "quit.lua")
	if err := os.WriteFile(path, []byte(`
function tick()
	assert(not goro.quit())
end
`), 0o600); err != nil {
		t.Fatal(err)
	}
	script, err := New(client.Context{}, path, Bindings{})
	if err != nil {
		t.Fatal(err)
	}
	defer script.Close()
	if err := script.Tick(); err != nil {
		t.Fatal(err)
	}
}
