-- 675_d4_governor_admits_function.sql — D4 stage B, revision per council r1 (corr 8f4bb57d).
-- THE ARCHITECTURE SEAT'S FIX, ADOPTED: the shed predicate becomes ONE canonical DB function,
-- `governor_admits(item_type)`, and every consumer (Go loader, Go claim backstop, the
-- selector's config text) emits a one-line call — the four-hand-copies byte-lockstep problem
-- the r1 round objected to ceases to exist rather than being guarded.
-- ALSO the bug_historian's observability fix: `governor_withheld_now`, a VIEW answering
-- "which items is the governor withholding right now, and why" — shedding is a computed
-- predicate, so its observable is a view, not per-item stamps; every existing
-- why-is-this-stuck query gains a one-join discriminator.
--
-- INERT ON APPLY three ways, same as all of stage A: nothing calls the function until the
-- held 674 applies (post-roll); governor_config.enabled=false makes it return true for every
-- row; monthly_budget_usd NULL keeps shed_level 0 regardless.
--
-- The three posture rules (unchanged from the r1 Go renderer, now in the ONE place):
--   1. FAIL-OPEN: an unreadable governor admits everything (outer NOT COALESCE(...,false)
--      inverted here — the function RETURNS TRUE on missing config/state rows).
--   2. An UNMAPPED item_type = maintenance + llm_bearing: sheds earliest.
--   3. enabled=false short-circuits to TRUE per row: disabled governor = identity.
-- Each rule is EXECUTED against synthetic states in the verify below — probes, not
-- string assertions (the r1 tests' assertions, made runnable).
-- Rollback: 675_..._ROLLBACK.sql (drops view then function; refuses while any live agent
-- config references governor_admits).

BEGIN;

DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM pg_proc WHERE proname='governor_admits') THEN
    RAISE EXCEPTION '675 REFUSED: governor_admits() already exists (replay).';
  END IF;
  IF NOT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name='governor_work_class_map') THEN
    RAISE EXCEPTION '675 REFUSED: stage A (671) not applied — governor_work_class_map missing.';
  END IF;
END $$;

CREATE FUNCTION governor_admits(p_item_type text) RETURNS boolean
LANGUAGE sql STABLE AS $FN$
  SELECT NOT COALESCE((
    SELECT gc.enabled
       AND COALESCE(m.llm_bearing, true)
       AND gs.shed_level >= CASE COALESCE(m.class, 'maintenance')
             WHEN 'maintenance' THEN 1
             WHEN 'build'       THEN 2
             ELSE                    3
           END
    FROM governor_config gc
    JOIN governor_state gs ON gs.id = 1
    LEFT JOIN governor_work_class_map m ON m.item_type = p_item_type
    WHERE gc.id = 1
  ), false)
$FN$;

COMMENT ON FUNCTION governor_admits(text) IS
'D4 spend governor (AGOV-013): TRUE unless the governor currently withholds this item_type.
The ONE canonical shed predicate — the Go loader/claim and the dispatch selector all call it;
do not re-spell the logic anywhere (council corr 8f4bb57d r1, architecture seat). Fail-open:
missing config/state rows admit everything. Unmapped types = maintenance+llm_bearing.';

-- The observability half (bug_historian, r1 gating objection): what is withheld RIGHT NOW,
-- and why — so "governor working as designed" is distinguishable from "dispatch broken"
-- with one query, during the event, by anyone.
CREATE VIEW governor_withheld_now AS
SELECT wi.id, wi.site_id, s.domain, wi.item_type,
       COALESCE(m.class, 'maintenance')     AS class,
       COALESCE(m.llm_bearing, true)        AS llm_bearing,
       gs.shed_level                        AS current_shed_level,
       wi.created_at, wi.priority, wi.status
FROM site_work_items wi
JOIN sites s ON s.id = wi.site_id
LEFT JOIN governor_work_class_map m ON m.item_type = wi.item_type
CROSS JOIN governor_state gs
WHERE wi.status IN ('triaged','approved')
  AND NOT governor_admits(wi.item_type);

