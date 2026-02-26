package main

import "net/http"

// handleUI serves the embedded single-page test UI for aegis-server.
func handleUI(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(uiHTML))
}

const uiHTML = `<!DOCTYPE html>
<html lang="en" data-theme="light">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>Aegis/core</title>
<style>
  :root, [data-theme="light"] {
    --bg:       #ffffff;
    --bg-alt:   #fafafa;
    --surface:  #ffffff;
    --surf-alt: #f5f5f5;
    --ink:      #1a1a1a;
    --ink-2:    #444;
    --ink-3:    #777;
    --ink-4:    #aaa;
    --brd:      #b0b0b0;
    --brd-lt:   #c0c0c0;
    --accent:   #7c3aed;
    --acc-soft: #f3efff;
    --acc-dim:  #6d28d9;
    --acc-wash: rgba(124,58,237,0.07);
    --green:    #22c55e;
    --red:      #ef4444;
    --red-bg:   rgba(239,68,68,0.06);
    --mono: "Cascadia Code","Fira Code","SF Mono","Consolas",monospace;
    --sans: -apple-system,BlinkMacSystemFont,"Segoe UI",Roboto,Helvetica,Arial,sans-serif;
    --ease: cubic-bezier(0.22,1,0.36,1);
    --c-person:#7c3aed;--c-email:#6366f1;--c-phone:#ea580c;
    --c-date:#2563eb;--c-org:#2563eb;--c-iban:#dc2626;
    --c-cc:#dc2626;--c-id:#dc2626;--c-ssn:#dc2626;
    --c-ip:#0d9488;--c-url:#0d9488;--c-fin:#0d9488;
    --c-med:#db2777;--c-addr:#16a34a;--c-secret:#991b1b;--c-def:#888;
  }

  [data-theme="dark"] {
    --bg:       #161616;
    --bg-alt:   #1c1c1c;
    --surface:  #242424;
    --surf-alt: #2c2c2c;
    --ink:      #f0f0f0;
    --ink-2:    #cccccc;
    --ink-3:    #999999;
    --ink-4:    #707070;
    --brd:      #606060;
    --brd-lt:   #555555;
    --accent:   #a78bfa;
    --acc-soft: rgba(167,139,250,0.12);
    --acc-dim:  #8b5cf6;
    --acc-wash: rgba(167,139,250,0.07);
    --green:    #4ade80;
    --red:      #f87171;
    --red-bg:   rgba(248,113,113,0.1);
    --c-person:#a78bfa;--c-email:#818cf8;--c-phone:#fb923c;
    --c-date:#60a5fa;--c-org:#60a5fa;--c-iban:#f87171;
    --c-cc:#f87171;--c-id:#f87171;--c-ssn:#f87171;
    --c-ip:#2dd4bf;--c-url:#2dd4bf;--c-fin:#2dd4bf;
    --c-med:#f472b6;--c-addr:#4ade80;--c-secret:#fca5a5;--c-def:#888;
  }

  *,*::before,*::after{box-sizing:border-box;margin:0;padding:0}
  html,body{height:100%;overflow:hidden}
  body{font-family:var(--sans);background:var(--bg);color:var(--ink);
    -webkit-font-smoothing:antialiased;
    transition:background .25s var(--ease),color .25s var(--ease)}
  ::selection{background:rgba(124,58,237,0.15)}
  [data-theme="dark"] ::selection{background:rgba(167,139,250,0.25)}

  @keyframes fadeUp{from{opacity:0;transform:translateY(8px)}to{opacity:1;transform:translateY(0)}}
  @keyframes fadeIn{from{opacity:0}to{opacity:1}}
  @keyframes spin{to{transform:rotate(360deg)}}

  /* ═ App ═ */
  .app{height:100vh;display:flex;flex-direction:column}

  /* ═ Header ═ */
  .hdr{
    display:flex;align-items:center;gap:0.75rem;
    padding:0 1.25rem;height:46px;flex-shrink:0;
    border-bottom:1px solid var(--brd-lt);
    background:var(--bg);
    transition:background .25s var(--ease),border-color .25s var(--ease);
  }
  .hdr-name{font-size:0.84rem;font-weight:700;color:var(--ink);letter-spacing:-0.01em}
  .hdr-name span{color:var(--ink-3);font-weight:400}
  .hdr-spacer{flex:1}
  .hdr-btn{
    width:30px;height:30px;display:flex;align-items:center;justify-content:center;
    background:none;border:1px solid var(--brd-lt);border-radius:7px;
    cursor:pointer;color:var(--ink-3);transition:all .2s var(--ease);flex-shrink:0;
  }
  .hdr-btn:hover{background:var(--bg-alt);color:var(--ink-2);border-color:var(--brd)}
  .hdr-btn svg{width:15px;height:15px}
  .hdr-btn-text{width:auto;padding:0 0.6rem;font-size:0.7rem;font-weight:600;font-family:var(--sans)}
  .sun-i,.moon-i{display:none}
  [data-theme="light"] .sun-i{display:block}
  [data-theme="dark"] .moon-i{display:block}

  /* ═ Split ═ */
  .split{flex:1;display:grid;grid-template-columns:1fr 1fr;min-height:0;overflow:hidden}
  .col{display:flex;flex-direction:column;min-height:0;overflow:hidden}
  .col+.col{border-left:1px solid var(--brd-lt)}
  .col-head{
    display:flex;align-items:center;gap:0.5rem;
    padding:0.45rem 1.25rem;border-bottom:1px solid var(--brd-lt);
    flex-shrink:0;background:var(--bg);
    transition:background .25s var(--ease),border-color .25s var(--ease);
  }
  .col-label{font-size:0.68rem;font-weight:600;text-transform:uppercase;letter-spacing:0.05em;color:var(--ink-3)}
  .col-sub{font-size:0.64rem;color:var(--ink-4);font-weight:400}
  .col-body{flex:1;overflow-y:auto;background:var(--bg-alt);transition:background .25s var(--ease)}

  /* ═ Input (left col) ═ */
  .input-area{
    width:100%;height:100%;
    padding:1rem 1.25rem;
    font-family:var(--mono);font-size:0.84rem;line-height:1.65;
    background:transparent;color:var(--ink);
    border:none;resize:none;outline:none;
    transition:color .25s var(--ease);
  }
  .input-area::placeholder{color:var(--ink-4)}

  /* ═ Output (right col) ═ */
  .out-body{padding:1rem 1.25rem}

  .empty-state{
    display:flex;flex-direction:column;align-items:center;justify-content:center;
    height:100%;color:var(--ink-4);gap:0.4rem;
    animation:fadeIn .4s var(--ease) both;
  }
  .empty-state svg{width:36px;height:36px;opacity:0.4}
  .empty-state p{font-size:0.8rem;text-align:center;line-height:1.5}
  .empty-state b{color:var(--ink-3)}
  .hint-err{color:var(--red)}
  .hint-err svg{opacity:0.6}

  .result{animation:fadeUp .3s var(--ease) both}

  .stats{display:flex;align-items:center;gap:0.4rem;margin-bottom:0.65rem;flex-wrap:wrap}
  .pill{
    font-size:0.65rem;font-family:var(--mono);font-weight:600;
    padding:0.12rem 0.45rem;border-radius:4px;
  }
  .pill-count{background:var(--acc-soft);color:var(--accent)}
  .pill-time{background:var(--surf-alt);color:var(--ink-3)}

  .san-box{
    background:var(--surface);border:1px solid var(--brd-lt);border-radius:8px;
    padding:0.8rem 1rem;font-family:var(--mono);font-size:0.8rem;line-height:1.7;
    white-space:pre-wrap;word-break:break-word;color:var(--ink);
    margin-bottom:0.8rem;animation:fadeIn .3s var(--ease) .05s both;
  }

  .slabel{
    font-size:0.63rem;font-weight:600;text-transform:uppercase;
    letter-spacing:0.06em;color:var(--ink-3);margin-bottom:0.35rem;
  }
  .slabel+.slabel{margin-top:0.9rem}

  /* ═ Table ═ */
  .tw{border:1px solid var(--brd-lt);border-radius:8px;overflow:hidden}
  table{width:100%;border-collapse:collapse;font-size:0.78rem}
  thead{background:var(--surface)}
  th{
    text-align:left;padding:0.45rem 0.65rem;font-weight:600;font-size:0.63rem;
    text-transform:uppercase;letter-spacing:0.05em;color:var(--ink-3);
    border-bottom:1px solid var(--brd-lt);white-space:nowrap;
  }
  td{padding:0.4rem 0.65rem;border-bottom:1px solid var(--brd-lt);vertical-align:middle}
  tr:last-child td{border-bottom:none}
  tbody tr{transition:background .1s}
  tbody tr:hover td{background:var(--acc-wash)}
  .mono{font-family:var(--mono);font-size:0.76rem;word-break:break-all}

  /* ═ Tags ═ */
  .tag{
    display:inline-block;padding:0.08rem 0.42rem;border-radius:3px;
    font-size:0.64rem;font-weight:600;font-family:var(--mono);white-space:nowrap;
  }
  .tag[data-t="PERSON"]{color:var(--c-person);background:color-mix(in srgb,var(--c-person) 12%,transparent)}
  .tag[data-t="EMAIL"]{color:var(--c-email);background:color-mix(in srgb,var(--c-email) 12%,transparent)}
  .tag[data-t="PHONE"]{color:var(--c-phone);background:color-mix(in srgb,var(--c-phone) 12%,transparent)}
  .tag[data-t="DATE"]{color:var(--c-date);background:color-mix(in srgb,var(--c-date) 12%,transparent)}
  .tag[data-t="ORG"]{color:var(--c-org);background:color-mix(in srgb,var(--c-org) 12%,transparent)}
  .tag[data-t="IBAN"]{color:var(--c-iban);background:color-mix(in srgb,var(--c-iban) 12%,transparent)}
  .tag[data-t="CREDIT_CARD"]{color:var(--c-cc);background:color-mix(in srgb,var(--c-cc) 12%,transparent)}
  .tag[data-t="ID_NUMBER"]{color:var(--c-id);background:color-mix(in srgb,var(--c-id) 12%,transparent)}
  .tag[data-t="SSN"]{color:var(--c-ssn);background:color-mix(in srgb,var(--c-ssn) 12%,transparent)}
  .tag[data-t="IP_ADDRESS"]{color:var(--c-ip);background:color-mix(in srgb,var(--c-ip) 12%,transparent)}
  .tag[data-t="URL"]{color:var(--c-url);background:color-mix(in srgb,var(--c-url) 12%,transparent)}
  .tag[data-t="FINANCIAL"]{color:var(--c-fin);background:color-mix(in srgb,var(--c-fin) 12%,transparent)}
  .tag[data-t="MEDICAL"]{color:var(--c-med);background:color-mix(in srgb,var(--c-med) 12%,transparent)}
  .tag[data-t="ADDRESS"]{color:var(--c-addr);background:color-mix(in srgb,var(--c-addr) 12%,transparent)}
  .tag[data-t="LOCATION"]{color:var(--c-addr);background:color-mix(in srgb,var(--c-addr) 12%,transparent)}
  .tag[data-t="SECRET"]{color:var(--c-secret);background:color-mix(in srgb,var(--c-secret) 12%,transparent)}

  /* ═ Score ═ */
  .sc{display:inline-flex;align-items:center;gap:0.35rem;font-family:var(--mono);font-size:0.72rem;color:var(--ink-3)}
  .sc-tr{width:34px;height:3px;background:var(--brd);border-radius:2px;overflow:hidden}
  .sc-fl{height:100%;border-radius:2px}

  /* ═ Clear state ═ */
  .ok-state{
    display:flex;flex-direction:column;align-items:center;justify-content:center;
    height:100%;padding:1.5rem 1rem;animation:fadeUp .35s var(--ease) both;
  }
  .ok-state svg{width:30px;height:30px;color:var(--green);margin-bottom:0.4rem}
  .ok-label{font-size:0.84rem;font-weight:600;color:var(--green)}
  .ok-sub{font-size:0.74rem;color:var(--ink-3);margin-top:0.1rem}

  /* ═ Error ═ */
  .err{
    display:flex;align-items:flex-start;gap:0.5rem;
    background:var(--red-bg);border:1px solid rgba(239,68,68,0.18);
    border-radius:8px;padding:0.65rem 0.85rem;color:var(--red);font-size:0.82rem;
  }
  .err svg{flex-shrink:0;width:14px;height:14px;margin-top:2px}

  /* ═ Bottom bar ═ */
  .bar{
    flex-shrink:0;display:flex;align-items:center;justify-content:center;
    gap:0.5rem;padding:0.65rem 1.25rem;
    border-top:1px solid var(--brd-lt);background:var(--bg);
    transition:background .25s var(--ease),border-color .25s var(--ease);
  }
  .btn{
    display:inline-flex;align-items:center;justify-content:center;gap:0.4rem;
    height:36px;padding:0 1.4rem;font-family:var(--sans);font-size:0.8rem;font-weight:600;
    border:none;border-radius:999px;cursor:pointer;white-space:nowrap;
    transition:all .2s var(--ease);
  }
  .btn:active:not(:disabled){transform:scale(0.96)}
  .btn:disabled{opacity:0.4;cursor:not-allowed}
  .btn svg{width:14px;height:14px;flex-shrink:0}

  .btn-scan,.btn-redact{background:var(--accent);color:#fff;box-shadow:0 1px 4px rgba(124,58,237,0.2)}
  .btn-scan:hover:not(:disabled),.btn-redact:hover:not(:disabled){background:var(--acc-dim);box-shadow:0 2px 10px rgba(124,58,237,0.28);transform:translateY(-1px)}

  .btn-sub{font-size:0.65rem;color:var(--ink-4);white-space:nowrap}
  .btn-div{color:var(--ink-4);font-size:0.95rem;font-weight:300;user-select:none}

  .btn-label{display:inline-flex;align-items:center;gap:0.4rem}
  .btn-label svg{width:14px;height:14px;flex-shrink:0}
  .btn-dots{position:absolute;inset:0;display:flex;align-items:center;justify-content:center;font-size:0.8rem;letter-spacing:0.1em}
  .dot-anim::after{content:"";animation:dots 1.2s steps(4,end) infinite}
  @keyframes dots{0%{content:""}25%{content:"."}50%{content:".."}75%{content:"..."}}

  /* ═ Responsive ═ */
  @media(max-width:700px){
    .split{grid-template-columns:1fr;grid-template-rows:1fr 1fr}
    .col+.col{border-left:none;border-top:1px solid var(--brd-lt)}
    .bar{flex-direction:column;gap:0.4rem;padding:0.75rem 1.25rem}
    .btn-div{display:none}
    .btn{width:100%}
    .btn-sub{text-align:center;width:100%}
  }
</style>
</head>
<body>
<div class="app">

  <div class="hdr">
    <span class="hdr-name">Aegis<span>/core</span></span>
    <span class="hdr-spacer"></span>
    <button class="hdr-btn hdr-btn-text" id="clear-btn" aria-label="Clear all">Clear</button>
    <button class="hdr-btn" id="theme-btn" aria-label="Toggle theme">
      <svg class="sun-i" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.4" stroke-linecap="round"><circle cx="8" cy="8" r="3"/><path d="M8 1.8v1M8 13.2v1M2.2 8h1M12.8 8h1M3.8 3.8l.7.7M11.5 11.5l.7.7M3.8 12.2l.7-.7M11.5 4.5l.7-.7"/></svg>
      <svg class="moon-i" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.4" stroke-linecap="round" stroke-linejoin="round"><path d="M13.2 9.5A5.2 5.2 0 016.5 2.8a6 6 0 106.7 6.7z"/></svg>
    </button>
  </div>

  <div class="split">
    <!-- Left: Input -->
    <div class="col">
      <div class="col-head">
        <span class="col-label">Input</span>
      </div>
      <div class="col-body">
        <textarea class="input-area" id="input" spellcheck="false" autocomplete="off" placeholder="Paste or type text here&#8230;"></textarea>
      </div>
    </div>

    <!-- Right: Output -->
    <div class="col">
      <div class="col-head">
        <span class="col-label">Output</span>
        <span id="live-ind" style="display:none;width:6px;height:6px;border-radius:50%;background:var(--accent);animation:spin 1s linear infinite"></span>
      </div>
      <div class="col-body out-body" id="output">
        <div class="empty-state">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1"><rect x="3" y="4.5" width="18" height="15" rx="2.5"/><path d="M7 9.5h10M7 13.5h6" stroke-width="1.2"/></svg>
          <p>Results will appear here.</p>
        </div>
      </div>
    </div>
  </div>

  <div class="bar">
    <span class="btn-sub">Detect PII entities</span>
    <button class="btn btn-scan" id="btn-scan" onclick="doScan()">
      <span class="btn-label"><svg viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.6" stroke-linecap="round"><circle cx="6.8" cy="6.8" r="4"/><path d="M9.8 9.8l3.5 3.5"/></svg> Scan</span><span class="btn-dots" style="visibility:hidden"><span class="dot-anim"></span></span>
    </button>
    <span class="btn-div">/</span>
    <button class="btn btn-redact" id="btn-redact" onclick="doRedact()">
      <span class="btn-label"><svg viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.6" stroke-linecap="round" stroke-linejoin="round"><rect x="2.5" y="3.5" width="11" height="9" rx="1.5"/><path d="M5 7h6M5 10h3.5"/></svg> Redact</span><span class="btn-dots" style="visibility:hidden"><span class="dot-anim"></span></span>
    </button>
    <span class="btn-sub">Replace with tokens</span>
    </button>
  </div>

</div>

<script>
(function(){
  "use strict";

  var html = document.documentElement;
  var st = localStorage.getItem("aegis-theme");
  if(st==="dark"||st==="light") html.setAttribute("data-theme",st);
  else if(matchMedia("(prefers-color-scheme:dark)").matches) html.setAttribute("data-theme","dark");

  document.getElementById("theme-btn").onclick=function(){
    var n=html.getAttribute("data-theme")==="dark"?"light":"dark";
    html.setAttribute("data-theme",n);
    localStorage.setItem("aegis-theme",n);
  };

  var inputEl=document.getElementById("input");
  var output=document.getElementById("output");
  var btnS=document.getElementById("btn-scan");
  var btnR=document.getElementById("btn-redact");
  var liveInd=document.getElementById("live-ind");
  var emptyHTML=output.innerHTML;

  function resetOutput(){clr(output);output.innerHTML=emptyHTML}

  document.getElementById("clear-btn").onclick=function(){
    inputEl.value="";resetOutput();inputEl.focus();
  };

  inputEl.addEventListener("input",function(){
    if(!inputEl.value.trim())resetOutput();
  });

  function el(t,c,txt){var e=document.createElement(t);if(c)e.className=c;if(txt!==undefined)e.textContent=txt;return e}
  function clr(n){while(n.firstChild)n.removeChild(n.firstChild)}
  function svgNS(t,a){var e=document.createElementNS("http://www.w3.org/2000/svg",t);if(a)for(var k in a)e.setAttribute(k,a[k]);return e}

  function tag(type){var t=el("span","tag",type);t.setAttribute("data-t",type);return t}

  function scoreCell(td,s){
    var p=Math.round(s*100),w=el("span","sc"),tr=el("span","sc-tr"),fl=el("span","sc-fl");
    fl.style.width=p+"%";
    fl.style.background=p>=90?"var(--green)":p>=75?"var(--accent)":"var(--c-phone)";
    tr.appendChild(fl);w.appendChild(tr);w.appendChild(document.createTextNode(p+"%"));
    td.appendChild(w);
  }

  function stats(count,ms){
    var d=el("div","stats");
    d.appendChild(el("span","pill pill-count",count+" "+(count===1?"match":"matches")));
    d.appendChild(el("span","pill pill-time",ms+" ms"));
    return d;
  }

  function entTable(ents){
    var w=el("div","tw"),t=document.createElement("table");
    var th=document.createElement("thead"),hr=document.createElement("tr");
    ["Type","Text","Score","Detector"].forEach(function(c){hr.appendChild(el("th",null,c))});
    th.appendChild(hr);t.appendChild(th);
    var tb=document.createElement("tbody");
    ents.forEach(function(e){
      var tr=document.createElement("tr");
      var td1=document.createElement("td");td1.appendChild(tag(e.type));tr.appendChild(td1);
      tr.appendChild(el("td","mono",e.text));
      var td3=document.createElement("td");scoreCell(td3,e.score);tr.appendChild(td3);
      tr.appendChild(el("td","mono",e.detector));
      tb.appendChild(tr);
    });
    t.appendChild(tb);w.appendChild(t);return w;
  }

  function mapTable(maps){
    var w=el("div","tw"),t=document.createElement("table");
    var th=document.createElement("thead"),hr=document.createElement("tr");
    ["Token","Original","Type"].forEach(function(c){hr.appendChild(el("th",null,c))});
    th.appendChild(hr);t.appendChild(th);
    var tb=document.createElement("tbody");
    maps.forEach(function(m){
      var tr=document.createElement("tr");
      tr.appendChild(el("td","mono",m.token));
      tr.appendChild(el("td","mono",m.original));
      var td3=document.createElement("td");td3.appendChild(tag(m.type));tr.appendChild(td3);
      tb.appendChild(tr);
    });
    t.appendChild(tb);w.appendChild(t);return w;
  }

  function noPII(){
    var d=el("div","ok-state");
    var s=svgNS("svg",{viewBox:"0 0 24 24",fill:"none",stroke:"currentColor","stroke-width":"1.3"});
    s.appendChild(svgNS("path",{d:"M12 2L3 6.5v5.5c0 5.25 3.82 10.15 9 11.5 5.18-1.35 9-6.25 9-11.5V6.5L12 2z","stroke-linejoin":"round"}));
    s.appendChild(svgNS("path",{d:"M8.5 12.5l2.5 2.5 5-5.5","stroke-width":"1.8","stroke-linecap":"round","stroke-linejoin":"round"}));
    d.appendChild(s);
    d.appendChild(el("div","ok-label","No PII detected"));
    d.appendChild(el("div","ok-sub","No personally identifiable information found."));
    return d;
  }

  function showHint(msg){
    var d=el("div","empty-state hint-err");
    var ic=svgNS("svg",{viewBox:"0 0 24 24",fill:"none",stroke:"currentColor","stroke-width":"1"});
    ic.appendChild(svgNS("rect",{x:"3",y:"4.5",width:"18",height:"15",rx:"2.5"}));
    ic.appendChild(svgNS("path",{d:"M7 9.5h10M7 13.5h6","stroke-width":"1.2"}));
    d.appendChild(ic);d.appendChild(el("p",null,msg));
    return d;
  }

  function showErr(msg){
    clr(output);
    var box=el("div","err");
    var ic=svgNS("svg",{viewBox:"0 0 16 16",fill:"none",stroke:"currentColor","stroke-width":"1.6"});
    ic.appendChild(svgNS("circle",{cx:"8",cy:"8",r:"6"}));
    ic.appendChild(svgNS("path",{d:"M8 5.5v3","stroke-linecap":"round"}));
    ic.appendChild(svgNS("circle",{cx:"8",cy:"10.8",r:".5",fill:"currentColor",stroke:"none"}));
    box.appendChild(ic);box.appendChild(el("span",null,msg));
    output.appendChild(box);
  }

  function renderScan(data){
    var ent=data.entities||[];
    clr(output);
    if(!ent.length){output.appendChild(noPII());return}
    var d=el("div","result");
    d.appendChild(stats(ent.length,data.processing_time_ms||0));
    d.appendChild(entTable(ent));
    output.appendChild(d);
  }

  function renderRedact(data){
    var ent=data.entities||[],san=data.sanitized_text||"",maps=data.mappings||[];
    clr(output);
    if(!ent.length){output.appendChild(noPII());return}
    var d=el("div","result");
    d.appendChild(stats(ent.length,data.processing_time_ms||0));
    d.appendChild(el("div","slabel","Sanitized text"));
    d.appendChild(el("div","san-box",san));
    if(maps.length){
      d.appendChild(el("div","slabel","Mapping table"));
      d.appendChild(mapTable(maps));
    }
    output.appendChild(d);
  }

  var busy=false;

  function api(url,render,btn){
    var text=inputEl.value.trim();
    if(!text||busy){if(!text){clr(output);output.appendChild(showHint("Please enter some text."));}return}
    busy=true;btnS.disabled=true;btnR.disabled=true;
    var label=btn.querySelector(".btn-label");
    var dots=btn.querySelector(".btn-dots");
    label.style.visibility="hidden";dots.style.visibility="visible";
    liveInd.style.display="inline-block";

    fetch(url,{method:"POST",headers:{"Content-Type":"application/json"},body:JSON.stringify({text:text})})
    .then(function(r){if(!r.ok)return r.json().then(function(b){throw new Error(b.error||"HTTP "+r.status)});return r.json()})
    .then(render)
    .catch(function(e){showErr(e.message||"Connection failed.")})
    .finally(function(){busy=false;btnS.disabled=false;btnR.disabled=false;label.style.visibility="visible";dots.style.visibility="hidden";liveInd.style.display="none"});
  }

  window.doScan=function(){api("/api/scan",renderScan,btnS)};
  window.doRedact=function(){api("/api/redact",renderRedact,btnR)};

  inputEl.addEventListener("keydown",function(e){
    if((e.ctrlKey||e.metaKey)&&e.key==="Enter"){e.preventDefault();doScan()}
  });
})();
</script>
</body>
</html>`
