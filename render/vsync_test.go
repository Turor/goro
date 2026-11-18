package render

import (
	"testing"

	"github.com/kivutar/goro/config"
)

func TestLowLatencyVSyncNeedsCap(t *testing.T) {
	for _, c := range []struct {
		vsync  bool
		maxFPS int
		want   bool
	}{{true, 144, true}, {true, 0, false}, {false, 144, false}} {
		if got := lowLatencyVSync(config.RenderConfig{VSync: c.vsync, MaxFPS: c.maxFPS}); got != c.want {
			t.Errorf("lowLatencyVSync(vsync=%v, max_fps=%d) = %v, want %v", c.vsync, c.maxFPS, got, c.want)
		}
	}
}
