-- p4_06_tool_patent_check.sql — idea.uk's FIRST free (Tier-1) tool: "Should you patent it?"
--
-- Owner 2026-07-25: "yes for copyright and the checker". This is the checker — the first Tier-1
-- free probe from features_open/013's three-tier funnel, and the first free tool idea.uk has built
-- (the existing /tools.html#audience-check is a pointer at the paid tool's own backend, not a
-- self-contained free probe).
--
-- WHY A NEW COMPONENT RATHER THAN REUSING `ai-readiness-quiz` (71a636a7, 2 live instances).
-- Reuse was the first thing checked, and it is the wrong instrument here. That component is a
-- fixed 5-question / 4-option **sum-score** quiz: every answer contributes points and the total
-- picks a result tier. Patentability does not work that way. "Have you already disclosed it
-- publicly?" is close to dispositive on its own — in the UK there is no general grace period, so a
-- public disclosure normally destroys novelty regardless of how good the idea is. Under a sum
-- score, someone who has already published would score well on the other five questions and be
-- told they look patent-ready. That is not a cosmetic mismatch; it is advice that could cost
-- somebody their rights. Same for subject-matter exclusions: a business method is excluded however
-- strong the commercial case.
--
-- So this component is GATED, not scored: disclosure and subject-matter are checked first and
-- short-circuit to their own outcomes; only if both pass does the commercial question (prior art,
-- detectability, ability to enforce, shelf life) get scored into three bands. The gating logic is
-- commented in the template's own script so the next reader does not "simplify" it back into a sum.
--
-- (Reusing the quiz would ALSO have required editing a shared component again: its
-- `quiz_badge_label` is `source: static` WITH a fallback, which p4_04 established is unoverridable
-- — the badge would have read "AI Readiness Assessment" on a patent checker.)
--
-- ZERO COST, ZERO BACKEND. Entirely client-side: no LLM call, no API, no data leaves the browser,
-- nothing stored. That is what makes it a Tier-1 probe rather than a paid operation — it captures
-- intent and funnels to /report.html without spending anything per use.
--
-- JS DELIVERY — inline <script> INSIDE html_template, deliberately.
-- The alternative (an external /tools/assets/*.js referenced by the template, as `ai-readiness-quiz`
-- does) depends on the asset-publishing path that produced bugs_open/041 (chrome JS published but
-- never loaded) and the js_content-vs-js_snippets trap. An inline script is part of the rendered
-- section HTML and cannot be published-but-not-loaded. Verified locally before insert: the template
-- parses and executes under text/template (0 `<no value>`, 0 unreplaced actions, `</section>`
-- present so it passes componentTemplateValid).
--
-- SCHEMA: every field is `source: static` with **NO fallback**, which after p4_04 is the shape that
-- lets content_data win (plan_sections_action.go:1556-1562 writes a fallback unconditionally and
-- resolved_data merges LAST, so a fallback here would make the field unoverridable). `items`-style
-- query sources are deliberately absent — this component derives nothing.
--
-- NGINX CHECK DONE BEFORE CHOOSING THE URL: idea.uk's box reserves tool routes by EXACT match
-- (`location = /request` etc.), with only /stripe/, /internal/ and /order/ as prefixes. `/tools/`
-- is served statically — probed live: /tools/assets/site-header.js 200, /tools/anything-random/ 404
-- (static miss, not the Go binary). So /tools/patent-check/index.html is safe to serve.
--
-- page_type='tool' means /tools.html lists it automatically: that page's `tool-list` sources items
-- from `query.pages_where_type:tool`, the same self-listing mechanism p4_02 gave the guides hub.
-- Ordering constraint is therefore the same: this page must SHIP before /tools.html is re-rendered.

\set ON_ERROR_STOP on

BEGIN;

DO $guard$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM pages
   WHERE site_id = '1244516d-014d-421c-88c6-090bb1e9552a'
     AND url = '/tools/patent-check/index.html';
  IF n > 0 THEN
    RAISE EXCEPTION 'ABORT: /tools/patent-check/index.html already exists — p4_06 already ran.';
  END IF;
  SELECT count(*) INTO n FROM content_components WHERE name = 'patent-check';
  IF n > 0 THEN
    RAISE EXCEPTION 'ABORT: a component named patent-check already exists.';
  END IF;
END
$guard$;

-- ---------------------------------------------------------------------------
-- 1. The component.
-- ---------------------------------------------------------------------------
INSERT INTO content_components
  (name, display_name, description, function, section_type, component_level, render_mode,
   category, is_active, created_from, suitable_site_types, suitable_page_types, html_template, input_schema)
VALUES (
  'patent-check',
  'Should you patent it? — free check',
  'Client-side, zero-cost patentability steer. GATED, not scored: public disclosure and excluded '
  || 'subject matter short-circuit to their own outcomes before any scoring happens, because either '
  || 'one is decisive on its own and a sum score would let a good total drown it out. No backend, no '
  || 'LLM, no data leaves the browser. Script is inline in the template on purpose (see p4_06).',
  'patent-check',
  'patent-check',
  'section',
  'template',
  'tool',
  true,
  'manual',
  '[]'::jsonb,
  '[]'::jsonb,
  $tpl$<section class="patent-check" data-component="patent-check" id="patent-check">
  <div class="pc-inner">
    <span class="pc-eyebrow">{{.eyebrow_label}}</span>
    <h2 class="pc-heading">{{.section_heading}}</h2>
    <p class="pc-intro">{{.section_intro}}</p>

    <form class="pc-form" id="pcForm" novalidate>
      <ol class="pc-questions">

        <li class="pc-q" data-q="disclosure">
          <p class="pc-q-text">1. Has it been shown or described outside a confidentiality agreement?</p>
          <p class="pc-q-help">A demo, a pitch without an NDA, a crowdfunding page, a talk, a paper, a listing, a video, a forum post explaining how it works.</p>
          <label><input type="radio" name="disclosure" value="never"> No — it has been kept confidential</label>
          <label><input type="radio" name="disclosure" value="nda"> Only under signed NDAs</label>
          <label><input type="radio" name="disclosure" value="recent"> Yes, publicly — within the last 6 months</label>
          <label><input type="radio" name="disclosure" value="old"> Yes, publicly — more than 6 months ago</label>
          <label><input type="radio" name="disclosure" value="unsure"> I am not sure</label>
        </li>

        <li class="pc-q" data-q="kind">
          <p class="pc-q-text">2. What is it, most accurately?</p>
          <p class="pc-q-help">Pick the closest. If two fit, pick the one a stranger would say describes it.</p>
          <label><input type="radio" name="kind" value="physical"> A physical product, device, material, or a way of making something</label>
          <label><input type="radio" name="kind" value="technical_sw"> Software that controls hardware, processes a signal, or changes how the computer itself works</label>
          <label><input type="radio" name="kind" value="plain_sw"> An app, website or software feature — no hardware, no technical effect beyond running normally</label>
          <label><input type="radio" name="kind" value="business"> A business model, pricing scheme, marketplace design, or a way of organising a service</label>
          <label><input type="radio" name="kind" value="creative"> A creative work — writing, art, music, game content, a visual design</label>
        </li>

        <li class="pc-q" data-q="priorart">
          <p class="pc-q-text">3. Have you checked whether it already exists?</p>
          <p class="pc-q-help">Patents, yes — but also products on sale, manuals, papers, videos and posts. All of it counts against novelty.</p>
          <label><input type="radio" name="priorart" value="searched_clear"> Yes — searched properly, found nothing close</label>
          <label><input type="radio" name="priorart" value="searched_similar"> Searched, and found things that are similar</label>
          <label><input type="radio" name="priorart" value="none"> Not yet</label>
        </li>

        <li class="pc-q" data-q="detect">
          <p class="pc-q-text">4. If a competitor copied it, could you tell from their product?</p>
          <p class="pc-q-help">A patent you cannot detect an infringement of is very hard to use.</p>
          <label><input type="radio" name="detect" value="yes"> Yes — it would be visible or testable from outside</label>
          <label><input type="radio" name="detect" value="maybe"> Possibly, with some effort</label>
          <label><input type="radio" name="detect" value="no"> No — it would be hidden in their process or on their servers</label>
        </li>

        <li class="pc-q" data-q="enforce">
          <p class="pc-q-text">5. Could you fund a legal fight over it?</p>
          <p class="pc-q-help">A patent is a right to sue at your own expense. Nobody enforces it for you.</p>
          <label><input type="radio" name="enforce" value="yes"> Yes, or an investor or licensee would back it</label>
          <label><input type="radio" name="enforce" value="maybe"> Maybe, against a serious infringer</label>
          <label><input type="radio" name="enforce" value="no"> No</label>
        </li>

        <li class="pc-q" data-q="life">
          <p class="pc-q-text">6. How long before it is superseded?</p>
          <p class="pc-q-help">A UK patent commonly takes several years to grant.</p>
          <label><input type="radio" name="life" value="long"> Five years or more</label>
          <label><input type="radio" name="life" value="mid"> Two to five years</label>
          <label><input type="radio" name="life" value="short"> Under two years</label>
        </li>

      </ol>

      <p class="pc-error" id="pcError" hidden>Please answer every question — each one changes the result.</p>
      <button type="submit" class="pc-submit">{{.submit_label}}</button>
    </form>

    <div class="pc-result" id="pcResult" hidden aria-live="polite">
      <span class="pc-verdict-tag" id="pcTag"></span>
      <h3 class="pc-verdict" id="pcVerdict"></h3>
      <p class="pc-verdict-lead" id="pcLead"></p>
      <ul class="pc-points" id="pcPoints"></ul>
      <div class="pc-actions">
        <a class="pc-btn pc-btn-primary" id="pcPrimary" href="{{.result_cta_url}}">{{.result_cta_label}}</a>
        <a class="pc-btn pc-btn-secondary" href="{{.guide_url}}">{{.guide_cta_label}}</a>
        <button type="button" class="pc-btn pc-btn-plain" id="pcRetake">Start again</button>
      </div>
      <p class="pc-disclaimer">{{.disclaimer}}</p>
    </div>
  </div>

<style>
.patent-check { padding: var(--spacing-section, 5rem 2rem); background: var(--color-background); color: var(--color-text); }
.patent-check .pc-inner { max-width: 780px; margin: 0 auto; }
.patent-check .pc-eyebrow { display:inline-block; font-size:.8rem; font-weight:700; letter-spacing:.1em; text-transform:uppercase; color: var(--color-primary); margin-bottom:.75rem; }
.patent-check .pc-heading { font-size: clamp(1.6rem, 3.2vw, 2.4rem); font-weight:800; color: var(--color-heading); margin:0 0 1rem; line-height:1.2; }
.patent-check .pc-intro { font-size:1.05rem; color: var(--color-text-muted); line-height:1.7; margin:0 0 2.5rem; }
.patent-check .pc-questions { list-style:none; margin:0; padding:0; }
.patent-check .pc-q { border:1px solid var(--color-border); border-radius: var(--border-radius, .5rem); padding:1.5rem; margin-bottom:1.25rem; background: var(--color-card-bg, var(--color-surface)); }
.patent-check .pc-q-text { font-weight:700; color: var(--color-heading); margin:0 0 .4rem; }
.patent-check .pc-q-help { font-size:.9rem; color: var(--color-text-muted); margin:0 0 1rem; line-height:1.5; }
.patent-check .pc-q label { display:flex; gap:.65rem; align-items:flex-start; padding:.55rem .7rem; border-radius:.35rem; cursor:pointer; line-height:1.45; }
.patent-check .pc-q label:hover { background: color-mix(in srgb, var(--color-primary) 8%, transparent); }
.patent-check .pc-q input { margin-top:.3rem; flex:none; }
.patent-check .pc-error { color:#b3261e; font-weight:600; margin:0 0 1rem; }
.patent-check .pc-submit, .patent-check .pc-btn { display:inline-block; padding:.9rem 1.75rem; border-radius:.4rem; font-weight:600; font-size:1rem; text-decoration:none; cursor:pointer; border:2px solid var(--color-primary); }
.patent-check .pc-submit, .patent-check .pc-btn-primary { background: var(--color-primary); color: var(--color-primary-text, #fff); }
.patent-check .pc-btn-secondary { background:transparent; color: var(--color-primary); }
.patent-check .pc-btn-plain { background:transparent; border-color:transparent; color: var(--color-text-muted); text-decoration:underline; }
.patent-check .pc-result { border:2px solid var(--color-primary); border-radius: var(--border-radius, .5rem); padding:2rem; background: var(--color-card-bg, var(--color-surface)); }
.patent-check .pc-verdict-tag { display:inline-block; font-size:.75rem; font-weight:700; letter-spacing:.08em; text-transform:uppercase; padding:.3rem .6rem; border-radius:.25rem; margin-bottom:.9rem; background: color-mix(in srgb, var(--color-primary) 15%, transparent); color: var(--color-primary); }
.patent-check .pc-verdict { font-size:1.5rem; font-weight:800; color: var(--color-heading); margin:0 0 .75rem; line-height:1.25; }
.patent-check .pc-verdict-lead { font-size:1.05rem; line-height:1.7; margin:0 0 1.25rem; }
.patent-check .pc-points { margin:0 0 1.75rem; padding-left:1.2rem; line-height:1.65; }
.patent-check .pc-points li { margin-bottom:.6rem; }
.patent-check .pc-actions { display:flex; gap:.75rem; flex-wrap:wrap; align-items:center; margin-bottom:1.5rem; }
.patent-check .pc-disclaimer { font-size:.85rem; color: var(--color-text-muted); line-height:1.6; margin:0; border-top:1px solid var(--color-border); padding-top:1rem; }
@media (max-width:768px){ .patent-check .pc-actions .pc-btn { width:100%; text-align:center; } }
</style>

<script>
(function () {
  var form = document.getElementById('pcForm');
  if (!form) return;
  var result = document.getElementById('pcResult');
  var errEl  = document.getElementById('pcError');
  var KEYS = ['disclosure','kind','priorart','detect','enforce','life'];

  function answers() {
    var a = {};
    for (var i = 0; i < KEYS.length; i++) {
      var el = form.querySelector('input[name="' + KEYS[i] + '"]:checked');
      if (!el) return null;
      a[KEYS[i]] = el.value;
    }
    return a;
  }

  // Gated, not scored. Some answers decide the outcome on their own — a sum would
  // let a strong score elsewhere drown out "already published", which is fatal.
  function verdict(a) {
    if (a.disclosure === 'recent' || a.disclosure === 'old' || a.disclosure === 'unsure') {
      var urgent = (a.disclosure === 'recent' || a.disclosure === 'unsure');
      return {
        tag: 'Time-critical',
        title: urgent ? 'Speak to a patent attorney this week — before you do anything else'
                      : 'Public disclosure has probably already ended the patent option',
        lead: 'In the UK and Europe an invention must be new at the moment you file, and there is no general grace period. A public disclosure — including your own — normally counts against you.',
        points: [
          urgent ? 'A small number of narrow exceptions exist (broadly, disclosure made in breach of confidence, or certain officially recognised international exhibitions, each within six months). Only a patent attorney can tell you whether one applies to you, and the clock is running.'
                 : 'The six-month window in which the narrow exceptions could have helped has most likely passed. It is still worth one conversation to be certain, because what exactly was disclosed matters.',
          'Do not disclose anything further in the meantime, and write down exactly what was shown, to whom, and when — that record is the first thing an attorney will ask for.',
          'If patenting really is closed off, you have not lost everything: copyright in your code and materials is automatic, and speed, brand and confidentiality about what has not yet been shown all still work.'
        ]
      };
    }

    if (a.kind === 'creative') {
      return {
        tag: 'Wrong tool',
        title: 'A patent is not the right instrument for this',
        lead: 'Patents cover technical inventions. Creative works — writing, art, music, game content, visual design — are excluded from patent protection, and they are already protected by something else.',
        points: [
          'Copyright is automatic, free, and lasts the author’s life plus seventy years. You already have it.',
          'If the value is in how the product looks, a registered design is inexpensive next to a patent.',
          'If the value is in the name, that is a trade mark.',
          'The thing most worth checking today: if anyone other than an employee made it for you, they own the copyright unless you have a signed written assignment.'
        ]
      };
    }

    if (a.kind === 'business') {
      return {
        tag: 'Likely excluded',
        title: 'This is very likely excluded from patent protection',
        lead: 'UK law excludes schemes, rules and methods for doing business, and presentations of information, from being patentable inventions "as such". Marketplace designs, pricing schemes and ways of organising a service normally fall on the wrong side of that line.',
        points: [
          'The exception is where there is a genuine technical effect underneath the commercial idea. If you believe there is one, that is a specific question worth putting to a patent attorney rather than assuming either way.',
          'For most business-model advantages the real protection is execution: getting there first, brand, distribution, data, and keeping the crucial part confidential.',
          'Confidentiality is free and available today. NDAs before disclosure, and keep the part that matters off the public side of the product.'
        ]
      };
    }

    if (a.kind === 'plain_sw') {
      return {
        tag: 'Probably excluded',
        title: 'Probably excluded — but there is a real question underneath',
        lead: 'Programs for a computer are excluded "as such". Software that produces a technical effect beyond a computer simply running — controlling a machine, processing a signal, or improving how the computer itself works — can still be patentable. Where the line falls is genuinely contested.',
        points: [
          'You answered that there is no hardware or technical effect involved, which puts this on the excluded side as described.',
          'If, thinking again, there is something technical underneath — performance, memory, a physical process, a signal — that changes the answer, and it is worth one paid hour with a patent attorney rather than a guess.',
          'Meanwhile: copyright in your source code is automatic and does stop someone copying it, though not someone writing their own.',
          a.detect === 'no' ? 'You also said an infringement would be invisible from outside. For server-side software that frequently makes a trade secret the stronger protection than a patent, which would publish the method at eighteen months.'
                            : 'Speed, brand and keeping the crucial part confidential are the usual protections in this category.'
        ]
      };
    }

    // Patentable subject matter in principle. Now the commercial question.
    var score = 0;
    if (a.priorart === 'searched_clear') score += 2; else if (a.priorart === 'searched_similar') score += 0; else score += 1;
    if (a.detect === 'yes') score += 2; else if (a.detect === 'maybe') score += 1;
    if (a.enforce === 'yes') score += 2; else if (a.enforce === 'maybe') score += 1;
    if (a.life === 'long') score += 2; else if (a.life === 'mid') score += 1;

    var pts = [];
    if (a.priorart === 'none') pts.push('Do a first-look prior-art search before you spend anything. Espacenet and Google Patents are free, an hour is usually enough for a first pass, and finding something close is a good outcome — it saves you the cost of finding out later.');
    if (a.priorart === 'searched_similar') pts.push('You found similar things. That is not necessarily fatal — patents turn on the specific claims, and "similar" is not "the same" — but it is the first thing to put in front of an attorney, because it shapes what can be claimed.');
    if (a.detect === 'no') pts.push('You could not tell if you were infringed. That is a serious argument for keeping it secret instead: a patent publishes your method at around eighteen months, and an unenforceable monopoly in exchange for a public recipe is a bad trade.');
    if (a.enforce === 'no') pts.push('You could not fund enforcement. A patent can still be worth having as an asset — for investors, a licensee or an acquirer — but be clear-eyed that as a weapon it would be a bluff.');
    if (a.life === 'short') pts.push('It is superseded within two years. A UK patent commonly takes several years to grant, so it may well be obsolete before you have the right, though the application itself still gives you a priority date and something to talk to investors about.');
    if (a.disclosure === 'nda') pts.push('Disclosure under signed NDAs is fine and does not count against novelty — keep it that way until you have filed.');

    if (score >= 6) {
      pts.unshift('File before you disclose anything publicly. The filing date is what the invention has to be new by, and everything else follows from it.');
      pts.push('Use the twelve-month priority year properly: once filed, you can talk, sell and raise money while deciding whether it is worth taking abroad.');
      return {
        tag: 'Worth pursuing',
        title: 'This looks worth a conversation with a patent attorney',
        lead: 'It is the right kind of subject matter, it has not been disclosed, and your answers on detectability, enforcement and shelf life all point the right way. That combination is not common.',
        points: pts
      };
    }
    if (score >= 3) {
      pts.unshift('Do not file yet, and do not disclose either — you still have the option, so keep it.');
      pts.push('A first consultation with a patent attorney is often free, and everything you tell them is confidential. One conversation will resolve most of what is above.');
      return {
        tag: 'Worth checking',
        title: 'Possibly — but do two things before you spend anything',
        lead: 'The subject matter is right and nothing has been disclosed, so the option is still open. Your answers flag some real doubts about whether a patent would earn its cost.',
        points: pts
      };
    }
    pts.unshift('Nothing here says the invention is bad — it says a patent is an expensive instrument that your answers suggest would not pay for itself.');
    pts.push('Keeping it confidential is a legitimate alternative with no expiry date, and it costs nothing. Its weakness is that it protects you not at all once it is out, or against someone who invents the same thing independently.');
    pts.push('Copyright in any code, drawings and materials is automatic and already yours.');
    return {
      tag: 'Probably not',
      title: 'Probably not worth patenting — here is what to do instead',
      lead: 'This is patentable subject matter in principle, but on detectability, enforcement, shelf life or prior art your answers point away from a patent being a good use of several thousand pounds.',
      points: pts
    };
  }

  form.addEventListener('submit', function (e) {
    e.preventDefault();
    var a = answers();
    if (!a) { errEl.hidden = false; return; }
    errEl.hidden = true;
    var v = verdict(a);
    document.getElementById('pcTag').textContent = v.tag;
    document.getElementById('pcVerdict').textContent = v.title;
    document.getElementById('pcLead').textContent = v.lead;
    var ul = document.getElementById('pcPoints');
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

  document.getElementById('pcRetake').addEventListener('click', function () {
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

-- ---------------------------------------------------------------------------
-- 2. The tool page.
-- ---------------------------------------------------------------------------
INSERT INTO pages
  (site_id, name, url, title, page_type, status, meta_description, topics,
   nav_label, nav_order, in_header, in_footer, build_status, sections)
VALUES
  ('1244516d-014d-421c-88c6-090bb1e9552a',
   'tool-patent-check',
   '/tools/patent-check/index.html',
   'Should you patent it? A free check',
   'tool',
   'active',
   'Six questions, no sign-up, nothing stored. A free steer on whether a UK patent is worth pursuing for your idea — or whether something cheaper protects it better.',
   ARRAY['patents','intellectual property','free tool','uk law'],
   'Should you patent it?',
   10,
   false,
   false,
   'planned',
   '[]'::jsonb);

-- Section 1 — hero.
INSERT INTO page_components (page_id, component_id, slot_name, position, content_data, build_status)
SELECT p.id, '23f95f00-f293-466e-b43a-81791ea0fc6c', 'hero', 1,
       jsonb_build_object(
         'headline',          'Should you patent it?',
         'subheadline',       'Six questions, about two minutes. No sign-up, nothing stored, nothing sent anywhere — it all runs in your browser. You will get an honest steer, including the fairly common answer that a patent is not the right tool and something cheaper protects you better.',
         'cta_text',          'Read the full patents guide',
         'cta_url',           '/guides/patents/index.html',
         'secondary_cta',     'Get a verified idea report',
         'secondary_cta_url', '/report.html'
       ),
       'pending'
FROM pages p
WHERE p.site_id = '1244516d-014d-421c-88c6-090bb1e9552a'
  AND p.url = '/tools/patent-check/index.html';

-- Section 2 — the checker.
INSERT INTO page_components (page_id, component_id, slot_name, position, content_data, build_status)
SELECT p.id, c.id, 'patent-check', 2,
       jsonb_build_object(
         'eyebrow_label',    'Free check',
         'section_heading',  'Answer six questions honestly and we will tell you if it is worth the money',
         'section_intro',    'Most ideas should not be patented, and knowing that early saves several thousand pounds. This asks the questions a patent attorney would ask in a first conversation. It is a steer, not an opinion on patentability — but it will tell you whether that conversation is worth booking, and it will tell you plainly if you have already done the one thing that closes the option off.',
         'submit_label',     'Show me the answer',
         'result_cta_url',   '/report.html',
         'result_cta_label', 'Get a verified idea report',
         'guide_url',        '/guides/patents/index.html',
         'guide_cta_label',  'Read the full guide',
         'disclaimer',       'This is general information about UK law and practice, not legal advice, and not an opinion on whether your invention is patentable. It cannot account for your circumstances, and patent deadlines are unforgiving. Speak to a registered patent attorney before acting on it — the Chartered Institute of Patent Attorneys publishes a directory, and a first conversation is often free and always confidential.'
       ),
       'pending'
FROM pages p, content_components c
WHERE p.site_id = '1244516d-014d-421c-88c6-090bb1e9552a'
  AND p.url = '/tools/patent-check/index.html'
  AND c.name = 'patent-check';

UPDATE pages p
SET sections = (
      SELECT COALESCE(jsonb_agg(pc.slot_name ORDER BY pc.position), '[]'::jsonb)
      FROM page_components pc WHERE pc.page_id = p.id AND COALESCE(pc.slot_name,'') <> ''
    ),
    updated_at = now()
WHERE p.site_id = '1244516d-014d-421c-88c6-090bb1e9552a'
  AND p.url = '/tools/patent-check/index.html';

DO $guard2$
DECLARE ntot int; nslot int; nbad int;
BEGIN
  SELECT count(*), count(*) FILTER (WHERE COALESCE(pc.slot_name,'') = ''),
         count(*) FILTER (WHERE pc.content_data IS NULL OR pc.content_data = '{}'::jsonb)
    INTO ntot, nslot, nbad
  FROM page_components pc JOIN pages p ON p.id = pc.page_id
  WHERE p.site_id = '1244516d-014d-421c-88c6-090bb1e9552a'
    AND p.url = '/tools/patent-check/index.html';
  IF ntot <> 2 THEN RAISE EXCEPTION 'ABORT: expected 2 sections, got %.', ntot; END IF;
  IF nslot > 0 THEN RAISE EXCEPTION 'ABORT: % section(s) have no slot_name.', nslot; END IF;
  IF nbad > 0 THEN RAISE EXCEPTION 'ABORT: % section(s) have empty content_data.', nbad; END IF;
END
$guard2$;

COMMIT;

SELECT p.id AS page_id, p.url, p.page_type, p.build_status, p.sections FROM pages p
WHERE p.site_id = '1244516d-014d-421c-88c6-090bb1e9552a' AND p.url = '/tools/patent-check/index.html';

SELECT id, name, function, section_type, length(html_template) AS tpl_bytes,
       (html_template LIKE '%</section>%') AS has_close_tag
FROM content_components WHERE name = 'patent-check';
