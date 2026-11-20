package ui

import (
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/widget"

	"github.com/kivutar/goro/ui/rotheme"
)

// scaleToken is the design size of one widget, captured before a resize
// multiplies it. Applying the scale from this token, not from the live
// style, keeps a second layout from compounding.
type scaleToken struct {
	width    float32
	height   float32
	minW     float32
	minH     float32
	gap      float32
	padding  geometry.Insets
	fontSize float32
	hasFont  bool
}

func (f *windowFrame) applyContentScale() {
	userW := f.Style().ExplicitWidth
	userH := f.Style().ExplicitHeight
	if f.designW <= 0 || f.designH <= 0 {
		f.designW, f.designH = userW, userH
	}
	scale := float32(1)
	if f.designW > 0 && f.designH > 0 && userW > 0 && userH > 0 {
		scale = userW / f.designW
		if heightScale := userH / f.designH; heightScale < scale {
			scale = heightScale
		}
	}
	if scale <= 0 {
		scale = 1
	}
	f.contentScale = scale
	if f.scaleTokens == nil {
		f.scaleTokens = make(map[widget.Widget]*scaleToken)
	}
	for _, child := range f.Children() {
		f.scaleWidgetTree(child)
	}
}

func (f *windowFrame) scaleWidgetTree(current widget.Widget) {
	if current == nil {
		return
	}
	switch node := current.(type) {
	case *primitives.BoxWidget:
		f.scaleBox(node)
	case *primitives.TextWidget:
		f.scaleText(node)
	case *windowTitleText:
		if node.TextWidget != nil {
			f.scaleText(node.TextWidget)
		}
	}
	for _, child := range current.Children() {
		f.scaleWidgetTree(child)
	}
}

func (f *windowFrame) scaleBox(box *primitives.BoxWidget) {
	token := f.boxToken(box)
	style := box.Style()
	if token.width > 0 {
		box.Width(token.width * f.contentScale)
	}
	if token.height > 0 {
		box.Height(token.height * f.contentScale)
	}
	if token.minW > 0 {
		box.MinWidthValue(token.minW * f.contentScale)
	} else if style.MinWidth > 0 && token.width == 0 {
		box.MinWidthValue(0)
	}
	if token.minH > 0 {
		box.MinHeightValue(token.minH * f.contentScale)
	}
	box.Gap(token.gap * f.contentScale)
	pad := token.padding
	box.PaddingLeft(pad.Left * f.contentScale)
	box.PaddingRight(pad.Right * f.contentScale)
	box.PaddingTop(pad.Top * f.contentScale)
	box.PaddingBottom(pad.Bottom * f.contentScale)
}

func (f *windowFrame) boxToken(box *primitives.BoxWidget) *scaleToken {
	if token := f.scaleTokens[box]; token != nil {
		return token
	}
	style := box.Style()
	token := &scaleToken{
		width:   style.ExplicitWidth,
		height:  style.ExplicitHeight,
		minW:    style.MinWidth,
		minH:    style.MinHeight,
		gap:     style.Gap,
		padding: style.Padding,
	}
	f.scaleTokens[box] = token
	return token
}

func (f *windowFrame) scaleText(text *primitives.TextWidget) {
	token := f.scaleTokens[text]
	if token == nil {
		size := text.Style().FontSize
		if size <= 0 {
			size = rotheme.Default.Typography.TextSize
		}
		token = &scaleToken{fontSize: size, hasFont: true}
		f.scaleTokens[text] = token
	}
	if token.hasFont && token.fontSize > 0 {
		text.FontSize(token.fontSize * f.contentScale)
	}
}

func (f *windowFrame) withContentScale(fn func()) {
	scale := f.contentScale
	if scale <= 0 {
		scale = 1
	}
	previous := rotheme.ContentScale()
	rotheme.SetContentScale(scale)
	defer rotheme.SetContentScale(previous)
	fn()
}

func withContentScale(scale float32, fn func()) {
	if scale <= 0 {
		scale = 1
	}
	previous := rotheme.ContentScale()
	rotheme.SetContentScale(scale)
	defer rotheme.SetContentScale(previous)
	fn()
}
