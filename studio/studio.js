(function(){
  const $ = (s)=>document.querySelector(s);
  const aos = $('#aos');
  const consoleEl = $('#console');
  const status = $('#status');
  const previewRoot = $('#previewRoot');
  const device = $('#device');
  const stateView = $('#state-view');
  const appName = $('#app-name');
  let deviceMode = 'mobile';
  let surfaces = {};
  let surfaceOID = '';
  let liveState = {};
  let sessionApp = '';

  const EXAMPLES = {
    'Cafetería mostrador': `(flow cafeteria-mostrador
  (version "1.0.0")
  (desc "Stock + venta + UI real")
  (orge stock
    (kind ORGES)
    (cap "stock.consultar" "stock.actualizar")
    (provides "stock.consulta" "1.0")
    (state (stock.unidades 10) (stock.umbral 3))
    (rule alerta (when (lte stock.unidades 3)) (then stock.alerta true)))
  (orge venta
    (kind ORGES)
    (cap "venta.registrar")
    (requires "stock.consulta" "1.0")
    (state (venta.habilitada false))
    (rule listo (when (gt stock.unidades 0)) (then venta.habilitada true) (indeterminate false)))
  (orge ui-mostrador
    (kind UI)
    (ui (column
      (label "Cafeteria")
      (badge (bind stock.unidades))
      (row
        (button "+" (emit stock.inc))
        (button "-" (emit stock.dec))
        (button "Vender" (emit venta.registrar))))))
  (assemble stock venta ui-mostrador)
  (surface ui-mostrador)
  (test stock-ini (eq stock.unidades 10))
  (deploy web (devices mobile tablet desktop) (out ./dist/cafeteria-mostrador))
)`,
    'Contador UI': `(flow contador
  (version "1.0.0")
  (orge counter
    (kind ORGES)
    (state (counter.unidades 0))
    (cap "counter.tick"))
  (orge ui-counter
    (kind UI)
    (ui (column
      (label "Contador")
      (badge (bind counter.unidades))
      (row
        (button "+" (emit counter.inc))
        (button "-" (emit counter.dec))
        (button "Reset" (emit counter.unidades=0))))))
  (assemble counter ui-counter)
  (surface ui-counter)
  (test zero (eq counter.unidades 0))
  (deploy web (devices mobile tablet desktop))
)`
  };

  const TOOLS = {
    orges: [
      {id:'orge', label:'ORGE dominio', snip:`\n  (orge nuevo\n    (kind ORGES)\n    (cap "x.hacer")\n    (state (x.v 1))\n    (rule r1 (when (gt x.v 0)) (then x.ok true)))\n`},
      {id:'provides', label:'provides contrato', snip:`\n    (provides "servicio.x" "1.0")\n`},
      {id:'requires', label:'requires contrato', snip:`\n    (requires "servicio.x" "1.0")\n`},
      {id:'bind', label:'BIND LEGO', snip:`\n  (bind consumidor proveedor servicio.x 1.0)\n`},
      {id:'assemble', label:'assemble', snip:`\n  (assemble org1 org2 ui1)\n`},
      {id:'test', label:'TEST', snip:`\n  (test t1 (eq x.v 1))\n`},
      {id:'deploy', label:'DEPLOY prod', snip:`\n  (deploy web (devices mobile tablet desktop) (out ./dist/app))\n`}
    ],
    ui: [
      {id:'ui-orge', label:'ORGE UI', snip:`\n  (orge ui-main\n    (kind UI)\n    (ui (column\n      (label "Titulo")\n      (badge (bind x.v))\n      (button "OK" (emit x.ping)))))\n  (surface ui-main)\n`},
      {id:'column', label:'column', snip:`(column (label "A") (button "B" (emit x.ok)))`},
      {id:'row', label:'row', snip:`(row (button "+" (emit x.inc)) (button "-" (emit x.dec)))`},
      {id:'label', label:'label', snip:`(label "Texto")`},
      {id:'button', label:'button + emit', snip:`(button "Accion" (emit x.hacer))`},
      {id:'badge', label:'badge bind', snip:`(badge (bind stock.unidades))`},
      {id:'field', label:'field', snip:`(field "entrada")`}
    ],
    actions: [
      {id:'inc', label:'emit *.inc', snip:`(emit stock.inc)`},
      {id:'dec', label:'emit *.dec', snip:`(emit stock.dec)`},
      {id:'venta', label:'emit venta.registrar', snip:`(emit venta.registrar)`},
      {id:'set', label:'emit campo=valor', snip:`(emit counter.unidades=0)`}
    ],
    templates: [
      {id:'tpl-cafe', label:'Plantilla cafetería', load:'Cafetería mostrador'},
      {id:'tpl-cnt', label:'Plantilla contador', load:'Contador UI'}
    ]
  };

  function log(msg, isErr){
    const line = '['+new Date().toISOString().slice(11,19)+'] '+msg+'\n';
    consoleEl.textContent = line + consoleEl.textContent;
    status.textContent = isErr ? 'error' : 'ok';
    status.className = 'badge '+(isErr?'err':'ok');
    if(isErr){ const e=$('#error-box'); e.textContent=msg; e.classList.remove('hidden'); }
  }
  function save(){ localStorage.setItem('alset.orges.aos', aos.value); localStorage.setItem('alset.orges.name', appName.value); }
  aos.value = localStorage.getItem('alset.orges.aos') || EXAMPLES['Cafetería mostrador'];
  appName.value = localStorage.getItem('alset.orges.name') || 'cafeteria-mostrador';

  function fillToolbox(el, items){
    el.innerHTML = '';
    items.forEach(it=>{
      const d=document.createElement('div');
      d.className='brick'; d.draggable=true; d.textContent=it.label;
      d.dataset.id=it.id;
      d.addEventListener('dragstart', e=>{
        e.dataTransfer.setData('text/snip', it.snip||'');
        e.dataTransfer.setData('text/load', it.load||'');
        e.dataTransfer.setData('text/label', it.label);
      });
      d.onclick = ()=>{
        if(it.load && EXAMPLES[it.load]){ aos.value = EXAMPLES[it.load]; save(); log('Plantilla: '+it.load); return; }
        if(it.snip){ aos.value += (aos.value.endsWith('\n')?'':'\n') + it.snip; save(); log('Insertado: '+it.label); }
      };
      el.appendChild(d);
    });
  }
  fillToolbox($('#toolbox-orges'), TOOLS.orges);
  fillToolbox($('#toolbox-ui'), TOOLS.ui);
  fillToolbox($('#toolbox-actions'), TOOLS.actions);
  fillToolbox($('#toolbox-templates'), TOOLS.templates);

  const canvas = $('#canvas');
  canvas.addEventListener('dragover', e=>e.preventDefault());
  canvas.addEventListener('drop', e=>{
    e.preventDefault();
    const load = e.dataTransfer.getData('text/load');
    const snip = e.dataTransfer.getData('text/snip');
    const label = e.dataTransfer.getData('text/label')||'pieza';
    if(load && EXAMPLES[load]){ aos.value = EXAMPLES[load]; }
    else if(snip){ aos.value += (aos.value.endsWith('\n')?'':'\n') + snip; }
    const chip=document.createElement('span'); chip.className='chip'; chip.textContent=label; canvas.appendChild(chip);
    save(); log('LEGO: '+label);
  });

  // Panel toggles
  document.querySelectorAll('.panel-tog').forEach(btn=>{
    btn.onclick = ()=>{
      const id = btn.dataset.panel;
      const pane = document.getElementById(id==='preview'?'pane-preview': id==='bottom'?'pane-bottom': id==='left'?'pane-left':'pane-right');
      if(!pane) return;
      pane.classList.toggle('collapsed');
      btn.classList.toggle('active');
    };
  });
  document.querySelectorAll('[data-close]').forEach(btn=>{
    btn.onclick = ()=>{
      const map={left:'pane-left',right:'pane-right',bottom:'pane-bottom',preview:'pane-preview'};
      const pane=document.getElementById(map[btn.dataset.close]);
      if(pane) pane.classList.add('collapsed');
      document.querySelector('.panel-tog[data-panel="'+btn.dataset.close+'"]')?.classList.remove('active');
    };
  });

  document.querySelectorAll('.dev').forEach(btn=>{
    btn.onclick = ()=>{
      document.querySelectorAll('.dev').forEach(x=>x.classList.remove('on'));
      btn.classList.add('on');
      deviceMode = btn.dataset.dev;
      device.className = 'device '+deviceMode;
      $('#devLabel').textContent = deviceMode;
    };
  });

  $('#btn-help').onclick = ()=>$('#help-drawer').classList.remove('hidden');
  $('#btn-close-help').onclick = ()=>$('#help-drawer').classList.add('hidden');
  $('#btn-examples').onclick = ()=>{
    const list=$('#examples-list'); list.innerHTML='';
    Object.keys(EXAMPLES).forEach(name=>{
      const b=document.createElement('button'); b.className='btn btn-ghost'; b.style.display='block'; b.style.width='100%'; b.style.marginBottom='6px';
      b.textContent=name; b.onclick=()=>{ aos.value=EXAMPLES[name]; appName.value=name.toLowerCase().replace(/\s+/g,'-'); save(); log('Ejemplo: '+name); };
      list.appendChild(b);
    });
    $('#examples-drawer').classList.remove('hidden');
  };
  $('#btn-close-examples').onclick = ()=>$('#examples-drawer').classList.add('hidden');

  async function api(path, body){
    const r = await fetch(path, {method:'POST', headers:{'Content-Type':'application/json'}, body: JSON.stringify(body||{})});
    const t = await r.text();
    try { return JSON.parse(t); } catch { return {ok:false, error:t, status:r.status}; }
  }

  function paint(tree){
    previewRoot.innerHTML='';
    if(!tree){ previewRoot.innerHTML='<p class="muted">Sin superficie UI</p>'; return; }
    previewRoot.appendChild(render(tree));
  }
  function val(bind){
    if(!bind) return '·';
    const v = liveState[bind];
    return v==null ? bind : v;
  }
  function render(n){
    if(!n) return document.createTextNode('');
    const t=(n.type||'').toLowerCase();
    if(t==='column'||t==='panel'){ const d=document.createElement('div'); d.className='col'; (n.kids||[]).forEach(k=>d.appendChild(render(k))); return d; }
    if(t==='row'){ const d=document.createElement('div'); d.className='row'; (n.kids||[]).forEach(k=>d.appendChild(render(k))); return d; }
    if(t==='label'){ const s=document.createElement('div'); s.textContent=n.text||''; return s; }
    if(t==='badge'){ const s=document.createElement('span'); s.className='badge-ui'; s.textContent=val(n.bind); return s; }
    if(t==='button'){
      const b=document.createElement('button'); b.textContent=n.text||'OK';
      b.onclick = async ()=>{
        const emit = n.emit||'';
        log('emit '+emit);
        const res = await api('/api/flow/action', {emit, app: sessionApp||appName.value});
        if(res.state){ liveState=res.state; stateView.textContent=JSON.stringify(liveState,null,2); }
        if(res.ok===false) log(res.error||'acción falló', true);
        else log('estado actualizado');
        paint(surfaces[surfaceOID]||Object.values(surfaces)[0]);
      };
      return b;
    }
    if(t==='field'){ const i=document.createElement('input'); i.className='rt-input'; i.placeholder=n.text||''; return i; }
    return document.createTextNode(n.type||'');
  }

  $('#btn-apply').onclick = async ()=>{
    save(); $('#error-box').classList.add('hidden');
    log('Aplicando flow…');
    const name = appName.value.trim()||'mi-app';
    // rewrite flow name in source lightly
    let src = aos.value;
    src = src.replace(/\(flow\s+[^\s)]+/, '(flow '+name);
    const res = await api('/api/flow/apply', {source: src, app: name});
    if(res.error){ log(res.error, true); return; }
    surfaces = res.surfaces||{};
    surfaceOID = res.surface||Object.keys(surfaces)[0]||'';
    sessionApp = res.app||name;
    if(res.state) liveState=res.state;
    stateView.textContent = JSON.stringify(liveState,null,2);
    log('Flow OK · '+ (res.organismos||0)+' organismos · surface '+surfaceOID);
    paint(surfaces[surfaceOID]||Object.values(surfaces)[0]);
    refreshApps();
  };

  $('#btn-run-preview').onclick = ()=>{
    paint(surfaces[surfaceOID]||Object.values(surfaces)[0]);
    log('Preview montado ('+deviceMode+')');
  };

  $('#btn-test').onclick = async ()=>{
    save();
    const res = await api('/api/flow/test', {source: aos.value, app: appName.value});
    if(res.ok) log('Tests OK');
    else log('Tests fallidos: '+JSON.stringify(res.errors||res.error), true);
  };

  $('#btn-deploy').onclick = async ()=>{
    save();
    log('Desplegando producción…');
    const res = await api('/api/flow/deploy', {source: aos.value, app: appName.value.trim()||'mi-app', production: true});
    if(res.error){ log(res.error, true); return; }
    log('Producción: '+(res.url||res.out));
    if(res.url) log('Abrir '+location.origin+res.url);
    refreshApps();
  };

  $('#btn-apply-prop').onclick = ()=>{
    const emit = $('#prop-emit').value.trim();
    const bind = $('#prop-bind').value.trim();
    if(emit) aos.value += `\n  ;; prop emit ${emit}\n`;
    if(bind) aos.value += `\n  ;; prop bind ${bind}\n`;
    save(); log('Propiedades anotadas en .aos (ajusta el botón/badge manualmente)');
  };

  async function refreshApps(){
    try{
      const r = await fetch('/api/apps');
      const j = await r.json();
      const box = $('#apps-list'); box.innerHTML='';
      (j.apps||[]).forEach(a=>{
        const link=document.createElement('a');
        link.href='/apps/'+encodeURIComponent(a)+'/';
        link.target='_blank'; link.textContent=a+' ↗';
        box.appendChild(link);
      });
    }catch(e){}
  }
  refreshApps();
  log('Alset Studio ORGES listo · Toolbox · Preview · Deploy producción');
})();
