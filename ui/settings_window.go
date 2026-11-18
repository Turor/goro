package ui

import (
	"fmt"

	"github.com/gogpu/ui/core/checkbox"
	"github.com/gogpu/ui/core/dropdown"
	"github.com/gogpu/ui/core/slider"
	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/widget"
	"github.com/kivutar/goro/client"
	"github.com/kivutar/goro/config"
	"github.com/kivutar/goro/glog"
	"github.com/kivutar/goro/ui/rotheme"
)

type settingsTab int

const (
	settingsTabDisplay settingsTab = iota
	settingsTabSound
	settingsTabGameplay
)

type SettingsWindow struct {
	Window
	tab settingsTab
}

func (w *SettingsWindow) OpenWindow(ctx client.Context) {
	w.EnsureWindow(settingsWindowW, settingsWindowH)
	w.ctx = ctx
	w.Open(ctx, w.widgetTree(ctx))
	w.OnLayoutReload(func() {
		w.SetSize(settingsWindowW, settingsWindowH)
		w.SetContent(w.widgetTree(w.ctx))
	})
	w.Publish(ctx)
}

func (w *SettingsWindow) Update(ctx client.Context) bool {
	w.EnsureWindow(settingsWindowW, settingsWindowH)
	w.ctx = ctx
	if !w.IsOpen() {
		return false
	}
	consumed := w.Window.Update(ctx)
	w.Publish(ctx)
	return consumed
}

func (w *SettingsWindow) Rebind(ctx client.Context) {
	if !w.IsOpen() {
		return
	}
	w.refresh(ctx)
}

func (w *SettingsWindow) widgetTree(ctx client.Context) widget.Widget {
	return Win(
		Title("Settings"),
		CloseButton(true),
		OnClose(w.Close),
		Size(float32(settingsWindowW), float32(settingsWindowH)),
		Content(w.contentTree(ctx)),
	)
}

func (w *SettingsWindow) contentTree(ctx client.Context) widget.Widget {
	return primitives.Box(
		w.tabStrip(),
		primitives.Box().Height(float32(settingsDividerH)).Background(rotheme.Default.Colors.WindowBorder),
		primitives.Expanded(w.tabContent(ctx)),
	).CrossAlign(primitives.CrossAxisStretch)
}

func (w *SettingsWindow) tabStrip() widget.Widget {
	var tabs []widget.Widget
	for _, def := range []struct {
		tab   settingsTab
		label string
	}{
		{settingsTabDisplay, "Display"},
		{settingsTabSound, "Sound"},
		{settingsTabGameplay, "Gameplay"},
	} {
		tabs = append(tabs, newTabWidget(tabWidgetConfig{
			label:  def.label,
			active: w.tab == def.tab,
			width:  settingsTabW,
			height: settingsTabH,
			onClick: func() {
				w.tab = def.tab
				w.refresh(w.ctx)
			},
		}))
	}
	tabs = append(tabs, primitives.Expanded(primitives.Box()))
	return primitives.HBox(tabs...).Gap(float32(settingsTabOverlap))
}

func (w *SettingsWindow) tabContent(ctx client.Context) widget.Widget {
	switch w.tab {
	case settingsTabSound:
		return w.soundTree(ctx)
	case settingsTabGameplay:
		return w.gameplayTree(ctx)
	default:
		return w.displayTree(ctx)
	}
}

