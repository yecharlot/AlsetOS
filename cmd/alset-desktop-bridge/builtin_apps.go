package main

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type builtinApp struct {
	Name        string
	Title       string
	Glyph       string
	MenuOnly    bool
	HTML        string
}

func (b *bridge) seedBuiltinApps() {
	apps := []builtinApp{
		{Name: "org-manager", Title: "Organismos", Glyph: "◎", HTML: appHTMLOrgManager()},
		{Name: "doc-reader", Title: "Documentos", Glyph: "📄", HTML: appHTMLDocReader()},
		{Name: "audio-player", Title: "Audio", Glyph: "♫", MenuOnly: true, HTML: appHTMLAudio()},
		{Name: "video-player", Title: "Video", Glyph: "▶", MenuOnly: true, HTML: appHTMLVideo()},
		{Name: "image-viewer", Title: "Imágenes", Glyph: "🖼", HTML: appHTMLImage()},
		{Name: "accounts", Title: "Cuentas", Glyph: "👤", HTML: appHTMLAccounts()},
		{Name: "ipfs-store", Title: "IPFS Store", Glyph: "⬡", HTML: appHTMLIPFS()},
		{Name: "files-ui", Title: "Archivos", Glyph: "📁", HTML: appHTMLFiles()},
		{Name: "calculadora", Title: "Calculadora", Glyph: "🔢", HTML: ""}, // uses default calc in deploy
	}
	for _, a := range apps {
		dir := filepath.Join(b.dataDir, "apps", a.Name)
		if _, err := os.Stat(filepath.Join(dir, "index.html")); err == nil {
			continue // don't overwrite user changes
		}
		_ = os.MkdirAll(dir, 0o755)
		html := a.HTML
		if html == "" {
			// calculator from buildAppHTML empty tree path
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
:root{--bg:#0c1018;--panel:#141824;--line:rgba(255,255,255,.08);--gold:#f0c14b;--text:#f2f4f8;--muted:#8b93a7;--accent:#5b9fd4}
*{box-sizing:border-box}body{margin:0;font-family:system-ui,sans-serif;background:var(--bg);color:var(--text);padding:14px}
h1{font-size:16px;color:var(--gold);margin:0 0 12px}
.card{background:var(--panel);border:1px solid var(--line);border-radius:12px;padding:12px;margin-bottom:10px}
button,.btn{background:var(--gold);color:#111;border:0;border-radius:8px;padding:8px 12px;font-weight:600;cursor:pointer;margin:4px 4px 0 0}
button.ghost{background:transparent;color:var(--muted);border:1px solid var(--line)}
pre,textarea,input{width:100%;background:#080a0e;border:1px solid var(--line);border-radius:8px;color:var(--text);padding:8px;font-size:12px}
pre{max-height:220px;overflow:auto;white-space:pre-wrap}
.muted{color:var(--muted);font-size:12px}
.row{display:flex;flex-wrap:wrap;gap:6px;align-items:center}
list-item, .item{display:block;padding:8px;border-radius:8px;border:1px solid transparent;cursor:pointer}
.item:hover{background:rgba(255,255,255,.05);border-color:var(--line)}
</style></head><body>
<h1>` + title + `</h1>
` + body + `
</body></html>`
}

func appHTMLOrgManager() string {
	return shellApp("Administrador de organismos", `
<p class="muted">Todo en AlsetOS es un organismo: apps, volúmenes, red, cuentas, IPFS.</p>
<div class="row">
<button onclick="load('')">Todos</button>
<button class="ghost" onclick="load('app')">Apps</button>
<button class="ghost" onclick="load('volume')">Volúmenes</button>
<button class="ghost" onclick="load('network')">Red</button>
<button class="ghost" onclick="load('ipfs')">IPFS</button>
</div>
<pre id="out">Cargando…</pre>
<script>
async function load(kind){
  const q=kind?('?kind='+kind):'';
  const r=await fetch('/v1/organisms'+q); const j=await r.json();
  document.getElementById('out').textContent=JSON.stringify(j.organisms||j,null,2);
}
load('');
</script>`)
}

func appHTMLDocReader() string {
	return shellApp("Lector de documentos", `
<p class="muted">Abre .txt y previsualiza rutas del data dir. DOCX/PDF: muestra metadatos / texto si es legible.</p>
<input id="path" placeholder="ruta relativa ej. readme.txt"/>
<div class="row"><button onclick="openDoc()">Abrir</button></div>
<pre id="out"></pre>
<script>
async function openDoc(){
  const p=document.getElementById('path').value;
  const r=await fetch('/v1/fs/read?path='+encodeURIComponent(p));
  const j=await r.json();
  document.getElementById('out').textContent=j.content||JSON.stringify(j,null,2);
}
</script>`)
}

func appHTMLAudio() string {
	return shellApp("Reproductor de audio", `
<p class="muted">Reproduce archivos de audio del sistema de archivos (ruta o URL).</p>
<input id="src" placeholder="URL o /apps/.../file.mp3"/>
<audio id="a" controls style="width:100%;margin-top:12px"></audio>
<div class="row"><button onclick="play()">Cargar</button></div>
<script>function play(){const a=document.getElementById('a');a.src=document.getElementById('src').value;a.play();}</script>`)
}

func appHTMLVideo() string {
	return shellApp("Reproductor de video", `
<p class="muted">mp4, webm y formatos soportados por el navegador.</p>
<input id="src" placeholder="URL del video"/>
<video id="v" controls style="width:100%;max-height:360px;margin-top:12px;background:#000"></video>
<div class="row"><button onclick="play()">Cargar</button></div>
<script>function play(){const v=document.getElementById('v');v.src=document.getElementById('src').value;v.play();}</script>`)
}

func appHTMLImage() string {
	return shellApp("Visor de imágenes", `
<p class="muted">jpg png gif webp — URL o ruta servida.</p>
<input id="src" placeholder="URL de imagen"/>
<div class="row"><button onclick="show()">Mostrar</button></div>
<img id="i" style="max-width:100%;margin-top:12px;border-radius:12px" alt=""/>
<script>function show(){document.getElementById('i').src=document.getElementById('src').value}</script>`)
}

func appHTMLAccounts() string {
	return shellApp("Cuentas de usuario", `
<p class="muted">Cuenta por defecto: <strong>Master</strong> / master (rol master).</p>
<div class="card">
<input id="name" value="Master"/><input id="pass" type="password" value="master" style="margin-top:6px"/>
<div class="row"><button onclick="login()">Entrar</button><button class="ghost" onclick="me()">Sesión</button><button class="ghost" onclick="list()">Listar (master)</button></div>
</div>
<pre id="out"></pre>
<script>
async function login(){
  const r=await fetch('/v1/auth/login',{method:'POST',headers:{'Content-Type':'application/json'},
    body:JSON.stringify({name:name.value,pass:pass.value})});
  const j=await r.json(); out.textContent=JSON.stringify(j,null,2);
  if(j.token) localStorage.setItem('alset_token',j.token);
}
async function me(){
  const t=localStorage.getItem('alset_token')||'';
  const r=await fetch('/v1/auth/me',{headers:{'X-Alset-Token':t}}); out.textContent=JSON.stringify(await r.json(),null,2);
}
async function list(){
  const t=localStorage.getItem('alset_token')||'';
  const r=await fetch('/v1/auth/accounts',{headers:{'X-Alset-Token':t}}); out.textContent=JSON.stringify(await r.json(),null,2);
}
</script>`)
}

func appHTMLIPFS() string {
	return shellApp("IPFS Store (nativo)", `
<p class="muted">Almacén content-addressed local (suelo IPFS de Alset). API: /v1/ipfs/*</p>
<textarea id="c" rows="4" placeholder="contenido a guardar"></textarea>
<input id="n" placeholder="nombre" value="nota.txt" style="margin-top:6px"/>
<div class="row"><button onclick="add()">Add</button><button class="ghost" onclick="list()">Listar</button></div>
<pre id="out"></pre>
<script>
async function add(){
  const r=await fetch('/v1/ipfs/add',{method:'POST',headers:{'Content-Type':'application/json'},
    body:JSON.stringify({name:n.value,content:c.value})});
  out.textContent=JSON.stringify(await r.json(),null,2);
}
async function list(){ out.textContent=JSON.stringify(await (await fetch('/v1/ipfs/list')).json(),null,2); }
list();
</script>`)
}

func appHTMLFiles() string {
	return shellApp("Gestor de archivos", `
<p class="muted">Explorador del data dir (organismos tipo file/volume).</p>
<div class="row"><button onclick="go('')">Raíz datos</button><button class="ghost" id="up">↑</button></div>
<pre id="out"></pre>
<script>
let cur='';
async function go(p){
  cur=p||'';
  const j=await (await fetch('/v1/fs/list?path='+encodeURIComponent(cur))).json();
  out.textContent=JSON.stringify(j,null,2);
}
document.getElementById('up').onclick=async()=>{
  const j=await (await fetch('/v1/fs/list?path='+encodeURIComponent(cur))).json();
  go(j.parent===j.root?'':(j.parent||''));
};
go('');
</script>`)
}
