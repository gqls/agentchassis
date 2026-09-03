-- 651_robot_hands_gripper_report_page.sql
-- robot-hands.com: the /gripper-report.html page + section component + widget
-- snippet — DESIGN §2 "Site side" of the gripper-dossier pilot, built after the
-- intake API went live end to end on 2026-08-26 (bugs_closed/409 hardening
-- included: hedge guard + travel/jaw-span reconcile, both live on tools-api
-- v1.0.1343).
--
-- DELIBERATE DEVIATIONS FROM DESIGN §2, recorded:
--   * url is /gripper-report.html (flat, .html) — the DESIGN's /gripper-report/
--     predates contact with this site's real url convention (cf. /about.html).
--   * in_footer=FALSE and noindex=TRUE — the soft-launch decision (unlinked vs
--     footer link) is the OWNER'S and is still pending; flipping is one UPDATE:
--     UPDATE pages SET in_footer=true, noindex=false WHERE site_id=<s> AND name='gripper-report';
--     (plus a site_nav_items row if the footer nav is table-driven for this site).
--
-- PROTECTION follows 249_oufe_legal_pages: rebuild_policy='owned' PLUS
-- lock_type='permanent' PLUS rendered_html written IN THE SAME STATEMENT —
-- save_page_sections PRESERVES locked rows rather than rendering them, so a
-- locked row without rendered_html serves as nothing for ever.
--
-- The retention facts in the page copy are VERIFIED against the live poller
-- (internal/tools-api/gripper/poller.go: transcript GC 24h idle; email/ip/ua
-- nulled 90 days after terminal) — if retention changes, this copy changes in
-- the same commit.
--
-- Idempotent: page/component INSERTs are guarded by NOT EXISTS; the widget
-- snippet UPSERTS on name so a re-run refreshes js_content.

BEGIN;

DO $$
DECLARE
  s   uuid;
  cc  uuid;
  pg  uuid;
  rh  text;
BEGIN
  SELECT id INTO s FROM sites WHERE domain = 'robot-hands.com';
  IF s IS NULL THEN RAISE EXCEPTION 'no site row for robot-hands.com'; END IF;

  SELECT id INTO cc FROM content_components WHERE function = 'gripper-report-intake' LIMIT 1;
  IF cc IS NULL THEN
    cc := gen_random_uuid();
    INSERT INTO content_components (id, name, description, html_template, function,
                                    display_name, category, component_level, section_type,
                                    render_mode, is_active, created_from)
    VALUES (cc, 'gripper-report-intake',
            'Chat intake for the gripper selection dossier (island API at tools.apis.uk). The page instance is locked-permanent with rendered_html seeded, so this template is a record of the markup, not a render source.',
            $sec$<section class="gripper-report-intake" data-component="gripper-report-intake">
  <div class="container">
    <h1>Gripper Selection Dossier</h1>
    <p>Tell us about one pick-and-place application in a short chat: the part, its weight, the speed, the mounting. We score it deterministically against the same published gripper index that drives every comparison on this site, write up what the numbers say, and email you a link to the finished dossier &mdash; normally within the hour.</p>
    <p>The dossier is computed and written automatically. Treat it as a shortlisting aid, not an engineering sign-off &mdash; check the numbers that matter against the manufacturer&rsquo;s datasheet before ordering.</p>
    <p>What happens with your details: the email address is used once, to send the link. Chat transcripts are deleted 24 hours after the conversation goes quiet, and the email address is removed from our records 90 days after the request completes.</p>
    <div data-gri-root data-gri-endpoint="https://tools.apis.uk/api/v1/tools/gripper"></div>
    <noscript><p>The intake needs JavaScript. If you would rather not enable it, the contact page works too &mdash; mention the part, weight, jaw opening, material, picks per minute and mounting.</p></noscript>
  </div>
</section>$sec$,
            'gripper-report-intake', 'Gripper dossier intake', 'tool',
            'section', 'gripper-report-intake', 'template', true, 'manual');
  END IF;

  rh := $sec$<section class="gripper-report-intake" data-component="gripper-report-intake">
  <div class="container">
    <h1>Gripper Selection Dossier</h1>
    <p>Tell us about one pick-and-place application in a short chat: the part, its weight, the speed, the mounting. We score it deterministically against the same published gripper index that drives every comparison on this site, write up what the numbers say, and email you a link to the finished dossier &mdash; normally within the hour.</p>
    <p>The dossier is computed and written automatically. Treat it as a shortlisting aid, not an engineering sign-off &mdash; check the numbers that matter against the manufacturer&rsquo;s datasheet before ordering.</p>
    <p>What happens with your details: the email address is used once, to send the link. Chat transcripts are deleted 24 hours after the conversation goes quiet, and the email address is removed from our records 90 days after the request completes.</p>
    <div data-gri-root data-gri-endpoint="https://tools.apis.uk/api/v1/tools/gripper"></div>
    <noscript><p>The intake needs JavaScript. If you would rather not enable it, the contact page works too &mdash; mention the part, weight, jaw opening, material, picks per minute and mounting.</p></noscript>
  </div>
