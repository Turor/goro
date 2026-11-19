package ui

import (
	"fmt"

	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/widget"
	"github.com/kivutar/goro/session"
	"github.com/kivutar/goro/ui/rotheme"
)

const (
	gearStatTabBasic = 0
	gearStatTabTrait = 1
)

// gearWindowSize is the player Gear window. The paperdoll keeps its width.
// The status column sits beside it and is tall enough for the open attribute
// tab plus the composite stats below it. Viewed equipment stays on the
// paperdoll size.
func gearWindowSize() (int, int) {
	width := equipmentWindowWidth + equipmentWindowPad + gearStatsColumnWidth()
	contentH := equipmentWindowContentH
	if statsH := gearStatsContentHeight(); statsH > contentH {
		contentH = statsH
	}
	return width, ROWindowTitleHeight + contentH + ROWindowFooterHeight
}

func gearStatsColumnWidth() int {
	width := int(statsPrimaryColumnWidth)
	if derived := int(statsDerivedColumnWidth); derived > width {
		width = derived
	}
	return width
}

func gearStatsContentHeight() int {
	rows := gearStatChromeRows + statsPrimaryRows + gearCompositeRows
	gaps := rows - 1
	if gaps < 0 {
		gaps = 0
	}
	return int(statsWindowPad*2) + rows*statsRowH + int(float32(gaps)*rotheme.TableGap)
}

func gearContentHeight() int {
	_, height := gearWindowSize()
	contentH := height - ROWindowTitleHeight - ROWindowFooterHeight
	if contentH < 1 {
		return 1
	}
	return contentH
}

func gearStatsColumn(ctx Context, tab int, onTab func(int)) widget.Widget {
	stats := sessionStats(ctx.Session)
	stub := &StatsWindow{}
	points := fmt.Sprintf("Points %d", stats.Points)
	allocation := stub.statRowsWidget(ctx)
	if tab == gearStatTabTrait {
		points = fmt.Sprintf("Trait Points %d", stats.TraitPoints)
		allocation = stub.traitRowsWidget(ctx)
	}
	body := primitives.Box(
		gearStatTabs(tab, onTab),
		rotheme.Text(points).MaxLines(1),
		allocation,
		rotheme.Label("Composite"),
		gearCompositeLines(stats),
	).Gap(rotheme.TableGap).Padding(statsWindowPad)
	return primitives.Box(body).
		Width(float32(gearStatsColumnWidth())).
		Height(float32(gearContentHeight()))
}

func gearStatTabs(tab int, onTab func(int)) widget.Widget {
	return primitives.HBox(
		gearStatTabButton("Basic", tab == gearStatTabBasic, func() {
			if onTab != nil {
				onTab(gearStatTabBasic)
			}
		}),
		gearStatTabButton("4th", tab == gearStatTabTrait, func() {
			if onTab != nil {
				onTab(gearStatTabTrait)
			}
		}),
	).Gap(rotheme.TableGap)
}

func gearStatTabButton(label string, active bool, onClick func()) widget.Widget {
	button := rotheme.Button(label, onClick)
	if active {
		button.Background(rotheme.Default.Colors.ButtonHover)
	}
	return button
}

func gearCompositeLines(stats session.Stats) widget.Widget {
	lines := []string{
		fmt.Sprintf("ATK %d + %d", stats.Attack, stats.AttackBonus),
		fmt.Sprintf("MATK %d - %d", stats.MatkMin, stats.MatkMax),
		fmt.Sprintf("DEF %d + %d", stats.Defense, stats.DefenseBonus),
		fmt.Sprintf("MDEF %d + %d", stats.MDefense, stats.MDefenseBonus),
		fmt.Sprintf("HIT %d", stats.Hit),
		fmt.Sprintf("FLEE %d + %d", stats.Flee, stats.FleeBonus),
		fmt.Sprintf("CRIT %d", stats.Critical),
		fmt.Sprintf("ASPD %d", displayASPD(stats.ASPD+stats.ASPDBonus)),
		fmt.Sprintf("P.ATK %d", stats.PAtk),
		fmt.Sprintf("S.MATK %d", stats.SMatk),
		fmt.Sprintf("RES %d", stats.Res),
		fmt.Sprintf("MRES %d", stats.MRes),
		fmt.Sprintf("H.PLUS %d", stats.HPlus),
		fmt.Sprintf("C.RATE %d", stats.CRate),
	}
	children := make([]widget.Widget, 0, len(lines))
	for _, line := range lines {
		children = append(children, rotheme.Text(line).MaxLines(1))
	}
	return primitives.Box(children...).Gap(rotheme.TableGap)
}

func (w *EquipmentWindow) OpenIfClosed(ctx Context) {
	w.EnsureWindow(gearWindowSize())
	if w.IsOpen() {
		w.Raise(ctx)
		w.Publish(ctx)
		return
	}
	w.Toggle(ctx)
}

func gearSnapshot(s *session.Session) string {
	return equipmentSnapshot(s) + "\n" + statsWindowSnapshot(s)
}
