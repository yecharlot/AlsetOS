# Tiny Core integration notes

## Recommended path

1. Start from **TinyCore** ISO (has X + FLWM) for fast graphics checks.  
2. Later: custom image from **Core** + only needed TCZ.  

## Persistence

Use Tiny Core boot codes, e.g.:

- `tce=UUID=...` or `tce=/mnt/sda1/tce`
- home / opt backup as documented upstream  

Do **not** rely only on the RAM tmpfs for Alset data.

## On-boot (conceptual)

```text
load TCZ (Xorg, browser or webview helper, ...)
start X + FLWM   # rescue WM always available
start alset-desktop-bridge
open Alset Shell (http://127.0.0.1:7420/) in minimal browser
optional: alsetos service
```

If Shell fails, user still has FLWM + terminal.

## Files in this tree

| Path | Role |
|------|------|
| `onboot/alset-desktop.sh` | Example onboot script |
| `tce/onboot.list.example` | Extension names to resolve on a real TC box |
| `recovery/README.md` | Rescue procedure |
