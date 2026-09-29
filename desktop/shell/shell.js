(function () {
  const statusBox = document.getElementById('status-box');
  const launcher = document.getElementById('launcher');
  const taskOrg = document.getElementById('task-org');

  function tick() {
    const d = new Date();
    document.getElementById('clock').textContent = d.toLocaleString();
  }
  setInterval(tick, 1000);
  tick();

  document.getElementById('btn-menu').onclick = function () {
    launcher.classList.toggle('hidden');
  };

  async function api(path, opts) {
    const r = await fetch(path, opts);
    if (!r.ok) throw new Error(await r.text());
    return r.json();
  }

  async function refreshStatus() {
    try {
      const st = await api('/v1/status');
      statusBox.textContent = JSON.stringify(st, null, 2);
      taskOrg.textContent = 'organismo: ' + (st.organism || '—') + (st.rootcid ? ' · ' + st.rootcid : '');
    } catch (e) {
      statusBox.textContent = 'Bridge no disponible: ' + e.message;
    }
  }

  document.querySelectorAll('#launcher [data-app]').forEach(function (btn) {
    btn.addEventListener('click', async function () {
      launcher.classList.add('hidden');
      const app = btn.getAttribute('data-app');
      if (app === 'status') return refreshStatus();
      if (app === 'organism') {
        statusBox.textContent = 'Ejecutando organismo…';
        try {
          const res = await api('/v1/organism/run', { method: 'POST' });
          statusBox.textContent = JSON.stringify(res, null, 2);
          taskOrg.textContent = 'organismo: ' + (res.organism || '—');
          if (res.rootcid) taskOrg.textContent += ' · ' + res.rootcid;
        } catch (e) {
          statusBox.textContent = String(e.message || e);
        }
        return;
      }
      if (app === 'files') {
        try {
          const res = await api('/v1/fs/list');
          statusBox.textContent = JSON.stringify(res, null, 2);
        } catch (e) {
          statusBox.textContent = String(e.message || e);
        }
        return;
      }
      if (app === 'terminal') {
        statusBox.textContent =
          'Recuperación:\n- FLWM / terminal del sistema siempre disponibles\n' +
          '- Bridge: alset-desktop-bridge\n- Datos: ~/.alset-desktop\n' +
          '- Si la shell falla, el escritorio TC no debe quedar bloqueado.';
      }
    });
  });

  refreshStatus();
  setInterval(refreshStatus, 15000);
})();
