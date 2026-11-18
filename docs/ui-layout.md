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

What updates live:

- **Every persistent in-game window** (status, inventory, equipment, skills,
  basic info, menu, chat, minimap, quests, social, companions, settings and
  the rest rebound on map entry) is rebuilt after a reload, so padding, row
  heights, labels and button heights follow the file without reopening. The
  NPC dialog, hotbar and inventory rebuild themselves too.
- **Not yet:** the login, server and character select screens, and transient
  pop-ups (shops, trades, prompts) keep their values until they are reopened.

## What the file covers

| Section | Keys |
|---|---|
| `[theme]` | All window, button, text and input colors, context menu hover, HP/SP and hunger bars, success and error text |
| `[text]` | UI text size, font family, optional font file |
| `[button]` | Radius, horizontal and vertical padding, large-button padding, highlight opacity |
| `[window]` | Corner radius, border width, title bar and footer heights, title padding, title button gap, footer padding and gap |
| `[hotbar]` | Slot size, slot gap, row gap, padding |
| `[inventory]` | Cell size, columns, visible rows |
| `[status]` | Row height, padding, gap between the stat columns (the window size follows) |
| `[npc_dialog]` | Dialog width, height, padding, line height, button width; menu width, row height, padding |
| `[progress_bar]` | Height |

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

Still compiled in: the remaining per-window sizes. They move into the file as
each window is tuned.

## For developers

A token is a package variable registered with `rotheme.RegisterInt`,
`RegisterFloat`, `RegisterColor` or `RegisterRGBA` (see `ui/layout_tokens.go`
and `ui/rotheme/theme_tokens.go`). Values computed from tokens are recomputed
in an `rotheme.OnLayoutApplied` hook. Build a window's tree from
`Window.baseHeight()` (not `height`), so the title/footer delta is added once. After adding a token, regenerate
`ui-layout.default.ini`; a test fails while it is stale.
