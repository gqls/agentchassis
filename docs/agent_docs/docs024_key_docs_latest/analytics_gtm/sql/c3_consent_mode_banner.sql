-- c3_consent_mode_banner.sql — Consent Mode v2 defaults (denied) + self-contained consent banner,
-- inserted INSIDE the {{if .gtm_container_id}} gate of the three head templates, BEFORE the GTM
-- loader. Owner decision 2026-09-02 ("let's do A"). bugs_open/397 lane (analytics_gtm).
--
-- Proven BEFORE this file touches anything: 26/26 behavioural assertions in headless Chromium
-- against the LIVE container (no _ga before consent; _ga only after Accept; wiped on withdrawal;
-- both choices persist) — analytics_gtm/NOTES §28. The block references no template variables, so
-- input_schema is untouched. Fail-safe: a broken banner leaves consent DENIED (less tracking).
--
-- ⚠ APPLYING THIS IS A FLEET REBUILD: the template md5 is in ChromeRenderInputsSQL, so every site
-- using these templates goes stale_chrome at its next discovery pass and fully re-renders.
--
-- Usage: -v GO=yes to run; add -v DRY=1 to roll back. Idempotent: refuses if the marker exists.
\set ON_ERROR_STOP on
\if :{?GO}
\else
  \echo 'REFUSED: pass -v GO=yes (and -v DRY=1 first).'
  DO $r$ BEGIN RAISE EXCEPTION 'refused without -v GO=yes'; END $r$;
\endif

BEGIN;

CREATE TEMP TABLE _cns(v text) ON COMMIT DROP;
INSERT INTO _cns VALUES ($c3$
<!-- Consent Mode v2 + banner (PECR; owner decision 2026-09-02). Must sit INSIDE the
     gtm_container_id gate, BEFORE the GTM loader. Self-contained; no external requests. -->
<script>
window.dataLayer=window.dataLayer||[];
function gtag(){dataLayer.push(arguments);}
gtag('consent','default',{ad_storage:'denied',ad_user_data:'denied',ad_personalization:'denied',analytics_storage:'denied'});
(function(){
var KEY='cc_v1',doc=document,ch='';
try{ch=localStorage.getItem(KEY)||'';}catch(e){}
if(ch==='g'){gtag('consent','update',{analytics_storage:'granted'});}
function save(v){try{localStorage.setItem(KEY,v);}catch(e){}}
function wipeGa(){
  var parts=doc.cookie.split(';'),i,name,host=location.hostname,doms=[host,'.'+host],ds,exp='; expires=Thu, 01 Jan 1970 00:00:00 GMT; path=/';
  if(host.indexOf('.')>0){doms.push('.'+host.split('.').slice(-2).join('.'));}
  for(i=0;i<parts.length;i++){
    name=parts[i].split('=')[0].replace(/^\s+/,'');
    if(name.indexOf('_ga')===0){
      doc.cookie=name+'='+exp;
      for(ds=0;ds<doms.length;ds++){doc.cookie=name+'='+exp+'; domain='+doms[ds];}
    }
  }
}
function el(tag,css,txt){var e=doc.createElement(tag);if(css)e.style.cssText=css;if(txt)e.textContent=txt;return e;}
var BTN='font:inherit;padding:9px 22px;border-radius:5px;cursor:pointer;min-width:110px';
function choose(v){
  save(v);
  if(v==='g'){gtag('consent','update',{analytics_storage:'granted'});}
  else{gtag('consent','update',{analytics_storage:'denied'});wipeGa();}
  var b=doc.getElementById('cc-bar');if(b)b.remove();
  pill();
}
function bar(){
  if(doc.getElementById('cc-bar'))return;
  var w=el('div','position:fixed;left:0;right:0;bottom:0;z-index:2147483000;background:#1c1c28;color:#fff;font:15px/1.5 system-ui,sans-serif;padding:14px 18px;display:flex;flex-wrap:wrap;gap:12px;align-items:center;justify-content:center;box-shadow:0 -2px 12px rgba(0,0,0,.35)');
  w.id='cc-bar';w.setAttribute('role','dialog');w.setAttribute('aria-label','Cookie consent');
  w.appendChild(el('span','max-width:560px','We would like to count visits using Google Analytics. No cookies are set unless you accept.'));
  var ok=el('button',BTN+';background:#fff;color:#1c1c28;border:1px solid #fff','Accept');
  var no=el('button',BTN+';background:none;color:#fff;border:1px solid #fff','No thanks');
  ok.onclick=function(){choose('g');};
  no.onclick=function(){choose('d');};
  w.appendChild(ok);w.appendChild(no);
  doc.body.appendChild(w);
}
function pill(){
  if(doc.getElementById('cc-pill'))return;
  var p=el('button','position:fixed;left:10px;bottom:10px;z-index:2147482999;background:#1c1c28;color:#fff;font:12px system-ui,sans-serif;padding:5px 10px;border-radius:12px;border:1px solid #555;cursor:pointer;opacity:.75','Cookies');
  p.id='cc-pill';p.setAttribute('aria-label','Cookie settings');
  p.onclick=function(){var b=doc.getElementById('cc-bar');if(b){b.remove();}else{bar();}};
  doc.body.appendChild(p);
}
function init(){if(ch===''){bar();}pill();}
if(doc.readyState==='loading'){doc.addEventListener('DOMContentLoaded',init);}else{init();}
})();
</script>
$c3$);

