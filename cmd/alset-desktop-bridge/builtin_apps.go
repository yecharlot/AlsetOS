package main

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type builtinApp struct {
	Name     string
	Title    string
	Glyph    string
	MenuOnly bool
	HTML     string
}

func (b *bridge) seedBuiltinApps() {
	apps := []builtinApp{
		{Name: "org-manager", Title: "Organismos", Glyph: "◎", HTML: appHTMLOrgManager()},
		{Name: "doc-reader", Title: "Documentos", Glyph: "📄", HTML: appHTMLDocReader()},
		{Name: "notes", Title: "Notas", Glyph: "📝", HTML: appHTMLNotes()},
		{Name: "audio-player", Title: "Audio", Glyph: "♫", MenuOnly: true, HTML: appHTMLAudio()},
		{Name: "video-player", Title: "Video", Glyph: "▶", MenuOnly: true, HTML: appHTMLVideo()},
		{Name: "image-viewer", Title: "Imágenes", Glyph: "🖼", HTML: appHTMLImage()},
		{Name: "accounts", Title: "Cuentas", Glyph: "👤", HTML: appHTMLAccounts()},
		{Name: "ipfs-store", Title: "IPFS Store", Glyph: "⬡", HTML: appHTMLIPFS()},
		{Name: "files-ui", Title: "Archivos", Glyph: "📁", HTML: appHTMLFiles()},
		{Name: "calculadora", Title: "Calculadora", Glyph: "🔢", HTML: ""},
	}
	for _, a := range apps {
		dir := filepath.Join(b.dataDir, "apps", a.Name)
		// Always refresh builtin HTML so UI evolves with releases
		_ = os.MkdirAll(dir, 0o755)
		html := a.HTML
		if html == "" {
			html = buildAppHTML(a.Name, a.Title, map[string]any{})
		}
		_ = os.WriteFile(filepath.Join(dir, "index.html"), []byte(html), 0o644)
		meta, _ := json.MarshalIndent(map[string]any{
			"name": a.Name, "title": a.Title, "glyph": a.Glyph,
			"menu_only": a.MenuOnly, "builtin": true, "desktop": !a.MenuOnly,
		}, "", "  ")
		_ = os.WriteFile(filepath.Join(dir, "app.json"), meta, 0o644)
		b.orgs.upsert(&sysOrganism{
			ID: "app:" + a.Name, Kind: "app", Name: a.Title, State: "installed",
			Meta: map[string]any{"builtin": true, "menu_only": a.MenuOnly},
		})
	}
	b.orgs.save()
}

