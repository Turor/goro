package ui

import (
	"testing"

	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/widget"
	"github.com/kivutar/goro/client"
	"github.com/kivutar/goro/input"
)

func TestWindowPressInsideGraceDoesNotDrag(t *testing.T) {
	app := &windowDragTestApp{}
	inputState := input.NewState()
	ctx := client.Context{Input: inputState, UIApp: app, ScreenW: 800, ScreenH: 600}
	window := NewWindow(100, 80)
	window.OpenAt(10, 20, primitives.Box())

	inputState.SetMousePosition(20, 25)
	inputState.SetMouseButton(input.MouseButtonLeft, true)
	if !window.Update(ctx) {
		t.Fatal("title press was not consumed")
	}
	if window.dragging || window.dragLayer || app.beginToken != nil {
		t.Fatal("press inside the grace distance started a drag")
	}

	inputState.EndFrame()
	inputState.SetMouseButton(input.MouseButtonLeft, false)
	if !window.Update(ctx) {
		t.Fatal("release inside the grace distance was not consumed")
	}
	if window.dragging || window.dragLayer || window.userMoved {
		t.Fatal("a click inside the grace distance dragged the window")
	}
}

func TestCoveredWindowDoesNotReceiveThePress(t *testing.T) {
	manager := NewManager()
	inputState := input.NewState()
	ctx := client.Context{Input: inputState, UIManager: manager, ScreenW: 800, ScreenH: 600}
	back := NewWindow(120, 90)
	back.OpenAt(10, 10, primitives.Box())
	back.Publish(ctx)
	front := NewWindow(80, 60)
	front.OpenAt(20, 20, primitives.Box())
	front.Publish(ctx)

	inputState.SetMousePosition(30, 30)
	inputState.SetMouseButton(input.MouseButtonLeft, true)
	if back.Update(ctx) {
		t.Fatal("covered window received the press")
	}
	if front.Update(ctx) != true || !front.dragPending {
		t.Fatal("top window did not arm the drag grace")
	}
	if back.dragPending || back.dragging {
		t.Fatal("covered window armed a drag")
	}
}

type pointerProbe struct {
	widget.WidgetBase
	events int
	accept bool
}

func (p *pointerProbe) Layout(widget.Context, geometry.Constraints) geometry.Size {
	return p.Bounds().Size()
}

func (p *pointerProbe) Draw(widget.Context, widget.Canvas) {}

func (p *pointerProbe) Event(widget.Context, event.Event) bool {
	p.events++
	return p.accept
}

func TestTopmostPointerTargetBlocksTheWidgetUnderIt(t *testing.T) {
	front := &pointerProbe{accept: false}
	front.SetVisible(true)
	front.SetEnabled(true)
	front.SetBounds(geometry.NewRect(0, 0, 40, 20))
	back := &pointerProbe{accept: true}
	back.SetVisible(true)
	back.SetEnabled(true)
	back.SetBounds(geometry.NewRect(0, 0, 40, 20))
	box := primitives.Box(back, front)
	box.SetBounds(geometry.NewRect(0, 0, 40, 20))
	box.Layout(widget.NewContext(), geometry.Constraints{MaxWidth: 40, MaxHeight: 20})
	front.SetBounds(geometry.NewRect(0, 0, 40, 20))
	back.SetBounds(geometry.NewRect(0, 0, 40, 20))

	mouse := event.NewMouseEvent(event.MousePress, event.ButtonLeft, event.ButtonStateLeft, geometry.Pt(8, 8), geometry.Pt(8, 8), event.ModNone)
	if !deliverTopmostPointer(widget.NewContext(), box, mouse) {
		t.Fatal("topmost target did not take the click")
	}
	if front.events != 1 {
		t.Fatalf("front events = %d, want 1", front.events)
	}
	if back.events != 0 {
		t.Fatalf("covered widget received %d click events", back.events)
	}
}
