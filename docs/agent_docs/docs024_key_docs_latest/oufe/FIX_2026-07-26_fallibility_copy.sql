-- ============================================================================
-- oufe.com — replace the "everything is sourced" promise with the fallibility
-- posture. Owner direction, 2026-07-26.
--
-- WHY
--   The generated copy adopted, and strengthened, exactly the slogan the owner
--   struck: "Every factual claim is sourced to a named, dated primary document",
--   "A claim without a named, dated source does not appear here", "If we can't
--   trace a number to a document we hold, it doesn't appear here", "This
--   discipline is not a disclaimer. It's the method."
--
--   That is a promise about our reliability, and it is not one we can keep. A
--   citation proves provenance, not correctness: it cannot show we read the
--   source properly, chose the right passage, or that the source is right — and
--   it does nothing at all about a model writing a convincing sentence around a
--   real link, which is the failure this estate has actually suffered.
--
--   Replacement posture: we cite everything SO YOU CAN CHECK US, and we can
--   still be wrong. Citation as the instrument the reader catches us with,
--   rather than a warrant that we are right.
--
-- NOTE ON THE SITE'S OWN CLAIMS MACHINERY
--   None of this was caught by the claims gate, and it would not be: these are
--   claims about US, not unregistered numbers about a third party, and the
--   evidence register's banned patterns target fabrication shapes. A promise of
--   infallibility is a different failure class from an invented figure. Worth
--   recording — it is a gap in what the automated layer can see.
--
-- SAFE TO RE-RENDER AFTER THIS
--   Every section on both pages has non-NULL content_data (checked), so the
--   `section_data_resolved` rerender re-renders from stored content_data through
--   the current template with NO LLM call. A NULL content_data on any section
--   would escalate the whole page to the content writer and regenerate the copy.
--
-- Applied out of band. Site a0d7f1ae-f37e-4ea5-b30c-9012d1d14f39.
-- ============================================================================

\set ON_ERROR_STOP on

BEGIN;

-- about / 1 — hero subheadline -----------------------------------------------
UPDATE page_components pc SET
  content_data = jsonb_set(pc.content_data, '{subheadline}', to_jsonb(
    'OUFE analyses the mechanics of UK restructuring: the legal frameworks, the creditor rankings, the arithmetic of who recovers what and why. We cite everything so that you can check it — and we can still be wrong, because a source can be wrong and our reading of it can be wrong. Treat what you find here as a worked example rather than an authority. This site is not investment advice and does not offer any.'::text)),
  updated_at = now()
FROM pages p
WHERE p.id = pc.page_id
  AND p.site_id = (SELECT id FROM sites WHERE domain='oufe.com')
  AND p.name = 'about' AND pc.position = 1;

-- about / 2 — first highlight becomes the fallibility statement ---------------
UPDATE page_components pc SET
  content_data = jsonb_set(pc.content_data, '{highlights,0}', jsonb_build_object(
    'title', 'We cite everything, and we can still be wrong',
    'description', 'Every factual claim about a named company names the document it came from and the date we read it. That is there so you can check us; it does not make us right. A source can be wrong, we can read it wrongly, and this site is assembled with AI assistance that can produce a convincing sentence around a perfectly real citation. Check anything that matters against the document itself.')),
  updated_at = now()
FROM pages p
WHERE p.id = pc.page_id
  AND p.site_id = (SELECT id FROM sites WHERE domain='oufe.com')
  AND p.name = 'about' AND pc.position = 2;

-- about / 3 — the method block, rewritten -------------------------------------
UPDATE page_components pc SET
  content_data = jsonb_set(
    jsonb_set(pc.content_data, '{heading}', to_jsonb('How we can be wrong'::text)),
    '{content}', to_jsonb(
      '<p>We cite everything. Where we give a figure, we name the document it came from and the date we read it, so you can go and look at the same thing we looked at.</p>' ||
      '<p><strong>That does not make us right.</strong> A citation shows you our source. It does not prove that we read it properly, that we picked the right passage, or that the source itself is accurate and current. This site is assembled with substantial AI assistance, and models invent detail fluently — the dangerous output is never the obviously wrong one, it is the plausible sentence sitting next to a real link.</p>' ||
      '<p>So please treat everything here as a possibly inaccurate worked example. We try hard to source real and current data, and you should still check anything that matters against the primary document before you rely on it. The citation is not a warranty. It is the instrument you use to catch us.</p>' ||
      '<p>The same goes for the interactive tools. They apply one arithmetic rule to the numbers you give them, and they leave out most of what decides a real case — security, guarantees, structural subordination, intercompany claims, contested valuation. A tool here can be arithmetically correct and still describe a real situation wrongly.</p>' ||
      '<p>Where we have not verified something, we say so and name the kind of document the figure will come from. A plausible estimate is not a substitute for a sourced figure, and "we do not know yet" is always publishable.</p>' ||
      '<p>If something here is wrong, tell us. We will correct it and say that we have.</p>'
    )),
  updated_at = now()
FROM pages p
WHERE p.id = pc.page_id
  AND p.site_id = (SELECT id FROM sites WHERE domain='oufe.com')
  AND p.name = 'about' AND pc.position = 3;

-- index / 1 — hero subheadline ------------------------------------------------
UPDATE page_components pc SET
  content_data = jsonb_set(pc.content_data, '{subheadline}', to_jsonb(
    'OUFE examines how distressed capital structures actually work: the statutory framework, the creditor waterfall, and the arithmetic behind the outcome. We name our sources so that you can check them, and we can still be wrong — so treat this as a worked example rather than an authority. Analysis of mechanism. Not investment advice, and not a substitute for legal advice on a specific matter.'::text)),
  updated_at = now()
FROM pages p
WHERE p.id = pc.page_id
  AND p.site_id = (SELECT id FROM sites WHERE domain='oufe.com')
  AND p.name = 'index' AND pc.position = 1;

-- index / 4 — closing call-to-action subheadline ------------------------------
UPDATE page_components pc SET
  content_data = jsonb_set(pc.content_data, '{subheadline}', to_jsonb(
    'This analysis works through how corporate finance structures fail under stress: the statutory framework, the creditor waterfall, and the arithmetic that determines who recovers what. We name the document behind each figure and the date we read it, so that you can check us — because we do make mistakes, and so do the sources. Where we have not verified something yet, we say so.'::text)),
  updated_at = now()
FROM pages p
WHERE p.id = pc.page_id
  AND p.site_id = (SELECT id FROM sites WHERE domain='oufe.com')
  AND p.name = 'index' AND pc.position = 4;

COMMIT;

-- Verify no survivor of the retired promise in content_data:
--   SELECT p.name, pc.position FROM page_components pc JOIN pages p ON p.id=pc.page_id
--    WHERE p.site_id=(SELECT id FROM sites WHERE domain='oufe.com')
--      AND pc.content_data::text ~* 'does not appear here|traceable to a primary document|is the method';
--
-- Then re-render each page (assemble-only will NOT pick this up — the stored
-- rendered_html must be regenerated from content_data):
--   ./docs/agent_docs/docs024_key_docs_latest/cta_link_integrity/scripts/049b_deploy_single_page.sh \
--     <page_id> a0d7f1ae-f37e-4ea5-b30c-9012d1d14f39 oufe.com section_data_resolved
