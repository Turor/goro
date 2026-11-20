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
// The status column sits beside it. Composite stats are a full-width band
// under both, not a list in the column. Viewed equipment stays on the
// paperdoll size.
func gearWindowSize() (int, int) {
	width := equipmentWindowWidth + equipmentWindowPad + gearStatsColumnWidth()
	contentH := equipmentWindowContentH
	if statsH := gearStatsContentHeight(); statsH > contentH {
		contentH = statsH
	}
	contentH += gearCompositeBandHeight()
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
	rows := gearStatChromeRows + statsPrimaryRows
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
	).Gap(rotheme.TableGap).Padding(statsWindowPad)
	return primitives.Box(body).
		Width(float32(gearStatsColumnWidth())).
		Height(float32(gearStatsContentHeight()))
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

func gearCompositeBandHeight() int {
	rows := gearCompositeBandRows
	if rows < 1 {
		rows = 1
	}
	gaps := rows - 1
	return int(statsWindowPad*2) + rows*statsRowH + int(float32(gaps)*rotheme.TableGap)
}

type gearCompositeStat struct {
	label string
	value string
}

func gearCompositeStats(stats session.Stats) []gearCompositeStat {
	return []gearCompositeStat{
		{label: "ATK", value: fmt.Sprintf("%d + %d", stats.Attack, stats.AttackBonus)},
		{label: "MATK", value: fmt.Sprintf("%d - %d", stats.MatkMin, stats.MatkMax)},
		{label: "HIT", value: fmt.Sprintf("%d", stats.Hit)},
		{label: "CRIT", value: fmt.Sprintf("%d", stats.Critical)},
		{label: "DEF", value: fmt.Sprintf("%d + %d", stats.Defense, stats.DefenseBonus)},
		{label: "MDEF", value: fmt.Sprintf("%d + %d", stats.MDefense, stats.MDefenseBonus)},
		{label: "FLEE", value: fmt.Sprintf("%d + %d", stats.Flee, stats.FleeBonus)},
		{label: "ASPD", value: fmt.Sprintf("%d", displayASPD(stats.ASPD+stats.ASPDBonus))},
		{label: "P.ATK", value: fmt.Sprintf("%d", stats.PAtk)},
		{label: "S.MATK", value: fmt.Sprintf("%d", stats.SMatk)},
		{label: "RES", value: fmt.Sprintf("%d", stats.Res)},
		{label: "MRES", value: fmt.Sprintf("%d", stats.MRes)},
		{label: "H.PLUS", value: fmt.Sprintf("%d", stats.HPlus)},
		{label: "C.RATE", value: fmt.Sprintf("%d", stats.CRate)},
	}
}

// gearCompositeBand is the classic status strip: equal boxes across the
// window, name in a header cell and the number beside it.
func gearCompositeBand(stats session.Stats) widget.Widget {
	all := gearCompositeStats(stats)
	return primitives.Box(
		rotheme.Label("Composite"),
		gearCompositeRow(all[:4]),
		gearCompositeRow(all[4:8]),
		gearCompositeRow(all[8:]),
	).Gap(rotheme.TableGap).Padding(statsWindowPad)
}

func gearCompositeRow(stats []gearCompositeStat) widget.Widget {
	cells := make([]widget.Widget, 0, len(stats))
	for _, stat := range stats {
		cells = append(cells, gearCompositeCell(stat))
	}
	return primitives.HBox(cells...).Gap(rotheme.TableGap)
}

func gearCompositeCell(stat gearCompositeStat) widget.Widget {
	return primitives.Expanded(
		primitives.HBox(
			primitives.Box(rotheme.Label(stat.label)).
				Background(rotheme.Default.Colors.ButtonHover).
				PaddingXY(rotheme.TableCellPadX, 0).
				Height(float32(statsRowH)),
			primitives.Expanded(
				primitives.Box(rotheme.Text(stat.value).MaxLines(1)).
					PaddingXY(rotheme.TableCellPadX, 0),
			),
		).
			Height(float32(statsRowH)).
			Background(rotheme.Default.Colors.WindowFooter).
			CrossAlign(primitives.CrossAxisCenter),
	)
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
