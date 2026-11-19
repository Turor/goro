package ui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/kivutar/goro/res"
	"github.com/kivutar/goro/session"
)

var itemGradeLabels = []string{"", "D", "C", "B", "A"}

func itemGradeLabel(grade uint8) string {
	if grade == 0 {
		return ""
	}
	if int(grade) < len(itemGradeLabels) {
		return "[" + itemGradeLabels[grade] + "]"
	}
	return "[" + strconv.Itoa(int(grade)) + "]"
}

func visibleRandomOptions(item session.InventoryItem) []session.ItemOption {
	if !item.Identified {
		return nil
	}
	out := make([]session.ItemOption, 0, len(item.Options))
	for _, opt := range item.Options {
		if opt.Index == 0 || res.IsForgeMetaOption(opt.Index) {
			continue
		}
		out = append(out, opt)
	}
	return out
}

func decorateIdentifiedItemName(name string, item session.InventoryItem) string {
	if !item.Identified {
		return name
	}
	if label := itemGradeLabel(item.Grade); label != "" {
		name = label + " " + name
	}
	if n := len(visibleRandomOptions(item)); n > 0 {
		name += fmt.Sprintf(" [%d Option]", n)
	}
	return name
}

func formatRandomOptionLine(manager *res.Manager, opt session.ItemOption) string {
	template := "UNKNOWN RANDOM OPTION"
	if manager != nil {
		if name, ok := manager.RandomOptionName(int(opt.Index)); ok {
			template = name
		}
	}
	value := strconv.Itoa(int(opt.Value))
	if strings.Contains(template, "%d") {
		return strings.ReplaceAll(strings.ReplaceAll(template, "%d", value), "%%", "%")
	}
	return template + " " + value
}

func itemRandomOptionLines(manager *res.Manager, item session.InventoryItem) []itemInfoTextLine {
	options := visibleRandomOptions(item)
	if len(options) == 0 {
		return nil
	}
	lines := make([]itemInfoTextLine, 0, len(options)+1)
	for _, opt := range options {
		lines = append(lines, itemInfoTextLine{Runs: []itemInfoTextRun{{
			Text:  formatRandomOptionLine(manager, opt),
			Color: inventoryTextColor,
		}}})
	}
	lines = append(lines, itemInfoTextLine{})
	return lines
}
