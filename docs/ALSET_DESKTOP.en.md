# Alset Desktop on Tiny Core Linux

**Strategy:** do not build a heavy distro and shrink it later.  
Start **minimal** (Tiny Core) and make **Alset** the primary experience.

## Stack

Alset Shell → Alset-JS (web container) → Desktop Bridge (Go) → AlsetOS Runtime → Tiny Core (kernel, BusyBox, X, FLWM, TCZ).

## Alset-JS decision

Current Alset-JS targets the **browser DOM**, not raw Xorg.  
0.1 uses FLWM + local bridge + shell in a minimal browser/webview + AlsetOS as a separate service.

## Order

1. Tiny Core bootable image + persistence  
2. Alset Shell + bridge  
3. One local organism in the panel  

See `desktop/ACCEPTANCE_0.1.md`.