func shellApp(title, body string) string {
	return `<!DOCTYPE html><html lang="es"><head><meta charset="utf-8"/>
<meta name="viewport" content="width=device-width,initial-scale=1"/>
<title>` + title + `</title>
<style>
:root{--bg:#0b0f16;--panel:#131926;--panel2:#1a2233;--line:rgba(255,255,255,.08);--gold:#f0c14b;--text:#eef1f6;--muted:#8b93a7;--accent:#5b9fd4;--ok:#3ddc97;--danger:#e85d5d}
*{box-sizing:border-box}html,body{height:100%}
body{margin:0;font-family:system-ui,-apple-system,sans-serif;background:var(--bg);color:var(--text);display:flex;flex-direction:column;min-height:100%}
header.app{padding:10px 14px;border-bottom:1px solid var(--line);display:flex;align-items:center;gap:10px;background:linear-gradient(180deg,#151c2a,#0f141e)}
header.app h1{font-size:14px;margin:0;color:var(--gold);letter-spacing:.04em;font-weight:700}
header.app .sub{color:var(--muted);font-size:11px;margin-left:auto}
main{flex:1;padding:12px;overflow:auto;display:flex;flex-direction:column;gap:10px}
.card{background:var(--panel);border:1px solid var(--line);border-radius:12px;padding:12px}
.card h3{margin:0 0 8px;font-size:12px;color:var(--muted);text-transform:uppercase;letter-spacing:.06em}
.row{display:flex;flex-wrap:wrap;gap:6px;align-items:center}
.col{display:flex;flex-direction:column;gap:8px}
button,.btn{background:var(--gold);color:#111;border:0;border-radius:8px;padding:8px 12px;font-weight:600;cursor:pointer;font-size:12px}
button.ghost{background:transparent;color:var(--muted);border:1px solid var(--line)}
button.danger{background:var(--danger);color:#fff}
button:disabled{opacity:.45;cursor:not-allowed}
input,textarea,select{width:100%;background:#080a0e;border:1px solid var(--line);border-radius:8px;color:var(--text);padding:8px 10px;font-size:13px}
textarea{min-height:120px;resize:vertical;font-family:ui-monospace,monospace;line-height:1.4}
pre{margin:0;max-height:240px;overflow:auto;white-space:pre-wrap;font-size:11px;color:var(--muted)}
.muted{color:var(--muted);font-size:12px}
.item{display:flex;align-items:center;gap:8px;padding:8px 10px;border-radius:8px;border:1px solid transparent;cursor:pointer}
.item:hover{background:rgba(255,255,255,.05);border-color:var(--line)}
.item.active{background:rgba(240,193,75,.12);border-color:rgba(240,193,75,.35)}
.item .name{flex:1;font-size:13px}
.item .meta{font-size:11px;color:var(--muted)}
.split{display:grid;grid-template-columns:220px 1fr;gap:10px;min-height:320px}
@media(max-width:560px){.split{grid-template-columns:1fr}}
.list{background:var(--panel);border:1px solid var(--line);border-radius:12px;overflow:auto;max-height:420px}
.toast{position:fixed;bottom:12px;right:12px;background:var(--panel2);border:1px solid var(--line);padding:8px 12px;border-radius:8px;font-size:12px;opacity:0;transition:.2s;z-index:9}
.toast.show{opacity:1}
.badge{display:inline-block;padding:2px 8px;border-radius:999px;font-size:10px;background:rgba(93,159,212,.2);color:var(--accent)}
table{width:100%;border-collapse:collapse;font-size:12px}
th,td{text-align:left;padding:8px;border-bottom:1px solid var(--line)}
th{color:var(--muted);font-weight:600}
img.preview{max-width:100%;max-height:280px;border-radius:10px;background:#000}
audio,video{width:100%;margin-top:8px}
</style></head><body>
<header class="app"><h1>` + title + `</h1><span class="sub">AlsetOS · organismo app</span></header>
<div class="toast" id="toast"></div>
<script>
const tok=()=>localStorage.getItem('alset_token')||'';
async function api(path,opts={}){
  const h=Object.assign({'Content-Type':'application/json'},opts.headers||{});
  if(tok())h['X-Alset-Token']=tok();
  const r=await fetch(path,Object.assign({},opts,{headers:h}));
  const j=await r.json().catch(()=>({}));
  if(!r.ok)throw new Error(j.error||j.message||r.statusText||String(r.status));
  return j;
}
function toast(m){const t=document.getElementById('toast');if(!t)return;t.textContent=m;t.classList.add('show');setTimeout(()=>t.classList.remove('show'),2200)}
function esc(s){return String(s??'').replace(/[&<>"']/g,c=>({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;',"'":'&#39;'}[c]))}
</script>
<main>` + body + `</main>
</body></html>`
}

// bodyScriptClose: if body already has scripts, shellApp embeds body as main content only.
// We restructure: body argument should NOT include closing - shellApp wraps body in main.
// Fix: shellApp already put body inside main. Scripts must be inside body param.


