//go:build slow

package sim

import (
	"os"
	"testing"
)

// TestSlowRequiresDataDir keeps real GRF/data work behind the slow tag.
// Default `go test` never compiles this file.
func TestSlowRequiresDataDir(t *testing.T) {
	if os.Getenv("GORO_DATA_DIR") == "" {
		t.Skip("slow tests need GORO_DATA_DIR pointing at client data")
	}
}