func (w *SettingsWindow) displayTree(ctx client.Context) widget.Widget {
	return primitives.Box(
		rotheme.Checkbox(
			checkbox.CheckedFn(func() bool { return settingsRuntimeFullscreen(ctx) }),
			checkbox.LabelOpt("Fullscreen"),
			checkbox.OnToggle(func(enabled bool) {
				if ctx.Runtime != nil {
					ctx.Runtime.SetFullscreen(enabled)
				}
				w.saveSettings(ctx)
				w.refresh(ctx)
			}),
		),

		rotheme.Checkbox(
			checkbox.Checked(settingsRuntimeVSync(ctx)),
			checkbox.LabelOpt("VSync (Restart)"),
			checkbox.OnToggle(func(enabled bool) {
				if ctx.Runtime != nil {
					ctx.Runtime.SetVSync(enabled)
				}
				w.saveSettings(ctx)
				w.refresh(ctx)
			}),
		),

		rotheme.Checkbox(
			checkbox.Checked(settingsRuntimeFPS(ctx)),
			checkbox.LabelOpt("FPS meter"),
			checkbox.OnToggle(func(enabled bool) {
				if ctx.Runtime != nil {
					ctx.Runtime.SetFPS(enabled)
				}
				w.saveSettings(ctx)
				w.refresh(ctx)
			}),
		),

		primitives.HBox(
			primitives.Expanded(rotheme.Text("Anisotropic filtering")),
			rotheme.Button(anisotropyLabel(settingsAnisotropy(ctx)), func() {
				level := settingsAnisotropy(ctx) * 2 // theme:ok anisotropy step
				if level == 0 {
					level = 2 // theme:ok anisotropy minimum
				} else if level > 16 { // theme:ok anisotropy maximum
					level = 0
				}
				if ctx.Runtime != nil {
					ctx.Runtime.SetAnisotropy(level)
				}
				w.saveSettings(ctx)
				w.refresh(ctx)
			}).Width(float32(settingsOptionButtonW)),
		).Gap(float32(settingsGap)).Height(rotheme.Default.Typography.TextSize+rotheme.ButtonPaddingY*2),

		primitives.HBox(
			primitives.Expanded(rotheme.Text("Upscaling (Next map)")),
			rotheme.Button(textureUpscalingLabel(settingsTextureUpscaling(ctx)), func() {
				if ctx.Runtime != nil {
					ctx.Runtime.SetTextureUpscaling(!ctx.Runtime.TextureUpscaling())
				}
				w.saveSettings(ctx)
				w.refresh(ctx)
			}).Width(float32(settingsOptionButtonW)),
		).Gap(float32(settingsGap)).Height(rotheme.Default.Typography.TextSize+rotheme.ButtonPaddingY*2),

		rotheme.Checkbox(
			checkbox.Checked(settingsSmoothSprites(ctx)),
			checkbox.LabelOpt("Smooth world sprites"),
			checkbox.OnToggle(func(enabled bool) {
				if ctx.Runtime != nil {
					ctx.Runtime.SetSmoothSprites(enabled)
				}
				w.saveSettings(ctx)
				w.refresh(ctx)
			}),
		),

		rotheme.Checkbox(
			checkbox.Checked(settingsMSAA(ctx)),
			checkbox.LabelOpt("Anti-aliasing 4x (Restart)"),
			checkbox.OnToggle(func(enabled bool) {
				if ctx.Runtime != nil {
					ctx.Runtime.SetMSAA(enabled)
				}
				w.saveSettings(ctx)
				w.refresh(ctx)
			}),
		),

		rotheme.Checkbox(
			checkbox.Checked(settingsBloom(ctx)),
			checkbox.LabelOpt("Bloom"),
			checkbox.OnToggle(func(enabled bool) {
				if ctx.Runtime != nil {
					ctx.Runtime.SetBloom(enabled)
				}
				w.saveSettings(ctx)
				w.refresh(ctx)
			}),
		),

		rotheme.Checkbox(
			checkbox.Checked(settingsSSAO(ctx)),
			checkbox.LabelOpt("Ambient occlusion"),
			checkbox.OnToggle(func(enabled bool) {
				if ctx.Runtime != nil {
					ctx.Runtime.SetSSAO(enabled)
				}
				w.saveSettings(ctx)
				w.refresh(ctx)
			}),
		),

		rotheme.Checkbox(
			checkbox.Checked(settingsFXAA(ctx)),
			checkbox.LabelOpt("FXAA"),
			checkbox.OnToggle(func(enabled bool) {
				if ctx.Runtime != nil {
					ctx.Runtime.SetFXAA(enabled)
				}
				w.saveSettings(ctx)
				w.refresh(ctx)
			}),
		),

		rotheme.Checkbox(
			checkbox.Checked(settingsEnhancedLighting(ctx)),
			checkbox.LabelOpt("Enhanced lighting"),
			checkbox.OnToggle(func(enabled bool) {
				if ctx.Runtime != nil {
					ctx.Runtime.SetEnhancedLighting(enabled)
				}
				w.saveSettings(ctx)
				w.refresh(ctx)
			}),
		),
	).Padding(float32(settingsPad)).Gap(float32(settingsGap))
}