func appHTMLFiles() string {
	body := `
<div class="split">
  <div class="col">
    <div class="row">
      <button type="button" onclick="go('')">Raíz</button>
      <button type="button" class="ghost" onclick="up()">↑</button>
      <button type="button" class="ghost" onclick="refresh()">↻</button>
    </div>
    <div class="list" id="list"></div>
    <div class="card">
      <h3>Nuevo</h3>
      <input id="newName" placeholder="nombre.txt o carpeta"/>
      <div class="row" style="margin-top:8px">
        <button type="button" onclick="createFile()">Archivo</button>
        <button type="button" class="ghost" onclick="createDir()">Carpeta</button>
      </div>
    </div>
  </div>
  <div class="col">
    <div class="card">
      <div class="row"><strong id="curPath" class="muted">/</strong>
        <button type="button" class="ghost" id="btnSave" disabled onclick="save()">Guardar</button>
        <button type="button" class="danger" id="btnDel" disabled onclick="del()">Eliminar</button>
      </div>
      <textarea id="editor" placeholder="Selecciona un archivo de texto…" disabled></textarea>
    </div>
  </div>
</div>
<script>
let cur='', selected=null, isDir=false;
async function refresh(){
  const j=await api('/v1/fs/list?path='+encodeURIComponent(cur));
  const list=document.getElementById('list'); list.innerHTML='';
  document.getElementById('curPath').textContent='/'+(cur||'');
  (j.entries||j.items||[]).forEach(e=>{
    const name=e.name||e;
    const dir=!!(e.dir||e.is_dir);
    const el=document.createElement('div');
    el.className='item'+(selected===name?' active':'');
    el.innerHTML='<span class="name">'+(dir?'📁 ':'📄 ')+esc(name)+'</span><span class="meta">'+(dir?'dir':'file')+'</span>';
    el.onclick=()=>select(name,dir);
    el.ondblclick=()=>{if(dir)go(join(cur,name)); else select(name,false)};
    list.appendChild(el);
  });
}
function join(a,b){return a?a.replace(/\/$/,'')+'/'+b:b}
async function go(p){cur=p||'';selected=null;isDir=false;document.getElementById('editor').value='';document.getElementById('editor').disabled=true;document.getElementById('btnSave').disabled=true;document.getElementById('btnDel').disabled=true;await refresh()}
async function up(){
  if(!cur)return go('');
  const parts=cur.split('/').filter(Boolean);parts.pop();
  return go(parts.join('/'));
}
async function select(name,dir){
  selected=name;isDir=dir;document.getElementById('btnDel').disabled=false;
  if(dir){document.getElementById('editor').value='(carpeta)';document.getElementById('editor').disabled=true;document.getElementById('btnSave').disabled=true;await refresh();return}
  const path=join(cur,name);
  try{
    const j=await api('/v1/fs/read?path='+encodeURIComponent(path));
    document.getElementById('editor').disabled=false;
    document.getElementById('editor').value=j.content??'';
    document.getElementById('btnSave').disabled=false;
  }catch(e){toast(e.message);document.getElementById('editor').disabled=true}
  await refresh();
}
async function save(){
  if(!selected||isDir)return;
  const path=join(cur,selected);
  await api('/v1/fs/write',{method:'POST',body:JSON.stringify({path,content:document.getElementById('editor').value})});
  toast('Guardado '+path);
}
async function del(){
  if(!selected)return;
  if(!confirm('¿Eliminar '+selected+'?'))return;
  const path=join(cur,selected);
  await api('/v1/fs/delete',{method:'POST',body:JSON.stringify({path})});
  selected=null;document.getElementById('editor').value='';toast('Eliminado');await refresh();
}
async function createFile(){
  const n=document.getElementById('newName').value.trim();if(!n)return;
  const path=join(cur,n);
  await api('/v1/fs/write',{method:'POST',body:JSON.stringify({path,content:''})});
  document.getElementById('newName').value='';await refresh();await select(n,false);toast('Creado '+n);
}
async function createDir(){
  const n=document.getElementById('newName').value.trim();if(!n)return;
  await api('/v1/fs/mkdir',{method:'POST',body:JSON.stringify({path:join(cur,n)})});
  document.getElementById('newName').value='';await refresh();toast('Carpeta '+n);
}
refresh().catch(e=>toast(e.message));
</script>`
	return shellApp("Archivos", body)
}

