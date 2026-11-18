package ui

import "github.com/kivutar/goro/ui/rotheme"

// Item icon size shared by inventory, shops, carts and trades.
var (
	inventoryIconSize = 24
)

func init() {
	s := rotheme.Section("inventory", "inventory bag")
	s.Int("icon_size", &inventoryIconSize, 8, 96, "item icon size")
}
