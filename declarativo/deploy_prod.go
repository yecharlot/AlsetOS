package declarativo

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// DeployProduction escribe una app lista para producción bajo appsRoot/name/
// Cada nombre es una instancia aislada (carpeta + source.aos + estado).
func DeployProduction(flow *FlowAOS, appsRoot string) (string, error) {
	if flow == nil || flow.Name == "" {
		return "", fmt.Errorf("flow sin nombre")
	}
	dir := filepath.Join(appsRoot, flow.Name)
	if err := DeployFlow(flow, dir); err != nil {
		return "", err
	}
	state := map[string]interface{}{}
	for _, o := range flow.Manifiesto.Organismos {
		for k, v := range o.InitialState {
			state[k] = v
		}
	}
	meta := map[string]interface{}{
		"app": flow.Name, "version": flow.Version,
		"deployed_at": time.Now().UTC().Format(time.RFC3339),
		"surface": flow.SurfaceOID, "production": true, "source": "lispai-aos",
		"isolated": true,
	}
	mb, _ := json.MarshalIndent(meta, "", "  ")
	_ = os.WriteFile(filepath.Join(dir, "production.json"), mb, 0o644)
	sb, _ := json.MarshalIndent(state, "", "  ")
	_ = os.WriteFile(filepath.Join(dir, "initial_state.json"), sb, 0o644)
	uiJSON, _ := json.Marshal(flow.Surfaces)
	prod := productionAppHTML(flow.Name, flow.SurfaceOID, string(uiJSON))
	if err := os.WriteFile(filepath.Join(dir, "app.html"), []byte(prod), 0o644); err != nil {
		return "", err
	}
	idx := fmt.Sprintf(`<!DOCTYPE html><html><head><meta charset="utf-8"/>
<meta http-equiv="refresh" content="0;url=app.html"/>
<title>%s</title></head><body>
<p><a href="app.html">Abrir %s</a></p>
<p><a href="mobile/">Mobile</a> · <a href="tablet/">Tablet</a> · <a href="desktop/">Desktop</a></p>
</body></html>`, flow.Name, flow.Name)
	_ = os.WriteFile(filepath.Join(dir, "index.html"), []byte(idx), 0o644)
	return dir, nil
}

