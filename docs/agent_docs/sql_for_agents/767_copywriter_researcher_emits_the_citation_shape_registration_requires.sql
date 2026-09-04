-- 767 — fix the copywriter researcher's output contract. 765's prompt asked the model for
-- `quote` + `source_url` and no publisher; `datahelpers.citations.go` requires **url**, **quote**
-- and **publisher** on every claim and rejects the rest. First live run (2026-09-04 13:38:48Z,
-- COMPLETED) extracted a real organisation — `stratton-craig` — and registration rejected every
-- claim with `class=citation_invalid, detail="citation missing required field(s): url, publisher"`.
-- Nothing was wrong with the research; the prompt named the wrong keys.
--
-- The corrected tail is taken from the SIBLING THAT WORKS (finance-directory-researcher's own
-- output contract, live row), so the two cannot drift on the part registration actually parses:
-- flat `quote` / `url` / `publisher` / `title` / `published`, plus `staleness_days` so the
-- re-verification sweep knows how long each field stays true.
--
-- WHY THE FAILURE WAS INVISIBLE UNTIL A RUN: the workflow completed, the step succeeded, the
-- model obeyed its instructions exactly, and `registration.candidates` read 0 with the rejects
-- in `collected_data`. A status-level check would call this a clean run. The artefact is the
-- `rejected` array — read it, not the status.

BEGIN;

DO $g$
DECLARE p text;
BEGIN
  SELECT default_config #>> '{workflow,steps,extract_claims,config,prompt_template}' INTO p
    FROM agent_definitions WHERE type='copywriter-directory-researcher'
     AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF p IS NULL THEN RAISE EXCEPTION '767 REFUSED: copywriter-directory-researcher not found (apply 765)'; END IF;
  IF position('source_url' in p) = 0 THEN RAISE EXCEPTION '767 REFUSED: prompt does not carry the defect (already fixed?)'; END IF;
  IF position('UK COPYWRITING ORGANISATIONS' in p) = 0 THEN RAISE EXCEPTION '767 REFUSED: this is not the copywriter prompt'; END IF;
END $g$;

UPDATE agent_definitions
   SET default_config = jsonb_set(default_config, '{workflow,steps,extract_claims,config,prompt_template}',
     to_jsonb($p$You are extracting ATOMIC, CITABLE facts about UK COPYWRITING ORGANISATIONS for a public cited directory.

Research question: {{.input_data.research_query}}

SCRAPED SOURCES (each has a url):
{{.scrape_results}}

Extract up to 15 candidate claims, one per (organisation, fact) pair. At most ONE claim per (organisation, field) pair — if several passages state the same field, emit only the single most complete enumeration: a later duplicate OVERWRITES the earlier one at registration, so a weaker duplicate destroys a stronger claim.

TWO HARD RULES. A claim that breaks either is worthless and must not be emitted.

1. AN ENTITY IS ONE NAMED ORGANISATION WITH ITS OWN PUBLIC WEBSITE — an agency, studio, consultancy or marketplace. NEVER a named individual or a sole trader trading under their own personal name; never a sector, a market segment, a "type of provider", a ranking list, or an article about copywriting. If the source is a listicle ("the 10 best copywriting agencies"), extract the ORGANISATIONS it names, never the list itself. If you cannot name the organisation and point at its own website, emit nothing for it.

2. EVERY CLAIM CARRIES A VERBATIM QUOTE FROM THE SOURCE THAT STATES IT, plus that source's url and publisher. No inference, no summary in place of a quote, no fact assembled from two passages.

FIELDS — use ONLY these names:
- specialisms — what kinds of copy the organisation writes (web, ads, email, product, tone-of-voice, technical, B2B), the source's own enumeration, never your synthesis across pages
- sectors_served — the industries it serves, likewise the source's own enumeration
- supplier_type — EXACTLY ONE OF "agency", "studio", "consultancy", "marketplace"
- location — the UK town or city it is based in
- established_year — four digits, only if the source states it explicitly
- team_size — only if the source states a number or a stated range
- website — the organisation's own site

NEVER extract prices, day rates or per-word rates — they go stale and a stale price under a named firm misrepresents it. NEVER extract testimonials, awards claimed without a named awarding body, or marketing superlatives ("the best", "leading", "award-winning"): they are not facts.

NEVER COPY A VALUE FROM THESE INSTRUCTIONS — they illustrate shape, not data.

Optionally, on the FIRST claim you emit for a given entity, also include entity_summary (one plain sentence: what the organisation is), entity_links (object: docs [its own about/services page]), and entity_attributes (object: supplier_type, region — coarse filing tags, not claims). Omit rather than guess.

Return ONLY a JSON array (no commentary). Each element:
{"entity_kind": "copywriter", "entity_slug": "lowercase-hyphenated", "entity_name": "Display Name", "entity_owner": "parent group if the source names one, else same as entity_name", "entity_summary": "..." (optional), "entity_links": {...} (optional), "entity_attributes": {...} (optional), "field": "...", "value": "...", "quote": "verbatim sentence from the source", "url": "the source page's url", "publisher": "the site or organisation that published that page", "title": "the source page's title", "published": "YYYY-MM or YYYY if shown", "staleness_days": <400 for established_year, location, supplier_type and website — structural facts; 200 for specialisms, sectors_served and team_size, which drift>}

If nothing extractable meets the bar, return [].$p$::text)),
       updated_at = NOW()
 WHERE type='copywriter-directory-researcher' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

DO $v$
DECLARE p text;
BEGIN
  SELECT default_config #>> '{workflow,steps,extract_claims,config,prompt_template}' INTO p
    FROM agent_definitions WHERE type='copywriter-directory-researcher'
     AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF position('source_url' in p) > 0 THEN RAISE EXCEPTION '767 VERIFY: the old key survived'; END IF;
  IF position('"publisher":' in p) = 0 THEN RAISE EXCEPTION '767 VERIFY: publisher not named in the output shape'; END IF;
  IF position('"url":' in p) = 0 THEN RAISE EXCEPTION '767 VERIFY: url not named in the output shape'; END IF;
  IF position('"quote":' in p) = 0 THEN RAISE EXCEPTION '767 VERIFY: quote not named in the output shape'; END IF;
  IF position('"entity_kind": "copywriter"' in p) = 0 THEN RAISE EXCEPTION '767 VERIFY: entity_kind literal lost'; END IF;
  IF position('NEVER a named individual' in p) = 0 THEN RAISE EXCEPTION '767 VERIFY: the organisations-only rule was dropped'; END IF;
END $v$;

COMMIT;
