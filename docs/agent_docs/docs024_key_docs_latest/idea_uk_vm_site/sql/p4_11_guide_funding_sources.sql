-- p4_11_guide_funding_sources.sql — idea.uk guide 4: "Funding: the sources" (stage 9 of features_open/014).
--
-- Companion to p4_10 (the ways). SAME FIGURE POLICY, plus an INSTITUTION POLICY: this guide
-- names only durable, major institutions whose existence is long-established public fact
-- (Innovate UK, the British Business Bank, UKRI, the devolved development agencies, the King's
-- Trust, Companies House-adjacent machinery like growth hubs). It names NO specific scheme
-- amounts, NO deadlines, NO individual funds or platforms by name where a class will do —
-- naming a specific crowdfunding platform or VC fund is both an implicit endorsement and a
-- staleness bomb. Everything points the reader at the institution's own site as the source of
-- truth. This is the anti-043 discipline applied to institutions rather than statistics.
--
-- nav_order 40 (Patents 10, Copyright 20, Ways 30).

\set ON_ERROR_STOP on

BEGIN;

DO $guard$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM pages
   WHERE site_id = '1244516d-014d-421c-88c6-090bb1e9552a'
     AND url = '/guides/funding-sources/index.html';
  IF n > 0 THEN
    RAISE EXCEPTION 'ABORT: /guides/funding-sources/index.html already exists — p4_11 already ran.';
  END IF;
END
$guard$;

INSERT INTO pages
  (site_id, name, url, title, page_type, status, meta_description, topics,
   nav_label, nav_order, in_header, in_footer, build_status, sections)
VALUES
  ('1244516d-014d-421c-88c6-090bb1e9552a',
   'guide-funding-sources',
   '/guides/funding-sources/index.html',
   'Funding an idea: the sources — where the money actually is in the UK',
   'guide',
   'active',
   'A plain-English UK map of who actually funds ideas — government bodies, banks, angels, platforms and trusts — and how to approach each without wasting months.',
   ARRAY['funding','grants','investors','startups','uk'],
   'Funding: the sources',
   40,
   false,
   false,
   'planned',
   '[]'::jsonb);

INSERT INTO page_components (page_id, component_id, slot_name, position, content_data, build_status)
SELECT p.id, '23f95f00-f293-466e-b43a-81791ea0fc6c', 'hero', 1,
       jsonb_build_object(
         'headline',          'Funding an idea: where the money actually is',
         'subheadline',       'Once you know which mechanism fits your stage, the question becomes who to actually ask. A UK map of the funders — government bodies, banks, angels, platforms and trusts — with an honest note on what each one is really looking for.',
         'cta_text',          'Get a verified idea report',
         'cta_url',           '/report.html',
         'secondary_cta',     'Start with the ways, if you haven''t',
         'secondary_cta_url', '/guides/funding-ways/index.html'
       ),
       'pending'
FROM pages p
WHERE p.site_id = '1244516d-014d-421c-88c6-090bb1e9552a'
  AND p.url = '/guides/funding-sources/index.html';

