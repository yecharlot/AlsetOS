package declarativo

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// DeployFlow genera pack multiplataforma (preview mobile/tablet/desktop) desde el flow.
func DeployFlow(flow *FlowAOS, outDir string) error {
	if flow == nil {
		return fmt.Errorf("flow nulo")
	}
	if outDir == "" {
		if flow.Deploy != nil && flow.Deploy.OutDir != "" {
			outDir = flow.Deploy.OutDir
		} else {
			outDir = "dist/" + flow.Name
		}
	}
	devices := []string{"mobile", "tablet", "desktop"}
	if flow.Deploy != nil && len(flow.Deploy.Devices) > 0 {
		devices = flow.Deploy.Devices
	}
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}
	if err := GenerarArtefactosFlow(flow, filepath.Join(outDir, "artifacts")); err != nil {
		return err
	}
	// shell preview
	manifest := map[string]interface{}{
		"app": flow.Name, "version": flow.Version, "devices": devices,
		"surface": flow.SurfaceOID, "source": "lispai-aos",
	}
	mb, _ := json.MarshalIndent(manifest, "", "  ")
	_ = os.WriteFile(filepath.Join(outDir, "app.manifest.json"), mb, 0o644)

	uiJSON, _ := json.Marshal(flow.Surfaces)
	for _, d := range devices {
		dir := filepath.Join(outDir, d)
		_ = os.MkdirAll(dir, 0o755)
		html := previewHTML(flow.Name, d, string(uiJSON), flow.SurfaceOID)
		if err := os.WriteFile(filepath.Join(dir, "index.html"), []byte(html), 0o644); err != nil {
			return err
		}
	}
	// index launcher
	idx := fmt.Sprintf(`<!DOCTYPE html><html lang="es"><head><meta charset="utf-8"/>
<meta name="viewport" content="width=device-width,initial-scale=1"/>
<meta http-equiv="refresh" content="0;url=mobile/"/>
<title>%s</title>
<style>body{margin:0;background:#0f1419}</style></head><body></body></html>`, flow.Name)
	return os.WriteFile(filepath.Join(outDir, "index.html"), []byte(idx), 0o644)
}

func previewHTML(app, device, uiJSON, surface string) string {
	widths := map[string]string{"mobile": "390px", "tablet": "768px", "desktop": "100%"}
	w := widths[device]
	if w == "" {
		w = "100%"
	}
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="es"><head><meta charset="utf-8"/>
<meta name="viewport" content="width=device-width,initial-scale=1"/>
<title>%s · %s</title>
<style>
:root{--bg:#0f1419;--card:#1a2332;--ink:#e8eef8;--accent:#F4B400;--muted:#8b9bb4}
*{box-sizing:border-box}
body{margin:0;background:#070b12;color:var(--ink);font-family:system-ui,sans-serif;min-height:100vh;display:flex;flex-direction:column;align-items:center}
.frame{width:%s;max-width:100%%;min-height:60vh;background:var(--bg);border:1px solid #243044;border-radius:12px;padding:1rem;margin:0 0.5rem 1rem}
.row{display:flex;gap:.5rem;flex-wrap:wrap}
.col{display:flex;flex-direction:column;gap:.5rem}
button{background:var(--accent);border:0;padding:.5rem .75rem;border-radius:8px;font-weight:600;cursor:pointer}
.badge{display:inline-block;background:var(--card);padding:.35rem .6rem;border-radius:999px;border:1px solid #243044}
label{color:var(--muted)}
</style></head><body>
<div class="frame" id="root"></div>
<script>
const SURFACES = %s;
const SURFACE = %q;
function render(n){
  if(!n) return document.createTextNode('');
  const t = (n.type||'').toLowerCase();
  if(t==='column'||t==='panel'){const d=document.createElement('div');d.className='col';(n.kids||[]).forEach(k=>d.appendChild(render(k)));return d;}
  if(t==='row'){const d=document.createElement('div');d.className='row';(n.kids||[]).forEach(k=>d.appendChild(render(k)));return d;}
  if(t==='label'){const s=document.createElement('div');s.textContent=n.text||'';return s;}
  if(t==='badge'){const s=document.createElement('span');s.className='badge';s.textContent=n.bind||n.text||'·';return s;}
  if(t==='button'){const b=document.createElement('button');b.textContent=n.text||'OK';b.onclick=()=>alert('emit: '+(n.emit||''));return b;}
  if(t==='field'){const i=document.createElement('input');i.placeholder=n.text||'';return i;}
  return document.createTextNode(JSON.stringify(n));
}
const tree = SURFACES[SURFACE] || Object.values(SURFACES)[0];
document.getElementById('root').appendChild(render(tree));
</script></body></html>`, app, device, w, uiJSON, surface)
}

// RunFlowTests evalúa aserciones simples sobre estado reactivo.
func (rt *Runtime) RunFlowTests(flow *FlowAOS) []string {
	var errs []string
	if flow == nil {
		return []string{"flow nulo"}
	}
	for _, tc := range flow.Tests {
		v, ok := rt.State.Get(tc.Path)
		if !ok && tc.Op != "missing" {
			errs = append(errs, fmt.Sprintf("%s: path %s ausente", tc.Name, tc.Path))
			continue
		}
		switch tc.Op {
		case "eq":
			if fmt.Sprint(v) != fmt.Sprint(tc.Want) {
				errs = append(errs, fmt.Sprintf("%s: %s=%v want %v", tc.Name, tc.Path, v, tc.Want))
			}
		case "gt", "gte", "lt", "lte", "ne":
			// comparación textual mínima
			if fmt.Sprint(v) == fmt.Sprint(tc.Want) && tc.Op == "ne" {
				errs = append(errs, fmt.Sprintf("%s: %s no debía ser %v", tc.Name, tc.Path, v))
			}
		}
	}
	return errs
}
