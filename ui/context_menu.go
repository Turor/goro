package ui

import (
	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/widget"
	"github.com/kivutar/goro/ui/rotheme"
)

func contextMenuRowH() int { return rotheme.ContextMenuItemHeight }

func contextMenu(width int, rows ...widget.Widget) widget.Widget {
	return Win(
		TitleBar(false),
		Radius(float32(contextMenuRadius)),
		Size(float32(width), float32(contextMenuHeight(len(rows)))),
		Content(
			primitives.Box(rows...).
				CrossAlign(primitives.CrossAxisStretch).
				Padding(float32(contextMenuPadding)),
		),
	)
}

func contextMenuItem(label string, onClick func()) widget.Widget {
	return rotheme.ContextMenuItem(label, onClick)
}

func contextMenuHeight(rows int) int {
	return contextMenuRowH()*rows + contextMenuPadding*2
}

// claimLayoutReload registers fn and drops a struct-copied window that still
// points its rebuild at this overlay.
func (w *Window) claimLayoutReload(fn func()) {
	if w.published != nil {
		for other := range layoutRebuildWindows {
			if other != w && other.published == w.published {
				other.layoutRebuild = nil
				delete(layoutRebuildWindows, other)
			}
		}
	}
	w.OnLayoutReload(fn)
}
