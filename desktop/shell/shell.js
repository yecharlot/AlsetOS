/**
 * Alset Desktop WM — icons, windows, files, apps, theme, LispAI control
 */
(function () {
  const ICON_KEY = 'alset-desktop-icons-v3';
  const THEME_KEY = 'alset-desktop-theme-v1';
  const WIN_KEY = 'alset-desktop-windows-v1';
  const SESSION_AUTH = 'alset-desktop-auth-v1';
  const iconLayer = document.getElementById('icon-layer');
  const windowLayer = document.getElementById('window-layer');
  const taskButtons = document.getElementById('task-buttons');
  const startMenu = document.getElementById('start-menu');
  const trayMind = document.getElementById('tray-mind');
  const trayClock = document.getElementById('tray-clock');

  let zTop = 30;
  const windows = new Map();
  const mem = Object.create(null);

  const DEFAULT_ICONS = [
    { id: 'studio', label: 'Studio', glyph: '◈', x: 24, y: 24 },
    { id: 'editor', label: 'Editor JS', glyph: '◇', x: 24, y: 120 },
    { id: 'files', label: 'Archivos', glyph: '📁', x: 24, y: 216 },
    { id: 'terminal', label: 'Terminal', glyph: '▣', x: 24, y: 312 },
    { id: 'mind', label: 'Mind', glyph: '◎', x: 24, y: 408 },
    { id: 'apps', label: 'Apps', glyph: '▦', x: 120, y: 24 },
    { id: 'settings', label: 'Ajustes', glyph: '⚙', x: 120, y: 120 },
    { id: 'organism', label: 'Organismo', glyph: '⚡', x: 120, y: 216 },
  ];

  // —— Theme ——
  function applyTheme(t) {
    const root = document.documentElement;
    const cfg = Object.assign({
      gold: '#f0c14b', accent: '#5b9fd4', bg0: '#06080c', bg1: '#0c1018',
      font: 'system-ui', iconScale: '1', wallpaper: 'default',
    }, t || {});
    root.style.setProperty('--gold', cfg.gold);
    root.style.setProperty('--accent', cfg.accent);
    root.style.setProperty('--bg0', cfg.bg0);
    root.style.setProperty('--bg1', cfg.bg1);
    document.body.style.fontFamily = cfg.font + ', system-ui, sans-serif';
    document.documentElement.style.setProperty('--icon-scale', cfg.iconScale);
    const desk = document.getElementById('desktop');
    if (cfg.wallpaper === 'aurora') {
      desk.style.background = 'radial-gradient(ellipse 80% 60% at 20% 0%, rgba(91,159,212,.28), transparent 55%), radial-gradient(ellipse at 100% 100%, rgba(240,193,75,.15), transparent 50%), #06080c';
    } else if (cfg.wallpaper === 'ember') {
      desk.style.background = 'radial-gradient(ellipse at 30% 20%, rgba(240,100,50,.2), transparent 50%), #0a0808';
    } else if (cfg.wallpaper === 'plain') {
      desk.style.background = cfg.bg0;
    } else {
      desk.style.background = '';
    }
    localStorage.setItem(THEME_KEY, JSON.stringify(cfg));
    document.querySelectorAll('.desk-icon .glyph').forEach((g) => {
      g.style.transform = 'scale(' + (cfg.iconScale || 1) + ')';
    });
  }
  try { applyTheme(JSON.parse(localStorage.getItem(THEME_KEY) || '{}')); } catch (_) { applyTheme({}); }

  function loadIcons() {
    try {
      const raw = JSON.parse(localStorage.getItem(ICON_KEY) || 'null');
      if (Array.isArray(raw) && raw.length) return raw;
    } catch (_) {}
    return DEFAULT_ICONS.map((i) => ({ ...i }));
  }
  function saveIcons(list) { localStorage.setItem(ICON_KEY, JSON.stringify(list)); }
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
      btn.addEventListener('contextmenu', (e) => {
        e.preventDefault();
        showIconMenu(e.clientX, e.clientY, ic);
      });
      btn.addEventListener('click', () => {
        if (btn._dragged) { btn._dragged = false; return; }
        document.querySelectorAll('.desk-icon').forEach((x) => x.classList.remove('selected'));
        btn.classList.add('selected');
      });
      iconLayer.appendChild(btn);
    });
    try {
      const th = JSON.parse(localStorage.getItem(THEME_KEY) || '{}');
      document.querySelectorAll('.desk-icon .glyph').forEach((g) => {
        g.style.transform = 'scale(' + (th.iconScale || 1) + ')';
      });
    } catch (_) {}
  }

  function enableIconDrag(el, ic) {
    let sx, sy, ox, oy, moved;
    el.addEventListener('pointerdown', (e) => {
      if (e.button !== 0) return;
      el.setPointerCapture(e.pointerId);
      sx = e.clientX; sy = e.clientY; ox = ic.x; oy = ic.y; moved = false;
      el.classList.add('dragging');
      const onMove = (ev) => {
        const dx = ev.clientX - sx, dy = ev.clientY - sy;
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
        if (moved) { el._dragged = true; saveIcons(iconState); }
      };
      el.addEventListener('pointermove', onMove);
      el.addEventListener('pointerup', onUp);
    });
  }

  function focusWin(id) {
    const w = windows.get(id);
    if (!w) return;
    zTop += 1;
    w.el.style.zIndex = String(zTop);
    w.el.classList.add('active');
    w.el.style.display = '';
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
  function createWindow({ id, title, width, height, x, y, contentHTML, iframeSrc }) {
    if (windows.has(id)) { focusWin(id); return windows.get(id); }
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
          <button type="button" data-act="min">–</button>
          <button type="button" data-act="max">□</button>
          <button type="button" class="close" data-act="close">×</button>
        </div>
      </div>
      <div class="win-body"></div>
      <div class="win-resize" data-resize></div>`;
    const body = el.querySelector('.win-body');
    if (iframeSrc) {
      const iframe = document.createElement('iframe');
      iframe.src = iframeSrc;
      iframe.title = title;
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
      if (el.style.display === 'none') { el.style.display = ''; focusWin(id); }
      else if (el.classList.contains('active')) el.style.display = 'none';
      else focusWin(id);
    };
    taskButtons.appendChild(taskBtn);
    el.querySelector('[data-act="close"]').onclick = (e) => { e.stopPropagation(); closeWin(id); };
    el.querySelector('[data-act="max"]').onclick = (e) => { e.stopPropagation(); el.classList.toggle('maximized'); };
    el.querySelector('[data-act="min"]').onclick = (e) => { e.stopPropagation(); el.style.display = 'none'; };
    el.addEventListener('mousedown', () => focusWin(id));
    enableWinDrag(el);
    enableWinResize(el);
    windowLayer.appendChild(el);
    windows.set(id, { el, taskBtn, title });
    focusWin(id);
    try { saveWindowsSession(); } catch (_) {}
    return windows.get(id);
  }
  function enableWinDrag(el) {
    const bar = el.querySelector('[data-drag]');
    bar.addEventListener('pointerdown', (e) => {
      if (e.target.closest('.win-controls') || el.classList.contains('maximized')) return;
      e.preventDefault();
      const rect = el.getBoundingClientRect();
      const desk = document.getElementById('desktop').getBoundingClientRect();
      const ox = e.clientX - rect.left, oy = e.clientY - rect.top;
      bar.setPointerCapture(e.pointerId);
      const onMove = (ev) => {
        el.style.left = Math.max(-rect.width + 80, Math.min(desk.width - 40, ev.clientX - desk.left - ox)) + 'px';
        el.style.top = Math.max(0, Math.min(desk.height - 40, ev.clientY - desk.top - oy)) + 'px';
      };
      const onUp = () => { bar.releasePointerCapture?.(e.pointerId); bar.removeEventListener('pointermove', onMove); bar.removeEventListener('pointerup', onUp); };
      bar.addEventListener('pointermove', onMove);
      bar.addEventListener('pointerup', onUp);
    });
  }
  function enableWinResize(el) {
    const handle = el.querySelector('[data-resize]');
    handle.addEventListener('pointerdown', (e) => {
      if (el.classList.contains('maximized')) return;
      e.preventDefault(); e.stopPropagation();
      const sx = e.clientX, sy = e.clientY, sw = el.offsetWidth, sh = el.offsetHeight;
      handle.setPointerCapture(e.pointerId);
      const onMove = (ev) => {
        el.style.width = Math.max(320, sw + (ev.clientX - sx)) + 'px';
        el.style.height = Math.max(200, sh + (ev.clientY - sy)) + 'px';
      };
      const onUp = () => { handle.releasePointerCapture?.(e.pointerId); handle.removeEventListener('pointermove', onMove); handle.removeEventListener('pointerup', onUp); };
      handle.addEventListener('pointermove', onMove);
      handle.addEventListener('pointerup', onUp);
    });
  }

  async function api(path, opts) {
    const r = await fetch(path, opts);
    const t = await r.text();
    try { return JSON.parse(t); } catch { return { raw: t, ok: r.ok, status: r.status }; }
  }

  function launch(id) {
    startMenu.classList.add('hidden');
    if (id === 'studio') {
      createWindow({
        id: 'app-studio', title: 'Alset Studio',
        width: Math.min(1100, window.innerWidth - 40),
        height: Math.min(720, window.innerHeight - 80),
        iframeSrc: '/tools/',
      });
      return;
    }
    if (id === 'editor') {
      createWindow({
        id: 'app-editor', title: 'Alset-JS Editor',
        width: Math.min(1100, window.innerWidth - 40),
        height: Math.min(720, window.innerHeight - 80),
        iframeSrc: '/alset-editor/',
      });
      return;
    }
    if (id === 'files') return openFiles();
    if (id === 'terminal') return openTerminal();
    if (id === 'mind') return openMind();
    if (id === 'apps') return openApps();
    if (id === 'settings') return openSettings();
    if (id === 'organism') return openOrganism();
    if (id === 'status') return openStatus();
    if (id === 'about') {
      createWindow({
        id: 'app-about', title: 'Acerca de Alset OS', width: 440, height: 300,
        contentHTML: `<p><strong>Alset Desktop</strong></p>
          <p class="status-pre">Ventanas · archivos · apps instaladas · Mind/Zyrion/Neural · LispAI en terminal.
Studio/Editor dentro del SO. Despliega apps y ábrelas aquí.</p>`,
      });
    }
  }

  // —— Files ——
  function openFiles(startPath) {
    const w = createWindow({
      id: 'app-files', title: 'Archivos · AlsetOS', width: 640, height: 480,
      contentHTML: `
        <div class="row" style="margin-bottom:8px;gap:8px;display:flex;flex-wrap:wrap;align-items:center">
          <button type="button" class="btn btn-ghost" id="fs-up">↑ Padre</button>
          <button type="button" class="btn btn-ghost" id="fs-refresh">Actualizar</button>
          <button type="button" class="btn btn-gold" id="fs-mkdir">Nueva carpeta</button>
          <span id="fs-path" style="color:var(--muted);font-size:12px;font-family:monospace"></span>
        </div>
        <div id="fs-list" class="fs-list"></div>
        <pre class="out" id="fs-preview" style="max-height:160px">Selecciona un archivo de texto para previsualizar</pre>`,
    });
    let cur = startPath || '';
    const listEl = w.el.querySelector('#fs-list');
    const pathEl = w.el.querySelector('#fs-path');
    const prev = w.el.querySelector('#fs-preview');

    async function load(path) {
      cur = path || '';
      const data = await api('/v1/fs/list?path=' + encodeURIComponent(cur));
      pathEl.textContent = data.rel || data.path || '/';
      listEl.innerHTML = '';
      (data.entries || []).forEach((e) => {
        const row = document.createElement('button');
        row.type = 'button';
        row.className = 'fs-row';
        row.innerHTML = `<span>${e.dir ? '📁' : '📄'}</span> <span>${e.name}</span>` +
          (e.size != null ? `<span class="sz">${e.size} B</span>` : '');
        row.onclick = async () => {
          const next = cur ? (cur.replace(/\/$/, '') + '/' + e.name) : e.name;
          if (e.dir) load(next);
          else {
            const r = await api('/v1/fs/read?path=' + encodeURIComponent(next));
            prev.textContent = r.content != null ? r.content : JSON.stringify(r);
          }
        };
        listEl.appendChild(row);
      });
    }
    w.el.querySelector('#fs-up').onclick = async () => {
      const data = await api('/v1/fs/list?path=' + encodeURIComponent(cur));
      load(data.parent === data.root ? '' : (data.parent || ''));
    };
    w.el.querySelector('#fs-refresh').onclick = () => load(cur);
    w.el.querySelector('#fs-mkdir').onclick = async () => {
      const name = prompt('Nombre de carpeta');
      if (!name) return;
      const path = cur ? cur + '/' + name : name;
      await api('/v1/fs/mkdir', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ path }) });
      load(cur);
    };
    load(cur);
  }

  // —— Apps ——
  async function openApps() {
    const w = createWindow({
      id: 'app-apps', title: 'Apps instaladas', width: 520, height: 420,
      contentHTML: `
        <div class="row" style="display:flex;gap:8px;margin-bottom:12px;flex-wrap:wrap">
          <button type="button" class="btn btn-gold" id="btn-deploy-calc">Instalar Calculadora</button>
          <button type="button" class="btn btn-ghost" id="btn-apps-refresh">Actualizar</button>
        </div>
        <div id="apps-list"></div>
        <p style="color:var(--muted);font-size:12px;margin-top:12px">Despliega desde Studio/Editor o terminal: <code>(deploy-app nombre)</code></p>`,
    });
    async function refresh() {
      const data = await api('/v1/apps/list');
      const box = w.el.querySelector('#apps-list');
      box.innerHTML = '';
      (data.apps || []).forEach((a) => {
        const row = document.createElement('div');
        row.className = 'card';
        row.style.marginBottom = '8px';
        row.innerHTML = `<strong>${a.title || a.name}</strong>
          <div class="row" style="margin-top:8px">
            <button type="button" class="btn btn-gold" data-open="${a.name}">Abrir</button>
          </div>`;
        row.querySelector('[data-open]').onclick = () => openInstalledApp(a.name, a.title || a.name);
        box.appendChild(row);
      });
      if (!(data.apps || []).length) box.innerHTML = '<p style="color:var(--muted)">Ninguna app aún. Instala la calculadora o despliega desde Studio.</p>';
    }
    w.el.querySelector('#btn-apps-refresh').onclick = refresh;
    w.el.querySelector('#btn-deploy-calc').onclick = async () => {
      await api('/v1/apps/deploy', {
        method: 'POST', headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ name: 'calculadora', title: 'Calculadora', kind: 'html' }),
      });
      // add desktop icon if missing
      if (!iconState.find((i) => i.id === 'app:seed')) {
        iconState.push({ id: 'app:seed', label: 'Seed', glyph: '🌱', x: 216, y: 24 });
      }
      if (!iconState.find((i) => i.id === 'app:calculadora')) {
        iconState.push({ id: 'app:calculadora', label: 'Calculadora', glyph: '🔢', x: 216, y: 120 });
        saveIcons(iconState);
        renderIcons();
      }
      refresh();
    };
    refresh();
  }

  function openInstalledApp(name, title) {
    const sizes = {
      'files-ui': [720, 520], 'doc-reader': [760, 540], 'notes': [720, 520],
      'org-manager': [700, 500], 'accounts': [640, 480], 'ipfs-store': [680, 500],
      'calculadora': [360, 480], 'seed': [900, 620], 'image-viewer': [560, 480],
      'audio-player': [480, 320], 'video-player': [640, 480],
    };
    const [w, h] = sizes[name] || [640, 520];
    createWindow({
      id: 'installed-' + name,
      title: title || name,
      width: w,
      height: h,
      iframeSrc: '/apps/' + encodeURIComponent(name) + '/',
    });
  }

  // Patch launch for app: icons
  const _launch = launch;
  launch = function (id) {
    if (id === 'logout') {
      try { localStorage.removeItem(SESSION_AUTH); localStorage.removeItem('alset_token'); } catch (_) {}
      location.reload();
      return;
    }
    if (id === 'shutdown') {
      if (confirm('¿Apagar el equipo?')) {
        api('/v1/power/shutdown', { method: 'POST' }).catch(() => {});
      }
      return;
    }
    if (id === 'reboot') {
      if (confirm('¿Reiniciar el equipo?')) {
        api('/v1/power/reboot', { method: 'POST' }).catch(() => {});
      }
      return;
    }
    if (id && id.startsWith('app:')) {
      const name = id.slice(4);
      const labels = {
        'files-ui': 'Archivos', 'doc-reader': 'Documentos', 'notes': 'Notas',
        'org-manager': 'Organismos', 'accounts': 'Cuentas', 'ipfs-store': 'IPFS Store',
        'calculadora': 'Calculadora', 'seed': 'Seed', 'image-viewer': 'Imágenes',
        'audio-player': 'Audio', 'video-player': 'Video',
      };
      openInstalledApp(name, labels[name] || name);
      return;
    }
    return _launch(id);
  };

  // —— Settings ——
  function openSettings() {
    let cur = {};
    try { cur = JSON.parse(localStorage.getItem(THEME_KEY) || '{}'); } catch (_) {}
    const w = createWindow({
      id: 'app-settings', title: 'Ajustes del escritorio', width: 480, height: 460,
      contentHTML: `
        <div class="card"><h3>Tema</h3>
          <label>Fondo
            <select id="set-wall">
              <option value="default">Default</option>
              <option value="aurora">Aurora</option>
              <option value="ember">Ember</option>
              <option value="plain">Liso</option>
            </select>
          </label>
          <label style="display:block;margin-top:8px">Acento <input id="set-gold" type="color" value="${cur.gold || '#f0c14b'}" /></label>
          <label style="display:block;margin-top:8px">Tipografía
            <select id="set-font">
              <option value="system-ui">System</option>
              <option value="Georgia">Georgia</option>
              <option value="ui-monospace">Mono</option>
              <option value="Inter">Inter</option>
            </select>
          </label>
          <label style="display:block;margin-top:8px">Tamaño iconos
            <input id="set-icon" type="range" min="0.8" max="1.4" step="0.1" value="${cur.iconScale || 1}" />
          </label>
          <div class="row" style="margin-top:12px">
            <button type="button" class="btn btn-gold" id="set-apply">Aplicar</button>
            <button type="button" class="btn btn-ghost" id="set-reset">Reset iconos</button>
          </div>
        </div>`,
    });
    w.el.querySelector('#set-wall').value = cur.wallpaper || 'default';
    w.el.querySelector('#set-font').value = cur.font || 'system-ui';
    w.el.querySelector('#set-apply').onclick = () => {
      applyTheme({
        gold: w.el.querySelector('#set-gold').value,
        font: w.el.querySelector('#set-font').value,
        iconScale: w.el.querySelector('#set-icon').value,
        wallpaper: w.el.querySelector('#set-wall').value,
      });
      renderIcons();
    };
    w.el.querySelector('#set-reset').onclick = () => {
      iconState = DEFAULT_ICONS.map((i) => ({ ...i }));
      saveIcons(iconState);
      renderIcons();
    };
  }

  // —— Terminal + LispAI ——
  function openTerminal() {
    const w = createWindow({
      id: 'app-terminal', title: 'Terminal · LispAI / AlsetOS', width: 680, height: 440,
      contentHTML: `<pre class="term-out" id="term-out">Alset LispAI Terminal
help · ls · cat path · open studio|editor|files|apps|settings
(theme gold "#f0c14b") (wallpaper aurora) (icon-scale 1.2)
(deploy-app calculadora) (open-app calculadora)
(recordar k v) (mind …) (zyrion a b)
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
        const res = await execLispAI(line);
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

  async function execLispAI(line) {
    const raw = (line || '').trim();
    if (!raw) return null;
    const low = raw.toLowerCase();
    if (low === 'help') {
      return `Comandos:
  ls [path] | cat path | open <studio|editor|files|apps|settings|mind>
  (theme gold "#hex") (wallpaper aurora|ember|plain|default)
  (icon-scale 1.2) (deploy-app name) (open-app name) (apps)
  (recordar k v) (leer k) (mind texto) (zyrion a b)
  (assert s r o) (infer) (neural k v) status organism`;
    }
    if (low === 'clear') {
      const o = document.getElementById('term-out');
      if (o) o.textContent = '';
      return null;
    }
    if (low.startsWith('ls')) {
      const path = raw.slice(2).trim();
      const data = await api('/v1/fs/list?path=' + encodeURIComponent(path));
      return (data.entries || []).map((e) => (e.dir ? '📁 ' : '📄 ') + e.name).join('\n') || '(vacío)';
    }
    if (low.startsWith('cat ')) {
      const r = await api('/v1/fs/read?path=' + encodeURIComponent(raw.slice(4).trim()));
      return r.content != null ? r.content : JSON.stringify(r);
    }
    if (low.startsWith('open ')) {
      launch(raw.slice(5).trim());
      return 'ok';
    }
    if (low === 'status') return JSON.stringify(await api('/v1/status'), null, 2);
    if (low === 'organism') return JSON.stringify(await api('/v1/organism/run', { method: 'POST' }), null, 2);
    if (low === 'apps') return JSON.stringify(await api('/v1/apps/list'), null, 2);

    const L = parseLisp(raw);
    if (!L) return 'desconocido — help';
    const { op, parts } = L;
    if (op === 'theme') {
      const key = parts[0], val = parts[1];
      const cur = JSON.parse(localStorage.getItem(THEME_KEY) || '{}');
      if (key === 'gold' || key === 'accent') cur[key] = val;
      applyTheme(cur);
      return 'theme ok';
    }
    if (op === 'wallpaper') {
      const cur = JSON.parse(localStorage.getItem(THEME_KEY) || '{}');
      cur.wallpaper = parts[0];
      applyTheme(cur);
      return 'wallpaper ' + parts[0];
    }
    if (op === 'icon-scale') {
      const cur = JSON.parse(localStorage.getItem(THEME_KEY) || '{}');
      cur.iconScale = parts[0];
      applyTheme(cur);
      renderIcons();
      return 'icon-scale ' + parts[0];
    }
    if (op === 'deploy-app') {
      const name = parts[0] || 'calculadora';
      const r = await api('/v1/apps/deploy', {
        method: 'POST', headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ name, title: name, kind: 'html' }),
      });
      if (!iconState.find((i) => i.id === 'app:' + name)) {
        iconState.push({ id: 'app:' + name, label: name, glyph: '📦', x: 216, y: 24 + iconState.length * 20 });
        saveIcons(iconState); renderIcons();
      }
      return JSON.stringify(r);
    }
    if (op === 'open-app') {
      openInstalledApp(parts[0], parts[0]);
      return 'ok';
    }
    if (op === 'open') { launch(parts[0]); return 'ok'; }
    if (op === 'recordar' || op === 'set-state') {
      mem[parts[0]] = parts.slice(1).join(' ');
      return 'ok';
    }
    if (op === 'leer' || op === 'get-state') return parts[0] + ' => ' + (mem[parts[0]] ?? 'nil');
    if (op === 'mind') {
      const r = await api('/v1/mind/tick', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ text: parts.join(' ') }) });
      updateTrayMind(r);
      return JSON.stringify(r, null, 2);
    }
    if (op === 'zyrion') {
      return JSON.stringify(await api('/v1/zyrion', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ a: Number(parts[0]), b: Number(parts[1]) }) }), null, 2);
    }
    if (op === 'assert') {
      return JSON.stringify(await api('/v1/syllogism/assert', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ s: parts[0], r: parts[1], o: parts[2], c: 1 }) }), null, 2);
    }
    if (op === 'infer' || op === 'ask') {
      return JSON.stringify(await api('/v1/syllogism/' + op, { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ s: parts[0], r: parts[1], o: parts[2] }) }), null, 2);
    }
    if (op === 'neural') {
      return JSON.stringify(await api('/v1/neural', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ key: parts[0], value: Number(parts[1]) }) }), null, 2);
    }
    if (op === 'ls') {
      const data = await api('/v1/fs/list?path=' + encodeURIComponent(parts[0] || ''));
      return (data.entries || []).map((e) => e.name).join('\n');
    }
    return 'forma no reconocida: ' + op;
  }

  function updateTrayMind(r) {
    if (!r) return;
    trayMind.textContent = 'Mind · ' + String(r.decision || r.voice || '—').slice(0, 28);
  }

  function openMind() {
    const w = createWindow({
      id: 'app-mind', title: 'Mind · Zyrion · Neural · Silogismos', width: 760, height: 520,
      contentHTML: `
<div class="cog-grid two">
  <div class="card"><h3>Mind</h3>
    <textarea id="mind-in" rows="2">escritorio alset</textarea>
    <div class="row"><button type="button" class="btn btn-gold" id="btn-mind">Latido</button></div>
    <pre class="out" id="mind-out">—</pre>
  </div>
  <div class="card"><h3>Zyrion</h3>
    <div class="row"><input id="zyr-a" type="number" step="0.1" value="0.2" style="width:80px"/><input id="zyr-b" type="number" step="0.1" value="0.8" style="width:80px"/></div>
    <div class="row"><button type="button" class="btn btn-gold" id="btn-zyr">Evaluar</button></div>
    <pre class="out" id="zyr-out">—</pre>
  </div>
</div>`,
    });
    w.el.querySelector('#btn-mind').onclick = async () => {
      const r = await api('/v1/mind/tick', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ text: w.el.querySelector('#mind-in').value }) });
      w.el.querySelector('#mind-out').textContent = JSON.stringify(r, null, 2);
      updateTrayMind(r);
    };
    w.el.querySelector('#btn-zyr').onclick = async () => {
      const r = await api('/v1/zyrion', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ a: Number(w.el.querySelector('#zyr-a').value), b: Number(w.el.querySelector('#zyr-b').value) }) });
      w.el.querySelector('#zyr-out').textContent = JSON.stringify(r, null, 2);
    };
  }

  async function openOrganism() {
    const w = createWindow({
      id: 'app-organism', title: 'Organismo', width: 480, height: 320,
      contentHTML: `<button type="button" class="btn btn-gold" id="btn-run-org">Ejecutar</button><pre class="out" id="org-out">—</pre>`,
    });
    w.el.querySelector('#btn-run-org').onclick = async () => {
      w.el.querySelector('#org-out').textContent = JSON.stringify(await api('/v1/organism/run', { method: 'POST' }), null, 2);
    };
  }

  async function openStatus() {
    createWindow({
      id: 'app-status', title: 'Estado', width: 520, height: 360,
      contentHTML: `<pre class="status-pre">${JSON.stringify(await api('/v1/status'), null, 2)}</pre>`,
    });
  }

  // start menu — ensure items exist in HTML; wire data-launch
  document.getElementById('btn-start').onclick = (e) => {
    e.stopPropagation();
    startMenu.classList.toggle('hidden');
  };
  document.addEventListener('click', (e) => {
    if (!startMenu.contains(e.target) && e.target.id !== 'btn-start') startMenu.classList.add('hidden');
  });
  document.querySelectorAll('[data-launch]').forEach((btn) => {
    btn.addEventListener('click', () => launch(btn.getAttribute('data-launch')));
  });

  function clock() { trayClock.textContent = new Date().toLocaleString(); }
  document.getElementById('tray-net')?.addEventListener('click', () => openNetTray());
  // refresh net tray label
  async function refreshNetChip() {
    try {
      const net = await api('/v1/network');
      const up = (net.interfaces || []).filter((i) => i.state === 'up').length;
      const el = document.getElementById('tray-net');
      if (el) el.textContent = '🌐 ' + up + ' up';
    } catch (_) {}
  }
  refreshNetChip();
  setInterval(refreshNetChip, 15000);
  setInterval(clock, 1000); clock();

  api('/v1/mind/tick', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ text: 'escritorio listo' }) })
    .then(updateTrayMind).catch(() => { trayMind.textContent = 'Mind · offline'; });



  function showIconMenu(x, y, ic) {
    document.querySelectorAll('.ctx-menu').forEach((n) => n.remove());
    const m = document.createElement('div');
    m.className = 'ctx-menu';
    m.style.left = x + 'px';
    m.style.top = y + 'px';
    const open = document.createElement('button');
    open.textContent = 'Abrir';
    open.onclick = () => { m.remove(); launch(ic.id); };
    const del = document.createElement('button');
    del.textContent = 'Quitar icono';
    del.onclick = () => {
      m.remove();
      iconState = iconState.filter((i) => i.id !== ic.id);
      saveIcons(iconState);
      renderIcons();
    };
    m.appendChild(open);
    m.appendChild(del);
    if (ic.id.startsWith('app:')) {
      const un = document.createElement('button');
      un.textContent = 'Desinstalar app';
      un.onclick = async () => {
        m.remove();
        const name = ic.id.slice(4);
        await api('/v1/apps/uninstall', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ name }),
        });
        iconState = iconState.filter((i) => i.id !== ic.id);
        saveIcons(iconState);
        renderIcons();
        closeWin('installed-' + name);
      };
      m.appendChild(un);
    }
    document.body.appendChild(m);
    const once = () => { m.remove(); document.removeEventListener('click', once); };
    setTimeout(() => document.addEventListener('click', once), 10);
  }

  async function openNetTray() {
    let net = {}, vol = {}, wifi = {};
    try { net = await api('/v1/network'); } catch (_) {}
    try { vol = await api('/v1/volumes'); } catch (_) {}
    try { wifi = await api('/v1/wifi/scan'); } catch (_) {}
    const nets = (wifi.networks || []).map((n) =>
      `<div class="item wifi-row" data-ssid="${(n.ssid || '').replace(/"/g, '')}">
        <span>${n.ssid || '—'}</span>
        <span class="muted">${n.signal || ''} ${n.security || ''}</span>
      </div>`
    ).join('') || '<p class="muted">Sin redes detectadas (o sin nmcli)</p>';
    createWindow({
      id: 'app-net-tray',
      title: 'Red · Wi‑Fi · Volúmenes',
      width: 520,
      height: 520,
      contentHTML: `
        <div class="card"><h3>Wi‑Fi / Hotspot</h3>
          <div id="wifi-list">${nets}</div>
          <div class="row" style="margin-top:8px">
            <input id="wifi-ssid" placeholder="SSID" style="flex:1"/>
            <input id="wifi-pass" type="password" placeholder="Clave" style="flex:1"/>
            <button type="button" class="btn btn-gold" id="wifi-connect">Conectar</button>
          </div>
          <p class="muted" id="wifi-status"></p>
        </div>
        <div class="card"><h3>Interfaces</h3><pre class="status-pre">${JSON.stringify(net.interfaces || [], null, 2)}</pre></div>
        <div class="card"><h3>Volúmenes</h3><pre class="status-pre">${JSON.stringify(vol.volumes || [], null, 2)}</pre></div>`,
    });
    setTimeout(() => {
      document.querySelectorAll('.wifi-row').forEach((el) => {
        el.addEventListener('click', () => {
          const ss = document.getElementById('wifi-ssid');
          if (ss) ss.value = el.getAttribute('data-ssid') || '';
        });
      });
      document.getElementById('wifi-connect')?.addEventListener('click', async () => {
        const ssid = document.getElementById('wifi-ssid')?.value || '';
        const pass = document.getElementById('wifi-pass')?.value || '';
        const st = document.getElementById('wifi-status');
        try {
          const j = await api('/v1/wifi/connect', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ ssid, pass }),
          });
          if (st) st.textContent = j.ok ? 'Conectado' : (j.error || JSON.stringify(j));
        } catch (e) {
          if (st) st.textContent = e.message || String(e);
        }
      });
    }, 50);
  }

  function saveWindowsSession() {
    const list = [];
    windows.forEach((w, id) => {
      if (!id.startsWith('installed-') && id !== 'app-studio' && id !== 'app-editor') return;
      const el = w.el;
      if (!el || el.style.display === 'none') return;
      list.push({
        id,
        title: w.title,
        left: el.style.left,
        top: el.style.top,
        width: el.style.width,
        height: el.style.height,
        url: el.querySelector('iframe')?.src || null,
      });
    });
    try { localStorage.setItem(WIN_KEY, JSON.stringify(list)); } catch (_) {}
  }

  function restoreWindowsSession() {
    let list = [];
    try { list = JSON.parse(localStorage.getItem(WIN_KEY) || '[]'); } catch (_) {}
    list.forEach((s) => {
      if (!s.url) return;
      const name = s.id.replace(/^installed-/, '');
      createWindow({
        id: s.id,
        title: s.title || name,
        width: parseInt(s.width, 10) || 400,
        height: parseInt(s.height, 10) || 480,
        x: parseInt(s.left, 10) || 80,
        y: parseInt(s.top, 10) || 60,
        iframeSrc: s.url,
      });
    });
  }

  // Studio/Editor deploy → icon + window inside desktop
  window.addEventListener('message', (ev) => {
    const d = ev.data;
    if (!d || d.type !== 'alset-desktop-deploy') return;
    const name = d.name || 'app';
    const title = d.title || name;
    const url = d.url || ('/apps/' + encodeURIComponent(name) + '/');
    const iconId = 'app:' + name;
    if (!iconState.find((i) => i.id === iconId)) {
      iconState.push({
        id: iconId,
        label: title.slice(0, 14),
        glyph: '📦',
        x: 216,
        y: 24 + (iconState.length % 6) * 96,
      });
      saveIcons(iconState);
      renderIcons();
    }
    createWindow({
      id: 'installed-' + name,
      title: title,
      width: 480,
      height: 560,
      iframeSrc: url,
    });
    saveWindowsSession();
  });

  // Persist after window moves (debounce)
  const _createWindow = createWindow;
  // wrap closeWin
  const _closeWin = closeWin;
  closeWin = function (id) {
    _closeWin(id);
    saveWindowsSession();
  };

  async function syncInstalledAppIcons() {
    try {
      const data = await api('/v1/apps/list');
      (data.apps || []).forEach((a) => {
        if (a.menu_only) return; // audio/video only in start menu
        if (a.desktop === false) return;
        const iconId = 'app:' + a.name;
        if (!iconState.find((i) => i.id === iconId)) {
          // only auto-pin a few builtins to avoid clutter
          const pin = ['org-manager', 'files-ui', 'seed', 'calculadora'].includes(a.name) || !a.builtin;
          if (!pin) return;
          iconState.push({
            id: iconId,
            label: (a.title || a.name || '').slice(0, 14),
            glyph: a.glyph || '📦',
            x: 216,
            y: 24 + (iconState.length % 6) * 96,
          });
        }
      });
      saveIcons(iconState);
      renderIcons();
    } catch (_) {}
  }


  // —— Splash + Login Master ——
  const splash = document.getElementById('splash');
  const login = document.getElementById('login');
  const osEl = document.getElementById('os');

  function showDesktop() {
    splash?.classList.add('hidden');
    login?.classList.add('hidden');
    osEl?.classList.remove('hidden');
    renderIcons();
    syncInstalledAppIcons().then(() => restoreWindowsSession());
  }

  function showLogin() {
    splash?.classList.add('hidden');
    login?.classList.remove('hidden');
    osEl?.classList.add('hidden');
  }

  async function tryLogin() {
    const user = document.getElementById('login-user')?.value || 'Master';
    const pass = document.getElementById('login-pass')?.value || '';
    const err = document.getElementById('login-err');
    try {
      const j = await api('/v1/auth/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ name: user, pass }),
      });
      if (!j.ok && !j.token) throw new Error('Credenciales inválidas');
      if (j.token) localStorage.setItem('alset_token', j.token);
      localStorage.setItem(SESSION_AUTH, JSON.stringify({ user, role: j.role || 'master', at: Date.now() }));
      showDesktop();
    } catch (e) {
      if (err) err.textContent = e.message || 'Error de acceso';
    }
  }

  document.getElementById('login-btn')?.addEventListener('click', tryLogin);
  document.getElementById('login-pass')?.addEventListener('keydown', (e) => {
    if (e.key === 'Enter') tryLogin();
  });

  // Splash then login or restore session
  setTimeout(() => {
    let sess = null;
    try { sess = JSON.parse(localStorage.getItem(SESSION_AUTH) || 'null'); } catch (_) {}
    // require login every boot for security on shared machine — only skip if same session < 8h
    const fresh = sess && sess.at && (Date.now() - sess.at) < 8 * 3600 * 1000;
    if (fresh && sess.user) {
      showDesktop();
    } else {
      showLogin();
    }
  }, 1600);

})();

