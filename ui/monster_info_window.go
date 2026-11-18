package ui

import (
	"fmt"
	"image"

	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/widget"
	"github.com/kivutar/goro/network"
	"github.com/kivutar/goro/ui/rotheme"
)

type MonsterInfoView struct {
	Info    network.MonsterInfo
	Name    string
	MaxHP   uint32
	Preview image.Image
}

type MonsterInfoWindow struct {
	Window
	view MonsterInfoView
}

func (w *MonsterInfoWindow) OpenInfo(ctx Context, view MonsterInfoView) {
	w.EnsureWindow(monsterInfoWindowW, monsterInfoWindowH)
	w.view = view
	w.Window.Open(ctx, w.widgetTree(ctx))
	w.OnLayoutReload(func() {
		w.SetSize(monsterInfoWindowW, monsterInfoWindowH)
		w.SetContent(w.widgetTree(w.ctx))
	})
	w.Publish(ctx)
}

func (w *MonsterInfoWindow) Update(ctx Context) bool {
	w.EnsureWindow(monsterInfoWindowW, monsterInfoWindowH)
	if !w.IsOpen() {
		return false
	}
	consumed := w.Window.Update(ctx)
	w.Publish(ctx)
	return consumed
}

func (w *MonsterInfoWindow) widgetTree(ctx Context) widget.Widget {
	return Win(
		Title("Monster Information"),
		CloseButton(true),
		OnClose(func() {
			w.Close()
			w.Publish(ctx)
		}),
		Size(float32(monsterInfoWindowW), float32(monsterInfoWindowH)),
		Content(
			primitives.HBox(
				w.previewPanel(),
				primitives.Box(
					w.statsPanel(),
					w.elementsPanel(),
				).
					Width(float32(monsterInfoStatsW)).
					Gap(float32(monsterInfoStatsGap)),
			).
				PaddingXY(float32(monsterInfoPadX), float32(monsterInfoPadY)).
				Gap(float32(monsterInfoContentGap)).
				CrossAlign(primitives.CrossAxisCenter),
		),
		Footer(
			primitives.Expanded(primitives.Box()),
			rotheme.Button("OK", func() {
				w.Close()
				w.Publish(ctx)
			}),
		),
	)
}

func (w *MonsterInfoWindow) previewPanel() widget.Widget {
	return primitives.Box(
		newStaticImageWidget(w.view.Preview, monsterInfoPreviewW, monsterInfoPreviewH),
	).
		Width(float32(monsterInfoPreviewW)).
		Height(float32(monsterInfoPreviewH))
}

func (w *MonsterInfoWindow) statsPanel() widget.Widget {
	info := w.view.Info
	return primitives.Box(
		monsterInfoStatRow("Name", w.displayName(), "Size", monsterSizeName(info.Size)),
		monsterInfoStatRow("Level", fmt.Sprintf("%d", info.Level), "Race", monsterRaceName(info.Race)),
		monsterInfoStatRow("HP", monsterHPText(info.HP, w.view.MaxHP), "MDEF", fmt.Sprintf("%d", info.MagicDefense)),
		monsterInfoStatRow("DEF", fmt.Sprintf("%d", info.Defense), "Property", monsterPropertyName(info.Property)),
	)
}

func (w *MonsterInfoWindow) displayName() string {
	if w.view.Name == "" {
		return fmt.Sprintf("Monster %d", w.view.Info.Class)
	}
	return w.view.Name
}

func monsterInfoStatRow(leftLabel, leftValue, rightLabel, rightValue string) widget.Widget {
	return primitives.HBox(
		monsterInfoStatLabel(leftLabel, float32(monsterInfoNameLabelW)),
		monsterInfoStatValue(leftValue, float32(monsterInfoNameValueW)),
		monsterInfoStatLabel(rightLabel, float32(monsterInfoExtraLabelW)),
		monsterInfoStatValue(rightValue, float32(monsterInfoExtraValueW)),
	).
		Height(float32(monsterInfoRowH)).
		CrossAlign(primitives.CrossAxisCenter)
}

func monsterInfoStatLabel(text string, width float32) widget.Widget {
	return primitives.Box(
		rotheme.Text(text).Color(rotheme.Default.Colors.MutedText),
	).Width(width)
}

func monsterInfoStatValue(text string, width float32) widget.Widget {
	return primitives.Box(
		rotheme.Text(text).MaxLines(1),
	).Width(width)
}

func (w *MonsterInfoWindow) elementsPanel() widget.Widget {
	e := w.view.Info.Elements
	return primitives.Box(
		monsterInfoElementRow(
			monsterElementCell("Water", e.Water),
			monsterElementCell("Wind", e.Wind),
			monsterElementCell("Shadow", e.Shadow),
		),
		monsterInfoElementRow(
			monsterElementCell("Earth", e.Earth),
			monsterElementCell("Poison", e.Poison),
			monsterElementCell("Ghost", e.Ghost),
		),
		monsterInfoElementRow(
			monsterElementCell("Fire", e.Fire),
			monsterElementCell("Holy", e.Holy),
			monsterElementCell("Undead", e.Undead),
		),
	).Gap(float32(monsterInfoElementRowGap))
}

func monsterInfoElementRow(cells ...widget.Widget) widget.Widget {
	return primitives.HBox(cells...).Gap(float32(monsterInfoElementGap)).Height(float32(monsterInfoElementH))
}

func monsterElementCell(name string, rate uint8) widget.Widget {
	textColor := rotheme.Default.Colors.Text
	if rate < 100 {
		textColor = Color(ErrorTextColor)
	} else if rate > 100 {
		textColor = Color(GoodTextColor)
	}
	label := rotheme.Text(fmt.Sprintf("%s: %d", name, rate)).
		Color(textColor).
		Align(widget.TextAlignCenter)
	if rate > 100 {
		label.Bold()
	}
	return primitives.HBox(
		primitives.Expanded(label),
	).
		Width(float32(monsterInfoElementW)).
		Height(float32(monsterInfoElementH)).
		CrossAlign(primitives.CrossAxisCenter).
		Background(rotheme.Default.Colors.PanelBody).
		BorderStyle(float32(monsterInfoElementBorder), rotheme.Default.Colors.WindowBorder).
		Rounded(float32(monsterInfoElementRadius))
}

func monsterHPText(current, maximum uint32) string {
	if maximum == 0 || maximum < current {
		return fmt.Sprintf("%d", current)
	}
	return fmt.Sprintf("%d / %d", current, maximum)
}

func monsterSizeName(size uint16) string {
	if int(size) < len(monsterSizeNames) {
		return monsterSizeNames[size]
	}
	return fmt.Sprintf("Unknown (%d)", size)
}

func monsterRaceName(race uint16) string {
	if int(race) < len(monsterRaceNames) {
		return monsterRaceNames[race]
	}
	return fmt.Sprintf("Unknown (%d)", race)
}

func monsterPropertyName(property uint16) string {
	if int(property) < len(monsterPropertyNames) {
		return monsterPropertyNames[property]
	}
	return fmt.Sprintf("Unknown (%d)", property)
}

var monsterSizeNames = [...]string{"Small", "Medium", "Large"}

var monsterRaceNames = [...]string{
	"Formless",
	"Undead",
	"Brute",
	"Plant",
	"Insect",
	"Fish",
	"Demon",
	"Demi-human",
	"Angel",
	"Dragon",
}

var monsterPropertyNames = [...]string{
	"Neutral",
	"Water",
	"Earth",
	"Fire",
	"Wind",
	"Poison",
	"Holy",
	"Shadow",
	"Ghost",
	"Undead",
}
