package declarativo

// SharedUISkin CSS idéntico en Studio preview y apps desplegadas.
const SharedUISkin = `
:root{
  --alset-bg:#0f1419; --alset-card:#1a2332; --alset-ink:#e8eef8;
  --alset-accent:#F4B400; --alset-muted:#8b9bb4; --alset-line:#243044;
  --alset-blue:#1b4965; --alset-orange:#e85d04;
}
*{box-sizing:border-box}
.alset-ui{font-family:system-ui,-apple-system,sans-serif;color:var(--alset-ink);line-height:1.4}
.alset-ui .col{display:flex;flex-direction:column;gap:.65rem}
.alset-ui .row{display:flex;gap:.5rem;flex-wrap:wrap;align-items:center}
.alset-ui button, .alset-ui .alset-tab-btn{
  background:var(--alset-accent);border:0;padding:.55rem .9rem;border-radius:10px;
  font-weight:700;cursor:pointer;color:#111;font-size:13px;
}
.alset-ui .alset-tab-btn{background:#243044;color:var(--alset-ink);border-radius:8px 8px 0 0}
.alset-ui .alset-tab-btn.on{background:var(--alset-accent);color:#111}
.alset-ui .badge,.alset-ui .badge-ui{
  background:var(--alset-card);padding:.35rem .7rem;border-radius:999px;
  border:1px solid var(--alset-line);display:inline-block;font-size:.9rem;
}
.alset-ui .rt-splash,.alset-ui .splash{
  background:linear-gradient(135deg,var(--alset-blue),var(--alset-orange) 80%);
  color:#fff;border-radius:14px;padding:1rem 1.15rem;font-weight:800;letter-spacing:.03em;
}
.alset-ui .gradient{
  min-height:48px;border-radius:12px;padding:.75rem;
  background:linear-gradient(120deg,var(--alset-blue),#5fa8d3,var(--alset-orange),#ffb703);
  background-size:300% 300%;
}
.alset-ui input.rt-input,.alset-ui input{
  width:100%;padding:.5rem .65rem;border-radius:8px;border:1px solid var(--alset-line);
  background:#0a0e14;color:var(--alset-ink);
}
.alset-ui .muted{color:var(--alset-muted);font-size:.85rem}
.alset-ui .alset-tabs{display:flex;flex-direction:column;gap:0;border:1px solid var(--alset-line);border-radius:12px;overflow:hidden;background:var(--alset-card)}
.alset-ui .alset-tabs-bar{display:flex;gap:4px;padding:6px;background:#121820;flex-wrap:wrap}
.alset-ui .alset-tabs.bottom-tabs .alset-tabs-bar{order:2;border-top:1px solid var(--alset-line);justify-content:space-around}
.alset-ui .alset-tabs.bottom-tabs .alset-tab-btn{border-radius:8px;flex:1}
.alset-ui .alset-tabs-panels{padding:10px 12px;min-height:48px}
.alset-ui .alset-drawer{border:1px solid var(--alset-line);border-radius:12px;overflow:hidden;background:var(--alset-card)}
.alset-ui .alset-drawer-head{
  display:flex;align-items:center;justify-content:space-between;padding:10px 12px;
  background:#121820;cursor:pointer;font-weight:700;user-select:none;
}
.alset-ui .alset-drawer-body{padding:10px 12px;display:none}
.alset-ui .alset-drawer.open .alset-drawer-body{display:block}
.alset-ui .alset-stack{display:flex;flex-direction:column;gap:8px}
.alset-ui .ui-selected{outline:2px solid var(--alset-accent)!important;outline-offset:2px}
.alset-ui .blob{width:64px;height:64px;border-radius:40%;background:linear-gradient(135deg,#f4b400,#5b9cf5)}
`

// ApplyPropsJS aplica props del IR al DOM.
const ApplyPropsJS = `
function applyAlsetProps(el, n){
  if(!el || !n || el.nodeType!==1) return el;
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
  return el;
}
`

