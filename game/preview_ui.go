package game

import (
	"strings"

	"github.com/kivutar/goro/client"
	"github.com/kivutar/goro/glog"
	"github.com/kivutar/goro/network"
)

// SetPreviewWindows makes the offline map preview (GORO_PREVIEW_MAP) open
// sample windows on its first frame, so themes and ui-layout.ini edits can be
// checked without a server. Developer only; nothing is sent.
//
//	GORO_PREVIEW_WINDOWS=shop
func (m *WorldMode) SetPreviewWindows(names []string) {
	m.previewWindows = names
}

func (m *WorldMode) openPreviewWindows(ctx client.Context) {
	if len(m.previewWindows) == 0 {
		return
	}
	names := m.previewWindows
	m.previewWindows = nil
	for _, name := range names {
		switch strings.ToLower(strings.TrimSpace(name)) {
		case "shop":
			// theme:ok sample item IDs and prices for the preview shop
			m.ui.npc.Shop.OpenBuy([]network.ShopBuyItem{
				{ItemID: 501, Type: 0, Price: 50, DiscountPrice: 50},
				{ItemID: 502, Type: 0, Price: 200, DiscountPrice: 200},
				{ItemID: 503, Type: 0, Price: 550, DiscountPrice: 550},
				{ItemID: 1201, Type: 5, Price: 50, DiscountPrice: 50},
				{ItemID: 2301, Type: 4, Price: 10, DiscountPrice: 10},
			}, ctx)
		case "":
		default:
			glog.Warnf("preview window %q is unknown (known: shop)", name)
		}
	}
}