INSERT INTO page_components (page_id, component_id, slot_name, position, content_data, build_status)
SELECT p.id, '8d81e665-3ee0-443d-a873-690268c15fbb', 'generic-text-block', 2,
       jsonb_build_object(
         'heading', $h$A working map of UK funders — and what each one is actually for$h$,
         'content', $c$
<p><strong>The short version.</strong> The UK funding landscape looks like alphabet soup from outside, but it reduces to a short list of institution types, each answering a different question. Schemes open, close and rename constantly, so this guide gives you the durable map and tells you where the current details live — always the institution's own site, never an article, including this one.</p>

<h3>1. The state's innovation money: Innovate UK and UKRI</h3>
<p><strong>Innovate UK</strong> is the UK's national innovation agency — the main route to grant funding for developing something genuinely new with commercial potential. It runs regular funding competitions, some open to any sector, some themed. Expect competition, real application effort, match funding, and payment against evidenced project spend — grants fund <em>projects</em>, not runway. It sits within <strong>UKRI</strong> (UK Research and Innovation), which also funds the research councils; if your idea is close to university research, that side matters too, usually through a university partner.</p>
<p>Two adjacent things worth knowing exist: <strong>Knowledge Transfer Partnerships</strong> place a graduate researcher inside your business, part-funded, to work on an innovation problem with a university; and the <strong>Catapult centres</strong> are national technology facilities (in areas like manufacturing, satellites, energy and medicines discovery) that let smaller companies use equipment and expertise they could never buy. Current competitions and rules: the Innovate UK and UKRI websites.</p>

<h3>2. The state's finance arm: the British Business Bank</h3>
<p>The <strong>British Business Bank</strong> is the government's economic development bank. You mostly do not deal with it directly — it works through partners — but two of its programmes matter to people with ideas: the <strong>Start Up Loans</strong> programme (government-backed personal loans for starting a business, with mentoring attached, delivered through partner organisations) and guarantee schemes that make ordinary bank lending available to businesses a bank would otherwise decline. If a bank has said no, it is worth asking whether a guarantee-backed product would change the answer. Current products and delivery partners: the British Business Bank's own site.</p>

<h3>3. The devolved and regional bodies</h3>
<p>Where you are changes what exists. <strong>Scottish Enterprise</strong> (and its Highlands and south-of-Scotland siblings), the <strong>Development Bank of Wales</strong>, and <strong>Invest Northern Ireland</strong> each run their own grants, loans and equity for their nations — often more accessible than the UK-wide schemes. In England, local <strong>Growth Hubs</strong> are the front door: free business support that knows what is currently open in your area. Councils and combined authorities run their own funds too. If you only check the national schemes, you are missing the pot most likely to say yes.</p>

<h3>4. Banks — and the debt you already qualify for</h3>
<p>High-street banks lend to trading businesses with predictable money coming in; for a pre-revenue idea they will mostly (sensibly) say no, and the Start Up Loans route above exists precisely for that gap. Once you are trading, the conversation changes: overdrafts, term loans, invoice finance against what customers owe you, asset finance for equipment. Compare more than one bank, ask about guarantee-backed schemes by name, and read the personal-guarantee clause twice.</p>

<h3>5. Angel investors — individuals backing you early</h3>
<p>Angels are experienced individuals investing their own money at the stage funds will not touch, usually in groups (syndicates) and usually expecting your company to qualify for the UK's early-investment tax reliefs — most serious angels will ask about HMRC advance assurance early, so read up on that before the first meeting (see the ways guide). Find them through regional angel networks and syndicates — the <strong>UK Business Angels Association</strong> is the trade body and keeps a directory — and through founders they have already backed, which is the introduction that actually works. What convinces an angel is evidence of demand plus a founder they believe; a warm, specific approach beats fifty cold decks.</p>

<h3>6. Venture capital — for the few ideas built for it</h3>
<p>VC funds invest other people's money in businesses designed to grow very large, very fast, and be sold. Most good businesses are not that shape, and taking VC money commits you to it. If yours genuinely is, research funds by stage and sector before approaching, and get introduced through their portfolio founders. If you are pre-evidence, you are early: use the free tools and the report first — a VC's first question is the demand question everyone else asks, with more zeros attached.</p>

<h3>7. Crowdfunding platforms</h3>
<p>Reward campaigns run on the big international platforms; equity crowdfunding runs on FCA-regulated UK platforms. Which platform matters far less than the campaign: successful raises arrive with an audience already warmed up, clear numbers, and — for equity — the same disclosure discipline as any other investment round. Re-read the patents warning in the ways guide before publishing how anything works.</p>

<h3>8. Trusts, foundations and social funders</h3>
<p>If you are young, the <strong>King's Trust</strong> (formerly the Prince's Trust) supports people starting businesses with funding, mentoring and courses. If the idea's point is social or environmental impact, a separate world exists — social investment funds, foundation grants, community shares — with its own funders who accept returns the commercial market would not. If your idea has a social mission at its core, search "social investment UK" and start from the infrastructure bodies you find there rather than from commercial investor lists.</p>

<h3>9. Universities — not just for students</h3>
<p>Enterprise programmes, incubators and competitions at your local university are often open to alumni and sometimes to the public; if your idea grew out of research, the university's technology transfer office owns that conversation. Also the cheapest route to expensive facilities and to the KTP scheme above.</p>

