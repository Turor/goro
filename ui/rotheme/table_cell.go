package rotheme

import (
	"image"

	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

type TableViewSimpleCell struct {
	Text               string
	Icon               image.Image
	Align              widget.TextAlign
	Color              widget.Color
	Hidden             bool
	IconButton         IconButtonKind
	HasIconButton      bool
	IconButtonDisabled bool
}

func (c TableViewSimpleCell) simpleCell() tableViewSimpleCell {
	color := c.Color
	if color.IsTransparent() {
		color = Default.Colors.Text
	}
	return tableViewSimpleCell{
		text:               c.Text,
		icon:               c.Icon,
		align:              c.Align,
		color:              color,
		visible:            !c.Hidden,
		iconButton:         c.IconButton,
		hasIconButton:      c.HasIconButton,
		iconButtonDisabled: c.IconButtonDisabled,
	}
}

func TableViewIconButtonCell(kind IconButtonKind, disabled bool) TableViewSimpleCell {
	return TableViewSimpleCell{
		IconButton:         kind,
		HasIconButton:      true,
		IconButtonDisabled: disabled,
	}
}

type tableViewSimpleCell struct {
	text               string
	icon               image.Image
	align              widget.TextAlign
	color              widget.Color
	visible            bool
	iconButton         IconButtonKind
	hasIconButton      bool
	iconButtonDisabled bool
}

func (c tableViewSimpleCell) draw(canvas widget.Canvas, bounds geometry.Rect) {
	if !c.visible {
		return
	}
	if c.hasIconButton {
		size := IconButtonSize
		if size > bounds.Width() {
			size = bounds.Width()
		}
		if size > bounds.Height() {
			size = bounds.Height()
		}
		buttonBounds := geometry.NewRect(
			bounds.Min.X+(bounds.Width()-size)/2,
			bounds.Min.Y+(bounds.Height()-size)/2,
			size,
			size,
		)
		DrawIconButton(canvas, buttonBounds, c.iconButton, false, c.iconButtonDisabled)
		return
	}
	textBounds := bounds
	if c.icon != nil {
		iconSize := bounds.Height() - tableIconInset*2
		if iconSize < 1 {
			iconSize = bounds.Height()
		}
		iconY := bounds.Min.Y + (bounds.Height()-iconSize)/2
		DrawFittedImage(canvas, c.icon, geometry.NewRect(bounds.Min.X+tableIconInset, iconY, iconSize, iconSize))
		textBounds.Min.X += iconSize + tableIconTextGap
	}
	textBounds.Min.X += TableCellPadX
	textBounds.Max.X -= TableCellPadX
	if textBounds.Width() < 0 {
		textBounds.Max.X = textBounds.Min.X
	}
	textHeight := Default.Typography.TextSize * tableCellLineScale
	if textHeight > textBounds.Height() {
		textHeight = textBounds.Height()
	}
	textY := textBounds.Min.Y + (textBounds.Height()-textHeight)/2
	DrawText(canvas, c.text, geometry.NewRect(textBounds.Min.X, textY, textBounds.Width(), textHeight), Default.Typography.TextSize, c.color, false, c.align)
}
