package render

import "github.com/kivutar/goro/config"

// PresentMode is a swapchain present mode, in the order gogpu should try.
type PresentMode string

const (
	PresentModeImmediate   PresentMode = "immediate"
	PresentModeMailbox     PresentMode = "mailbox"
	PresentModeFifoRelaxed PresentMode = "fifo-relaxed"
	PresentModeFifo        PresentMode = "fifo"
)

// FrameLimiter is the frame-loop cap. MaxFPS is the same value passed to
// gogpu Config.WithMaxFPS. gogpu sleeps; this client does not sleep again.
type FrameLimiter struct {
	MaxFPS int
}

// PresentRequest is the present choice Run and the tests share.
type PresentRequest struct {
	VSync      bool
	LowLatency bool
	MaxFPS     int
	ModeOrder  []PresentMode
	Limiter    FrameLimiter
}

// BuildPresentRequest selects mailbox ahead of FIFO when vsync is on and a
// frame cap is set, and applies that same cap to the limiter.
func BuildPresentRequest(cfg config.RenderConfig) PresentRequest {
	lowLatency := cfg.VSync && cfg.MaxFPS > 0
	var order []PresentMode
	switch {
	case cfg.VSync && lowLatency:
		order = []PresentMode{PresentModeMailbox, PresentModeFifoRelaxed, PresentModeFifo}
	case cfg.VSync:
		order = []PresentMode{PresentModeFifoRelaxed, PresentModeFifo}
	default:
		order = []PresentMode{PresentModeImmediate, PresentModeMailbox, PresentModeFifo}
	}
	return PresentRequest{
		VSync:      cfg.VSync,
		LowLatency: lowLatency,
		MaxFPS:     cfg.MaxFPS,
		ModeOrder:  order,
		Limiter:    FrameLimiter{MaxFPS: cfg.MaxFPS},
	}
}
