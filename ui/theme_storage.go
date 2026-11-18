package ui

import "github.com/kivutar/goro/ui/rotheme"

// [storage] defaults: the kafra storage window. Width and height follow the table.
var (
	storageTabW       = 32
	storageTabH       = 40
	storageTableViewW = 312
	storageRowH       = 32
	storageRows       = 9
	storageItemMinW   = float32(120)
	storageAmountW    = float32(76)
	storageScreenTop  = 118

	storageTabRailW     int
	storageWindowWidth  int
	storageWindowTitleH int
	storageTableViewH   int
	storageWindowHeight int
)

func init() {
	s := rotheme.Section("storage", "kafra storage window")
	s.Int("tab_width", &storageTabW, 16, 96, "category tab width")
	s.Int("tab_height", &storageTabH, 16, 120, "category tab height")
	s.Int("table_width", &storageTableViewW, 160, 900, "item table width")
	s.Int("row_height", &storageRowH, 12, 96, "item row height")
	s.Int("visible_rows", &storageRows, 1, 30, "visible item rows")
	s.Float("item_min_width", &storageItemMinW, 40, 480, "item column minimum width")
	s.Float("amount_width", &storageAmountW, 24, 240, "quantity column width")
	s.Int("screen_top", &storageScreenTop, 0, 400, "distance from the screen top when opened")
	rotheme.OnLayoutApplied(recomputeStorageMetrics)
	recomputeStorageMetrics()
}

func recomputeStorageMetrics() {
	storageTabRailW = storageTabW + storageTabOver*2
	storageWindowWidth = storageTabRailW + verticalTabDividerW + storageTableViewW
	storageWindowTitleH = ROWindowTitleHeight
	storageTableViewH = storageRows * storageRowH
	storageWindowHeight = storageWindowTitleH + storageTableViewH + ROWindowFooterHeight
}