// themeRow is the Theme dropdown at the top of the Gameplay tab: the built-in themes, then the player's
// files in the themes folder next to goro.ini. A choice applies at once (the
// theme sits under ui-layout.ini, so its overrides still win) and is saved in
// goro.ini.
func (w *SettingsWindow) themeRow(ctx client.Context) widget.Widget {
	themes := rotheme.ListThemes()
	items := make([]dropdown.ItemDef, 0, len(themes))
	selected := -1
	active := rotheme.ActiveTheme()
	for i, theme := range themes {
		if theme.ID == active {
			selected = i
		}
		items = append(items, dropdown.ItemDef{Value: theme.ID, Label: theme.Name})
	}
	return primitives.HBox(
		primitives.Expanded(rotheme.Text("Theme")),
		primitives.Box(
			dropdown.New(
				dropdown.ItemDefs(items),
				dropdown.Selected(selected),
				dropdown.MaxVisibleItems(settingsThemeMenuRows),
				dropdown.ItemHeight(float32(settingsThemeItemH)),
				dropdown.PainterOpt(rotheme.DropdownPainter{}),
				dropdown.OnChange(func(_ int, id string) {
					w.selectTheme(ctx, id)
				}),
			),
		).Width(float32(settingsThemeDropdownW)).CrossAlign(primitives.CrossAxisStretch),
	).Gap(float32(settingsGap)).Height(rotheme.Default.Typography.TextSize + rotheme.ButtonPaddingY*2)
}

// selectTheme applies and saves a theme. A theme that fails to load (a user
// file with an error) leaves the current one in place.
func (w *SettingsWindow) selectTheme(ctx client.Context, id string) {
	if id == rotheme.ActiveTheme() {
		return
	}
	warnings, err := rotheme.SetTheme(id)
	for _, warning := range warnings {
		glog.Warnf("ui theme: %s", warning)
	}
	if err != nil {
		glog.Warnf("ui theme %q not applied: %v", id, err)
		w.refresh(ctx)
		return
	}
	if _, err := ctx.Config.SaveUITheme(id); err != nil {
		glog.Warnf("settings save failed: %v", err)
	}
	w.refresh(ctx)
}

func (w *SettingsWindow) soundTree(ctx client.Context) widget.Widget {
	return primitives.Box(
		primitives.HBox(
			rotheme.Text("BGM Vol"),
			primitives.Expanded(
				rotheme.Slider(
					slider.Min(0),
					slider.Max(1),
					slider.Value(float32(settingsVolumeBGM(ctx))),
					slider.OnChange(func(v float32) {
						if ctx.Audio != nil {
							ctx.Audio.SetBGMVolume(float64(v))
						}
						w.saveSettings(ctx)
						w.refresh(ctx)
					}),
				),
			),
		).Gap(float32(settingsGap)),

		primitives.HBox(
			rotheme.Text("SFX Vol"),
			primitives.Expanded(
				rotheme.Slider(
					slider.Min(0),
					slider.Max(1),
					slider.Value(float32(settingsVolumeSFX(ctx))),
					slider.OnChange(func(v float32) {
						if ctx.Audio != nil {
							ctx.Audio.SetSFXVolume(float64(v))
						}
						w.saveSettings(ctx)
						w.refresh(ctx)
					}),
				),
			),
		).Gap(float32(settingsGap)),
	).Padding(float32(settingsPad)).Gap(float32(settingsGap))
}