COMMENT ON VIEW governor_withheld_now IS
'Items the spend governor is withholding at this instant (empty whenever the governor is
disabled or at level 0). THE first query during any shed event: rows here are working-as-
designed, not stuck. AGOV-013.';

-- Posture probes: each rule DRIVEN against synthetic state inside this transaction and
-- restored; a wrong function body cannot COMMIT.
DO $$
DECLARE r boolean;
BEGIN
  -- Baseline (disabled, budget NULL, level 0): admits everything, mapped or not.
  IF NOT governor_admits('content_rewrite') OR NOT governor_admits('never_seen_type_xyz') THEN
    RAISE EXCEPTION '675 VERIFY: disabled governor must admit everything';
  END IF;

  -- Rule 3 + shed order: enable at level 1 — maintenance sheds, build/research/llm-free stay.
  UPDATE governor_config SET enabled=true WHERE id=1;
  UPDATE governor_state  SET shed_level=1 WHERE id=1;
  IF governor_admits('content_rewrite') THEN
    RAISE EXCEPTION '675 VERIFY: L1 must shed llm-bearing maintenance (content_rewrite)';
  END IF;
  IF governor_admits('never_seen_type_xyz') THEN
    RAISE EXCEPTION '675 VERIFY: L1 must shed an UNMAPPED type (defaults maintenance+bearing)';
  END IF;
  IF NOT governor_admits('page_rerender') THEN
    RAISE EXCEPTION '675 VERIFY: llm-free maintenance must NEVER shed';
  END IF;
  IF NOT governor_admits('needs_page') OR NOT governor_admits('needs_vertical_research') THEN
    RAISE EXCEPTION '675 VERIFY: L1 must not touch build or research';
  END IF;

  UPDATE governor_state SET shed_level=2 WHERE id=1;
  IF governor_admits('needs_page') OR NOT governor_admits('needs_vertical_research') THEN
    RAISE EXCEPTION '675 VERIFY: L2 = maintenance+build shed, research protected';
  END IF;

  UPDATE governor_state SET shed_level=3 WHERE id=1;
  IF governor_admits('needs_vertical_research') OR NOT governor_admits('page_rerender') THEN
    RAISE EXCEPTION '675 VERIFY: L3 sheds research too; llm-free still exempt';
  END IF;

  -- Rule 3 again from the shed state: disabling restores identity at any level.
  UPDATE governor_config SET enabled=false WHERE id=1;
  IF NOT governor_admits('content_rewrite') THEN
    RAISE EXCEPTION '675 VERIFY: disabling must admit everything even at level 3';
  END IF;

  -- Rule 1, fail-open: with the state row GONE, the function must admit, not error/shed.
  DELETE FROM governor_state WHERE id=1;
  IF NOT governor_admits('content_rewrite') THEN
    RAISE EXCEPTION '675 VERIFY: missing governor_state must FAIL OPEN (admit)';
  END IF;
  INSERT INTO governor_state (id, shed_level) VALUES (1, 0);

  -- Restore the exact shipped-inert state and prove it.
  UPDATE governor_state SET shed_level=0, month=NULL, mtd_usd=NULL, unpriced_io_tokens=NULL, computed_at=NULL WHERE id=1;
  PERFORM 1 FROM governor_config WHERE id=1 AND enabled=false AND monthly_budget_usd IS NULL;
  IF NOT FOUND THEN RAISE EXCEPTION '675 VERIFY: config not restored to shipped-inert'; END IF;
  SELECT count(*)=0 INTO r FROM governor_withheld_now;
  IF NOT r THEN RAISE EXCEPTION '675 VERIFY: withheld view must be empty at level 0'; END IF;

  RAISE NOTICE '675 OK: governor_admits() is the one canonical predicate — all three posture rules and the full shed order PROVEN by execution; withheld-now view live and empty; state restored inert. (The 120s task will repopulate the meter fields within 2 min.)';
END $$;

COMMIT;
