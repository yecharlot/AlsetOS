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

  function setStatus(msg, isErr){
    const el=document.getElementById('status');
    if(!el) return;
    el.textContent=msg||'listo';
    el.classList.toggle('err', !!isErr);
    el.classList.toggle('ok', !isErr && msg && msg!=='listo');
  }
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


  /** Extrae IR UI mínimo del texto .aos si el API no devuelve surfaces */
  function clientUIFromAOS(src){
    if(!src) return null;
    const hasUI = /\(kind\s+UI\)/.test(src) || /\(ui\s+\(/.test(src);
    if(!hasUI) return null;
    const kids = [];
    const splash = src.match(/\(splash\s+"([^"]*)"/);
    if(splash) kids.push({type:'splash', text:splash[1]});
    if(/\(gradient/.test(src)) kids.push({type:'gradient', text:''});
    const badges = [...src.matchAll(/\(badge\s+\(bind\s+([a-zA-Z0-9_.-]+)\)/g)];
    badges.forEach(m => kids.push({type:'badge', bind:m[1]}));
    const labels = [...src.matchAll(/\(label\s+"([^"]*)"/g)];
    labels.forEach(m => kids.push({type:'label', text:m[1]}));
    const buttons = [...src.matchAll(/\(button\s+"([^"]*)"\s+\(emit\s+([a-zA-Z0-9_.=-]+)\)/g)];
    if(buttons.length){
      const row = {type:'row', kids: buttons.map(m => ({type:'button', text:m[1], emit:m[2]}))};
      kids.push(row);
    }
    if(!kids.length) kids.push({type:'label', text:'UI (sin nodos parseables en cliente)'});
    return {type:'column', kids};
  }

  function applyAlsetProps(el, n){
    if(!el || !n || el.nodeType!==1) return el;
    const p = n.props || {};
    const st = el.style;
    if(p.padding) st.padding = p.padding;
    if(p.margin) st.margin = p.margin;
    if(p.gap) st.gap = p.gap;
    if(p.radius) st.borderRadius = p.radius;
    if(p.color) st.color = p.color;
    if(p.bg || p.background) st.background = p.bg || p.background;
    if(p.width) st.width = p.width;
    if(p.height) st.height = p.height;
    if(p.fontsize || p['font-size']) st.fontSize = p.fontsize || p['font-size'];
    if(p.weight) st.fontWeight = p.weight;
    if(p.opacity) st.opacity = p.opacity;
    if(p.border) st.border = p.border;
    if(p.shadow) st.boxShadow = p.shadow;
    if(p.align) st.alignItems = p.align;
    if(p.justify) st.justifyContent = p.justify;
    return el;
  }
  function pickSurface(res){
    const surfacesMap = res && res.surfaces ? res.surfaces : surfaces;
    const oid = (res && res.surface) || surfaceOID || Object.keys(surfacesMap||{})[0] || '';
    let tree = oid ? surfacesMap[oid] : null;
    if(!tree) tree = Object.values(surfacesMap||{})[0];
    if(!tree) tree = clientUIFromAOS(aos && aos.value);
    return { tree, oid, surfacesMap };
  }

  async function api(path, body){
    const r=await fetch(path,{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify(body||{})});
    const t=await r.text();
    let j; try{j=JSON.parse(t);}catch{j={ok:false,error:t};}
    if(!r.ok){ throw new Error(j.error||t||r.statusText); }
    return j;
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
    if(stateView) stateView.textContent=JSON.stringify(liveState,null,2);
    log('OK instancia '+sessionApp+' · orgs '+(res.organismos||0)+' · surface '+(surfaceOID||'—'));
    const picked = pickSurface(res);
    surfaces = picked.surfacesMap || surfaces;
    surfaceOID = picked.oid || surfaceOID;
    paint(picked.tree);
    try{ inspectorAnalyze(); }catch(_){}
    refreshApps();
  };
  $('#btn-run-preview').onclick=()=>{
    const picked = pickSurface({surfaces, surface: surfaceOID});
    paint(picked.tree);
    log('Preview '+deviceMode+(picked.tree?' · UI ok':' · sin UI'));
  };
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
      log(editMode ? 'Layout ON: usa el handle dorado para mover. Clic selecciona props.' : 'Clic en un componente del preview para editar props');
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
    if(!el || el.nodeType!==1) return;
    el.dataset.path = path;
    el.classList.add('ui-selectable');
    // Click = seleccionar para props (no mueve el layout)
    el.addEventListener('click', function(ev){
      // no interceptar botones funcionales salvo en editMode
      if(ev.target.closest && ev.target.closest('button') && !editMode) return;
      if(ev.target.classList && ev.target.classList.contains('ui-drag-handle')) return;
      ev.stopPropagation();
      document.querySelectorAll('#previewRoot .ui-selected').forEach(function(x){ x.classList.remove('ui-selected'); });
      el.classList.add('ui-selected');
      // Recuperar nodo del IR si está en dataset — selectUINode se llama desde fuera con n
      if(el.__uiNode && typeof selectUINode==='function') selectUINode(el.__uiNode, path);
    });
    if(!editMode) return;
    el.style.outline = '1px dashed rgba(244,180,0,.35)';
    if(getComputedStyle(el).position==='static') el.style.position='relative';
    applyLayout(el, path);
    // Solo el handle mueve/redimensiona — evita desajustar al hacer clic
    const handle = document.createElement('div');
    handle.className = 'ui-drag-handle';
    handle.title = 'Arrastrar / redimensionar';
    handle.style.cssText = 'position:absolute;right:2px;bottom:2px;width:14px;height:14px;background:#f4b400;border-radius:3px;cursor:nwse-resize;z-index:6';
    el.appendChild(handle);
    let mode=null, sx, sy, sl, st, sw, sh;
    handle.addEventListener('mousedown', function(ev){
      mode = (ev.shiftKey ? 'resize' : 'drag');
      ev.preventDefault(); ev.stopPropagation();
      const r = el.getBoundingClientRect();
      const pr = previewRoot.getBoundingClientRect();
      sx=ev.clientX; sy=ev.clientY;
      sl=r.left-pr.left; st=r.top-pr.top; sw=r.width; sh=r.height;
      el.style.position='absolute';
      el.style.left=sl+'px'; el.style.top=st+'px';
      const move=function(e){
        if(mode==='drag'){
          el.style.left=(sl+e.clientX-sx)+'px';
          el.style.top=(st+e.clientY-sy)+'px';
        } else {
          el.style.width=Math.max(40, sw+e.clientX-sx)+'px';
          el.style.height=Math.max(24, sh+e.clientY-sy)+'px';
        }
      };
      const up=function(){
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
    if(!previewRoot) return;
    previewRoot.innerHTML='';
    previewRoot.style.position='relative';
    previewRoot.style.minHeight='160px';
    previewRoot.classList.add('alset-ui');
    if(!tree || (typeof tree==='object' && !tree.type && !(tree.kids||[]).length)){
      const fb = clientUIFromAOS(aos && aos.value);
      tree = fb;
    }
    if(!tree){ previewRoot.innerHTML='<p class="muted">Sin UI — añade (orge … (kind UI) (ui …)) y (surface …), luego Aplicar flow</p>'; return; }
    try {
      const rootEl = renderEditable(tree, '0');
      if(rootEl) previewRoot.appendChild(rootEl);
      else previewRoot.innerHTML='<p class="muted">UI vacía</p>';
    } catch(err) {
      console.error(err);
      previewRoot.innerHTML='<p class="muted">Error render: '+(err.message||err)+'</p>';
    }
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
        else if(t==='top-tabs'||t==='tabs'||t==='bottom-tabs'){
      el=document.createElement('div'); el.className='alset-tabs '+t;
      const bar=document.createElement('div'); bar.className='alset-tabs-bar';
      const panels=document.createElement('div'); panels.className='alset-tabs-panels';
      (n.kids||[]).forEach(function(k,i){
        const btn=document.createElement('button'); btn.type='button';
        btn.className='alset-tab-btn'+(i===0?' on':'');
        btn.textContent=k.text||('Tab '+(i+1));
        const panel=document.createElement('div'); panel.className='alset-tab-panel';
        panel.style.display=i===0?'block':'none';
        (k.kids||[]).forEach(function(c,j){ panel.appendChild(renderEditable(c, path+'.'+i+'.'+j)); });
        btn.onclick=function(ev){
          ev.stopPropagation();
          bar.querySelectorAll('.alset-tab-btn').forEach(function(b){ b.classList.remove('on'); });
          btn.classList.add('on');
          panels.querySelectorAll('.alset-tab-panel').forEach(function(p){ p.style.display='none'; });
          panel.style.display='block';
        };
        bar.appendChild(btn); panels.appendChild(panel);
      });
      if(t==='bottom-tabs'){ el.appendChild(panels); el.appendChild(bar); }
      else { el.appendChild(bar); el.appendChild(panels); }
    }
    else if(t==='drawer'){
      el=document.createElement('div'); el.className='alset-drawer open';
      const head=document.createElement('div'); head.className='alset-drawer-head';
      head.innerHTML='<span>'+(n.text||'Menu')+'</span><span>▾</span>';
      const body=document.createElement('div'); body.className='alset-drawer-body'; body.style.display='block';
      (n.kids||[]).forEach(function(k,i){ body.appendChild(renderEditable(k, path+'.'+i)); });
      head.onclick=function(ev){ ev.stopPropagation(); el.classList.toggle('open'); body.style.display=el.classList.contains('open')?'block':'none'; };
      el.appendChild(head); el.appendChild(body);
    }
    else if(t==='stack'){
      el=document.createElement('div'); el.className='alset-stack';
      (n.kids||[]).forEach(function(k,i){ el.appendChild(renderEditable(k, path+'.'+i)); });
    }
    else if(t==='tab'){
      el=document.createElement('div'); el.className='col';
      (n.kids||[]).forEach(function(k,i){ el.appendChild(renderEditable(k, path+'.'+i)); });
    }
    else if(t==='field'){ el=document.createElement('input'); el.className='rt-input'; el.placeholder=n.text||''; }
    else { el=document.createElement('div'); el.textContent=n.type||''; }
    if(el && el.nodeType===1){
      el.__uiNode = n;
      applyAlsetProps(el, n);
      el.dataset.uiPath = path;
      makeEditable(el, path);
    }
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
        if(res && res.error) { try{ inspectorAnalyze(); }catch(_){} return; }
        surfaces = (res && res.surfaces) || {};
        surfaceOID = (res && res.surface) || Object.keys(surfaces)[0] || '';
        if(res && res.state) liveState = res.state;
        if(stateView) stateView.textContent = JSON.stringify(liveState,null,2);
        var picked = pickSurface(res || {});
        paint(picked.tree);
        try{ inspectorAnalyze(); }catch(_){}
      }catch(e){
        // API falló: al menos pintar fallback cliente + inspector local
        paint(clientUIFromAOS(aos.value));
        try{ inspectorAnalyze(); }catch(_){}
      }
    }, 700);
  }
  document.querySelectorAll('[data-aos-tab]').forEach(function(btn){
    btn.onclick = function(){
      document.querySelectorAll('[data-aos-tab]').forEach(function(b){ b.classList.remove('on'); });
      btn.classList.add('on');
      var tab = btn.getAttribute('data-aos-tab');
      var ed = document.getElementById('aos-editor-wrap');
      var mind = document.getElementById('mind-panel');
      var decl = document.getElementById('declare-panel');
      if(ed) ed.classList.toggle('hidden', tab!=='editor');
      if(mind) mind.classList.toggle('hidden', tab!=='mind');
      if(decl) decl.classList.toggle('hidden', tab!=='declare');
      if(tab==='mind') mindAnalyze();
      if(tab==='declare') renderEmitComponents();
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
          if(active){ var it=(sug._items&&sug._items[+active.getAttribute('data-i')])||{insert:active.textContent}; applySuggest(it.insert||it.label||active.textContent); }
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
        if(e.key==='Enter' && sug.querySelector('li.active')){ e.preventDefault(); var ai=sug.querySelector('li.active'); var it=(sug._items&&sug._items[+ai.getAttribute('data-i')])||{insert:ai.textContent}; applySuggest(it.insert||it.label||ai.textContent); }
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
  async function mindAnalyze(){
    const src = aos.value || '';
    setStatus('Mind analiza…');
    try {
      const res = await api('/api/mind/analyze', { source: src, mode: 'analyze', message: '' });
      renderMindResponse(res);
      setStatus('Mind listo');
    } catch (e) {
      // fallback local if API down
      document.getElementById('mind-out').textContent = 'Mind local (API no disponible): ' + (e.message||e);
      setStatus('Mind offline', true);
    }
  }

  function renderMindResponse(res){
    if(!res) return;
    const out = document.getElementById('mind-out');
    if(out) out.textContent = res.reply || '';
    const diags = document.getElementById('mind-diags');
    if(diags){
      diags.innerHTML = (res.diagnostics||[]).map(d=>'<div class="diag">⚠ '+escapeHtml(d)+'</div>').join('');
    }
    const sugs = document.getElementById('mind-sugs');
    mindSuggestionsList = res.suggestions || [];
    if(sugs){
      sugs.innerHTML = mindSuggestionsList.map((s,i)=>
        '<button type="button" class="sug-btn" data-sug="'+i+'"><b>'+escapeHtml(s.title)+'</b> <span class="muted">'+escapeHtml(s.kind)+'</span><br/><small>'+escapeHtml(s.reason||'')+'</small></button>'
      ).join('');
      sugs.querySelectorAll('[data-sug]').forEach(btn=>{
        btn.onclick = ()=>{
          const s = mindSuggestionsList[+btn.dataset.sug];
          if(!s) return;
          mindSuggestion = s.snippet || '';
          document.getElementById('mind-code').textContent = mindSuggestion;
        };
      });
    }
    if(mindSuggestionsList[0]){
      mindSuggestion = mindSuggestionsList[0].snippet || '';
      document.getElementById('mind-code').textContent = mindSuggestion;
    }
  }

  function escapeHtml(s){
    return String(s||'').replace(/&/g,'&amp;').replace(/</g,'&lt;').replace(/>/g,'&gt;');
  }

  let mindSuggestionsList = [];

  async function mindChat(){
    const input = document.getElementById('mind-chat-input');
    const msg = (input && input.value || '').trim();
    if(!msg) return;
    const log = document.getElementById('mind-chat-log');
    if(log) log.innerHTML += '<div class="chat-user"><b>Tú:</b> '+escapeHtml(msg)+'</div>';
    if(input) input.value = '';
    setStatus('Mind piensa…');
    try {
      const res = await api('/api/mind/chat', { source: aos.value||'', message: msg, mode: 'chat' });
      renderMindResponse(res);
      if(log) log.innerHTML += '<div class="chat-mind"><b>Mind:</b> '+escapeHtml(res.reply||'').replace(/\n/g,'<br/>')+'</div>';
      log && (log.scrollTop = log.scrollHeight);
      setStatus('Mind listo');
    } catch(e){
      if(log) log.innerHTML += '<div class="chat-mind"><b>Mind:</b> error '+escapeHtml(e.message||e)+'</div>';
      setStatus('Mind error', true);
    }
  }
  if((_b=document.getElementById('btn-mind-analyze'))) _b.onclick=mindAnalyze;
  if((_b=document.getElementById('btn-mind-chat'))) _b.onclick=mindChat;
  if((_b=document.getElementById('mind-chat-input'))) _b.addEventListener('keydown',function(e){ if(e.key==='Enter') mindChat(); });
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


  /* Emits como componentes reutilizables */
  var emitLibrary = [
    {name:'stock.inc', effect:'stock.unidades +1', snip:'(button "+" (emit stock.inc))'},
    {name:'stock.dec', effect:'stock.unidades -1', snip:'(button "-" (emit stock.dec))'},
    {name:'venta.registrar', effect:'descuenta stock', snip:'(button "Vender" (emit venta.registrar))'},
    {name:'pago.estado=pagado', effect:'marca pagado', snip:'(button "Pagar" (emit pago.estado=pagado))'}
  ];
  var declareDraft = { states:[], rules:[], orges:[], emits:[], flow:'' };

  function renderEmitComponents(){
    var box = document.getElementById('emit-components');
    if(!box) return;
    box.innerHTML = '';
    emitLibrary.forEach(function(e){
      var c = document.createElement('span');
      c.className = 'emit-chip';
      c.draggable = true;
      c.textContent = e.name;
      c.title = e.effect;
      c.ondragstart = function(ev){ ev.dataTransfer.setData('text/plain', e.snip); };
      c.onclick = function(){ insertAtCursor('\n  '+e.snip+'\n'); };
      box.appendChild(c);
    });
  }
  function refreshDeclareDraft(){
    var lines = [];
    declareDraft.states.forEach(function(s){ lines.push('    ('+s.key+' '+s.val+')'); });
    var stateBlock = declareDraft.states.length ? '    (state\n'+declareDraft.states.map(function(s){return '      ('+s.key+' '+s.val+')';}).join('\n')+'\n    )' : '';
    declareDraft.orges.forEach(function(o){
      lines.push('  (orge '+o.name+' (kind ORGES) (cap "'+o.cap+'")');
      if(stateBlock) lines.push(stateBlock);
      declareDraft.rules.forEach(function(r){
        lines.push('    (rule '+r.id+' (when '+r.when+') (then '+r.then+'))');
      });
      lines.push('  )');
    });
    declareDraft.emits.forEach(function(e){
      lines.push('  ;; emit componente: '+e.name+' → '+e.effect);
    });
    if(declareDraft.flow){
      lines.push('(flow '+declareDraft.flow.name);
      lines.push('  (version "1.0.0")');
      lines = lines.concat(/* orges already */);
      lines.push('  (assemble '+declareDraft.flow.assemble+')');
      lines.push('  (surface '+declareDraft.flow.surface+')');
      lines.push(')');
    }
    var pre = document.getElementById('declare-draft');
    if(pre) pre.textContent = lines.join('\n') || '(vacío — declara estado, reglas, orges…)';
  }
  function wireDeclare(){
    function val(id){ var el=document.getElementById(id); return el?el.value.trim():''; }
    var b;
    if((b=document.getElementById('d-add-state'))) b.onclick=function(){
      var k=val('d-state-key'), v=val('d-state-val')||'0';
      if(!k) return; declareDraft.states.push({key:k,val:v}); refreshDeclareDraft();
    };
    if((b=document.getElementById('d-add-rule'))) b.onclick=function(){
      declareDraft.rules.push({id:val('d-rule-id')||'r1', when:val('d-rule-when')||'(true)', then:val('d-rule-then')||'ok true'});
      refreshDeclareDraft();
    };
    if((b=document.getElementById('d-add-emit'))) b.onclick=function(){
      var name=val('d-emit-name'), effect=val('d-emit-effect');
      if(!name) return;
      emitLibrary.push({name:name, effect:effect||name, snip:'(button "'+name+'" (emit '+name+'))'});
      declareDraft.emits.push({name:name, effect:effect});
      renderEmitComponents();
      // también al toolbox acciones
      var tb=document.getElementById('toolbox-actions');
      if(tb){
        var d=document.createElement('div'); d.className='tool'; d.draggable=true;
        d.textContent=name; d.dataset.snip='(emit '+name+')';
        d.ondragstart=function(ev){ ev.dataTransfer.setData('text/plain', d.dataset.snip); };
        tb.appendChild(d);
      }
      refreshDeclareDraft();
      log('Emit componente: '+name);
    };
    if((b=document.getElementById('d-add-orge'))) b.onclick=function(){
      declareDraft.orges.push({name:val('d-orge-name')||'orge1', cap:val('d-orge-cap')||'cap.default'});
      refreshDeclareDraft();
    };
    if((b=document.getElementById('d-build-flow'))) b.onclick=function(){
      declareDraft.flow = {
        name: val('d-flow-name')||'mi-app',
        assemble: val('d-assemble')||declareDraft.orges.map(function(o){return o.name;}).join(' '),
        surface: val('d-surface')||'ui-main'
      };
      var body = [];
      body.push('(flow '+declareDraft.flow.name);
      body.push('  (version "1.0.0")');
      body.push('  (desc "Declarado con Mind")');
      declareDraft.orges.forEach(function(o){
        body.push('  (orge '+o.name+' (kind ORGES) (cap "'+o.cap+'")');
        if(declareDraft.states.length){
          body.push('    (state');
          declareDraft.states.forEach(function(s){ body.push('      ('+s.key+' '+s.val+')'); });
          body.push('    )');
        }
        declareDraft.rules.forEach(function(r){
          body.push('    (rule '+r.id+' (when '+r.when+') (then '+r.then+'))');
        });
        body.push('  )');
      });
      // UI mínima si no hay ui en assemble
      if((declareDraft.flow.assemble||'').indexOf('ui-')>=0 || declareDraft.flow.surface){
        body.push('  (orge '+(declareDraft.flow.surface||'ui-main')+' (kind UI) (ui (column');
        body.push('    (splash "'+declareDraft.flow.name+'")');
        if(declareDraft.states[0]) body.push('    (badge (bind '+declareDraft.states[0].key+'))');
        emitLibrary.slice(-5).forEach(function(e){ body.push('    '+e.snip); });
        body.push('  )))');
      }
      body.push('  (assemble '+declareDraft.flow.assemble+(declareDraft.flow.surface?' '+declareDraft.flow.surface:'')+')');
      body.push('  (surface '+declareDraft.flow.surface+')');
      body.push('  (deploy web (devices mobile tablet desktop))');
      body.push(')');
      var pre=document.getElementById('declare-draft');
      if(pre) pre.textContent = body.join('\n');
      declareDraft._full = body.join('\n');
    };
    if((b=document.getElementById('d-to-aos'))) b.onclick=function(){
      var pre=document.getElementById('declare-draft');
      var src = (declareDraft._full) || (pre && pre.textContent) || '';
      if(!src || src.indexOf('vacío')>=0) return;
      aos.value = src;
      syncHl(); scheduleLive();
      // switch to editor tab
      var tab=document.querySelector('[data-aos-tab="editor"]');
      if(tab) tab.click();
      log('Borrador Mind pasado al .aos');
    };
  }
  wireDeclare();
  renderEmitComponents();

  /* Sugerencias en tiempo real enriquecidas (keywords + emits + snippets) */
  function richSuggest(word){
    var items = [];
    var w = (word||'').toLowerCase();
    AOS_KEYWORDS.forEach(function(k){
      if(!w || k.indexOf(w)===0) items.push({label:k, detail:'keyword', insert:k});
    });
    emitLibrary.forEach(function(e){
      if(!w || e.name.indexOf(w)===0 || e.name.indexOf(w)>=0)
        items.push({label:e.name, detail:'emit · '+e.effect, insert:e.name});
    });
    if(!w || 'rule'.indexOf(w)===0) items.push({label:'(rule …)', detail:'snippet', insert:'(rule r1 (when (true)) (then ok true))'});
    if(!w || 'state'.indexOf(w)===0) items.push({label:'(state …)', detail:'snippet', insert:'(state (campo 0))'});
    if(!w || 'orge'.indexOf(w)===0) items.push({label:'(orge …)', detail:'snippet', insert:'(orge nombre (kind ORGES) (state (x 0)))'});
    if(!w || 'assemble'.indexOf(w)===0) items.push({label:'(assemble …)', detail:'snippet', insert:'(assemble a b ui-main)'});
    if(!w || 'pulse'.indexOf(w)===0) items.push({label:'(pulse …)', detail:'snippet', insert:SNIPPETS.pulse});
    return items.slice(0, 14);
  }
  // Override showSuggest to support objects
  var _origShow = showSuggest;
  showSuggest = function(items){
    if(!items || !items.length){ sug.classList.add('hidden'); return; }
    // normalize
    items = items.map(function(x){
      if(typeof x==='string') return {label:x, detail:'', insert:x};
      return x;
    });
    sug._items = items;
    sug.innerHTML = items.map(function(x,i){
      return '<li data-i="'+i+'" class="'+(i===0?'active':'')+'">'+x.label+(x.detail?'<small>'+x.detail+'</small>':'')+'</li>';
    }).join('');
    sug.classList.remove('hidden');
    suggestIdx = 0;
    sug.querySelectorAll('li').forEach(function(li){
      li.onmousedown = function(e){
        e.preventDefault();
        var it = sug._items[+li.getAttribute('data-i')];
        applySuggest(it.insert || it.label);
      };
    });
  };
  // Replace keyup handler by re-binding: patch via capturing on aos
  if(aos){
    aos.addEventListener('keyup', function(e){
      if(['ArrowDown','ArrowUp','Enter','Escape','Tab','Control'].indexOf(e.key)>=0) return;
      var w = wordAtCursor().word;
      if(w.length>=1) showSuggest(richSuggest(w));
      else sug.classList.add('hidden');
    });
  }

  log('Studio · .aos live + Mind declarar + emits componentes · ' + Object.keys(EXAMPLES).length + ' plantillas');

  /* ——— AlsetInspector permanente (derecha del preview) ——— */
  let inspectorSnippet = '';
  let inspectorBuilt = '';

  async function inspectorAnalyze(){
    const src = (aos && aos.value) || '';
    const badge = document.getElementById('inspector-badge');
    if(badge){ badge.textContent='analizando…'; badge.classList.remove('ok'); }
    try {
      const res = await api('/api/mind/analyze', { source: src, mode: 'analyze', message: '' });
      renderInspector(res);
      if(badge){ badge.textContent='activo'; badge.classList.add('ok'); }
    } catch(e) {
      const out = document.getElementById('inspector-out');
      if(out) out.textContent = 'Inspector offline: '+(e.message||e)+'\nRevisa sintaxis localmente (paréntesis, flow, surface).';
      if(badge) badge.textContent='offline';
      // local paren check
      let bal=0; for(const c of src){ if(c==='(')bal++; else if(c===')')bal--; if(bal<0)break; }
      const diags = document.getElementById('inspector-diags');
      if(diags && bal!==0) diags.innerHTML = '<div class="diag">⚠ Paréntesis desbalanceados (Δ '+bal+')</div>';
    }
  }

  function renderInspector(res){
    if(!res) return;
    const out = document.getElementById('inspector-out');
    if(out) out.textContent = res.reply || '';
    const diags = document.getElementById('inspector-diags');
    if(diags) diags.innerHTML = (res.diagnostics||[]).map(d=>'<div class="diag">⚠ '+escapeHtml(d)+'</div>').join('');
    const sugs = document.getElementById('inspector-sugs');
    const list = res.suggestions || [];
    if(sugs){
      sugs.innerHTML = list.map((s,i)=>
        '<button type="button" class="sug-btn" data-isug="'+i+'"><b>'+escapeHtml(s.title)+'</b> <span class="muted">'+escapeHtml(s.kind)+'</span><br/><small>'+escapeHtml(s.reason||'')+'</small></button>'
      ).join('');
      sugs.querySelectorAll('[data-isug]').forEach(btn=>{
        btn.onclick = ()=>{
          const s = list[+btn.dataset.isug];
          if(!s) return;
          inspectorSnippet = s.snippet || '';
          if(s.kind==='build') inspectorBuilt = s.snippet || '';
          const code = document.getElementById('inspector-code');
          if(code) code.textContent = inspectorSnippet;
        };
      });
    }
    if(res.built_aos){ inspectorBuilt = res.built_aos; inspectorSnippet = res.built_aos; }
    else if(list[0]){ inspectorSnippet = list[0].snippet || ''; }
    const code = document.getElementById('inspector-code');
    if(code && inspectorSnippet) code.textContent = inspectorSnippet;
  }

  async function inspectorChat(){
    const input = document.getElementById('inspector-chat-input');
    const msg = (input && input.value || '').trim();
    if(!msg) return;
    const logEl = document.getElementById('inspector-chat-log');
    if(logEl) logEl.innerHTML += '<div class="chat-user"><b>Tú:</b> '+escapeHtml(msg)+'</div>';
    if(input) input.value = '';
    try {
      const res = await api('/api/mind/chat', { source: (aos&&aos.value)||'', message: msg, mode: 'chat' });
      renderInspector(res);
      if(res.built_aos){
        inspectorBuilt = res.built_aos;
        const code = document.getElementById('inspector-code');
        if(code) code.textContent = res.built_aos;
      }
      if(logEl) logEl.innerHTML += '<div class="chat-mind"><b>Inspector:</b> '+escapeHtml(res.reply||'').replace(/\n/g,'<br/>')+'</div>';
      if(logEl) logEl.scrollTop = logEl.scrollHeight;
    } catch(e) {
      if(logEl) logEl.innerHTML += '<div class="chat-mind">Error: '+escapeHtml(e.message||e)+'</div>';
    }
  }

  (function wireInspector(){
    const ins = document.getElementById('btn-inspector-insert');
    if(ins) ins.onclick = ()=>{
      if(!inspectorSnippet || !aos) return;
      insertAtCursor('\n'+inspectorSnippet+'\n');
      log('Inspector: snippet insertado');
      scheduleLive();
    };
    const apl = document.getElementById('btn-inspector-apply-build');
    if(apl) apl.onclick = ()=>{
      const src = inspectorBuilt || inspectorSnippet;
      if(!src || !aos) return;
      aos.value = src;
      if(typeof syncHl==='function') syncHl();
      log('Inspector: flow cargado en editor');
      scheduleLive();
      $('#btn-apply') && $('#btn-apply').click();
    };
    const chat = document.getElementById('btn-inspector-chat');
    if(chat) chat.onclick = inspectorChat;
    const cin = document.getElementById('inspector-chat-input');
    if(cin) cin.addEventListener('keydown', function(e){ if(e.key==='Enter') inspectorChat(); });
    // continuous
    setTimeout(inspectorAnalyze, 400);
    setInterval(function(){ if(document.getElementById('pane-inspector')) inspectorAnalyze(); }, 12000);
  })();

  // Device tabs if present
  document.querySelectorAll('#device-tabs [data-device]').forEach(function(btn){
    btn.onclick = function(){
      document.querySelectorAll('#device-tabs [data-device]').forEach(b=>b.classList.remove('on'));
      btn.classList.add('on');
      deviceMode = btn.getAttribute('data-device')||'mobile';
      if(device){ device.className = 'device '+deviceMode; }
      const lab = document.getElementById('devLabel');
      if(lab) lab.textContent = deviceMode;
      const picked = pickSurface({surfaces, surface: surfaceOID});
      paint(picked.tree);
    };
  });

  // Initial paint attempt from editor
  setTimeout(function(){ try{ scheduleLive(); }catch(_){} }, 600);


  /* ——— Props de órganos UI + Inspector chrome ——— */
  var selectedNode = null, selectedPath = '';
  function selectUINode(n, path){
    selectedNode = n; selectedPath = path || '';
    var sel = document.getElementById('prop-sel');
    if(sel) sel.value = ((n && n.type) || '') + (path ? (' @ ' + path) : '');
    var emit = document.getElementById('prop-emit');
    if(emit) emit.value = (n && n.emit) || '';
    var bind = document.getElementById('prop-bind');
    if(bind) bind.value = (n && n.bind) || '';
    var p = (n && n.props) || {};
    function set(id, k){ var el=document.getElementById(id); if(el) el.value = p[k] || ''; }
    set('prop-padding','padding'); set('prop-margin','margin'); set('prop-gap','gap');
    set('prop-radius','radius'); set('prop-color','color'); set('prop-bg','bg');
    set('prop-width','width'); set('prop-height','height'); set('prop-fontsize','fontsize');
    log('Seleccionado: ' + ((n && n.type) || '?') + ' ' + selectedPath);
  }

  function applyPropsToAOS(){
    if(!selectedNode){ log('Haz clic en un componente del preview para seleccionarlo', true); return; }
    if(!selectedNode.props) selectedNode.props = {};
    var map = {
      padding:'prop-padding', margin:'prop-margin', gap:'prop-gap', radius:'prop-radius',
      color:'prop-color', bg:'prop-bg', width:'prop-width', height:'prop-height', fontsize:'prop-fontsize'
    };
    Object.keys(map).forEach(function(k){
      var el = document.getElementById(map[k]);
      if(el && el.value.trim()) selectedNode.props[k] = el.value.trim();
      else if(selectedNode.props[k]) delete selectedNode.props[k];
    });
    var emit = document.getElementById('prop-emit');
    var bind = document.getElementById('prop-bind');
    if(emit && emit.value.trim()) selectedNode.emit = emit.value.trim();
    if(bind && bind.value.trim()) selectedNode.bind = bind.value.trim();
    // Anotar en .aos si hay texto de botón
    var propsPairs = Object.keys(selectedNode.props||{}).map(function(k){ return '('+k+' '+selectedNode.props[k]+')'; }).join(' ');
    if(propsPairs && selectedNode.text && aos){
      var label = selectedNode.text;
      var re = new RegExp('(\\(button\\s+"'+label.replace(/[.*+?^${}()|[\\]\\\\]/g,'\\$&')+'"[^\\)]*?)(\\))');
      if(re.test(aos.value)){
        aos.value = aos.value.replace(re, function(m,a,b){
          if(/\bpadding\b|\bradius\b|\bbg\b/.test(a)) return m;
          return a + ' ' + propsPairs + b;
        });
        if(typeof syncHl==='function') syncHl();
      }
    }
    log('Props del órgano aplicadas');
    paint(pickSurface({surfaces: surfaces, surface: surfaceOID}).tree);
    if(typeof scheduleLive==='function') scheduleLive();
  }
  var _bp = document.getElementById('btn-apply-prop');
  if(_bp) _bp.onclick = applyPropsToAOS;

  (function wireInspectorChrome(){
    var insp = document.getElementById('pane-inspector');
    var row = document.getElementById('preview-row');
    if(!insp || !row) return;
    var head = insp.querySelector('.pane-head');
    if(head && !head.querySelector('[data-close-inspector]')){
      var x = document.createElement('button');
      x.type = 'button'; x.className = 'icon-btn'; x.setAttribute('data-close-inspector','1');
      x.textContent = '×'; x.title = 'Cerrar inspector';
      head.appendChild(x);
      x.onclick = function(){ insp.classList.add('collapsed'); };
    }
    var top = document.getElementById('topbar');
    if(top && !document.getElementById('btn-toggle-inspector')){
      var t = document.createElement('button');
      t.type = 'button'; t.id = 'btn-toggle-inspector'; t.className = 'btn btn-ghost';
      t.textContent = 'Inspector';
      top.appendChild(t);
      t.onclick = function(){ insp.classList.toggle('collapsed'); };
    }
    if(!document.getElementById('splitter-inspector')){
      var sp = document.createElement('div');
      sp.id = 'splitter-inspector'; sp.className = 'splitter-inspector';
      row.insertBefore(sp, insp);
      var startX=0, startW=0;
      sp.addEventListener('mousedown', function(ev){
        ev.preventDefault();
        startX = ev.clientX;
        startW = insp.getBoundingClientRect().width;
        function move(e){
          var w = Math.max(200, Math.min(window.innerWidth*0.55, startW - (e.clientX - startX)));
          insp.style.flex = '0 0 ' + w + 'px';
          insp.style.maxWidth = w + 'px';
          insp.style.width = w + 'px';
        }
        function up(){
          window.removeEventListener('mousemove', move);
          window.removeEventListener('mouseup', up);
        }
        window.addEventListener('mousemove', move);
        window.addEventListener('mouseup', up);
      });
    }
  })();

})();

