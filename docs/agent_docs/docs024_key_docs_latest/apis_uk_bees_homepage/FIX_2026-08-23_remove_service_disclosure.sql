-- ============================================================================
-- apis.uk — REMOVE THE OTHER-SERVICE DISCLOSURE, AND CLOSE THE DOOR ON IT
-- Written 2026-08-23 after the owner found it live on the site.
--
-- WHAT WENT WRONG. This is not a framework failure. The framework did exactly
-- what it was told: the 2026-08-22 mission brief asked for "a single short,
-- plainly worded line ... acknowledging that the domain also hosts an unrelated
-- technical service ... so that a developer who lands here by mistake is not
-- confused", and roadmap_brief repeated it. The writer complied — FOUR times,
-- once per section, because "somewhere unobtrusive" names no single place.
--
-- WHY IT IS SERIOUS. It publishes, on a public page, the fact that this domain
-- carries a technical service on another hostname. Nobody asked for that
-- disclosure, it serves no visitor, and a personal page about an insect has no
-- business advertising infrastructure. The instruction was mine and the
-- judgement behind it was wrong: a developer arriving at a bees page by mistake
-- is not a problem worth solving by telling every visitor the service exists.
--
-- ORDER MATTERS. The specs are fixed BEFORE the bytes, because the specs are
-- what a regeneration reads. Strip the sentences first and the next rerender
-- writes them straight back.
--
-- FAIL-CLOSED, NOT JUST CORRECTED. Removing the instruction stops the writer
-- being ASKED for it; the banned_claims patterns stop it being ACCEPTED if it
-- is ever written again for any other reason. Both, because an instruction
-- removed from a prompt is a decision no future reader can see, and a ban is.
-- ============================================================================

\set ON_ERROR_STOP on

BEGIN;

-- ------------------------------------------------- 1. evidence_base: bans --
CREATE TEMP TABLE _eb AS
SELECT * FROM site_specs
WHERE site_id = (SELECT id FROM sites WHERE domain='apis.uk')
  AND aspect = 'evidence_base' AND is_current;

UPDATE site_specs SET is_current = false, superseded_at = now()
WHERE id IN (SELECT id FROM _eb);

INSERT INTO site_specs (site_id, aspect, data, source, notes, is_current, pinned, created_by)
SELECT
  site_id, 'evidence_base',
  jsonb_set(
    jsonb_set(data, '{banned_claims}',
      (data->'banned_claims') || $bans$[
        {"pattern": "unrelated (technical |other )?(service|system|application|platform|tool|api)",
         "reason": "OTHER-SERVICE DISCLOSURE class. Shipped live 2026-08-22 in four sections and caught by the owner. This page must not mention, allude to, or explain the existence of anything else running on this domain."},
        {"pattern": "technical service",
         "reason": "same class, the exact phrase the writer used. Kept as its own literal so the ban is obvious to a reader of this list."},
        {"pattern": "(different|another|separate|other|second) (hostname|host name|subdomain|sub-domain|address on)",
         "reason": "same class, the mechanism half of the sentence. The disclosure is recognisable by its shape: this-domain-also + a different-host clause."},
        {"pattern": "(this|the) (domain|address|site|name) (also )?(hosts|carries|runs|serves|lends|is home to|is also)",
         "reason": "same class, the possessive half. Catches 'this domain also hosts', 'this address also carries', 'apis.uk is also the address of'."},
        {"pattern": "(nothing|no connection) (here |on this page )?(to do with|with) (it|that|the service)",
         "reason": "same class, the disclaimer half — the sentence that follows the disclosure and draws further attention to it."}
      ]$bans$::jsonb),
    '{writer_block}',
    to_jsonb((data->>'writer_block') || E'\n\nAND ONE ABSOLUTE PROHIBITION, ADDED 2026-08-23 BECAUSE IT WAS GOT WRONG ONCE ALREADY.\n\nDo not mention, hint at, allude to, explain, disclaim or apologise for anything else that runs on this domain. Not in the body, not in a footer, not in an aside, not in a single quiet sentence at the end of a section, and not as a kindness to some imagined visitor who has arrived at the wrong place. There is no such visitor to serve, and naming infrastructure on a public page about an insect is a disclosure nobody asked for.\n\nThis page is about bees. That is the whole of its subject. If you find yourself writing a sentence whose purpose is to tell the reader what this page is NOT about, delete the sentence: it is not orienting anybody, and the reader who needed it does not exist.')
  ),
  source,
  'Superseded 2026-08-23: added the OTHER-SERVICE DISCLOSURE ban class (5 patterns) and an absolute prohibition in writer_block, after the 2026-08-22 brief asked for exactly this sentence and the writer shipped it live in four sections. facts[] still deliberately empty.',
  true, true, 'apis-uk-bees-2026-08-23'
FROM _eb;

-- --------------------------------------- 2. roadmap_brief: drop the ask ----
UPDATE site_specs SET is_current = false, superseded_at = now()
WHERE site_id = (SELECT id FROM sites WHERE domain='apis.uk')
  AND aspect = 'roadmap_brief' AND is_current;

INSERT INTO site_specs (site_id, aspect, data, source, notes, is_current, pinned, created_by)
SELECT
  id, 'roadmap_brief',
  $rb${
    "text": "PHASE 1 — AND PHASE 1 IS THE WHOLE OF THIS SITE FOR NOW. Build exactly ONE page and no others: the home page, at the site root, page_type index. Do not plan, propose or create an about page, a contact page, a guides index, a blog, a glossary, a species directory, a tools page, a legal page or a privacy page. Do not add navigation items pointing at pages that do not exist. If a subject seems to deserve its own page, it belongs in a section of the home page instead, or it waits for a later phase the owner has not yet asked for. The home page is a single scrolling page about bees, written for a general visitor who arrived out of curiosity: a hero that says plainly what this page is, then a small number of content sections that each take one genuinely interesting aspect of bees and explain it well in plain prose, and a quiet footer. No pricing, no offer, no signup, no lead capture, no testimonials, no statistics band, no client logos, no call-to-action urging the visitor to do anything commercial. The page is about bees and about nothing else whatsoever: it must not mention, allude to, disclaim or explain anything else that runs on this domain, in any section, in the footer, or in a closing aside. A sentence whose job is to tell the reader what this page is not about does not belong on it."
  }$rb$::jsonb,
  'manual',
  'Superseded 2026-08-23. The previous version ASKED for a line acknowledging an unrelated technical service on another hostname; the writer obliged in four sections and it went live. That clause is removed and replaced with an explicit prohibition. Page-count constraint unchanged.',
  true, true, 'apis-uk-bees-2026-08-23'
FROM sites WHERE domain = 'apis.uk';

COMMIT;

-- Verify
--   SELECT jsonb_array_length(data->'banned_claims') FROM site_specs ss JOIN sites s ON s.id=ss.site_id
--    WHERE s.domain='apis.uk' AND aspect='evidence_base' AND is_current;   -- expect 32
--   SELECT data->>'text' ILIKE '%must not mention%' FROM site_specs ss JOIN sites s ON s.id=ss.site_id
--    WHERE s.domain='apis.uk' AND aspect='roadmap_brief' AND is_current;   -- expect t
