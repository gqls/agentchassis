-- p4_05_guide_copyright.sql — idea.uk guide 2: "Copyright: what you already own, and what it does not protect".
--
-- Stage 7 of features_open/014 (owner asked for copyright next, 2026-07-25). Follows the recipe
-- proven by p4_01 and written up as RUNBOOK Phase 5 — including slot_name (the trap that cost a
-- round on the patents guide) set from content_components.function.
--
-- CONTENT POLICY (unchanged from p4_01, and this is stage 7 of the "authored" band): hand-written,
-- not LLM-generated, because claims-verification V5 is inert and bugs_open/043 is a live
-- fabrication lane. Sections locked by p4_03 once verified live.
--
-- WHY THIS GUIDE EARNS ITS PLACE rather than being a patents footnote: the single most expensive
-- copyright mistake a small business makes is assuming it owns what it paid a freelancer to make.
-- It does not — a contractor keeps copyright unless there is a WRITTEN, SIGNED assignment
-- (CDPA 1988 s.90(3)); only employees' work vests in the employer by default (s.11(2)). That one
-- fact is worth the page on its own, and it is invisible from the patents guide.
--
-- The AI section deliberately says the law is UNSETTLED rather than picking a side. Whether
-- AI-generated output attracts copyright in the UK, and who would own it, is contested and under
-- active government review, and training-data cases are live. Stating a confident answer there
-- would be exactly the fabrication this content policy exists to prevent.
--
-- No new components: hero (23f95f00) + Generic Text Block (8d81e665) + call-to-action (0197e8d7),
-- same as the patents guide. nav_order 20 so it sorts after Patents (10) in the hub, which orders
-- by COALESCE(nav_order,100) then name.

\set ON_ERROR_STOP on

BEGIN;

DO $guard$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM pages
   WHERE site_id = '1244516d-014d-421c-88c6-090bb1e9552a'
     AND url = '/guides/copyright/index.html';
  IF n > 0 THEN
    RAISE EXCEPTION 'ABORT: /guides/copyright/index.html already exists (% row(s)) — p4_05 already ran.', n;
  END IF;
END
$guard$;

INSERT INTO pages
  (site_id, name, url, title, page_type, status, meta_description, topics,
   nav_label, nav_order, in_header, in_footer, build_status, sections)
VALUES
  ('1244516d-014d-421c-88c6-090bb1e9552a',
   'guide-copyright',
   '/guides/copyright/index.html',
   'Copyright: what you already own, and what it does not protect',
   'guide',
   'active',
   'A plain-English UK guide to copyright — why it is automatic, why your freelancer may own your website, what it protects, and where it stops.',
   ARRAY['copyright','intellectual property','protecting ideas','uk law'],
   'Copyright',
   20,
   false,
   false,
   'planned',
   '[]'::jsonb);

-- Section 1 — hero.
INSERT INTO page_components (page_id, component_id, slot_name, position, content_data, build_status)
SELECT p.id, '23f95f00-f293-466e-b43a-81791ea0fc6c', 'hero', 1,
       jsonb_build_object(
         'headline',          'Copyright: you already have it. The question is who owns it.',
         'subheadline',       'Copyright costs nothing and happens automatically — which is exactly why people never check who ended up with it. A plain-English guide to what UK copyright covers, where it stops, and the one paragraph you need in every freelancer contract.',
         'cta_text',          'Get a verified idea report',
         'cta_url',           '/report.html',
         'secondary_cta',     'Read the patents guide',
         'secondary_cta_url', '/guides/patents/index.html'
       ),
       'pending'
FROM pages p
WHERE p.site_id = '1244516d-014d-421c-88c6-090bb1e9552a'
  AND p.url = '/guides/copyright/index.html';

