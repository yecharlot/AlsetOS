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
	w := "360px"
	switch device {
	case "tablet":
		w = "768px"
	case "desktop":
		w = "100%"
	}
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="es"><head>
<meta charset="utf-8"/><meta name="viewport" content="width=device-width,initial-scale=1"/>
<title>%s · %s</title>
<style>
%s
body{margin:0;background:#0a0e14;display:flex;justify-content:center;padding:1rem}
.frame{width:%s;max-width:100%%;min-height:60vh;background:var(--alset-bg);border:1px solid #243044;border-radius:12px;padding:1rem}
</style></head><body>
<div class="frame alset-ui" id="root"></div>
<script>
const SURFACES = %s;
const SURFACE = %q;
%s
function render(n){
  if(!n) return document.createTextNode('');
  const t=(n.type||'').toLowerCase();
  let el;
  if(t==='column'||t==='panel'){ el=document.createElement('div'); el.className='col'; (n.kids||[]).forEach(k=>el.appendChild(render(k))); }
  else if(t==='row'){ el=document.createElement('div'); el.className='row'; (n.kids||[]).forEach(k=>el.appendChild(render(k))); }
  else if(t==='label'){ el=document.createElement('div'); el.textContent=n.text||''; }
  else if(t==='badge'){ el=document.createElement('span'); el.className='badge'; el.textContent=n.bind||n.text||'·'; }
  else if(t==='button'){ el=document.createElement('button'); el.textContent=n.text||'OK'; el.onclick=()=>alert('emit: '+(n.emit||'')); }
  else if(t==='splash'){ el=document.createElement('div'); el.className='rt-splash splash'; el.textContent=n.text||'Splash'; }
  else if(t==='gradient'){ el=document.createElement('div'); el.className='gradient'; el.textContent=n.text||''; }
  else if(t==='field'){ el=document.createElement('input'); el.className='rt-input'; el.placeholder=n.text||''; }
  else { el=document.createElement('div'); el.textContent=n.type||''; }
  return applyAlsetProps(el, n);
}
const tree = SURFACES[SURFACE] || Object.values(SURFACES)[0];
document.getElementById('root').appendChild(render(tree));
</script></body></html>`, app, device, SharedUISkin, w, uiJSON, surface, ApplyPropsJS)
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
