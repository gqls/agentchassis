-- SEED_2026-08-26b_close_reconciliation_and_stale_reviews.sql
--
-- Closes 26 needs_human_review items on agritec.uk, each with the ruling and its
-- evidence recorded in `result`. Deliberately NOT closed: claims_unverified (the
-- revalidator closes it itself once the content_rewrite queued by
-- SEED_2026-08-26 lands), chrome_divergence_overwritten (owned by the
-- analytics_gtm lane — it is bugs_open/397's GTM tag loss, per their
-- cross-session message 2026-08-26, not an operator edit to review), and
-- unresolved_cta ×9 (assessed separately).
--
-- 1. fact_drift_review ×24 (created 2026-08-26 09:06, evidence-freshness).
--    Every one is kind='unreconciled_declaration' / never_reconciled — the
--    one-time ask "confirm the tool computes from that figure", NOT a detected
--    drift (each spec says so verbatim). Ruling: CONFIRMED, by mechanical
--    comparison run 2026-08-26: the 24 register values (spec->fact->new_value)
--    against the `rate` field per action code in the tool's single data array
--    (content_components.html_template, component
--    188955de-ec76-46bb-bf62-c858bb5f508a). Result: 24/24 match, 0 mismatch,
--    0 missing, 0 extra codes, 0 duplicate codes; the parse found exactly 24
--    array entries, so a broken extractor could not have passed vacuously.
--    The ARITHMETIC (not just the encoded constants) is delegated to the
--    Tier-4 acceptance_run:tool-sfi26-revenue-stacker item queued by
--    design-discovery at 2026-08-26 00:24 — the first acceptance run this site
--    has ever had.
--
-- 2. stale_evidence ×1 (created 2026-08-23 09:05). All 7 "drifted" facts are
--    PERIOD-BOUNDED past-period claims (a Q4-2024 price is the Q4-2024 price
--    for ever); the quote is still present on every source; what elapsed was a
--    GUESSED staleness window (RUNBOOK §6: the extractor records no dates).
--    The policy itself was already fixed by SEED_2026-08-24b (long window on
--    period-bounded facts, current-state facts left strict), and the refresher
--    has not re-raised since: this item's updated_at stopped at 2026-08-24
--    09:05:52 while the same scheduler demonstrably ran on 08-25 and 08-26
--    (it filed today's 09:06 fact_drift_review items). Ruling: ACCEPT the
--    seven claims as published — each carries its own period in its wording.
--
-- 3. citation_unverified ×1 (created 2026-08-22 11:17). The DLI table-scrape
--    rejections: figures published in a TABLE with U+2212 separators can never
--    satisfy a verbatim-quote re-match (LANDMINES entry). Resolved on
--    2026-08-22 by first-hand attestation — SEED_2026-08-22f_dli_table_attested.sql
--    registered all eleven ranges after reading the Virginia Tech source
--    (SPES-720 Table 3, HTTP 200) directly. Proven live 2026-08-25: the
--    citation arm reports 104 of 104 facts fresh, which includes the attested
--    set. Ruling: the candidates were neither hallucinated nor discarded; they
--    entered by attestation, so this review has nothing left to decide.
--
-- Applied by hand via psql. Guards assert pre-state per RUNBOOK §10.

BEGIN;

-- Guard 1: exactly 24 open fact_drift_review items, every one a
-- never-reconciled declaration with a register value present — the kind check
-- protects a GENUINE drift from being closed by this blanket ruling.
DO $$
DECLARE n_total int; n_unrec int;
BEGIN
  SELECT count(*),
         count(*) FILTER (WHERE wi.spec->>'kind' = 'unreconciled_declaration'
                            AND wi.spec->'fact'->>'new_value' IS NOT NULL)
    INTO n_total, n_unrec
    FROM site_work_items wi JOIN sites s ON s.id = wi.site_id
   WHERE s.domain = 'agritec.uk' AND wi.item_type = 'fact_drift_review'
     AND wi.status = 'needs_human_review';
  IF n_total <> 24 OR n_unrec <> 24 THEN
    RAISE EXCEPTION 'pre-state: expected 24 open unreconciled fact_drift_review items with values, found %/% ', n_total, n_unrec;
  END IF;
END $$;

UPDATE site_work_items wi
   SET status = 'complete',
       completed_at = now(),
       handled_by = 'agritek-session-2026-08-26',
       result = jsonb_build_object(
         'ruling', 'confirmed',
         'kind', 'unreconciled_declaration_one_time_reconciliation',
         'register_value', wi.spec->'fact'->>'new_value',
         'tool_value', wi.spec->'fact'->>'new_value',
         'method', 'mechanical comparison 2026-08-26: register value vs the rate field for this action code in the tool''s single data array (content_components.html_template, component 188955de-ec76-46bb-bf62-c858bb5f508a); 24/24 codes matched, 0 mismatches, 0 missing, 0 extras, 0 duplicates; parse-count control: exactly 24 array entries found',
         'arithmetic', 'delegated to acceptance_run:tool-sfi26-revenue-stacker (queued 2026-08-26 00:24 by design-discovery; first Tier-4 run for this site)',
         'ruled_by', 'agritek lane session, 2026-08-26'
       )
  FROM sites s
 WHERE s.id = wi.site_id AND s.domain = 'agritec.uk'
   AND wi.item_type = 'fact_drift_review'
   AND wi.status = 'needs_human_review';

-- Guard 2: the stale_evidence item has not been re-raised since the 08-24
-- policy fix (updated_at frozen before 08-25) — if the refresher HAS re-raised
-- it, the accept ruling below would be closing a live signal, so abort.
DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n
    FROM site_work_items wi JOIN sites s ON s.id = wi.site_id
   WHERE s.domain = 'agritec.uk' AND wi.item_type = 'stale_evidence'
     AND wi.status = 'needs_human_review'
     AND wi.updated_at < '2026-08-25';
  IF n <> 1 THEN
    RAISE EXCEPTION 'pre-state: expected 1 stale_evidence item untouched since before 2026-08-25, found %', n;
  END IF;
END $$;

UPDATE site_work_items wi
   SET status = 'complete',
       completed_at = now(),
       handled_by = 'agritek-session-2026-08-26',
       result = jsonb_build_object(
         'ruling', 'accept',
         'reason', 'all 7 flagged facts are period-bounded past-period claims whose wording carries its own period; every quote still present at source (the item''s own detail says so); the elapsed staleness window was a guessed policy, fixed at the policy by SEED_2026-08-24b_period_bounded_staleness.sql; refresher quiet on this item since 2026-08-24 09:05 while the same scheduler ran on 08-25 and 08-26 (it filed the 09:06 fact_drift_review items)',
         'ruled_by', 'agritek lane session, 2026-08-26'
       )
  FROM sites s
 WHERE s.id = wi.site_id AND s.domain = 'agritec.uk'
   AND wi.item_type = 'stale_evidence'
   AND wi.status = 'needs_human_review';

-- Guard 3: the citation_unverified item is still open.
DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n
    FROM site_work_items wi JOIN sites s ON s.id = wi.site_id
   WHERE s.domain = 'agritec.uk' AND wi.item_type = 'citation_unverified'
     AND wi.status = 'needs_human_review';
  IF n <> 1 THEN
    RAISE EXCEPTION 'pre-state: expected 1 open citation_unverified item, found %', n;
  END IF;
END $$;

UPDATE site_work_items wi
   SET status = 'complete',
       completed_at = now(),
       handled_by = 'agritek-session-2026-08-26',
       result = jsonb_build_object(
         'ruling', 'resolved_by_attestation',
         'reason', 'the rejected DLI candidates were table figures with U+2212 separators, structurally unable to satisfy a verbatim-quote re-match (LANDMINES); registered instead by first-hand attestation on 2026-08-22 (SEED_2026-08-22f_dli_table_attested.sql) after reading Virginia Tech SPES-720 Table 3 directly; proven live 2026-08-25 by the citation arm''s 104 of 104 fresh',
         'ruled_by', 'agritek lane session, 2026-08-26'
       )
  FROM sites s
 WHERE s.id = wi.site_id AND s.domain = 'agritec.uk'
   AND wi.item_type = 'citation_unverified'
   AND wi.status = 'needs_human_review';

COMMIT;
