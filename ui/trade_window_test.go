package ui

import (
	"testing"

	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

func TestTradePanelsSplitAvailableWidth(t *testing.T) {
	w := &TradeWindow{partnerName: "Zambla"}
	panels := w.tradePanels()
	availableWidth := float32(tradeWindowW - 2*tradePanelPad)
	panels.Layout(widget.NewContext(), geometry.Tight(geometry.Sz(availableWidth, float32(tradePanelH+tradePanelExtra))))

	children := panels.Children()
	if len(children) != 2 {
		t.Fatalf("trade panes = %d, want 2", len(children))
	}
	wantWidth := (availableWidth - float32(tradePanelGap)) / 2
	left := children[0].(interface{ Bounds() geometry.Rect }).Bounds()
	right := children[1].(interface{ Bounds() geometry.Rect }).Bounds()
	if left.Width() != wantWidth || right.Width() != wantWidth {
		t.Fatalf("trade pane widths = %.1f and %.1f, want %.1f each", left.Width(), right.Width(), wantWidth)
	}
	if right.Min.X != left.Max.X+float32(tradePanelGap) || right.Max.X != availableWidth {
		t.Fatalf("trade pane bounds leave unused space: left=%v right=%v available=%.1f", left, right, availableWidth)
	}
	if float32(tradePanelW) != wantWidth {
		t.Fatalf("trade drop target width = %d, want %.1f", tradePanelW, wantWidth)
	}
}

func TestTradeItemRowUsesFullPaneWidth(t *testing.T) {
	row := tradeItemRowWidget(tradeWindowItem{name: "Yggdrasil Leaf"})
	row.Layout(widget.NewContext(), geometry.Tight(geometry.Sz(float32(tradePanelW), float32(tradeRowH))))

	children := row.Children()
	if len(children) != 3 {
		t.Fatalf("trade row children = %d, want icon, name and quantity", len(children))
	}
	quantity := children[2].(interface{ Bounds() geometry.Rect }).Bounds()
	if quantity.Max.X != float32(tradePanelW) {
		t.Fatalf("trade quantity right edge = %.1f, want pane edge %d", quantity.Max.X, tradePanelW)
	}
}
