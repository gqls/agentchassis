-- 651_robot_hands_gripper_report_page_ROLLBACK.sql
--
-- Undoes ONLY the DOM-ready-guard change made to the js_snippets row
-- 'gripper-report-intake-widget' by 651 as committed in 991cf8b8b
-- (2026-09-03). It restores that row's js_content byte-for-byte to the value
-- that was live immediately before, which is also this file's own dump of it.
--
-- It deliberately does NOT touch the page or page_components rows 651 creates:
-- those pre-date this change, 651 guards them with NOT EXISTS, and the widget
-- fix never altered them. The full teardown (delete the page, deactivate the
-- snippet) is the comment block at the foot of 651 itself and is a different,
-- larger operation -- do not confuse the two.
--
-- WHY THIS FILE EXISTS. The council round on the guard change (correlation
-- 5775dc10-c791-4285-9f4c-249a055b5aa3) objected, correctly, that the change
-- was a production text mutation on a live content row with no dump-first step
-- and no rollback artefact. This is both: the literal below IS the pre-change
-- dump, so restoring does not depend on git still being reachable.
--
-- WHAT YOU LOSE BY RUNNING IT. The widget goes back to being INERT on
-- /gripper-report.html: it queries [data-gri-root] during <head> parse, before
-- <body> exists, gets null and silently returns, so no Start button is ever
-- drawn. That is the bug 651's change fixed -- only run this if the guard has
-- caused a worse problem than the one it solved.
--
-- AFTER RUNNING IT the DB and the served bundle disagree until a rerender
-- reships /assets/js/snippets.js. Restoring the row is not restoring the site.

BEGIN;

UPDATE js_snippets
   SET js_content = $grirb$// gripper-report-intake widget 2026-08-26. textContent-only rendering.
