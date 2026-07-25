-- p4_10_guide_funding_ways.sql — idea.uk guide 3: "Funding: the ways" (stage 8 of features_open/014).
--
-- Owner 2026-07-25: "please go ahead" (the funding pair, ways then sources). Same recipe as
-- p4_01/p4_05 (RUNBOOK Phase 5), same content policy: HAND-AUTHORED because claims-verification
-- V5 is inert and funding eligibility/figures are prime fabrication territory (bugs_open/043,
-- and 061's invented price tables are the exact shape a "grant amounts" paragraph would take).
--
-- FIGURE POLICY, stricter here than the patents guide needed to be: this guide names
-- MECHANISMS, never rates, caps, thresholds or deadlines. SEIS/EIS relief percentages, loan
-- limits and grant sizes all change with fiscal events; a stated figure would be stale within
-- the year and indistinguishable from a fabricated one. Every section that touches numbers says
-- "check the current figures" instead. The SOURCES guide (p4_11) carries the same rule for
-- institutions.
--
-- nav_order 30 (after Patents 10, Copyright 20).

\set ON_ERROR_STOP on

BEGIN;

DO $guard$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM pages
   WHERE site_id = '1244516d-014d-421c-88c6-090bb1e9552a'
     AND url = '/guides/funding-ways/index.html';
  IF n > 0 THEN
    RAISE EXCEPTION 'ABORT: /guides/funding-ways/index.html already exists — p4_10 already ran.';
  END IF;
END
$guard$;

INSERT INTO pages
  (site_id, name, url, title, page_type, status, meta_description, topics,
   nav_label, nav_order, in_header, in_footer, build_status, sections)
VALUES
  ('1244516d-014d-421c-88c6-090bb1e9552a',
   'guide-funding-ways',
   '/guides/funding-ways/index.html',
   'Funding an idea: the ways — and which one fits which stage',
   'guide',
   'active',
   'A plain-English UK guide to the funding mechanisms — bootstrapping, grants, loans, equity, crowdfunding — what each really costs you, and which stage each fits.',
   ARRAY['funding','grants','investment','startups','uk'],
   'Funding: the ways',
   30,
   false,
   false,
   'planned',
   '[]'::jsonb);

INSERT INTO page_components (page_id, component_id, slot_name, position, content_data, build_status)
SELECT p.id, '23f95f00-f293-466e-b43a-81791ea0fc6c', 'hero', 1,
       jsonb_build_object(
         'headline',          'Funding an idea: the ways',
         'subheadline',       'Money for an idea comes in about eight distinct shapes, and each one costs you something different — equity, interest, obligations, or time. Most early ideas should take none of them yet. A plain-English guide to the mechanisms, so you can tell which one fits where you actually are.',
         'cta_text',          'Get a verified idea report',
         'cta_url',           '/report.html',
         'secondary_cta',     'The sources: where the money actually is',
         'secondary_cta_url', '/guides/funding-sources/index.html'
       ),
       'pending'
FROM pages p
WHERE p.site_id = '1244516d-014d-421c-88c6-090bb1e9552a'
  AND p.url = '/guides/funding-ways/index.html';

INSERT INTO page_components (page_id, component_id, slot_name, position, content_data, build_status)
SELECT p.id, '8d81e665-3ee0-443d-a873-690268c15fbb', 'generic-text-block', 2,
       jsonb_build_object(
         'heading', $h$The eight shapes money comes in, and what each one really costs$h$,
         'content', $c$
<p><strong>The short version.</strong> Almost every way of funding an idea is one of eight mechanisms. They differ in what they cost you — a share of the business, interest, obligations, or months of your time — and in which stage of an idea they will actually fund. The most common funding mistake is not picking the wrong source; it is picking the wrong <em>mechanism</em> for the stage you are at, usually by chasing investment before there is any evidence anyone wants the thing.</p>

<h3>0. First, the uncomfortable question</h3>
<p>Funding is fuel, and fuel is only useful if the engine works. If you cannot yet show that anyone wants what you are building, money mostly lets you build the wrong thing faster. The cheapest capital at the idea stage is <strong>evidence</strong>: a test that shows real people doing something costly-to-them — pre-ordering, signing up, paying a deposit — before you commit. Everything below gets dramatically easier once you have it, because every funder on this list is trying to answer the same question: is there demand?</p>

<h3>1. Bootstrapping — funding it from yourself and your customers</h3>
<p>Your savings, your salary from the day job, and — best of all — your customers' money, by selling early versions, pre-orders or services while you build. It costs you no equity and no interest, and it forces the discipline every other mechanism lets you postpone: making something people pay for.</p>
<p>Its limits are real: it is slow, it caps how big a swing you can take, and in a market with a closing window it can mean arriving second. But if the idea can be started small, this should usually be the default — every other row on this list is more expensive than it looks.</p>

<h3>2. Friends and family</h3>
<p>Common, fast, and the most dangerous money on this page — not commercially, but personally. If it must happen, treat it more formally than feels natural: write down whether it is a gift, a loan or a share purchase; what happens if the idea fails; and never take money the giver cannot afford to lose. The default failure mode is not the business — it is Christmas dinner.</p>

<h3>3. Grants — money you do not pay back</h3>
<p>Public bodies and foundations fund work they want to exist: innovation, research and development, regional growth, social outcomes. A grant takes no equity and charges no interest, which is why everybody wants one and why they are competitive, slow, and demanding — applications take real weeks, most schemes want <em>match funding</em> (you bring part of the cost yourself), the money is usually paid in arrears against evidenced spend, and it comes with reporting obligations. Grants suit a specific shape of idea: genuine innovation with a clear project plan. They are a poor fit for "I need money to live on while I figure this out". Schemes, amounts and deadlines change constantly — check the funder's own site, and be wary of anyone quoting you figures from an article, including this one.</p>

<h3>4. Competitions and awards</h3>
<p>Pitch competitions, innovation prizes, university enterprise awards. The money is usually modest, but it is non-dilutive, the deadline forces you to sharpen the pitch, and the credibility and contacts often outlast the cheque. Worth entering when one fits; not a funding strategy on its own.</p>

<h3>5. Debt — borrowing it</h3>
<p>Bank loans, government-backed start-up loan schemes, and later on invoice finance and asset finance. Debt keeps your equity — nobody owns a piece of your idea — but it must be repaid whether or not the idea works, and for early businesses lenders usually want a <strong>personal guarantee</strong>, which means the risk lands on you, not the company. Debt fits a business with predictable money coming in: known costs, orders in hand, invoices to bridge. It is a bad fit for open-ended product development, where the repayments start before the revenue does.</p>

<h3>6. Equity — selling a piece of it</h3>
<p>Angel investors (individuals investing their own money) and venture capital funds (investing other people's, at larger scale, later). Equity is the only mechanism on this list built for the case where the idea needs serious money <em>before</em> it can prove itself, and the investor shares the downside — if it fails, you owe nothing. What it costs is permanent: a share of ownership, usually a say in decisions, and an expectation of growth and eventual sale that you cannot quietly retire from. Take equity when the ambition genuinely needs it, not because it is the kind of funding that gets written about.</p>
<p>The UK does something unusual here: government schemes exist that give individual investors substantial income-tax relief for backing early companies — which is a large part of why UK angels invest at all. Most serious UK angels will expect your company to qualify, and qualifying (an "advance assurance" from HMRC) is a normal early step. The reliefs, caps and rules change; check the current position on gov.uk or with an accountant before promising anything to an investor.</p>

<h3>7. Crowdfunding — many small backers</h3>
<p>Three different things share the name, and they belong to different stages. <strong>Reward</strong> crowdfunding pre-sells the product — which makes it a demand test that pays you, the best kind, and a genuine option at the idea stage if you have an audience. <strong>Equity</strong> crowdfunding sells small stakes to many investors on a regulated platform — real fundraising, with real obligations, suited to consumer businesses whose customers want to be owners. <strong>Debt</strong> crowdfunding is borrowing from a crowd, and belongs with the other debt on this page. A warning that applies to all three: a public campaign is a public disclosure. If any part of the idea might be patentable, read the patents guide <em>before</em> you launch one — publishing the how can end your patent rights.</p>

<h3>8. Money you are already owed — the tax system</h3>
<p>Not funding, but frequently worth more than a grant: if you are building something genuinely new and technically uncertain, UK research-and-development tax relief can return a meaningful part of your development spend, as reduced tax or cash. The rules, rates and definitions have been reorganised more than once in recent years and this is now firmly accountant territory — but knowing the mechanism exists changes what you can afford to attempt. Ask an accountant who handles R&D claims routinely, and be wary of cold-calling "R&D specialists" who promise the moon on commission.</p>

<h3>9. Accelerators and incubators</h3>
<p>Programmes that trade a small amount of money and a lot of structure — mentoring, workspace, a cohort, a demo day — usually for a small equity stake, sometimes for nothing. The good ones are worth more than the cheque; the mediocre ones cost you equity and three months. Judge them like an investor would judge you: who are the alumni, and what did the programme actually do for them?</p>

<h3>Matching the mechanism to the stage</h3>
<ul>
  <li><strong>Idea, no evidence yet</strong> — bootstrap, and buy evidence cheaply: a test, a pre-order page, a reward crowdfunding campaign if you have an audience. Most other doors open with what this stage produces.</li>
  <li><strong>Evidence, building the first real version</strong> — bootstrapping plus a grant if the work is genuinely innovative; a start-up loan if there is early revenue; angels if it truly cannot be built small.</li>
  <li><strong>Selling, growing</strong> — debt against predictable revenue; equity if the opportunity outruns what revenue can fund; R&D relief on qualifying development all the while.</li>
</ul>

<h3>The whole thing, on one line</h3>
<p>Get evidence before you get money → prefer money that keeps your equity and your freedom → match the mechanism to the stage → and read the terms of anything you sign, because every mechanism on this page is cheap in the column you are looking at and expensive in one you are not.</p>

<p><em>This guide is general information about funding in the UK, written to help you ask better questions. It is not financial advice, and funding schemes, tax reliefs and their rules change frequently — check current details with the scheme operator, gov.uk, or a qualified accountant or adviser before acting. Nothing here recommends any specific provider.</em></p>
$c$
       ),
       'pending'
FROM pages p
WHERE p.site_id = '1244516d-014d-421c-88c6-090bb1e9552a'
  AND p.url = '/guides/funding-ways/index.html';

INSERT INTO page_components (page_id, component_id, slot_name, position, content_data, build_status)
SELECT p.id, '0197e8d7-1adc-43d6-ab32-d0716e013175', 'call-to-action', 3,
       jsonb_build_object(
         'headline',          'Every funder will ask the same question. Answer it first.',
         'subheadline',       'Grant assessors, lenders and investors all want the same thing: evidence that somebody wants this. A Verified Idea Report researches one idea properly — the market, the competition, and a specific, affordable way to test real demand — for £29.',
         'primary_cta',       'Get a verified idea report',
         'primary_cta_url',   '/report.html',
         'secondary_cta',     'Now read: where the money actually is',
         'secondary_cta_url', '/guides/funding-sources/index.html'
       ),
       'pending'
FROM pages p
WHERE p.site_id = '1244516d-014d-421c-88c6-090bb1e9552a'
  AND p.url = '/guides/funding-ways/index.html';

UPDATE pages p
SET sections = (
      SELECT COALESCE(jsonb_agg(pc.slot_name ORDER BY pc.position), '[]'::jsonb)
      FROM page_components pc WHERE pc.page_id = p.id AND COALESCE(pc.slot_name,'') <> ''
    ),
    updated_at = now()
WHERE p.site_id = '1244516d-014d-421c-88c6-090bb1e9552a'
  AND p.url = '/guides/funding-ways/index.html';

DO $guard2$
DECLARE nbad int; ntot int; nslot int;
BEGIN
  SELECT count(*) FILTER (WHERE pc.content_data IS NULL OR pc.content_data = '{}'::jsonb),
         count(*),
         count(*) FILTER (WHERE COALESCE(pc.slot_name,'') = '')
    INTO nbad, ntot, nslot
  FROM page_components pc JOIN pages p ON p.id = pc.page_id
  WHERE p.site_id = '1244516d-014d-421c-88c6-090bb1e9552a'
    AND p.url = '/guides/funding-ways/index.html';
  IF ntot <> 3 THEN RAISE EXCEPTION 'ABORT: expected 3 sections, got %.', ntot; END IF;
  IF nbad > 0 THEN RAISE EXCEPTION 'ABORT: % section(s) empty content_data.', nbad; END IF;
  IF nslot > 0 THEN RAISE EXCEPTION 'ABORT: % section(s) missing slot_name.', nslot; END IF;
END
$guard2$;

COMMIT;

SELECT p.id AS page_id, p.url, p.build_status, p.nav_order, p.sections FROM pages p
WHERE p.site_id = '1244516d-014d-421c-88c6-090bb1e9552a' AND p.url = '/guides/funding-ways/index.html';
