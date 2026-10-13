package declarativo

// SharedUISkin CSS idéntico en Studio preview y apps desplegadas (coherencia visual).
const SharedUISkin = `
:root{
  --alset-bg:#0f1419; --alset-card:#1a2332; --alset-ink:#e8eef8;
  --alset-accent:#F4B400; --alset-muted:#8b9bb4; --alset-line:#243044;
  --alset-blue:#1b4965; --alset-orange:#e85d04;
}
*{box-sizing:border-box}
.alset-ui{font-family:system-ui,-apple-system,sans-serif;color:var(--alset-ink)}
.alset-ui .col{display:flex;flex-direction:column;gap:.65rem}
.alset-ui .row{display:flex;gap:.5rem;flex-wrap:wrap;align-items:center}
.alset-ui button{
  background:var(--alset-accent);border:0;padding:.55rem .9rem;border-radius:10px;
  font-weight:700;cursor:pointer;color:#111;transition:transform .15s,box-shadow .15s;
}
.alset-ui button:hover{transform:translateY(-1px);box-shadow:0 4px 12px rgba(244,180,0,.35)}
.alset-ui .badge,.alset-ui .badge-ui{
  background:var(--alset-card);padding:.35rem .7rem;border-radius:999px;
  border:1px solid var(--alset-line);display:inline-block;font-size:.9rem;
}
.alset-ui .rt-splash,.alset-ui .splash{
  background:linear-gradient(135deg,var(--alset-blue),var(--alset-orange) 80%);
  color:#fff;border-radius:14px;padding:1rem 1.15rem;font-weight:800;letter-spacing:.03em;
  animation:alsetSplash .9s ease;
}
@keyframes alsetSplash{from{opacity:0;transform:translateY(10px) scale(.97)}to{opacity:1;transform:none}}
.alset-ui .gradient{
  min-height:48px;border-radius:12px;padding:.75rem;
  background:linear-gradient(120deg,var(--alset-blue),#5fa8d3,var(--alset-orange),#ffb703);
  background-size:300% 300%;animation:alsetGrad 6s ease infinite;
}
@keyframes alsetGrad{0%{background-position:0% 50%}50%{background-position:100% 50%}100%{background-position:0% 50%}}
.alset-ui input.rt-input,.alset-ui input{
  width:100%;padding:.5rem .65rem;border-radius:8px;border:1px solid var(--alset-line);
  background:#0a0e14;color:var(--alset-ink);
}
.alset-ui .muted{color:var(--alset-muted);font-size:.85rem}
`

// ApplyPropsJS función JS compartida para aplicar props del IR al DOM.
const ApplyPropsJS = `
function applyAlsetProps(el, n){
  if(!el || !n) return el;
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
  if(n.style && String(n.style).includes('animated') && el.classList) el.classList.add('animated');
  return el;
}
`
