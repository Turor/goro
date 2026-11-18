package rotheme

import "github.com/gogpu/ui/widget"

var (
	ContextMenuItemHeight       = 18
	contextMenuItemTextPaddingX = float32(4)
	// Registered in theme_tokens.go as [theme] context_menu_hover.
	contextMenuItemHover = widget.RGBA8(115, 156, 239, 255)
)

func init() {
	s := Section("context_menu", "shared right-click menu chrome")
	s.Int("item_height", &ContextMenuItemHeight, 12, 48, "context menu row height")
	s.Float("item_padding_x", &contextMenuItemTextPaddingX, 0, 24, "context menu item text inset")
}
