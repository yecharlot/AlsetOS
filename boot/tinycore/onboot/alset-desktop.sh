#!/bin/sh
# Example Tiny Core onboot hook — run after X is up.
# Install under /opt/bootlocal.sh or tce onboot as appropriate for your image.

set -eu
BRIDGE_BIN="${ALSET_BRIDGE:-/usr/local/bin/alset-desktop-bridge}"
SHELL_URL="${ALSET_SHELL_URL:-http://127.0.0.1:7420/}"
BROWSER="${ALSET_BROWSER:-}"

log() { echo "[alset-desktop] $*" >&2; }

if [ ! -x "$BRIDGE_BIN" ]; then
  log "bridge missing: $BRIDGE_BIN — recovery: use FLWM/terminal"
  exit 0
fi

# Start bridge (serves shell + IPC)
if ! pgrep -f alset-desktop-bridge >/dev/null 2>&1; then
  log "starting bridge"
  "$BRIDGE_BIN" -addr 127.0.0.1:7420 -data "${HOME:-/tmp}/.alset-desktop" &
  sleep 1
fi

# Open shell UI if a browser helper exists
if [ -n "$BROWSER" ] && command -v "$BROWSER" >/dev/null 2>&1; then
  log "opening shell with $BROWSER"
  "$BROWSER" "$SHELL_URL" &
elif command -v chromium-browser >/dev/null 2>&1; then
  chromium-browser --app="$SHELL_URL" &
elif command -v firefox >/dev/null 2>&1; then
  firefox "$SHELL_URL" &
else
  log "no browser found — open $SHELL_URL manually; FLWM remains available"
fi
