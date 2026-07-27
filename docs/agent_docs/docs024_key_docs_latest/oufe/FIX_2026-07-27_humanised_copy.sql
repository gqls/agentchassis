-- ============================================================================
-- oufe.com — rewrite the live copy in the house voice
--
-- Owner, 2026-07-27: "all the text is very AI-sounding … staccato and full of
-- 'it's this not that'." Measured on the live pages before this ran:
--     /            3 em dashes,  7 negative-frame constructions
--     /about.html  3 em dashes, 19 negative-frame constructions,
--                  6 sentences opening "It is…" / "It does…"
--
-- Source of the voice: REVERSE_ENGINEERED_STYLE_PROMPT_v3.md, built 2026-07-17
-- across three rounds of the owner critiquing the prompt's own output. Rule 3 is
-- the one he named: say what a thing is before saying what it isn't.
--
-- Some of the worst offenders were MY OWN hand-written copy from the 07-26
-- fallibility rewrite, not the generator's. "That does not make us right."
-- opens on a negative. "A citation shows you our source. It does not prove that
-- we read it properly." is a contrastive pair doing the same work twice. Both
-- are fixed below.
--
-- Every fact is preserved. This changes pacing and phrasing only. The honesty
-- content the owner approved stays exactly as approved: we cite our sources, we
-- can be wrong, the tools can be wrong, check anything that matters.
--
-- Companion: mig 228 makes this voice the writers' default so the next page
-- starts here, and the oufe `voice` spec turns on check_voice_tells so drift is
-- detectable afterwards.
-- ============================================================================

\set ON_ERROR_STOP on

BEGIN;

-- index / 1 — hero -----------------------------------------------------------
UPDATE page_components pc SET
  content_data = pc.content_data || jsonb_build_object(
    'subheadline',
    'OUFE explains how distressed capital structures actually work. The statutory framework, the creditor waterfall, and the arithmetic that decides who gets paid. We name our sources so you can check them, and we still get things wrong sometimes, so treat what you read here as a worked example rather than an authority. This is analysis of mechanism. It has nothing to say about what you should buy, and it is no substitute for legal advice on your own situation.'),
  updated_at = now()
FROM pages p WHERE p.id = pc.page_id
  AND p.site_id = (SELECT id FROM sites WHERE domain='oufe.com')
  AND p.name='index' AND pc.position=1;

-- index / 4 — closing call to action ------------------------------------------
UPDATE page_components pc SET
  content_data = pc.content_data || jsonb_build_object(
    'subheadline',
    'This analysis works through how corporate finance structures fail under stress. The statutory framework, the creditor waterfall, and the arithmetic that decides who recovers what. We name the document behind each figure and the date we read it, so you can check us. We do make mistakes, and so do the sources. Where we haven''t verified something yet, we say so.'),
  updated_at = now()
FROM pages p WHERE p.id = pc.page_id
  AND p.site_id = (SELECT id FROM sites WHERE domain='oufe.com')
  AND p.name='index' AND pc.position=4;

-- about / 1 — hero ------------------------------------------------------------
UPDATE page_components pc SET
  content_data = pc.content_data || jsonb_build_object(
    'subheadline',
    'OUFE covers the mechanics of UK restructuring. The legal frameworks, the creditor rankings, and the arithmetic of who recovers what. We cite everything so you can check it. A source can be wrong and our reading of it can be wrong, so treat what you find here as a worked example rather than an authority. Nothing on this site is investment advice.'),
  updated_at = now()
FROM pages p WHERE p.id = pc.page_id
  AND p.site_id = (SELECT id FROM sites WHERE domain='oufe.com')
  AND p.name='about' AND pc.position=1;

-- about / 2 — body + first highlight ------------------------------------------
UPDATE page_components pc SET
  content_data = jsonb_set(
    jsonb_set(pc.content_data, '{content}', to_jsonb(
      '<p>OUFE covers the mechanics of UK restructuring. Each case analysis works through what the statutory framework permits, who ranks where in the creditor waterfall, and what the arithmetic shows at different recovery assumptions. We build it from primary documents: court filings, company announcements, regulatory submissions and published accounts. Where a figure hasn''t been checked against one of those, we say so.</p>' ||
      '<p>Nothing here is investment advice, and none of it is a financial promotion. We don''t recommend buying, selling or holding anything. Explaining how a restructuring plan under Part 26A works, or how a court can bind a class that voted against a plan, is an explanation of legal mechanism. Your own matter needs a lawyer.</p>' ||
      '<p>The interactive tools can give you a wrong answer. Each one applies a single arithmetic rule to the numbers you type, and leaves out most of what decides a real case: security, guarantees, structural subordination, intercompany claims, contingent liabilities, contested valuation. We state the logic openly so you can judge it. Treat any result as a sketch, and please don''t rely on it.</p>' ||
      '<p>Where we don''t yet have a verified figure, we say which kind of document it will come from. We''d rather publish a gap than a plausible estimate. Citing a source is how you check us, though. It''s no guarantee we got it right.</p>'
    )),
    '{highlights,0}', jsonb_build_object(
      'title', 'We cite everything, and we still get things wrong',
      'description', 'Every factual claim about a named company names the document it came from and the date we read it, so you can go and look at the same thing we looked at. A source can be wrong. Our reading of it can be wrong. This site is put together with a lot of AI assistance, which will happily write a convincing sentence around a perfectly real citation. Check anything that matters against the document itself.')),
  updated_at = now()
