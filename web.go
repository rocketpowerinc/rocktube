package main

// indexHTML is the embedded single-page app. No external requests — everything
// (CSS, JS, icons) lives inline so the binary stays truly self-contained.
const indexHTML = `<!doctype html>
<html lang="en" data-theme="dark">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width,initial-scale=1">
<title>RockTube</title>
<link rel="icon" href="/favicon.ico">
<style>
:root{
  --bg:#0f0f0f; --bg-2:#1f1f1f; --bg-3:#272727; --bg-hover:#3f3f3f;
  --text:#f1f1f1; --text-dim:#aaaaaa; --text-faint:#717171;
  --accent:#ff0033; --accent-2:#3ea6ff;
  --radius:12px; --radius-sm:8px;
}
*{box-sizing:border-box;margin:0;padding:0}
html,body{height:100%}
body{
  background:var(--bg);color:var(--text);
  font-family:"Roboto","Segoe UI",system-ui,Arial,sans-serif;
  -webkit-font-smoothing:antialiased;overflow-x:hidden;
}
a{color:inherit;text-decoration:none}
img{display:block;max-width:100%}
button{font-family:inherit;cursor:pointer;border:none;background:none;color:inherit}

/* ---------- top bar ---------- */
.topbar{
  position:sticky;top:0;z-index:50;
  display:flex;align-items:center;gap:16px;
  height:56px;padding:0 16px;
  background:var(--bg);
}
.logo{display:flex;align-items:center;gap:8px;min-width:200px}
.logo-icon{
  width:32px;height:32px;border-radius:50%;
  background:var(--accent);
  display:grid;place-items:center;flex:none;
}
.logo-icon svg{width:16px;height:16px}
.logo-text{font-size:20px;font-weight:700;letter-spacing:-.5px}
.logo-text b{color:var(--accent)}

.search{
  flex:1;display:flex;align-items:center;justify-content:center;max-width:640px;margin:0 auto;
}
.search-box{
  display:flex;width:100%;height:40px;
  border:1px solid var(--bg-3);border-radius:40px;overflow:hidden;
  background:var(--bg);
}
.search-box input{
  flex:1;background:transparent;border:none;outline:none;color:var(--text);
  font-size:16px;padding:0 16px;
}
.search-box input::placeholder{color:var(--text-faint)}
.search-box .btn{
  width:64px;display:grid;place-items:center;background:var(--bg-3);
}
.search-box .btn:hover{background:var(--bg-hover)}
.top-actions{min-width:200px;display:flex;justify-content:flex-end;gap:8px;align-items:center}
.chip{
  display:flex;align-items:center;gap:6px;padding:8px 12px;border-radius:18px;
  background:var(--bg-3);font-size:14px;
}
.chip:hover{background:var(--bg-hover)}
.scan-btn{
  width:36px;height:36px;border-radius:50%;display:grid;place-items:center;
  background:var(--bg-3);color:var(--text-dim);
}
.scan-btn:hover{background:var(--bg-hover);color:var(--text)}
.scan-btn svg{width:20px;height:20px}

/* ---------- layout ---------- */
.shell{display:flex;min-height:calc(100vh - 56px)}
.sidebar{
  width:240px;flex:none;padding:12px 8px;position:sticky;top:56px;height:calc(100vh - 56px);
  overflow-y:auto;
}
.sidebar::-webkit-scrollbar{width:8px}
.sidebar::-webkit-scrollbar-thumb{background:var(--bg-3);border-radius:4px}
.nav-item{
  display:flex;align-items:center;gap:20px;padding:10px 12px;border-radius:10px;
  font-size:14px;color:var(--text);cursor:pointer;
}
.nav-item:hover{background:var(--bg-3)}
.nav-item.active{background:var(--bg-3);font-weight:600}
.nav-item svg{width:24px;height:24px;flex:none}
.nav-divider{height:1px;background:var(--bg-3);margin:12px 0}
.nav-section{padding:8px 12px;font-size:14px;color:var(--text-faint);text-transform:none}
.folder-tree{display:flex;flex-direction:column;gap:2px}
.folder-node{display:flex;flex-direction:column}
.folder-row{
  display:grid;grid-template-columns:24px 1fr auto;align-items:center;gap:6px;
  min-height:36px;padding:6px 8px;border-radius:10px;font-size:14px;color:var(--text);
}
.folder-row:hover{background:var(--bg-3)}
.folder-row.active{background:var(--bg-3);font-weight:600}
.folder-toggle{width:24px;height:24px;border-radius:50%;display:grid;place-items:center;color:var(--text-dim)}
.folder-toggle:hover{background:var(--bg-hover);color:var(--text)}
.folder-toggle svg{width:16px;height:16px;transition:transform .15s}
.folder-node.collapsed>.folder-row .folder-toggle svg{transform:rotate(-90deg)}
.folder-name{overflow:hidden;text-overflow:ellipsis;white-space:nowrap;cursor:pointer}
.folder-count{font-size:12px;color:var(--text-faint);font-weight:500}
.folder-children{margin-left:16px}
.folder-node.collapsed>.folder-children{display:none}
.main{flex:1;padding:24px 24px 60px;min-width:0}
.page-head{margin-bottom:18px}
.page-head h1{font-size:20px;line-height:1.25;margin-bottom:4px}
.page-head p{font-size:14px;color:var(--text-dim)}
.crumbs{display:flex;align-items:center;gap:8px;flex-wrap:wrap;margin-bottom:18px;color:var(--text-dim);font-size:14px}
.crumbs button{color:var(--text);font-weight:600}
.crumbs button:hover{color:var(--accent-2)}
.crumbs .sep{color:var(--text-faint)}

/* ---------- video grid ---------- */
.grid{
  display:grid;gap:16px 12px;
  grid-template-columns:repeat(auto-fill,minmax(300px,1fr));
}
.card{cursor:pointer;display:flex;flex-direction:column}
.thumb{
  position:relative;width:100%;aspect-ratio:16/9;border-radius:var(--radius);
  overflow:hidden;background:var(--bg-3);
}
.thumb img{width:100%;height:100%;object-fit:cover;transition:transform .25s ease}
.card:hover .thumb img{transform:scale(1.04)}
.thumb .dur{
  position:absolute;right:8px;bottom:8px;
  background:rgba(0,0,0,.85);color:#fff;font-size:12px;font-weight:600;
  padding:2px 5px;border-radius:4px;
}
.thumb .placeholder{width:100%;height:100%;display:grid;place-items:center;color:var(--text-faint)}
.card-body{display:flex;gap:12px;margin-top:12px}
.avatar{
  width:36px;height:36px;border-radius:50%;flex:none;
  background:linear-gradient(135deg,var(--accent),#7a0020);
  display:grid;place-items:center;font-weight:700;color:#fff;
}
.meta{flex:1;min-width:0}
.meta .title{
  font-size:14px;line-height:1.4;font-weight:600;
  display:-webkit-box;-webkit-line-clamp:2;-webkit-box-orient:vertical;overflow:hidden;
}
.meta .sub{font-size:13px;color:var(--text-dim);margin-top:4px}
.meta .sub span{display:block}
.path-badge{
  display:inline-flex;max-width:100%;margin-bottom:3px;padding:2px 6px;border-radius:4px;
  background:var(--bg-3);color:var(--text-dim);font-size:12px;overflow:hidden;text-overflow:ellipsis;white-space:nowrap;
}

/* ---------- watch page ---------- */
.watch{display:flex;gap:24px;align-items:flex-start;max-width:1280px;margin:0 auto}
.player-wrap{flex:1;min-width:0}
.player{
  width:100%;aspect-ratio:16/9;background:#000;border-radius:var(--radius);
  overflow:hidden;
}
.player video{width:100%;height:100%}
.watch-title{font-size:18px;font-weight:600;margin:12px 0 8px;line-height:1.4}
.watch-meta{
  display:flex;justify-content:space-between;align-items:center;flex-wrap:wrap;gap:12px;
  padding:12px 0;border-bottom:1px solid var(--bg-3);
}
.channel{display:flex;align-items:center;gap:12px}
.channel .avatar{width:40px;height:40px}
.channel .name{font-weight:600}
.channel .sub-count{font-size:13px;color:var(--text-dim)}
.views-line{color:var(--text-dim);font-size:14px}
.actions{display:flex;align-items:center;gap:8px;flex-wrap:wrap}
.rate-pill{
  display:flex;align-items:stretch;background:var(--bg-3);border-radius:18px;overflow:hidden;
}
.rate-pill button{
  display:flex;align-items:center;gap:8px;padding:8px 16px;font-weight:600;font-size:14px;
  color:var(--text);transition:background .15s;
}
.rate-pill button:hover{background:var(--bg-hover)}
.rate-pill button.active{color:#fff}
.rate-pill .like.active{background:var(--accent-2);color:#062b4a}
.rate-pill .dislike.active{background:var(--accent);color:#fff}
.rate-pill .div{width:1px;background:var(--bg-hover);align-self:center;height:24px}
.rate-pill svg{width:20px;height:20px;flex:none}
.rate-pill .cnt{font-size:13px;font-weight:600;min-width:14px;text-align:center}
.action-btn{
  display:flex;align-items:center;gap:8px;padding:8px 16px;border-radius:18px;background:var(--bg-3);
  font-weight:600;font-size:14px;color:var(--text);
}
.action-btn:hover{background:var(--bg-hover)}
.action-btn svg{width:20px;height:20px;flex:none}
.watch-desc{
  background:var(--bg-2);border-radius:12px;padding:12px;margin-top:12px;font-size:14px;
  color:var(--text-dim);line-height:1.5;
}
/* ---------- comments ---------- */
.comments{margin-top:24px;max-width:720px}
.comments-head{display:flex;align-items:center;gap:24px;margin-bottom:20px}
.comments-head .count{font-size:16px;font-weight:600}
.comment-form{display:flex;gap:12px;margin-bottom:24px}
.comment-form .avatar{width:40px;height:40px;flex:none}
.comment-form .fields{flex:1;min-width:0}
.comment-form input.name{
  width:100%;max-width:280px;background:transparent;border:none;border-bottom:1px solid var(--bg-3);
  color:var(--text);font-size:14px;padding:6px 0;margin-bottom:8px;outline:none;
}
.comment-form input.name:focus{border-bottom-color:var(--accent-2)}
.comment-form textarea{
  width:100%;min-height:24px;resize:none;background:transparent;border:none;border-bottom:1px solid var(--bg-3);
  color:var(--text);font-size:14px;font-family:inherit;padding:6px 0;outline:none;
}
.comment-form textarea:focus{border-bottom-color:var(--accent-2)}
.comment-form .row{display:flex;justify-content:flex-end;gap:8px;margin-top:8px}
.comment-form .cancel{padding:8px 16px;border-radius:18px;font-weight:600;color:var(--text-dim)}
.comment-form .cancel:hover{background:var(--bg-3);color:var(--text)}
.comment-form .submit{padding:8px 16px;border-radius:18px;font-weight:600;background:var(--bg-3);color:var(--text-faint)}
.comment-form .submit.ready{background:var(--bg-hover);color:var(--text)}
.comment-form .submit.ready:hover{background:var(--accent-2);color:#062b4a}
.comment{display:flex;gap:12px;margin-bottom:20px}
.comment .avatar{width:40px;height:40px;flex:none}
.comment .body{flex:1;min-width:0}
.comment .meta{font-size:13px;color:var(--text-dim);margin-bottom:4px}
.comment .meta b{color:var(--text);font-weight:500;margin-right:8px}
.comment .text{font-size:14px;line-height:1.5;word-wrap:break-word;white-space:pre-wrap}
.comment .del{font-size:12px;color:var(--text-faint);margin-top:4px;cursor:pointer;background:none}
.comment .del:hover{color:var(--accent)}
.comments-empty{color:var(--text-faint);font-size:14px;padding:8px 0}
.up-next{width:400px;flex:none}
.up-next h3{font-size:16px;margin-bottom:12px}
.up-list{display:flex;flex-direction:column;gap:8px}
.up-card{display:flex;gap:8px;cursor:pointer;padding:6px;border-radius:8px}
.up-card:hover{background:var(--bg-3)}
.up-thumb{width:168px;flex:none;aspect-ratio:16/9;border-radius:8px;overflow:hidden;background:var(--bg-3);position:relative}
.up-thumb img{width:100%;height:100%;object-fit:cover}
.up-thumb .dur{position:absolute;right:4px;bottom:4px;background:rgba(0,0,0,.85);font-size:11px;padding:1px 4px;border-radius:3px}
.up-body{flex:1;min-width:0}
.up-body .t{font-size:14px;font-weight:600;line-height:1.3;display:-webkit-box;-webkit-line-clamp:2;-webkit-box-orient:vertical;overflow:hidden}
.up-body .s{font-size:12px;color:var(--text-dim);margin-top:4px}

/* ---------- skeletons / empty ---------- */
.sk{background:linear-gradient(90deg,var(--bg-3) 25%,var(--bg-hover) 37%,var(--bg-3) 63%);background-size:400% 100%;animation:sh 1.4s infinite}
@keyframes sh{0%{background-position:100% 50%}100%{background-position:0 50%}}
.sk-card{display:flex;flex-direction:column}
.sk-card .thumb{border-radius:var(--radius)}
.empty{text-align:center;padding:80px 20px;color:var(--text-faint)}
.empty svg{width:64px;height:64px;margin-bottom:16px;opacity:.5}
.cache-panel{max-width:760px;margin:32px auto 0;padding:0 4px}
.cache-head{display:flex;align-items:flex-start;justify-content:space-between;gap:18px;margin-bottom:22px}
.cache-head h1{font-size:22px;line-height:1.25;margin-bottom:6px}
.cache-head p{color:var(--text-dim);font-size:14px;line-height:1.45}
.cache-meter{height:10px;border-radius:999px;background:var(--bg-3);overflow:hidden;margin:18px 0 12px}
.cache-meter span{display:block;height:100%;width:0;background:var(--accent-2);transition:width .25s ease}
.cache-stats{display:grid;grid-template-columns:repeat(4,minmax(0,1fr));gap:10px;margin-top:16px}
.cache-stat{background:var(--bg-2);border-radius:var(--radius-sm);padding:12px}
.cache-stat b{display:block;font-size:20px;line-height:1.1}
.cache-stat span{display:block;margin-top:4px;color:var(--text-dim);font-size:12px}
.cache-current{margin-top:14px;color:var(--text-dim);font-size:13px;white-space:nowrap;overflow:hidden;text-overflow:ellipsis}
.cache-actions{display:flex;gap:10px;flex-wrap:wrap;margin-top:22px}
.action-btn{display:inline-flex;align-items:center;gap:8px;height:36px;padding:0 14px;border-radius:18px;background:var(--bg-3);font-size:14px;font-weight:600}
.action-btn:hover{background:var(--bg-hover)}
.action-btn.primary{background:var(--text);color:var(--bg)}
.action-btn.primary:hover{background:#fff}
.action-btn.danger{color:#ffb4c2}
.spinner{
  width:40px;height:40px;border:3px solid var(--bg-3);border-top-color:var(--accent-2);
  border-radius:50%;animation:spin 1s linear infinite;margin:60px auto;
}
@keyframes spin{to{transform:rotate(360deg)}}

.back{display:inline-flex;align-items:center;gap:6px;color:var(--text-dim);margin-bottom:16px;font-size:14px}
.back:hover{color:var(--text)}

@media(max-width:1000px){.sidebar{display:none}.watch{flex-direction:column}.up-next{width:100%}}
@media(max-width:600px){.logo span,.top-actions{display:none}.main{padding:16px}.cache-stats{grid-template-columns:repeat(2,minmax(0,1fr))}.cache-head{display:block}}
</style>
</head>
<body>
<header class="topbar">
  <div class="logo">
    <button class="logo-icon" aria-label="RockTube home">
      <svg viewBox="0 0 24 24" fill="white"><path d="M8 5v14l11-7z"/></svg>
    </button>
    <span class="logo-text">Rock<b>Tube</b></span>
  </div>
  <form class="search" id="searchForm">
    <div class="search-box">
      <input id="searchInput" type="text" placeholder="Search videos" autocomplete="off">
      <button class="btn" type="submit" aria-label="Search">
        <svg viewBox="0 0 24 24" width="22" height="22" fill="var(--text-dim)"><path d="M20.49 19l-5.73-5.73C15.53 12.2 16 10.91 16 9.5 16 5.91 13.09 3 9.5 3S3 5.91 3 9.5 5.91 16 9.5 16c1.41 0 2.7-.47 3.77-1.24L19 20.49 20.49 19zM5 9.5C5 7.01 7.01 5 9.5 5S14 7.01 14 9.5 11.99 14 9.5 14 5 11.99 5 9.5z"/></svg>
      </button>
    </div>
  </form>
  <div class="top-actions">
    <button class="scan-btn" title="Scan for changes" id="scanBtn" aria-label="Scan for changes">
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 12a9 9 0 0 1-15.5 6.2"/><path d="M3 12A9 9 0 0 1 18.5 5.8"/><path d="M18 2v4h4"/><path d="M6 22v-4H2"/></svg>
    </button>
    <button class="chip" title="Open cache status" id="folderChip">Local</button>
  </div>
</header>

<div class="shell">
  <aside class="sidebar">
    <div class="nav-item active" data-nav="home">
      <svg viewBox="0 0 24 24" fill="currentColor"><path d="M12 3l9 8h-3v9h-4v-6H10v6H6v-9H3z"/></svg> Home
    </div>
    <div class="nav-item" data-nav="recent">
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="9"/><path d="M12 7v5l3 2"/></svg> Recent
    </div>
    <div class="nav-item" data-nav="liked">
      <svg viewBox="0 0 24 24" fill="currentColor"><path d="M1 21h4V9H1v12zm22-11c0-1.1-.9-2-2-2h-6.31l.95-4.57.03-.32c0-.41-.17-.79-.44-1.06L14.17 1 7.59 7.59C7.22 7.95 7 8.45 7 9v10c0 1.1.9 2 2 2h9c.83 0 1.54-.5 1.84-1.22l3.02-7.05c.09-.23.14-.47.14-.73v-2z"/></svg> Liked videos
    </div>
    <div class="nav-item" data-nav="commented">
      <svg viewBox="0 0 24 24" fill="currentColor"><path d="M20 2H4c-1.1 0-2 .9-2 2v18l4-4h14c1.1 0 2-.9 2-2V4c0-1.1-.9-2-2-2zm-2 12H6v-2h12v2zm0-3H6V9h12v2zm0-3H6V6h12v2z"/></svg> Commented Videos
    </div>
    <div class="nav-divider"></div>
    <div class="nav-section">Library</div>
    <div class="nav-item" data-nav="all">
      <svg viewBox="0 0 24 24" fill="currentColor"><path d="M4 6H2v14c0 1.1.9 2 2 2h14v-2H4V6zm16-4H8c-1.1 0-2 .9-2 2v12c0 1.1.9 2 2 2h12c1.1 0 2-.9 2-2V4c0-1.1-.9-2-2-2zm-1 9h-4v4h-2v-4H9V9h4V5h2v4h4v2z"/></svg> All videos
    </div>
    <div class="nav-item" data-nav="cache">
      <svg viewBox="0 0 24 24" fill="currentColor"><path d="M12 3C7.03 3 3 4.79 3 7v10c0 2.21 4.03 4 9 4s9-1.79 9-4V7c0-2.21-4.03-4-9-4zm0 2c4.42 0 7 1.43 7 2s-2.58 2-7 2-7-1.43-7-2 2.58-2 7-2zm0 14c-4.42 0-7-1.43-7-2v-2.08C6.61 15.58 9.12 16 12 16s5.39-.42 7-1.08V17c0 .57-2.58 2-7 2zm0-5c-4.42 0-7-1.43-7-2V9.92C6.61 10.58 9.12 11 12 11s5.39-.42 7-1.08V12c0 .57-2.58 2-7 2z"/></svg> Cache
    </div>
    <div class="folder-tree" id="folderTree"></div>
  </aside>
  <main class="main" id="app"></main>
</div>

<script>
const app = document.getElementById('app');
let CACHE = [];      // all videos seen so far
let VIEW = 'home';   // current nav selection
let FOLDER_TREE = [];
let CURRENT_FOLDER = '';
let CACHE_STATUS = null;
let CACHE_TIMER = null;
let CACHE_PROMPTED = false;

const esc = s => String(s).replace(/[&<>"']/g, c => ({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;',"'":'&#39;'}[c]));
const fmtViews = n => {
  if (!n) return 'No views';
  if (n === 1) return '1 view';
  if (n < 1000) return n + ' views';
  if (n < 1e6) return (n/1000).toFixed(0) + 'K views';
  return (n/1e6).toFixed(1) + 'M views';
};
const fmtBytes = b => {
  if (!b) return '';
  const u=['B','KB','MB','GB','TB']; let i=0,f=b;
  while(f>=1024 && i<u.length-1){f/=1024;i++}
  return f.toFixed(f<10&&i>0?1:0)+' '+u[i];
};
const mediaPath = id => String(id).split('/').map(encodeURIComponent).join('/');
const folderLabel = p => p ? p.split('/').filter(Boolean).pop() : 'All videos';
const RECENT_KEY = 'rt_recent';

function shuffled(list, limit){
  const a = list.slice();
  for (let i=a.length-1;i>0;i--){
    const j = Math.floor(Math.random() * (i + 1));
    [a[i], a[j]] = [a[j], a[i]];
  }
  return a.slice(0, limit);
}

function recentIDs(){
  try {
    const ids = JSON.parse(localStorage.getItem(RECENT_KEY) || '[]');
    return Array.isArray(ids) ? ids : [];
  } catch(e){
    return [];
  }
}

function rememberRecent(id){
  const ids = recentIDs().filter(x => x !== id);
  ids.unshift(id);
  localStorage.setItem(RECENT_KEY, JSON.stringify(ids.slice(0, 20)));
}

function likedIDs(){
  const ids = [];
  for (let i=0;i<localStorage.length;i++){
    const key = localStorage.key(i);
    if (key && key.startsWith('rt_vote:') && localStorage.getItem(key) === 'like'){
      ids.push(key.slice('rt_vote:'.length));
    }
  }
  return ids;
}

function videosByIDs(ids){
  const byID = new Map(CACHE.map(v => [v.id, v]));
  return ids.map(id => byID.get(id)).filter(Boolean);
}

async function fetchVideosByIDs(ids){
  const unique = Array.from(new Set(ids.filter(Boolean)));
  if (!unique.length) return [];
  const qs = new URLSearchParams();
  unique.forEach(id => qs.append('id', id));
  const data = await api('/api/videos?' + qs.toString());
  const videos = data.videos || [];
  mergeCache(videos);
  return videosByIDs(unique);
}

// ---------- routing (hash based) ----------
function route(){
  const h = location.hash;
  if (h.startsWith('#/cache')) {
    return renderCache(false);
  }
  if (h.startsWith('#/watch')) {
    const id = new URLSearchParams(h.split('?')[1]).get('id');
    if (id) return renderWatch(id);
  }
  if (h.startsWith('#/search')) {
    const q = new URLSearchParams(h.split('?')[1]).get('q') || '';
    return renderSearch(q);
  }
  if (h.startsWith('#/folder')) {
    const folder = new URLSearchParams(h.split('?')[1]).get('path') || '';
    return renderFolder(folder);
  }
  if (!CACHE_PROMPTED && cacheNeedsWork(CACHE_STATUS)) {
    CACHE_PROMPTED = true;
    return renderCache(true);
  }
  renderHome();
}
window.addEventListener('hashchange', route);

function openWatch(id){
  const next = '#/watch?id=' + encodeURIComponent(id);
  if (location.hash === next) {
    renderWatch(id);
    return;
  }
  location.hash = next;
}

// ---------- API ----------
async function api(path, options){
  const r = await fetch(path, options);
  if (!r.ok) throw new Error('request failed');
  return r.json();
}

function cacheNeedsWork(st){
  return st && st.total > 0 && (st.running || st.metaCached < st.total || st.thumbCached < st.total);
}

function cachePercent(st){
  if (!st || !st.total) return 0;
  if (st.running) return Math.round((st.done / st.total) * 100);
  return Math.round((Math.min(st.metaCached || 0, st.thumbCached || 0) / st.total) * 100);
}

async function loadCacheStatus(){
  CACHE_STATUS = await api('/api/cache/status');
  updateCacheChip(CACHE_STATUS);
  return CACHE_STATUS;
}

function updateCacheChip(st){
  const chip = document.getElementById('folderChip');
  if (!chip || !st) return;
  if (st.running) {
    const pct = cachePercent(st);
    chip.textContent = 'Caching ' + pct + '%';
  } else if (st.total) {
    chip.textContent = st.thumbCached + '/' + st.total + ' cached';
  } else {
    chip.textContent = 'Local';
  }
}

async function startCache(force){
  CACHE_STATUS = await api('/api/cache/start?force=' + (force ? '1' : '0'), {method:'POST'});
}

async function renderCache(autoStart){
  CURRENT_FOLDER = '';
  setActiveNav('cache');
  clearInterval(CACHE_TIMER);
  app.innerHTML = cacheHTML(CACHE_STATUS || {total:0,done:0,metaCached:0,thumbCached:0,message:'Checking cache'});
  wireCacheActions();
  try {
    let st = await loadCacheStatus();
    if (autoStart && cacheNeedsWork(st) && !st.running) {
      await startCache(false);
      st = await loadCacheStatus();
    }
    drawCacheStatus(st);
    if (st.running) {
      CACHE_TIMER = setInterval(async () => {
        try {
          const next = await loadCacheStatus();
          drawCacheStatus(next);
          if (!next.running) clearInterval(CACHE_TIMER);
        } catch(e) {}
      }, 1000);
    }
  } catch(e) {
    app.innerHTML = errHTML(e);
  }
}

function cacheHTML(st){
  const pct = cachePercent(st);
  return '<div class="cache-panel">' +
    '<div class="cache-head">' +
      '<div><h1>Library cache</h1><p id="cacheMessage">'+esc(st.message || 'Checking cache')+'</p></div>' +
      '<div style="color:var(--text-dim);font-size:14px;white-space:nowrap" id="cachePct">'+pct+'%</div>' +
    '</div>' +
    '<div class="cache-meter"><span id="cacheBar" style="width:'+pct+'%"></span></div>' +
    '<div class="cache-current" id="cacheCurrent">'+esc(st.current || '')+'</div>' +
    '<div class="cache-stats">' +
      '<div class="cache-stat"><b id="cacheDone">'+(st.done||0)+'</b><span>Scanned</span></div>' +
      '<div class="cache-stat"><b id="cacheTotal">'+(st.total||0)+'</b><span>Videos</span></div>' +
      '<div class="cache-stat"><b id="cacheMeta">'+(st.metaCached||0)+'</b><span>Metadata</span></div>' +
      '<div class="cache-stat"><b id="cacheThumbs">'+(st.thumbCached||0)+'</b><span>Thumbnails</span></div>' +
    '</div>' +
    '<div class="cache-actions">' +
      '<button class="action-btn primary" id="cacheScanBtn">Scan for changes</button>' +
      '<button class="action-btn" id="cacheBrowseBtn">Browse anyway</button>' +
      '<button class="action-btn danger" id="cacheForceBtn">Rebuild all</button>' +
    '</div>' +
  '</div>';
}

function drawCacheStatus(st){
  const pct = cachePercent(st);
  const msg = st.scanning ? 'Scanning library' : (st.message || (st.running ? 'Building cache' : 'Cache ready'));
  document.getElementById('cacheMessage').textContent = msg;
  document.getElementById('cachePct').textContent = pct + '%';
  document.getElementById('cacheBar').style.width = pct + '%';
  document.getElementById('cacheCurrent').textContent = st.current || '';
  document.getElementById('cacheDone').textContent = st.done || 0;
  document.getElementById('cacheTotal').textContent = st.total || 0;
  document.getElementById('cacheMeta').textContent = st.metaCached || 0;
  document.getElementById('cacheThumbs').textContent = st.thumbCached || 0;
}

function wireCacheActions(){
  document.getElementById('cacheScanBtn').addEventListener('click', async () => {
    await startCache(false);
    renderCache(false);
  });
  document.getElementById('cacheForceBtn').addEventListener('click', async () => {
    if (!confirm('Rebuild all cached thumbnails and metadata?')) return;
    await startCache(true);
    renderCache(false);
  });
  document.getElementById('cacheBrowseBtn').addEventListener('click', () => {
    clearInterval(CACHE_TIMER);
    CACHE_STATUS = null;
    location.hash = '#/folder?path=';
  });
}

// ---------- home ----------
async function renderHome(){
  clearInterval(CACHE_TIMER);
  CURRENT_FOLDER = '';
  setActiveNav('home');
  app.innerHTML = '<div class="spinner"></div>';
  try {
    const data = await api('/api/videos?random=1');
    const videos = data.videos || [];
    mergeCache(videos);
    drawGrid(videos, pageHeadHTML('Home', '20 random videos from your library'));
  } catch(e){
    app.innerHTML = errHTML(e);
  }
}

async function renderNav(mode){
  CURRENT_FOLDER = '';
  setActiveNav(mode);
  VIEW = mode;
  if (!CACHE.length){
    app.innerHTML = '<div class="spinner"></div>';
    const data = await api('/api/videos?random=1');
    mergeCache(data.videos || []);
  }
  if (mode === 'recent'){
    app.innerHTML = '<div class="spinner"></div>';
    const videos = await fetchVideosByIDs(recentIDs());
    drawGrid(videos, pageHeadHTML('Recent', 'The last 20 videos played in this browser'));
    return;
  }
  if (mode === 'liked'){
    app.innerHTML = '<div class="spinner"></div>';
    const videos = await fetchVideosByIDs(likedIDs());
    drawGrid(videos, pageHeadHTML('Liked videos', 'Videos liked in this browser'));
    return;
  }
  if (mode === 'commented'){
    app.innerHTML = '<div class="spinner"></div>';
    const data = await api('/api/videos?commented=1');
    const videos = data.videos || [];
    mergeCache(videos);
    drawGrid(videos, pageHeadHTML('Commented Videos', videos.length + ' videos with comments'));
    return;
  }
  drawGrid(CACHE, pageHeadHTML('All videos', CACHE.length + ' videos in your library'));
}

function renderSearch(q){
  CURRENT_FOLDER = '';
  setActiveNav(null);
  app.innerHTML = '<div class="spinner"></div>';
  api('/api/search?q=' + encodeURIComponent(q)).then(data => {
    CACHE = data.videos || [];
    const head = '<div style="margin-bottom:16px;font-size:16px;color:var(--text-dim)">Results for <b style="color:var(--text)">'+esc(q)+'</b> — '+CACHE.length+' found</div>';
    drawGrid(CACHE, head);
  }).catch(e => app.innerHTML = errHTML(e));
}

async function renderFolder(folder){
  CURRENT_FOLDER = folder || '';
  setActiveNav(null);
  setActiveFolder(CURRENT_FOLDER);
  app.innerHTML = '<div class="spinner"></div>';
  try {
    const path = CURRENT_FOLDER ? '/api/videos?folder=' + encodeURIComponent(CURRENT_FOLDER) : '/api/videos';
    const data = await api(path);
    const videos = data.videos || [];
    mergeCache(videos);
    drawGrid(videos, breadcrumbsHTML(CURRENT_FOLDER));
  } catch(e){
    app.innerHTML = errHTML(e);
  }
}

function drawGrid(videos, prependHTML){
  if (!videos || !videos.length){
    app.innerHTML = (prependHTML||'') + '<div class="empty">' +
      '<svg viewBox="0 0 24 24" fill="currentColor"><path d="M4 6H2v14c0 1.1.9 2 2 2h14v-2H4V6zm16-4H8c-1.1 0-2 .9-2 2v12c0 1.1.9 2 2 2h12c1.1 0 2-.9 2-2V4c0-1.1-.9-2-2-2zm-1 9h-4v4h-2v-4H9V9h4V5h2v4h4v2z"/></svg>' +
      '<div style="font-size:18px;margin-bottom:8px">No videos here yet</div>' +
      '<div>Drop some video files in the folder and refresh.</div></div>';
    wireBreadcrumbs();
    return;
  }
  app.innerHTML = (prependHTML||'') + '<div class="grid" id="grid"></div>';
  wireBreadcrumbs();
  const grid = document.getElementById('grid');
  // skeleton first for snappy feel
  for (let i=0;i<videos.length && i<8;i++){
    const d = document.createElement('div');
    d.className = 'sk-card';
    d.innerHTML = '<div class="thumb sk"></div><div style="display:flex;gap:12px;margin-top:12px"><div class="sk" style="width:36px;height:36px;border-radius:50%"></div><div style="flex:1"><div class="sk" style="height:14px;border-radius:4px;margin-bottom:6px"></div><div class="sk" style="height:12px;width:60%;border-radius:4px"></div></div></div>';
    grid.appendChild(d);
  }
  // then real cards
  requestAnimationFrame(() => {
    grid.innerHTML = '';
    videos.forEach(v => grid.appendChild(cardEl(v)));
  });
}

function mergeCache(videos){
  const byID = new Map(CACHE.map(v => [v.id, v]));
  videos.forEach(v => byID.set(v.id, v));
  CACHE = Array.from(byID.values());
}

function breadcrumbsHTML(folder){
  const parts = (folder || '').split('/').filter(Boolean);
  let html = '<div class="crumbs"><button data-folder="">Home</button>';
  let acc = '';
  parts.forEach(part => {
    acc = acc ? acc + '/' + part : part;
    html += '<span class="sep">/</span><button data-folder="'+esc(acc)+'">'+esc(part)+'</button>';
  });
  html += '</div>';
  return html;
}

function pageHeadHTML(title, subtitle){
  return '<div class="page-head"><h1>'+esc(title)+'</h1><p>'+esc(subtitle)+'</p></div>';
}

function wireBreadcrumbs(){
  document.querySelectorAll('.crumbs button[data-folder]').forEach(btn => {
    btn.addEventListener('click', () => {
      const folder = btn.dataset.folder || '';
      location.hash = folder ? '#/folder?path=' + encodeURIComponent(folder) : '';
    });
  });
}

function cardEl(v){
  const el = document.createElement('div');
  el.className = 'card';
  const initial = esc((v.title||'?').trim().charAt(0).toUpperCase() || '?');
  const sub = v.width && v.height ? v.width+'×'+v.height+' · ' : '';
  const pathBadge = v.path ? '<div class="path-badge" title="'+esc(v.path)+'">'+esc(v.path)+'</div>' : '';
  el.innerHTML =
    '<div class="thumb">' +
      '<img loading="lazy" src="/thumb/'+mediaPath(v.id)+'" alt="" onerror="this.style.display=\'none\';this.nextElementSibling.style.display=\'grid\'">' +
      '<div class="placeholder" style="display:none"><svg viewBox="0 0 24 24" width="48" height="48" fill="currentColor"><path d="M4 6H2v14c0 1.1.9 2 2 2h14v-2H4V6zm16-4H8c-1.1 0-2 .9-2 2v12c0 1.1.9 2 2 2h12c1.1 0 2-.9 2-2V4c0-1.1-.9-2-2-2zm-2 10l-6 3.5V7l6 3.5v1z"/></svg></div>' +
      (v.duration ? '<span class="dur">'+esc(v.duration)+'</span>' : '') +
    '</div>' +
    '<div class="card-body">' +
      '<div class="avatar">'+initial+'</div>' +
      '<div class="meta">' +
        pathBadge +
        '<div class="title" title="'+esc(v.title)+'">'+esc(v.title)+'</div>' +
        '<div class="sub">'+sub+esc(v.name)+'<span>'+fmtViews(v.views)+' · '+esc(v.uploaded||'')+(v.size?' · '+fmtBytes(v.size):'')+'</span></div>' +
      '</div>' +
    '</div>';
  el.addEventListener('click', () => openWatch(v.id));
  return el;
}

// ---------- watch ----------
async function renderWatch(id){
  window.scrollTo(0,0);
  app.innerHTML = '<div class="spinner"></div>';
  let v = CACHE.find(x => x.id === id);
  if (!v) {
    try {
      const d = await api('/api/videos?id=' + encodeURIComponent(id));
      mergeCache(d.videos || []);
      v = CACHE.find(x => x.id === id);
    } catch(e) {}
  }
  v = v || {id, name: id, title: id, path: ''};
  let up = [];
  try {
    const d = await api('/api/videos?direct=1&folder=' + encodeURIComponent(v.path || ''));
    up = (d.videos || []).filter(x => x.id !== id).slice(0,10);
    mergeCache(up);
  } catch(e) {}
  rememberRecent(v.id);
  const src = '/stream/'+mediaPath(v.id);
  const initial = esc((v.title||'?').trim().charAt(0).toUpperCase() || '?');
  const subTrack = v.hasSubs ? '<track kind="subtitles" src="/subtitle/'+mediaPath(v.id)+'" srclang="en" label="Subtitles" default>' : '';
  const backFolder = CURRENT_FOLDER || v.path || '';
  const backHref = backFolder ? '#/folder?path=' + encodeURIComponent(backFolder) : '#/';
  const watchCrumbs = breadcrumbsHTML(v.path || '');

  app.innerHTML =
    '<div class="watch">' +
      '<div class="player-wrap">' +
        '<a class="back" href="'+backHref+'"><svg viewBox="0 0 24 24" width="18" height="18" fill="currentColor"><path d="M20 11H7.83l5.59-5.59L12 4l-8 8 8 8 1.41-1.41L7.83 13H20v-2z"/></svg> Back</a>' +
        watchCrumbs +
        '<div class="player"><video controls autoplay playsinline src="'+src+'">'+subTrack+'</video></div>' +
        '<div class="watch-title">'+esc(v.title)+'</div>' +
        '<div class="watch-meta">' +
          '<div class="channel">' +
            '<div class="avatar">'+initial+'</div>' +
            '<div><div class="name">'+esc(folderLabel(v.path||''))+'</div><div class="sub-count">'+esc(v.path || 'Local library')+'</div></div>' +
          '</div>' +
          '<div class="actions">' +
            '<div class="rate-pill" id="ratePill">' +
              '<button class="like" id="likeBtn" title="I like this">' +
                '<svg viewBox="0 0 24 24" fill="currentColor"><path d="M1 21h4V9H1v12zm22-11c0-1.1-.9-2-2-2h-6.31l.95-4.57.03-.32c0-.41-.17-.79-.44-1.06L14.17 1 7.59 7.59C7.22 7.95 7 8.45 7 9v10c0 1.1.9 2 2 2h9c.83 0 1.54-.5 1.84-1.22l3.02-7.05c.09-.23.14-.47.14-.73v-2z"/></svg>' +
                '<span class="cnt" id="likeCnt">'+fmtCount(v.likes)+'</span>' +
              '</button>' +
              '<span class="div"></span>' +
              '<button class="dislike" id="dislikeBtn" title="I dislike this">' +
                '<svg viewBox="0 0 24 24" fill="currentColor"><path d="M15 3H6c-.83 0-1.54.5-1.84 1.22l-3.02 7.05c-.09.23-.14.47-.14.73v2c0 1.1.9 2 2 2h6.31l-.95 4.57-.03.32c0 .41.17.79.44 1.06L9.83 23l6.59-6.59c.36-.36.58-.86.58-1.41V5c0-1.1-.9-2-2-2zm4 0v12h4V3h-4z"/></svg>' +
              '</button>' +
            '</div>' +
            '<div class="views-line"><span id="vw">'+fmtViews(v.views)+'</span> · '+esc(v.uploaded||'')+'</div>' +
          '</div>' +
        '</div>' +
        '<div class="watch-desc"><b style="color:var(--text)">File:</b> '+esc(v.name)+(v.width?' · '+v.width+'×'+v.height:'')+(v.size?' · '+fmtBytes(v.size):'')+'</div>' +
        '<div class="comments" id="comments">' +
          '<div class="comments-head"><div class="count" id="cmtCount">Comments</div></div>' +
          '<div class="comment-form">' +
            '<div class="avatar">'+userInitial()+'</div>' +
            '<div class="fields">' +
              '<input class="name" id="cmtName" type="text" placeholder="Name (optional)" maxlength="40" value="'+esc(userName())+'">' +
              '<textarea id="cmtText" placeholder="Add a comment..." rows="1" maxlength="2000"></textarea>' +
              '<div class="row">' +
                '<button class="cancel" id="cmtCancel">Cancel</button>' +
                '<button class="submit" id="cmtSubmit">Comment</button>' +
              '</div>' +
            '</div>' +
          '</div>' +
          '<div id="cmtList"></div>' +
        '</div>' +
      '</div>' +
      '<div class="up-next">' +
        '<h3>Up next</h3>' +
        '<div class="up-list" id="upList"></div>' +
      '</div>' +
    '</div>';

  const upList = document.getElementById('upList');
  if (!up.length){ upList.innerHTML = '<div style="color:var(--text-faint);font-size:14px">No other videos.</div>'; }
  up.forEach(u => {
    const d = document.createElement('div');
    d.className = 'up-card';
    d.innerHTML =
      '<div class="up-thumb"><img loading="lazy" src="/thumb/'+mediaPath(u.id)+'" alt="" onerror="this.style.opacity=.2">'+(u.duration?'<span class="dur">'+esc(u.duration)+'</span>':'')+'</div>' +
      '<div class="up-body"><div class="t">'+esc(u.title)+'</div><div class="s">'+esc(u.path ? u.path+' · ' : '')+fmtViews(u.views)+' · '+esc(u.uploaded||'')+'</div></div>';
    d.addEventListener('click', () => openWatch(u.id));
    upList.appendChild(d);
  });

  // count a view, fire-and-forget
  fetch('/api/view?id='+encodeURIComponent(v.id), {method:'POST'}).then(r=>r.json()).then(j=>{
    const el=document.getElementById('vw'); if(el) el.textContent=fmtViews(j.views||0);
  }).catch(()=>{});

  wireRating(v);
  wireComments(v);
  wireBreadcrumbs();
}

// ---- like / dislike --------------------------------------------------------
function fmtCount(n){
  if (!n) return '';
  if (n < 1000) return String(n);
  if (n < 1e6) return (n/1000).toFixed(n<10000?1:0)+'K';
  return (n/1e6).toFixed(1)+'M';
}

function wireRating(v){
  const likeBtn = document.getElementById('likeBtn');
  const dlikeBtn = document.getElementById('dislikeBtn');
  const likeCnt = document.getElementById('likeCnt');
  const likeTotal = v.likes || 0;
  const dislikeTotal = v.dislikes || 0;
  // dislike pill shows a count once there are any dislikes
  if (dislikeTotal && !dlikeBtn.querySelector('.cnt')){
    dlikeBtn.insertAdjacentHTML('beforeend','<span class="cnt">'+fmtCount(dislikeTotal)+'</span>');
  }

  // restore this browser's previous vote so the pill reflects it
  const myVote = localStorage.getItem('rt_vote:'+v.id) || '';
  setVoteUI(myVote, likeTotal, dislikeTotal);

  function setVoteUI(vote, l, d){
    likeBtn.classList.toggle('active', vote==='like');
    dlikeBtn.classList.toggle('active', vote==='dislike');
    likeCnt.textContent = fmtCount(l);
    const dCnt = dlikeBtn.querySelector('.cnt');
    if (dCnt) dCnt.textContent = fmtCount(d);
  }

  function vote(action){
    fetch('/api/rate?id='+encodeURIComponent(v.id)+'&action='+action, {method:'POST', credentials:'same-origin'})
      .then(r=>r.json())
      .then(j => {
        localStorage.setItem('rt_vote:'+v.id, j.myVote==='none' ? '' : j.myVote);
        setVoteUI(j.myVote, j.likes, j.dislikes);
      })
      .catch(()=>{});
  }

  likeBtn.addEventListener('click', () => {
    const cur = localStorage.getItem('rt_vote:'+v.id);
    vote(cur==='like' ? 'none' : 'like');
  });
  dlikeBtn.addEventListener('click', () => {
    const cur = localStorage.getItem('rt_vote:'+v.id);
    vote(cur==='dislike' ? 'none' : 'dislike');
  });
}

// ---- comments --------------------------------------------------------------
function fmtAgo(ts){
  if (!ts) return 'just now';
  const s = Math.max(1, Math.floor(Date.now()/1000 - ts));
  if (s < 60) return s+' second'+(s===1?'':'s')+' ago';
  const m = Math.floor(s/60);
  if (m < 60) return m+' minute'+(m===1?'':'s')+' ago';
  const h = Math.floor(m/60);
  if (h < 24) return h+' hour'+(h===1?'':'s')+' ago';
  const d = Math.floor(h/24);
  if (d < 30) return d+' day'+(d===1?'':'s')+' ago';
  const mo = Math.floor(d/30);
  if (mo < 12) return mo+' month'+(mo===1?'':'s')+' ago';
  return Math.floor(mo/12)+' year'+(mo<24?'':'s')+' ago';
}

function userName(){
  return localStorage.getItem('rt_user') || '';
}
function userInitial(){
  const n = (userName()||'Y').trim();
  return esc((n.charAt(0)||'Y').toUpperCase());
}

function wireComments(v){
  const list = document.getElementById('cmtList');
  const countEl = document.getElementById('cmtCount');
  const ta = document.getElementById('cmtText');
  const nameIn = document.getElementById('cmtName');
  const submit = document.getElementById('cmtSubmit');
  const cancel = document.getElementById('cmtCancel');

  function refreshUIState(){
    const has = ta.value.trim().length > 0;
    submit.classList.toggle('ready', has);
  }

  function render(list2){
    countEl.textContent = (list2.length+' Comment'+(list2.length===1?'':'s'));
    if (!list2.length){
      list.innerHTML = '<div class="comments-empty">Be the first to comment.</div>';
      return;
    }
    list.innerHTML = '';
    list2.slice().reverse().forEach(c => list.appendChild(commentEl(v, c)));
  }

  function load(){
    fetch('/api/comments?id='+encodeURIComponent(v.id), {credentials:'same-origin'})
      .then(r=>r.json())
      .then(d => render(d.comments||[]))
      .catch(()=>{});
  }

  function commentEl(v, c){
    const el = document.createElement('div');
    el.className = 'comment';
    const initial = esc((c.author||'?').trim().charAt(0).toUpperCase()||'?');
    el.innerHTML =
      '<div class="avatar">'+initial+'</div>' +
      '<div class="body">' +
        '<div class="meta"><b>'+esc(c.author||'Anonymous')+'</b>'+esc(fmtAgo(c.createdAt))+'</div>' +
        '<div class="text">'+esc(c.text)+'</div>' +
      '</div>';
    // delete button — local tool, so let any viewer remove. (Persisted server-side.)
    const del = document.createElement('button');
    del.className = 'del'; del.textContent = 'Delete';
    del.addEventListener('click', () => {
      if (!confirm('Delete this comment?')) return;
      fetch('/api/comments?id='+encodeURIComponent(v.id)+'&cid='+encodeURIComponent(c.id), {method:'DELETE', credentials:'same-origin'})
        .then(r=>r.json()).then(()=>load()).catch(()=>{});
    });
    el.querySelector('.body').appendChild(del);
    return el;
  }

  ta.addEventListener('input', () => {
    ta.style.height = 'auto';
    ta.style.height = Math.min(ta.scrollHeight, 200) + 'px';
    refreshUIState();
  });
  nameIn.addEventListener('input', () => {
    localStorage.setItem('rt_user', nameIn.value.trim());
  });
  cancel.addEventListener('click', () => { ta.value=''; ta.style.height='auto'; refreshUIState(); });
  submit.addEventListener('click', () => {
    const text = ta.value.trim();
    if (!text) return;
    const author = nameIn.value.trim();
    localStorage.setItem('rt_user', author);
    submit.disabled = true;
    fetch('/api/comments?id='+encodeURIComponent(v.id), {
      method:'POST', credentials:'same-origin',
      headers:{'Content-Type':'application/json'},
      body: JSON.stringify({author, text})
    }).then(r=>r.ok ? r.json() : Promise.reject(new Error('failed')))
      .then(() => { ta.value=''; ta.style.height='auto'; refreshUIState(); load(); })
      .catch(()=>{ alert('Could not post comment'); })
      .finally(()=>{ submit.disabled = false; });
  });

  load();
}

// ---------- nav helpers ----------
function setActiveNav(which){
  document.querySelectorAll('.nav-item').forEach(n => n.classList.remove('active'));
  if (which) setActiveFolder('__none__');
  if (which){ const m = document.querySelector('.nav-item[data-nav="'+which+'"]'); if(m) m.classList.add('active'); }
}
document.querySelectorAll('.nav-item').forEach(n => {
  n.addEventListener('click', () => {
    const v = n.dataset.nav;
    if (v === 'home'){
      CACHE_PROMPTED = true;
      if (location.hash === '') renderHome();
      else location.hash = '';
    }
    else if (v === 'cache'){ location.hash = '#/cache'; }
    else { renderNav(v); }
  });
});

async function loadFolders(){
  try {
    const data = await api('/api/folders');
    FOLDER_TREE = data.folders || [];
    renderFolderTree();
  } catch(e){
    const tree = document.getElementById('folderTree');
    if (tree) tree.innerHTML = '<div class="nav-section">Folders unavailable</div>';
  }
}

function renderFolderTree(){
  const tree = document.getElementById('folderTree');
  if (!tree) return;
  tree.innerHTML = '';
  FOLDER_TREE.forEach(node => tree.appendChild(folderNodeEl(node, true)));
  setActiveFolder(CURRENT_FOLDER);
}

function folderNodeEl(node, isRoot){
  const wrap = document.createElement('div');
  wrap.className = 'folder-node' + (isRoot ? '' : ' collapsed');
  const children = node.children || [];
  const count = node.totalCount ?? node.count ?? 0;
  wrap.innerHTML =
    '<div class="folder-row" data-folder="'+esc(node.path||'')+'">' +
      '<button class="folder-toggle" title="Expand folder" '+(!children.length?'style="visibility:hidden"':'')+'>' +
        '<svg viewBox="0 0 24 24" fill="currentColor"><path d="M7 10l5 5 5-5z"/></svg>' +
      '</button>' +
      '<div class="folder-name" title="'+esc(node.path || node.name)+'">'+esc(isRoot ? 'Folders' : node.name)+'</div>' +
      '<div class="folder-count">'+count+'</div>' +
    '</div>' +
    '<div class="folder-children"></div>';
  const row = wrap.querySelector('.folder-row');
  const toggle = wrap.querySelector('.folder-toggle');
  row.querySelector('.folder-name').addEventListener('click', () => {
    const folder = node.path || '';
    location.hash = folder ? '#/folder?path=' + encodeURIComponent(folder) : '';
  });
  toggle.addEventListener('click', e => {
    e.stopPropagation();
    wrap.classList.toggle('collapsed');
  });
  const childBox = wrap.querySelector('.folder-children');
  children.forEach(child => childBox.appendChild(folderNodeEl(child, false)));
  return wrap;
}

function setActiveFolder(folder){
  document.querySelectorAll('.folder-row').forEach(r => r.classList.toggle('active', (r.dataset.folder||'') === (folder||'')));
}

// ---------- search form ----------
document.getElementById('searchForm').addEventListener('submit', e => {
  e.preventDefault();
  const q = document.getElementById('searchInput').value.trim();
  location.hash = '#/search?q=' + encodeURIComponent(q);
});
document.getElementById('scanBtn').addEventListener('click', () => {
  location.hash = '#/cache';
});
document.getElementById('folderChip').addEventListener('click', () => {
  location.hash = '#/cache';
});

function errHTML(e){
  return '<div class="empty"><div style="font-size:18px;margin-bottom:8px;color:var(--text)">Something went wrong</div><div>'+esc(e.message)+'</div></div>';
}

// ---------- boot ----------
Promise.allSettled([loadFolders(), loadCacheStatus()]).finally(route);
</script>
</body>
</html>`

// faviconSVG is a tiny inline play-button icon.
const faviconSVG = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24"><rect width="24" height="24" rx="6" fill="#ff0033"/><path d="M9 7v10l8-5z" fill="#fff"/></svg>`

// placeholderSVG is shown when a thumbnail can't be generated (no ffmpeg, etc).
const placeholderSVG = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 320 180"><rect width="320" height="180" fill="#272727"/><path fill="#717171" d="M128 70v40l32-20z"/></svg>`
