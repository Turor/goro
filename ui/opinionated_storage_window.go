package ui

// presentClassicStorage opens the separate storage window. The opinionated
// bag already shows that list, so selecting its Storage or Guild tab must not
// pop the old window. The server storage stays open so a drop can still store.
func PresentClassicStorage(storage *StorageWindow, bag *InventoryBagWindow, ctx Context) {
	if storage == nil {
		return
	}
	if bag != nil && bag.HidesClassicStorage() {
		storage.SetOpen(false)
		return
	}
	storage.OpenWindow(ctx)
}

func (w *InventoryBagWindow) HidesClassicStorage() bool {
	return w != nil && w.opinionated && w.storageSource()
}
