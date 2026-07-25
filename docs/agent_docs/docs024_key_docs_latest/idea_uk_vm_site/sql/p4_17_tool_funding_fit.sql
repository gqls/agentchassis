-- p4_17_tool_funding_fit.sql — idea.uk's SECOND free (Tier-1) tool: the funding-fit finder.
--
-- Owner 2026-07-25: "please go ahead with both" (this + the /report.html copy pass, p4_16).
-- This is the tool features_open/014 sketched as "funding-fit finder (which funding route/source
-- matches your stage)", and the second application of the rule the patent checker established
-- and 014 now records: **where a single answer can be decisive, the instrument must GATE before
-- it scores.** Funding has two such answers, and both are checked before anything composes:
--   GATE 1 — "the money is for living costs": almost nothing funds runway, and a sum-score
--            would cheerfully route a runway-seeker to Innovate UK. The honest verdict is its
--            own outcome.
--   GATE 2 — "no evidence yet": every route on the map asks the demand question first; the
--            verdict is "buy evidence cheaply", funnelled at the testing/building guides + the
--            report, not a funder list.
-- Past both gates the remaining answers COMPOSE a route map (grants / equity / debt / customer
-- money) rather than picking a single winner — funding routes are not mutually exclusive, which
-- is also why this is not a scored quiz.
--
-- CONTENT POLICY: same as the funding guides (p4_10/p4_11) — durable institutions only, zero
-- amounts/rates/deadlines, free front doors emphasised, paid-intermediary warning included.
-- Everything the verdicts name is already named in the guides.
--
-- DELIVERY: identical posture to patent-check (p4_06) — zero backend, zero LLM, nothing leaves
-- the browser; JS inline in html_template (off the bugs_open/041 publish-but-never-load path);
-- template parse+execute verified locally before insert (0 <no value>, 0 unreplaced,
-- </section> present); schema fields all source=static with NO fallback (the p4_04 rule);
-- /tools/ confirmed static in nginx (probed for p4_06). page_type='tool' → /tools.html lists it
-- automatically; nav_order 20 (report 3, patent-check 10, audience-check 30).
--
-- One template difference from patent-check: question 7 is CHECKBOXES (18–30 / social purpose /
-- devolved nation) because those are not mutually exclusive — validation requires only the six
-- radio questions.

\set ON_ERROR_STOP on

BEGIN;

DO $guard$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM pages
   WHERE site_id = '1244516d-014d-421c-88c6-090bb1e9552a'
     AND url = '/tools/funding-fit/index.html';
  IF n > 0 THEN
    RAISE EXCEPTION 'ABORT: /tools/funding-fit/index.html already exists — p4_17 already ran.';
  END IF;
  SELECT count(*) INTO n FROM content_components WHERE name = 'funding-fit';
  IF n > 0 THEN
    RAISE EXCEPTION 'ABORT: a component named funding-fit already exists.';
  END IF;
END
$guard$;

INSERT INTO content_components
  (name, display_name, description, function, section_type, component_level, render_mode,
   category, is_active, created_from, suitable_site_types, suitable_page_types, html_template, input_schema)
