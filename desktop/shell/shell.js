/**
 * Alset Desktop Window Manager
 * Icons movable · windows move/resize/close · apps in-desktop (iframe)
 * Mind · Zyrion · Neural · Syllogisms as system services
 */
(function () {
  const ICON_KEY = 'alset-desktop-icons-v2';
  const iconLayer = document.getElementById('icon-layer');
  const windowLayer = document.getElementById('window-layer');
  const taskButtons = document.getElementById('task-buttons');
  const startMenu = document.getElementById('start-menu');
  const trayMind = document.getElementById('tray-mind');
  const trayClock = document.getElementById('tray-clock');

  let zTop = 30;
  const windows = new Map(); // id -> { el, taskBtn, title }
  const mem = Object.create(null);
  const neural = Object.create(null); // key -> weight 0..1
  const facts = []; // syllogism triples {s,r,o,c}

  const DEFAULT_ICONS = [
    { id: 'studio', label: 'Studio', glyph: '◈', x: 24, y: 24 },
    { id: 'editor', label: 'Editor JS', glyph: '◇', x: 24, y: 120 },
    { id: 'terminal', label: 'Terminal', glyph: '▣', x: 24, y: 216 },
    { id: 'mind', label: 'Mind', glyph: '◎', x: 24, y: 312 },
    { id: 'organism', label: 'Organismo', glyph: '⚡', x: 24, y: 408 },
    { id: 'status', label: 'Sistema', glyph: '▤', x: 120, y: 24 },
  ];

  function loadIcons() {
    try {
      const raw = JSON.parse(localStorage.getItem(ICON_KEY) || 'null');
      if (Array.isArray(raw) && raw.length) return raw;
    } catch (_) {}
    return DEFAULT_ICONS.map((i) => ({ ...i }));
  }

  function saveIcons(list) {
    localStorage.setItem(ICON_KEY, JSON.stringify(list));
  }

  let iconState = loadIcons();

  function renderIcons() {
    iconLayer.innerHTML = '';
    iconState.forEach((ic) => {
      const btn = document.createElement('button');
      btn.type = 'button';
      btn.className = 'desk-icon';
      btn.style.left = ic.x + 'px';
      btn.style.top = ic.y + 'px';
      btn.dataset.id = ic.id;
      btn.innerHTML = `<span class="glyph">${ic.glyph}</span><span class="label">${ic.label}</span>`;
      enableIconDrag(btn, ic);
      btn.addEventListener('dblclick', () => launch(ic.id));
      btn.addEventListener('click', (e) => {
        if (btn._dragged) { btn._dragged = false; return; }
        document.querySelectorAll('.desk-icon').forEach((x) => x.classList.remove('selected'));
        btn.classList.add('selected');
      });
      iconLayer.appendChild(btn);
    });
  }

  function enableIconDrag(el, ic) {
    let sx, sy, ox, oy, moved;
    el.addEventListener('pointerdown', (e) => {
      if (e.button !== 0) return;
      el.setPointerCapture(e.pointerId);
      sx = e.clientX; sy = e.clientY;
      ox = ic.x; oy = ic.y; moved = false;
      el.classList.add('dragging');
      const onMove = (ev) => {
        const dx = ev.clientX - sx;
        const dy = ev.clientY - sy;
        if (Math.abs(dx) + Math.abs(dy) > 4) moved = true;
        const desk = document.getElementById('desktop').getBoundingClientRect();
        ic.x = Math.max(0, Math.min(desk.width - 90, ox + dx));
        ic.y = Math.max(0, Math.min(desk.height - 90, oy + dy));
        el.style.left = ic.x + 'px';
        el.style.top = ic.y + 'px';
      };
      const onUp = () => {
        el.classList.remove('dragging');
        el.releasePointerCapture?.(e.pointerId);
        el.removeEventListener('pointermove', onMove);
        el.removeEventListener('pointerup', onUp);
        if (moved) {
          el._dragged = true;
          saveIcons(iconState);
        }
      };
      el.addEventListener('pointermove', onMove);
      el.addEventListener('pointerup', onUp);
    });
  }

  // ——— Window manager ———
  function focusWin(id) {
    const w = windows.get(id);
    if (!w) return;
    zTop += 1;
    w.el.style.zIndex = String(zTop);
    w.el.classList.add('active');
    windows.forEach((other, oid) => {
      if (oid !== id) other.el.classList.remove('active');
      if (other.taskBtn) other.taskBtn.classList.toggle('active', oid === id);
    });
  }

  function closeWin(id) {
    const w = windows.get(id);
    if (!w) return;
    w.el.remove();
    w.taskBtn?.remove();
    windows.delete(id);
  }

  function toggleMax(id) {
    const w = windows.get(id);
    if (!w) return;
    w.el.classList.toggle('maximized');
  }

  function createWindow({ id, title, width, height, x, y, contentHTML, iframeSrc }) {
    if (windows.has(id)) {
      focusWin(id);
      const existing = windows.get(id);
      existing.el.classList.remove('hidden-min');
      existing.el.style.display = '';
      return existing;
    }
    const desk = document.getElementById('desktop').getBoundingClientRect();
    const el = document.createElement('div');
    el.className = 'win active';
    el.dataset.win = id;
    el.style.width = (width || 720) + 'px';
    el.style.height = (height || 480) + 'px';
    el.style.left = (x != null ? x : Math.max(40, (desk.width - (width || 720)) / 2)) + 'px';
    el.style.top = (y != null ? y : Math.max(30, (desk.height - (height || 480)) / 3)) + 'px';

    el.innerHTML = `
      <div class="win-title" data-drag>
        <span class="title-text">${title}</span>
        <div class="win-controls">
          <button type="button" data-act="min" title="Minimizar">–</button>
          <button type="button" data-act="max" title="Maximizar">□</button>
          <button type="button" class="close" data-act="close" title="Cerrar">×</button>
        </div>
      </div>
      <div class="win-body"></div>
      <div class="win-resize" data-resize></div>
    `;
    const body = el.querySelector('.win-body');
    if (iframeSrc) {
      const iframe = document.createElement('iframe');
      iframe.src = iframeSrc;
      iframe.title = title;
      iframe.setAttribute('allow', 'clipboard-read; clipboard-write');
      body.appendChild(iframe);
    } else {
      const inner = document.createElement('div');
      inner.className = 'inner';
      inner.innerHTML = contentHTML || '';
      body.appendChild(inner);
    }

    const taskBtn = document.createElement('button');
    taskBtn.type = 'button';
    taskBtn.className = 'task-btn active';
    taskBtn.textContent = title;
    taskBtn.onclick = () => {
      if (el.style.display === 'none') {
        el.style.display = '';
        focusWin(id);
      } else if (el.classList.contains('active')) {
        el.style.display = 'none';
      } else {
        focusWin(id);
      }
    };
    taskButtons.appendChild(taskBtn);

    el.querySelector('[data-act="close"]').onclick = (e) => { e.stopPropagation(); closeWin(id); };
    el.querySelector('[data-act="max"]').onclick = (e) => { e.stopPropagation(); toggleMax(id); };
    el.querySelector('[data-act="min"]').onclick = (e) => {
      e.stopPropagation();
      el.style.display = 'none';
    };
    el.addEventListener('mousedown', () => focusWin(id));

    enableWinDrag(el);
    enableWinResize(el);
    windowLayer.appendChild(el);
    windows.set(id, { el, taskBtn, title });
    focusWin(id);
    return windows.get(id);
  }

  function enableWinDrag(el) {
    const bar = el.querySelector('[data-drag]');
    bar.addEventListener('pointerdown', (e) => {
      if (e.target.closest('.win-controls')) return;
      if (el.classList.contains('maximized')) return;
      e.preventDefault();
      const rect = el.getBoundingClientRect();
      const desk = document.getElementById('desktop').getBoundingClientRect();
      const ox = e.clientX - rect.left;
      const oy = e.clientY - rect.top;
      bar.setPointerCapture(e.pointerId);
      const onMove = (ev) => {
        let left = ev.clientX - desk.left - ox;
        let top = ev.clientY - desk.top - oy;
        left = Math.max(-rect.width + 80, Math.min(desk.width - 40, left));
        top = Math.max(0, Math.min(desk.height - 40, top));
        el.style.left = left + 'px';
        el.style.top = top + 'px';
      };
      const onUp = () => {
        bar.releasePointerCapture?.(e.pointerId);
        bar.removeEventListener('pointermove', onMove);
        bar.removeEventListener('pointerup', onUp);
      };
      bar.addEventListener('pointermove', onMove);
      bar.addEventListener('pointerup', onUp);
    });
  }

  function enableWinResize(el) {
    const handle = el.querySelector('[data-resize]');
    handle.addEventListener('pointerdown', (e) => {
      if (el.classList.contains('maximized')) return;
      e.preventDefault();
      e.stopPropagation();
      const startX = e.clientX, startY = e.clientY;
      const startW = el.offsetWidth, startH = el.offsetHeight;
      handle.setPointerCapture(e.pointerId);
      const onMove = (ev) => {
        el.style.width = Math.max(320, startW + (ev.clientX - startX)) + 'px';
        el.style.height = Math.max(200, startH + (ev.clientY - startY)) + 'px';
      };
      const onUp = () => {
        handle.releasePointerCapture?.(e.pointerId);
        handle.removeEventListener('pointermove', onMove);
        handle.removeEventListener('pointerup', onUp);
      };
      handle.addEventListener('pointermove', onMove);
      handle.addEventListener('pointerup', onUp);
    });
  }

  // ——— API helpers ———
  async function api(path, opts) {
    const r = await fetch(path, opts);
    const t = await r.text();
    try { return JSON.parse(t); } catch { return { raw: t, ok: r.ok }; }
  }

  // ——— Apps ———
  function launch(id) {
    startMenu.classList.add('hidden');
    switch (id) {
      case 'studio':
        createWindow({
          id: 'app-studio',
          title: 'Alset Studio',
          width: Math.min(1100, window.innerWidth - 40),
          height: Math.min(700, window.innerHeight - 80),
          iframeSrc: '/tools/',
        });
        break;
      case 'editor':
        createWindow({
          id: 'app-editor',
          title: 'Alset-JS Editor',
          width: Math.min(1100, window.innerWidth - 40),
          height: Math.min(700, window.innerHeight - 80),
          iframeSrc: '/tools/alset-editor/',
        });
        break;
      case 'terminal':
        openTerminal();
        break;
      case 'mind':
        openMind();
        break;
      case 'organism':
        openOrganism();
        break;
      case 'status':
        openStatus();
        break;
      case 'files':
        openFiles();
        break;
      case 'about':
        createWindow({
          id: 'app-about',
          title: 'Acerca de Alset OS',
          width: 420,
          height: 280,
          contentHTML: `<p><strong>Alset Desktop</strong> — gestor de ventanas nativo del ecosistema Alset.</p>
            <p class="status-pre">Mind · Zyrion · Neural · Silogismos disponibles como servicios del sistema.
Studio y Editor se ejecutan dentro del escritorio (no en pestañas externas).</p>
            <p class="status-pre">Iconos arrastrables · ventanas movibles · persistencia local de layout de iconos.</p>`,
        });
        break;
      default:
        break;
    }
  }

  function openTerminal() {
    const w = createWindow({
      id: 'app-terminal',
      title: 'Terminal · LispAI / alsetState',
      width: 640,
      height: 420,
      contentHTML: `<pre class="term-out" id="term-out">Alset Terminal
help · status · organism · (recordar k v) · (leer k) · (set-state k v) · (get-state k)
(mind texto) · (zyrion a b) · (assert s r o) · (infer) · (neural k v)
</pre>
<form class="term-form" id="term-form"><span class="prompt">›</span><input id="term-in" autocomplete="off" /></form>`,
    });
    const form = w.el.querySelector('#term-form');
    const input = w.el.querySelector('#term-in');
    const out = w.el.querySelector('#term-out');
    if (form && !form._bound) {
      form._bound = true;
      form.addEventListener('submit', async (e) => {
        e.preventDefault();
        const line = input.value;
        input.value = '';
        out.textContent += '› ' + line + '\n';
        const res = await execTerm(line);
        if (res != null) out.textContent += res + '\n';
        out.scrollTop = out.scrollHeight;
      });
    }
    setTimeout(() => input?.focus(), 50);
  }

  function parseLisp(line) {
    const s = line.trim();
    const m = /^\(\s*(\S+)(?:\s+([\s\S]*))?\)\s*$/.exec(s);
    if (!m) return null;
    const op = m[1];
    const rest = (m[2] || '').trim();
    const parts = [];
    let cur = '', q = false;
    for (let i = 0; i < rest.length; i++) {
      const c = rest[i];
      if (c === '"') { q = !q; continue; }
      if (!q && /\s/.test(c)) { if (cur) parts.push(cur); cur = ''; continue; }
      cur += c;
    }
    if (cur) parts.push(cur);
    return { op, parts };
  }

  async function execTerm(line) {
    const raw = (line || '').trim();
    if (!raw) return null;
    const low = raw.toLowerCase();
    if (low === 'help') {
      return 'status organism clear mem studio editor mind\n(recordar k v) (leer k) (set-state k v) (get-state k)\n(mind texto…) (zyrion a b) (assert s r o [c]) (infer) (ask s r)\n(neural k peso) (neural-get k)';
    }
    if (low === 'clear') {
      const o = document.getElementById('term-out');
      if (o) o.textContent = '';
      return null;
    }
    if (low === 'mem') return JSON.stringify(mem, null, 2);
    if (low === 'studio' || low === 'editor' || low === 'mind') { launch(low); return 'ok'; }
    if (low === 'status') {
      const st = await api('/v1/status');
      return JSON.stringify(st, null, 2);
    }
    if (low === 'organism') {
      const r = await api('/v1/organism/run', { method: 'POST' });
      return JSON.stringify(r, null, 2);
    }
    const L = parseLisp(raw);
    if (!L) return 'desconocido — help';
    const { op, parts } = L;
    if (op === 'recordar' || op === 'set-state') {
      mem[parts[0]] = parts.slice(1).join(' ');
      return 'ok ' + parts[0];
    }
    if (op === 'leer' || op === 'get-state') {
      return parts[0] + ' => ' + (mem[parts[0]] ?? 'nil');
    }
    if (op === 'mind') {
      const text = parts.join(' ');
      const r = await api('/v1/mind/tick', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ text }),
      });
      updateTrayMind(r);
      return JSON.stringify(r, null, 2);
    }
    if (op === 'zyrion') {
      const a = Number(parts[0]), b = Number(parts[1]);
      const r = await api('/v1/zyrion', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ a, b }),
      });
      return JSON.stringify(r, null, 2);
    }
    if (op === 'assert') {
      const [s, r, o, c] = parts;
      const res = await api('/v1/syllogism/assert', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ s, r, o, c: c != null ? Number(c) : 1 }),
      });
      return JSON.stringify(res, null, 2);
    }
    if (op === 'infer' || op === 'ask') {
      const res = await api('/v1/syllogism/' + (op === 'ask' ? 'ask' : 'infer'), {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ s: parts[0], r: parts[1], o: parts[2] }),
      });
      return JSON.stringify(res, null, 2);
    }
    if (op === 'neural') {
      const k = parts[0];
      const v = Number(parts[1]);
      const res = await api('/v1/neural', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ key: k, value: v }),
      });
      return JSON.stringify(res, null, 2);
    }
    if (op === 'neural-get') {
      const res = await api('/v1/neural?key=' + encodeURIComponent(parts[0]));
      return JSON.stringify(res, null, 2);
    }
    return 'forma no reconocida: ' + op;
  }

  function updateTrayMind(r) {
    if (!r) return;
    const d = r.decision || r.voice || r.estado || '—';
    trayMind.textContent = 'Mind · ' + String(d).slice(0, 24);
  }

  function openMind() {
    const w = createWindow({
      id: 'app-mind',
      title: 'Mind · Zyrion · Neural · Silogismos',
      width: 760,
      height: 560,
      contentHTML: `
<div class="cog-grid two">
  <div class="card">
    <h3>Mind · latido</h3>
    <textarea id="mind-in" rows="3" placeholder="observa el escritorio · quién eres"></textarea>
    <div class="row"><button type="button" class="btn btn-gold" id="btn-mind">Enviar latido</button></div>
    <pre class="out" id="mind-out">—</pre>
  </div>
  <div class="card">
    <h3>Zyrion · ternario</h3>
    <div class="row">
      <label>a <input id="zyr-a" type="number" step="0.1" value="0.2" style="width:80px" /></label>
      <label>b <input id="zyr-b" type="number" step="0.1" value="0.8" style="width:80px" /></label>
    </div>
    <div class="row"><button type="button" class="btn btn-gold" id="btn-zyr">Evaluar</button></div>
    <pre class="out" id="zyr-out">—</pre>
  </div>
  <div class="card">
    <h3>Silogismos</h3>
    <input id="syl-s" placeholder="sujeto" value="organismo" />
    <input id="syl-r" placeholder="relación" value="tiene_capacidad" style="margin-top:6px" />
    <input id="syl-o" placeholder="objeto" value="backup" style="margin-top:6px" />
    <div class="row">
      <button type="button" class="btn btn-ghost" id="btn-assert">Assert</button>
      <button type="button" class="btn btn-ghost" id="btn-infer">Infer</button>
      <button type="button" class="btn btn-ghost" id="btn-ask">Ask</button>
    </div>
    <pre class="out" id="syl-out">—</pre>
  </div>
  <div class="card">
    <h3>Red neuronal (pesos)</h3>
    <input id="neu-k" placeholder="clave" value="atencion.ui" />
    <input id="neu-v" type="number" step="0.05" min="0" max="1" value="0.5" style="margin-top:6px" />
    <div class="row">
      <button type="button" class="btn btn-gold" id="btn-neu-set">Guardar peso</button>
      <button type="button" class="btn btn-ghost" id="btn-neu-list">Listar</button>
    </div>
    <pre class="out" id="neu-out">—</pre>
  </div>
</div>
<p style="color:var(--muted);font-size:12px;margin-top:12px">Estos servicios viven en el <strong>sistema</strong> (bridge). Las apps Studio/Editor pueden llamar las mismas rutas <code>/v1/mind/*</code>, <code>/v1/zyrion</code>, <code>/v1/syllogism/*</code>, <code>/v1/neural</code>.</p>`,
    });
    const root = w.el;
    root.querySelector('#btn-mind').onclick = async () => {
      const text = root.querySelector('#mind-in').value;
      const r = await api('/v1/mind/tick', {
        method: 'POST', headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ text }),
      });
      root.querySelector('#mind-out').textContent = JSON.stringify(r, null, 2);
      updateTrayMind(r);
    };
    root.querySelector('#btn-zyr').onclick = async () => {
      const a = Number(root.querySelector('#zyr-a').value);
      const b = Number(root.querySelector('#zyr-b').value);
      const r = await api('/v1/zyrion', {
        method: 'POST', headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ a, b }),
      });
      root.querySelector('#zyr-out').textContent = JSON.stringify(r, null, 2);
    };
    const sylBody = () => ({
      s: root.querySelector('#syl-s').value,
      r: root.querySelector('#syl-r').value,
      o: root.querySelector('#syl-o').value,
      c: 1,
    });
    root.querySelector('#btn-assert').onclick = async () => {
      const r = await api('/v1/syllogism/assert', {
        method: 'POST', headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(sylBody()),
      });
      root.querySelector('#syl-out').textContent = JSON.stringify(r, null, 2);
    };
    root.querySelector('#btn-infer').onclick = async () => {
      const r = await api('/v1/syllogism/infer', {
        method: 'POST', headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(sylBody()),
      });
      root.querySelector('#syl-out').textContent = JSON.stringify(r, null, 2);
    };
    root.querySelector('#btn-ask').onclick = async () => {
      const r = await api('/v1/syllogism/ask', {
        method: 'POST', headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(sylBody()),
      });
      root.querySelector('#syl-out').textContent = JSON.stringify(r, null, 2);
    };
    root.querySelector('#btn-neu-set').onclick = async () => {
      const r = await api('/v1/neural', {
        method: 'POST', headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          key: root.querySelector('#neu-k').value,
          value: Number(root.querySelector('#neu-v').value),
        }),
      });
      root.querySelector('#neu-out').textContent = JSON.stringify(r, null, 2);
    };
    root.querySelector('#btn-neu-list').onclick = async () => {
      const r = await api('/v1/neural');
      root.querySelector('#neu-out').textContent = JSON.stringify(r, null, 2);
    };
  }

  async function openOrganism() {
    const w = createWindow({
      id: 'app-organism',
      title: 'Organismo local',
      width: 520,
      height: 360,
      contentHTML: `<p>Ejecuta un organismo AlsetOS vía el bridge.</p>
        <div class="row"><button type="button" class="btn btn-gold" id="btn-run-org">Ejecutar</button></div>
        <pre class="out" id="org-out">—</pre>`,
    });
    w.el.querySelector('#btn-run-org').onclick = async () => {
      const r = await api('/v1/organism/run', { method: 'POST' });
      w.el.querySelector('#org-out').textContent = JSON.stringify(r, null, 2);
    };
  }

  async function openStatus() {
    const st = await api('/v1/status');
    createWindow({
      id: 'app-status',
      title: 'Estado del sistema',
      width: 520,
      height: 400,
      contentHTML: `<pre class="status-pre">${JSON.stringify(st, null, 2)}</pre>`,
    });
  }

  async function openFiles() {
    const st = await api('/v1/fs/list');
    createWindow({
      id: 'app-files',
      title: 'Archivos · data dir',
      width: 480,
      height: 360,
      contentHTML: `<pre class="status-pre">${JSON.stringify(st, null, 2)}</pre>`,
    });
  }

  // ——— Chrome ——
  document.getElementById('btn-start').onclick = (e) => {
    e.stopPropagation();
    startMenu.classList.toggle('hidden');
  };
  document.addEventListener('click', (e) => {
    if (!startMenu.contains(e.target) && e.target.id !== 'btn-start') {
      startMenu.classList.add('hidden');
    }
  });
  document.querySelectorAll('[data-launch]').forEach((btn) => {
    btn.addEventListener('click', () => launch(btn.getAttribute('data-launch')));
  });

  function clock() {
    trayClock.textContent = new Date().toLocaleString();
  }
  setInterval(clock, 1000);
  clock();

  // Boot cognitive ping
  api('/v1/mind/tick', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ text: 'escritorio listo' }),
  }).then(updateTrayMind).catch(() => {
    trayMind.textContent = 'Mind · offline';
  });

  renderIcons();
})();
