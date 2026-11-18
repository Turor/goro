package rotheme

import (
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

type TableCell struct {
	Text  string
	Width float32
	Align widget.TextAlign
	Head  bool
}

type TableRow []TableCell

type TableOption func(*tableWidget)

type tableWidget struct {
	widget.WidgetBase
	rows         []TableRow
	rowHeight    float32
	rowHeightSet bool
	headBG       widget.Color
	cellBG       widget.Color
	colorsSet    bool
}

func Table(rows []TableRow, opts ...TableOption) widget.Widget {
	w := &tableWidget{rows: rows}
	for _, opt := range opts {
		opt(w)
	}
	w.SetVisible(true)
	w.SetEnabled(false)
	return w
}

func TableRowHeightOpt(height float32) TableOption {
	return func(w *tableWidget) {
		w.rowHeight = height
		w.rowHeightSet = true
	}
}

func TableColors(headBG, cellBG widget.Color) TableOption {
	return func(w *tableWidget) {
		w.headBG = headBG
		w.cellBG = cellBG
		w.colorsSet = true
	}
}

func (w *tableWidget) metrics() (rowHeight, gap, pad float32) {
	rowHeight = TableRowHeight
	if w.rowHeightSet {
		rowHeight = w.rowHeight
	}
	return rowHeight, TableGap, TableCellPadX
}

func (w *tableWidget) colors() (head, cell widget.Color) {
	if w.colorsSet {
		return w.headBG, w.cellBG
	}
	return Default.Colors.ButtonHover, Default.Colors.WindowFooter
}

func (w *tableWidget) Layout(_ widget.Context, constraints geometry.Constraints) geometry.Size {
	size := constraints.Constrain(geometry.Sz(w.width(), w.height()))
	w.SetBounds(geometry.FromPointSize(w.Position(), size))
	return size
}

func (w *tableWidget) Draw(_ widget.Context, canvas widget.Canvas) {
	bounds := w.Bounds()
	rowHeight, gap, pad := w.metrics()
	headBG, cellBG := w.colors()
	for rowIndex, row := range w.rows {
		x := bounds.Min.X
		y := bounds.Min.Y + float32(rowIndex)*(rowHeight+gap)
		for _, cell := range row {
			bg := cellBG
			if cell.Head {
				bg = headBG
			}
			rect := geometry.NewRect(x, y, cell.Width, rowHeight)
			canvas.DrawRect(rect, bg)
			textRect := geometry.NewRect(rect.Min.X+pad, rect.Min.Y, rect.Width()-pad*2, rect.Height())
			if cell.Head {
				DrawLabel(canvas, cell.Text, textRect, cell.Align)
			} else {
				DrawText(canvas, cell.Text, textRect, Default.Typography.TextSize, Default.Colors.Text, false, cell.Align)
			}
			x += cell.Width + gap
		}
	}
}

func (w *tableWidget) Event(widget.Context, event.Event) bool {
	return false
}

func (w *tableWidget) Children() []widget.Widget {
	return nil
}

func (w *tableWidget) width() float32 {
	_, gap, _ := w.metrics()
	width := float32(0)
	for _, row := range w.rows {
		rowWidth := float32(0)
		for i, cell := range row {
			rowWidth += cell.Width
			if i > 0 {
				rowWidth += gap
			}
		}
		if rowWidth > width {
			width = rowWidth
		}
	}
	return width
}

func (w *tableWidget) height() float32 {
	if len(w.rows) == 0 {
		return 0
	}
	rowHeight, gap, _ := w.metrics()
	return float32(len(w.rows))*rowHeight + float32(len(w.rows)-1)*gap
}
