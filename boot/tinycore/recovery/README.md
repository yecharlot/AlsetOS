# Recovery

1. Boot Tiny Core normally (or `tinycore text` if needed).  
2. Start X + FLWM if not already.  
3. Terminal: check `ps` for `alset-desktop-bridge`.  
4. Data dir default: `~/.alset-desktop`  
5. Re-run `/opt/alset/alset-desktop.sh` or the bridge binary manually.  

Alset Shell must **never** be the only way to reach a terminal.
