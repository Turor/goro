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

## What the file covers

| Section | Keys |
|---|---|
| `[theme]` | All window, button, text and input colors, context menu hover, HP/SP and hunger bars, success and error text |
| `[text]` | UI text size |
| `[button]` | Radius, horizontal and vertical padding, large-button padding, highlight opacity |
| `[window]` | Corner radius, border width, title padding, title button gap, footer padding and gap |
| `[hotbar]` | Slot size, slot gap, row gap, padding |
| `[inventory]` | Cell size, columns, visible rows |
| `[progress_bar]` | Height |

Still compiled in: window title and footer heights (about 90 window sizes are
constant sums of them), the font family, and the remaining per-window sizes.
They move into the file as each window is tuned.

## For developers

A token is a package variable registered with `rotheme.RegisterInt`,
`RegisterFloat`, `RegisterColor` or `RegisterRGBA` (see `ui/layout_tokens.go`
and `ui/rotheme/theme_tokens.go`). Values computed from tokens are recomputed
in an `rotheme.OnLayoutApplied` hook. After adding a token, regenerate
`ui-layout.default.ini`; a test fails while it is stale.
