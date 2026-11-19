package ui

import (
	"fmt"

	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/widget"
	"github.com/kivutar/goro/ui/rotheme"
)

type AttendanceWindow struct {
	Window
	day     int
	claimed bool
	message string
}

func (w *AttendanceWindow) Open(ctx Context, data int64) {
	w.EnsureWindow(attendanceWindowW, attendanceWindowH)
	w.day = int(data / 10)
	w.claimed = data%10 != 0
	if w.claimed {
		w.message = fmt.Sprintf("Day %d claimed.", w.day)
	} else {
		w.message = fmt.Sprintf("Day %d ready to claim.", w.day+1)
	}
	w.Window.Open(ctx, w.widgetTree(ctx))
	w.OnLayoutReload(func() {
		w.SetSize(attendanceWindowW, attendanceWindowH)
		w.SetContent(w.widgetTree(w.ctx))
	})
	w.Publish(ctx)
}

func (w *AttendanceWindow) SetProgress(ctx Context, day int) {
	w.day = day
	w.claimed = true
	w.message = fmt.Sprintf("Claimed day %d.", day)
	if w.IsOpen() {
		w.SetContent(w.widgetTree(ctx))
		w.Publish(ctx)
	}
}

func (w *AttendanceWindow) Update(ctx Context) bool {
	w.EnsureWindow(attendanceWindowW, attendanceWindowH)
	if !w.IsOpen() {
		return false
	}
	consumed := w.Window.Update(ctx)
	w.Publish(ctx)
	return consumed
}

func (w *AttendanceWindow) widgetTree(ctx Context) widget.Widget {
	return Win(
		Title("Attendance"),
		CloseButton(true),
		OnClose(func() { w.Close(); w.Publish(ctx) }),
		Size(float32(attendanceWindowW), float32(attendanceWindowH)),
		Content(primitives.Box(rotheme.Label(w.message)).Padding(float32(attendancePad))),
		Footer(rotheme.Button("Claim", func() {
			if ctx.Network != nil && !w.claimed {
				_ = ctx.Network.SendAttendanceReward()
			}
		})),
	)
}
