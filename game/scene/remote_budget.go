package scene

import "time"

// FrameRemoteBudget is how long a per-frame asset lookup (icons, board
// textures) may wait for a streamed file before drawing without it; the
// download finishes in the background and a later frame picks it up.
const FrameRemoteBudget = 15 * time.Millisecond
