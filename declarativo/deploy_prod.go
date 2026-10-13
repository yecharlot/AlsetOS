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
const APP_NAME = %q;
%s
%s
window.__alsetAction = async function(emit){
  const urls = [
    '/api/apps/' + encodeURIComponent(APP_NAME) + '/action',
    'api/action',
    './api/action'
  ];
  let lastErr = null;
  for (const u of urls) {
    try {
      const r = await fetch(u, { method:'POST', headers:{'Content-Type':'application/json'}, body: JSON.stringify({emit: emit, app: APP_NAME}) });
      const txt = await r.text();
      let j; try { j = JSON.parse(txt); } catch(e) { lastErr = e; continue; }
      if (!r.ok) { lastErr = new Error(j.error||txt); continue; }
      return j;
    } catch(e) { lastErr = e; }
  }
  console.warn('action failed', lastErr);
  return { ok:false };
};
const tree = SURFACES[SURFACE] || Object.values(SURFACES)[0];
window.__tree = tree;
window.__state = {};
document.getElementById('root').appendChild(render(tree, window.__state));
</script></body></html>`, app, SharedUISkin, uiJSON, surface, app, ApplyPropsJS, RenderUIJS)
}