func (w *SettingsWindow) gameplayTree(ctx client.Context) widget.Widget {
	return primitives.Box(
		w.themeRow(ctx),

		rotheme.Checkbox(
			checkbox.Checked(settingsNoShift(ctx)),
			checkbox.LabelOpt("No Shift"),
			checkbox.OnToggle(func(enabled bool) {
				if ctx.Session != nil {
					ctx.Session.NoShift = enabled
				}
				w.saveSettings(ctx)
				w.refresh(ctx)
			}),
		),

		rotheme.Checkbox(
			checkbox.Checked(settingsNoCtrl(ctx)),
			checkbox.LabelOpt("No Ctrl"),
			checkbox.OnToggle(func(enabled bool) {
				if ctx.Session != nil {
					ctx.Session.NoCtrl = enabled
				}
				w.saveSettings(ctx)
				w.refresh(ctx)
			}),
		),

		rotheme.Checkbox(
			checkbox.Checked(settingsSnapTargets(ctx)),
			checkbox.LabelOpt("Snap to targets"),
			checkbox.OnToggle(func(enabled bool) {
				if ctx.Session != nil {
					ctx.Session.SnapTargets = enabled
				}
				w.saveSettings(ctx)
				w.refresh(ctx)
			}),
		),

		rotheme.Checkbox(
			checkbox.Checked(settingsSnapItems(ctx)),
			checkbox.LabelOpt("Snap to items"),
			checkbox.OnToggle(func(enabled bool) {
				if ctx.Session != nil {
					ctx.Session.SnapItems = enabled
				}
				w.saveSettings(ctx)
				w.refresh(ctx)
			}),
		),

		rotheme.Checkbox(
			checkbox.Checked(ctx.ScriptPath() == "builtin:wasd"),
			checkbox.LabelOpt("Keyboard & joypad controls"),
			checkbox.OnToggle(func(enabled bool) {
				if ctx.Session == nil {
					return
				}
				path := "none"
				if enabled {
					path = "builtin:wasd"
				}
				ctx.Session.ScriptPath = path
				if _, err := ctx.Config.SaveScriptPath(path); err != nil {
					glog.Warnf("settings save failed: %v", err)
				}
				w.refresh(ctx)
			}),
		),
	).
		Padding(float32(settingsPad)).
		Gap(float32(settingsGap))
}

func (w *SettingsWindow) refresh(ctx client.Context) {
	w.EnsureWindow(settingsWindowW, settingsWindowH)
	w.ctx = ctx
	w.SetContent(w.widgetTree(ctx))
	w.Publish(ctx)
}

func (w *SettingsWindow) saveSettings(ctx client.Context) {
	settings := config.UserSettings{
		Fullscreen:       settingsRuntimeFullscreen(ctx),
		VSync:            settingsRuntimeVSync(ctx),
		MaxFPS:           max(ctx.Config.Render.MaxFPS, 0),
		FPS:              settingsRuntimeFPS(ctx),
		Anisotropy:       settingsAnisotropy(ctx),
		TextureUpscaling: settingsTextureUpscaling(ctx),
		SmoothSprites:    settingsSmoothSprites(ctx),
		MSAA:             settingsMSAA(ctx),
		Bloom:            settingsBloom(ctx),
		SSAO:             settingsSSAO(ctx),
		FXAA:             settingsFXAA(ctx),
		EnhancedLighting: settingsEnhancedLighting(ctx),
		BGMVolume:        settingsVolumeBGM(ctx),
		SFXVolume:        settingsVolumeSFX(ctx),
		NoShift:          settingsNoShift(ctx),
		NoCtrl:           settingsNoCtrl(ctx),
		LessEffects:      settingsLessEffects(ctx),
		SnapTargets:      settingsSnapTargets(ctx),
		SnapItems:        settingsSnapItems(ctx),
	}
	path, err := ctx.Config.SaveUserSettings(settings)
	if err != nil {
		glog.Warnf("settings save failed: %v", err)
		return
	}
	glog.Debugf("settings saved path=%s", path)
}

