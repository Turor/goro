package ui

import (
	"fmt"
	"strings"

	"github.com/gogpu/ui/core/textfield"
	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/widget"
	"github.com/kivutar/goro/glog"
	"github.com/kivutar/goro/input"
	"github.com/kivutar/goro/network"
	"github.com/kivutar/goro/ui/rotheme"
)

type PetInfoWindow struct {
	Window
	property  network.PetProperty
	name      string
	nameField *textfield.Widget
}

func (w *PetInfoWindow) OpenInfo(ctx Context, property network.PetProperty) {
	w.EnsureWindow(petInfoWindowW, petInfoWindowH)
	// Status updates reuse the editor so a hunger tick cannot interrupt typing.
	if !w.IsOpen() || property.Modified || property.Name != w.property.Name || property.Job != w.property.Job {
		if w.nameField != nil {
			releaseWindowFocus(ctx, w.nameField)
		}
		w.nameField = nil
	}
	if w.nameField == nil || !w.nameField.IsFocused() {
		w.name = property.Name
		if w.nameField != nil {
			w.nameField.SetText(w.name)
		}
	}
	w.property = property
	w.Window.Open(ctx, w.widgetTree(ctx))
	w.OnLayoutReload(func() {
		w.SetSize(petInfoWindowW, petInfoWindowH)
		w.SetContent(w.widgetTree(w.ctx))
	})
	w.Publish(ctx)
}

func (w *PetInfoWindow) Update(ctx Context) bool {
	w.EnsureWindow(petInfoWindowW, petInfoWindowH)
	if !w.IsOpen() {
		return false
	}
	if w.submitFromFocusedEnter(ctx) {
		w.Publish(ctx)
		return true
	}
	consumed := w.Window.Update(ctx)
	w.Publish(ctx)
	return consumed
}

func (w *PetInfoWindow) KeyboardShortcutsBlocked() bool {
	return w.IsOpen() && w.nameField != nil && w.nameField.IsFocused()
}

func (w *PetInfoWindow) widgetTree(ctx Context) widget.Widget {
	return Win(
		Title("Pet Info"),
		CloseButton(true),
		OnClose(func() {
			w.Close()
			w.Publish(ctx)
		}),
		Size(float32(petInfoWindowW), float32(petInfoWindowH)),
		Content(
			primitives.Box(
				w.nameRow(ctx),
				w.infoRow("Level", fmt.Sprintf("%d", w.property.Level)),
				w.infoRow("Hunger", petHungerText(w.property.Fullness)),
				w.infoRow("Intimacy", petIntimacyText(w.property.Relationship)),
				w.infoRow("Accessory", w.accessoryText(ctx)),
			).
				Padding(float32(petInfoContentPad)).
				Gap(float32(petInfoRowGap)),
		),
	)
}

func (w *PetInfoWindow) nameRow(ctx Context) widget.Widget {
	if w.property.Modified {
		return w.infoRow("Name", w.property.Name)
	}
	return primitives.HBox(
		w.rowLabel("Name"),
		primitives.Box(w.nameInput(ctx)).
			Width(float32(petInfoFieldW)).
			Height(float32(petInfoNameFieldH)).
			CrossAlign(primitives.CrossAxisStretch),
		rotheme.Button("Edit", func() {
			w.rename(ctx)
		}).Width(float32(petInfoRenameButtonW)),
	).
		Gap(float32(petInfoRowGap)).
		CrossAlign(primitives.CrossAxisCenter).
		Height(float32(petInfoNameFieldH))
}

func (w *PetInfoWindow) nameInput(ctx Context) *textfield.Widget {
	if w.nameField != nil {
		return w.nameField
	}
	w.nameField = rotheme.TextField(
		w.name,
		textfield.TypeText,
		func(value string) {
			w.name = value
		},
		func(string) {
			w.rename(ctx)
		},
		textfield.MaxLength(24), // theme:ok name length limit
	)
	return w.nameField
}

func (w *PetInfoWindow) infoRow(label, value string) widget.Widget {
	return primitives.HBox(
		w.rowLabel(label),
		primitives.Box(rotheme.Text(value)).
			Width(float32(petInfoFieldW)),
	).
		Gap(float32(petInfoRowGap)).
		CrossAlign(primitives.CrossAxisCenter).
		Height(float32(petInfoRowH))
}

func (w *PetInfoWindow) rowLabel(label string) widget.Widget {
	return primitives.Box(
		rotheme.Text(label).
			Color(rotheme.Default.Colors.MutedText),
	).
		Width(float32(petInfoLabelW))
}

func (w *PetInfoWindow) accessoryText(ctx Context) string {
	if w.property.AccessoryID == 0 {
		return "Unequipped"
	}
	if ctx.Resources != nil {
		if name, ok := ctx.Resources.ItemDisplayName(int(w.property.AccessoryID), true); ok && strings.TrimSpace(name) != "" {
			return name
		}
	}
	return fmt.Sprintf("Item %d", w.property.AccessoryID)
}

func (w *PetInfoWindow) rename(ctx Context) {
	name := strings.TrimSpace(w.name)
	if name == "" || name == strings.TrimSpace(w.property.Name) || ctx.Network == nil {
		return
	}
	if err := ctx.Network.SendRenamePet(name); err != nil {
		glog.Warnf("pet rename failed: %v", err)
		return
	}
}

func (w *PetInfoWindow) submitFromFocusedEnter(ctx Context) bool {
	if ctx.Input == nil || w.nameField == nil || !w.nameField.IsFocused() {
		return false
	}
	if !ctx.Input.JustPressed(input.KeyEnter) {
		return false
	}
	w.rename(ctx)
	return true
}

func petHungerText(value uint16) string {
	switch {
	case value < 10:
		return "Very Hungry"
	case value < 25:
		return "Hungry"
	case value < 75:
		return "Satisfied"
	case value < 90:
		return "Stuffed"
	default:
		return "Full"
	}
}

func petIntimacyText(value uint16) string {
	switch {
	case value < 100:
		return "Awkward"
	case value < 250:
		return "Shy"
	case value < 600:
		return "Neutral"
	case value < 900:
		return "Cordial"
	default:
		return "Loyal"
	}
}