</section>$sec$;

  IF NOT EXISTS (SELECT 1 FROM pages WHERE site_id = s AND name = 'gripper-report') THEN
    pg := gen_random_uuid();
    INSERT INTO pages (id, site_id, name, url, title, page_type, status, build_status,
                       meta_description, nav_label, nav_order, in_header, in_footer,
                       sections, rebuild_policy, noindex, deployed_at)
    VALUES (pg, s, 'gripper-report', '/gripper-report.html',
            'Gripper Selection Dossier', 'content', 'active', 'needs_rebuild',
            'Describe your pick-and-place application in a short chat and get an emailed dossier scoring our published gripper index against your numbers.',
            'Gripper Dossier', 50, false, false,
            '["gripper-report-intake"]'::jsonb, 'owned', true, NULL);

    INSERT INTO page_components (id, page_id, slot_name, component_id, content_data, rendered_html,
                                 build_status, position, locked_at, locked_by, lock_type)
    VALUES (gen_random_uuid(), pg, 'gripper-report-intake', cc,
            jsonb_build_object(
              'heading', 'Gripper Selection Dossier',
              'endpoint', 'https://tools.apis.uk/api/v1/tools/gripper',
              'note', 'locked-permanent: rendered_html is the served artefact; content_data is the record'),
            rh, 'pending', 0, now(), '651_robot_hands_gripper_report_page', 'permanent');
  END IF;
  -- The locked row is guarded by NOT EXISTS above, so a re-apply would never
  -- heal its markup. The seed is the OWNER of this section's content (owned
  -- page, section-editor edits only): converge the live row to this file's
  -- markup on every apply. Added 2026-08-26 with the council round's caveat
  -- sentence (corr de0068fd, compliance advisory).
  UPDATE page_components pc
     SET rendered_html = rh
    FROM pages p
   WHERE p.id = pc.page_id AND p.site_id = s AND p.name = 'gripper-report'
     AND pc.slot_name = 'gripper-report-intake'
     AND pc.rendered_html IS DISTINCT FROM rh;
END $$;