FROM pages p WHERE p.id = pc.page_id
  AND p.site_id = (SELECT id FROM sites WHERE domain='oufe.com')
  AND p.name='about' AND pc.position=2;

-- about / 3 — the method block ------------------------------------------------
UPDATE page_components pc SET
  content_data = jsonb_set(
    jsonb_set(pc.content_data, '{heading}', to_jsonb('How we get things wrong'::text)),
    '{content}', to_jsonb(
      '<p>We cite everything. Where we give a figure we name the document it came from and the date we read it, so you can go and look at the same thing we looked at.</p>' ||
      '<p>A citation only shows you our source. It says nothing about whether we read it properly, picked the right passage, or whether the source itself is accurate and current. This site is assembled with a lot of AI assistance, and models invent detail fluently. The dangerous output is never the obviously wrong one. It''s the plausible sentence sitting next to a real link.</p>' ||
      '<p>So treat everything here as a worked example that might be inaccurate. We try hard to find real, current sources. You should still check anything that matters against the document itself before you rely on it.</p>' ||
      '<p>The same goes for the tools. They apply one arithmetic rule to the numbers you give them, and leave out most of what decides a real case. A tool here can be arithmetically correct and still describe a real situation wrongly.</p>' ||
      '<p>Where we haven''t verified something we say so, and we name the kind of document the figure will come from. A plausible estimate is no substitute for a sourced one. "We don''t know yet" is always publishable.</p>' ||
      '<p>Tell us if something here is wrong. We''ll correct it and say that we have.</p>'
    )),
  updated_at = now()
FROM pages p WHERE p.id = pc.page_id
  AND p.site_id = (SELECT id FROM sites WHERE domain='oufe.com')
  AND p.name='about' AND pc.position=3;

-- cases-index -----------------------------------------------------------------
UPDATE page_components pc SET
  content_data = jsonb_set(pc.content_data, '{content}', to_jsonb(
    '<p>A case here is a worked example rather than a verdict. We take a real situation, explain the mechanism being used, and show the arithmetic that follows from it. You should be able to apply the same reasoning to the next situation you meet.</p>' ||
    '<p>Treat every case on this site as a worked example that might be inaccurate. We try hard to find real, current sources, and we name the document behind each figure and the date we read it, so you can check us. That''s no guarantee we got it right. A source can be wrong, our reading of it can be wrong, and situations move faster than pages do.</p>' ||
    '<h3>In preparation</h3>' ||
    '<p><strong>Thames Water.</strong> A restructuring of a regulated utility, which makes it unusually well documented, because the regulator publishes and so does the court. We''re working through the primary documents now. Nothing is published here yet, because we don''t have the verified figures and we won''t estimate them.</p>' ||
    '<h3>What comes first</h3>' ||
    '<p>Ahead of the cases we''re writing the mechanism explainers they depend on. How a restructuring plan can bind a class that voted against it. How the creditor waterfall decides who gets paid and in what order. What the statutory framework actually permits. Those can be explained exactly, using figures that are openly made up, and they''re the part that carries over to every case rather than just one.</p>' ||
    '<p>If you want to see the arithmetic behind a waterfall before any of that is written, the recovery simulator lets you set a capital structure and watch which class the value runs out in.</p>'
  )),
  updated_at = now()
FROM pages p WHERE p.id = pc.page_id
  AND p.site_id = (SELECT id FROM sites WHERE domain='oufe.com')
  AND p.name='cases-index';

COMMIT;

-- Then re-render each page (content_data alone changes nothing a visitor sees):
--   ./docs/agent_docs/docs024_key_docs_latest/oufe/TRIGGER_rerender_page.sh <name> oufe.com
--
-- Verify on the RENDERED page, not the row — count what the owner counted:
--   em dashes, "is not / isn't / not a … but" constructions, and sentences
--   opening "It is" or "This is".
