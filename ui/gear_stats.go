package ui

import (
	"fmt"

	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/widget"
	"github.com/kivutar/goro/session"
	"github.com/kivutar/goro/ui/rotheme"
)

// gearWindowSize is the player Gear window: the paperdoll on the left and the
// status column on the right. The viewed-equipment pop-up stays on the
// paperdoll size.
func gearWindowSize() (int, int) {
	width := equipmentWindowWidth + equipmentWindowPad + int(statsPrimaryColumnWidth)
	rows := 2 + statsPrimaryRows + 1 + 8 + 1 + statsPrimaryRows + 2
	height := ROWindowTitleHeight + ROWindowFooterHeight + equipmentWindowPad*2 + rows*statsRowH + rows*int(rotheme.TableGap)
	if height < equipmentWindowHeight {
		height = equipmentWindowHeight
	}
	return width, height
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

func gearStatsColumn(ctx Context) widget.Widget {
	stats := sessionStats(ctx.Session)
	stub := &StatsWindow{}
	return primitives.Box(
		rotheme.Label("Base"),
		rotheme.Text(fmt.Sprintf("Points %d", stats.Points)).MaxLines(1),
		stub.statRowsWidget(ctx),
		rotheme.Label("Combat"),
		gearCombatLines(stats),
		rotheme.Label("Traits"),
		stub.traitRowsWidget(ctx),
	).Width(statsPrimaryColumnWidth).Gap(rotheme.TableGap)
}

func gearCombatLines(stats session.Stats) widget.Widget {
	lines := []string{
		fmt.Sprintf("ATK %d + %d", stats.Attack, stats.AttackBonus),
		fmt.Sprintf("MATK %d - %d", stats.MatkMin, stats.MatkMax),
		fmt.Sprintf("DEF %d + %d", stats.Defense, stats.DefenseBonus),
		fmt.Sprintf("MDEF %d + %d", stats.MDefense, stats.MDefenseBonus),
		fmt.Sprintf("HIT %d", stats.Hit),
		fmt.Sprintf("FLEE %d + %d", stats.Flee, stats.FleeBonus),
		fmt.Sprintf("CRIT %d", stats.Critical),
		fmt.Sprintf("ASPD %d", displayASPD(stats.ASPD+stats.ASPDBonus)),
		fmt.Sprintf("P.ATK %d  S.MATK %d  RES %d", stats.PAtk, stats.SMatk, stats.Res),
		fmt.Sprintf("MRES %d  H.PLUS %d  C.RATE %d", stats.MRes, stats.HPlus, stats.CRate),
	}
	children := make([]widget.Widget, 0, len(lines))
	for _, line := range lines {
		children = append(children, rotheme.Text(line).MaxLines(1).Ellipsis())
	}
	return primitives.Box(children...).Gap(rotheme.TableGap)
}
