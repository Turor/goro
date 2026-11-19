package ui

import (
	"fmt"
	"sort"

	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/state"
	"github.com/gogpu/ui/widget"
	"github.com/kivutar/goro/network"
	"github.com/kivutar/goro/res"
	"github.com/kivutar/goro/ui/rotheme"
)

type AchievementWindow struct {
	Window
	List        network.AchievementList
	category    int
	selectedID  uint32
	message     string
	selectedRow state.Signal[int]
	scrollY     state.Signal[float32]
}

func (w *AchievementWindow) Open(ctx Context) {
	w.EnsureWindow(achievementWindowW, achievementWindowH)
	if w.selectedRow == nil {
		w.selectedRow = state.NewSignal(-1)
	}
	if w.scrollY == nil {
		w.scrollY = state.NewSignal[float32](0)
	}
	w.Window.Open(ctx, w.widgetTree(ctx))
	w.OnLayoutReload(func() {
		w.SetSize(achievementWindowW, achievementWindowH)
		w.SetContent(w.widgetTree(w.ctx))
	})
	w.Publish(ctx)
}

func (w *AchievementWindow) SetMessage(ctx Context, message string) {
	w.message = message
	if w.IsOpen() {
		w.SetContent(w.widgetTree(ctx))
		w.Publish(ctx)
	}
}

