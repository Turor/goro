package ui

import (
	"github.com/gogpu/ui/widget"
	"github.com/kivutar/goro/ui/rotheme"
)

var (
	footerStripeH     float32 = 2
	footerStripeColor         = widget.RGBA(1, 1, 1, 0.5)
)

func init() {
	s := rotheme.Section("footer", "stripes drawn across a window footer")
	s.Float("stripe_height", &footerStripeH, 1, 16, "footer stripe thickness")
	s.Color("stripe_color", &footerStripeColor, "footer stripe color")
}