func appHTMLDocReader() string {
	body := `
<div class="split">
  <div class="col">
    <div class="card">
      <h3>Biblioteca (docs/)</h3>
      <div class="row">
        <button type="button" onclick="listDocs()">Actualizar</button>
        <button type="button" class="ghost" onclick="newDoc()">Nuevo .txt</button>
      </div>
      <div class="list" id="list" style="margin-top:8px;max-height:360px"></div>
    </div>
  </div>
  <div class="col">
    <div class="card">
      <div class="row">
        <input id="docName" placeholder="documento.txt" style="flex:1"/>
        <button type="button" onclick="saveDoc()">Guardar</button>
        <button type="button" class="danger" onclick="delDoc()">Eliminar</button>
      </div>
      <textarea id="editor" style="min-height:340px;margin-top:8px" placeholder="Escribe o abre un documento…"></textarea>
      <p class="muted">CRUD sobre archivos de texto en la carpeta docs del data dir. Formatos binarios (docx/pdf) se listan; el contenido binario no se edita aquí.</p>
    </div>
  </div>
</div>
<script>
const ROOT='docs';
async function ensure(){try{await api('/v1/fs/mkdir',{method:'POST',body:JSON.stringify({path:ROOT})})}catch(e){}}
async function listDocs(){
  await ensure();
  const j=await api('/v1/fs/list?path='+encodeURIComponent(ROOT));
  const list=document.getElementById('list');list.innerHTML='';
  (j.entries||[]).forEach(e=>{
    const name=e.name||e;
    const el=document.createElement('div');el.className='item';
    el.innerHTML='<span class="name">📄 '+esc(name)+'</span>';
    el.onclick=()=>openDoc(name);
    list.appendChild(el);
  });
}
async function openDoc(name){
  document.getElementById('docName').value=name;
  const j=await api('/v1/fs/read?path='+encodeURIComponent(ROOT+'/'+name));
  document.getElementById('editor').value=j.content??'';
  toast('Abierto '+name);
}
async function saveDoc(){
  let name=document.getElementById('docName').value.trim()||'sin-titulo.txt';
  if(!/\.[a-z0-9]+$/i.test(name))name+='.txt';
  document.getElementById('docName').value=name;
  await ensure();
  await api('/v1/fs/write',{method:'POST',body:JSON.stringify({path:ROOT+'/'+name,content:document.getElementById('editor').value})});
  toast('Guardado');await listDocs();
}
async function delDoc(){
  const name=document.getElementById('docName').value.trim();if(!name)return;
  if(!confirm('¿Eliminar '+name+'?'))return;
  await api('/v1/fs/delete',{method:'POST',body:JSON.stringify({path:ROOT+'/'+name})});
  document.getElementById('editor').value='';document.getElementById('docName').value='';
  toast('Eliminado');await listDocs();
}
async function newDoc(){
  document.getElementById('docName').value='nuevo-'+Date.now()+'.txt';
  document.getElementById('editor').value='';
  document.getElementById('editor').focus();
}
listDocs().catch(e=>toast(e.message));
</script>`
	return shellApp("Documentos", body)
}

func appHTMLNotes() string {
	body := `
<div class="split">
  <div class="list" id="list"></div>
  <div class="col">
    <div class="card">
      <div class="row">
        <input id="title" placeholder="Título de la nota" style="flex:1"/>
        <button type="button" onclick="save()">Guardar</button>
        <button type="button" class="ghost" onclick="create()">Nueva</button>
        <button type="button" class="danger" onclick="remove()">Eliminar</button>
      </div>
      <textarea id="body" style="min-height:300px;margin-top:8px" placeholder="Contenido…"></textarea>
      <p class="muted" id="meta"></p>
    </div>
  </div>
</div>
<script>
const DIR='notes';
let current=null;
async function ensure(){try{await api('/v1/fs/mkdir',{method:'POST',body:JSON.stringify({path:DIR})})}catch(e){}}
async function refresh(){
  await ensure();
  const j=await api('/v1/fs/list?path='+encodeURIComponent(DIR));
  const list=document.getElementById('list');list.innerHTML='';
  (j.entries||[]).filter(e=>String(e.name||e).endsWith('.json')).forEach(e=>{
    const name=e.name||e;
    const el=document.createElement('div');el.className='item'+(current===name?' active':'');
    el.innerHTML='<span class="name">📝 '+esc(name.replace(/\.json$/,''))+'</span>';
    el.onclick=()=>load(name);
    list.appendChild(el);
  });
}
async function load(name){
  current=name;
  const j=await api('/v1/fs/read?path='+encodeURIComponent(DIR+'/'+name));
  let data={};try{data=JSON.parse(j.content||'{}')}catch(e){data={title:name,body:j.content}}
  document.getElementById('title').value=data.title||'';
  document.getElementById('body').value=data.body||'';
  document.getElementById('meta').textContent='Archivo: '+name+(data.updated?' · '+data.updated:'');
  await refresh();
}
async function save(){
  const title=document.getElementById('title').value.trim()||'Sin título';
  if(!current)current=slug(title)+'.json';
  const payload={title,body:document.getElementById('body').value,updated:new Date().toISOString()};
  await ensure();
  await api('/v1/fs/write',{method:'POST',body:JSON.stringify({path:DIR+'/'+current,content:JSON.stringify(payload,null,2)})});
  toast('Nota guardada');await refresh();
}
async function create(){current=null;document.getElementById('title').value='';document.getElementById('body').value='';document.getElementById('meta').textContent='Nueva nota';}
async function remove(){
  if(!current)return;if(!confirm('¿Eliminar nota?'))return;
  await api('/v1/fs/delete',{method:'POST',body:JSON.stringify({path:DIR+'/'+current})});
  current=null;document.getElementById('title').value='';document.getElementById('body').value='';
  toast('Eliminada');await refresh();
}
function slug(s){return s.toLowerCase().replace(/[^a-z0-9]+/g,'-').replace(/^-|-$/g,'').slice(0,40)||('nota-'+Date.now())}
refresh().catch(e=>toast(e.message));
</script>`
	return shellApp("Notas", body)
}