VALUES (
  'funding-fit',
  'Which funding route fits? — free finder',
  'Client-side, zero-cost funding-route finder. GATED, not scored (the p4_06 rule): "money for '
  || 'living costs" and "no evidence yet" are dispositive and short-circuit to honest verdicts '
  || 'before any composition happens; past the gates the answers COMPOSE a route map (grants/'
  || 'equity/debt/customer money) because routes are not mutually exclusive. Durable institutions '
  || 'only, no amounts or scheme details. No backend, no LLM, nothing leaves the browser. Script '
  || 'inline in the template on purpose (see p4_17).',
  'funding-fit',
  'funding-fit',
  'section',
  'template',
  'tool',
  true,
  'manual',
  '[]'::jsonb,
  '[]'::jsonb,
  $tpl$<section class="funding-fit" data-component="funding-fit" id="funding-fit">
  <div class="ff-inner">
    <span class="ff-eyebrow">{{.eyebrow_label}}</span>
    <h2 class="ff-heading">{{.section_heading}}</h2>
    <p class="ff-intro">{{.section_intro}}</p>

    <form class="ff-form" id="ffForm" novalidate>
      <ol class="ff-questions">

        <li class="ff-q" data-q="stage">
          <p class="ff-q-text">1. Where is the idea, honestly?</p>
          <p class="ff-q-help">Not where you hope it will be — where it is today.</p>
          <label><input type="radio" name="stage" value="idea"> An idea — no evidence yet that anyone wants it</label>
          <label><input type="radio" name="stage" value="evidence"> Evidence of demand, building the first real version</label>
          <label><input type="radio" name="stage" value="trading"> Trading — real customers, some money coming in</label>
          <label><input type="radio" name="stage" value="growing"> Established and growing</label>
        </li>

        <li class="ff-q" data-q="need">
          <p class="ff-q-text">2. What is the money actually for?</p>
          <label><input type="radio" name="need" value="runway"> Living costs while I work it out</label>
          <label><input type="radio" name="need" value="build"> Building the first version</label>
          <label><input type="radio" name="need" value="rd"> A genuinely innovative research-and-development project</label>
          <label><input type="radio" name="need" value="working"> Stock, equipment or cashflow</label>
          <label><input type="radio" name="need" value="growth"> Marketing and growth</label>
        </li>

        <li class="ff-q" data-q="ambition">
          <p class="ff-q-text">3. How big are you honestly trying to make this?</p>
          <p class="ff-q-help">There is no right answer — but the funding routes differ completely.</p>
          <label><input type="radio" name="ambition" value="lifestyle"> A good living for me (and maybe a few others)</label>
          <label><input type="radio" name="ambition" value="solid"> A solid company with a team</label>
          <label><input type="radio" name="ambition" value="large"> Very large — this only works at scale</label>
        </li>

        <li class="ff-q" data-q="equity">
          <p class="ff-q-text">4. Would you sell a share of the business?</p>
          <label><input type="radio" name="equity" value="yes"> Yes, for the right backer</label>
          <label><input type="radio" name="equity" value="reluctant"> Prefer not to</label>
          <label><input type="radio" name="equity" value="never"> Never</label>
        </li>

        <li class="ff-q" data-q="repay">
          <p class="ff-q-text">5. Could you take on repayments — probably personally guaranteed?</p>
          <p class="ff-q-help">Early business lending usually means the risk lands on you, not the company.</p>
          <label><input type="radio" name="repay" value="yes"> Yes, against money I can see coming in</label>
          <label><input type="radio" name="repay" value="no"> No — repayments would sink me if it went slowly</label>
        </li>

        <li class="ff-q" data-q="rd">
          <p class="ff-q-text">6. Is the work genuinely new and technically uncertain?</p>
          <p class="ff-q-help">Not "new to me" — something whose outcome a competent expert could not predict. This is the bar innovation grants and R&amp;D relief actually apply.</p>
          <label><input type="radio" name="rd" value="yes"> Yes — nobody is sure this can be done</label>
          <label><input type="radio" name="rd" value="partial"> Partly — some real technical unknowns</label>
          <label><input type="radio" name="rd" value="no"> No — it is new packaging of things that exist</label>
        </li>

        <li class="ff-q" data-q="flags">
          <p class="ff-q-text">7. Tick any that apply to you</p>
          <label><input type="checkbox" name="flag_young" value="1"> I am 18–30</label>
          <label><input type="checkbox" name="flag_social" value="1"> The core purpose is social or environmental impact</label>
          <label><input type="checkbox" name="flag_devolved" value="1"> Based in Scotland, Wales or Northern Ireland</label>
        </li>

      </ol>

      <p class="ff-error" id="ffError" hidden>Please answer questions 1–6 — each one changes the result.</p>
      <button type="submit" class="ff-submit">{{.submit_label}}</button>
    </form>

    <div class="ff-result" id="ffResult" hidden aria-live="polite">
      <span class="ff-verdict-tag" id="ffTag"></span>
      <h3 class="ff-verdict" id="ffVerdict"></h3>
      <p class="ff-verdict-lead" id="ffLead"></p>
      <ul class="ff-points" id="ffPoints"></ul>
      <div class="ff-actions">
        <a class="ff-btn ff-btn-primary" href="{{.result_cta_url}}">{{.result_cta_label}}</a>
        <a class="ff-btn ff-btn-secondary" href="{{.guide_url}}">{{.guide_cta_label}}</a>
        <button type="button" class="ff-btn ff-btn-plain" id="ffRetake">Start again</button>
      </div>
      <p class="ff-disclaimer">{{.disclaimer}}</p>
    </div>
  </div>

<style>
.funding-fit { padding: var(--spacing-section, 5rem 2rem); background: var(--color-background); color: var(--color-text); }
.funding-fit .ff-inner { max-width: 780px; margin: 0 auto; }
.funding-fit .ff-eyebrow { display:inline-block; font-size:.8rem; font-weight:700; letter-spacing:.1em; text-transform:uppercase; color: var(--color-primary); margin-bottom:.75rem; }
.funding-fit .ff-heading { font-size: clamp(1.6rem, 3.2vw, 2.4rem); font-weight:800; color: var(--color-heading); margin:0 0 1rem; line-height:1.2; }
.funding-fit .ff-intro { font-size:1.05rem; color: var(--color-text-muted); line-height:1.7; margin:0 0 2.5rem; }
.funding-fit .ff-questions { list-style:none; margin:0; padding:0; }
.funding-fit .ff-q { border:1px solid var(--color-border); border-radius: var(--border-radius, .5rem); padding:1.5rem; margin-bottom:1.25rem; background: var(--color-card-bg, var(--color-surface)); }
.funding-fit .ff-q-text { font-weight:700; color: var(--color-heading); margin:0 0 .4rem; }
.funding-fit .ff-q-help { font-size:.9rem; color: var(--color-text-muted); margin:0 0 1rem; line-height:1.5; }
.funding-fit .ff-q label { display:flex; gap:.65rem; align-items:flex-start; padding:.55rem .7rem; border-radius:.35rem; cursor:pointer; line-height:1.45; }
.funding-fit .ff-q label:hover { background: color-mix(in srgb, var(--color-primary) 8%, transparent); }
.funding-fit .ff-q input { margin-top:.3rem; flex:none; }
.funding-fit .ff-error { color:#b3261e; font-weight:600; margin:0 0 1rem; }
.funding-fit .ff-submit, .funding-fit .ff-btn { display:inline-block; padding:.9rem 1.75rem; border-radius:.4rem; font-weight:600; font-size:1rem; text-decoration:none; cursor:pointer; border:2px solid var(--color-primary); }
.funding-fit .ff-submit, .funding-fit .ff-btn-primary { background: var(--color-primary); color: var(--color-primary-text, #fff); }
.funding-fit .ff-btn-secondary { background:transparent; color: var(--color-primary); }
.funding-fit .ff-btn-plain { background:transparent; border-color:transparent; color: var(--color-text-muted); text-decoration:underline; }
.funding-fit .ff-result { border:2px solid var(--color-primary); border-radius: var(--border-radius, .5rem); padding:2rem; background: var(--color-card-bg, var(--color-surface)); }
.funding-fit .ff-verdict-tag { display:inline-block; font-size:.75rem; font-weight:700; letter-spacing:.08em; text-transform:uppercase; padding:.3rem .6rem; border-radius:.25rem; margin-bottom:.9rem; background: color-mix(in srgb, var(--color-primary) 15%, transparent); color: var(--color-primary); }
.funding-fit .ff-verdict { font-size:1.5rem; font-weight:800; color: var(--color-heading); margin:0 0 .75rem; line-height:1.25; }
.funding-fit .ff-verdict-lead { font-size:1.05rem; line-height:1.7; margin:0 0 1.25rem; }
.funding-fit .ff-points { margin:0 0 1.75rem; padding-left:1.2rem; line-height:1.65; }
.funding-fit .ff-points li { margin-bottom:.6rem; }
.funding-fit .ff-actions { display:flex; gap:.75rem; flex-wrap:wrap; align-items:center; margin-bottom:1.5rem; }
.funding-fit .ff-disclaimer { font-size:.85rem; color: var(--color-text-muted); line-height:1.6; margin:0; border-top:1px solid var(--color-border); padding-top:1rem; }
@media (max-width:768px){ .funding-fit .ff-actions .ff-btn { width:100%; text-align:center; } }
</style>

<script>
(function () {
  var form = document.getElementById('ffForm');
  if (!form) return;
  var result = document.getElementById('ffResult');
  var errEl  = document.getElementById('ffError');
  var RADIOS = ['stage','need','ambition','equity','repay','rd'];

  function answers() {
    var a = {};
    for (var i = 0; i < RADIOS.length; i++) {
      var el = form.querySelector('input[name="' + RADIOS[i] + '"]:checked');
      if (!el) return null;
      a[RADIOS[i]] = el.value;
    }
    a.young    = !!form.querySelector('input[name="flag_young"]:checked');
    a.social   = !!form.querySelector('input[name="flag_social"]:checked');
    a.devolved = !!form.querySelector('input[name="flag_devolved"]:checked');
    return a;
  }

  // Signposts appended to every verdict — durable institutions only, no amounts,
  // no scheme details: those change every fiscal year, and the guides carry the map.
  function signposts(a, pts) {
    if (a.young)    pts.push('You are 18–30: the King’s Trust (formerly the Prince’s Trust) exists for exactly this — funding, mentoring and courses for young people starting businesses. Free to approach, and worth doing early.');
    if (a.social)   pts.push('Your core purpose is social or environmental: a separate funding world exists — social investment funds, foundation grants, community shares — that accepts returns the commercial market would not. Start from “social investment UK” and the infrastructure bodies, not from commercial investor lists.');
    if (a.devolved) pts.push('Scotland, Wales and Northern Ireland each run their own development agency (Scottish Enterprise, the Development Bank of Wales, Invest NI) with grants, loans and equity that are often more accessible than the UK-wide schemes. Check yours before anything national.');
    else            pts.push('Your local Growth Hub is the free front door in England — it knows what is currently open in your area and costs nothing to ask.');
    pts.push('Be wary of anyone charging upfront to “find you grants” or “introduce you to investors” — every front door above is free.');
  }

  // GATED, not scored — same rule as the patent check. Two answers are close to
  // dispositive on their own and are checked FIRST: needing the money for living
  // costs (almost nothing funds runway, and pretending otherwise wastes months),
  // and having no evidence yet (every route asks the same demand question).
  // Only past both gates do the remaining answers COMPOSE a route map.
  function verdict(a) {
    if (a.need === 'runway') {
      var pts = [
        'Grants fund projects with deliverables, not salaries while you think. Lenders want visible repayment. Investors fund growth, not runway. None of them is designed for this, and applications to them from here mostly cost you months.',
        'The honest routes are unglamorous: keep or take paid work and build in the margins; cut the idea’s scope until it can be started with what you have; or sell a rough version of the service by hand so the customers fund it.',
        'If the work itself is genuinely innovative, project-based grant funding may cover the PROJECT (not your living costs) — see the funding-sources guide for the map.'
      ];
      signposts(a, pts);
      return { tag: 'Honest answer', title: 'Almost nothing funds living costs at this stage — knowing that now saves you months', lead: 'This is the most common funding question and the one with the least comfortable answer. The funding world funds projects, repayable lending and scalable growth — not runway while an idea takes shape.', points: pts };
    }

    if (a.stage === 'idea') {
      var pts2 = [
        'Grant assessors, lenders and investors all ask the same first question: where is the evidence anyone wants this? Without it, applications fail slowly and politely, which is worse than quickly.',
        'The cheapest capital available to you is evidence: a test with a pass mark, a pre-order page, deposits, a hand-delivered version for the first few customers. The testing guide covers how.',
        a.need === 'build' ? 'For building the first version: read the building guide before spending — the first version is usually buildable without funding at all (by hand, no-code, one narrow path), and what it teaches you is what unlocks the money.' : '',
        'If you have an existing audience, reward crowdfunding is the one route that works AT this stage — it is a demand test that pays you. Mind the patents warning first if the idea might be patentable.'
      ].filter(Boolean);
      signposts(a, pts2);
      return { tag: 'Evidence first', title: 'You are one step early — buy evidence before you ask anyone for money', lead: 'Nothing is wrong with the idea stage; it is just not the funding stage. Every route on the map opens noticeably easier with evidence of demand in your hand.', points: pts2 };
    }

    // Past the gates: compose the route map.
    var routes = [];
    if (a.rd === 'yes' || a.rd === 'partial') {
      routes.push('Innovation grants fit the R&D half of what you described: Innovate UK nationally' + (a.devolved ? ', plus your national development agency' : '') + '. Expect competition, real application effort, match funding, and payment against evidenced project spend — grants fund the project, not the runway around it.');
      routes.push('The same genuinely-uncertain work may qualify for R&D tax relief — accountant territory (use one who handles claims routinely, not a cold-calling “specialist”), but worth knowing before you decide what you can afford to attempt.');
    }
    if (a.equity === 'yes' && a.ambition === 'large') {
      routes.push('Your ambition and willingness to sell a share point at equity: angels first (find them through networks and warm introductions via founders they have backed — the UK Business Angels Association keeps a directory), and read up on HMRC advance assurance before the first meeting, because serious UK angels will ask. Venture capital only if it genuinely must be huge to work — taking VC money commits you to that.');
    } else if (a.equity === 'yes' && a.ambition !== 'lifestyle') {
      routes.push('Equity is open to you but not compulsory: angels via networks and warm introductions where the growth story justifies it, or equity crowdfunding if your customers would want to be owners. Read up on HMRC advance assurance first.');
    }
    if (a.equity !== 'yes' && a.ambition === 'large') {
      routes.push('One tension to resolve honestly: very-large ambitions are usually equity-funded, and you said no to selling a share. One of the two normally has to give — better to decide which now than mid-negotiation.');
    }
    if (a.repay === 'yes' && (a.stage === 'trading' || a.stage === 'growing')) {
      routes.push((a.need === 'working' ? 'For stock, equipment and cashflow specifically, debt is the natural shape: ' : 'With money visibly coming in, debt keeps your equity: ') + 'banks first (ask about government guarantee-backed schemes by name if they hesitate), the Start Up Loans programme if you are early, and invoice or asset finance where the need is bridging what customers owe you or buying equipment. Read the personal-guarantee clause twice.');
    }
    if (a.repay === 'no') {
      routes.push('You said repayments would sink you if things went slowly — believe yourself, and treat debt as closed for now. That mainly leaves grants (project-shaped), equity (growth-shaped) and customer money.');
    }
    if (a.equity === 'never' && a.repay === 'no' && a.rd === 'no') {
      routes.push('With equity, debt and grants all effectively closed by your answers, the funding source left is the best one anyway: customers. Price properly, invoice deposits, and let revenue fund the growth — slower, and entirely yours.');
    }
    if (a.stage === 'evidence' && a.need === 'build') {
      routes.push('At your stage the amounts needed are usually smaller than they look: the building guide’s ladder (by hand → no-code → real build) cuts the bill before you fund it.');
    }
    signposts(a, routes);
    var tag = (a.rd !== 'no') ? 'Grant-led mix' : (a.equity === 'yes' && a.ambition !== 'lifestyle') ? 'Equity-led mix' : (a.repay === 'yes') ? 'Debt-led mix' : 'Customer-funded';
    return { tag: tag, title: 'Your route map, in order of fit', lead: 'Past the two gates — you have evidence, and the money is for the business rather than runway — your answers compose a route map rather than a single answer. In rough order of fit:', points: routes };
  }

  form.addEventListener('submit', function (e) {
    e.preventDefault();
    var a = answers();
    if (!a) { errEl.hidden = false; return; }
    errEl.hidden = true;
    var v = verdict(a);
    document.getElementById('ffTag').textContent = v.tag;
    document.getElementById('ffVerdict').textContent = v.title;
    document.getElementById('ffLead').textContent = v.lead;
    var ul = document.getElementById('ffPoints');
    ul.innerHTML = '';
    for (var i = 0; i < v.points.length; i++) {
      if (!v.points[i]) continue;
      var li = document.createElement('li');
      li.textContent = v.points[i];
      ul.appendChild(li);
    }
    form.hidden = true;
    result.hidden = false;
    result.scrollIntoView({ behavior: 'smooth', block: 'start' });
  });

  document.getElementById('ffRetake').addEventListener('click', function () {
    form.reset();
    result.hidden = true;
    form.hidden = false;
    form.scrollIntoView({ behavior: 'smooth', block: 'start' });
  });
})();
</script>
</section>
$tpl$,
  jsonb_build_object('fields', jsonb_build_object(
    'eyebrow_label',   jsonb_build_object('type','text','source','static','required',false),
    'section_heading', jsonb_build_object('type','text','source','static','required',false),
    'section_intro',   jsonb_build_object('type','text','source','static','required',false),
    'submit_label',    jsonb_build_object('type','text','source','static','required',false),
    'result_cta_url',  jsonb_build_object('type','url', 'source','static','required',false),
    'result_cta_label',jsonb_build_object('type','text','source','static','required',false),
    'guide_url',       jsonb_build_object('type','url', 'source','static','required',false),
    'guide_cta_label', jsonb_build_object('type','text','source','static','required',false),
    'disclaimer',      jsonb_build_object('type','text','source','static','required',false)
  ))
);

INSERT INTO pages
  (site_id, name, url, title, page_type, status, meta_description, topics,
   nav_label, nav_order, in_header, in_footer, build_status, sections)
VALUES
  ('1244516d-014d-421c-88c6-090bb1e9552a',
   'tool-funding-fit',
   '/tools/funding-fit/index.html',
   'Which funding route fits? A free finder',
   'tool',
   'active',
   'Seven questions, no sign-up, nothing stored. An honest steer on which funding routes fit your stage — including the answer nobody advertises: sometimes none of them do yet.',
   ARRAY['funding','grants','investment','free tool','uk'],
   'Which funding fits?',
   20,
   false,
   false,
   'planned',
   '[]'::jsonb);

INSERT INTO page_components (page_id, component_id, slot_name, position, content_data, build_status)
SELECT p.id, '23f95f00-f293-466e-b43a-81791ea0fc6c', 'hero', 1,
       jsonb_build_object(
         'headline',          'Which funding route actually fits?',
         'subheadline',       'Seven questions, about two minutes. No sign-up, nothing stored — it runs in your browser. It will tell you which of the UK routes fit where you actually are, and it will tell you honestly when the answer is "none of them yet", which is the answer that saves the most months.',
         'cta_text',          'Read the funding guides first',
         'cta_url',           '/guides/funding-ways/index.html',
         'secondary_cta',     'Get a verified idea report',
         'secondary_cta_url', '/report.html'
       ),
       'pending'
FROM pages p
WHERE p.site_id = '1244516d-014d-421c-88c6-090bb1e9552a'
  AND p.url = '/tools/funding-fit/index.html';

INSERT INTO page_components (page_id, component_id, slot_name, position, content_data, build_status)
SELECT p.id, c.id, 'funding-fit', 2,
       jsonb_build_object(
         'eyebrow_label',    'Free finder',
         'section_heading',  'Answer honestly and we will point you at the right doors — or tell you to wait',
         'section_intro',    'The funding world funds projects, repayable lending and scalable growth — and quietly refuses everything else. This asks what a good adviser would ask first, then maps your answers onto the routes and institutions from our funding guides. It is a steer, not advice — but it will stop you spending months applying to doors that were never going to open from where you stand.',
         'submit_label',     'Show me the routes',
         'result_cta_url',   '/report.html',
         'result_cta_label', 'Get a verified idea report',
         'guide_url',        '/guides/funding-sources/index.html',
         'guide_cta_label',  'Read the sources guide',
         'disclaimer',       'This is general information about the UK funding landscape, not financial advice, and it names no amounts or scheme details because they change frequently. Check the institutions'' own websites or gov.uk for the current position, and take regulated advice where a decision is significant. Nothing here is an endorsement of any provider.'
       ),
       'pending'
FROM pages p, content_components c
WHERE p.site_id = '1244516d-014d-421c-88c6-090bb1e9552a'
  AND p.url = '/tools/funding-fit/index.html'
  AND c.name = 'funding-fit';

UPDATE pages p
SET sections = (
      SELECT COALESCE(jsonb_agg(pc.slot_name ORDER BY pc.position), '[]'::jsonb)
      FROM page_components pc WHERE pc.page_id = p.id AND COALESCE(pc.slot_name,'') <> ''
    ),
    updated_at = now()
WHERE p.site_id = '1244516d-014d-421c-88c6-090bb1e9552a'
  AND p.url = '/tools/funding-fit/index.html';

DO $guard2$
DECLARE ntot int; nslot int; nbad int;
BEGIN
  SELECT count(*), count(*) FILTER (WHERE COALESCE(pc.slot_name,'') = ''),
         count(*) FILTER (WHERE pc.content_data IS NULL OR pc.content_data = '{}'::jsonb)
    INTO ntot, nslot, nbad
  FROM page_components pc JOIN pages p ON p.id = pc.page_id
  WHERE p.site_id = '1244516d-014d-421c-88c6-090bb1e9552a'
    AND p.url = '/tools/funding-fit/index.html';
  IF ntot <> 2 THEN RAISE EXCEPTION 'ABORT: expected 2 sections, got %.', ntot; END IF;
  IF nslot > 0 THEN RAISE EXCEPTION 'ABORT: % section(s) missing slot_name.', nslot; END IF;
  IF nbad > 0 THEN RAISE EXCEPTION 'ABORT: % section(s) empty content_data.', nbad; END IF;
END
$guard2$;

COMMIT;

SELECT p.id AS page_id, p.url, p.page_type, p.build_status, p.sections FROM pages p
WHERE p.site_id = '1244516d-014d-421c-88c6-090bb1e9552a' AND p.url = '/tools/funding-fit/index.html';

SELECT id, name, function, length(html_template) AS tpl_bytes,
       (html_template LIKE '%</section>%') AS has_close_tag
FROM content_components WHERE name = 'funding-fit';
