package accesslog

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes 注册访问日志查看页面与 JSON API。
func RegisterRoutes(r *gin.Engine, store *Store, uiPath string) {
	if store == nil || !store.Enabled() {
		return
	}
	if uiPath == "" {
		uiPath = "/debug/access-logs"
	}
	r.GET(uiPath, func(c *gin.Context) {
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.String(http.StatusOK, pageHTML(uiPath))
	})
	r.GET(uiPath+"/api", func(c *gin.Context) {
		limit := 100
		if v := c.Query("limit"); v != "" {
			if n, err := strconv.Atoi(v); err == nil && n > 0 {
				limit = n
			}
		}
		c.JSON(http.StatusOK, gin.H{
			"items": store.List(limit),
		})
	})
}

func pageHTML(apiBase string) string {
	return `<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="utf-8"/>
<meta name="viewport" content="width=device-width, initial-scale=1"/>
<title>Access Logs</title>
<style>
  :root {
    --bg: #0b0f14;
    --surface: #121820;
    --surface-2: #1a222d;
    --border: #2a3444;
    --text: #e8edf4;
    --muted: #8b97a8;
    --accent: #4c8dff;
    --accent-dim: rgba(76, 141, 255, 0.12);
    --ok: #34d399;
    --warn: #fbbf24;
    --err: #f87171;
    --radius: 10px;
    --ease: cubic-bezier(0.4, 0, 0.2, 1);
  }
  * { box-sizing: border-box; }
  body {
    font-family: "SF Pro Text", ui-sans-serif, system-ui, sans-serif;
    background: var(--bg);
    color: var(--text);
    margin: 0;
    padding: 1.25rem 1.5rem 2rem;
    line-height: 1.45;
  }
  .header { margin-bottom: 1.25rem; }
  .header h1 { font-size: 1.35rem; font-weight: 600; margin: 0 0 .35rem; letter-spacing: -0.02em; }
  .header p { margin: 0; color: var(--muted); font-size: .875rem; }
  .toolbar {
    display: flex; gap: .65rem; align-items: center; flex-wrap: wrap;
    margin-bottom: 1rem; padding: .75rem 1rem;
    background: var(--surface); border: 1px solid var(--border); border-radius: var(--radius);
  }
  .toolbar input {
    flex: 1; min-width: 180px; max-width: 320px;
    background: var(--surface-2); border: 1px solid var(--border); color: var(--text);
    padding: .45rem .65rem; border-radius: 8px; font-size: .85rem;
  }
  .toolbar input:focus { outline: 2px solid var(--accent-dim); border-color: var(--accent); }
  .btn {
    background: var(--accent); color: #fff; border: 0;
    padding: .45rem .85rem; border-radius: 8px; cursor: pointer; font-size: .85rem; font-weight: 500;
    transition: transform .15s var(--ease), filter .15s;
  }
  .btn:hover { filter: brightness(1.06); }
  .btn:active { transform: scale(0.98); }
  .btn-ghost {
    background: transparent; color: var(--muted); border: 1px solid var(--border);
  }
  .btn-ghost:hover { color: var(--text); border-color: var(--muted); }
  .btn-ghost.active { color: var(--accent); border-color: var(--accent); background: var(--accent-dim); }
  .meta { color: var(--muted); font-size: .8rem; margin-left: auto; }
  .table-wrap {
    border: 1px solid var(--border); border-radius: var(--radius);
    overflow: hidden; background: var(--surface);
  }
  table { width: 100%; border-collapse: collapse; font-size: .8125rem; }
  th {
    text-align: left; padding: .65rem .75rem; color: var(--muted); font-weight: 600;
    background: var(--surface-2); border-bottom: 1px solid var(--border);
    position: sticky; top: 0; z-index: 1;
  }
  td { padding: .55rem .75rem; border-bottom: 1px solid var(--border); vertical-align: middle; }
  tr.log-row { cursor: pointer; transition: background .18s var(--ease); }
  tr.log-row:hover td { background: rgba(255,255,255,.03); }
  tr.log-row.selected td { background: var(--accent-dim); }
  tr.log-row.selected td:first-child { box-shadow: inset 3px 0 0 var(--accent); }
  .drawer-backdrop {
    position: fixed; inset: 0; z-index: 40;
    background: rgba(0,0,0,.45); opacity: 0; pointer-events: none;
    transition: opacity .32s var(--ease);
  }
  .drawer-backdrop.open { opacity: 1; pointer-events: auto; }
  .drawer {
    position: fixed; top: 0; right: 0; z-index: 50;
    width: min(520px, 100vw); height: 100vh;
    background: var(--surface); border-left: 1px solid var(--border);
    box-shadow: -12px 0 40px rgba(0,0,0,.35);
    transform: translateX(100%);
    transition: transform .38s var(--ease);
    display: flex; flex-direction: column;
  }
  .drawer.open { transform: translateX(0); }
  .drawer-header {
    flex-shrink: 0; padding: 1rem 1.1rem; border-bottom: 1px solid var(--border);
    display: flex; align-items: flex-start; gap: .75rem;
  }
  .drawer-header-main { flex: 1; min-width: 0; }
  .drawer-title { font-size: .95rem; font-weight: 600; margin: 0 0 .35rem; word-break: break-all; }
  .drawer-sub { font-size: .78rem; color: var(--muted); display: flex; flex-wrap: wrap; gap: .5rem .75rem; }
  .drawer-close {
    flex-shrink: 0; width: 2rem; height: 2rem; border: 1px solid var(--border);
    background: var(--surface-2); color: var(--muted); border-radius: 8px; cursor: pointer;
    font-size: 1.1rem; line-height: 1; transition: color .15s, border-color .15s;
  }
  .drawer-close:hover { color: var(--text); border-color: var(--muted); }
  .drawer-body {
    flex: 1; overflow-y: auto; padding: 1rem 1.1rem 1.5rem;
    overscroll-behavior: contain;
  }
  .drawer .grid-2 { grid-template-columns: 1fr; }
  .drawer pre.code { max-height: 280px; }
  body.drawer-open { overflow: hidden; }
  .detail-body { padding: 0; }
  .method {
    display: inline-block; font-size: .68rem; font-weight: 700; letter-spacing: .03em;
    padding: .15rem .45rem; border-radius: 5px; font-family: ui-monospace, monospace;
  }
  .method-GET { background: rgba(76,141,255,.2); color: #7eb6ff; }
  .method-POST { background: rgba(52,211,153,.18); color: #6ee7b7; }
  .method-PATCH, .method-PUT { background: rgba(251,191,36,.15); color: #fcd34d; }
  .method-DELETE { background: rgba(248,113,113,.18); color: #fca5a5; }
  .path { font-family: ui-monospace, SFMono-Regular, monospace; font-size: .78rem; color: #c5d0de; }
  .status-pill {
    display: inline-block; min-width: 2.5rem; text-align: center;
    padding: .12rem .4rem; border-radius: 6px; font-weight: 600; font-size: .75rem;
  }
  .status-ok { background: rgba(52,211,153,.15); color: var(--ok); }
  .status-warn { background: rgba(251,191,36,.12); color: var(--warn); }
  .status-err { background: rgba(248,113,113,.15); color: var(--err); }
  .dur { font-variant-numeric: tabular-nums; color: var(--muted); }
  .dur-slow { color: var(--warn); }
  .sql-badge {
    font-size: .72rem; padding: .1rem .45rem; border-radius: 999px;
    background: var(--surface-2); color: var(--muted); border: 1px solid var(--border);
  }
  .sql-badge.has { color: var(--accent); border-color: rgba(76,141,255,.35); }
  .grid-2 { display: grid; grid-template-columns: 1fr 1fr; gap: .75rem; }
  @media (max-width: 900px) { .grid-2 { grid-template-columns: 1fr; } }
  .block { margin-bottom: .85rem; }
  .block-title {
    font-size: .72rem; text-transform: uppercase; letter-spacing: .06em;
    color: var(--muted); margin-bottom: .35rem; font-weight: 600;
  }
  pre.code {
    margin: 0; white-space: pre-wrap; word-break: break-word;
    background: #080b10; border: 1px solid var(--border); padding: .65rem .75rem;
    border-radius: 8px; font-size: .74rem; line-height: 1.5;
    max-height: 220px; overflow: auto; font-family: ui-monospace, monospace;
  }
  .kv { font-size: .8rem; color: var(--muted); margin-bottom: .5rem; }
  .kv b { color: var(--text); font-weight: 500; }
  .err-box {
    color: var(--err); background: rgba(248,113,113,.08); border: 1px solid rgba(248,113,113,.25);
    padding: .5rem .65rem; border-radius: 8px; font-size: .8rem; margin-bottom: .75rem;
  }
  .sql-list { display: flex; flex-direction: column; gap: .5rem; }
  .sql-card {
    border: 1px solid var(--border); border-radius: 8px; overflow: hidden;
    background: var(--surface);
  }
  .sql-card summary {
    list-style: none; cursor: pointer; padding: .5rem .65rem;
    display: flex; align-items: center; gap: .5rem; font-size: .78rem;
    transition: background .15s;
  }
  .sql-card summary::-webkit-details-marker { display: none; }
  .sql-card summary:hover { background: rgba(255,255,255,.04); }
  .sql-card summary .sql-chev { transition: transform .25s var(--ease); color: var(--muted); }
  .sql-card[open] summary .sql-chev { transform: rotate(90deg); color: var(--accent); }
  .sql-card .sql-meta { color: var(--muted); font-variant-numeric: tabular-nums; }
  .sql-card .sql-preview {
    flex: 1; overflow: hidden; text-overflow: ellipsis; white-space: nowrap;
    font-family: ui-monospace, monospace; font-size: .72rem; color: #a8b4c4;
  }
  .sql-card pre { border: 0; border-top: 1px solid var(--border); border-radius: 0; max-height: 180px; }
  .empty { text-align: center; padding: 2.5rem; color: var(--muted); }
</style>
</head>
<body>
<div class="header">
  <h1>HTTP 访问日志</h1>
  <p>点击表格行在右侧查看详情；SQL 可单独展开</p>
</div>
<div class="toolbar">
  <input type="search" id="filter" placeholder="筛选路径 / 方法…" autocomplete="off"/>
  <button type="button" class="btn" id="refresh">刷新</button>
  <button type="button" class="btn btn-ghost" id="auto" title="每 10 秒自动刷新">自动刷新</button>
  <button type="button" class="btn btn-ghost" id="collapse">关闭侧栏</button>
  <span class="meta" id="status">加载中…</span>
</div>
<div class="table-wrap">
<table>
<thead><tr>
  <th>时间</th><th>方法</th><th>路径</th><th>状态</th><th>耗时</th><th>SQL</th>
</tr></thead>
<tbody id="rows"></tbody>
</table>
</div>
<div class="drawer-backdrop" id="backdrop" aria-hidden="true"></div>
<aside class="drawer" id="drawer" aria-hidden="true">
  <div class="drawer-header">
    <div class="drawer-header-main">
      <h2 class="drawer-title" id="drawer-title">请求详情</h2>
      <div class="drawer-sub" id="drawer-sub"></div>
    </div>
    <button type="button" class="drawer-close" id="drawer-close" title="关闭 (Esc)">×</button>
  </div>
  <div class="drawer-body" id="drawer-content"></div>
</aside>
<script>
const api = "` + apiBase + `/api";
let items = [];
let expandedId = null;
let autoTimer = null;

function esc(s){const d=document.createElement('div');d.textContent=s??'';return d.innerHTML;}
function statusClass(code){if(code<400)return'status-ok';if(code<500)return'status-warn';return'status-err';}
function methodClass(m){return 'method method-'+(m||'GET');}
function prettyJson(raw){
  if(!raw) return '';
  try { return JSON.stringify(JSON.parse(raw), null, 2); } catch { return raw; }
}
function sqlPreview(q){ return (q||'').replace(/\s+/g,' ').trim().slice(0, 80); }

function buildDetail(e){
  const sqlCount = (e.sql && e.sql.length) || 0;
  const sqlHtml = sqlCount
    ? e.sql.map((s,i)=>'<details class="sql-card"><summary><span class="sql-chev">▶</span><span class="sql-meta">#'+(i+1)+' · '+s.duration_ms.toFixed(2)+' ms</span><span class="sql-preview">'+esc(sqlPreview(s.query))+'</span></summary><pre class="code">'+esc(s.query)+(s.args?'\n\n-- args\n'+esc(s.args):'')+(s.error?'\n\n-- error\n'+esc(s.error):'')+'</pre></details>').join('')
    : '<p class="kv">无 SQL 记录</p>';
  return '<div class="detail-body">'+
    (e.error?'<div class="err-box">'+esc(e.error)+'</div>':'')+
    '<div class="kv"><b>Query</b> '+esc(e.query||'—')+' &nbsp;·&nbsp; <b>IP</b> '+esc(e.client_ip)+'</div>'+
    '<div class="grid-2">'+
      '<div class="block"><div class="block-title">Request Body</div><pre class="code">'+esc(prettyJson(e.request_body)||'—')+'</pre></div>'+
      '<div class="block"><div class="block-title">Response Body</div><pre class="code">'+esc(prettyJson(e.response_body)||'—')+'</pre></div>'+
    '</div>'+
    '<div class="block"><div class="block-title">SQL ('+sqlCount+')</div><div class="sql-list">'+sqlHtml+'</div></div>'+
  '</div>';
}

function closeDrawer(){
  expandedId = null;
  document.getElementById('drawer').classList.remove('open');
  document.getElementById('backdrop').classList.remove('open');
  document.body.classList.remove('drawer-open');
  document.getElementById('drawer').setAttribute('aria-hidden','true');
  document.getElementById('backdrop').setAttribute('aria-hidden','true');
  document.querySelectorAll('tr.log-row').forEach(r => r.classList.remove('selected'));
}

function openDrawer(e){
  expandedId = e.id;
  document.querySelectorAll('tr.log-row').forEach(r => {
    r.classList.toggle('selected', r.dataset.id === e.id);
  });
  const sqlCount = (e.sql && e.sql.length) || 0;
  document.getElementById('drawer-title').textContent = (e.method||'') + ' ' + (e.path||'');
  document.getElementById('drawer-sub').innerHTML =
    '<span class="status-pill '+statusClass(e.status)+'">'+e.status+'</span>'+
    '<span>'+e.duration_ms.toFixed(2)+' ms</span>'+
    '<span>SQL '+sqlCount+'</span>'+
    '<span>'+esc(new Date(e.time).toLocaleString())+'</span>';
  document.getElementById('drawer-content').innerHTML = buildDetail(e);
  document.getElementById('drawer').classList.add('open');
  document.getElementById('backdrop').classList.add('open');
  document.body.classList.add('drawer-open');
  document.getElementById('drawer').setAttribute('aria-hidden','false');
  document.getElementById('backdrop').setAttribute('aria-hidden','false');
}

function toggleRow(id){
  const e = items.find(x => x.id === id);
  if(!e) return;
  if(expandedId === id){
    closeDrawer();
    return;
  }
  openDrawer(e);
}

function render(list){
  const q = (document.getElementById('filter').value||'').trim().toLowerCase();
  const filtered = q ? list.filter(e =>
    (e.path||'').toLowerCase().includes(q) ||
    (e.method||'').toLowerCase().includes(q) ||
    String(e.status).includes(q)
  ) : list;

  const tb = document.getElementById('rows');
  tb.innerHTML = '';
  if(!filtered.length){
    tb.innerHTML = '<tr><td colspan="6" class="empty">暂无匹配记录</td></tr>';
    document.getElementById('status').textContent = '0 条';
    return;
  }

  for(const e of filtered){
    const sqlCount = (e.sql && e.sql.length) || 0;
    const slow = e.duration_ms > 300;
    const row = document.createElement('tr');
    row.className = 'log-row' + (e.id === expandedId ? ' selected' : '');
    row.dataset.id = e.id;
    row.innerHTML =
      '<td>'+esc(new Date(e.time).toLocaleString())+'</td>'+
      '<td><span class="'+methodClass(e.method)+'">'+esc(e.method)+'</span></td>'+
      '<td class="path">'+esc(e.path)+'</td>'+
      '<td><span class="status-pill '+statusClass(e.status)+'">'+e.status+'</span></td>'+
      '<td class="dur'+(slow?' dur-slow':'')+'">'+e.duration_ms.toFixed(2)+' ms</td>'+
      '<td><span class="sql-badge'+(sqlCount?' has':'')+'">'+sqlCount+'</span></td>';
    row.addEventListener('click', () => toggleRow(e.id));
    tb.appendChild(row);
  }
  if(expandedId){
    const cur = filtered.find(x => x.id === expandedId);
    if(cur) openDrawer(cur);
    else closeDrawer();
  }
  document.getElementById('status').textContent = filtered.length + ' / ' + list.length + ' 条';
}

async function load(){
  document.getElementById('status').textContent = '加载中…';
  try {
    const r = await fetch(api + '?limit=200');
    const j = await r.json();
    items = j.items || [];
    const ids = new Set(items.map(x => x.id));
    if(expandedId && !ids.has(expandedId)) closeDrawer();
    render(items);
  } catch(err) {
    document.getElementById('status').textContent = '加载失败';
  }
}

document.getElementById('refresh').onclick = load;
document.getElementById('filter').oninput = () => render(items);
document.getElementById('collapse').onclick = closeDrawer;
document.getElementById('drawer-close').onclick = closeDrawer;
document.getElementById('backdrop').onclick = closeDrawer;
document.addEventListener('keydown', ev => { if(ev.key === 'Escape') closeDrawer(); });
document.getElementById('auto').onclick = function(){
  this.classList.toggle('active');
  if(this.classList.contains('active')){
    autoTimer = setInterval(load, 10000);
  } else {
    clearInterval(autoTimer);
    autoTimer = null;
  }
};

load();
</script>
</body>
</html>`
}