func appHTMLOrgManager() string {
	body := `
<div class="row">
  <button type="button" onclick="load('')">Todos</button>
  <button type="button" class="ghost" onclick="load('app')">Apps</button>
  <button type="button" class="ghost" onclick="load('volume')">Volúmenes</button>
  <button type="button" class="ghost" onclick="load('network')">Red</button>
  <button type="button" class="ghost" onclick="load('ipfs')">IPFS</button>
  <button type="button" class="ghost" onclick="load('user')">Usuarios</button>
</div>
<div class="card">
  <table>
    <thead><tr><th>Nombre</th><th>Tipo</th><th>Estado</th><th>RootCID</th><th></th></tr></thead>
    <tbody id="tb"></tbody>
  </table>
</div>
<div class="card"><h3>Detalle</h3><pre id="detail">Selecciona un organismo</pre></div>
<script>
async function load(kind){
  const q=kind?('?kind='+encodeURIComponent(kind)):'';
  const j=await api('/v1/organisms'+q);
  const tb=document.getElementById('tb');tb.innerHTML='';
  (j.organisms||[]).forEach(o=>{
    const tr=document.createElement('tr');
    tr.innerHTML='<td>'+esc(o.name)+'</td><td><span class="badge">'+esc(o.kind)+'</span></td><td>'+esc(o.state)+'</td><td class="muted">'+esc((o.rootcid||'').slice(0,28))+'</td><td><button type="button" class="ghost" data-id="'+esc(o.id)+'">Ver</button></td>';
    tr.querySelector('button').onclick=async()=>{
      const d=await api('/v1/organisms/get?id='+encodeURIComponent(o.id));
      document.getElementById('detail').textContent=JSON.stringify(d.organism||d,null,2);
    };
    tb.appendChild(tr);
  });
}
load('').catch(e=>toast(e.message));
</script>`
	return shellApp("Organismos", body)
}

func appHTMLAccounts() string {
	body := `
<div class="split">
  <div class="card">
    <h3>Sesión</h3>
    <input id="name" value="Master" placeholder="usuario"/>
    <input id="pass" type="password" value="master" style="margin-top:6px" placeholder="clave"/>
    <div class="row" style="margin-top:8px">
      <button type="button" onclick="login()">Entrar</button>
      <button type="button" class="ghost" onclick="me()">Mi sesión</button>
    </div>
    <pre id="sess" style="margin-top:10px"></pre>
  </div>
  <div class="card">
    <h3>Cuentas (Master)</h3>
    <div class="row">
      <button type="button" onclick="listAcc()">Listar</button>
    </div>
    <table style="margin-top:8px"><thead><tr><th>ID</th><th>Nombre</th><th>Rol</th><th>Organismo</th></tr></thead>
    <tbody id="tb"></tbody></table>
    <p class="muted">Cuenta por defecto <strong>Master</strong> / master — control total del sistema. Cada cuenta es un organismo.</p>
  </div>
</div>
<script>
async function login(){
  const j=await api('/v1/auth/login',{method:'POST',body:JSON.stringify({name:name.value,pass:pass.value})});
  if(j.token)localStorage.setItem('alset_token',j.token);
  sess.textContent=JSON.stringify(j,null,2);toast('Sesión iniciada');
}
async function me(){
  sess.textContent=JSON.stringify(await api('/v1/auth/me'),null,2);
}
async function listAcc(){
  const j=await api('/v1/auth/accounts');
  const tb=document.getElementById('tb');tb.innerHTML='';
  (j.accounts||[]).forEach(a=>{
    const tr=document.createElement('tr');
    tr.innerHTML='<td>'+esc(a.id)+'</td><td>'+esc(a.name)+'</td><td>'+esc(a.role)+'</td><td class="muted">'+esc(a.organism_cid||'')+'</td>';
    tb.appendChild(tr);
  });
}
me().catch(()=>{});listAcc().catch(()=>{});
</script>`
	return shellApp("Cuentas", body)
}

