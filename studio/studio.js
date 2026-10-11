const DEFAULT_AOS = `(flow mi-app
  (version "1.0.0")
  (desc "Flow desde Alset Studio")

  (orge stock
    (kind ORGES)
    (cap "stock.consultar")
    (provides "stock.consulta" "1.0")
    (state (stock.unidades 5))
    (rule ok (when (gt stock.unidades 0)) (then stock.ok true)))

  (orge ui-main
    (kind UI)
    (ui (column
      (label "Mi app")
      (badge (bind stock.unidades))
      (button "Hola" (emit stock.ping)))))

  (assemble stock ui-main)
  (surface ui-main)
  (test base (eq stock.unidades 5))
  (deploy web (devices mobile tablet desktop) (out ./dist/mi-app))
)`;

const aosEl = document.getElementById('aos');
const canvas = document.getElementById('canvas');
const logEl = document.getElementById('log');
const previewRoot = document.getElementById('previewRoot');
const device = document.getElementById('device');
let deviceMode = 'mobile';

aosEl.value = localStorage.getItem('alset.studio.aos') || DEFAULT_AOS;

function log(m){ logEl.textContent = (new Date().toISOString().slice(11,19)) + ' ' + m + '\n' + logEl.textContent; }
function save(){ localStorage.setItem('alset.studio.aos', aosEl.value); }

function parseSurfacesFromAos(src){
  // Preview local minimal: extract (ui ...) tree roughly via API if available
  return null;
}

function renderPreviewLocal(){
  // Client-side lightweight: show placeholder until API returns surfaces
  previewRoot.innerHTML = '<div class="col"><div class="muted">Aplica el flow para refrescar preview IR</div></div>';
}

document.querySelectorAll('.brick').forEach(b=>{
  b.addEventListener('dragstart', e=>{
    e.dataTransfer.setData('text/kind', b.dataset.kind);
  });
});
canvas.addEventListener('dragover', e=>e.preventDefault());
canvas.addEventListener('drop', e=>{
  e.preventDefault();
  const kind = e.dataTransfer.getData('text/kind');
  const chip = document.createElement('span');
  chip.className = 'chip';
  chip.textContent = kind;
  canvas.appendChild(chip);
  const snippets = {
    orge: `\n  (orge nuevo\n    (kind ORGES)\n    (cap "x.hacer")\n    (state (x.v 1)))\n`,
    ui: `\n  (orge ui-x\n    (kind UI)\n    (ui (column (label "Vista") (button "OK" (emit x.ok)))))\n`,
    bind: `\n  (bind consumidor proveedor contrato 1.0)\n`,
    test: `\n  (test t1 (eq x.v 1))\n`,
    deploy: `\n  (deploy web (devices mobile tablet desktop) (out ./dist/app))\n`
  };
  aosEl.value += snippets[kind] || '';
  save();
  log('LEGO añadido: ' + kind);
});

document.querySelectorAll('.dev').forEach(btn=>{
  btn.onclick = ()=>{
    document.querySelectorAll('.dev').forEach(x=>x.classList.remove('on'));
    btn.classList.add('on');
    deviceMode = btn.dataset.dev;
    device.className = 'device ' + deviceMode;
    document.getElementById('devLabel').textContent = deviceMode;
  };
});

async function api(path, body){
  const r = await fetch(path, {method:'POST', headers:{'Content-Type':'application/json'}, body: JSON.stringify(body||{})});
  const t = await r.text();
  try { return JSON.parse(t); } catch { return {ok:false, raw:t, status:r.status}; }
}

document.getElementById('btnApply').onclick = async ()=>{
  save();
  log('Aplicando flow…');
  const res = await api('/api/flow/apply', {source: aosEl.value});
  log(JSON.stringify(res).slice(0,300));
  if(res.surfaces){
    const surface = res.surface || Object.keys(res.surfaces)[0];
    paint(res.surfaces[surface] || Object.values(res.surfaces)[0]);
  }
};
document.getElementById('btnTest').onclick = async ()=>{
  save();
  const res = await api('/api/flow/test', {source: aosEl.value});
  log('tests: ' + JSON.stringify(res));
};
document.getElementById('btnDeploy').onclick = async ()=>{
  save();
  const res = await api('/api/flow/deploy', {source: aosEl.value, device: deviceMode});
  log('deploy: ' + JSON.stringify(res));
  if(res.url) log('abrir ' + res.url);
};

function paint(n){
  previewRoot.innerHTML = '';
  previewRoot.appendChild(render(n));
}
function render(n){
  if(!n) return document.createTextNode('');
  const t=(n.type||'').toLowerCase();
  if(t==='column'||t==='panel'){const d=document.createElement('div');d.className='col';(n.kids||[]).forEach(k=>d.appendChild(render(k)));return d;}
  if(t==='row'){const d=document.createElement('div');d.className='row';(n.kids||[]).forEach(k=>d.appendChild(render(k)));return d;}
  if(t==='label'){const s=document.createElement('div');s.textContent=n.text||'';return s;}
  if(t==='badge'){const s=document.createElement('span');s.className='badge';s.textContent=n.bind||'·';return s;}
  if(t==='button'){const b=document.createElement('button');b.textContent=n.text||'OK';return b;}
  return document.createTextNode(n.type||'');
}
renderPreviewLocal();
aosEl.addEventListener('change', save);
log('Studio listo · arrastra piezas LEGO o edita .aos');