func (w *AchievementWindow) SetList(ctx Context, list network.AchievementList) {
	w.List = list
	if len(w.List.Items) > 0 && w.selectedID == 0 {
		w.selectedID = w.List.Items[0].ID
	}
	if w.IsOpen() {
		w.SetContent(w.widgetTree(ctx))
		w.Publish(ctx)
	} else {
		w.Open(ctx)
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
	return Win(
		Title(msgOr(ctx, 2646, "Achievement Challenges")),
		CloseButton(true),
		OnClose(w.Close),
		Size(float32(achievementWindowW), float32(achievementWindowH)),
		Content(w.content(ctx)),
	)
}

func (w *AchievementWindow) content(ctx Context) widget.Widget {
	header := primitives.Box(
		rotheme.Label(fmt.Sprintf("Rank %d", w.List.Rank)).MaxLines(1),
		rotheme.Label(fmt.Sprintf("Points %d", w.List.Points)).MaxLines(1),
		rotheme.Label(fmt.Sprintf("Progress %d / %d", w.List.RankPoints, w.List.NextPoints)).MaxLines(1),
	).Gap(float32(achievementGap))

	cats := achievementCategories(ctx)
	var catButtons []widget.Widget
	for i, name := range cats {
		i, name := i, name
		catButtons = append(catButtons, rotheme.Button(name, func() {
			w.category = i
			w.scrollY.Set(0)
			w.SetContent(w.widgetTree(ctx))
			w.Publish(ctx)
		}))
	}
	sidebar := primitives.Box(catButtons...).Width(float32(achievementSidebarW)).Gap(float32(achievementGap))

	if w.category == 0 {
		return w.overview(ctx, header, sidebar)
	}
	rows := w.filteredRows(ctx)
	titles := make([]string, len(rows))
	for i, a := range rows {
		titles[i] = achievementTitle(ctx, a)
		if a.Completed {
			titles[i] += "  (Complete)"
		}
	}
	list := rotheme.TableView(
		rotheme.TableViewColumns([]rotheme.TableViewColumn{{Key: "ach", Flex: 1}}),
		rotheme.TableViewShowHeader(false),
		rotheme.TableViewRowCount(len(rows)),
		rotheme.TableViewRowHeight(float32(achievementRowH)),
		rotheme.TableViewEmptyText("No achievements"),
		rotheme.TableViewScrollYSignal(w.scrollY),
		rotheme.TableViewSelectedRow(w.selectedRow),
		rotheme.TableViewDispatchHoverToCells(false),
		rotheme.TableViewBuildSimpleCell(func(cell rotheme.TableViewCellContext) rotheme.TableViewSimpleCell {
			return rotheme.TableViewSimpleCell{Text: titles[cell.Row]}
		}),
		rotheme.TableViewOnRowClick(func(row int) {
			if row >= 0 && row < len(rows) {
				w.selectedID = rows[row].ID
				w.SetContent(w.widgetTree(ctx))
				w.Publish(ctx)
			}
		}),
	)

	detail := w.detail(ctx)
	body := primitives.HBox(
		sidebar,
		primitives.Expanded(primitives.Box(header, primitives.Expanded(list), detail).Gap(float32(achievementGap))),
	).Gap(float32(achievementGap))
	return primitives.Box(body).Padding(float32(achievementPad))
}

func (w *AchievementWindow) overview(ctx Context, header, sidebar widget.Widget) widget.Widget {
	counts := map[int][2]int{}
	for _, a := range w.List.Items {
		info := achievementMeta(ctx, a.ID)
		if info.Major <= 0 {
			continue
		}
		c := counts[info.Major]
		c[1]++
		if a.Completed {
			c[0]++
		}
		counts[info.Major] = c
	}
	cats := achievementCategories(ctx)
	lines := []widget.Widget{rotheme.Label(msgOr(ctx, 2656, "General")).MaxLines(1)}
	for i := 1; i < len(cats); i++ {
		c := counts[i]
		lines = append(lines, rotheme.Label(fmt.Sprintf("%s  %d/%d", cats[i], c[0], c[1])).MaxLines(1))
	}
	body := primitives.HBox(
		sidebar,
		primitives.Expanded(primitives.Box(append([]widget.Widget{header}, lines...)...).Gap(float32(achievementGap))),
	).Gap(float32(achievementGap))
	return primitives.Box(body).Padding(float32(achievementPad))
}

func (w *AchievementWindow) detail(ctx Context) widget.Widget {
	a, ok := w.selected(ctx)
	if !ok {
		return rotheme.Label("Select an achievement.")
	}
	info := achievementMeta(ctx, a.ID)
	rows := []widget.Widget{
		rotheme.Label(achievementTitle(ctx, a)).MaxLines(2),
	}
	if info.Summary != "" {
		rows = append(rows, rotheme.Label(info.Summary).MaxLines(3))
	}
	if info.Contents != "" {
		rows = append(rows, rotheme.Label(info.Contents).MaxLines(6))
	}
	for i, count := range a.Objectives {
		if count == 0 {
			continue
		}
		rows = append(rows, rotheme.Label(fmt.Sprintf("Objective %d: %d", i+1, count)).MaxLines(1))
	}
	if w.message != "" {
		rows = append(rows, rotheme.Label(w.message).MaxLines(1))
	}
	if a.Completed && !a.Rewarded {
		rows = append(rows, rotheme.Button(msgOr(ctx, 2653, "Claim Reward"), func() {
			if ctx.Network != nil {
				_ = ctx.Network.SendAchievementReward(a.ID)
			}
		}))
	} else if a.Rewarded {
		rows = append(rows, rotheme.Label("Reward claimed."))
	}
	return primitives.Box(rows...).Gap(float32(achievementGap)).Height(float32(achievementDetailH))
}

func (w *AchievementWindow) selected(ctx Context) (network.Achievement, bool) {
	for _, a := range w.List.Items {
		if a.ID == w.selectedID {
			return a, true
		}
	}
	if len(w.List.Items) > 0 {
		return w.List.Items[0], true
	}
	return network.Achievement{}, false
}

func (w *AchievementWindow) filteredRows(ctx Context) []network.Achievement {
	items := append([]network.Achievement(nil), w.List.Items...)
	sort.Slice(items, func(i, j int) bool { return items[i].ID < items[j].ID })
	var out []network.Achievement
	for _, a := range items {
		info := achievementMeta(ctx, a.ID)
		if info.Major == w.category {
			out = append(out, a)
		}
	}
	return out
}

func achievementMeta(ctx Context, id uint32) res.AchievementInfo {
	if ctx.Resources == nil {
		return res.AchievementInfo{}
	}
	info, _ := ctx.Resources.AchievementInfo(id)
	return info
}

func achievementTitle(ctx Context, a network.Achievement) string {
	info := achievementMeta(ctx, a.ID)
	if info.Title != "" {
		return info.Title
	}
	return fmt.Sprintf("Achievement %d", a.ID)
}

func achievementCategories(ctx Context) []string {
	return []string{
		msgOr(ctx, 2656, "General"),
		msgOr(ctx, 2657, "Adventure"),
		msgOr(ctx, 2661, "Battle"),
		msgOr(ctx, 2668, "Memorial"),
		msgOr(ctx, 2671, "Feat"),
		msgOr(ctx, 2674, "Quest"),
		msgOr(ctx, 2678, "Event"),
	}
}

func msgOr(ctx Context, id int, fallback string) string {
	if ctx.Resources == nil {
		return fallback
	}
	if s, ok := ctx.Resources.MsgString(id); ok && s != "" {
		return s
	}
	return fallback
}