<h3>10. How to work this map without losing six months</h3>
<ul>
  <li><strong>Start local and free:</strong> your Growth Hub (or national equivalent) knows what is currently open and will tell you for nothing.</li>
  <li><strong>Match before you apply:</strong> pick the two or three sources whose question you can actually answer today. A grant body asks "is this innovative?"; a lender asks "how do you repay?"; an investor asks "how does this get big?". Applying to the wrong one is the main way people burn months.</li>
  <li><strong>Check currency on everything:</strong> scheme names, amounts and windows change several times a year. The institution's own website is the only source that counts.</li>
  <li><strong>Beware paid intermediaries:</strong> anyone charging upfront to "find you grants" or "introduce you to investors" deserves hard questions. The front doors above are free.</li>
</ul>

<h3>The whole thing, on one line</h3>
<p>Growth Hub first for what is open now → Innovate UK for innovation grants → British Business Bank routes for early debt → your nation's development agency → angels via networks and warm introductions → and match the funder's question to the evidence you actually have, which is the real gate on all of it.</p>

<p><em>This guide is general information about the UK funding landscape, written to help you ask better questions. It is not financial advice and it deliberately avoids amounts, rates and scheme details, because they change frequently — always check the institution's own website or gov.uk for the current position, and take regulated advice where a decision is significant. Nothing here is an endorsement of any provider.</em></p>
$c$
       ),
       'pending'
FROM pages p
WHERE p.site_id = '1244516d-014d-421c-88c6-090bb1e9552a'
  AND p.url = '/guides/funding-sources/index.html';

INSERT INTO page_components (page_id, component_id, slot_name, position, content_data, build_status)
SELECT p.id, '0197e8d7-1adc-43d6-ab32-d0716e013175', 'call-to-action', 3,
       jsonb_build_object(
         'headline',          'Know which door to knock on — and what to hold in your hand',
         'subheadline',       'Every source on this page asks for evidence before money. A Verified Idea Report gives you the researched case — the market, the competition, and a specific next step — for £29. Cheaper than six months of applying to the wrong doors.',
         'primary_cta',       'Get a verified idea report',
         'primary_cta_url',   '/report.html',
         'secondary_cta',     'Should you patent it first? Free check',
         'secondary_cta_url', '/tools/patent-check/index.html'
       ),
       'pending'
FROM pages p
WHERE p.site_id = '1244516d-014d-421c-88c6-090bb1e9552a'
  AND p.url = '/guides/funding-sources/index.html';

UPDATE pages p
SET sections = (
      SELECT COALESCE(jsonb_agg(pc.slot_name ORDER BY pc.position), '[]'::jsonb)
      FROM page_components pc WHERE pc.page_id = p.id AND COALESCE(pc.slot_name,'') <> ''
    ),
    updated_at = now()
WHERE p.site_id = '1244516d-014d-421c-88c6-090bb1e9552a'
  AND p.url = '/guides/funding-sources/index.html';

DO $guard2$
DECLARE nbad int; ntot int; nslot int;
BEGIN
  SELECT count(*) FILTER (WHERE pc.content_data IS NULL OR pc.content_data = '{}'::jsonb),
         count(*),
         count(*) FILTER (WHERE COALESCE(pc.slot_name,'') = '')
    INTO nbad, ntot, nslot
  FROM page_components pc JOIN pages p ON p.id = pc.page_id
  WHERE p.site_id = '1244516d-014d-421c-88c6-090bb1e9552a'
    AND p.url = '/guides/funding-sources/index.html';
  IF ntot <> 3 THEN RAISE EXCEPTION 'ABORT: expected 3 sections, got %.', ntot; END IF;
  IF nbad > 0 THEN RAISE EXCEPTION 'ABORT: % section(s) empty content_data.', nbad; END IF;
  IF nslot > 0 THEN RAISE EXCEPTION 'ABORT: % section(s) missing slot_name.', nslot; END IF;
END
$guard2$;

COMMIT;

SELECT p.id AS page_id, p.url, p.build_status, p.nav_order, p.sections FROM pages p
WHERE p.site_id = '1244516d-014d-421c-88c6-090bb1e9552a' AND p.url = '/guides/funding-sources/index.html';
