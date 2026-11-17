// Package updater is the TuroranRO signed auto-updater for goro.
//
// On launch a release build fetches the main-channel catalog, verifies an
// Ed25519-signed manifest against the embedded Turoran release public key,
// downloads the binary for the running platform, checks its size and SHA-256,
// and swaps it in place next to the running executable. The new binary must
// pass a self-check before the swap is kept; any failure restores the previous
// executable. Development builds (no Channel set at link time) never update.
//
// # Manifest (GORO1)
//
// The signature covers this UTF-8 payload only. Files are sorted by path
// bytes, and there is no insignificant whitespace:
//
//	GORO1
//	<version>
//	<build>
//	<file count>
//	for each file:
//	<relative path>
//	<sha256 lowercase hex>
//	<size decimal>
//	<url>
//
// The on-disk envelope prepends "GORO1SIG\n" and a 128-character lowercase hex
// signature line. The layout mirrors Korangar's KRUP1 so the same controller
// seed can sign both, while the different magic stops a Korangar manifest
// from ever being accepted by goro (and the other way round).
//
// See docs/turoran-auto-update.md.
package updater