INSERT INTO js_snippets (name, description, js_content, applies_to, is_active)
VALUES ('gripper-report-intake-widget',
        'Chat widget for the gripper dossier intake on robot-hands.com. One IIFE, textContent-only rendering, honeypot + elapsed gate, degraded plain-form mode on 503. <=8KB by design (DESIGN section 2).',
        $grijs$(function () {
  'use strict';
  function init() {
  var root = document.querySelector('[data-gri-root]');
  if (!root) return;
  var API = root.getAttribute('data-gri-endpoint') || '';
  if (!API) return;
  var t0 = Date.now();
  var sid = null;
  var busy = false;

  var css = '.gri-box{border:1px solid #d0d4da;border-radius:8px;padding:1rem;max-width:44rem}.gri-log{max-height:22rem;overflow-y:auto;margin:0 0 .75rem;padding:0;list-style:none}.gri-log li{margin:.4rem 0;padding:.55rem .8rem;border-radius:8px;max-width:90%;color:#111}.gri-a{background:#eef1f5;margin-right:auto}.gri-v{background:#dbe9ff;margin-left:auto;text-align:right}.gri-row{display:flex;gap:.5rem}.gri-row input{flex:1;padding:.55rem;border:1px solid #b8bec8;border-radius:6px}' +
    '.gri-btn{padding:.55rem 1.1rem;border:0;border-radius:6px;background:#1d4ed8;color:#fff;cursor:pointer}.gri-btn[disabled]{opacity:.5;cursor:default}.gri-note{font-size:.85rem;opacity:.7}.gri-hp{position:absolute;left:-9999px;top:-9999px;height:1px;width:1px;overflow:hidden}.gri-form label{display:block;margin:.5rem 0 .15rem}.gri-form input,.gri-form select{width:100%;padding:.5rem;border:1px solid #b8bec8;border-radius:6px}';
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
  }
  if (document.readyState === 'loading') document.addEventListener('DOMContentLoaded', init);
  else init();
})();
$grijs$,
        '["gripper-report-intake"]'::jsonb, true)
ON CONFLICT (name) DO UPDATE
  SET js_content = EXCLUDED.js_content,
      description = EXCLUDED.description,
      applies_to  = EXCLUDED.applies_to,
      is_active   = true;

-- Verify (DO/RAISE so a failure ABORTS the transaction, per LANDMINES).
DO $$
DECLARE
  s uuid; n int; l int;
BEGIN
  SELECT id INTO s FROM sites WHERE domain = 'robot-hands.com';

  SELECT count(*) INTO n FROM pages
   WHERE site_id = s AND name = 'gripper-report' AND rebuild_policy = 'owned'
     AND in_footer = false AND noindex = true;
  IF n <> 1 THEN RAISE EXCEPTION '651: page row wrong (count %)', n; END IF;

  SELECT count(*) INTO n FROM page_components pc
    JOIN pages p ON p.id = pc.page_id
   WHERE p.site_id = s AND p.name = 'gripper-report'
     AND pc.lock_type = 'permanent'
     AND pc.rendered_html LIKE '%data-gri-root%'
     AND length(pc.rendered_html) > 500;
  IF n <> 1 THEN RAISE EXCEPTION '651: locked component with mount div missing (count %)', n; END IF;

  SELECT octet_length(js_content) INTO l FROM js_snippets
   WHERE name = 'gripper-report-intake-widget' AND is_active;
  IF l IS NULL OR l < 4000 OR l > 8192 THEN
    RAISE EXCEPTION '651: widget snippet missing or out of the <=8KB bound (octet_length %)', l;
  END IF;

  -- The bundle-selection join the renderer actually runs (render_js_snippets_
  -- for_site_action loadJSSnippetsForSite): the snippet must be selectable for
  -- THIS site via the component-function overlap, or it ships to nobody.
  SELECT count(*) INTO n
    FROM js_snippets js
   WHERE js.name = 'gripper-report-intake-widget' AND js.is_active
     AND EXISTS (
       SELECT 1 FROM jsonb_array_elements_text(js.applies_to) a(elem)
        WHERE a.elem IN (
          SELECT DISTINCT cc.function
            FROM page_components pc
            JOIN content_components cc ON cc.id = pc.component_id
            JOIN pages p ON p.id = pc.page_id
           WHERE p.site_id = s));
  IF n <> 1 THEN RAISE EXCEPTION '651: widget not selectable for robot-hands.com by the renderer join'; END IF;

  RAISE NOTICE '651 verified: page + locked mount section + active <=8KB widget, selectable for robot-hands.com';
END $$;

COMMIT;

-- To undo (the page was never linked):
--   DELETE FROM page_components WHERE page_id=(SELECT id FROM pages WHERE name='gripper-report' AND site_id=(SELECT id FROM sites WHERE domain='robot-hands.com'));
--   DELETE FROM pages WHERE name='gripper-report' AND site_id=(SELECT id FROM sites WHERE domain='robot-hands.com');
--   UPDATE js_snippets SET is_active=false WHERE name='gripper-report-intake-widget';
