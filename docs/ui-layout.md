# UI layout and theme file

goro reads optional UI overrides from `ui-layout.ini`, next to `goro.ini`
(or the path in `--ui-layout` / `[ui_layout] file`). Without that file the
built-in layout is used, exactly as before, and nothing else is read.

- Every key, with its default, range and meaning, is in
  [`ui-layout.default.ini`](ui-layout.default.ini). Regenerate it with
  `goro --dump-ui-layout docs/ui-layout.default.ini` (`-` prints to stdout).
- Copy only the lines you want to change. Omitted keys keep the built-in value,
  and deleting a line puts the default back.
- Colors are `#rrggbb` or `#rrggbbaa`. The alpha byte makes a surface
  translucent, for example `window_body = #ffffffc0`. A color can also be a
  palette reference such as `@panel` or `@accent darken(15%)`; see
  [Palette](#palette-color-themes) below.
- A file with any malformed line or out-of-range value is rejected as a whole
  with `file:line: reason` in the log, and the previous values stay. Unknown
  keys are only warnings, so a file from an older or newer build still loads.

## Palette (color themes)

A color theme is set at a high level in `[palette]`, the first section of
the file. It holds the base colors:

| key | used for |
| --- | --- |
| `accent` | window borders, title bars, selection, labels |
| `background` | window background, boards, toasts |
| `panel` | inset panels and light fills |
| `panel_alt` | buttons, alternate rows, secondary fills |
| `text` | body text on windows |
| `text_dim` | secondary, muted and disabled text |
| `border` | button, field and control borders |
| `highlight` | admin names, critical hits, item labels, notices |
| `good` | positive values, success, healing |
| `bad` | errors, damage taken, danger |
| `warning` | warnings, guild markers |
| `hp`, `sp`, `exp` | HP, SP and experience bars |
| `world_text` | text over the game world: names, damage numbers, dark chat panels |
| `shadow` | outlines, shadows, dark translucent backdrops |

Every other color setting defaults to a reference to one of them, written
`@name` and optionally followed by modifiers, applied left to right in HSL
space (the same meaning as in Sass):

- `lighten(N%)`, `darken(N%)`: add or subtract N points of lightness;
- `saturate(N%)`, `desaturate(N%)`: add or subtract N points of saturation;
- `spin(D)`: rotate the hue by D degrees (-360 to 360);
- `alpha(A)`: set the opacity, 0 to 1 (or 0% to 100%).

For example `button_down = @panel_alt spin(-0.6) desaturate(2.6%) darken(8.8%)`
or `panel_background = @shadow ... alpha(0.737)`. The defaults are computed so
that with the default palette every reference gives exactly the built-in
color: nothing changes until the palette does. `docs/ui-layout.default.ini`
shows each reference with the color it gives, for example
`# window background ... (= #ffffff)`.

Changing a palette color re-colors every setting that refers to it, live with
`--ui-dev` (the whole file is resolved again on every reload). To change one
setting only, give it an explicit `#rrggbb[aa]`, which always wins, or its own
reference (`window_title = @panel_alt lighten(4%)`). Palette entries
themselves must be explicit colors. The F11 tuner shows and saves references
as written; after a palette change, Save writes only the palette lines.

[`ui-palette-dark.example.ini`](ui-palette-dark.example.ini) is a complete
dark theme: a palette plus a few overrides for colors the light theme derives
as pale fills (title bars, hovers, highlights). It is the same as the built-in
Dark theme below.

## Themes (Settings > Gameplay > Theme)

A theme is a palette file that sits under `ui-layout.ini`. Pick one from the
Theme dropdown at the top of the settings window's Gameplay tab: it applies at
once, and the choice is saved in `goro.ini`:

```ini
[ui_layout]
theme = dark
```

The order, lowest first, is: compiled default, then the theme, then
`ui-layout.ini` (and the F11 tuner). So anything set in `ui-layout.ini` still
wins over the theme key by key, and the tuner saves only your own changes,
never the theme's values.

Built in (embedded in the executable, in `ui/rotheme/themes/`):

| Theme | Look |
|---|---|
| Default | The light blue Ragnarok look; sets nothing. |
| Dark | Charcoal panels with a gold accent. |
| Complementary | Blue base hue and its orange complement on dark blue panels. |
| Analogous | Neighbouring red, orange, yellow and yellow-green hues on light warm panels. |
| Triadic | Violet, rose and green, 120 degrees apart, on dark violet panels. |

Each color-theory theme builds its accent, highlight, good, bad, warning and
panel colors from its scheme's hues. Every built-in theme keeps body text at
WCAG AA contrast (at least 4.5:1) on the background and panel colors, and the
window text, title and label colors on what they are drawn on;
`TestBuiltinThemeTextContrast` checks it.

Your own themes: put `<name>.ini` in a `themes` folder next to `goro.ini`
(names use `a-z`, `0-9`, `_` and `-`). It is an ordinary layout file, usually
just a `[palette]` section and a few `[theme]` lines; keys it leaves out keep
the compiled defaults. A `# name: My Theme` comment at the top sets the label
shown in the dropdown. A file with a built-in theme's name is ignored, and a
file with an error is not applied (the current theme stays).

## Live editing (--ui-dev)

Start goro with `--ui-dev` (or `[ui_layout] watch = true` in goro.ini). The
client then checks the file's size and modification time every 0.5 s, so a
saved edit shows within about a second; F10 reloads at once. A banner at the
top of the screen says `UI layout reloaded`, or `UI layout NOT applied:
file:line: reason` for a bad edit, which keeps the last good values. Deleting
the file restores the built-in layout. Without `--ui-dev` nothing is polled and
F10 does nothing.

What updates live: **everything.** After a reload every window and screen
that is on screen rebuilds once, in its place in the window stack, with the
new sizes and colors:

- persistent in-game windows (status, inventory, equipment, skills, basic
  info, menu, chat, minimap, quests, social, companions, settings, ...);
- the login, server select and character select/create screens;
- transient pop-ups (NPC shops, trades, vending, storage, mail, prompts,
  confirm dialogs, context menus);
- paint-time HUD elements (hotbar, status icons, notifications, minimap
  markers) read their values on every draw.
- the map loading screen progress bar (`[loading]`);
- the game world overlays, read on every draw: HP/SP/hunger/cast bars
  (`[actor_bars]`), damage, heal, miss and critical numbers
  (`[damage_numbers]`), names under actors (`[actor_names]`), chat room and
  vending boards (`[boards]`, redrawn once per reload), speech bubbles
  (`[speech_bubble]`), ground item names and the fallback item marker
  (`[item_labels]`), and the Blind/Darkness screen tint (`[blind]`).

## In-game tuner (F11)

With `--ui-dev`, or `[ui_layout] watch = true` in goro.ini (the same switch
that turns on the file watcher), F11 opens the UI Tuner. The header row picks
a section with `<` and `>` (`all` lists every key, shown as `section.key`) and
shows the section's description; the search box filters keys and their
descriptions by text (Enter or **Find**). **Prev** and **Next** page through
the filtered keys. Numbers change with `-` and `+` (by 1, or 0.05 for 0 to 1
values such as alpha) and apply at once; colors and names are typed, then
**Apply** (in key order; the first invalid value stops and is reported). A
changed key is shown in the title color. **Save** writes the changed values
(only those) to ui-layout.ini through a temporary file and a rename; this
replaces the file, so comments in it are not kept. **Revert** reloads the
file. Players never see the tuner: without the watch switch F11 does nothing.

## What the file covers

Every window and screen has its own section, named after it (`[shop]`,
`[trade]`, `[char_select]`, `[login]`, `[guild]`, ...), holding its width and
height, paddings, gaps, row and header heights, column widths, visible rows,
offsets and colors. The base colors are in `[palette]`; shared values live in
`[theme]` (shared window colors), `[text]`,
`[button]`, `[window]` (frame), `[progress_bar]` and the widget sections. The
full list, with every default, range and description, is
[`ui-layout.default.ini`](ui-layout.default.ini); in the tuner, pick the
section or search for a word.

Window sizes in the code are written against the built-in 28 px title bar
and 42 px footer. Changing `title_height` or `footer_height` grows or shrinks
every window by the difference, so each window keeps its content area.

Fonts: `font_family` names the font the UI draws with. The built-in font is
`Inter`; to use another, set `font_file` to a TTF or OTF (relative to
ui-layout.ini, up to 16 MiB) and `font_family` to the name to register it
under. A family that is not loaded falls back to Inter, and a font file that
cannot be read is reported in the banner and log.

The base gogpu/ui theme (text fields, scrollbars, focus rings) is rebuilt
from these values on every reload too.

## For developers

No UI size or color is hardcoded. Each window keeps its compiled defaults in a
`theme_<section>.go` file next to it (`ui/theme_shop.go`,
`game/login/theme_login.go`, ...): a `var` block with the values and an
`init()` that registers each one with `rotheme.Section(name, doc)` and
`.Int`, `.Float`, `.Color`, `.RGBA` or `.String`. Window code only reads those
variables. `ui/hardcoded_literals_test.go` fails on any color literal, or any
size, offset, padding or row height written as a number, in `ui`,
`ui/rotheme`, `game/login` or the map loading screen (`game/map/loading.go`,
whose defaults are in `ui/rotheme/theme_loading.go`) outside the `theme.go`/`theme_*.go` defaults
tables, and on a defaults-table value that is not registered. A number that is
not visual (packet flag, skill ID, text limit, timer, capacity) stays inline
with `// theme:ok <reason>`.

A window that copies values when it builds its tree registers a rebuild where
it opens: `w.OnLayoutReload(func() { w.SetSize(...); w.SetContent(tree()) })`.
`Manager.RebuildStaleWindows` runs it once per reload for every window still
on screen (persistent world windows are also rebound by the world mode).
Values computed from tokens are computed where used, or in an
`rotheme.OnLayoutApplied` hook. Build a window's tree from
`Window.baseHeight()` (not `height`), so the title/footer delta is added once.
After adding a token, regenerate `ui-layout.default.ini`; a test in package
main (which links every package that registers tokens) fails while it is
stale.

Every color token needs a default palette reference in
`ui/rotheme/palette_refs.go`. Add `"section.key": "@base",` with the palette
color it belongs to and run
`GORO_UPDATE_PALETTE_REFS=1 go test -run TestPaletteDefaults .`, which writes
the exact modifiers (a new key missing from the file gets the nearest palette
color). `TestPaletteDefaultsReproduceBuiltInColors` fails while any color has
no reference or its reference does not give exactly the built-in color, and
`TestPaletteChangeReachesEveryReference` checks that changing each palette
color reaches every setting that refers to it and nothing else.

The literal guard also covers the world overlays: `game/battle`,
`game/actors`, `game/items.go` (defaults in `game/theme_items.go`) and the
Blind overlay in `game/effects/blind.go` (`theme_blind.go`). Their
non-visual numbers (timers, motion curves taken from roBrowser, sprite anchors
from the original client, drop physics, packet codes) carry `// theme:ok`.
The skill effects in the rest of `game/effects` reproduce the original
client's effect data and are not themed.

Offline previews for theming without a server: `GORO_PREVIEW_MAP=prontera`
with `GORO_PREVIEW_WINDOWS=shop`, or `GORO_PREVIEW_SCREEN=charselect`.
