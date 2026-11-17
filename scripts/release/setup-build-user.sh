#!/usr/bin/env bash
# One-time, as root on the build host (CT119): create the unprivileged user
# that runs the release gate and build. It gets no sudo, no SSH keys, and no
# access to the controller or the signing seed.
set -euo pipefail
USER_NAME="${1:-gorobuild}"
id "$USER_NAME" >/dev/null 2>&1 || useradd --create-home --shell /bin/bash "$USER_NAME"
chmod 700 "/home/$USER_NAME"
passwd -l "$USER_NAME" >/dev/null
[[ -x /usr/local/go/bin/go ]] || { echo "install Go under /usr/local/go first" >&2; exit 1; }
runuser -u "$USER_NAME" -- env -i HOME="/home/$USER_NAME" PATH=/usr/local/go/bin:/usr/bin:/bin go version
