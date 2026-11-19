package ui

import (
	"fmt"

	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/widget"
	"github.com/kivutar/goro/network"
	"github.com/kivutar/goro/ui/rotheme"
)

type AchievementWindow struct {
	Window
	List    network.AchievementList
	message string
}

func (w *AchievementWindow) Open(ctx Context) {
	w.EnsureWindow(achievementWindowW, achievementWindowH)
	w.message = ""
	w.Window.Open(ctx, w.widgetTree(ctx))
	w.OnLayoutReload(func() {
		w.SetSize(achievementWindowW, achievementWindowH)
		w.SetContent(w.widgetTree(w.ctx))
	})
	w.Publish(ctx)
}

func (w *AchievementWindow) SetList(ctx Context, list network.AchievementList) {
	w.List = list
	if !w.IsOpen() {
		w.Open(ctx)
		return
	}
	w.SetContent(w.widgetTree(ctx))
	w.Publish(ctx)
}

func (w *AchievementWindow) SetMessage(ctx Context, message string) {
	w.message = message
	if w.IsOpen() {
		w.SetContent(w.widgetTree(ctx))
		w.Publish(ctx)
	}
}

func (w *AchievementWindow) Update(ctx Context) bool {
	w.EnsureWindow(achievementWindowW, achievementWindowH)
	if !w.IsOpen() {
		return false
	}
	consumed := w.Window.Update(ctx)
	w.Publish(ctx)
	return consumed
}

func (w *AchievementWindow) widgetTree(ctx Context) widget.Widget {
	rows := []widget.Widget{
		rotheme.Label(fmt.Sprintf("Points %d  Rank %d", w.List.Points, w.List.Rank)),
	}
	if w.message != "" {
		rows = append(rows, rotheme.Label(w.message))
	}
	for _, a := range w.List.Items {
		a := a
		label := fmt.Sprintf("#%d", a.ID)
		if a.Completed {
			label += " done"
		}
		if a.Rewarded {
			label += " claimed"
		}
		rows = append(rows, rotheme.Button(label, func() {
			if ctx.Network != nil && a.Completed && !a.Rewarded {
				_ = ctx.Network.SendAchievementReward(a.ID)
			}
		}).Height(float32(achievementRowH)))
	}
	return Win(
		Title("Achievements"),
		CloseButton(true),
		OnClose(func() { w.Close(); w.Publish(ctx) }),
		Size(float32(achievementWindowW), float32(achievementWindowH)),
		Content(primitives.Box(rows...).Padding(float32(achievementPad)).Gap(float32(achievementGap))),
	)
}