DO $pre$
DECLARE r record; n int;
BEGIN
  FOR r IN SELECT id, name, html_template FROM content_components
            WHERE id IN ('116c5f91-bc0d-439d-9e13-a3ba2d145571','aec98dbe-76b7-4e13-9641-e5b6ba2502aa','14cf6193-c8f0-4640-9cf1-f8b5347e6885') LOOP
    n := (length(r.html_template) - length(replace(r.html_template, '{{if .gtm_container_id}}<!-- Google Tag Manager -->',''))) / length('{{if .gtm_container_id}}<!-- Google Tag Manager -->');
    IF n <> 1 THEN RAISE EXCEPTION 'pre: anchor occurs % times in % (want 1) — template drifted, re-read before editing', n, r.name; END IF;
    IF position('cc_v1' in r.html_template) > 0 THEN RAISE EXCEPTION 'pre: consent block already present in % — idempotency refusal', r.name; END IF;
  END LOOP;
END $pre$;

UPDATE content_components cc
   SET html_template = replace(cc.html_template,
         '{{if .gtm_container_id}}<!-- Google Tag Manager -->',
         '{{if .gtm_container_id}}' || (SELECT v FROM _cns) || '<!-- Google Tag Manager -->'),
       updated_at = now()
 WHERE cc.id IN ('116c5f91-bc0d-439d-9e13-a3ba2d145571','aec98dbe-76b7-4e13-9641-e5b6ba2502aa','14cf6193-c8f0-4640-9cf1-f8b5347e6885');

DO $post$
DECLARE r record; a int; b int;
BEGIN
  FOR r IN SELECT id, name, html_template FROM content_components
            WHERE id IN ('116c5f91-bc0d-439d-9e13-a3ba2d145571','aec98dbe-76b7-4e13-9641-e5b6ba2502aa','14cf6193-c8f0-4640-9cf1-f8b5347e6885') LOOP
    a := (length(r.html_template) - length(replace(r.html_template,'cc_v1','')))/length('cc_v1');
    IF a <> 1 THEN RAISE EXCEPTION 'post: cc_v1 occurs % times in % (want 1: the KEY constant)', a, r.name; END IF;
    b := position('consent' in r.html_template);
    IF b = 0 OR b > position('googletagmanager.com/gtm.js' in r.html_template)
      THEN RAISE EXCEPTION 'post: consent default is not BEFORE the GTM loader in %', r.name; END IF;
    IF (length(r.html_template) - length(replace(r.html_template,'{{if .gtm_container_id}}','')))/length('{{if .gtm_container_id}}') <> 1
      THEN RAISE EXCEPTION 'post: gate count changed in %', r.name; END IF;
  END LOOP;
END $post$;

SELECT name, length(html_template) AS tpl_len, updated_at FROM content_components
 WHERE id IN ('116c5f91-bc0d-439d-9e13-a3ba2d145571','aec98dbe-76b7-4e13-9641-e5b6ba2502aa','14cf6193-c8f0-4640-9cf1-f8b5347e6885');

\if :{?DRY}
  ROLLBACK;
  \echo 'DRY RUN — rolled back.'
\else
  COMMIT;
  \echo 'APPLIED. Every site on these templates goes stale_chrome at its next discovery pass and re-renders with consent-gated GTM.'
\endif
