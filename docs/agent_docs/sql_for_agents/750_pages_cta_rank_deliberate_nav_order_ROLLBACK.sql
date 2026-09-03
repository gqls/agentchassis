-- ROLLBACK for 750 (bugs_open/436). Hand-apply only; the runner never applies
-- an UPPERCASE-suffixed sidecar.
--
-- ⚠ THIS DROPS ACKNOWLEDGEMENTS, WHICH ARE HUMAN JUDGEMENTS AND ARE NOT
-- RECOVERABLE FROM ANYWHERE ELSE. Every non-NULL cta_rank_deliberate_nav_order
-- is a person having looked at a site's primary button and accepted it; nothing
-- else records that they did. The check does not persist its own silence, and a
-- retracted work item records only that the condition cleared, not why.
-- So: PRINT THEM FIRST (the block below refuses to run blind), keep the output,
-- and expect every dropped acknowledgement to come back as a fresh
-- needs_human_review item on the next discovery pass.
--
-- ORDERING: drop the column only when no running binary reads it. A chassis
-- carrying check_cta_rank_anomaly's acknowledgement lookup will error on every
-- completeness pass for every site if the column is gone — and the discovery
-- runner FAILS THE WHOLE STEP on a check error, so this takes down all 46
-- checks, not just this one. Roll back the image first, or accept that.

BEGIN;

DO $$
DECLARE
    r      record;
    n      integer;
BEGIN
    SELECT count(*) INTO n FROM pages WHERE cta_rank_deliberate_nav_order IS NOT NULL;
    RAISE NOTICE '750_ROLLBACK: % acknowledgement(s) about to be destroyed:', n;
    FOR r IN
        SELECT s.domain, p.name, p.cta_rank_deliberate_nav_order AS ack, COALESCE(p.nav_order,100) AS nav
        FROM pages p JOIN sites s ON s.id = p.site_id
        WHERE p.cta_rank_deliberate_nav_order IS NOT NULL
        ORDER BY s.domain, p.name
    LOOP
        RAISE NOTICE '  %  %  acknowledged_at_nav=%  current_nav=%', r.domain, r.name, r.ack, r.nav;
    END LOOP;
END $$;

ALTER TABLE pages DROP COLUMN IF EXISTS cta_rank_deliberate_nav_order;

DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.columns
                WHERE table_name = 'pages' AND column_name = 'cta_rank_deliberate_nav_order') THEN
        RAISE EXCEPTION '750_ROLLBACK: column still present after DROP';
    END IF;
END $$;

COMMIT;