// RenderUIJS render completo compartido (preview deploy + producción).
// Placeholders: no — se embebe tal cual; APP_NAME y fetch los pone productionAppHTML.
const RenderUIJS = `
function val(bind, state){ if(!bind) return '·'; const v=(state||{})[bind]; return v==null?bind:v; }
function render(n, state){
  if(!n) return document.createTextNode('');
  const t=(n.type||'').toLowerCase();
  let el;
  if(t==='column'||t==='panel'){
    el=document.createElement('div'); el.className='col';
    (n.kids||[]).forEach(k=>el.appendChild(render(k,state)));
  } else if(t==='row'){
    el=document.createElement('div'); el.className='row';
    (n.kids||[]).forEach(k=>el.appendChild(render(k,state)));
  } else if(t==='label'){
    el=document.createElement('div'); el.textContent=n.text||'';
  } else if(t==='badge'){
    el=document.createElement('span'); el.className='badge'; el.textContent=val(n.bind,state);
  } else if(t==='button'){
    el=document.createElement('button'); el.textContent=n.text||'OK';
    el.onclick=async function(ev){
      ev && ev.stopPropagation && ev.stopPropagation();
      const emit=n.emit||'';
      if(!emit || !window.__alsetAction) return;
      try {
        const j = await window.__alsetAction(emit);
        if(j && j.state){
          window.__state = j.state;
          const root=document.getElementById('root');
          if(root){ root.innerHTML=''; root.appendChild(render(window.__tree, j.state)); }
        }
      } catch(e){ console.warn(e); }
    };
  } else if(t==='splash'){
    el=document.createElement('div'); el.className='rt-splash splash'; el.textContent=n.text||'Splash';
  } else if(t==='gradient'){
    el=document.createElement('div'); el.className='gradient'; el.textContent=n.text||'';
  } else if(t==='field'){
    el=document.createElement('input'); el.className='rt-input'; el.placeholder=n.text||'';
  } else if(t==='blob'){
    el=document.createElement('div'); el.className='blob';
  } else if(t==='image'){
    el=document.createElement('img'); el.src=n.src||''; el.style.maxWidth='100%';
  } else if(t==='top-tabs'||t==='tabs'||t==='bottom-tabs'){
    el=document.createElement('div'); el.className='alset-tabs '+t;
    const bar=document.createElement('div'); bar.className='alset-tabs-bar';
    const panels=document.createElement('div'); panels.className='alset-tabs-panels';
    (n.kids||[]).forEach(function(k,i){
      const btn=document.createElement('button'); btn.type='button';
      btn.className='alset-tab-btn'+(i===0?' on':'');
      btn.textContent=k.text||('Tab '+(i+1));
      const panel=document.createElement('div'); panel.className='alset-tab-panel';
      panel.style.display=i===0?'block':'none';
      (k.kids||[]).forEach(function(c){ panel.appendChild(render(c,state)); });
      btn.onclick=function(ev){
        ev.stopPropagation();
        bar.querySelectorAll('.alset-tab-btn').forEach(function(b){ b.classList.remove('on'); });
        btn.classList.add('on');
        panels.querySelectorAll('.alset-tab-panel').forEach(function(p){ p.style.display='none'; });
        panel.style.display='block';
      };
      bar.appendChild(btn); panels.appendChild(panel);
    });
    if(t==='bottom-tabs'){ el.appendChild(panels); el.appendChild(bar); }
    else { el.appendChild(bar); el.appendChild(panels); }
  } else if(t==='drawer'){
    el=document.createElement('div'); el.className='alset-drawer open';
    const head=document.createElement('div'); head.className='alset-drawer-head';
    head.innerHTML='<span>'+(n.text||'Menu')+'</span><span>▾</span>';
    const body=document.createElement('div'); body.className='alset-drawer-body';
    body.style.display='block';
    (n.kids||[]).forEach(function(k){ body.appendChild(render(k,state)); });
    head.onclick=function(ev){
      ev.stopPropagation();
      el.classList.toggle('open');
      body.style.display = el.classList.contains('open') ? 'block' : 'none';
    };
    el.appendChild(head); el.appendChild(body);
  } else if(t==='stack'){
    el=document.createElement('div'); el.className='alset-stack';
    (n.kids||[]).forEach(function(k){ el.appendChild(render(k,state)); });
  } else if(t==='tab'){
    el=document.createElement('div'); el.className='col';
    (n.kids||[]).forEach(function(k){ el.appendChild(render(k,state)); });
  } else {
    el=document.createElement('div'); el.textContent=n.type||'';
  }
  return applyAlsetProps(el, n);
}
`
