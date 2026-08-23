-- 569 — repair the two live phantom-source fields whose fix needs no new machinery
-- (bugs_open/362, owner direction 2026-08-23: "do the six live ones next").
--
-- Two of the six are repairable as pure declaration fixes. The other four are NOT, and
-- this file deliberately does not touch them — see the FOOT of this comment.
--
-- ── 1. info-card-grid.carousel : source 'config' -> 'static' ──────────────────
-- `config` with NO DOT is outside the resolver's vocabulary: plan_sections' resolve()
-- splits on the first '.', and `len(parts) < 2` returns (nil,false). The field has been
-- dropped on all 32 live instances since the component was born (2026-03-31).
--
-- ⚠ THIS IS A BEHAVIOUR-IDENTICAL CHANGE, AND THAT IS THE POINT. Measured before
-- writing: 1 of the 32 instances carries `carousel: true` in its stored content_data.
-- It survives today because handleMissingField() calls carryStored() FIRST, which
-- re-reads the stored value (bugs_open/238's carry fix). Bare `static` resolves to
-- (nil, true) — found, but nil — and the field loop only assigns `if found && value !=
-- nil`, so it falls through to the SAME handleMissingField() and the SAME carry. The
-- one opt-in instance keeps its carousel; the other 31 keep no key. Nothing rendered
-- changes; what changes is that the declaration stops naming a source that cannot
-- resolve, so the component survives regeneration through the CLC-018 birth gate.
--
-- `static` is the honest vocabulary term here: the template's own comment says the flag
-- is "emitted only when content_data sets carousel" — i.e. it is set per instance at
-- render time, never fetched from a data source.
--
-- ── 2. Latest News Feed.insights_url : 'query.pages' -> 'pages.news' ──────────
-- `query.pages` names a query the resolver has never registered, so it errors at plan
-- time and the field is dropped on all 6 live instances. The field is used ONCE, inside
-- a <noscript> fallback: "Enable JavaScript to see the latest news{{if .insights_url}},
-- or visit <a href=...>our insights page</a>{{end}}".
--
-- WHY pages.news AND NOT llm. An LLM-authored URL here is precisely bugs_open/203's
-- class — a fabricated destination shipped inside an anchor. `pages.<name>` resolves a
-- LIVE page name to its real URL and, on a miss, returns (nil,false) with an explicit
-- "do NOT fabricate a URL" in the resolver, so the {{if}} simply drops the link.
--
-- MEASURED 2026-08-23, the six sites carrying this component: ai-agent-orchestration.com
-- and gaswholesalers.com have a page named exactly `news` (the link starts working);
-- idea.uk and robot-hands.com have `news-index`, relojistas.com and vetcomparison.uk
-- have neither (behaviour unchanged — no link, exactly as today). So this is a strict
-- improvement with no regression surface: 2 sites gain a real link, 4 keep the status
-- quo, and none can gain a fabricated one.
--
-- ── WHAT THIS FILE DELIBERATELY DOES NOT REPAIR ──────────────────────────────
--   featured_article (7 fields, query.featured_post, 3 live instances) and
--   category-listing (3 fields, query.category/query.category_posts, 2 live instances)
--     — both need queries the resolver does not register. That is a Go change in
--       queryresolve plus a migration, not a declaration fix, and it is sized in
--       bugs_open/362.
--   testimonials + social_proof (site_specs.social_proof.testimonials, 3 live instances)
--     — BLOCKED ON AN OWNER CONTENT DECISION, not on effort. No site_specs aspect holds
--       testimonial data, and the only cheap declaration fix (source: llm) would license
--       a model to write customer testimonials, which is the fabrication class this
--       estate polices. Detail in bugs_open/362.
--
-- APPLY BY HAND (the runner takes EVERY pending file; 560-568 are other sessions' or
-- already applied):
--   kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
--     psql -U clients_user -d clients_db -v ON_ERROR_STOP=1 \
--     < docs/agent_docs/sql_for_agents/569_repair_two_live_phantom_sources.sql
--   then: scripts/migration/run-migrations.sh --record-only 569_repair_two_live_phantom_sources.sql
--
-- ⚠ AFTER APPLYING, THE DAILY CHECK WILL GO RED, AND THAT IS CORRECT. The two repaired
-- entries become STALE baseline entries (they match nothing live any more). That is the
-- ratchet's pawl working. Trim those lines from
-- deployments/kustomize/services/component-source-vocabulary-check/base/component_source_baseline.json
-- and RE-APPLY the overlay, or the cluster keeps the old ConfigMap.

BEGIN;

CREATE TABLE IF NOT EXISTS bak_569_phantom_source_repair_20260823 AS
SELECT id, name, input_schema, now() AS backed_up_at
FROM content_components
WHERE is_active AND name IN ('info-card-grid','Latest News Feed');

-- ── PRE-STATE GUARD ───────────────────────────────────────────────────────────
DO $guard$
DECLARE n_carousel int; n_insights int;
BEGIN
  SELECT count(*) INTO n_carousel FROM content_components
  WHERE is_active AND name='info-card-grid'
    AND input_schema->'fields'->'carousel'->>'source' = 'config';
  IF n_carousel <> 1 THEN
    RAISE EXCEPTION '569 ABORT: expected exactly 1 info-card-grid with carousel source=config, found %.', n_carousel;
  END IF;

  SELECT count(*) INTO n_insights FROM content_components
  WHERE is_active AND name='Latest News Feed'
    AND input_schema->'fields'->'insights_url'->>'source' = 'query.pages';
  IF n_insights <> 1 THEN
    RAISE EXCEPTION '569 ABORT: expected exactly 1 Latest News Feed with insights_url source=query.pages, found %.', n_insights;
  END IF;
END
$guard$;

-- ── THE CHANGE ────────────────────────────────────────────────────────────────
UPDATE content_components
SET input_schema = jsonb_set(input_schema, '{fields,carousel,source}', '"static"'::jsonb, false)
WHERE is_active AND name='info-card-grid'
  AND input_schema->'fields'->'carousel'->>'source' = 'config';

UPDATE content_components
SET input_schema = jsonb_set(input_schema, '{fields,insights_url,source}', '"pages.news"'::jsonb, false)
WHERE is_active AND name='Latest News Feed'
  AND input_schema->'fields'->'insights_url'->>'source' = 'query.pages';

-- ── VERIFY (DO/RAISE — a block of SELECTs cannot stop the COMMIT) ─────────────
DO $verify$
DECLARE
  n_ok int;
  n_left int;
  n_fields_carousel int;
  n_fields_insights int;
BEGIN
  SELECT count(*) INTO n_ok FROM content_components
  WHERE is_active AND (
    (name='info-card-grid'   AND input_schema->'fields'->'carousel'->>'source'     = 'static') OR
    (name='Latest News Feed' AND input_schema->'fields'->'insights_url'->>'source' = 'pages.news'));
  IF n_ok <> 2 THEN
    RAISE EXCEPTION '569 VERIFY FAILED: expected 2 repaired declarations, found %.', n_ok;
  END IF;

  -- Neither dead source may survive anywhere in the active library.
  SELECT count(*) INTO n_left FROM content_components cc,
       LATERAL jsonb_each(CASE WHEN jsonb_typeof(cc.input_schema->'fields')='object'
                               THEN cc.input_schema->'fields' ELSE '{}'::jsonb END) AS f(k,v)
  WHERE cc.is_active AND v->>'source' IN ('config','query.pages');
  IF n_left <> 0 THEN
    RAISE EXCEPTION '569 VERIFY FAILED: % field(s) still declare config or query.pages.', n_left;
  END IF;

  -- The FIELDS themselves must still exist — a jsonb_set typo can silently create a
  -- sibling key and leave the original untouched, which would read as success above.
  SELECT count(*) INTO n_fields_carousel FROM content_components
  WHERE is_active AND name='info-card-grid' AND input_schema->'fields' ? 'carousel';
  SELECT count(*) INTO n_fields_insights FROM content_components
  WHERE is_active AND name='Latest News Feed' AND input_schema->'fields' ? 'insights_url';
  IF n_fields_carousel <> 1 OR n_fields_insights <> 1 THEN
    RAISE EXCEPTION '569 VERIFY FAILED: field presence is carousel=% insights_url=%, expected 1 and 1.',
      n_fields_carousel, n_fields_insights;
  END IF;

  RAISE NOTICE '569 OK: carousel -> static, insights_url -> pages.news; no config/query.pages sources remain.';
END
$verify$;

COMMIT;
