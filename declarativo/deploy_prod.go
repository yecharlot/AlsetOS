package declarativo

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// DeployProduction escribe una app lista para producción bajo appsRoot/name/
// Servida por alset-plataforma studio en /apps/{name}/
func DeployProduction(flow *FlowAOS, appsRoot string) (string, error) {
	if flow == nil || flow.Name == "" {
		return "", fmt.Errorf("flow sin nombre")
	}
	dir := filepath.Join(appsRoot, flow.Name)
	if err := DeployFlow(flow, dir); err != nil {
		return "", err
	}
	// Runtime bootstrap state for production shell
	state := map[string]interface{}{}
	for _, o := range flow.Manifiesto.Organismos {
		for k, v := range o.InitialState {
			state[k] = v
		}
	}
	meta := map[string]interface{}{
		"app": flow.Name, "version": flow.Version, "deployed_at": time.Now().UTC().Format(time.RFC3339),
		"surface": flow.SurfaceOID, "production": true, "source": "lispai-aos",
	}
	mb, _ := json.MarshalIndent(meta, "", "  ")
	_ = os.WriteFile(filepath.Join(dir, "production.json"), mb, 0o644)
	sb, _ := json.MarshalIndent(state, "", "  ")
	_ = os.WriteFile(filepath.Join(dir, "initial_state.json"), sb, 0o644)
	// Production SPA that talks to /api/apps/{name}/action
	uiJSON, _ := json.Marshal(flow.Surfaces)
	prod := productionAppHTML(flow.Name, flow.SurfaceOID, string(uiJSON))
	if err := os.WriteFile(filepath.Join(dir, "app.html"), []byte(prod), 0o644); err != nil {
		return "", err
	}
	// default index redirects to app
	idx := fmt.Sprintf(`<!DOCTYPE html><html><head><meta charset="utf-8"/>
<meta http-equiv="refresh" content="0;url=app.html"/>
<title>%s</title></head><body>
<p><a href="app.html">Abrir %s (producción)</a></p>
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
<link rel="manifest" href="manifest.webmanifest"/>
<style>
:root{--bg:#0f1419;--card:#1a2332;--ink:#e8eef8;--accent:#F4B400;--muted:#8b9bb4}
*{box-sizing:border-box}body{margin:0;font-family:system-ui,sans-serif;background:var(--bg);color:var(--ink);min-height:100vh}
header{padding:.75rem 1rem;border-bottom:1px solid #243044;display:flex;justify-content:space-between;align-items:center}
#root{padding:1rem;max-width:480px;margin:0 auto}
.col{display:flex;flex-direction:column;gap:.6rem}.row{display:flex;gap:.5rem;flex-wrap:wrap}
button{background:var(--accent);border:0;padding:.55rem .85rem;border-radius:10px;font-weight:700;cursor:pointer}
.badge{background:var(--card);padding:.4rem .7rem;border-radius:999px;border:1px solid #243044}
#status{font-size:.75rem;color:var(--muted)}
</style></head><body>
<header><strong>%s</strong><span id="status">listo</span></header>
<div id="root"></div>
<script>
const APP=%q, SURFACE=%q, SURFACES=%s;
let STATE={};
async function action(emit){
  document.getElementById('status').textContent='…';
  const r=await fetch('/api/apps/'+encodeURIComponent(APP)+'/action',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({emit})});
  const j=await r.json();
  if(j.state) STATE=j.state;
  document.getElementById('status').textContent=j.ok?'ok':'error';
  paint();
}
function val(bind){ if(!bind)return '·'; const v=STATE[bind]; return v==null?bind:v; }
function render(n){
  if(!n)return document.createTextNode('');
  const t=(n.type||'').toLowerCase();
  if(t==='column'||t==='panel'){const d=document.createElement('div');d.className='col';(n.kids||[]).forEach(k=>d.appendChild(render(k)));return d;}
  if(t==='row'){const d=document.createElement('div');d.className='row';(n.kids||[]).forEach(k=>d.appendChild(render(k)));return d;}
  if(t==='label'){const s=document.createElement('div');s.textContent=n.text||'';return s;}
  if(t==='badge'){const s=document.createElement('span');s.className='badge';s.textContent=val(n.bind);s.dataset.bind=n.bind||'';return s;}
  if(t==='button'){const b=document.createElement('button');b.textContent=n.text||'OK';b.onclick=()=>action(n.emit||'');return b;}
  if(t==='field'){const i=document.createElement('input');i.placeholder=n.text||'';return i;}
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