-- Section 2 — the guide body.
INSERT INTO page_components (page_id, component_id, slot_name, position, content_data, build_status)
SELECT p.id, '8d81e665-3ee0-443d-a873-690268c15fbb', 'generic-text-block', 2,
       jsonb_build_object(
         'heading', $h$What copyright gives you for free, and what it quietly refuses to do$h$,
         'content', $c$
<p><strong>The short version.</strong> Copyright is automatic, free, and lasts a very long time. It protects the <em>thing you made</em> — the actual words, code, drawing or recording — and not the idea behind it. There is nothing to register in the UK, which is a relief and also the problem: because nobody ever files anything, nobody ever checks who owns what, and businesses routinely discover years later that their website, their logo or their app belongs to somebody they once paid an invoice to.</p>

<h3>1. It happens by itself. There is no UK copyright register.</h3>
<p>The moment an original work is recorded in some form — typed, saved, drawn, filmed, recorded — copyright exists in it. You do not apply, you do not pay, and you do not need to put a © symbol anywhere. Unlike the United States, the UK has no copyright register and no registration to make.</p>
<p>The © notice is still worth using. It does not create the right, but it tells people the work is claimed, names who to ask, and makes "I assumed it was free to use" a much harder argument. Write it as <em>© 2026 Your Name or Company</em>.</p>

<h3>2. It protects the expression, not the idea</h3>
<p>This is the line that catches people out. Copyright stops someone <strong>copying</strong> your work. It does not stop them having the same idea, solving the same problem, or building a product that does exactly what yours does — as long as they wrote their own version.</p>
<p>So copyright protects your source code, but not the feature. It protects your article, but not the argument. It protects your drawings of the device, but not the device. If what is valuable is the <em>concept</em> rather than the words you used to express it, copyright is not the tool you are looking for — see the patents guide, or think about keeping it confidential.</p>

<h3>3. What is actually covered</h3>
<ul>
  <li><strong>Literary works</strong> — which in UK law explicitly includes <strong>computer programs</strong>, both source and compiled, plus preparatory design material.</li>
  <li><strong>Artistic works</strong> — drawings, photographs, diagrams, logos, maps, sculpture, and works of artistic craftsmanship.</li>
  <li><strong>Dramatic and musical works</strong>, and separately the <strong>sound recordings</strong>, <strong>films</strong> and <strong>broadcasts</strong> made of them.</li>
  <li><strong>The typographical arrangement</strong> of a published edition — the layout itself, as distinct from the text.</li>
  <li><strong>Databases</strong> get two bites: copyright if the selection or arrangement is original, plus a separate <em>database right</em> protecting substantial investment in obtaining or verifying the contents. Database right is much shorter-lived and has its own rules.</li>
</ul>
<p>"Original" sets a low bar — it means you created it rather than copied it, not that it is any good.</p>

<h3>4. How long it lasts</h3>
<ul>
  <li>Literary, dramatic, musical and artistic works — <strong>the author's life plus 70 years</strong>. For software written by a person, that is effectively forever in business terms.</li>
  <li>Sound recordings — 70 years from release.</li>
  <li>Broadcasts — 50 years.</li>
  <li>Typographical arrangement of a published edition — 25 years.</li>
  <li>Works with no human author, generated by a computer, sit under a separate 50-year rule — and see the AI section below, because how that applies now is genuinely unsettled.</li>
</ul>

<h3>5. Who owns it — the part that costs people real money</h3>
<p>The default is that <strong>the author owns the copyright</strong>. There are two things you must know about how that default plays out.</p>
<p><strong>Employees:</strong> work made by an employee in the course of their employment belongs to the <strong>employer</strong> automatically. Nothing to sign.</p>
<p><strong>Freelancers, contractors and agencies: they keep it.</strong> Paying someone to make something does not transfer the copyright to you. The designer who drew your logo owns your logo. The developer who built your site owns the code. The photographer owns the photographs of your own product. You have, at best, an implied licence to use the work for the purpose you commissioned it for — which is not the same as owning it, and it will not survive due diligence when you raise money or sell.</p>
<p><strong>The fix is one clause and it must be in writing.</strong> An assignment of copyright is only effective if it is in writing and signed by the person giving it up. Get it signed <em>before</em> the work starts, when you still have leverage, and make it cover everything created for you. A licence, however generous, is not ownership. If you have already paid for work without one, ask for a retrospective assignment now, politely, while the relationship is good — this gets very much harder later.</p>
<p>Also worth knowing: <strong>moral rights</strong> are separate and stay with the human author even after they assign the copyright — the right to be identified as author (which in the UK has to be positively asserted) and to object to derogatory treatment of the work. They can be waived in writing but not transferred. For most commercial work a waiver is standard; just know it is a separate line in the contract.</p>

<h3>6. Where copyright stops</h3>
<ul>
  <li>It does <strong>not</strong> protect a name, title or slogan — that is trade marks.</li>
  <li>It does <strong>not</strong> protect the shape or appearance of a product — that is registered designs and design right.</li>
  <li>It does <strong>not</strong> protect a method, system, business process or invention — that is patents, if anything.</li>
  <li>It does <strong>not</strong> stop independent creation. If they genuinely did not copy you, there is no infringement, however similar the result.</li>
  <li>It does <strong>not</strong> help much if you cannot show what you had and when.</li>
</ul>

<h3>7. Proving what you had, and when</h3>
<p>With no register, evidence is your own responsibility, and it is easy. Keep dated version control history — a git log with real commit dates is unusually good evidence, and signed commits better still. Keep dated backups and original files with their metadata intact. Keep the drafts, not just the final. For something significant you can deposit a copy with a solicitor.</p>
<p>The old advice about posting yourself a sealed envelope is folklore, not law: it proves very little and is trivially faked. Dated, systematic records that you kept anyway are worth far more.</p>

<h3>8. Licensing, and the licences you are already relying on</h3>
<p>You can license your work without giving up ownership, which is usually what you actually want: keep the copyright, grant what is needed. An <em>exclusive</em> licence has to be in writing to be effective. Be specific about territory, duration, medium and whether the other side can sub-license.</p>
<p>Remember that open-source licences are copyright licences and using the code obliges you to their terms. Permissive ones typically only require attribution. <strong>Copyleft licences can require you to release your own source code</strong> if you distribute software built on them — which for a product company can be an existential surprise rather than a legal footnote. Know what is in your dependency tree before you ship, not after.</p>

<h3>9. AI-generated work — genuinely unsettled, and we are not going to pretend otherwise</h3>
<p>Two questions matter commercially, and neither has a safe, settled UK answer right now: whether material generated by an AI system attracts copyright at all and who would own it, and whether training AI systems on copyright works is lawful without permission. The UK has a long-standing provision for "computer-generated" works with no human author, but how it applies to modern generative systems is contested, the government has been consulting on changing it, and training-data disputes are live in the courts.</p>
<p>The practical posture until that settles: <strong>do not assume you own AI output</strong> that you intend to build a business on, keep a record of the human contribution to anything that matters, check what rights your AI vendor's terms actually grant you, and <strong>do not assume you may freely train on other people's works</strong>. If a specific piece of this is commercially load-bearing for you, that is a question for a solicitor, not an article.</p>

<h3>10. If someone copies you</h3>
<p>Start with the cheap routes: a clear, polite letter setting out what was copied and what you want; a takedown request to the platform, host or app store, which is often faster than anything legal; and a search-engine removal request where relevant.</p>
<p>If it goes further, copyright is in a much better position than patents for a smaller business: copyright, trade mark, passing off and unregistered design right claims <strong>can</strong> be brought in the Intellectual Property Enterprise Court's small claims track, which is designed to be usable without heavy legal costs. Patents specifically cannot. So of all the rights on this site, copyright is the one you are most realistically able to enforce.</p>

<h3>The whole thing, on one line</h3>
<p>You already own it → unless a freelancer made it, in which case get a signed assignment today → keep dated records → license deliberately, and read your open-source terms → and remember it protects what you wrote, not what you thought of.</p>

<p><em>This guide is general information about UK law and practice, written to help you ask better questions. It is not legal advice, it does not create a professional relationship, and it cannot take account of your circumstances — particularly on contracts, moral rights and anything involving AI, where the details decide the answer. Take advice from a solicitor with intellectual property experience before relying on any of it.</em></p>
$c$
       ),
       'pending'
