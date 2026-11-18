//go:build unix

package rotheme

import (
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"
)

// A FIFO at the layout path must be refused, not opened (which would block).
func TestLoadLayoutFileRefusesFIFO(t *testing.T) {
	path := filepath.Join(t.TempDir(), "ui-layout.ini")
	if err := syscall.Mkfifo(path, 0o644); err != nil {
		t.Skip("mkfifo:", err)
	}
	done := make(chan error, 1)
	go func() {
		_, _, err := LoadLayoutFile(path)
		done <- err
	}()
	select {
	case err := <-done:
		if err == nil || !strings.Contains(err.Error(), "not a regular file") {
			t.Fatalf("err = %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("LoadLayoutFile blocked on a FIFO")
	}
}