func productionAppHTML(app, surface, uiJSON string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="es"><head>
<meta charset="utf-8"/><meta name="viewport" content="width=device-width,initial-scale=1"/>
<meta name="theme-color" content="#F4B400"/>
<title>%s</title>
<style>
:root{--bg:#0f1419;--card:#1a2332;--ink:#e8eef8;--accent:#F4B400;--muted:#8b9bb4}
*{box-sizing:border-box}body{margin:0;font-family:system-ui,sans-serif;background:var(--bg);color:var(--ink);min-height:100vh}
header{padding:.75rem 1rem;border-bottom:1px solid #243044;display:flex;justify-content:space-between}
#root{padding:1rem;max-width:560px;margin:0 auto;position:relative}
.col{display:flex;flex-direction:column;gap:.6rem}.row{display:flex;gap:.5rem;flex-wrap:wrap}
button,.tab{background:var(--accent);border:0;padding:.5rem .8rem;border-radius:10px;font-weight:700;cursor:pointer;color:#111}
.tab{background:#243044;color:var(--ink)}.tab.on{background:var(--accent);color:#111}
.badge{background:var(--card);padding:.35rem .65rem;border-radius:999px;border:1px solid #243044;display:inline-block}
.splash{text-align:center;padding:2rem 1rem;animation:fade .8s ease}
@keyframes fade{from{opacity:0;transform:translateY(12px)}to{opacity:1;transform:none}}
.blob{width:120px;height:120px;margin:0 auto;border-radius:40%% 60%% 55%% 45%%;background:linear-gradient(135deg,#F4B400,#5b9cf5);animation:blob 4s ease-in-out infinite}
@keyframes blob{0%%,100%%{border-radius:40%% 60%% 55%% 45%%}50%%{border-radius:55%% 45%% 40%% 60%%}}
.gradient{min-height:80px;border-radius:12px;padding:1rem;background:linear-gradient(135deg,#1a2332,#3d2a5a)}
.drawer{position:fixed;top:0;bottom:0;width:78%%;max-width:320px;background:#151c28;padding:1rem;z-index:20;transform:translateX(-105%%);transition:transform .25s}
.drawer.open{transform:none}.drawer.right{right:0;left:auto;transform:translateX(105%%)}.drawer.right.open{transform:none}
.backdrop{position:fixed;inset:0;background:rgba(0,0,0,.45);z-index:15;display:none}.backdrop.show{display:block}
.tabs{display:flex;gap:4px;margin-bottom:.5rem;flex-wrap:wrap}
.bottom-tabs{position:fixed;bottom:0;left:0;right:0;display:flex;background:#151c28;border-top:1px solid #243044;padding:.35rem}
.bottom-tabs .tab{flex:1;border-radius:8px;margin:2px;font-size:.75rem}
.stack>*{display:none}.stack>.on{display:block}
img.ui-img{max-width:100%%;border-radius:12px;display:block}
.qr-box{font-family:monospace;background:#111;padding:1rem;border-radius:8px;word-break:break-all}
#status{font-size:.75rem;color:var(--muted)}
input[type=file]{font-size:.8rem;color:var(--muted)}
.doclist li{margin:.35rem 0}
</style></head><body>
<header><strong>%s</strong><span id="status">listo</span></header>
<div class="backdrop" id="bd"></div>
<div id="root"></div>
<script>
const APP=%q, SURFACE=%q, SURFACES=%s;
let STATE={}; let tabSel={};
async function action(emit){
  if(!emit) return;
  document.getElementById('status').textContent='…';
  const r=await fetch('/api/apps/'+encodeURIComponent(APP)+'/action',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({emit})});
  const j=await r.json();
  if(j.state) STATE=j.state;
  document.getElementById('status').textContent=j.ok?'ok':'error';
  paint();
}
function val(bind){ if(!bind)return '·'; const v=STATE[bind]; return v==null?bind:String(v); }
function render(n, path){
  if(!n) return document.createTextNode('');
  const t=(n.type||'').toLowerCase();
  path = path||'root';
  if(t==='column'||t==='panel'||t==='row'){
    const d=document.createElement('div'); d.className=t==='row'?'row':'col';
    (n.kids||[]).forEach((k,i)=>d.appendChild(render(k, path+'.'+i))); return d;
  }
  if(t==='label'){ const s=document.createElement('div'); s.textContent=n.text||''; return s; }
  if(t==='badge'){ const s=document.createElement('span'); s.className='badge'; s.textContent=val(n.bind); return s; }
  if(t==='button'){ const b=document.createElement('button'); b.textContent=n.text||'OK'; b.onclick=()=>action(n.emit||''); return b; }
  if(t==='field'){ const i=document.createElement('input'); i.placeholder=n.text||''; i.style.width='100%%'; i.style.padding='8px'; return i; }
  if(t==='splash'){ const d=document.createElement('div'); d.className='splash'; const h=document.createElement('h2'); h.textContent=n.text||'Alset'; d.appendChild(h); (n.kids||[]).forEach(k=>d.appendChild(render(k))); return d; }
  if(t==='blob'){ const d=document.createElement('div'); d.className='blob'; return d; }
  if(t==='gradient'){ const d=document.createElement('div'); d.className='gradient'; if(n.src) d.style.background=n.src; d.textContent=n.text||''; (n.kids||[]).forEach(k=>d.appendChild(render(k))); return d; }
  if(t==='image'){ const i=document.createElement('img'); i.className='ui-img'; i.src=n.src||''; i.alt=n.text||''; if((n.style||'').includes('blur')) i.style.filter='blur(2px)'; return i; }
  if(t==='file'){ const w=document.createElement('div'); const i=document.createElement('input'); i.type='file'; i.accept=n.src||'*/*';
    i.onchange=()=>{ const f=i.files&&i.files[0]; if(f) action((n.emit||'file.selected')+'='+f.name); }; w.appendChild(i); return w; }
  if(t==='qr'){ const d=document.createElement('div'); d.className='qr-box'; d.textContent=val(n.bind)||n.text||'QR'; return d; }
  if(t==='doclist'){ const ul=document.createElement('ul'); ul.className='doclist'; const raw=val(n.bind); (String(raw).split('|').filter(Boolean)).forEach(x=>{ const li=document.createElement('li'); li.textContent=x; ul.appendChild(li); }); return ul; }
  if(t==='drawer'){
    const wrap=document.createElement('div');
    const btn=document.createElement('button'); btn.textContent=n.text||'Menu';
    const dr=document.createElement('div'); dr.className='drawer'+(String(n.style||'').includes('right')?' right':'');
    (n.kids||[]).forEach(k=>dr.appendChild(render(k)));
    btn.onclick=()=>{ dr.classList.add('open'); document.getElementById('bd').classList.add('show'); };
    document.getElementById('bd').onclick=()=>{ document.querySelectorAll('.drawer').forEach(x=>x.classList.remove('open')); document.getElementById('bd').classList.remove('show'); };
    wrap.appendChild(btn); wrap.appendChild(dr); return wrap;
  }
  if(t==='tabs'||t==='top-tabs'||t==='bottom-tabs'||t==='stack'){
    const kids=n.kids||[];
    const key=path;
    if(tabSel[key]==null) tabSel[key]=0;
    const wrap=document.createElement('div');
    if(t!=='stack'){
      const bar=document.createElement('div'); bar.className=t==='bottom-tabs'?'bottom-tabs':'tabs';
      kids.forEach((k,i)=>{ const b=document.createElement('button'); b.className='tab'+(tabSel[key]===i?' on':''); b.textContent=k.text||('Tab '+(i+1));
        b.onclick=()=>{ tabSel[key]=i; paint(); }; bar.appendChild(b); });
      if(t==='bottom-tabs') document.body.appendChild(bar); else wrap.appendChild(bar);
    }
    const body=document.createElement('div'); body.className='stack';
    kids.forEach((k,i)=>{ const pane=document.createElement('div'); if(tabSel[key]===i) pane.className='on';
      (k.kids||[k]).forEach(c=>{ if(c===k && k.type==='tab') (k.kids||[]).forEach(cc=>pane.appendChild(render(cc))); else if(c!==k||k.type!=='tab') pane.appendChild(render(c)); });
      // simpler: render kid content
      pane.innerHTML='';
      if((k.type||'').toLowerCase()==='tab') (k.kids||[]).forEach(c=>pane.appendChild(render(c)));
      else pane.appendChild(render(k));
      if(tabSel[key]===i) pane.classList.add('on');
      body.appendChild(pane);
    });
    wrap.appendChild(body); return wrap;
  }
  return document.createTextNode(n.type||'');
}
function paint(){
  const root=document.getElementById('root'); root.innerHTML='';
  const tree=SURFACES[SURFACE]||Object.values(SURFACES)[0];
  root.appendChild(render(tree));
}
fetch('/api/apps/'+encodeURIComponent(APP)+'/state').then(r=>r.json()).then(j=>{if(j.state)STATE=j.state;paint();}).catch(paint);
</script></body></html>`, app, app, app, surface, uiJSON)
}
