-- 696_lendzy_correct_two_fca_rule_citations_ROLLBACK.sql
--
-- Restores the two WRONG rule citations. Say that plainly: rolling back puts a
-- known-wrong legal reference back on two live finance sites. The only reason to
-- run this is that 696 itself broke something worse.
--
-- ⚠ THE REVERSE IS ENUMERATED AND ORDERED, AND CANNOT BE A BLANKET REPLACE:
--   * order: 6.7.23 -> 6.7.17 must run FIRST (while every remaining 6.7.23 is a
--     rollover cite), then 7.6.12 -> 6.7.23 — the mirror of the forward order.
--   * enumeration: the tool-cpa-cancellation-checker template and its page
--     carried a CORRECT 'CONC 7.6.12' BEFORE 696 (census 2026-09-02). A blanket
--     7.6.12 -> 6.7.23 would corrupt that pre-existing correct citation, so the
--     CPA reverse is scoped to the rows 696 actually changed, by name.
-- ⚠ If the queued rerenders have already run, deployed artefacts carry the
--   corrected numbers; rolling the store back re-creates the store/artefact
--   split this lane exists to close. Check the rerender items first.

BEGIN;

-- 0. Withdraw 696's rerenders still waiting. Claimed/complete ones are left.
UPDATE site_work_items SET status='cancelled', updated_at=NOW()
 WHERE source='lendzy_co_uk lane (migration 696)' AND item_type='page_rerender' AND status='triaged';

-- 1. Rollover reverse FIRST (all remaining 6.7.23s are rollover cites).
UPDATE page_components pc
   SET content_data = replace(pc.content_data::text,'CONC 6.7.23','CONC 6.7.17')::jsonb, updated_at=NOW()
  FROM pages p
 WHERE pc.page_id=p.id AND p.site_id='8ff093d5-1f19-453b-9439-a10379bbcd76'
   AND p.name IN ('rollover-rules','tool-rollover-limit-checker-guide')
   AND pc.content_data::text LIKE '%CONC 6.7.23%';

UPDATE page_components pc
   SET rendered_html = replace(pc.rendered_html,'CONC 6.7.23','CONC 6.7.17'), updated_at=NOW()
  FROM pages p
 WHERE pc.page_id=p.id
   AND ((p.site_id='8ff093d5-1f19-453b-9439-a10379bbcd76'
         AND p.name IN ('rollover-rules','tool-rollover-limit-checker-guide','tool-rollover-limit-checker'))
     OR (p.site_id='ee4a8199-4f5b-4e2e-88ce-01e600721b74' AND p.name='tool-rollover-limit-checker'))
   AND pc.rendered_html LIKE '%CONC 6.7.23%';

UPDATE content_components
   SET html_template = replace(html_template,'CONC 6.7.23','CONC 6.7.17'), updated_at=NOW()
 WHERE id IN ('1fbbd1da-a467-468d-99dd-7e56cfeb78d9','6525121a-1d06-44b5-bd16-e551d45167b2');

-- 2. CPA reverse SECOND, scoped to the rows 696 changed — NOT the cpa tool.
UPDATE page_components pc
   SET content_data = replace(pc.content_data::text,'CONC 7.6.12','CONC 6.7.23')::jsonb, updated_at=NOW()
  FROM pages p
 WHERE pc.page_id=p.id AND p.site_id='8ff093d5-1f19-453b-9439-a10379bbcd76'
   AND p.name IN ('cant-pay','check-your-loan','continuous-payment-authority','how-to-complain',
                  'index','your-rights','tool-cpa-cancellation-checker-guide')
   AND pc.content_data::text LIKE '%CONC 7.6.12%';

UPDATE page_components pc
   SET rendered_html = replace(pc.rendered_html,'CONC 7.6.12','CONC 6.7.23'), updated_at=NOW()
  FROM pages p
 WHERE pc.page_id=p.id AND p.site_id='8ff093d5-1f19-453b-9439-a10379bbcd76'
   AND p.name IN ('cant-pay','check-your-loan','continuous-payment-authority','how-to-complain',
                  'index','your-rights','tool-cpa-cancellation-checker-guide')
   AND pc.rendered_html LIKE '%CONC 7.6.12%';

-- 3. The spec: re-current the superseded original; retire 696's row.
UPDATE site_specs SET is_current=false, superseded_at=NOW()
 WHERE site_id='8ff093d5-1f19-453b-9439-a10379bbcd76' AND aspect='content_direction'
   AND is_current AND created_by='lendzy_co_uk lane (migration 696)';

UPDATE site_specs SET is_current=true, superseded_at=NULL
 WHERE id = (SELECT id FROM site_specs
              WHERE site_id='8ff093d5-1f19-453b-9439-a10379bbcd76' AND aspect='content_direction'
                AND NOT is_current AND created_by <> 'lendzy_co_uk lane (migration 696)'
              ORDER BY superseded_at DESC NULLS LAST LIMIT 1);

DO $$
DECLARE c int; r int;
BEGIN
  SELECT COALESCE(sum((length(pc.content_data::text)-length(replace(pc.content_data::text,'CONC 6.7.17','')))/11),0) INTO c
    FROM pages p JOIN page_components pc ON pc.page_id=p.id
   WHERE p.site_id='8ff093d5-1f19-453b-9439-a10379bbcd76';
  IF c <> 6 THEN RAISE EXCEPTION '696 ROLLBACK: lendzy content_data CONC 6.7.17 = % (expected 6)', c; END IF;
  SELECT count(*) INTO r FROM site_specs
   WHERE site_id='8ff093d5-1f19-453b-9439-a10379bbcd76' AND aspect='content_direction' AND is_current
     AND data::text LIKE '%CONC 6.7.23%';
  IF r <> 1 THEN RAISE EXCEPTION '696 ROLLBACK: original content_direction not restored as current (matching rows: %)', r; END IF;
  -- the cpa tool's own correct citation must have survived the reverse untouched
  SELECT count(*) INTO r FROM content_components
   WHERE function='tool-cpa-cancellation-checker' AND html_template LIKE '%CONC 7.6.12%';
  IF r < 1 THEN RAISE EXCEPTION '696 ROLLBACK: the cpa tool template lost its pre-existing correct 7.6.12 — the scoping failed'; END IF;
  RAISE NOTICE '696 ROLLBACK OK: the two wrong citations are back on the record — this state is known-wrong and should not persist';
END $$;

COMMIT;