(function () {
  'use strict';
  var root = document.querySelector('[data-gri-root]');
  if (!root) return;
  var API = root.getAttribute('data-gri-endpoint') || '';
  if (!API) return;
  var t0 = Date.now();
  var sid = null;
  var busy = false;

  var css = '.gri-box{border:1px solid #d0d4da;border-radius:8px;padding:1rem;max-width:44rem}' +
    '.gri-log{max-height:22rem;overflow-y:auto;margin:0 0 .75rem;padding:0;list-style:none}' +
    '.gri-log li{margin:.4rem 0;padding:.55rem .8rem;border-radius:8px;max-width:90%}' +
    '.gri-a{background:#eef1f5;margin-right:auto}.gri-v{background:#dbe9ff;margin-left:auto;text-align:right}' +
    '.gri-row{display:flex;gap:.5rem}.gri-row input{flex:1;padding:.55rem;border:1px solid #b8bec8;border-radius:6px}' +
    '.gri-btn{padding:.55rem 1.1rem;border:0;border-radius:6px;background:#1d4ed8;color:#fff;cursor:pointer}' +
    '.gri-btn[disabled]{opacity:.5;cursor:default}.gri-note{font-size:.85rem;color:#556}' +
    '.gri-hp{position:absolute;left:-9999px;top:-9999px;height:1px;width:1px;overflow:hidden}' +
    '.gri-form label{display:block;margin:.5rem 0 .15rem}.gri-form input,.gri-form select{width:100%;padding:.5rem;border:1px solid #b8bec8;border-radius:6px}';
  var st = document.createElement('style');
  st.textContent = css;
  document.head.appendChild(st);

  function el(tag, cls, text) {
    var e = document.createElement(tag);
    if (cls) e.className = cls;
    if (text) e.textContent = text;
    return e;
  }

  function post(path, body) {
    return fetch(API + path, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(body || {})
    }).then(function (r) {
      return r.json().catch(function () { return {}; }).then(function (j) {
        return { status: r.status, body: j };
      });
    });
  }

  var box = el('div', 'gri-box');
  root.appendChild(box);

  var log, input, sendBtn, emailRow;

  function say(who, text) {
    var li = el('li', who === 'v' ? 'gri-v' : 'gri-a', text);
    log.appendChild(li);
    log.scrollTop = log.scrollHeight;
  }

  function start() {
    box.textContent = '';
    log = el('ul', 'gri-log');
    log.setAttribute('aria-live', 'polite');
    box.appendChild(log);
    var row = el('div', 'gri-row');
    input = el('input');
    input.maxLength = 2000;
    input.setAttribute('aria-label', 'Your answer');
    sendBtn = el('button', 'gri-btn', 'Send');
    row.appendChild(input); row.appendChild(sendBtn);
    box.appendChild(row);
    box.appendChild(el('p', 'gri-note', 'No account needed. We ask for an email only at the end.'));
    sendBtn.addEventListener('click', send);
    input.addEventListener('keydown', function (ev) { if (ev.key === 'Enter') send(); });
    busy = true;
    post('/session').then(function (r) {
      busy = false;
      if (r.status !== 200 || !r.body.session_id) return degrade();
      sid = r.body.session_id;
      say('a', r.body.greeting || 'Tell me about your application.');
      input.focus();
    }).catch(degrade);
  }

  function send() {
    var msg = (input.value || '').trim();
    if (!msg || busy || !sid) return;
    busy = true; sendBtn.disabled = true;
    say('v', msg);
    input.value = '';
    post('/chat', { session_id: sid, message: msg }).then(function (r) {
      busy = false; sendBtn.disabled = false;
      if (r.status === 429) return say('a', 'One moment - a little too fast for me. Please send that again.');
      if (r.status === 503) return degrade();
      if (r.status !== 200) return say('a', (r.body && r.body.error) || 'Something went wrong - please try that again.');
      say('a', r.body.reply || '');
      if (r.body.complete && !emailRow) showEmail();
      input.focus();
    }).catch(function () { busy = false; sendBtn.disabled = false; degrade(); });
  }

  function showEmail() {
    emailRow = el('div', 'gri-row');
    var em = el('input');
    em.type = 'email'; em.placeholder = 'you@company.com';
    em.setAttribute('aria-label', 'Email for the dossier link');
    var go = el('button', 'gri-btn', 'Email me the dossier');
    emailRow.appendChild(em); emailRow.appendChild(go);
    box.appendChild(emailRow);
    go.addEventListener('click', function () {
      var v = (em.value || '').trim();
      if (!v) { em.focus(); return; }
      go.disabled = true;
      submit({ session_id: sid, email: v }, function () { go.disabled = false; });
    });
  }

  function submit(payload, onFail) {
    payload._elapsed = Date.now() - t0;
    payload.company_website = hpVal();
    post('/submit', payload).then(function (r) {
      if (r.status === 200 && r.body.accepted) return done();
      if (onFail) onFail();
      say('a', (r.body && r.body.error) || 'That did not go through - please check the email address and try again.');
    }).catch(function () { if (onFail) onFail(); });
  }

  var hp;
  function hpVal() { return hp ? hp.value : ''; }
  function addHp(parent) {
    hp = el('input', 'gri-hp');
    hp.name = 'company_website';
    hp.tabIndex = -1;
    hp.setAttribute('autocomplete', 'off');
    hp.setAttribute('aria-hidden', 'true');
    parent.appendChild(hp);
  }

  function done() {
    box.textContent = '';
    box.appendChild(el('p', null, 'Request received. The dossier link normally arrives by email within the hour.'));
  }

  var MATERIALS = ['steel', 'aluminium', 'plastic', 'glass', 'cardboard', 'rubber'];
  function degrade() {
    box.textContent = '';
    var f = el('div', 'gri-form');
    f.appendChild(el('p', null, 'The chat assistant is unavailable just now. The short form below does the same job.'));
    function fld(label, name, type) {
      f.appendChild(el('label', null, label));
      var i = el(type === 'select' ? 'select' : 'input');
      if (type && type !== 'select') i.type = type;
      i.setAttribute('data-gri-f', name);
      if (type === 'select') MATERIALS.forEach(function (m) { i.appendChild(el('option', null, m)); });
      f.appendChild(i);
      return i;
    }
    fld('Workpiece mass, kg', 'mass_kg', 'number');
    fld('Part shape and dimensions, mm', 'part_geometry', 'text');
    fld('Jaw opening needed, mm (usually the width you grip across)', 'travel_mm', 'number');
    fld('Surface material', 'surface_material', 'select');
    fld('Minimum IP rating, if any', 'ip_min', 'number');
    fld('Picks per minute', 'cycle_rate', 'number');
    fld('Robot or arm, and flange standard if known', 'mounting', 'text');
    fld('Email for the dossier link', 'gri_email', 'email');
    addHp(f);
    var go = el('button', 'gri-btn', 'Request the dossier');
    go.style.marginTop = '.75rem';
    f.appendChild(go);
    var msg = el('p', 'gri-note', '');
    f.appendChild(msg);
    box.appendChild(f);
    go.addEventListener('click', function () {
      var spec = {};
      var email = '';
      f.querySelectorAll('[data-gri-f]').forEach(function (i) {
        var k = i.getAttribute('data-gri-f');
        var v = (i.value || '').trim();
        if (k === 'gri_email') { email = v; return; }
        if (!v) return;
        spec[k] = (i.type === 'number') ? Number(v) : v;
      });
      if (!email) { msg.textContent = 'The email address is required - it is how the dossier reaches you.'; return; }
      go.disabled = true;
      post('/submit', { email: email, spec: spec, _elapsed: Date.now() - t0, company_website: hpVal() }).then(function (r) {
        if (r.status === 200 && r.body.accepted) return done();
        go.disabled = false;
        msg.textContent = (r.body && r.body.error) || 'That did not go through - please check the fields and try again.';
      }).catch(function () { go.disabled = false; msg.textContent = 'Network problem - please try again.'; });
    });
  }

  var intro = el('p', null, 'Describe your pick-and-place application in a short chat. We compute the physics against our published gripper index and email you a dossier, normally within the hour.');
  var startBtn = el('button', 'gri-btn', 'Start');
  box.appendChild(intro);
  box.appendChild(startBtn);
  addHp(box);
  startBtn.addEventListener('click', start);
})();
$grirb$
 WHERE name = 'gripper-report-intake-widget';

-- Verify (DO/RAISE so a failure ABORTS -- a verify block of bare SELECTs
-- cannot stop the COMMIT; see LANDMINES).
DO $$
DECLARE
  l int; guarded boolean;
BEGIN
  SELECT octet_length(js_content), js_content LIKE '%DOMContentLoaded%'
    INTO l, guarded
    FROM js_snippets WHERE name = 'gripper-report-intake-widget';

  IF l IS NULL THEN
    RAISE EXCEPTION '651_ROLLBACK: snippet row is missing';
  END IF;
  IF l <> 8175 THEN
    RAISE EXCEPTION '651_ROLLBACK: restored octet_length is %, expected 8175', l;
  END IF;
  IF guarded THEN
    RAISE EXCEPTION '651_ROLLBACK: the DOM-ready guard is still present after restore';
  END IF;

  RAISE NOTICE '651_ROLLBACK verified: widget restored to the pre-guard body (8175 B, unguarded)';
END $$;

COMMIT;
