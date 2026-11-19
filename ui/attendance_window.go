package ui

import (
	"fmt"

	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/widget"
	"github.com/kivutar/goro/res"
	"github.com/kivutar/goro/ui/rotheme"
)

type AttendanceWindow struct {
	Window
	data    int64
	day     int
	claimed bool
}

func (w *AttendanceWindow) Open(ctx Context, data int64) {
	w.EnsureWindow(attendanceWindowW, attendanceWindowH)
	w.applyData(data)
	w.Window.Open(ctx, w.widgetTree(ctx))
	w.OnLayoutReload(func() {
		w.SetSize(attendanceWindowW, attendanceWindowH)
		w.SetContent(w.widgetTree(w.ctx))
	})
	w.Publish(ctx)
}

func (w *AttendanceWindow) SetProgress(ctx Context, data int) {
	w.Open(ctx, int64(data))
}

func (w *AttendanceWindow) applyData(data int64) {
	w.data = data
	w.claimed = data%10 != 0 // theme:ok attendance wire encoding (count*10 + claimed)
	w.day = int(data / 10)   // theme:ok attendance wire encoding
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
		Title(msgOr(ctx, 3349, "Attendance Check")),
		CloseButton(true),
		OnClose(func() { w.Close(); w.Publish(ctx) }),
		Size(float32(attendanceWindowW), float32(attendanceWindowH)),
		Content(w.content(ctx)),
		Footer(rotheme.Button("Check Reward", func() { w.claim(ctx) })),
	)
}

func (w *AttendanceWindow) content(ctx Context) widget.Widget {
	current := w.day + 1
	if current > attendanceDays {
		current = attendanceDays
	}
	status := fmt.Sprintf("Click the item to claim day %d reward", current)
	if w.claimed || w.day >= attendanceDays {
		status = fmt.Sprintf("%d Day attendance success", w.day)
	}
	remaining := attendanceDays - w.day
	if remaining < 0 {
		remaining = 0
	}
	header := primitives.Box(
		rotheme.Label("Event Period: Ongoing (20-day cycle, repeats)").MaxLines(1),
		rotheme.Label(status).MaxLines(1),
		rotheme.Label(fmt.Sprintf("Remaining days: %d", remaining)).MaxLines(1),
	).Gap(float32(attendanceGap))

	rewards := attendanceRewards(ctx)
	byDay := map[int]res.AttendanceReward{}
	for _, r := range rewards {
		byDay[r.Day] = r
	}

	var rows []widget.Widget
	for row := 0; row < attendanceDays/attendanceCols; row++ {
		var cells []widget.Widget
		for col := 0; col < attendanceCols; col++ {
			day := row*attendanceCols + col + 1
			cells = append(cells, w.dayCell(ctx, day, current, byDay[day]))
		}
		rows = append(rows, primitives.HBox(cells...).Gap(float32(attendanceGap)))
	}
	return primitives.Box(append([]widget.Widget{header}, rows...)...).Padding(float32(attendancePad)).Gap(float32(attendanceGap))
}

func (w *AttendanceWindow) dayCell(ctx Context, day, current int, reward res.AttendanceReward) widget.Widget {
	label := fmt.Sprintf("%d Day", day)
	if reward.ItemID != 0 {
		name := fmt.Sprintf("Item %d", reward.ItemID)
		if ctx.Resources != nil {
			if n, ok := ctx.Resources.ItemDisplayName(reward.ItemID, true); ok && n != "" {
				name = n
			}
		}
		label = fmt.Sprintf("%d Day\n%s x%d", day, name, reward.Quantity)
	}
	if day <= w.day {
		label += "\nChecked"
	} else if day == current && !w.claimed {
		label += "\nClaim"
	}
	btn := rotheme.Button(label, func() {
		if day == current && !w.claimed {
			w.claim(ctx)
		}
	})
	return primitives.Box(btn).Width(float32(attendanceCellW)).Height(float32(attendanceCellH))
}

func (w *AttendanceWindow) claim(ctx Context) {
	if ctx.Network != nil && !w.claimed {
		_ = ctx.Network.SendAttendanceReward()
	}
}

func attendanceRewards(ctx Context) []res.AttendanceReward {
	if ctx.Resources == nil {
		return nil
	}
	return ctx.Resources.AttendanceRewards()
}
