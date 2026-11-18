package rotheme

import (
	"strings"

	"github.com/gogpu/ui/core/textfield"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

func TextField(value string, inputType textfield.InputType, onChange func(string), onSubmit func(string), extra ...textfield.Option) *textfield.Widget {
	opts := []textfield.Option{
		textfield.InitialValue(value),
		textfield.InputTypeOpt(inputType),
		textfield.OnChange(onChange),
		textfield.OnSubmit(onSubmit),
		textfield.PainterOpt(TextFieldPainter{}),
	}
	opts = append(opts, extra...)
	return textfield.New(opts...)
}

type TextFieldPainter struct{}

func (TextFieldPainter) PaintTextField(canvas widget.Canvas, state *textfield.PaintState) {
	border := Default.Colors.FooterLine
	if state.Focused {
		border = Default.Colors.InputFocus
	}
	canvas.DrawRect(state.Bounds, Default.Colors.PanelBody)
	drawTextFieldInsetShadow(canvas, state.Bounds)
	canvas.StrokeRect(state.Bounds, border, 1)

	content := geometry.NewRect(state.Bounds.Min.X+textFieldPadX, state.Bounds.Min.Y+textFieldPadY, state.Bounds.Width()-textFieldPadX*2, state.Bounds.Height()-textFieldPadY*2)
	text := state.Text
	if state.InputType == textfield.TypePassword {
		text = strings.Repeat("*", len([]rune(text)))
	}
	textColor := Default.Colors.Text
	if text == "" && !state.Focused {
		text = state.Placeholder
		textColor = Default.Colors.MutedText
	}
	DrawText(canvas, text, content, Default.Typography.TextSize, textColor, false, widget.TextAlignLeft)
	if state.Focused && state.SelectStart == state.SelectEnd {
		runes := []rune(text)
		cursor := state.CursorPos
		if cursor > len(runes) {
			cursor = len(runes)
		}
		x := min(content.Max.X, content.Min.X+MeasureText(canvas, string(runes[:cursor]), Default.Typography.TextSize, false))
		canvas.DrawLine(geometry.Pt(x, state.Bounds.Min.Y+textFieldPadY), geometry.Pt(x, state.Bounds.Max.Y-textFieldPadY), Default.Colors.Text, 1)
	}
}

func drawTextFieldInsetShadow(canvas widget.Canvas, bounds geometry.Rect) {
	rows := min(textFieldShadowRows, int(bounds.Height())-textFieldShadowInset)
	for row := 0; row < rows; row++ {
		alpha := uint8(textFieldShadowAlpha - row*textFieldShadowAlphaStep)
		canvas.DrawRect(
			geometry.NewRect(bounds.Min.X, bounds.Min.Y+float32(row), bounds.Width(), 1),
			textFieldShadow.WithAlpha(float32(alpha)/255),
		)
	}
}
