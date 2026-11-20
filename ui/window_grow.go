package ui

import (
	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/widget"
)

func growWindowPanes(root widget.Widget, availW, availH float32) {
	growWindowHeight(root, availH)
	growWindowWidth(root, availW)
}

func growWindowHeight(root widget.Widget, availH float32) {
	if availH <= 0 {
		return
	}
	boxes := windowBoxes(root)
	maxH := float32(0)
	for _, box := range boxes {
		if h := box.Style().ExplicitHeight; h > maxH {
			maxH = h
		}
	}
	if maxH < float32(windowMinHeight) {
		return
	}
	for _, box := range boxes {
		style := box.Style()
		if style.ExplicitHeight != maxH && style.MinHeight != maxH {
			continue
		}
		token := style.MinHeight
		if token <= 0 || token > style.ExplicitHeight {
			token = style.ExplicitHeight
			box.MinHeightValue(token)
		}
		want := token
		if availH > token {
			want = availH
		}
		if style.ExplicitHeight != want {
			box.Height(want)
		}
	}
}

func growWindowWidth(root widget.Widget, availW float32) {
	box, ok := root.(*primitives.BoxWidget)
	if !ok || availW <= 0 {
		return
	}
	var tokenSum float32
	var count int
	var rightmost *primitives.BoxWidget
	for _, child := range box.Children() {
		childBox, ok := child.(*primitives.BoxWidget)
		if !ok {
			continue
		}
		width := childBox.Style().ExplicitWidth
		if width <= 0 {
			continue
		}
		token := childBox.Style().MinWidth
		if token <= 0 || token > width {
			token = width
		}
		tokenSum += token
		count++
		rightmost = childBox
	}
	if rightmost == nil || count == 0 {
		return
	}
	token := rightmost.Style().MinWidth
	current := rightmost.Style().ExplicitWidth
	if token <= 0 || token > current {
		token = current
		rightmost.MinWidthValue(token)
	}
	extra := availW - tokenSum - box.Style().Gap*float32(count-1)
	want := token
	if extra > 0 {
		want = token + extra
	}
	if current != want {
		rightmost.Width(want)
	}
}

func windowBoxes(root widget.Widget) []*primitives.BoxWidget {
	var boxes []*primitives.BoxWidget
	var walk func(widget.Widget)
	walk = func(current widget.Widget) {
		if current == nil {
			return
		}
		if box, ok := current.(*primitives.BoxWidget); ok {
			boxes = append(boxes, box)
		}
		for _, child := range current.Children() {
			walk(child)
		}
	}
	walk(root)
	return boxes
}