func appHTMLIPFS() string {
	body := `
<div class="split">
  <div class="card">
    <h3>Añadir contenido</h3>
    <input id="n" placeholder="nombre" value="nota.txt"/>
    <textarea id="c" style="margin-top:6px" placeholder="contenido…"></textarea>
    <div class="row" style="margin-top:8px">
      <button type="button" onclick="add()">Guardar (CID)</button>
      <button type="button" class="ghost" onclick="list()">Actualizar lista</button>
    </div>
  </div>
  <div class="card">
    <h3>Almacén</h3>
    <div class="list" id="list" style="max-height:280px"></div>
    <pre id="out" style="margin-top:8px"></pre>
  </div>
</div>
<script>
async function add(){
  const j=await api('/v1/ipfs/add',{method:'POST',body:JSON.stringify({name:n.value,content:c.value})});
  out.textContent=JSON.stringify(j,null,2);toast('CID '+j.cid);await list();
}
async function list(){
  const j=await api('/v1/ipfs/list');
  const list=document.getElementById('list');list.innerHTML='';
  (j.items||[]).forEach(it=>{
    const el=document.createElement('div');el.className='item';
    el.innerHTML='<span class="name">⬡ '+esc(it.name||it.cid)+'</span><span class="meta">'+esc((it.cid||'').slice(0,20))+'</span>';
    el.onclick=async()=>{
      const g=await api('/v1/ipfs/get?cid='+encodeURIComponent(it.cid));
      out.textContent=typeof g.content==='string'?g.content:JSON.stringify(g,null,2);
    };
    list.appendChild(el);
  });
}
list().catch(e=>toast(e.message));
</script>`
	return shellApp("IPFS Store", body)
}

func appHTMLAudio() string {
	body := `
<div class="split">
  <div class="col">
    <div class="card">
      <h3>Explorar medios</h3>
      <select id="rootSel"></select>
      <div class="list" id="browser" style="margin-top:8px;max-height:280px"></div>
    </div>
  </div>
  <div class="card">
    <h3>Reproductor</h3>
    <p class="muted" id="now">Ningún archivo</p>
    <audio id="a" controls></audio>
  </div>
</div>
<script>
window.__mediaPick=(url,path,name)=>{document.getElementById('a').src=url;document.getElementById('a').play();document.getElementById('now').textContent=name||path;toast(name)};

async function mediaRoots(){return (await api('/v1/media/roots')).roots||[]}
let mPath='';
async function mediaGo(path,filter){
  mPath=path||'';
  const j=await api('/v1/media/list?path='+encodeURIComponent(mPath)+'&filter='+(filter||'all'));
  const box=document.getElementById('browser');
  box.innerHTML='';
  const up=document.createElement('div');up.className='item';up.innerHTML='<span class="name">↑ Subir</span>';
  up.onclick=()=>{
    if(!mPath||mPath==='.')return;
    const parts=mPath.replace(/\/$/,'').split('/');parts.pop();
    mediaGo(parts.join('/')||'',filter);
  };
  box.appendChild(up);
  (j.entries||[]).forEach(e=>{
    const el=document.createElement('div');el.className='item';
    el.innerHTML='<span class="name">'+(e.dir?'📁 ':'📄 ')+esc(e.name)+'</span>';
    el.onclick=()=>{
      if(e.dir)mediaGo(e.path,filter);
      else if(e.url){window.__mediaPick&&window.__mediaPick(e.url,e.path,e.name)}
    };
    box.appendChild(el);
  });
}
async function mediaInit(filter){
  const roots=await mediaRoots();
  const sel=document.getElementById('rootSel');
  sel.innerHTML='';
  roots.forEach(r=>{
    const o=document.createElement('option');o.value=r.path;o.textContent=r.label;sel.appendChild(o);
  });
  sel.onchange=()=>mediaGo(sel.value,filter);
  if(roots[0])mediaGo(roots[0].path,filter);
}

mediaInit('audio').catch(e=>toast(e.message));
</script>`
	return shellApp("Audio", body)
}