FROM pages p
WHERE p.site_id = '1244516d-014d-421c-88c6-090bb1e9552a'
  AND p.url = '/guides/copyright/index.html';

-- Section 3 — the funnel.
INSERT INTO page_components (page_id, component_id, slot_name, position, content_data, build_status)
SELECT p.id, '0197e8d7-1adc-43d6-ab32-d0716e013175', 'call-to-action', 3,
       jsonb_build_object(
         'headline',          'Protecting it is the easy part. Knowing whether it is worth protecting is not.',
         'subheadline',       'Copyright you already have, for free. The harder question is whether the thing you are protecting is one anybody wants. A Verified Idea Report researches one idea properly — the market, who else is there, and a specific next step. £29, and it will tell you if the answer is no.',
         'primary_cta',       'Get a verified idea report',
         'primary_cta_url',   '/report.html',
         'secondary_cta',     'Should you patent it? Free check',
         'secondary_cta_url', '/tools/patent-check/index.html'
       ),
       'pending'
FROM pages p
WHERE p.site_id = '1244516d-014d-421c-88c6-090bb1e9552a'
  AND p.url = '/guides/copyright/index.html';

-- Backfill pages.sections (the rerender path never writes it — see p4_01c).
UPDATE pages p
SET sections = (
      SELECT COALESCE(jsonb_agg(pc.slot_name ORDER BY pc.position), '[]'::jsonb)
      FROM page_components pc WHERE pc.page_id = p.id AND COALESCE(pc.slot_name,'') <> ''
    ),
    updated_at = now()
WHERE p.site_id = '1244516d-014d-421c-88c6-090bb1e9552a'
  AND p.url = '/guides/copyright/index.html';

DO $guard2$
DECLARE nbad int; ntot int; nslot int;
BEGIN
  SELECT count(*) FILTER (WHERE pc.content_data IS NULL OR pc.content_data = '{}'::jsonb),
         count(*),
         count(*) FILTER (WHERE COALESCE(pc.slot_name,'') = '')
    INTO nbad, ntot, nslot
  FROM page_components pc JOIN pages p ON p.id = pc.page_id
  WHERE p.site_id = '1244516d-014d-421c-88c6-090bb1e9552a'
    AND p.url = '/guides/copyright/index.html';
  IF ntot <> 3 THEN RAISE EXCEPTION 'ABORT: expected 3 sections, got %.', ntot; END IF;
  IF nbad > 0 THEN RAISE EXCEPTION 'ABORT: % section(s) have empty content_data.', nbad; END IF;
  IF nslot > 0 THEN RAISE EXCEPTION 'ABORT: % section(s) have no slot_name — they would be CARRIED, not rendered.', nslot; END IF;
END
$guard2$;

COMMIT;

SELECT p.id AS page_id, p.url, p.page_type, p.build_status, p.nav_order, p.sections
FROM pages p
WHERE p.site_id = '1244516d-014d-421c-88c6-090bb1e9552a'
  AND p.url = '/guides/copyright/index.html';
