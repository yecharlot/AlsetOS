#!/bin/sh
# bootlocal.d — arranque tardío TinyCore
# Instalar sesión Alset como .xsession del usuario tc (kiosk)
if [ -f /opt/alset/xsession-alset ] && [ -d /home/tc ]; then
  cp /opt/alset/xsession-alset /home/tc/.xsession
  chmod +x /home/tc/.xsession
  chown tc:staff /home/tc/.xsession 2>/dev/null || true
fi
# Si X ya corre, lanzar bridge+kiosk ya
[ -x /opt/alset/boot-alset.sh ] && /opt/alset/boot-alset.sh &
