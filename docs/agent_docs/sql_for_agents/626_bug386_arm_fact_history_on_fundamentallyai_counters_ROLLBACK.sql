-- 626_bug386_arm_fact_history_on_fundamentallyai_counters_ROLLBACK.sql
--
-- Reverses 626: strips `retain_history` and `history` from every fact in
-- fundamentallyai.com's current evidence_base.
--
-- Effect of rolling back: the five stale-render findings (11513, 10194, 428,
-- 483, 23 on capabilities/evidence-chart) return at ERROR severity, and a
-- rebuild of that page is refused again. Nothing else changes — the keys are
-- inert to every other consumer, and no binary depends on their presence.
--
-- Supersede-then-insert for the same reason as the forward file:
-- idx_site_specs_current is UNIQUE on (site_id, aspect) WHERE is_current.

BEGIN;

DO $$
DECLARE n int;
BEGIN
    SELECT count(*) INTO n
    FROM site_specs, jsonb_array_elements(data->'facts') f
    WHERE site_id = '199733a8-ac9c-4c30-b2ce-65ecdac6f3bd'
      AND aspect = 'evidence_base' AND is_current
      AND (f ? 'retain_history' OR f ? 'history');
    IF n = 0 THEN
        RAISE EXCEPTION 'bug386 arming ROLLBACK: no fact carries retain_history/history — nothing to roll back (already reversed?).';
    END IF;
END $$;

UPDATE site_specs
SET is_current = false, superseded_at = now()
WHERE site_id = '199733a8-ac9c-4c30-b2ce-65ecdac6f3bd'
  AND aspect = 'evidence_base' AND is_current;

WITH base AS (
    SELECT * FROM site_specs
     WHERE site_id = '199733a8-ac9c-4c30-b2ce-65ecdac6f3bd'
       AND aspect = 'evidence_base'
       AND is_current = false AND superseded_at IS NOT NULL
     ORDER BY superseded_at DESC
     LIMIT 1
)
INSERT INTO site_specs (site_id, aspect, data, source, source_agent, created_by, pinned, notes)
SELECT
    b.site_id,
    'evidence_base',
    jsonb_set(
        b.data,
        '{facts}',
        (
            SELECT jsonb_agg((f - 'retain_history') - 'history' ORDER BY ord)
            FROM jsonb_array_elements(b.data->'facts') WITH ORDINALITY AS t(f, ord)
        )
    ),
    'manual',
    NULL,
    'session-2026-08-25-bug386-arm-fact-history-ROLLBACK',
    b.pinned,
    'ROLLBACK of 626 (bugs_open/386 Phase 2): retain_history/history stripped from every fact. The five stale-render findings on capabilities/evidence-chart return at error severity and that page is refusable again.'
FROM base b;

DO $$
DECLARE n int; nfacts int;
BEGIN
    SELECT jsonb_array_length(data->'facts') INTO nfacts
    FROM site_specs
    WHERE site_id = '199733a8-ac9c-4c30-b2ce-65ecdac6f3bd'
      AND aspect = 'evidence_base' AND is_current;
    IF nfacts <> 16 THEN
        RAISE EXCEPTION 'bug386 arming ROLLBACK VERIFY: fact count changed to %.', nfacts;
    END IF;

    SELECT count(*) INTO n
    FROM site_specs, jsonb_array_elements(data->'facts') f
    WHERE site_id = '199733a8-ac9c-4c30-b2ce-65ecdac6f3bd'
      AND aspect = 'evidence_base' AND is_current
      AND (f ? 'retain_history' OR f ? 'history');
    IF n <> 0 THEN
        RAISE EXCEPTION 'bug386 arming ROLLBACK VERIFY: % fact(s) still carry the keys.', n;
    END IF;
END $$;

COMMIT;
