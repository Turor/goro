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
  translucent, for example `window_body = #ffffffc0`.
- A file with any malformed line or out-of-range value is rejected as a whole
  with `file:line: reason` in the log, and the previous values stay. Unknown
  keys are only warnings, so a file from an older or newer build still loads.

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
offsets and colors. Shared values live in `[theme]` (base colors), `[text]`,
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
`ui/rotheme` or `game/login` outside the `theme.go`/`theme_*.go` defaults
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
After adding a token, regenerate `ui-layout.default.ini`; a test fails while
it is stale.

Offline previews for theming without a server: `GORO_PREVIEW_MAP=prontera`
with `GORO_PREVIEW_WINDOWS=shop`, or `GORO_PREVIEW_SCREEN=charselect`.