func appHTMLVideo() string {
	body := `
<div class="split">
  <div class="col">
    <div class="card">
      <h3>Explorar</h3>
      <select id="rootSel"></select>
      <div class="list" id="browser" style="margin-top:8px;max-height:280px"></div>
    </div>
  </div>
  <div class="card">
    <h3>Reproductor</h3>
    <p class="muted" id="now">Ningún archivo</p>
    <video id="v" controls></video>
  </div>
</div>
<script>
window.__mediaPick=(url,path,name)=>{document.getElementById('v').src=url;document.getElementById('v').play();document.getElementById('now').textContent=name||path;toast(name)};

async function mediaRoots(){return (await api('/v1/media/roots')).roots||[]}
let mPath='';
async function mediaGo(path,filter){
  mPath=path||'';
  const j=await api('/v1/media/list?path='+encodeURIComponent(mPath)+'&filter='+(filter||'all'));
  const box=document.getElementById('browser');
  box.innerHTML='';
  const up=document.createElement('div');up.className='item';up.innerHTML='<span class="name">↑ Subir</span>';
  up.onclick=()=>{
    if(!mPath||mPath==='.')return;
    const parts=mPath.replace(/\/$/,'').split('/');parts.pop();
    mediaGo(parts.join('/')||'',filter);
  };
  box.appendChild(up);
  (j.entries||[]).forEach(e=>{
    const el=document.createElement('div');el.className='item';
    el.innerHTML='<span class="name">'+(e.dir?'📁 ':'📄 ')+esc(e.name)+'</span>';
    el.onclick=()=>{
      if(e.dir)mediaGo(e.path,filter);
      else if(e.url){window.__mediaPick&&window.__mediaPick(e.url,e.path,e.name)}
    };
    box.appendChild(el);
  });
}
async function mediaInit(filter){
  const roots=await mediaRoots();
  const sel=document.getElementById('rootSel');
  sel.innerHTML='';
  roots.forEach(r=>{
    const o=document.createElement('option');o.value=r.path;o.textContent=r.label;sel.appendChild(o);
  });
  sel.onchange=()=>mediaGo(sel.value,filter);
  if(roots[0])mediaGo(roots[0].path,filter);
}

mediaInit('video').catch(e=>toast(e.message));
</script>`
	return shellApp("Video", body)
}

func appHTMLImage() string {
	body := `
<div class="split">
  <div class="col">
    <div class="card">
      <h3>Explorar imágenes</h3>
      <select id="rootSel"></select>
      <div class="list" id="browser" style="margin-top:8px;max-height:320px"></div>
    </div>
  </div>
  <div class="card">
    <h3>Vista</h3>
    <p class="muted" id="now">Selecciona una imagen</p>
    <img class="preview" id="i" alt=""/>
  </div>
</div>
<script>
window.__mediaPick=(url,path,name)=>{document.getElementById('i').src=url;document.getElementById('now').textContent=name||path;toast(name)};

async function mediaRoots(){return (await api('/v1/media/roots')).roots||[]}
let mPath='';
async function mediaGo(path,filter){
  mPath=path||'';
  const j=await api('/v1/media/list?path='+encodeURIComponent(mPath)+'&filter='+(filter||'all'));
  const box=document.getElementById('browser');
  box.innerHTML='';
  const up=document.createElement('div');up.className='item';up.innerHTML='<span class="name">↑ Subir</span>';
  up.onclick=()=>{
    if(!mPath||mPath==='.')return;
    const parts=mPath.replace(/\/$/,'').split('/');parts.pop();
    mediaGo(parts.join('/')||'',filter);
  };
  box.appendChild(up);
  (j.entries||[]).forEach(e=>{
    const el=document.createElement('div');el.className='item';
    el.innerHTML='<span class="name">'+(e.dir?'📁 ':'📄 ')+esc(e.name)+'</span>';
    el.onclick=()=>{
      if(e.dir)mediaGo(e.path,filter);
      else if(e.url){window.__mediaPick&&window.__mediaPick(e.url,e.path,e.name)}
    };
    box.appendChild(el);
  });
}
async function mediaInit(filter){
  const roots=await mediaRoots();
  const sel=document.getElementById('rootSel');
  sel.innerHTML='';
  roots.forEach(r=>{
    const o=document.createElement('option');o.value=r.path;o.textContent=r.label;sel.appendChild(o);
  });
  sel.onchange=()=>mediaGo(sel.value,filter);
  if(roots[0])mediaGo(roots[0].path,filter);
}

mediaInit('image').catch(e=>toast(e.message));
</script>`
	return shellApp("Imágenes", body)
}

