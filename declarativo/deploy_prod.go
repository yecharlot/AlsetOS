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
	idx := fmt.Sprintf(`<!DOCTYPE html><html lang="es"><head><meta charset="utf-8"/>
<meta name="viewport" content="width=device-width,initial-scale=1"/>
<meta http-equiv="refresh" content="0;url=app.html"/>
<title>%s</title>
<style>body{margin:0;background:#0f1419}</style>
</head><body></body></html>`, flow.Name)
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
%s
body{margin:0;background:var(--alset-bg);color:var(--alset-ink);min-height:100vh}
#root{padding:1rem;max-width:560px;margin:0 auto;min-height:100vh}
</style></head><body>
<div id="root" class="alset-ui"></div>
<script>
const SURFACES = %s;
const SURFACE = %q;
%s
function val(bind, state){ if(!bind) return '·'; const v=(state||{})[bind]; return v==null?bind:v; }
function render(n, state){
  if(!n) return document.createTextNode('');
  const t=(n.type||'').toLowerCase();
  let el;
  if(t==='column'||t==='panel'){ el=document.createElement('div'); el.className='col'; (n.kids||[]).forEach(k=>el.appendChild(render(k,state))); }
  else if(t==='row'){ el=document.createElement('div'); el.className='row'; (n.kids||[]).forEach(k=>el.appendChild(render(k,state))); }
  else if(t==='label'){ el=document.createElement('div'); el.textContent=n.text||''; }
  else if(t==='badge'){ el=document.createElement('span'); el.className='badge'; el.textContent=val(n.bind,state); }
  else if(t==='button'){
    el=document.createElement('button'); el.textContent=n.text||'OK';
    el.onclick=async()=>{
      const emit=n.emit||'';
      try{
        const r=await fetch('api/action',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({emit:emit})});
        const j=await r.json();
        if(j.state){ window.__state=j.state; const root=document.getElementById('root'); root.innerHTML=''; root.appendChild(render(tree,j.state)); }
      }catch(e){ console.warn(e); }
    };
  }
  else if(t==='splash'){ el=document.createElement('div'); el.className='rt-splash splash'; el.textContent=n.text||'Splash'; }
  else if(t==='gradient'){ el=document.createElement('div'); el.className='gradient'; el.textContent=n.text||''; }
  else if(t==='field'){ el=document.createElement('input'); el.className='rt-input'; el.placeholder=n.text||''; }
  else if(t==='blob'){ el=document.createElement('div'); el.style.cssText='width:64px;height:64px;border-radius:40%%;background:linear-gradient(135deg,#f4b400,#5b9cf5)'; }
  else if(t==='image'){ el=document.createElement('img'); el.src=n.src||''; el.style.maxWidth='100%%'; }
  else if(t==='drawer'||t==='tabs'||t==='stack'||t==='top-tabs'||t==='bottom-tabs'){
    el=document.createElement('div'); el.className='col';
    const lab=document.createElement('div'); lab.className='muted'; lab.textContent=t; el.appendChild(lab);
    (n.kids||[]).forEach(k=>el.appendChild(render(k,state)));
  } else { el=document.createElement('div'); el.textContent=n.type||''; }
  return applyAlsetProps(el, n);
}
const tree = SURFACES[SURFACE] || Object.values(SURFACES)[0];
window.__state = {};
document.getElementById('root').appendChild(render(tree, window.__state));
</script></body></html>`, app, SharedUISkin, uiJSON, surface, ApplyPropsJS)
}
