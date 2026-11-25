package game

import (
	"github.com/kivutar/goro/client"
	gameui "github.com/kivutar/goro/ui"
)

func (m *WorldMode) presentClassicStorage(ctx client.Context) {
	if m == nil {
		return
	}
	gameui.PresentClassicStorage(&m.ui.storageWindow, &m.ui.inventoryBag, ctx)
}
