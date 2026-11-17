# Credits

TuroranRO's goro client is a fork. Everything below is third-party work we
build on; thank you to every author.

## goro

- Upstream: <https://github.com/kivutar/goro>
- Author: Jean-André Santoni and goro contributors
- License: MIT (see `LICENSE`, kept unchanged)

goro's own README credits ROBrowser Legacy and Open Midgard for their reverse
engineering work, which goro builds on.

## Libraries

Go module dependencies and their licenses are listed in `go.mod`; notable
ones include GoGPU (`gogpu/gogpu`, `gogpu/wgpu`, `gogpu/ui`, `gogpu/gg`,
`gogpu/naga`, via kivutar's forks), gopher-lua, ebitengine oto and purego,
charmbracelet log, godexture, and xbrscaler.

## Turoran additions

Turoran-specific code (for example `updater/`, `cmd/goro-sign-release/`,
`scripts/release/`) is © Turoran contributors under the same MIT license. The
auto-update design follows our Korangar fork (`Turor/korangar`, itself a fork
of <https://github.com/vE5li/korangar>, MIT).