func settingsVolumeBGM(ctx client.Context) float64 {
	if ctx.Audio != nil {
		return ctx.Audio.BGMVolume()
	}
	return ctx.Config.Audio.BGMVolume
}

func settingsVolumeSFX(ctx client.Context) float64 {
	if ctx.Audio != nil {
		return ctx.Audio.SFXVolume()
	}
	return ctx.Config.Audio.SFXVolume
}

func settingsRuntimeFullscreen(ctx client.Context) bool {
	if ctx.Runtime != nil {
		return ctx.Runtime.Fullscreen()
	}
	return ctx.Config.Window.Fullscreen
}

func settingsRuntimeVSync(ctx client.Context) bool {
	if ctx.Runtime != nil {
		return ctx.Runtime.VSync()
	}
	return ctx.Config.Render.VSync
}

func settingsRuntimeFPS(ctx client.Context) bool {
	if ctx.Runtime != nil {
		return ctx.Runtime.FPS()
	}
	return ctx.Config.Render.FPS
}

func settingsAnisotropy(ctx client.Context) int {
	if ctx.Runtime != nil {
		return ctx.Runtime.Anisotropy()
	}
	return ctx.Config.Render.Anisotropy
}

func settingsSmoothSprites(ctx client.Context) bool {
	if ctx.Runtime != nil {
		return ctx.Runtime.SmoothSprites()
	}
	return ctx.Config.Render.SmoothSprites
}

func settingsTextureUpscaling(ctx client.Context) bool {
	if ctx.Runtime != nil {
		return ctx.Runtime.TextureUpscaling()
	}
	return ctx.Config.Render.TextureUpscaling
}

func textureUpscalingLabel(enabled bool) string {
	if enabled {
		return "2x"
	}
	return "Off"
}

func anisotropyLabel(level int) string {
	if level == 0 {
		return "Off"
	}
	return fmt.Sprintf("%dx", level)
}

func settingsNoShift(ctx client.Context) bool {
	if ctx.Session != nil {
		return ctx.Session.NoShift
	}
	return ctx.Config.Gameplay.NoShift
}

func settingsNoCtrl(ctx client.Context) bool {
	if ctx.Session != nil {
		return ctx.Session.NoCtrl
	}
	return ctx.Config.Gameplay.NoCtrl
}

func settingsLessEffects(ctx client.Context) bool {
	if ctx.Session != nil {
		return ctx.Session.LessEffects
	}
	return ctx.Config.Gameplay.LessEffects
}

func settingsSnapTargets(ctx client.Context) bool {
	if ctx.Session != nil {
		return ctx.Session.SnapTargets
	}
	return ctx.Config.Gameplay.SnapTargets
}

func settingsSnapItems(ctx client.Context) bool {
	if ctx.Session != nil {
		return ctx.Session.SnapItems
	}
	return ctx.Config.Gameplay.SnapItems
}

func settingsMSAA(ctx client.Context) bool {
	if ctx.Runtime != nil {
		return ctx.Runtime.MSAA()
	}
	return ctx.Config.Render.MSAA
}

func settingsBloom(ctx client.Context) bool {
	if ctx.Runtime != nil {
		return ctx.Runtime.Bloom()
	}
	return ctx.Config.Render.Bloom
}

func settingsSSAO(ctx client.Context) bool {
	if ctx.Runtime != nil {
		return ctx.Runtime.SSAO()
	}
	return ctx.Config.Render.SSAO
}

func settingsFXAA(ctx client.Context) bool {
	if ctx.Runtime != nil {
		return ctx.Runtime.FXAA()
	}
	return ctx.Config.Render.FXAA
}

func settingsEnhancedLighting(ctx client.Context) bool {
	if ctx.Runtime != nil {
		return ctx.Runtime.EnhancedLighting()
	}
	return ctx.Config.Render.EnhancedLighting
}
