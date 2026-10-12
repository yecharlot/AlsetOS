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
  let surfaces = {}, surfaceOID = '', liveState = {}, sessionApp = '';

  function flowShell(name, desc, body){
    return `(flow ${name}
  (version "1.0.0")
  (desc ${JSON.stringify(desc)})
${body}
  (deploy web (devices mobile tablet desktop))
)`;
  }

  const EXAMPLES = {
    'cafeteria-mostrador': flowShell('cafeteria-mostrador','Stock + venta',`
  (orge stock (kind ORGES) (cap "stock.consultar") (provides "stock.consulta" "1.0")
    (state (stock.unidades 10)) (rule a (when (lte stock.unidades 3)) (then stock.alerta true)))
  (orge venta (kind ORGES) (requires "stock.consulta" "1.0") (state (venta.habilitada false))
    (rule l (when (gt stock.unidades 0)) (then venta.habilitada true) (indeterminate false)))
  (orge ui-main (kind UI) (ui (column
    (splash "Cafeteria")
    (badge (bind stock.unidades))
    (row (button "+" (emit stock.inc)) (button "-" (emit stock.dec)) (button "Vender" (emit venta.registrar))))))
  (assemble stock venta ui-main) (surface ui-main)
  (test t1 (eq stock.unidades 10))`),

    'tienda-online': flowShell('tienda-online','Tienda online con catalogo',`
  (orge catalogo (kind ORGES) (state (catalogo.items "Cafe|Te|Pan") (cart.count 0)))
  (orge ui-shop (kind UI) (ui (column
    (splash "Tienda Alset")
    (gradient "Ofertas de la semana" (style "animated"))
    (badge (bind cart.count))
    (label "Productos")
    (row (button "Cafe" (emit cart.count=1)) (button "Te" (emit cart.count=2)) (button "Pan" (emit cart.count=3)))
    (button "Checkout" (emit checkout.start)))))
  (assemble catalogo ui-shop) (surface ui-shop)`),

    'pagos-checkout': flowShell('pagos-checkout','App de pagos',`
  (orge pagos (kind ORGES) (state (pago.monto 0) (pago.estado "pendiente")))
  (orge ui-pay (kind UI) (ui (column
    (label "Checkout seguro")
    (badge (bind pago.estado))
    (row (button "10 CUP" (emit pago.monto=10)) (button "50 CUP" (emit pago.monto=50)))
    (button "Pagar" (emit pago.estado=pagado))
    (button "Reembolsar" (emit pago.estado=reembolsado)))))
  (assemble pagos ui-pay) (surface ui-pay)`),

    'ecommerce-full': flowShell('ecommerce-full','Comercio electronico tabs',`
  (orge shop (kind ORGES) (state (shop.tab 0) (shop.cart 0)))
  (orge ui-ecom (kind UI) (ui (column
    (splash "E-Commerce")
    (bottom-tabs
      (tab "Inicio" (column (label "Destacados") (button "Add" (emit shop.cart=1))))
      (tab "Cart" (column (label "Carrito") (badge (bind shop.cart))))
      (tab "Perfil" (column (label "Mi cuenta")))))))
  (assemble shop ui-ecom) (surface ui-ecom)`),

    'billetera-virtual': flowShell('billetera-virtual','Billetera virtual',`
  (orge wallet (kind ORGES) (state (wallet.saldo 1000) (wallet.moneda "CUP")))
  (orge ui-wallet (kind UI) (ui (column
    (blob)
    (label "Saldo")
    (badge (bind wallet.saldo))
    (row (button "Enviar" (emit wallet.saldo=900)) (button "Recibir" (emit wallet.saldo=1100)))
    (qr (bind wallet.saldo)))))
  (assemble wallet ui-wallet) (surface ui-wallet)`),

    'landing-page': flowShell('landing-page','Landing marketing',`
  (orge land (kind ORGES) (state (lead.email "")))
  (orge ui-land (kind UI) (ui (column
    (splash "Lanza tu producto")
    (gradient "Alset · organismos digitales")
    (image (src "https://picsum.photos/640/240") (style "rounded"))
    (label "Une ORGES como LEGO")
    (button "Empezar" (emit lead.email=ok)))))
  (assemble land ui-land) (surface ui-land)`),

    'blog-notas': flowShell('blog-notas','Blog y notas',`
  (orge notes (kind ORGES) (state (notes.list "Hola|Idea 2") (notes.draft "")))
  (orge ui-blog (kind UI) (ui (column
    (label "Mis notas")
    (doclist (bind notes.list))
    (field "Nueva nota")
    (file "Adjuntar" (accept "image/*,.txt") (emit file.note))
    (button "Publicar" (emit notes.list=Hola|Idea 2|Nueva)))))
  (assemble notes ui-blog) (surface ui-blog)`),

    'generacion-contenido': flowShell('generacion-contenido','Generacion de contenido',`
  (orge gen (kind ORGES) (state (gen.out "Escribe un prompt")))
  (orge ui-gen (kind UI) (ui (column
    (label "Studio creativo")
    (field "Prompt")
    (button "Generar" (emit gen.out=Borrador listo))
    (badge (bind gen.out)))))
  (assemble gen ui-gen) (surface ui-gen)`),

    'streaming-vivo': flowShell('streaming-vivo','Streaming en vivo',`
  (orge stream (kind ORGES) (state (stream.live false) (stream.viewers 0)))
  (orge ui-stream (kind UI) (ui (column
    (splash "Live")
    (badge (bind stream.viewers))
    (row (button "Go Live" (emit stream.live=true)) (button "End" (emit stream.live=false)))
    (button "+ Viewer" (emit stream.viewers=1)))))
  (assemble stream ui-stream) (surface ui-stream)`),

    'medico-citas': flowShell('medico-citas','Aplicacion medica citas',`
  (orge clinic (kind ORGES) (state (cita.paciente "") (cita.estado "libre")))
  (orge ui-med (kind UI) (ui (column
    (drawer "Menu"
      (button "Citas" (emit cita.estado=libre))
      (button "Historial" (emit cita.estado=historial)))
    (label "Consulta")
    (badge (bind cita.estado))
    (button "Reservar" (emit cita.estado=reservada)))))
  (assemble clinic ui-med) (surface ui-med)`),

    'finanzas-pyme': flowShell('finanzas-pyme','Finanzas y caja',`
  (orge fin (kind ORGES) (state (fin.caja 500) (fin.gasto 0)))
  (orge ui-fin (kind UI) (ui (column
    (label "Caja del dia")
    (badge (bind fin.caja))
    (row (button "Ingreso +50" (emit fin.caja=550)) (button "Gasto -20" (emit fin.caja=480)))
    (tabs
      (tab "Resumen" (column (label "Hoy OK")))
      (tab "Movimientos" (column (label "Ver libro")))))))
  (assemble fin ui-fin) (surface ui-fin)`),

    'logistica-rutas': flowShell('logistica-rutas','Seguimiento logistico',`
  (orge logi (kind ORGES) (state (ruta.estado "en_transito") (ruta.paquete "PKG-1")))
  (orge ui-logi (kind UI) (ui (column
    (label "Tracking")
    (badge (bind ruta.estado))
    (qr (bind ruta.paquete))
    (row (button "En ruta" (emit ruta.estado=en_transito)) (button "Entregado" (emit ruta.estado=entregado))))))
  (assemble logi ui-logi) (surface ui-logi)`),

    'lector-qr': flowShell('lector-qr','Lector QR y codigo',`
  (orge scan (kind ORGES) (state (qr.payload "alset://org/demo")))
  (orge ui-qr (kind UI) (ui (column
    (splash "Escanear")
    (qr (bind qr.payload))
    (button "Simular scan" (emit qr.payload=alset://org/scanned))
    (badge (bind qr.payload)))))
  (assemble scan ui-qr) (surface ui-qr)`),

    'gestor-documentos': flowShell('gestor-documentos','Gestor de documentos',`
  (orge docs (kind ORGES) (state (docs.list "Contrato.pdf|Factura.pdf")))
  (orge ui-docs (kind UI) (ui (column
    (label "Documentos")
    (doclist (bind docs.list))
    (file "Subir desde disco" (accept ".pdf,image/*") (emit file.doc))
    (button "Nuevo" (emit docs.list=Contrato.pdf|Factura.pdf|Nota.txt)))))
  (assemble docs ui-docs) (surface ui-docs)`),

    'notas-rapidas': flowShell('notas-rapidas','App de notas',`
  (orge memo (kind ORGES) (state (memo.text "Hola Alset")))
  (orge ui-memo (kind UI) (ui (column
    (label "Notas")
    (badge (bind memo.text))
    (field "Escribe")
    (button "Guardar" (emit memo.text=Nota guardada)))))
  (assemble memo ui-memo) (surface ui-memo)`),

    'ui-kit-navegacion': flowShell('ui-kit-navegacion','Drawer stack tabs animados',`
  (orge kit (kind ORGES) (state (kit.n 1)))
  (orge ui-kit (kind UI) (ui (column
    (splash "UI Kit")
    (blob)
    (drawer "Menu lateral"
      (button "A" (emit kit.n=1))
      (button "B" (emit kit.n=2)))
    (top-tabs
      (tab "Uno" (column (label "Stack 1") (badge (bind kit.n))))
      (tab "Dos" (column (label "Stack 2"))))
    (bottom-tabs
      (tab "Home" (column (label "Home")))
      (tab "More" (column (label "More")))))))
  (assemble kit ui-kit) (surface ui-kit)`)
  };

  const TOOLS = {
    orges: [
      {label:'ORGE dominio', snip:`\n  (orge nuevo (kind ORGES) (state (x.v 1)) (cap "x.hacer"))\n`},
      {label:'provides', snip:`\n    (provides "svc.x" "1.0")\n`},
      {label:'requires', snip:`\n    (requires "svc.x" "1.0")\n`},
      {label:'BIND', snip:`\n  (bind consumidor proveedor svc.x 1.0)\n`},
      {label:'TEST', snip:`\n  (test t1 (eq x.v 1))\n`},
      {label:'DEPLOY', snip:`\n  (deploy web (devices mobile tablet desktop))\n`}
    ],
    ui: [
      {label:'ORGE UI', snip:`\n  (orge ui-main (kind UI) (ui (column (label "Hola") (button "OK" (emit x.ok)))))\n  (surface ui-main)\n`},
      {label:'splash', snip:`(splash "Bienvenido")`},
      {label:'blob', snip:`(blob)`},
      {label:'gradient', snip:`(gradient "Titulo")`},
      {label:'drawer', snip:`(drawer "Menu" (button "A" (emit x.a)))`},
      {label:'tabs', snip:`(tabs (tab "Uno" (label "A")) (tab "Dos" (label "B")))`},
      {label:'bottom-tabs', snip:`(bottom-tabs (tab "Home" (label "H")) (tab "More" (label "M")))`},
      {label:'image', snip:`(image (src "https://picsum.photos/400/200"))`},
      {label:'file disco', snip:`(file "Elegir archivo" (accept "image/*,.pdf") (emit file.pick))`},
      {label:'qr', snip:`(qr (bind x.code))`},
      {label:'doclist', snip:`(doclist (bind docs.list))`},
      {label:'button emit', snip:`(button "Accion" (emit stock.inc))`},
      {label:'badge', snip:`(badge (bind stock.unidades))`}
    ]
  };

  function log(msg, isErr){
    consoleEl.textContent = '['+new Date().toISOString().slice(11,19)+'] '+msg+'\n'+consoleEl.textContent;
    status.textContent = isErr?'error':'ok';
    status.className = 'badge '+(isErr?'err':'ok');
    if(isErr){ const e=$('#error-box'); e.textContent=msg; e.classList.remove('hidden'); }
  }
  function save(){ localStorage.setItem('alset.orges.aos', aos.value); localStorage.setItem('alset.orges.name', appName.value); }

  // default first example
  const firstKey = 'cafeteria-mostrador';
  aos.value = localStorage.getItem('alset.orges.aos') || EXAMPLES[firstKey];
  appName.value = localStorage.getItem('alset.orges.name') || firstKey;

  function fillToolbox(el, items){
    el.innerHTML='';
    items.forEach(it=>{
      const d=document.createElement('div'); d.className='brick'; d.draggable=true; d.textContent=it.label;
      d.addEventListener('dragstart', e=>{ e.dataTransfer.setData('text/snip', it.snip||''); e.dataTransfer.setData('text/label', it.label); });
      d.onclick=()=>{ if(it.snip){ aos.value+=(aos.value.endsWith('\n')?'':'\n')+it.snip; save(); log('Insertado '+it.label); } };
      el.appendChild(d);
    });
  }
  fillToolbox($('#toolbox-orges'), TOOLS.orges);
  fillToolbox($('#toolbox-ui'), TOOLS.ui);
  // templates list = all examples
  const tplBox = $('#toolbox-templates');
  tplBox.innerHTML='';
  Object.keys(EXAMPLES).forEach(k=>{
    const d=document.createElement('div'); d.className='brick'; d.textContent=k;
    d.onclick=()=>{ aos.value=EXAMPLES[k]; appName.value=k; save(); log('Plantilla '+k); };
    d.draggable=true;
    d.addEventListener('dragstart', e=>{ e.dataTransfer.setData('text/tpl', k); e.dataTransfer.setData('text/label', k); });
    tplBox.appendChild(d);
  });

  const canvas=$('#canvas');
  canvas.addEventListener('dragover', e=>e.preventDefault());
  canvas.addEventListener('drop', e=>{
    e.preventDefault();
    const tpl=e.dataTransfer.getData('text/tpl');
    const snip=e.dataTransfer.getData('text/snip');
    const label=e.dataTransfer.getData('text/label')||'pieza';
    if(tpl && EXAMPLES[tpl]){ aos.value=EXAMPLES[tpl]; appName.value=tpl; }
    else if(snip) aos.value+=(aos.value.endsWith('\n')?'':'\n')+snip;
    const chip=document.createElement('span'); chip.className='chip'; chip.textContent=label; canvas.appendChild(chip);
    save(); log('LEGO '+label);
  });

  // --- Resizable panels (real) ---
  function enableSplit(selector, orient, onMove){
    const el = document.querySelector(selector);
    if(!el) return;
    el.addEventListener('mousedown', (ev)=>{
      ev.preventDefault();
      const startX=ev.clientX, startY=ev.clientY;
      const move=(e)=>{
        onMove(e.clientX-startX, e.clientY-startY, e);
      };
      const up=()=>{ window.removeEventListener('mousemove', move); window.removeEventListener('mouseup', up); document.body.style.cursor=''; };
      window.addEventListener('mousemove', move);
      window.addEventListener('mouseup', up);
      document.body.style.cursor = orient==='v'?'col-resize':'row-resize';
    });
  }
  const leftPane=$('#pane-left'), rightPane=$('#pane-right'), bottomPane=$('#pane-bottom'), previewPane=$('#pane-preview');
  let leftW=leftPane.offsetWidth||240, rightW=rightPane.offsetWidth||280, bottomH=140, previewH=280;
  enableSplit('.splitter.v[data-split="left"]','v',(dx)=>{ leftW=Math.max(160, Math.min(420, leftW+dx)); leftPane.style.width=leftW+'px'; leftPane.style.flex='0 0 auto'; });
  // right splitter is before right pane - need second splitter handler: use data-split right
  document.querySelectorAll('.splitter.v').forEach((sp, idx)=>{
    if(sp.getAttribute('data-split')==='right'){
      let base=rightW;
      sp.addEventListener('mousedown', (ev)=>{
        ev.preventDefault();
        const sx=ev.clientX; base=rightPane.offsetWidth;
        const move=(e)=>{ const w=Math.max(180, Math.min(480, base-(e.clientX-sx))); rightPane.style.width=w+'px'; rightPane.style.flex='0 0 auto'; };
        const up=()=>{ window.removeEventListener('mousemove',move); window.removeEventListener('mouseup',up); };
        window.addEventListener('mousemove',move); window.addEventListener('mouseup',up);
      });
    }
  });
  enableSplit('.splitter.h[data-split="bottom"]','h',(dx,dy)=>{
    // This splitter is above bottom - actually structure may need fix
  });
  // Preview height drag: top border of preview
  previewPane.style.flex='0 0 auto';
  previewPane.style.height=previewH+'px';
  bottomPane.style.flex='0 0 auto';
  bottomPane.style.height=bottomH+'px';
  const prevHead = previewPane.querySelector('.pane-head');
  if(prevHead){
    prevHead.style.cursor='ns-resize';
    prevHead.title='Arrastra para redimensionar preview';
    prevHead.addEventListener('mousedown', (ev)=>{
      if(ev.target.closest('.icon-btn')) return;
      ev.preventDefault();
      const sy=ev.clientY; const base=previewPane.offsetHeight;
      const move=(e)=>{ const h=Math.max(120, Math.min(window.innerHeight*0.7, base-(e.clientY-sy))); previewPane.style.height=h+'px'; };
      const up=()=>{ window.removeEventListener('mousemove',move); window.removeEventListener('mouseup',up); };
      window.addEventListener('mousemove',move); window.addEventListener('mouseup',up);
    });
  }
  const botHead = bottomPane.querySelector('.pane-head');
  if(botHead){
    botHead.style.cursor='ns-resize';
    botHead.addEventListener('mousedown', (ev)=>{
      if(ev.target.closest('.icon-btn')) return;
      ev.preventDefault();
      const sy=ev.clientY; const base=bottomPane.offsetHeight;
      const move=(e)=>{ const h=Math.max(80, Math.min(400, base-(e.clientY-sy))); bottomPane.style.height=h+'px'; };
      const up=()=>{ window.removeEventListener('mousemove',move); window.removeEventListener('mouseup',up); };
      window.addEventListener('mousemove',move); window.addEventListener('mouseup',up);
    });
  }

  document.querySelectorAll('.panel-tog').forEach(btn=>{
    btn.onclick=()=>{
      const id=btn.dataset.panel;
      const pane=document.getElementById(id==='preview'?'pane-preview':id==='bottom'?'pane-bottom':id==='left'?'pane-left':'pane-right');
      if(!pane) return;
      pane.classList.toggle('collapsed');
      btn.classList.toggle('active');
    };
  });
  document.querySelectorAll('[data-close]').forEach(btn=>{
    btn.onclick=()=>{
      const map={left:'pane-left',right:'pane-right',bottom:'pane-bottom',preview:'pane-preview'};
      document.getElementById(map[btn.dataset.close])?.classList.add('collapsed');
      document.querySelector('.panel-tog[data-panel="'+btn.dataset.close+'"]')?.classList.remove('active');
    };
  });
  document.querySelectorAll('.dev').forEach(btn=>{
    btn.onclick=()=>{
      document.querySelectorAll('.dev').forEach(x=>x.classList.remove('on'));
      btn.classList.add('on'); deviceMode=btn.dataset.dev;
      device.className='device '+deviceMode; $('#devLabel').textContent=deviceMode;
    };
  });

  $('#btn-help').onclick=()=>$('#help-drawer').classList.remove('hidden');
  $('#btn-close-help').onclick=()=>$('#help-drawer').classList.add('hidden');
  $('#btn-examples').onclick=()=>{
    const list=$('#examples-list'); list.innerHTML='';
    Object.keys(EXAMPLES).forEach(name=>{
      const b=document.createElement('button'); b.className='btn btn-ghost'; b.style.cssText='display:block;width:100%;margin-bottom:6px;text-align:left';
      b.textContent=name;
      b.onclick=()=>{ aos.value=EXAMPLES[name]; appName.value=name; save(); log('Ejemplo '+name); $('#examples-drawer').classList.add('hidden'); };
      list.appendChild(b);
    });
    $('#examples-drawer').classList.remove('hidden');
  };
  $('#btn-close-examples').onclick=()=>$('#examples-drawer').classList.add('hidden');

  async function api(path, body){
    const r=await fetch(path,{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify(body||{})});
    const t=await r.text();
    try{return JSON.parse(t);}catch{return {ok:false,error:t};}
  }

  function paint(tree){
    previewRoot.innerHTML='';
    if(!tree){ previewRoot.innerHTML='<p class="muted">Sin UI</p>'; return; }
    previewRoot.appendChild(render(tree));
  }
  function val(bind){ if(!bind)return '·'; const v=liveState[bind]; return v==null?bind:v; }
  function render(n){
    if(!n) return document.createTextNode('');
    const t=(n.type||'').toLowerCase();
    if(t==='column'||t==='panel'){ const d=document.createElement('div'); d.className='col'; (n.kids||[]).forEach(k=>d.appendChild(render(k))); return d; }
    if(t==='row'){ const d=document.createElement('div'); d.className='row'; (n.kids||[]).forEach(k=>d.appendChild(render(k))); return d; }
    if(t==='label'){ const s=document.createElement('div'); s.textContent=n.text||''; return s; }
    if(t==='badge'){ const s=document.createElement('span'); s.className='badge-ui'; s.textContent=val(n.bind); return s; }
    if(t==='button'){ /* beautiful */ const b=document.createElement('button'); b.textContent=n.text||'OK';
      b.onclick=async()=>{ const emit=n.emit||''; log('emit '+emit);
        const res=await api('/api/flow/action',{emit, app: sessionApp||appName.value});
        if(res.state){ liveState=res.state; stateView.textContent=JSON.stringify(liveState,null,2); }
        if(res.ok===false) log(res.error||'fail', true); paint(surfaces[surfaceOID]||Object.values(surfaces)[0]);
      }; return b; }
    if(t==='splash'){ el=document.createElement('div'); el.textContent=n.text||'Splash'; el.className='rt-splash'; }
    if(t==='blob'){ const d=document.createElement('div'); d.style.cssText='width:64px;height:64px;border-radius:40%;background:linear-gradient(135deg,#f4b400,#5b9cf5)'; return d; }
    if(t==='gradient'){ const d=document.createElement('div'); d.textContent=n.text||''; d.style.cssText='padding:12px;border-radius:10px;background:linear-gradient(135deg,#1a2332,#3d2a5a)'; return d; }
    if(t==='image'){ const i=document.createElement('img'); i.src=n.src||''; i.style.maxWidth='100%'; return i; }
    if(t==='qr'){ const d=document.createElement('code'); d.textContent=val(n.bind)||n.text||'QR'; return d; }
    if(t==='file'){ const i=document.createElement('input'); i.type='file'; return i; }
    if(t==='doclist'){ const ul=document.createElement('ul'); String(val(n.bind)).split('|').filter(Boolean).forEach(x=>{ const li=document.createElement('li'); li.textContent=x; ul.appendChild(li); }); return ul; }
    if(t==='drawer'||t==='tabs'||t==='top-tabs'||t==='bottom-tabs'||t==='stack'){
      const d=document.createElement('div'); d.className='col';
      const lab=document.createElement('div'); lab.className='muted'; lab.textContent=t+(n.text?': '+n.text:''); d.appendChild(lab);
      (n.kids||[]).forEach(k=>d.appendChild(render(k))); return d;
    }
    if(t==='tab'){ const d=document.createElement('div'); d.textContent='Tab '+(n.text||''); (n.kids||[]).forEach(k=>d.appendChild(render(k))); return d; }
    if(t==='field'){ const i=document.createElement('input'); i.className='rt-input'; i.placeholder=n.text||''; return i; }
    return document.createTextNode(n.type||'');
  }

  $('#btn-apply').onclick=async()=>{
    save(); $('#error-box').classList.add('hidden');
    const name=appName.value.trim()||'mi-app';
    let src=aos.value.replace(/\(flow\s+[^\s)]+/, '(flow '+name);
    log('Aplicando '+name+'…');
    const res=await api('/api/flow/apply',{source:src, app:name});
    if(res.error){ log(res.error,true); return; }
    surfaces=res.surfaces||{}; surfaceOID=res.surface||Object.keys(surfaces)[0]||'';
    sessionApp=res.app||name; appName.value=sessionApp;
    if(res.state) liveState=res.state;
    stateView.textContent=JSON.stringify(liveState,null,2);
    log('OK instancia '+sessionApp+' · orgs '+(res.organismos||0));
    paint(surfaces[surfaceOID]||Object.values(surfaces)[0]);
    refreshApps();
  };
  $('#btn-run-preview').onclick=()=>{ paint(surfaces[surfaceOID]||Object.values(surfaces)[0]); log('Preview '+deviceMode); };
  $('#btn-test').onclick=async()=>{
    const res=await api('/api/flow/test',{source:aos.value, app:appName.value});
    if(res.ok) log('Tests OK · '+res.app); else log('Tests: '+JSON.stringify(res.errors||res.error), true);
  };
  $('#btn-deploy').onclick=async()=>{
    const name=appName.value.trim()||'mi-app';
    let src=aos.value.replace(/\(flow\s+[^\s)]+/, '(flow '+name);
    log('Deploy producción instancia '+name+'…');
    const res=await api('/api/flow/deploy',{source:src, app:name, production:true});
    if(res.error){ log(res.error,true); return; }
    log('Aislada: '+res.app+' → '+res.url+(res.isolated?' (isolated)':''));
    sessionApp=res.app; refreshApps();
  };
  $('#btn-apply-prop')&&($('#btn-apply-prop').onclick=()=>log('Ajusta emit/bind en el .aos'));

  async function refreshApps(){
    try{
      const r=await fetch('/api/apps'); const j=await r.json();
      const box=$('#apps-list'); box.innerHTML='';
      (j.apps||[]).forEach(a=>{
        const link=document.createElement('a'); link.href='/apps/'+encodeURIComponent(a)+'/'; link.target='_blank';
        link.textContent=a+((j.live||[]).includes(a)?' ●':'')+' ↗'; box.appendChild(link);
      });
    }catch(e){}
  }
  refreshApps();
  

  // --- Editor visual en preview (adaptar a mobile/tablet/desktop) ---
  let editMode = false;
  const layoutStore = JSON.parse(localStorage.getItem('alset.orges.layouts')||'{}');
  function layoutKey(){ return (sessionApp||appName.value||'app')+'::'+deviceMode; }
  function saveLayouts(){ localStorage.setItem('alset.orges.layouts', JSON.stringify(layoutStore)); }

  // botón en top si existe, o crear
  (function(){
    let b = document.getElementById('btn-edit-layout');
    if(!b){
      b = document.createElement('button');
      b.type='button'; b.id='btn-edit-layout'; b.className='btn btn-ghost';
      b.textContent='Editar layout';
      const top = document.getElementById('topbar');
      if(top) top.appendChild(b);
    }
    b.onclick = ()=>{
      editMode = !editMode;
      b.classList.toggle('active', editMode);
      b.textContent = editMode ? 'Layout ON' : 'Editar layout';
      log(editMode ? 'Modo edición: arrastra y redimensiona elementos en el preview' : 'Modo edición off');
      paint(surfaces[surfaceOID]||Object.values(surfaces)[0]);
    };
  })();

  function applyLayout(el, path){
    const L = (layoutStore[layoutKey()]||{})[path];
    if(!L) return;
    el.style.position = 'absolute';
    if(L.left!=null) el.style.left = L.left+'px';
    if(L.top!=null) el.style.top = L.top+'px';
    if(L.width!=null) el.style.width = L.width+'px';
    if(L.height!=null) el.style.height = L.height+'px';
  }

  function makeEditable(el, path){
    if(!editMode) return;
    el.dataset.path = path;
    el.style.outline = '1px dashed rgba(244,180,0,.45)';
    el.style.cursor = 'move';
    el.style.position = el.style.position || 'relative';
    const handle = document.createElement('div');
    handle.style.cssText = 'position:absolute;right:0;bottom:0;width:12px;height:12px;background:#f4b400;cursor:nwse-resize;z-index:5';
    el.style.minHeight = '24px';
    if(getComputedStyle(el).position==='static') el.style.position='relative';
    el.appendChild(handle);
    applyLayout(el, path);

    let mode=null, sx, sy, sl, st, sw, sh;
    el.addEventListener('mousedown', (ev)=>{
      if(ev.target===handle) mode='resize'; else mode='drag';
      ev.preventDefault(); ev.stopPropagation();
      const r = el.getBoundingClientRect();
      const pr = previewRoot.getBoundingClientRect();
      sx=ev.clientX; sy=ev.clientY;
      sl=r.left-pr.left; st=r.top-pr.top; sw=r.width; sh=r.height;
      el.style.position='absolute';
      el.style.left=sl+'px'; el.style.top=st+'px';
      const move=(e)=>{
        if(mode==='drag'){
          el.style.left=(sl+e.clientX-sx)+'px';
          el.style.top=(st+e.clientY-sy)+'px';
        } else {
          el.style.width=Math.max(40, sw+e.clientX-sx)+'px';
          el.style.height=Math.max(24, sh+e.clientY-sy)+'px';
        }
      };
      const up=()=>{
        window.removeEventListener('mousemove', move);
        window.removeEventListener('mouseup', up);
        const key=layoutKey();
        layoutStore[key]=layoutStore[key]||{};
        layoutStore[key][path]={
          left: parseFloat(el.style.left)||0,
          top: parseFloat(el.style.top)||0,
          width: parseFloat(el.style.width)||el.offsetWidth,
          height: parseFloat(el.style.height)||el.offsetHeight
        };
        saveLayouts();
        log('layout '+deviceMode+' · '+path);
      };
      window.addEventListener('mousemove', move);
      window.addEventListener('mouseup', up);
    });
  }

  // Patch paint/render to attach paths — replace paint function body usage
  const _paint = paint;
  paint = function(tree){
    previewRoot.innerHTML='';
    previewRoot.style.position='relative';
    previewRoot.style.minHeight='200px';
    if(!tree){ previewRoot.innerHTML='<p class="muted">Sin UI</p>'; return; }
    const rootEl = renderEditable(tree, '0');
    previewRoot.appendChild(rootEl);
  };

  function renderEditable(n, path){
    if(!n) return document.createTextNode('');
    const t=(n.type||'').toLowerCase();
    let el;
    if(t==='column'||t==='panel'){ el=document.createElement('div'); el.className='col'; (n.kids||[]).forEach((k,i)=>el.appendChild(renderEditable(k, path+'.'+i))); }
    else if(t==='row'){ el=document.createElement('div'); el.className='row'; (n.kids||[]).forEach((k,i)=>el.appendChild(renderEditable(k, path+'.'+i))); }
    else if(t==='label'){ el=document.createElement('div'); el.textContent=n.text||''; }
    else if(t==='badge'){ el=document.createElement('span'); el.className='badge-ui'; el.textContent=val(n.bind); }
    else if(t==='button'){ /* beautiful */
      el=document.createElement('button'); el.textContent=n.text||'OK';
      el.onclick=async (ev)=>{
        if(editMode){ ev.preventDefault(); return; }
        const emit=n.emit||''; log('emit '+emit);
        const res=await api('/api/flow/action',{emit, app: sessionApp||appName.value});
        if(res.state){ liveState=res.state; stateView.textContent=JSON.stringify(liveState,null,2); }
        if(res.ok===false) log(res.error||'fail', true);
        paint(surfaces[surfaceOID]||Object.values(surfaces)[0]);
      };
    }
    else if(t==='splash'){ el=document.createElement('div'); el.textContent=n.text||'Splash'; el.className='rt-splash'; }
    else if(t==='blob'){ el=document.createElement('div'); el.style.cssText='width:64px;height:64px;border-radius:40%;background:linear-gradient(135deg,#f4b400,#5b9cf5)'; }
    else if(t==='gradient'){ el=document.createElement('div'); el.textContent=n.text||''; el.style.cssText='padding:12px;border-radius:10px;background:linear-gradient(135deg,#1a2332,#3d2a5a)'; }
    else if(t==='image'){ el=document.createElement('img'); el.src=n.src||''; el.style.maxWidth='100%'; }
    else if(t==='qr'){ el=document.createElement('code'); el.textContent=val(n.bind)||n.text||'QR'; }
    else if(t==='file'){ el=document.createElement('input'); el.type='file'; }
    else if(t==='doclist'){ el=document.createElement('ul'); String(val(n.bind)).split('|').filter(Boolean).forEach(x=>{ const li=document.createElement('li'); li.textContent=x; el.appendChild(li); }); }
    else if(t==='drawer'||t==='tabs'||t==='top-tabs'||t==='bottom-tabs'||t==='stack'){
      el=document.createElement('div'); el.className='col';
      const lab=document.createElement('div'); lab.className='muted'; lab.textContent=t+(n.text?': '+n.text:''); el.appendChild(lab);
      (n.kids||[]).forEach((k,i)=>el.appendChild(renderEditable(k, path+'.'+i)));
    }
    else if(t==='tab'){ el=document.createElement('div'); el.textContent='Tab '+(n.text||''); (n.kids||[]).forEach((k,i)=>el.appendChild(renderEditable(k, path+'.'+i))); }
    else if(t==='field'){ el=document.createElement('input'); el.className='rt-input'; el.placeholder=n.text||''; }
    else { el=document.createElement('div'); el.textContent=n.type||''; }
    if(el && el.nodeType===1) makeEditable(el, path);
    return el;
  }

  // include layouts on deploy
  const _deployBtn = document.getElementById('btn-deploy');
  if(_deployBtn){
    const prev = _deployBtn.onclick;
    _deployBtn.onclick = async ()=>{
      const name=appName.value.trim()||'mi-app';
      let src=aos.value.replace(/\(flow\s+[^\s)]+/, '(flow '+name);
      const layouts = {};
      Object.keys(layoutStore).forEach(k=>{ if(k.startsWith(name+'::') || k.startsWith(sessionApp+'::')) layouts[k]=layoutStore[k]; });
      log('Deploy producción instancia '+name+'…');
      const res=await api('/api/flow/deploy',{source:src, app:name, production:true, layouts});
      if(res.error){ log(res.error,true); return; }
      log('Aislada: '+res.app+' → '+res.url);
      sessionApp=res.app; refreshApps();
    };
  }


  /* ——— Panel inteligente de reglas + Mind ——— */
  const AOS_KEYWORDS = ['flow','orge','version','desc','kind','cap','provides','requires','state','rule','when','then','indeterminate','ui','column','row','splash','gradient','button','badge','label','field','bind','emit','assemble','surface','test','deploy','devices','mobile','tablet','desktop','blob','tabs','tab','stack','drawer','image','qr','file','doclist','bottom-tabs','top-tabs','pulse','secure'];
  const SNIPPETS = {
    orge: "  (orge nombre (kind ORGES)\n    (cap \"recurso.consultar\")\n    (provides \"recurso.consulta\" \"1.0\")\n    (state (recurso.valor 0))\n    (rule r1 (when (gt recurso.valor 0)) (then recurso.ok true)))\n",
    rule: "    (rule id (when (eq atomo valor)) (then destino true) (indeterminate false))\n",
    ui: "  (orge ui-main (kind UI) (ui (column\n    (splash \"Mi app\")\n    (gradient \"Espacio Alset\" (style \"animated\"))\n    (badge (bind recurso.valor))\n    (row (button \"+\" (emit recurso.inc)) (button \"Acción\" (emit recurso.ok=true)))))\n",
    pulse: "  ;; Pulse seguro (tipo, origen, destino, contenido) — firma Ed25519 + nonce en runtime\n  (pulse (tipo \"evento\") (origen stock) (destino venta) (contenido \"stock.changed\"))\n"
  };
  let liveOn = true, mindSuggestion = '', suggestIdx = 0;
  const KW_RE = new RegExp('\\b(' + AOS_KEYWORDS.join('|') + ')\\b', 'g');
  function highlightAos(src){
    let esc = src.replace(/&/g,'&amp;').replace(/</g,'&lt;');
    esc = esc.replace(/("(?:\\.|[^"\\])*")/g, '<span class="s">$1</span>');
    esc = esc.replace(/(\b\d+(?:\.\d+)?\b)/g, '<span class="n">$1</span>');
    esc = esc.replace(KW_RE, '<span class="k">$1</span>');
    esc = esc.replace(/([()])/g, '<span class="p">$1</span>');
    return esc;
  }
  function syncHl(){
    const hl = document.getElementById('aos-hl');
    if(!hl || !aos) return;
    hl.innerHTML = highlightAos(aos.value) + '\n';
    hl.scrollTop = aos.scrollTop; hl.scrollLeft = aos.scrollLeft;
  }
  if(aos){
    aos.addEventListener('scroll', function(){ var hl=document.getElementById('aos-hl'); if(hl){ hl.scrollTop=aos.scrollTop; hl.scrollLeft=aos.scrollLeft; }});
    aos.addEventListener('input', function(){ syncHl(); scheduleLive(); });
    syncHl();
  }
  var liveTimer = null;
  function scheduleLive(){
    if(!liveOn) return;
    clearTimeout(liveTimer);
    liveTimer = setTimeout(async function(){
      try{
        var name = (appName && appName.value.trim()) || 'mi-app';
        var src = aos.value.replace(/\(flow\s+[^\s)]+/, '(flow '+name);
        var res = await api('/api/flow/apply',{source:src, app:name});
        if(res.error) return;
        surfaces = res.surfaces||{}; surfaceOID = res.surface||Object.keys(surfaces)[0]||'';
        if(res.state) liveState = res.state;
        if(stateView) stateView.textContent = JSON.stringify(liveState,null,2);
        paint(surfaces[surfaceOID]||Object.values(surfaces)[0]);
      }catch(e){}
    }, 700);
  }
  document.querySelectorAll('[data-aos-tab]').forEach(function(btn){
    btn.onclick = function(){
      document.querySelectorAll('[data-aos-tab]').forEach(function(b){ b.classList.remove('on'); });
      btn.classList.add('on');
      var tab = btn.getAttribute('data-aos-tab');
      document.getElementById('aos-editor-wrap').classList.toggle('hidden', tab!=='editor');
      document.getElementById('mind-panel').classList.toggle('hidden', tab!=='mind');
      if(tab==='mind') mindAnalyze();
    };
  });
  function insertAtCursor(text){
    if(!aos) return;
    var start = aos.selectionStart, end = aos.selectionEnd;
    aos.value = aos.value.slice(0,start) + text + aos.value.slice(end);
    aos.selectionStart = aos.selectionEnd = start + text.length;
    aos.focus(); syncHl(); scheduleLive();
  }
  var _b;
  if((_b=document.getElementById('btn-snippet-orge'))) _b.onclick=function(){ insertAtCursor(SNIPPETS.orge); };
  if((_b=document.getElementById('btn-snippet-rule'))) _b.onclick=function(){ insertAtCursor(SNIPPETS.rule); };
  if((_b=document.getElementById('btn-snippet-ui'))) _b.onclick=function(){ insertAtCursor(SNIPPETS.ui); };
  if((_b=document.getElementById('btn-indent'))) _b.onclick=function(){
    aos.value = aos.value.split('\n').map(function(l){ return l.trim()?('  '+l):l; }).join('\n');
    syncHl();
  };
  var liveBtn = document.getElementById('btn-live-toggle');
  if(liveBtn){
    liveBtn.classList.add('live-on');
    liveBtn.onclick = function(){
      liveOn = !liveOn;
      liveBtn.classList.toggle('live-on', liveOn);
      liveBtn.textContent = liveOn ? '\u25cf Live' : '\u25cb Live';
      if(liveOn) scheduleLive();
    };
  }
  var sug = document.getElementById('aos-suggest');
  function wordAtCursor(){
    var pos = aos.selectionStart;
    var left = aos.value.slice(0, pos);
    var m = left.match(/[a-zA-Z0-9_.-]*$/);
    return { word: m?m[0]:'', start: pos-(m?m[0].length:0), end: pos };
  }
  function showSuggest(items){
    if(!items.length){ sug.classList.add('hidden'); return; }
    sug.innerHTML = items.map(function(x,i){ return '<li data-i="'+i+'" class="'+(i===0?'active':'')+'">'+x+'</li>'; }).join('');
    sug.classList.remove('hidden');
    suggestIdx = 0;
    sug.querySelectorAll('li').forEach(function(li){
      li.onmousedown = function(e){ e.preventDefault(); applySuggest(li.textContent); };
    });
  }
  function applySuggest(text){
    var w = wordAtCursor();
    aos.value = aos.value.slice(0,w.start) + text + aos.value.slice(w.end);
    aos.selectionStart = aos.selectionEnd = w.start + text.length;
    sug.classList.add('hidden');
    syncHl(); scheduleLive();
  }
  function updateSugActive(){
    var lis = [].slice.call(sug.querySelectorAll('li'));
    if(!lis.length) return;
    suggestIdx = (suggestIdx + lis.length) % lis.length;
    lis.forEach(function(li,i){ li.classList.toggle('active', i===suggestIdx); });
  }
  if(aos){
    aos.addEventListener('keydown', function(e){
      if(e.key==='Tab'){
        e.preventDefault();
        if(!sug.classList.contains('hidden')){
          var active = sug.querySelector('li.active');
          if(active) applySuggest(active.textContent);
          return;
        }
        insertAtCursor('  ');
        return;
      }
      if(e.ctrlKey && e.key===' '){
        e.preventDefault();
        var w = wordAtCursor().word;
        showSuggest(AOS_KEYWORDS.filter(function(k){ return !w || k.indexOf(w.toLowerCase())===0; }).slice(0,12));
      }
      if(!sug.classList.contains('hidden')){
        if(e.key==='ArrowDown'){ e.preventDefault(); suggestIdx++; updateSugActive(); }
        if(e.key==='ArrowUp'){ e.preventDefault(); suggestIdx--; updateSugActive(); }
        if(e.key==='Enter' && sug.querySelector('li.active')){ e.preventDefault(); applySuggest(sug.querySelector('li.active').textContent); }
        if(e.key==='Escape') sug.classList.add('hidden');
      }
    });
    aos.addEventListener('keyup', function(e){
      if(['ArrowDown','ArrowUp','Enter','Escape','Tab','Control'].indexOf(e.key)>=0) return;
      var w = wordAtCursor().word;
      if(w.length>=2) showSuggest(AOS_KEYWORDS.filter(function(k){ return k.indexOf(w.toLowerCase())===0; }).slice(0,10));
      else sug.classList.add('hidden');
    });
  }
  function mindAnalyze(){
    var src = aos.value||'';
    var orges=[], rules=[], binds=[], emits=[], m, re;
    re=/\(orge\s+([^\s)]+)/g; while((m=re.exec(src))) orges.push(m[1]);
    re=/\(rule\s+([^\s)]+)/g; while((m=re.exec(src))) rules.push(m[1]);
    re=/\(bind\s+([^\s)]+)/g; while((m=re.exec(src))) binds.push(m[1]);
    re=/\(emit\s+([^\s)]+)/g; while((m=re.exec(src))) emits.push(m[1]);
    var hasUI = /\(kind\s+UI\)|\(ui\s/.test(src);
    var parts = [];
    parts.push('ORGES: '+(orges.join(', ')||'ninguno'));
    parts.push('Reglas: '+(rules.join(', ')||'ninguna'));
    parts.push('Binds: '+(binds.join(', ')||'—'));
    parts.push('Emits: '+(emits.join(', ')||'—'));
    parts.push(hasUI ? 'Superficie UI presente (vista = organismo).' : 'Sin UI-orge: añade vista que reaccione al estado.');
    if(rules.length && !hasUI) parts.push('Mind: reglas sin vista — sugiere UI con bind a los átomos que escriben las reglas.');
    if(orges.length>=2 && !/\(assemble/.test(src)) parts.push('Mind: falta (assemble …) para acoplar LEGO.');
    if(!/\(pulse/.test(src) && orges.length>=2) parts.push('Mind: documenta Pulse firmado entre organismos (Ed25519 + nonce).');
    document.getElementById('mind-out').textContent = parts.join('\n');
    if(!hasUI && orges[0]){
      mindSuggestion = '  (orge ui-main (kind UI) (ui (column\n    (splash "'+orges[0]+'")\n    (gradient "Reactivo Alset" (style "animated"))\n    (badge (bind '+(binds[0]||orges[0]+'.estado')+'))\n    (row (button "Acción" (emit '+(emits[0]||orges[0]+'.tick')+')))))\n  (assemble '+orges.join(' ')+' ui-main)\n  (surface ui-main)\n';
    } else if(rules.length){
      mindSuggestion = '    (rule mind-check (when (exists '+(binds[0]||'estado.atom')+')) (then mind.visto true))\n';
    } else {
      mindSuggestion = SNIPPETS.orge + SNIPPETS.ui + '  (assemble nombre ui-main)\n  (surface ui-main)\n';
    }
    document.getElementById('mind-code').textContent = mindSuggestion.replace(/\\n/g, '\n');
  }
  if((_b=document.getElementById('btn-mind-analyze'))) _b.onclick=mindAnalyze;
  if((_b=document.getElementById('btn-mind-suggest-ui'))) _b.onclick=function(){
    mindSuggestion = SNIPPETS.ui + '  (assemble stock ui-main)\n  (surface ui-main)\n';
    document.getElementById('mind-code').textContent = mindSuggestion.replace(/\\n/g, '\n');
    document.getElementById('mind-out').textContent = 'Mind propone UI con splash animado, gradient y botones Alset.';
  };
  if((_b=document.getElementById('btn-mind-pulse'))) _b.onclick=function(){
    mindSuggestion = SNIPPETS.pulse;
    document.getElementById('mind-code').textContent = mindSuggestion.replace(/\\n/g, '\n');
    document.getElementById('mind-out').textContent = 'Pulse: protocolo entre ORGES. Seguro (firma+nonce), legible y resiliente en el runtime.';
  };
  if((_b=document.getElementById('btn-mind-insert'))) _b.onclick=function(){
    if(mindSuggestion) insertAtCursor('\n'+mindSuggestion.replace(/\\n/g, '\n'));
  };

  log('Studio · editor inteligente .aos · Mind agente · Live preview · ' + Object.keys(EXAMPLES).length + ' plantillas');
})();

