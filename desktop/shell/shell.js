(function () {
  const statusBox = document.getElementById('status-box');
  const launcher = document.getElementById('launcher');
  const taskOrg = document.getElementById('task-org');
  const termOut = document.getElementById('term-out');
  const termIn = document.getElementById('term-in');
  const mem = Object.create(null);

  function tick() {
    document.getElementById('clock').textContent = new Date().toLocaleString();
  }
  setInterval(tick, 1000);
  tick();

  document.getElementById('btn-menu').onclick = function () {
    launcher.classList.toggle('hidden');
  };

  document.querySelectorAll('[data-close]').forEach(function (btn) {
    btn.addEventListener('click', function () {
      const id = btn.getAttribute('data-close');
      const el = document.getElementById(id);
      if (el) el.classList.add('hidden');
    });
  });

  async function api(path, opts) {
    const r = await fetch(path, opts);
    if (!r.ok) throw new Error(await r.text());
    return r.json();
  }

  function termLog(line) {
    termOut.textContent += line + '\n';
    termOut.scrollTop = termOut.scrollHeight;
  }

  async function refreshStatus() {
    try {
      const st = await api('/v1/status');
      statusBox.textContent = JSON.stringify(st, null, 2);
      taskOrg.textContent = 'organismo: ' + (st.organism || '—') + (st.rootcid ? ' · ' + st.rootcid : '');
      return st;
    } catch (e) {
      statusBox.textContent = 'Bridge no disponible: ' + e.message;
      throw e;
    }
  }

  function openTools(which) {
    // Prefer same-origin paths served by bridge -web
    const paths = {
      studio: '/tools/studio/',
      editor: '/tools/alset-editor/',
    };
    const url = paths[which];
    if (!url) return;
    window.open(url, '_blank', 'noopener');
  }

  async function runOrganism() {
    statusBox.textContent = 'Ejecutando organismo…';
    const res = await api('/v1/organism/run', { method: 'POST' });
    statusBox.textContent = JSON.stringify(res, null, 2);
    taskOrg.textContent = 'organismo: ' + (res.organism || '—') + (res.rootcid ? ' · ' + res.rootcid : '');
    return res;
  }

  function openTerminal() {
    const w = document.getElementById('win-terminal');
    w.classList.remove('hidden');
    termIn.focus();
  }

  async function handleApp(app) {
    launcher.classList.add('hidden');
    if (app === 'status') return refreshStatus();
    if (app === 'organism') {
      try { await runOrganism(); } catch (e) { statusBox.textContent = String(e.message || e); }
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
    if (app === 'terminal') return openTerminal();
    if (app === 'studio' || app === 'editor') {
      openTools(app);
      statusBox.textContent =
        'Abriendo ' + app + '…\nSi no carga, arranca el bridge con -web apuntando al runtime:\n' +
        '  alset-desktop-bridge -web /ruta/Alset-LISPAI-Runtime/web';
      return;
    }
  }

  document.querySelectorAll('[data-app]').forEach(function (btn) {
    btn.addEventListener('click', function () {
      handleApp(btn.getAttribute('data-app'));
    });
  });

  // ——— Terminal LispAI / state ———
  function parseSimpleLisp(line) {
    const s = line.trim();
    const m = /^\(\s*(\S+)(?:\s+([\s\S]*))?\)\s*$/.exec(s);
    if (!m) return null;
    const op = m[1];
    const rest = (m[2] || '').trim();
    const parts = [];
    let cur = '';
    let inQ = false;
    for (let i = 0; i < rest.length; i++) {
      const c = rest[i];
      if (c === '"') { inQ = !inQ; cur += c; continue; }
      if (!inQ && /\s/.test(c)) {
        if (cur) parts.push(cur.replace(/^"|"$/g, ''));
        cur = '';
        continue;
      }
      cur += c;
    }
    if (cur) parts.push(cur.replace(/^"|"$/g, ''));
    return { op, parts };
  }

  async function execTerm(line) {
    const raw = line.trim();
    if (!raw) return;
    termLog('› ' + raw);
    const low = raw.toLowerCase();
    if (low === 'help' || raw === '(help)') {
      termLog('help | status | organism | clear | studio | editor');
      termLog('(recordar clave valor)  (leer clave)');
      termLog('(set-state clave valor)  (get-state clave)');
      termLog('mem — lista memoria local del terminal');
      return;
    }
    if (low === 'clear') { termOut.textContent = ''; return; }
    if (low === 'studio' || low === 'editor') return handleApp(low);
    if (low === 'status') {
      try {
        const st = await refreshStatus();
        termLog(JSON.stringify(st, null, 2));
      } catch (e) { termLog('error: ' + e.message); }
      return;
    }
    if (low === 'organism') {
      try {
        const res = await runOrganism();
        termLog(JSON.stringify(res, null, 2));
      } catch (e) { termLog('error: ' + e.message); }
      return;
    }
    if (low === 'mem') {
      termLog(JSON.stringify(mem, null, 2));
      return;
    }
    const lisp = parseSimpleLisp(raw);
    if (lisp) {
      const op = lisp.op;
      if (op === 'recordar' || op === 'set-state') {
        const k = lisp.parts[0];
        const v = lisp.parts.slice(1).join(' ') || '';
        mem[k] = v;
        termLog('ok ' + k + ' = ' + v);
        return;
      }
      if (op === 'leer' || op === 'get-state') {
        const k = lisp.parts[0];
        termLog(k + ' => ' + (mem[k] !== undefined ? mem[k] : 'nil'));
        return;
      }
      if (op === 'status' || op === 'organism') return execTerm(op);
      termLog('forma no reconocida: ' + op);
      return;
    }
    termLog('desconocido. escribe help');
  }

  document.getElementById('term-form').addEventListener('submit', function (e) {
    e.preventDefault();
    const v = termIn.value;
    termIn.value = '';
    execTerm(v);
  });

  refreshStatus().catch(function () {});
  setInterval(function () { refreshStatus().catch(function () {}); }, 20000);
})();
