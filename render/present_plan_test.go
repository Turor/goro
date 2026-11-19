package render

import (
	"testing"

	"github.com/kivutar/goro/config"
)

func TestPresentRequestMailboxBeforeFifo(t *testing.T) {
	cfg := config.RenderConfig{VSync: true, MaxFPS: 144}
	req := BuildPresentRequest(cfg)
	if !req.VSync || !req.LowLatency {
		t.Fatalf("present %+v", req)
	}
	if req.MaxFPS != cfg.MaxFPS || req.Limiter.MaxFPS != cfg.MaxFPS {
		t.Fatalf("limiter %d max %d want %d", req.Limiter.MaxFPS, req.MaxFPS, cfg.MaxFPS)
	}
	mail, fifo := indexOfMode(req.ModeOrder, PresentModeMailbox), indexOfMode(req.ModeOrder, PresentModeFifo)
	if mail < 0 || fifo < 0 || mail >= fifo {
		t.Fatalf("mode order %v", req.ModeOrder)
	}
	if lowLatencyVSync(cfg) != req.LowLatency {
		t.Fatal("lowLatencyVSync diverges from BuildPresentRequest")
	}

	uncapped := BuildPresentRequest(config.RenderConfig{VSync: true, MaxFPS: 0})
	if uncapped.LowLatency || uncapped.Limiter.MaxFPS != 0 {
		t.Fatalf("uncapped %+v", uncapped)
	}
	if len(uncapped.ModeOrder) == 0 || uncapped.ModeOrder[0] != PresentModeFifoRelaxed {
		t.Fatalf("uncapped order %v", uncapped.ModeOrder)
	}
	if indexOfMode(uncapped.ModeOrder, PresentModeMailbox) >= 0 {
		t.Fatalf("uncapped vsync must not request mailbox: %v", uncapped.ModeOrder)
	}
}

func indexOfMode(order []PresentMode, mode PresentMode) int {
	for i, m := range order {
		if m == mode {
			return i
		}
	}
	return -1
}
